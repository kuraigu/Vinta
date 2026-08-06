#include "winrt_bridge.h"
#include "notification_service.h"
#include "media_service.h"
#include "audio_peak_service.h"

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
