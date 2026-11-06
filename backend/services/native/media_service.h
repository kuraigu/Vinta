/**
 * @file media_service.h
 * @brief Windows System Media Transport Controls (SMTC) integration.
 *
 * Utilizes Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager
 * to track active playback sessions across Spotify, web browsers, and media players.
 * Captures timeline progress, metadata changes, and extracts album artwork.
 */

#ifndef MEDIA_SERVICE_H
#define MEDIA_SERVICE_H

#include "winrt_bridge.h"

void StartMediaService(NativeMediaCallback cb);
void StopMediaService();
int ExecuteMediaCommand(const char* command);
bool IsMediaPlayingState();
int SelectMediaSession(const char* sessionId);
int FocusMediaApp(const char* sourceApp);
int FocusApp(const char* sourceApp, const char* aumid);
/**
 * Re-queries the media properties and thumbnail stream for the current session.
 */
void RequeryMediaArtwork();

/**
 * Forces an immediate emission of the active media state to the IPC callback.
 */
void ForceEmitMediaState();

void FlushMediaArtCache();

int GetAllAudioSessionsJson(char* outBuf, int maxLen);
int SetAppAudioVolumeByPID(uint32_t pid, int volumePercent);
int SetAppAudioMuteByPID(uint32_t pid, bool mute);
int GetAudioEndpointsJson(char* outBuf, int maxLen);
int SetDefaultAudioEndpoint(const wchar_t* endpointId);
int SetMasterVolumeNative(int percent);
int ToggleMasterMuteNative();

#endif // MEDIA_SERVICE_H
