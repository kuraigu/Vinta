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

	mediaCallbackPtr uintptr
	notifCallbackPtr uintptr
	peakCallbackPtr  uintptr

	wailsCtx  context.Context
	serviceMu sync.Mutex
	isStarted bool
)

// StartServicesListener initializes the native C++/WinRT in-process bridge.
func StartServicesListener(ctx context.Context) {
	serviceMu.Lock()
	defer serviceMu.Unlock()

	if isStarted {
		return
	}

	wailsCtx = ctx

	dllPath := findDLLPath()
	if dllPath == "" {
		logMsg("vintaservices.dll not found; media/notification services disabled")
		return
	}

	handle, err := syscall.LoadLibrary(dllPath)
	if err != nil {
		logMsg(fmt.Sprintf("Failed to load %s: %v", dllPath, err))
		return
	}
	dllHandle = handle

	procStart, err := syscall.GetProcAddress(dllHandle, "StartNativeWinRTServices")
	if err != nil {
		logMsg("Failed to find StartNativeWinRTServices proc: " + err.Error())
		return
	}
	procStartNativeWinRTServices = procStart

	procStop, err := syscall.GetProcAddress(dllHandle, "StopNativeWinRTServices")
	if err != nil {
		logMsg("Failed to find StopNativeWinRTServices proc: " + err.Error())
		return
	}
	procStopNativeWinRTServices = procStop

	procCmd, err := syscall.GetProcAddress(dllHandle, "SendNativeMediaCommand")
	if err != nil {
		logMsg("Failed to find SendNativeMediaCommand proc: " + err.Error())
		return
	}
	procSendNativeMediaCommand = procCmd

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
	}

	isStarted = false
	logMsg("Native WinRT services stopped")
}

// SendMediaControl dispatches an action directly to the native GSMTC media controller.
func SendMediaControl(action string) {
	serviceMu.Lock()
	proc := procSendNativeMediaCommand
	serviceMu.Unlock()

	if proc == 0 || action == "" {
		return
	}

	cmdBytes := append([]byte(action), 0)
	_, _, _ = syscall.SyscallN(proc, uintptr(unsafe.Pointer(&cmdBytes[0])))
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
