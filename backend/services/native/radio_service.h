/**
 * @file radio_service.h
 * @brief Windows Runtime radio control and status query interface.
 *
 * Queries and toggles Wi-Fi, Bluetooth, and system-wide Airplane Mode using
 * WinRT Windows.Devices.Radios and the native COM IRadioManager interface.
 */

#ifndef RADIO_SERVICE_H
#define RADIO_SERVICE_H

#include "winrt_common.h"

/**
 * @brief Toggles the power state of a system radio adapter or Airplane Mode.
 * @param radioKind 1 = Wi-Fi, 2 = MobileBroadband, 3 = Bluetooth, 255 = Airplane Mode.
 * @param targetState 1 = On, 2 = Off (for Airplane Mode: 2 enables Airplane Mode, 1 disables it).
 * @return 1 on success, 0 on failure.
 */
int SetRadioState(int radioKind, int targetState);

/**
 * @brief Serializes the current Wi-Fi, Bluetooth, and Airplane Mode status into a JSON buffer.
 * @param outBuf Destination character buffer allocated by the caller.
 * @param maxBufLen Capacity of outBuf in bytes.
 * @return 1 on success, 0 on failure or insufficient buffer space.
 */
int GetRadiosStateJson(char* outBuf, int maxBufLen);

#endif // RADIO_SERVICE_H