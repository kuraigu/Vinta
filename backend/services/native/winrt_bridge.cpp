/**
 * @file winrt_bridge.cpp
 * @brief Thread-safe C-ABI boundary and exception firewall for native WinRT services.
 *
 * Guards service startup and teardown with a mutex so concurrent Go calls cannot
 * race, and wraps every exported C function in an exception handler so WinRT or
 * STL exceptions never escape across the CGO/syscall boundary.
 */

#include "winrt_bridge.h"
#include "notification_service.h"
#include "media_service.h"
#include "audio_peak_service.h"
#include "radio_service.h"
#include "icon_service.h"

#include <atomic>
#include <mutex>

static std::mutex g_lifecycleMutex;
static std::atomic<bool> g_bridgeRunning{false};

/**
 * @brief Executes an integer-returning callable and catches any C++ exceptions at the ABI boundary.
 */
template <typename Fn> static inline int SafeInvokeInt(Fn&& fn) noexcept {
    try {
        return fn();
    } catch (...) {
        return 0;
    }
}

/**
 * @brief Executes a void-returning callable and catches any C++ exceptions at the ABI boundary.
 */
template <typename Fn> static inline void SafeInvokeVoid(Fn&& fn) noexcept {
    try {
        fn();
    } catch (...) {
    }
}

extern "C" BRIDGE_API int StartNativeWinRTServices(NativeMediaCallback mediaCb, NativeNotificationCallback notifCb,
                                                   NativeAudioPeakCallback peakCb) {
    return SafeInvokeInt([&]() {
        std::lock_guard<std::mutex> lock(g_lifecycleMutex);
        if (g_bridgeRunning.exchange(true, std::memory_order_acq_rel)) {
            return 0;
        }
        StartMediaService(mediaCb);
        StartNotificationService(notifCb);
        StartAudioPeakService(peakCb);
        return 1;
    });
}

extern "C" BRIDGE_API void StopNativeWinRTServices(void) {
    SafeInvokeVoid([]() {
        std::lock_guard<std::mutex> lock(g_lifecycleMutex);
        if (!g_bridgeRunning.exchange(false, std::memory_order_acq_rel)) {
            return;
        }
        StopAudioPeakService();
        StopNotificationService();
        StopMediaService();
    });
}

extern "C" BRIDGE_API int SendNativeMediaCommand(const char* command) {
    return SafeInvokeInt([&]() { return command ? ExecuteMediaCommand(command) : 0; });
}

extern "C" BRIDGE_API int SetNativeRadioState(int radioKind, int targetState) {
    return SafeInvokeInt([&]() { return SetRadioState(radioKind, targetState); });
}

extern "C" BRIDGE_API int GetNativeRadiosStatus(char* outJson, int maxLen) {
    return SafeInvokeInt([&]() { return GetRadiosStateJson(outJson, maxLen); });
}

extern "C" BRIDGE_API int SelectNativeMediaSession(const char* sessionId) {
    return SafeInvokeInt([&]() { return sessionId ? SelectMediaSession(sessionId) : 0; });
}

extern "C" BRIDGE_API int FocusNativeMediaApp(const char* sourceApp) {
    return SafeInvokeInt([&]() { return sourceApp ? FocusMediaApp(sourceApp) : 0; });
}

extern "C" BRIDGE_API int FocusNativeApp(const char* sourceApp, const char* aumid) {
    return SafeInvokeInt([&]() { return FocusApp(sourceApp, aumid); });
}

extern "C" BRIDGE_API int ExtractAppIconToFile(const char* sourceApp, const char* aumid, const wchar_t* destFilePath,
                                               int targetSize) {
    return SafeInvokeInt(
        [&]() { return destFilePath ? ExtractAppIconNativeToFile(sourceApp, aumid, destFilePath, targetSize) : 0; });
}

extern "C" BRIDGE_API void ClearNativeMediaArtCache(void) {
    SafeInvokeVoid([]() { FlushMediaArtCache(); });
}

extern "C" BRIDGE_API void RequeryNativeMediaArtwork(void) {
    SafeInvokeVoid([]() { RequeryMediaArtwork(); });
}

extern "C" BRIDGE_API void ForceEmitNativeMediaState(void) {
    SafeInvokeVoid([]() { ForceEmitMediaState(); });
}

extern "C" BRIDGE_API int SendNativeToast(const wchar_t* title, const wchar_t* message) {
    return SafeInvokeInt([&]() { return SendToastNotification(title, message); });
}

extern "C" BRIDGE_API int GetNativeAudioSessions(char* outJson, int maxLen) {
    return SafeInvokeInt([&]() { return GetAllAudioSessionsJson(outJson, maxLen); });
}

extern "C" BRIDGE_API int SetNativeSessionVolume(uint32_t pid, int volumePercent) {
    return SafeInvokeInt([&]() { return SetAppAudioVolumeByPID(pid, volumePercent); });
}

extern "C" BRIDGE_API int SetNativeSessionMute(uint32_t pid, int mute) {
    return SafeInvokeInt([&]() { return SetAppAudioMuteByPID(pid, mute != 0); });
}

extern "C" BRIDGE_API int GetNativeAudioEndpoints(char* outJson, int maxLen) {
    return SafeInvokeInt([&]() { return GetAudioEndpointsJson(outJson, maxLen); });
}

extern "C" BRIDGE_API int SetNativeDefaultAudioEndpoint(const wchar_t* endpointId) {
    return SafeInvokeInt([&]() { return endpointId ? SetDefaultAudioEndpoint(endpointId) : 0; });
}

extern "C" BRIDGE_API int SetNativeMasterAudioVolume(int percent) {
    return SafeInvokeInt([&]() { return SetMasterVolumeNative(percent); });
}

extern "C" BRIDGE_API int ToggleNativeMasterAudioMute(void) {
    return SafeInvokeInt([]() { return ToggleMasterMuteNative(); });
}