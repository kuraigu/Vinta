#ifndef AUDIO_PEAK_SERVICE_H
#define AUDIO_PEAK_SERVICE_H

#include "winrt_bridge.h"

void StartAudioPeakService(NativeAudioPeakCallback cb);
void StopAudioPeakService();

#endif // AUDIO_PEAK_SERVICE_H
