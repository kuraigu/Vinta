// Package services provides native in-process WinRT integration for media and notifications.
package services

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/sys/windows"
)

var (
	dllHandle                     syscall.Handle
	procStartNativeWinRTServices  uintptr
	procStopNativeWinRTServices   uintptr
	procSendNativeMediaCommand    uintptr
	procSetNativeRadioState       uintptr
	procGetNativeRadiosStatus     uintptr
	procSelectNativeMediaSession  uintptr
	procFocusNativeMediaApp       uintptr
	procFocusNativeApp            uintptr
	procExtractAppIconToFile      uintptr
	procClearNativeMediaArtCache  uintptr
	procRequeryNativeMediaArtwork uintptr
	procForceEmitNativeMediaState uintptr

	mediaCallbackPtr uintptr
	notifCallbackPtr uintptr
	peakCallbackPtr  uintptr

	serviceMu sync.Mutex
	isStarted bool

	isFrontendReady atomic.Bool

	cachedMediaState   map[string]interface{}
	cachedMediaStateMu sync.Mutex

	pendingNotifs   []map[string]interface{}
	pendingNotifsMu sync.Mutex
)

// SetFrontendReady updates the readiness status of the frontend WebView runtime.
// When transitioning to ready, buffered initial media updates and pending notifications
// received during window initialization are dispatched immediately.
func SetFrontendReady(ready bool) {
	wasReady := isFrontendReady.Swap(ready)
	if !ready || wasReady {
		return
	}

	pendingNotifsMu.Lock()
	bufferedNotifs := pendingNotifs
	pendingNotifs = nil
	pendingNotifsMu.Unlock()

	if app := application.Get(); app != nil {
		for _, item := range bufferedNotifs {
			msgType, _ := item["type"].(string)
			if msgType == "" {
				msgType = "notification"
			}
			app.Event.Emit(msgType, item)
		}
	}

	cachedMediaStateMu.Lock()
	media := cachedMediaState
	cachedMediaStateMu.Unlock()

	if media != nil {
		if app := application.Get(); app != nil {
			app.Event.Emit("media-update", media)
		}
	}
}

// IsFrontendReady reports whether the frontend WebView is active and accepting events.
func IsFrontendReady() bool {
	return isFrontendReady.Load()
}

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
	procSelectNativeMediaSession, _ = syscall.GetProcAddress(dllHandle, "SelectNativeMediaSession")
	procFocusNativeMediaApp, _ = syscall.GetProcAddress(dllHandle, "FocusNativeMediaApp")
	procFocusNativeApp, _ = syscall.GetProcAddress(dllHandle, "FocusNativeApp")
	procExtractAppIconToFile, _ = syscall.GetProcAddress(dllHandle, "ExtractAppIconToFile")
	procClearNativeMediaArtCache, _ = syscall.GetProcAddress(dllHandle, "ClearNativeMediaArtCache")
	procRequeryNativeMediaArtwork, _ = syscall.GetProcAddress(dllHandle, "RequeryNativeMediaArtwork")
	procForceEmitNativeMediaState, _ = syscall.GetProcAddress(dllHandle, "ForceEmitNativeMediaState")
	return true
}

// StartServicesListener initializes the native C++/WinRT in-process bridge.
func StartServicesListener() {
	serviceMu.Lock()
	defer serviceMu.Unlock()

	if isStarted {
		return
	}

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
	SetFrontendReady(false)

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
		procFocusNativeMediaApp = 0
		procFocusNativeApp = 0
		procExtractAppIconToFile = 0
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

// SelectMediaSession asks vintaservices.dll to switch active playback session to sessionID.
func SelectMediaSession(sessionID string) bool {
	serviceMu.Lock()
	if !ensureDLLLoadedLocked() {
		serviceMu.Unlock()
		return false
	}
	proc := procSelectNativeMediaSession
	serviceMu.Unlock()

	if proc == 0 {
		return false
	}

	idBytes := append([]byte(sessionID), 0)
	ret, _, _ := syscall.SyscallN(proc, uintptr(unsafe.Pointer(&idBytes[0])))
	return ret == 1
}

// FocusMediaApp asks vintaservices.dll to find and focus the source application's window.
func FocusMediaApp(sourceApp string) bool {
	serviceMu.Lock()
	if !ensureDLLLoadedLocked() {
		serviceMu.Unlock()
		return false
	}
	proc := procFocusNativeMediaApp
	serviceMu.Unlock()

	if proc == 0 || sourceApp == "" {
		return false
	}

	appBytes := append([]byte(sourceApp), 0)
	ret, _, _ := syscall.SyscallN(proc, uintptr(unsafe.Pointer(&appBytes[0])))
	return ret == 1
}

// FocusNotificationApp asks vintaservices.dll to focus the window matching either
// the sourceApp process name or the App User Model ID (aumid).
func FocusNotificationApp(sourceApp, aumid string) bool {
	serviceMu.Lock()
	if !ensureDLLLoadedLocked() {
		serviceMu.Unlock()
		return false
	}
	proc := procFocusNativeApp
	serviceMu.Unlock()

	if proc == 0 {
		return false
	}

	appBytes := append([]byte(sourceApp), 0)
	aumidBytes := append([]byte(aumid), 0)
	ret, _, _ := syscall.SyscallN(
		proc,
		uintptr(unsafe.Pointer(&appBytes[0])),
		uintptr(unsafe.Pointer(&aumidBytes[0])),
	)
	return ret == 1
}

// ExtractAppIconToFile invokes vintaservices.dll to extract the native icon for sourceApp/aumid to destFilePath.
func ExtractAppIconToFile(sourceApp, aumid, destFilePath string, size int) bool {
	serviceMu.Lock()
	if !ensureDLLLoadedLocked() {
		serviceMu.Unlock()
		return false
	}
	proc := procExtractAppIconToFile
	serviceMu.Unlock()

	if proc == 0 || destFilePath == "" {
		return false
	}

	appBytes := append([]byte(sourceApp), 0)
	aumidBytes := append([]byte(aumid), 0)
	destPathW, err := syscall.UTF16PtrFromString(destFilePath)
	if err != nil {
		return false
	}

	ret, _, _ := syscall.SyscallN(
		proc,
		uintptr(unsafe.Pointer(&appBytes[0])),
		uintptr(unsafe.Pointer(&aumidBytes[0])),
		uintptr(unsafe.Pointer(destPathW)),
		uintptr(size),
	)
	return ret == 1
}

var (
	mediaMu             sync.Mutex
	lastTrackKey        string
	lastPlayState       bool
	lastArtSentTrackKey string
	lastSentArt         string
	isAppHidden         bool
	hiddenMu            sync.RWMutex
)

// SetAppHidden informs the native services whether Vinta is currently hidden to throttle high-frequency events.
func SetAppHidden(hidden bool) {
	hiddenMu.Lock()
	isAppHidden = hidden
	hiddenMu.Unlock()
}

// ResetMediaArtCacheState flushes in-memory media artwork deduplication state.
func ResetMediaArtCacheState() {
	mediaMu.Lock()
	defer mediaMu.Unlock()
	lastArtSentTrackKey = ""
	lastSentArt = ""
	lastTrackKey = ""
}

// ClearNativeMediaArtCache calls into the native WinRT bridge to wipe cached media artwork.
func ClearNativeMediaArtCache() {
	serviceMu.Lock()
	if dllHandle != 0 && procClearNativeMediaArtCache != 0 {
		_, _, _ = syscall.SyscallN(procClearNativeMediaArtCache)
	}
	serviceMu.Unlock()
	ResetMediaArtCacheState()
}

// RequeryMediaArt instructs the native WinRT bridge to re-extract the thumbnail
// for the current session and forces an immediate IPC push.
func RequeryMediaArt() {
	ResetMediaArtCacheState()
	serviceMu.Lock()
	if dllHandle != 0 && procRequeryNativeMediaArtwork != 0 {
		_, _, _ = syscall.SyscallN(procRequeryNativeMediaArtwork)
	}
	serviceMu.Unlock()
}

// RequestMediaUpdate resets the cached delivered artwork state and requests
// the native bridge to immediately emit the current media snapshot.
// This is called when the frontend mounts or reconnects to avoid stale blank state.
func RequestMediaUpdate() {
	SetFrontendReady(true)
	ResetMediaArtCacheState()
	serviceMu.Lock()
	if dllHandle != 0 && procForceEmitNativeMediaState != 0 {
		_, _, _ = syscall.SyscallN(procForceEmitNativeMediaState)
	}
	serviceMu.Unlock()
}

func onNativeMedia(cStr *byte) uintptr {
	if cStr == nil {
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
		sourceApp, _ := payload["sourceApp"].(string)
		activeSessionID, _ := payload["activeSessionId"].(string)
		isPlaying, _ := payload["isPlaying"].(bool)

		trackKey := sourceApp + "::" + activeSessionID + "::" + title + "::" + artist

		hiddenMu.RLock()
		hidden := isAppHidden
		hiddenMu.RUnlock()

		mediaMu.Lock()
		trackChanged := (trackKey != lastTrackKey)
		playStateChanged := (isPlaying != lastPlayState)
		lastPlayState = isPlaying

		art, _ := payload["albumArt"].(string)
		if art != "" {
			isNewOrDifferentArt := (lastArtSentTrackKey != trackKey) || (art != lastSentArt)
			if isNewOrDifferentArt {
				lastArtSentTrackKey = trackKey
				lastSentArt = art
			} else if !trackChanged {
				// Once delivered, suppress multi-megabyte base64 on subsequent routine position updates
				payload["albumArt"] = ""
			}
		} else if !trackChanged && lastArtSentTrackKey == trackKey {
			payload["albumArt"] = ""
		}

		if trackChanged {
			lastTrackKey = trackKey
			if art == "" {
				lastArtSentTrackKey = ""
				lastSentArt = ""
			}
		}
		mediaMu.Unlock()

		// When the island is in hidden mode, avoid emitting high-frequency timeline
		// scrub ticks into WebView2. Only notify when track changes or play state toggles.
		if hidden && !trackChanged && !playStateChanged {
			return 0
		}

		cachedMediaStateMu.Lock()
		cachedMediaState = payload
		cachedMediaStateMu.Unlock()

		if !isFrontendReady.Load() {
			return 0
		}

		if app := application.Get(); app != nil {
			app.Event.Emit("media-update", payload)
		}
	}

	return 0
}

func onNativeNotification(cStr *byte) uintptr {
	if cStr == nil {
		return 0
	}

	str := windows.BytePtrToString(cStr)
	if str == "" {
		return 0
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(str), &payload); err == nil {
		if !isFrontendReady.Load() {
			pendingNotifsMu.Lock()
			if len(pendingNotifs) < 50 {
				pendingNotifs = append(pendingNotifs, payload)
			}
			pendingNotifsMu.Unlock()
			return 0
		}

		if app := application.Get(); app != nil {
			msgType, _ := payload["type"].(string)
			if msgType == "" {
				msgType = "notification"
			}
			app.Event.Emit(msgType, payload)
		}
	}

	return 0
}

var (
	lastAudioPeak float32 = -1.0
	peakMu        sync.Mutex
)

func onNativeAudioPeak(peakBits uintptr) uintptr {
	if !isFrontendReady.Load() {
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
	if app := application.Get(); app != nil {
		app.Event.Emit("audio-peak", payload)
	}
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
