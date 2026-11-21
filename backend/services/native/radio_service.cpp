/**
 * @file radio_service.cpp
 * @brief WinRT and COM implementation for Wi-Fi, Bluetooth, and Airplane Mode.
 *
 * Caches radio access permissions to avoid redundant out-of-process RPC round-trips
 * during status polling, dispatches multi-radio state changes in parallel, and formats
 * connection status without temporary heap string concatenations.
 */

#include "radio_service.h"
#include <winrt/Windows.Devices.Radios.h>
#include <winrt/Windows.Foundation.h>
#include <winrt/Windows.Networking.Connectivity.h>

#include <atomic>
#include <cstdio>
#include <vector>

using namespace winrt;
using namespace Windows::Devices::Radios;
using namespace Windows::Networking::Connectivity;

static std::atomic<bool> g_explicitAirplaneMode{false};
static std::atomic<bool> g_radioAccessGranted{false};

static const GUID CLSID_RadioManagementAPI = {
    0x581333f6, 0x28db, 0x41be, {0xbc, 0x7a, 0xff, 0x20, 0x1f, 0x12, 0xf3, 0xf6}};
static const GUID IID_IRadioManager = {0xdb3afbfb, 0x08e6, 0x46c6, {0xaa, 0x70, 0xbf, 0x9a, 0x34, 0xc3, 0x0a, 0xb7}};

/**
 * @struct IRadioManager
 * @brief Undocumented Windows COM interface for system-wide Airplane Mode control.
 */
struct IRadioManager : public IUnknown {
    virtual HRESULT STDMETHODCALLTYPE IsRMSupported(DWORD* pdwState) = 0;
    virtual HRESULT STDMETHODCALLTYPE GetUIRadioInstances(void** ppInstances) = 0;
    virtual HRESULT STDMETHODCALLTYPE GetSystemRadioState(BOOL* pbEnabled, int* p2, int* p3) = 0;
    virtual HRESULT STDMETHODCALLTYPE SetSystemRadioState(BOOL bEnabled) = 0;
    virtual HRESULT STDMETHODCALLTYPE Refresh(void) = 0;
    virtual HRESULT STDMETHODCALLTYPE OnHardwareSliderChange(int p1, int p2) = 0;
};

/**
 * @brief Queries or updates the system-wide radio master switch via COM IRadioManager.
 * @param setMode True to apply enableAirplane, false to query current state into isAirplaneOut.
 */
static bool AccessSystemAirplaneMode(bool setMode, bool enableAirplane, bool& isAirplaneOut) {
    IRadioManager* pMgr = nullptr;
    HRESULT hr = CoCreateInstance(CLSID_RadioManagementAPI, nullptr, CLSCTX_INPROC_SERVER | CLSCTX_LOCAL_SERVER,
                                  IID_IRadioManager, reinterpret_cast<void**>(&pMgr));
    if (FAILED(hr) || !pMgr)
        return false;

    if (setMode) {
        hr = pMgr->SetSystemRadioState(enableAirplane ? FALSE : TRUE);
    } else {
        BOOL radiosEnabled = TRUE;
        int p2 = 0, p3 = 0;
        hr = pMgr->GetSystemRadioState(&radiosEnabled, &p2, &p3);
        if (SUCCEEDED(hr)) {
            isAirplaneOut = (radiosEnabled == FALSE);
        }
    }
    pMgr->Release();
    return SUCCEEDED(hr);
}

/**
 * @brief Checks and caches WinRT Radio access permission so status polls skip RPC overhead.
 */
static bool EnsureRadioAccess() {
    if (g_radioAccessGranted.load(std::memory_order_acquire)) {
        return true;
    }
    if (Radio::RequestAccessAsync().get() == RadioAccessStatus::Allowed) {
        g_radioAccessGranted.store(true, std::memory_order_release);
        return true;
    }
    return false;
}

int SetRadioState(int radioKind, int targetState) {
    try {
        ApartmentScope apt;
        bool unused = false;

        if (radioKind == 255) {
            bool enableAirplane = (targetState == 2);
            g_explicitAirplaneMode.store(enableAirplane, std::memory_order_release);
            AccessSystemAirplaneMode(true, enableAirplane, unused);

            if (EnsureRadioAccess()) {
                auto radios = Radio::GetRadiosAsync().get();
                std::vector<Windows::Foundation::IAsyncOperation<RadioAccessStatus>> pendingOps;
                for (auto&& radio : radios) {
                    auto kind = radio.Kind();
                    if (enableAirplane || kind == RadioKind::WiFi || kind == RadioKind::Bluetooth) {
                        try {
                            pendingOps.push_back(
                                radio.SetStateAsync(enableAirplane ? RadioState::Off : RadioState::On));
                        } catch (...) {
                        }
                    }
                }
                for (auto& op : pendingOps) {
                    try {
                        op.get();
                    } catch (...) {
                    }
                }
            }
            return 1;
        }

        if (targetState == 1) {
            g_explicitAirplaneMode.store(false, std::memory_order_release);
            AccessSystemAirplaneMode(true, false, unused);
        }

        if (!EnsureRadioAccess())
            return 0;

        auto radios = Radio::GetRadiosAsync().get();
        RadioState desired = (targetState == 1) ? RadioState::On : RadioState::Off;
        std::vector<Windows::Foundation::IAsyncOperation<RadioAccessStatus>> pendingOps;

        for (auto&& radio : radios) {
            if (static_cast<int>(radio.Kind()) == radioKind) {
                try {
                    pendingOps.push_back(radio.SetStateAsync(desired));
                } catch (...) {
                }
            }
        }
        for (auto& op : pendingOps) {
            try {
                op.get();
            } catch (...) {
            }
        }
        return !pendingOps.empty() ? 1 : 0;
    } catch (...) {
        return 0;
    }
}

int GetRadiosStateJson(char* outBuf, int maxBufLen) {
    if (!outBuf || maxBufLen <= 0)
        return 0;
    try {
        ApartmentScope apt;
        if (!EnsureRadioAccess())
            return 0;

        bool hasWifi = false, wifiEnabled = false, hasBt = false, btEnabled = false;
        for (auto&& radio : Radio::GetRadiosAsync().get()) {
            auto kind = radio.Kind();
            bool isOn = (radio.State() == RadioState::On);
            if (kind == RadioKind::WiFi) {
                hasWifi = true;
                wifiEnabled |= isOn;
            } else if (kind == RadioKind::Bluetooth) {
                hasBt = true;
                btEnabled |= isOn;
            }
        }

        bool wifiConnected = false;
        std::string wifiSsid;
        int wifiSignal = 0;

        if (wifiEnabled) {
            try {
                auto profile = NetworkInformation::GetInternetConnectionProfile();
                if (profile && profile.IsWlanConnectionProfile() &&
                    profile.GetNetworkConnectivityLevel() != NetworkConnectivityLevel::None) {
                    wifiConnected = true;
                    if (auto wlan = profile.WlanConnectionProfileDetails()) {
                        wifiSsid = winrt::to_string(wlan.GetConnectedSsid());
                    }
                    if (wifiSsid.empty()) {
                        wifiSsid = winrt::to_string(profile.ProfileName());
                    }
                    auto bars = profile.GetSignalBars();
                    wifiSignal = bars ? static_cast<int>(bars.Value()) * 20 : 80;
                }
            } catch (...) {
            }
        }

        bool isAirplane = false;
        if (!AccessSystemAirplaneMode(false, false, isAirplane)) {
            isAirplane = g_explicitAirplaneMode.load(std::memory_order_acquire) ||
                         (hasWifi && !wifiEnabled && (!hasBt || !btEnabled));
        }
        if (wifiEnabled || (hasBt && btEnabled)) {
            isAirplane = false;
            g_explicitAirplaneMode.store(false, std::memory_order_release);
        }

        std::string escapedSsid = EscapeJsonString(wifiSsid);
        int written =
            std::snprintf(outBuf, static_cast<size_t>(maxBufLen),
                          "{\"hasWifi\":%s,\"wifiEnabled\":%s,\"wifiConnected\":%s,\"wifiSsid\":\"%s\","
                          "\"wifiSignal\":%d,\"hasBluetooth\":%s,\"bluetoothEnabled\":%s,\"airplaneMode\":%s}",
                          hasWifi ? "true" : "false", wifiEnabled ? "true" : "false", wifiConnected ? "true" : "false",
                          escapedSsid.c_str(), wifiSignal, hasBt ? "true" : "false", btEnabled ? "true" : "false",
                          isAirplane ? "true" : "false");

        return (written > 0 && written < maxBufLen) ? 1 : 0;
    } catch (...) {
        return 0;
    }
}