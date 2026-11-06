/**
 * @file icon_service.h
 * @brief Native Windows application icon extraction service.
 *
 * Extracts high-resolution icons from Win32 executable files and Windows Store (UWP)
 * application packages using IShellItemImageFactory and SHGetFileInfoW, saving
 * them to the local application cache as PNG images.
 */

#ifndef ICON_SERVICE_H
#define ICON_SERVICE_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// ExtractAppIconNativeToFile extracts the native icon of a Win32 executable or UWP/Store AUMID
// and saves it directly as a PNG file at destFilePath.
// Returns 1 on success, 0 on failure.
int ExtractAppIconNativeToFile(const char* sourceApp, const char* aumid, const wchar_t* destFilePath, int targetSize);

#ifdef __cplusplus
}
#endif

#endif // ICON_SERVICE_H
