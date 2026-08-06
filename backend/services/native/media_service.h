#ifndef MEDIA_SERVICE_H
#define MEDIA_SERVICE_H

#include "winrt_bridge.h"

void StartMediaService(NativeMediaCallback cb);
void StopMediaService();
int ExecuteMediaCommand(const char* command);
bool IsMediaPlayingState();

#endif // MEDIA_SERVICE_H
