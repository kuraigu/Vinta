/**
 * @file audio_peak_service.h
 * @brief Windows Core Audio (WASAPI) endpoint and session volume manager.
 *
 * Interacts with IAudioSessionManager2 and IAudioEndpointVolume to enumerate
 * running audio processes, adjust per-app volume sliders, mute streams, and
 * sample real-time master output peak levels for visualizer widgets.
 */

#ifndef AUDIO_PEAK_SERVICE_H
#define AUDIO_PEAK_SERVICE_H

#include "winrt_bridge.h"

void StartAudioPeakService(NativeAudioPeakCallback cb);
void StopAudioPeakService();

#endif // AUDIO_PEAK_SERVICE_H
