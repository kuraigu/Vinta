#include "radio_service.h"
#include <winrt/Windows.Devices.Radios.h>
#include <winrt/Windows.Networking.Connectivity.h>

using namespace winrt;
using namespace Windows::Devices::Radios;
using namespace Windows::Networking::Connectivity;

static bool g_explicitAirplaneMode = false;

// Windows Radio Management COM API (for system-wide Airplane Mode)
static const GUID CLSID_RadioManagementAPI = { 0x581333f6, 0x28db, 0x41be, { 0xbc, 0x7a, 0xff, 0x20, 0x1f, 0x12, 0xf3, 0xf6 } };
static const GUID IID_IRadioManager = { 0xdb3afbfb, 0x08e6, 0x46c6, { 0xaa, 0x70, 0xbf, 0x9a, 0x34, 0xc3, 0x0a, 0xb7 } };

struct IRadioManager : public IUnknown {
    virtual HRESULT STDMETHODCALLTYPE IsRMSupported(DWORD *pdwState) = 0;
    virtual HRESULT STDMETHODCALLTYPE GetUIRadioInstances(void **ppInstances) = 0;
    virtual HRESULT STDMETHODCALLTYPE GetSystemRadioState(BOOL *pbEnabled, int *param_2, int *param_3) = 0;
    virtual HRESULT STDMETHODCALLTYPE SetSystemRadioState(BOOL bEnabled) = 0;
    virtual HRESULT STDMETHODCALLTYPE Refresh(void) = 0;
    virtual HRESULT STDMETHODCALLTYPE OnHardwareSliderChange(int param_1, int param_2) = 0;
};

static bool SetSystemAirplaneMode(bool enableAirplane) {
    ComInitScope com;
    IRadioManager* pRadioMgr = nullptr;
    HRESULT hr = CoCreateInstance(
        CLSID_RadioManagementAPI,
        nullptr,
        CLSCTX_INPROC_SERVER | CLSCTX_LOCAL_SERVER,
        IID_IRadioManager,
        reinterpret_cast<void**>(&pRadioMgr)
    );
    if (SUCCEEDED(hr) && pRadioMgr) {
        // bEnabled: TRUE = Radios On (Airplane Off), FALSE = Radios Off (Airplane On)
        BOOL bSystemRadiosEnabled = enableAirplane ? FALSE : TRUE;
        pRadioMgr->SetSystemRadioState(bSystemRadiosEnabled);
        pRadioMgr->Release();
        return true;
    }
    return false;
}

static bool GetSystemAirplaneMode(bool& isAirplane) {
    ComInitScope com;
    IRadioManager* pRadioMgr = nullptr;
    HRESULT hr = CoCreateInstance(
        CLSID_RadioManagementAPI,
        nullptr,
        CLSCTX_INPROC_SERVER | CLSCTX_LOCAL_SERVER,
        IID_IRadioManager,
        reinterpret_cast<void**>(&pRadioMgr)
    );
    if (SUCCEEDED(hr) && pRadioMgr) {
        BOOL bSystemRadiosEnabled = TRUE;
        int p2 = 0, p3 = 0;
        hr = pRadioMgr->GetSystemRadioState(&bSystemRadiosEnabled, &p2, &p3);
        pRadioMgr->Release();
        if (SUCCEEDED(hr)) {
            isAirplane = (bSystemRadiosEnabled == FALSE);
            return true;
        }
    }
    return false;
}

int SetRadioState(int radioKind, int targetState) {
    try {
        ApartmentScope apt;

        if (radioKind == 255) {
            // Airplane Mode toggle
            bool enableAirplane = (targetState == 2);
            g_explicitAirplaneMode = enableAirplane;
            SetSystemAirplaneMode(enableAirplane);

            // Also set individual radios for immediate hardware enforcement
            auto access = Radio::RequestAccessAsync().get();
            if (access == RadioAccessStatus::Allowed) {
                auto radios = Radio::GetRadiosAsync().get();
                for (auto&& radio : radios) {
                    auto kind = radio.Kind();
                    if (enableAirplane) {
                        try {
                            radio.SetStateAsync(RadioState::Off).get();
                        } catch (...) {}
                    } else {
                        if (kind == RadioKind::WiFi || kind == RadioKind::Bluetooth) {
                            try {
                                radio.SetStateAsync(RadioState::On).get();
                            } catch (...) {}
                        }
                    }
                }
            }
            return 1;
        }

        // Individual radio toggle (WiFi = 1, Bluetooth = 3)
        if (targetState == 1) {
            g_explicitAirplaneMode = false;
            // If turning an individual radio on, ensure Airplane Mode is disabled in OS
            SetSystemAirplaneMode(false);
        }

        auto access = Radio::RequestAccessAsync().get();
        if (access != RadioAccessStatus::Allowed) {
            return 0;
        }

        auto radios = Radio::GetRadiosAsync().get();
        RadioState desiredState = (targetState == 1) ? RadioState::On : RadioState::Off;

        bool anyFound = false;
        for (auto&& radio : radios) {
            if (static_cast<int>(radio.Kind()) == radioKind) {
                anyFound = true;
                try {
                    radio.SetStateAsync(desiredState).get();
                } catch (...) {}
            }
        }
        return anyFound ? 1 : 0;
    } catch (...) {
        return 0;
    }
}

int GetRadiosStateJson(char* outBuf, int maxBufLen) {
    if (!outBuf || maxBufLen <= 0) return 0;
    try {
        ApartmentScope apt;
        auto access = Radio::RequestAccessAsync().get();
        if (access != RadioAccessStatus::Allowed) {
            return 0;
        }

        auto radios = Radio::GetRadiosAsync().get();
        bool hasWifi = false;
        bool wifiEnabled = false;
        bool hasBt = false;
        bool btEnabled = false;

        for (auto&& radio : radios) {
            auto kind = radio.Kind();
            auto state = radio.State();

            if (kind == RadioKind::WiFi) {
                hasWifi = true;
                if (state == RadioState::On) {
                    wifiEnabled = true;
                }
            } else if (kind == RadioKind::Bluetooth) {
                hasBt = true;
                if (state == RadioState::On) {
                    btEnabled = true;
                }
            }
        }

        // Query Wi-Fi connection profile in-process via WinRT NetworkInformation (0ms latency, no netsh popup)
        bool wifiConnected = false;
        std::string wifiSsid = "";
        int wifiSignal = 0;

        if (wifiEnabled) {
            try {
                auto profile = NetworkInformation::GetInternetConnectionProfile();
                if (profile && profile.IsWlanConnectionProfile()) {
                    auto level = profile.GetNetworkConnectivityLevel();
                    if (level != NetworkConnectivityLevel::None) {
                        wifiConnected = true;
                        auto wlanDetails = profile.WlanConnectionProfileDetails();
                        if (wlanDetails) {
                            wifiSsid = winrt::to_string(wlanDetails.GetConnectedSsid());
                        }
                        if (wifiSsid.empty()) {
                            wifiSsid = winrt::to_string(profile.ProfileName());
                        }
                        auto bars = profile.GetSignalBars();
                        if (bars) {
                            wifiSignal = static_cast<int>(bars.Value()) * 20; // 1-5 bars -> 20-100%
                        } else {
                            wifiSignal = 80;
                        }
                    }
                }
            } catch (...) {}
        }

        // Query system-level Airplane Mode state
        bool isAirplane = false;
        if (!GetSystemAirplaneMode(isAirplane)) {
            isAirplane = g_explicitAirplaneMode;
            if (!isAirplane && hasWifi && !wifiEnabled && (!hasBt || !btEnabled)) {
                isAirplane = true;
            }
        }
        if (wifiEnabled || (hasBt && btEnabled)) {
            isAirplane = false;
            g_explicitAirplaneMode = false;
        }

        std::string json = "{";
        json += "\"hasWifi\":" + std::string(hasWifi ? "true" : "false") + ",";
        json += "\"wifiEnabled\":" + std::string(wifiEnabled ? "true" : "false") + ",";
        json += "\"wifiConnected\":" + std::string(wifiConnected ? "true" : "false") + ",";
        json += "\"wifiSsid\":\"" + EscapeJsonString(wifiSsid) + "\",";
        json += "\"wifiSignal\":" + std::to_string(wifiSignal) + ",";
        json += "\"hasBluetooth\":" + std::string(hasBt ? "true" : "false") + ",";
        json += "\"bluetoothEnabled\":" + std::string(btEnabled ? "true" : "false") + ",";
        json += "\"airplaneMode\":" + std::string(isAirplane ? "true" : "false");
        json += "}";

        if (json.length() >= static_cast<size_t>(maxBufLen)) return 0;
        strcpy_s(outBuf, maxBufLen, json.c_str());
        return 1;
    } catch (...) {
        return 0;
    }
}
