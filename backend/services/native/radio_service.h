/**
 * @file radio_service.h
 * @brief Windows Runtime Radio control service for Wi-Fi and Bluetooth.
 *
 * Enumerates system radio adapters using Windows.Devices.Radios.Radio and
 * toggles their power state asynchronously.
 */

#ifndef RADIO_SERVICE_H
#define RADIO_SERVICE_H

#include "winrt_common.h"

// radioKind: 1 = WiFi, 2 = MobileBroadband, 3 = Bluetooth, 255 = All (Airplane Mode)
// targetState: 1 = On, 2 = Off
int SetRadioState(int radioKind, int targetState);

// Returns JSON string with native radio and Wi-Fi connection states
// Returns 1 on success, 0 on failure
int GetRadiosStateJson(char* outBuf, int maxBufLen);

#endif // RADIO_SERVICE_H
