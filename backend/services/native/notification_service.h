#ifndef NOTIFICATION_SERVICE_H
#define NOTIFICATION_SERVICE_H

#include "winrt_bridge.h"

void StartNotificationService(NativeNotificationCallback cb);
void StopNotificationService();

#endif // NOTIFICATION_SERVICE_H
