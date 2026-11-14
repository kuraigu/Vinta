/**
 * @file media_service.h
 * @brief Windows System Media Transport Controls (SMTC) and WASAPI session mixer.
 *
 * Tracks active media sessions across desktop apps and browsers, manages timeline
 * extrapolation and album art caching, and provides per-application and master
 * volume controls via WASAPI.
 */

#ifndef MEDIA_SERVICE_H
#define MEDIA_SERVICE_H

#include "winrt_bridge.h"

/** @brief Starts the SMTC session manager and background polling thread. */
void StartMediaService(NativeMediaCallback cb);

/** @brief Stops the media service, unregisters WinRT tokens, and joins worker threads. */
void StopMediaService();

/** @brief Executes a playback, seek, or volume command on the active media session. */
int ExecuteMediaCommand(const char* command);

/** @brief Returns true if the current media session is actively playing. */
bool IsMediaPlayingState();

/** @brief Pins media tracking to a specific session unique ID. */
int SelectMediaSession(const char* sessionId);

/** @brief Brings the window of the specified media source application to the foreground. */
int FocusMediaApp(const char* sourceApp);

/** @brief Brings an application window to the foreground by executable name or AUMID. */
int FocusApp(const char* sourceApp, const char* aumid);

/** @brief Invalidates the current track's cached artwork and triggers a fresh thumbnail fetch. */
void RequeryMediaArtwork();

/** @brief Forces an immediate state broadcast of the active media session. */
void ForceEmitMediaState();

/** @brief Clears all cached album artwork and track snapshots from memory. */
void FlushMediaArtCache();

/** @brief Serializes all active per-application WASAPI audio sessions as JSON into outBuf. */
int GetAllAudioSessionsJson(char* outBuf, int maxLen);

/** @brief Sets the audio session volume (0-100) for the specified process ID. */
int SetAppAudioVolumeByPID(uint32_t pid, int volumePercent);

/** @brief Mutes or unmutes the audio session for the specified process ID. */
int SetAppAudioMuteByPID(uint32_t pid, bool mute);

/** @brief Serializes all active render audio endpoints as JSON into outBuf. */
int GetAudioEndpointsJson(char* outBuf, int maxLen);

/** @brief Sets the system default render audio endpoint across console, multimedia, and comms roles. */
int SetDefaultAudioEndpoint(const wchar_t* endpointId);

/** @brief Sets the master system output volume (0-100). */
int SetMasterVolumeNative(int percent);

/** @brief Toggles the master system output mute state. */
int ToggleMasterMuteNative();

#endif // MEDIA_SERVICE_H