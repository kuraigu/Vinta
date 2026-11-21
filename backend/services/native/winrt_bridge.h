/**
 * @file winrt_bridge.h
 * @brief Exported C-ABI interface for the native WinRT companion DLL.
 *
 * Exposes flat C functions loaded by the Go backend to start and stop
 * background media, notification, and audio peak workers, as well as
 * on-demand calls for radio control, app icon extraction, and volume mixing.
 */

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

/** @brief Callback invoked when media playback metadata or state changes. */
typedef void (*NativeMediaCallback)(const char* jsonPayload);

/** @brief Callback invoked when a toast notification arrives or updates. */
typedef void (*NativeNotificationCallback)(const char* jsonPayload);

/** @brief Callback invoked with raw IEEE-754 float bits of the master audio peak. */
typedef void (*NativeAudioPeakCallback)(uint32_t peakBits);

/**
 * @brief Starts the media, notification, and audio peak background services.
 * @param mediaCb Callback for media session updates.
 * @param notifCb Callback for system toast notifications.
 * @param peakCb Callback for real-time audio output peak levels.
 * @return 1 if started, 0 if already running.
 */
BRIDGE_API int StartNativeWinRTServices(NativeMediaCallback mediaCb, NativeNotificationCallback notifCb,
                                        NativeAudioPeakCallback peakCb);

/**
 * @brief Stops all active native background services and joins worker threads.
 */
BRIDGE_API void StopNativeWinRTServices(void);

/** @brief Dispatches a playback command ("play", "pause", "next", "prev", etc.) to the active session. */
BRIDGE_API int SendNativeMediaCommand(const char* command);

/** @brief Toggles Wi-Fi, Bluetooth, or Airplane Mode state. */
BRIDGE_API int SetNativeRadioState(int radioKind, int targetState);

/** @brief Writes the current radio and Wi-Fi connection status as JSON into outJson. */
BRIDGE_API int GetNativeRadiosStatus(char* outJson, int maxLen);

/** @brief Locks media control onto a specific GSMTC session ID. */
BRIDGE_API int SelectNativeMediaSession(const char* sessionId);

/** @brief Brings the window of the specified media source application to the foreground. */
BRIDGE_API int FocusNativeMediaApp(const char* sourceApp);

/** @brief Brings an application window to the foreground by process name or AUMID. */
BRIDGE_API int FocusNativeApp(const char* sourceApp, const char* aumid);

/** @brief Extracts an application's icon and writes it as a PNG file to destFilePath. */
BRIDGE_API int ExtractAppIconToFile(const char* sourceApp, const char* aumid, const wchar_t* destFilePath,
                                    int targetSize);

/** @brief Clears cached media album artwork from memory. */
BRIDGE_API void ClearNativeMediaArtCache(void);

/** @brief Forces a fresh fetch of the active media session's album artwork. */
BRIDGE_API void RequeryNativeMediaArtwork(void);

/** @brief Forces an immediate broadcast of the current media state to the Go callback. */
BRIDGE_API void ForceEmitNativeMediaState(void);

/** @brief Displays a native Windows toast notification under the Vinta AUMID. */
BRIDGE_API int SendNativeToast(const wchar_t* title, const wchar_t* message);

/** @brief Writes active per-application audio sessions as JSON into outJson. */
BRIDGE_API int GetNativeAudioSessions(char* outJson, int maxLen);

/** @brief Sets the volume level (0-100) for a specific process ID's audio session. */
BRIDGE_API int SetNativeSessionVolume(uint32_t pid, int volumePercent);

/** @brief Mutes or unmutes a specific process ID's audio session. */
BRIDGE_API int SetNativeSessionMute(uint32_t pid, int mute);

/** @brief Writes available audio output endpoints as JSON into outJson. */
BRIDGE_API int GetNativeAudioEndpoints(char* outJson, int maxLen);

/** @brief Switches the system default audio output device to the given endpoint ID. */
BRIDGE_API int SetNativeDefaultAudioEndpoint(const wchar_t* endpointId);

/** @brief Sets the master system output volume (0-100). */
BRIDGE_API int SetNativeMasterAudioVolume(int percent);

/** @brief Toggles the master system output mute state. */
BRIDGE_API int ToggleNativeMasterAudioMute(void);

#ifdef __cplusplus
}
#endif

#endif // WINRT_BRIDGE_H