// Package backend provides core application services and state management for Vinta.
package backend

import (
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

// GetCalendarData reads calendar.json from %APPDATA%/FreeMatrix Studios/Vinta.
func (a *App) GetCalendarData() (string, error) {
	filePath := filepath.Join(GetAppDataDir(), "calendar.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "{}", nil
	}
	return string(data), nil
}

// SaveCalendarData writes calendar.json to %APPDATA%/FreeMatrix Studios/Vinta.
func (a *App) SaveCalendarData(calendarJSON string) error {
	dir := GetAppDataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	filePath := filepath.Join(dir, "calendar.json")
	return os.WriteFile(filePath, []byte(calendarJSON), 0644)
}

// GetTimerData reads timer.json from %APPDATA%/FreeMatrix Studios/Vinta.
func (a *App) GetTimerData() (string, error) {
	filePath := filepath.Join(GetAppDataDir(), "timer.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "{}", nil
	}
	return string(data), nil
}

// SaveTimerData writes timer.json to %APPDATA%/FreeMatrix Studios/Vinta.
func (a *App) SaveTimerData(timerJSON string) error {
	dir := GetAppDataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	filePath := filepath.Join(dir, "timer.json")
	return os.WriteFile(filePath, []byte(timerJSON), 0644)
}

var dpapiNotificationHeader = []byte("VNDP")

// encryptDPAPI encrypts plaintext bytes using the Windows Data Protection API.
func encryptDPAPI(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, nil
	}
	var inBlob windows.DataBlob
	inBlob.Size = uint32(len(plaintext))
	inBlob.Data = &plaintext[0]

	var outBlob windows.DataBlob
	err := windows.CryptProtectData(&inBlob, nil, nil, 0, nil, 0, &outBlob)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = windows.LocalFree(windows.Handle(unsafe.Pointer(outBlob.Data)))
	}()

	ciphertext := unsafe.Slice(outBlob.Data, outBlob.Size)
	result := make([]byte, len(ciphertext))
	copy(result, ciphertext)
	return result, nil
}

// decryptDPAPI decrypts ciphertext bytes using the Windows Data Protection API.
func decryptDPAPI(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, nil
	}
	var inBlob windows.DataBlob
	inBlob.Size = uint32(len(ciphertext))
	inBlob.Data = &ciphertext[0]

	var outBlob windows.DataBlob
	err := windows.CryptUnprotectData(&inBlob, nil, nil, 0, nil, 0, &outBlob)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = windows.LocalFree(windows.Handle(unsafe.Pointer(outBlob.Data)))
	}()

	plaintext := unsafe.Slice(outBlob.Data, outBlob.Size)
	result := make([]byte, len(plaintext))
	copy(result, plaintext)
	return result, nil
}

// GetNotifications reads notifications.json from %APPDATA%/FreeMatrix Studios/Vinta, decrypting DPAPI ciphertext if encrypted.
func (a *App) GetNotifications() (string, error) {
	filePath := filepath.Join(GetAppDataDir(), "notifications.json")
	data, err := os.ReadFile(filePath)
	if err != nil || len(data) == 0 {
		return "[]", nil
	}

	if len(data) >= len(dpapiNotificationHeader) && string(data[:len(dpapiNotificationHeader)]) == string(dpapiNotificationHeader) {
		decrypted, err := decryptDPAPI(data[len(dpapiNotificationHeader):])
		if err != nil {
			return "[]", nil
		}
		return string(decrypted), nil
	}

	return string(data), nil
}

// SaveNotifications encrypts and writes notifications.json to %APPDATA%/FreeMatrix Studios/Vinta using Windows DPAPI.
func (a *App) SaveNotifications(notificationsJSON string) error {
	dir := GetAppDataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	encrypted, err := encryptDPAPI([]byte(notificationsJSON))
	if err != nil {
		return err
	}

	payload := make([]byte, 0, len(dpapiNotificationHeader)+len(encrypted))
	payload = append(payload, dpapiNotificationHeader...)
	payload = append(payload, encrypted...)

	filePath := filepath.Join(dir, "notifications.json")
	return os.WriteFile(filePath, payload, 0600)
}
