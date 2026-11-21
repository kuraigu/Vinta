/**
 * @file audio_peak_service.cpp
 * @brief WASAPI audio peak polling and default endpoint change tracking.
 *
 * Samples the active multimedia output device at 20 Hz while media is playing
 * and 5 Hz when paused, automatically rebinding to the new default endpoint
 * whenever the user switches audio output devices.
 */

#include "audio_peak_service.h"
#include "winrt_common.h"
#include "media_service.h"

#include <mmdeviceapi.h>
#include <endpointvolume.h>
#include <audioclient.h>

#include <atomic>
#include <condition_variable>
#include <cstring>
#include <mutex>
#include <thread>

static std::atomic<NativeAudioPeakCallback> g_peakCb{nullptr};
static std::atomic<bool> g_peakRunning{false};
static std::atomic<bool> g_deviceDirty{true};
static std::mutex g_sleepMutex;
static std::condition_variable g_sleepCv;
static std::thread g_audioThread;

/**
 * @class EndpointChangeNotice
 * @brief Lightweight IMMNotificationClient that flags when the default output device changes.
 */
class EndpointChangeNotice final : public IMMNotificationClient {
public:
    ULONG STDMETHODCALLTYPE AddRef() override { return 1; }
    ULONG STDMETHODCALLTYPE Release() override { return 1; }

    HRESULT STDMETHODCALLTYPE QueryInterface(REFIID riid, VOID** ppvInterface) override {
        if (!ppvInterface)
            return E_POINTER;
        if (riid == __uuidof(IUnknown) || riid == __uuidof(IMMNotificationClient)) {
            *ppvInterface = static_cast<IMMNotificationClient*>(this);
            return S_OK;
        }
        *ppvInterface = nullptr;
        return E_NOINTERFACE;
    }

    HRESULT STDMETHODCALLTYPE OnDefaultDeviceChanged(EDataFlow flow, ERole role, LPCWSTR) override {
        if (flow == eRender && (role == eMultimedia || role == eConsole)) {
            g_deviceDirty.store(true, std::memory_order_release);
        }
        return S_OK;
    }

    HRESULT STDMETHODCALLTYPE OnDeviceAdded(LPCWSTR) override { return S_OK; }

    HRESULT STDMETHODCALLTYPE OnDeviceRemoved(LPCWSTR) override {
        g_deviceDirty.store(true, std::memory_order_release);
        return S_OK;
    }

    HRESULT STDMETHODCALLTYPE OnDeviceStateChanged(LPCWSTR, DWORD) override {
        g_deviceDirty.store(true, std::memory_order_release);
        return S_OK;
    }

    HRESULT STDMETHODCALLTYPE OnPropertyValueChanged(LPCWSTR, const PROPERTYKEY) override { return S_OK; }
};

/**
 * @brief Converts a float peak level to raw uint32 bits and invokes the registered callback.
 * @param value Normalized peak value in the range [0.0, 1.0].
 */
static inline void EmitPeakValue(float value) {
    NativeAudioPeakCallback cb = g_peakCb.load(std::memory_order_acquire);
    if (!cb)
        return;
    uint32_t bits = 0;
    std::memcpy(&bits, &value, sizeof(bits));
    cb(bits);
}

/**
 * @brief Sleeps for the specified duration or wakes early if the service is stopping.
 * @param ms Sleep duration in milliseconds.
 * @return True if the full sleep completed, false if interrupted by shutdown.
 */
static bool WaitInterruptible(int ms) {
    std::unique_lock<std::mutex> lock(g_sleepMutex);
    return !g_sleepCv.wait_for(lock, std::chrono::milliseconds(ms),
                               [] { return !g_peakRunning.load(std::memory_order_acquire); });
}

/**
 * @brief Background loop that queries IAudioMeterInformation on the default render endpoint.
 */
static void AudioPeakWorker() {
    ComInitScope com;

    IMMDeviceEnumerator* pEnumerator = nullptr;
    HRESULT hr = CoCreateInstance(__uuidof(MMDeviceEnumerator), nullptr, CLSCTX_INPROC_SERVER,
                                  __uuidof(IMMDeviceEnumerator), reinterpret_cast<void**>(&pEnumerator));
    if (FAILED(hr) || !pEnumerator) {
        return;
    }

    EndpointChangeNotice notifyClient;
    bool notifyRegistered = SUCCEEDED(pEnumerator->RegisterEndpointNotificationCallback(&notifyClient));

    IMMDevice* pDefaultDevice = nullptr;
    IAudioMeterInformation* pDefaultMeter = nullptr;

    auto releaseMeter = [&]() {
        if (pDefaultMeter) {
            pDefaultMeter->Release();
            pDefaultMeter = nullptr;
        }
        if (pDefaultDevice) {
            pDefaultDevice->Release();
            pDefaultDevice = nullptr;
        }
    };

    bool wasPlaying = false;

    while (g_peakRunning.load(std::memory_order_acquire)) {
        if (!IsMediaPlayingState()) {
            if (wasPlaying) {
                wasPlaying = false;
                EmitPeakValue(0.0f);
            }
            if (!WaitInterruptible(250))
                break;
            continue;
        }
        wasPlaying = true;

        if (!WaitInterruptible(50))
            break; // 20 FPS peak polling when playing media

        if (g_deviceDirty.exchange(false, std::memory_order_acq_rel) || !pDefaultMeter) {
            releaseMeter();
            if (SUCCEEDED(pEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &pDefaultDevice)) &&
                pDefaultDevice) {
                pDefaultDevice->Activate(__uuidof(IAudioMeterInformation), CLSCTX_INPROC_SERVER, nullptr,
                                         reinterpret_cast<void**>(&pDefaultMeter));
            }
        }

        float maxPeak = 0.0f;
        if (pDefaultMeter) {
            float peak = 0.0f;
            if (FAILED(pDefaultMeter->GetPeakValue(&peak))) {
                releaseMeter();
                g_deviceDirty.store(true, std::memory_order_release);
            } else {
                maxPeak = peak;
            }
        }

        EmitPeakValue(maxPeak);
    }

    releaseMeter();
    if (notifyRegistered) {
        pEnumerator->UnregisterEndpointNotificationCallback(&notifyClient);
    }
    pEnumerator->Release();
}

void StartAudioPeakService(NativeAudioPeakCallback cb) {
    if (g_peakRunning.exchange(true, std::memory_order_acq_rel)) {
        return;
    }
    if (g_audioThread.joinable()) {
        g_audioThread.join();
    }
    g_peakCb.store(cb, std::memory_order_release);
    g_deviceDirty.store(true, std::memory_order_release);
    g_audioThread = std::thread(AudioPeakWorker);
}

void StopAudioPeakService() {
    {
        std::lock_guard<std::mutex> sleepLock(g_sleepMutex);
        if (!g_peakRunning.exchange(false, std::memory_order_acq_rel)) {
            return;
        }
    }
    g_peakCb.store(nullptr, std::memory_order_release);
    g_sleepCv.notify_all();
    if (g_audioThread.joinable()) {
        g_audioThread.join();
    }
}