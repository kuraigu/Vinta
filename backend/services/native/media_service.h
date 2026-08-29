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

#endif // MEDIA_SERVICE_H
