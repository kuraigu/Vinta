#ifndef WINRT_BRIDGE_H
#define WINRT_BRIDGE_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#ifdef _WIN32
#define BRIDGE_API __declspec(dllexport)
#else
#define BRIDGE_API
#endif

typedef void (*NativeMediaCallback)(const char* jsonPayload);
typedef void (*NativeNotificationCallback)(const char* jsonPayload);
typedef void (*NativeAudioPeakCallback)(uint32_t peakBits);

BRIDGE_API int StartNativeWinRTServices(
    NativeMediaCallback mediaCb,
    NativeNotificationCallback notifCb,
    NativeAudioPeakCallback peakCb
);

BRIDGE_API void StopNativeWinRTServices(void);

BRIDGE_API int SendNativeMediaCommand(const char* command);

BRIDGE_API int SetNativeRadioState(int radioKind, int targetState);

BRIDGE_API int GetNativeRadiosStatus(char* outJson, int maxLen);

BRIDGE_API int SelectNativeMediaSession(const char* sessionId);

BRIDGE_API int FocusNativeMediaApp(const char* sourceApp);
BRIDGE_API int FocusNativeApp(const char* sourceApp, const char* aumid);

BRIDGE_API int ExtractAppIconToFile(const char* sourceApp, const char* aumid, const wchar_t* destFilePath, int targetSize);
BRIDGE_API void ClearNativeMediaArtCache(void);
BRIDGE_API void RequeryNativeMediaArtwork(void);
BRIDGE_API void ForceEmitNativeMediaState(void);
BRIDGE_API int SendNativeToast(const wchar_t* title, const wchar_t* message);

#ifdef __cplusplus
}
#endif

#endif // WINRT_BRIDGE_H
