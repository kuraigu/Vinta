#include "audio_peak_service.h"
#include "winrt_common.h"
#include "media_service.h"

#include <mmdeviceapi.h>
#include <endpointvolume.h>
#include <audioclient.h>

#include <atomic>
#include <cstring>
#include <thread>

static NativeAudioPeakCallback g_peakCb = nullptr;
static std::atomic<bool> g_peakRunning{false};
static std::thread g_audioThread;

static void AudioPeakWorker() {
    ComInitScope com;
    IMMDeviceEnumerator* pEnumerator = nullptr;
    HRESULT hr = CoCreateInstance(__uuidof(MMDeviceEnumerator), NULL, CLSCTX_ALL, __uuidof(IMMDeviceEnumerator), (void**)&pEnumerator);
    if (FAILED(hr) || !pEnumerator) {
        return;
    }

    IMMDevice* pDefaultDevice = nullptr;
    IAudioMeterInformation* pDefaultMeter = nullptr;

    while (g_peakRunning) {
        if (!IsMediaPlayingState()) {
            std::this_thread::sleep_for(std::chrono::milliseconds(200));
            float zeroPeak = 0.0f;
            if (g_peakCb) {
                uint32_t bits = 0;
                std::memcpy(&bits, &zeroPeak, sizeof(bits));
                g_peakCb(bits);
            }
            continue;
        }

        std::this_thread::sleep_for(std::chrono::milliseconds(50)); // 20 FPS peak polling when playing media

        if (!pDefaultMeter) {
            if (pDefaultDevice) { pDefaultDevice->Release(); pDefaultDevice = nullptr; }
            if (SUCCEEDED(pEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &pDefaultDevice)) && pDefaultDevice) {
                pDefaultDevice->Activate(__uuidof(IAudioMeterInformation), CLSCTX_ALL, NULL, (void**)&pDefaultMeter);
            }
        }

        float maxPeak = 0.0f;
        if (pDefaultMeter) {
            float peak = 0.0f;
            if (FAILED(pDefaultMeter->GetPeakValue(&peak))) {
                pDefaultMeter->Release();
                pDefaultMeter = nullptr;
            } else {
                maxPeak = peak;
            }
        }

        if (g_peakCb) {
            uint32_t bits = 0;
            std::memcpy(&bits, &maxPeak, sizeof(bits));
            g_peakCb(bits);
        }
    }

    if (pDefaultMeter) pDefaultMeter->Release();
    if (pDefaultDevice) pDefaultDevice->Release();
    if (pEnumerator) pEnumerator->Release();
}

void StartAudioPeakService(NativeAudioPeakCallback cb) {
    if (g_peakRunning) return;
    g_peakCb = cb;
    g_peakRunning = true;
    g_audioThread = std::thread(AudioPeakWorker);
}

void StopAudioPeakService() {
    if (!g_peakRunning) return;
    g_peakRunning = false;
    if (g_audioThread.joinable()) {
        g_audioThread.join();
    }
}
