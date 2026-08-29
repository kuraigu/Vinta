// Package services provides native in-process WinRT integration for media and notifications.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

var (
	dllHandle                    syscall.Handle
	procStartNativeWinRTServices uintptr
	procStopNativeWinRTServices  uintptr
	procSendNativeMediaCommand   uintptr
	procSetNativeRadioState      uintptr
	procGetNativeRadiosStatus    uintptr

	mediaCallbackPtr uintptr
	notifCallbackPtr uintptr
	peakCallbackPtr  uintptr

	wailsCtx  context.Context
	serviceMu sync.Mutex
	isStarted bool
)

func ensureDLLLoadedLocked() bool {
	if dllHandle != 0 {
		return true
	}
	dllPath := findDLLPath()
	if dllPath == "" {
		return false
	}
	handle, err := syscall.LoadLibrary(dllPath)
	if err != nil {
		return false
	}
	dllHandle = handle
	procStartNativeWinRTServices, _ = syscall.GetProcAddress(dllHandle, "StartNativeWinRTServices")
	procStopNativeWinRTServices, _ = syscall.GetProcAddress(dllHandle, "StopNativeWinRTServices")
	procSendNativeMediaCommand, _ = syscall.GetProcAddress(dllHandle, "SendNativeMediaCommand")
	procSetNativeRadioState, _ = syscall.GetProcAddress(dllHandle, "SetNativeRadioState")
	procGetNativeRadiosStatus, _ = syscall.GetProcAddress(dllHandle, "GetNativeRadiosStatus")
	return true
}

// StartServicesListener initializes the native C++/WinRT in-process bridge.
func StartServicesListener(ctx context.Context) {
	serviceMu.Lock()
	defer serviceMu.Unlock()

	if isStarted {
		return
	}

	wailsCtx = ctx

	if !ensureDLLLoadedLocked() {
		logMsg("vintaservices.dll not found; native WinRT services disabled")
		return
	}

	if procStartNativeWinRTServices == 0 {
		logMsg("Failed to find StartNativeWinRTServices proc")
		return
	}

	mediaCallbackPtr = syscall.NewCallback(onNativeMedia)
	notifCallbackPtr = syscall.NewCallback(onNativeNotification)
	peakCallbackPtr = syscall.NewCallback(onNativeAudioPeak)

	ret, _, _ := syscall.SyscallN(
		procStartNativeWinRTServices,
		mediaCallbackPtr,
		notifCallbackPtr,
		peakCallbackPtr,
	)

	if ret == 1 {
		isStarted = true
		logMsg("Native C++/WinRT in-process bridge started successfully")
	} else {
		logMsg("Failed to initialize native WinRT services")
	}
}

// StopServicesListener halts the native WinRT worker threads and unloads the bridge.
func StopServicesListener() {
	serviceMu.Lock()
	defer serviceMu.Unlock()

	if !isStarted {
		return
	}

	if procStopNativeWinRTServices != 0 {
		_, _, _ = syscall.SyscallN(procStopNativeWinRTServices)
	}

	if dllHandle != 0 {
		_ = syscall.FreeLibrary(dllHandle)
		dllHandle = 0
		procStartNativeWinRTServices = 0
		procStopNativeWinRTServices = 0
		procSendNativeMediaCommand = 0
		procSetNativeRadioState = 0
		procGetNativeRadiosStatus = 0
	}

	isStarted = false
	logMsg("Native WinRT services stopped")
}

// SendMediaControl dispatches an action directly to the native GSMTC media controller.
func SendMediaControl(action string) {
	serviceMu.Lock()
	if !ensureDLLLoadedLocked() {
		serviceMu.Unlock()
		return
	}
	proc := procSendNativeMediaCommand
	serviceMu.Unlock()

	if proc == 0 || action == "" {
		return
	}

	cmdBytes := append([]byte(action), 0)
	_, _, _ = syscall.SyscallN(proc, uintptr(unsafe.Pointer(&cmdBytes[0])))
}

// SetNativeRadioState sends radio state request to vintaservices.dll.
// radioKind: 1 = WiFi, 2 = MobileBroadband, 3 = Bluetooth, 255 = All (Airplane Mode)
// targetState: 1 = On, 2 = Off
func SetNativeRadioState(radioKind int, targetState int) bool {
	serviceMu.Lock()
	if !ensureDLLLoadedLocked() {
		serviceMu.Unlock()
		return false
	}
	proc := procSetNativeRadioState
	serviceMu.Unlock()

	if proc == 0 {
		return false
	}

	ret, _, _ := syscall.SyscallN(proc, uintptr(radioKind), uintptr(targetState))
	return ret == 1
}

// GetNativeRadiosStatus queries radio states from vintaservices.dll.
func GetNativeRadiosStatus() (string, bool) {
	serviceMu.Lock()
	if !ensureDLLLoadedLocked() {
		serviceMu.Unlock()
		return "", false
	}
	proc := procGetNativeRadiosStatus
	serviceMu.Unlock()

	if proc == 0 {
		return "", false
	}

	buf := make([]byte, 1024)
	ret, _, _ := syscall.SyscallN(proc, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if ret == 1 {
		return windows.BytePtrToString(&buf[0]), true
	}
	return "", false
}

var (
	mediaMu       sync.Mutex
	lastTrackKey  string
	lastPlayState bool
	isAppHidden   bool
	hiddenMu      sync.RWMutex
)

// SetAppHidden informs the native services whether Vinta is currently hidden to throttle high-frequency events.
func SetAppHidden(hidden bool) {
	hiddenMu.Lock()
	isAppHidden = hidden
	hiddenMu.Unlock()
}

func onNativeMedia(cStr *byte) uintptr {
	if cStr == nil || wailsCtx == nil {
		return 0
	}

	str := windows.BytePtrToString(cStr)
	if str == "" {
		return 0
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(str), &payload); err == nil {
		title, _ := payload["title"].(string)
		artist, _ := payload["artist"].(string)
		isPlaying, _ := payload["isPlaying"].(bool)
		trackKey := title + "_" + artist

		hiddenMu.RLock()
		hidden := isAppHidden
		hiddenMu.RUnlock()

		mediaMu.Lock()
		trackChanged := (trackKey != lastTrackKey)
		playStateChanged := (isPlaying != lastPlayState)
		lastPlayState = isPlaying

		if !trackChanged {
			// Do not re-send the multi-megabyte albumArt on routine 500ms position updates.
			// The frontend preserves the album art for the same track.
			payload["albumArt"] = ""
		} else {
			lastTrackKey = trackKey
		}
		mediaMu.Unlock()

		// When the island is in hidden mode, avoid emitting high-frequency timeline
		// scrub ticks into WebView2. Only notify when track changes or play state toggles.
		if hidden && !trackChanged && !playStateChanged {
			return 0
		}

		runtime.EventsEmit(wailsCtx, "media-update", payload)
	}

	return 0
}

func onNativeNotification(cStr *byte) uintptr {
	if cStr == nil || wailsCtx == nil {
		return 0
	}

	str := windows.BytePtrToString(cStr)
	if str == "" {
		return 0
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(str), &payload); err == nil {
		runtime.EventsEmit(wailsCtx, "notification", payload)
	}

	return 0
}

var (
	lastAudioPeak float32 = -1.0
	peakMu        sync.Mutex
)

func onNativeAudioPeak(peakBits uintptr) uintptr {
	if wailsCtx == nil {
		return 0
	}

	hiddenMu.RLock()
	hidden := isAppHidden
	hiddenMu.RUnlock()
	if hidden {
		return 0
	}

	peak := math.Float32frombits(uint32(peakBits))

	peakMu.Lock()
	if peak == 0 && lastAudioPeak == 0 {
		peakMu.Unlock()
		return 0
	}
	if math.Abs(float64(peak-lastAudioPeak)) < 0.012 && peak > 0 {
		peakMu.Unlock()
		return 0
	}
	lastAudioPeak = peak
	peakMu.Unlock()

	payload := map[string]interface{}{
		"type": "audio_peak",
		"peak": float64(peak),
	}
	runtime.EventsEmit(wailsCtx, "audio-peak", payload)
	return 0
}

func findDLLPath() string {
	exePath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exePath)
		candidates := []string{
			filepath.Join(dir, "vintaservices.dll"),
			filepath.Join(dir, "build", "bin", "vintaservices.dll"),
			filepath.Join(dir, "backend", "services", "vintaservices.dll"),
			filepath.Join(dir, "..", "..", "backend", "services", "vintaservices.dll"),
			filepath.Join(dir, "..", "..", "build", "bin", "vintaservices.dll"),
		}
		for _, candidate := range candidates {
			if fileExists(candidate) {
				return candidate
			}
		}
	}

	// Working directory fallbacks
	wdCandidates := []string{
		"build\\bin\\vintaservices.dll",
		"backend\\services\\vintaservices.dll",
		"vintaservices.dll",
	}
	for _, c := range wdCandidates {
		if abs, err := filepath.Abs(c); err == nil && fileExists(abs) {
			return abs
		}
	}

	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func logMsg(msg string) {
	fmt.Printf("[VintaServices] %s\n", msg)
}
