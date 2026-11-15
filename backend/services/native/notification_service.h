/**
 * @file notification_service.h
 * @brief Windows toast notification listener and sender service.
 *
 * Monitors the Windows notification database and WinRT UserNotificationListener
 * to capture incoming system toasts, extract media attachments and app icons,
 * and forward structured JSON payloads to the Go backend.
 */

#ifndef NOTIFICATION_SERVICE_H
#define NOTIFICATION_SERVICE_H

#include "winrt_bridge.h"

/**
 * @brief Starts the background toast notification listener thread.
 * @param cb Callback invoked with JSON notification payloads.
 */
void StartNotificationService(NativeNotificationCallback cb);

/**
 * @brief Stops the notification listener thread and releases OS handles.
 */
void StopNotificationService();

/**
 * @brief Sends a native Windows toast notification under the Vinta AUMID.
 * @param title Notification header text.
 * @param message Notification body text.
 * @return 1 on success, 0 on failure.
 */
int SendToastNotification(const wchar_t* title, const wchar_t* message);

#endif // NOTIFICATION_SERVICE_H