#ifndef NOTIFICATION_SERVICE_H
#define NOTIFICATION_SERVICE_H

#include "winrt_bridge.h"

void StartNotificationService(NativeNotificationCallback cb);
void StopNotificationService();
int SendToastNotification(const wchar_t* title, const wchar_t* message);

#endif // NOTIFICATION_SERVICE_H
