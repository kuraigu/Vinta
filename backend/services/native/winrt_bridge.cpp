#include "winrt_bridge.h"
#include "notification_service.h"
#include "media_service.h"
#include "audio_peak_service.h"
#include "radio_service.h"
#include "icon_service.h"

static bool g_bridgeRunning = false;

extern "C" BRIDGE_API int StartNativeWinRTServices(
    NativeMediaCallback mediaCb,
    NativeNotificationCallback notifCb,
    NativeAudioPeakCallback peakCb
) {
    if (g_bridgeRunning) return 0;
    g_bridgeRunning = true;

    StartMediaService(mediaCb);
    StartNotificationService(notifCb);
    StartAudioPeakService(peakCb);

    return 1;
}

extern "C" BRIDGE_API void StopNativeWinRTServices(void) {
    if (!g_bridgeRunning) return;
    g_bridgeRunning = false;

    StopAudioPeakService();
    StopNotificationService();
    StopMediaService();
}

extern "C" BRIDGE_API int SendNativeMediaCommand(const char* command) {
    return ExecuteMediaCommand(command);
}

extern "C" BRIDGE_API int SetNativeRadioState(int radioKind, int targetState) {
    return SetRadioState(radioKind, targetState);
}

extern "C" BRIDGE_API int GetNativeRadiosStatus(char* outJson, int maxLen) {
    return GetRadiosStateJson(outJson, maxLen);
}

extern "C" BRIDGE_API int SelectNativeMediaSession(const char* sessionId) {
    return SelectMediaSession(sessionId);
}

extern "C" BRIDGE_API int FocusNativeMediaApp(const char* sourceApp) {
    return FocusMediaApp(sourceApp);
}

extern "C" BRIDGE_API int FocusNativeApp(const char* sourceApp, const char* aumid) {
    return FocusApp(sourceApp, aumid);
}

extern "C" BRIDGE_API int ExtractAppIconToFile(const char* sourceApp, const char* aumid, const wchar_t* destFilePath, int targetSize) {
    try {
        return ExtractAppIconNativeToFile(sourceApp, aumid, destFilePath, targetSize);
    } catch (...) {
        return 0;
    }
}

extern "C" BRIDGE_API void ClearNativeMediaArtCache(void) {
    FlushMediaArtCache();
}

extern "C" BRIDGE_API void RequeryNativeMediaArtwork(void) {
    RequeryMediaArtwork();
}

extern "C" BRIDGE_API void ForceEmitNativeMediaState(void) {
    ForceEmitMediaState();
}
