/**
 * @file audio_peak_service.h
 * @brief Master audio output peak meter service using WASAPI.
 *
 * Polls the active Windows render endpoint via IAudioMeterInformation on a
 * background thread and sends normalized peak values to the Go backend for
 * audio visualizer widgets.
 */

#ifndef AUDIO_PEAK_SERVICE_H
#define AUDIO_PEAK_SERVICE_H

#include "winrt_bridge.h"

/**
 * @brief Starts the background audio peak polling thread.
 * @param cb Callback invoked with IEEE-754 float bits representing the current peak level.
 */
void StartAudioPeakService(NativeAudioPeakCallback cb);

/**
 * @brief Stops the audio peak polling thread and releases COM audio interfaces.
 */
void StopAudioPeakService();

#endif // AUDIO_PEAK_SERVICE_H