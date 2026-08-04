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

#ifdef __cplusplus
}
#endif

#endif // WINRT_BRIDGE_H
