// Package services provides native system integration hooks for Vinta.
package services

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	modUser32   = syscall.NewLazyDLL("user32.dll")
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")
	modShell32  = syscall.NewLazyDLL("shell32.dll")

	procAddClipboardFormatListener    = modUser32.NewProc("AddClipboardFormatListener")
	procRemoveClipboardFormatListener = modUser32.NewProc("RemoveClipboardFormatListener")
	procOpenClipboard                 = modUser32.NewProc("OpenClipboard")
	procCloseClipboard                = modUser32.NewProc("CloseClipboard")
	procGetClipboardData              = modUser32.NewProc("GetClipboardData")
	procIsClipboardFormatAvailable    = modUser32.NewProc("IsClipboardFormatAvailable")
	procGetClipboardSequenceNumber    = modUser32.NewProc("GetClipboardSequenceNumber")
	procRegisterClipboardFormatW      = modUser32.NewProc("RegisterClipboardFormatW")
	procGetForegroundWindow           = modUser32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessID      = modUser32.NewProc("GetWindowThreadProcessId")
	procGetWindowTextW                = modUser32.NewProc("GetWindowTextW")
	procCreateWindowExW               = modUser32.NewProc("CreateWindowExW")
	procDestroyWindow                 = modUser32.NewProc("DestroyWindow")
	procDefWindowProcW                = modUser32.NewProc("DefWindowProcW")
	procRegisterClassExW              = modUser32.NewProc("RegisterClassExW")
	procUnregisterClassW              = modUser32.NewProc("UnregisterClassW")
	procGetMessageW                   = modUser32.NewProc("GetMessageW")
	procTranslateMessage              = modUser32.NewProc("TranslateMessage")
	procDispatchMessageW              = modUser32.NewProc("DispatchMessageW")
	procPostQuitMessage               = modUser32.NewProc("PostQuitMessage")
	procSetWindowsHookExW             = modUser32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx           = modUser32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx                = modUser32.NewProc("CallNextHookEx")
	procGetKeyState                   = modUser32.NewProc("GetKeyState")

	procGlobalLock                 = modKernel32.NewProc("GlobalLock")
	procGlobalUnlock               = modKernel32.NewProc("GlobalUnlock")
	procOpenProcess                = modKernel32.NewProc("OpenProcess")
	procQueryFullProcessImageNameW = modKernel32.NewProc("QueryFullProcessImageNameW")
	procGetModuleHandleW           = modKernel32.NewProc("GetModuleHandleW")

	procDragQueryFileW = modShell32.NewProc("DragQueryFileW")
)

const (
	cfText        = 1
	cfBitmap      = 2
	cfDib         = 8
	cfUnicodeText = 13
	cfHDrop       = 15
	cfDibV5       = 17

	wmClipboardUpdate = 0x031D
	wmKeyDown         = 0x0100
	wmSysKeyDown      = 0x0104
	whKeyboardLl      = 13

	vkShift   = 0x10
	vkControl = 0x11
	vkInsert  = 0x2D

	processQueryLimitedInformation = 0x1000
)

// ClipboardEventPayload represents the structured notification data sent to frontend.
type ClipboardEventPayload struct {
	Action         string `json:"action"` // "copy" or "paste"
	AppName        string `json:"appName"`
	Aumid          string `json:"aumid"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	ContentType    string `json:"contentType"` // "text", "files", "image", "data"
	ContentSnippet string `json:"contentSnippet"`
	CanOpen        bool   `json:"canOpen"`
	Timestamp      int64  `json:"timestamp"`
}

type kbdLLHookStruct struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

var (
	clipMutex          sync.Mutex
	clipStarted        bool
	copyEnabledState   atomic.Bool
	pasteEnabledState  atomic.Bool
	lastSequenceNumber uint32
	lastEventTime      time.Time
	lastSnippet        string
	clipWindowHwnd     uintptr
	keyboardHook       uintptr
	pasteTriggerChan   = make(chan struct{}, 10)
	clipContext        context.Context
)

func init() {
	// Copy notifications enabled by default; paste notifications disabled by default
	copyEnabledState.Store(true)
	pasteEnabledState.Store(false)
}

// SetClipboardMonitoring toggles active copy and paste event processing.
func SetClipboardMonitoring(copyEnabled bool, pasteEnabled bool) {
	copyEnabledState.Store(copyEnabled)
	pasteEnabledState.Store(pasteEnabled)

	clipMutex.Lock()
	defer clipMutex.Unlock()

	// Manage low-level keyboard hook dynamically based on pasteEnabled setting
	if pasteEnabled && keyboardHook == 0 && clipStarted {
		installKeyboardHookLocked()
	} else if !pasteEnabled && keyboardHook != 0 {
		uninstallKeyboardHookLocked()
	}
}

// StartClipboardListener initializes background clipboard monitoring and message handling.
func StartClipboardListener(ctx context.Context) {
	clipMutex.Lock()
	if clipStarted {
		clipMutex.Unlock()
		return
	}
	clipStarted = true
	clipContext = ctx
	clipMutex.Unlock()

	// Launch worker for processing paste trigger events asynchronously
	go pasteWorkerLoop()

	// Run message loop on a dedicated, locked OS thread for Win32 notifications
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		hInstance, _, _ := procGetModuleHandleW.Call(0)
		className, _ := syscall.UTF16PtrFromString("VintaClipboardListenerClass")
		windowName, _ := syscall.UTF16PtrFromString("VintaClipboardListenerWindow")

		wndProcCallback := syscall.NewCallback(clipboardWndProc)

		type wndClassExW struct {
			CbSize        uint32
			Style         uint32
			LpfnWndProc   uintptr
			CbClsExtra    int32
			CbWndExtra    int32
			HInstance     uintptr
			HIcon         uintptr
			HCursor       uintptr
			HbrBackground uintptr
			LpszMenuName  *uint16
			LpszClassName *uint16
			HIconSm       uintptr
		}

		var wc wndClassExW
		wc.CbSize = uint32(unsafe.Sizeof(wc))
		wc.LpfnWndProc = wndProcCallback
		wc.HInstance = hInstance
		wc.LpszClassName = className

		_, _, _ = procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

		// HWND_MESSAGE = -3 (0xFFFFFFFFFFFFFFFD or 0xFFFFFFFD)
		hwndMessage := uintptr(^uintptr(2))
		hwnd, _, _ := procCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(className)),
			uintptr(unsafe.Pointer(windowName)),
			0,
			0, 0, 0, 0,
			hwndMessage,
			0,
			hInstance,
			0,
		)

		if hwnd == 0 {
			return
		}

		clipMutex.Lock()
		clipWindowHwnd = hwnd
		if pasteEnabledState.Load() {
			installKeyboardHookLocked()
		}
		clipMutex.Unlock()

		_, _, _ = procAddClipboardFormatListener.Call(hwnd)

		type winMsg struct {
			Hwnd    uintptr
			Message uint32
			WParam  uintptr
			LParam  uintptr
			Time    uint32
			Pt      [2]int32
		}

		var msg winMsg
		for {
			res, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if res == 0 || res == ^uintptr(0) {
				break
			}
			_, _, _ = procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			_, _, _ = procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
		}

		_, _, _ = procRemoveClipboardFormatListener.Call(hwnd)
		_, _, _ = procDestroyWindow.Call(hwnd)
		_, _, _ = procUnregisterClassW.Call(uintptr(unsafe.Pointer(className)), hInstance)

		clipMutex.Lock()
		clipWindowHwnd = 0
		uninstallKeyboardHookLocked()
		clipStarted = false
		clipMutex.Unlock()
	}()
}

// StopClipboardListener stops the background clipboard listener and unhooks native resources.
func StopClipboardListener() {
	clipMutex.Lock()
	defer clipMutex.Unlock()

	if !clipStarted {
		return
	}

	uninstallKeyboardHookLocked()

	if clipWindowHwnd != 0 {
		_, _, _ = procPostQuitMessage.Call(0)
	}
}

func installKeyboardHookLocked() {
	if keyboardHook != 0 {
		return
	}
	hookCb := syscall.NewCallback(lowLevelKeyboardProc)
	hHook, _, _ := procSetWindowsHookExW.Call(
		uintptr(whKeyboardLl),
		hookCb,
		0,
		0,
	)
	keyboardHook = hHook
}

func uninstallKeyboardHookLocked() {
	if keyboardHook != 0 {
		_, _, _ = procUnhookWindowsHookEx.Call(keyboardHook)
		keyboardHook = 0
	}
}

func clipboardWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	if msg == wmClipboardUpdate {
		if copyEnabledState.Load() {
			handleClipboardUpdate()
		}
		return 0
	}
	res, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return res
}

func lowLevelKeyboardProc(nCode int32, wParam uintptr, kbd *kbdLLHookStruct) uintptr {
	if nCode >= 0 && (wParam == uintptr(wmKeyDown) || wParam == uintptr(wmSysKeyDown)) {
		if pasteEnabledState.Load() {
			if kbd != nil {
				// Check for Ctrl+V
				ctrlState, _, _ := procGetKeyState.Call(uintptr(vkControl))
				isCtrlDown := (ctrlState & 0x8000) != 0

				// Check for Shift+Insert
				shiftState, _, _ := procGetKeyState.Call(uintptr(vkShift))
				isShiftDown := (shiftState & 0x8000) != 0

				if (isCtrlDown && (kbd.VkCode == 'V' || kbd.VkCode == 'v')) ||
					(isShiftDown && kbd.VkCode == uint32(vkInsert)) {
					select {
					case pasteTriggerChan <- struct{}{}:
					default:
					}
				}
			}
		}
	}
	res, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, uintptr(unsafe.Pointer(kbd)))
	return res
}

func pasteWorkerLoop() {
	for range pasteTriggerChan {
		if !pasteEnabledState.Load() {
			continue
		}
		// Small delay to allow target application to process the paste operation
		time.Sleep(75 * time.Millisecond)
		dispatchPasteNotification()
	}
}

func handleClipboardUpdate() {
	seq, _, _ := procGetClipboardSequenceNumber.Call()
	currSeq := uint32(seq)

	clipMutex.Lock()
	if currSeq == lastSequenceNumber && time.Since(lastEventTime) < 500*time.Millisecond {
		clipMutex.Unlock()
		return
	}
	lastSequenceNumber = currSeq
	lastEventTime = time.Now()
	clipMutex.Unlock()

	// Inspect foreground application that performed the copy
	fgHwnd, _, _ := procGetForegroundWindow.Call()
	appName, processName := getProcessInfo(fgHwnd)

	payload, ok := readClipboardContent("copy", appName, processName)
	if !ok {
		return
	}

	clipMutex.Lock()
	if payload.ContentSnippet != "" && payload.ContentSnippet == lastSnippet && time.Since(lastEventTime) < 1*time.Second {
		clipMutex.Unlock()
		return
	}
	lastSnippet = payload.ContentSnippet
	ctx := clipContext
	clipMutex.Unlock()

	if ctx != nil {
		wailsRuntime.EventsEmit(ctx, "clipboard:event", payload)
	}
}

func dispatchPasteNotification() {
	fgHwnd, _, _ := procGetForegroundWindow.Call()
	appName, processName := getProcessInfo(fgHwnd)

	payload, ok := readClipboardContent("paste", appName, processName)
	if !ok {
		return
	}

	clipMutex.Lock()
	ctx := clipContext
	clipMutex.Unlock()

	if ctx != nil {
		wailsRuntime.EventsEmit(ctx, "clipboard:event", payload)
	}
}

func readClipboardContent(action, appName, processName string) (ClipboardEventPayload, bool) {
	// Retry opening the clipboard with short backoff as apps copying large items may briefly lock it
	var openSuccess bool
	for i := 0; i < 5; i++ {
		ret, _, _ := procOpenClipboard.Call(0)
		if ret != 0 {
			openSuccess = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if !openSuccess {
		return ClipboardEventPayload{}, false
	}
	defer procCloseClipboard.Call()

	// Privacy guard: Ignore passwords from password managers (1Password, Bitwarden, KeePass)
	formatIgnoreName, _ := syscall.UTF16PtrFromString("Clipboard Viewer Ignore")
	cfIgnore, _, _ := procRegisterClipboardFormatW.Call(uintptr(unsafe.Pointer(formatIgnoreName)))
	if cfIgnore != 0 {
		isIgnoreAvail, _, _ := procIsClipboardFormatAvailable.Call(cfIgnore)
		if isIgnoreAvail != 0 {
			return ClipboardEventPayload{}, false
		}
	}

	formatExcludeName, _ := syscall.UTF16PtrFromString("ExcludeClipboardContentFromMonitorProcessing")
	cfExclude, _, _ := procRegisterClipboardFormatW.Call(uintptr(unsafe.Pointer(formatExcludeName)))
	if cfExclude != 0 {
		isExcludeAvail, _, _ := procIsClipboardFormatAvailable.Call(cfExclude)
		if isExcludeAvail != 0 {
			return ClipboardEventPayload{}, false
		}
	}

	nowTs := time.Now().UnixMilli()
	displayTitle := "Copied to Clipboard"
	if action == "paste" {
		displayTitle = "Pasted from Clipboard"
	}

	effectiveAppName := "Clipboard"
	processName = "Clipboard"

	// 1. Check for Unicode Text
	availText, _, _ := procIsClipboardFormatAvailable.Call(uintptr(cfUnicodeText))
	if availText != 0 {
		hData, _, _ := procGetClipboardData.Call(uintptr(cfUnicodeText))
		if hData != 0 {
			rawPtr, _, _ := syscall.SyscallN(procGlobalLock.Addr(), hData)
			if rawPtr != 0 {
				var uPtr unsafe.Pointer
				*(*uintptr)(unsafe.Pointer(&uPtr)) = rawPtr
				str := utf16PtrToString((*uint16)(uPtr))
				_, _, _ = procGlobalUnlock.Call(hData)

				trimmed := strings.TrimSpace(str)
				if trimmed != "" {
					preview := formatSnippet(trimmed, 100)
					return ClipboardEventPayload{
						Action:         action,
						AppName:        effectiveAppName,
						Aumid:          processName,
						Title:          displayTitle,
						Message:        preview,
						ContentType:    "text",
						ContentSnippet: preview,
						CanOpen:        false,
						Timestamp:      nowTs,
					}, true
				}
			}
		}
	}

	// 2. Check for Copied Files/Folders (CF_HDROP)
	availHdrop, _, _ := procIsClipboardFormatAvailable.Call(uintptr(cfHDrop))
	if availHdrop != 0 {
		hData, _, _ := procGetClipboardData.Call(uintptr(cfHDrop))
		if hData != 0 {
			rawPtr, _, _ := syscall.SyscallN(procGlobalLock.Addr(), hData)
			if rawPtr != 0 {
				hDrop := rawPtr
				defer procGlobalUnlock.Call(hData)

				retCount, _, _ := procDragQueryFileW.Call(hDrop, 0xFFFFFFFF, 0, 0)
				fileCount := int(retCount)
				if fileCount <= 0 {
					return ClipboardEventPayload{}, false
				}

				var firstFileBuf [260]uint16
				_, _, _ = procDragQueryFileW.Call(hDrop, 0, uintptr(unsafe.Pointer(&firstFileBuf[0])), uintptr(len(firstFileBuf)))
				firstFile := syscall.UTF16ToString(firstFileBuf[:])
				fileName := filepath.Base(firstFile)

				var msg string
				if fileCount == 1 {
					msg = fileName
				} else {
					msg = fileName + " and " + itoa(fileCount-1) + " other file(s)"
				}

				itemTitle := "Copied File"
				if fileCount > 1 {
					itemTitle = "Copied Files (" + itoa(fileCount) + ")"
				}
				if action == "paste" {
					itemTitle = "Pasted File"
					if fileCount > 1 {
						itemTitle = "Pasted Files (" + itoa(fileCount) + ")"
					}
				}

				return ClipboardEventPayload{
					Action:         action,
					AppName:        effectiveAppName,
					Aumid:          processName,
					Title:          itemTitle,
					Message:        formatSnippet(msg, 100),
					ContentType:    "files",
					ContentSnippet: fileName,
					CanOpen:        false,
					Timestamp:      nowTs,
				}, true
			}
		}
	}

	// 3. Check for Copied Images (CF_DIB / CF_DIBV5 / CF_BITMAP)
	availDib, _, _ := procIsClipboardFormatAvailable.Call(uintptr(cfDib))
	availDibV5, _, _ := procIsClipboardFormatAvailable.Call(uintptr(cfDibV5))
	availBmp, _, _ := procIsClipboardFormatAvailable.Call(uintptr(cfBitmap))
	if availDib != 0 || availDibV5 != 0 || availBmp != 0 {
		imgTitle := "Copied Image"
		if action == "paste" {
			imgTitle = "Pasted Image"
		}
		return ClipboardEventPayload{
			Action:         action,
			AppName:        effectiveAppName,
			Aumid:          processName,
			Title:          imgTitle,
			Message:        "Image on clipboard",
			ContentType:    "image",
			ContentSnippet: "[Image]",
			CanOpen:        false,
			Timestamp:      nowTs,
		}, true
	}

	return ClipboardEventPayload{}, false
}

func getProcessInfo(hwnd uintptr) (appName string, processName string) {
	if hwnd == 0 {
		return "", ""
	}

	var pid uint32
	_, _, _ = procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return "", ""
	}

	// Self-process check
	if pid == uint32(os.Getpid()) {
		return "Vinta", "Vinta.exe"
	}

	hProcess, _, _ := procOpenProcess.Call(
		uintptr(processQueryLimitedInformation),
		0,
		uintptr(pid),
	)
	if hProcess != 0 {
		defer procCloseHandle.Call(hProcess)

		var pathBuf [1024]uint16
		size := uint32(len(pathBuf))
		ret, _, _ := procQueryFullProcessImageNameW.Call(
			hProcess,
			0,
			uintptr(unsafe.Pointer(&pathBuf[0])),
			uintptr(unsafe.Pointer(&size)),
		)
		if ret != 0 && size > 0 {
			fullPath := syscall.UTF16ToString(pathBuf[:size])
			exeName := filepath.Base(fullPath)
			processName = exeName
		}
	}

	// Optional window title fallback for display name
	var titleBuf [256]uint16
	tLen, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&titleBuf[0])), uintptr(len(titleBuf)))
	if tLen > 0 {
		appName = syscall.UTF16ToString(titleBuf[:tLen])
	}

	return appName, processName
}

func formatSnippet(text string, maxLen int) string {
	clean := strings.ReplaceAll(text, "\r\n", " ")
	clean = strings.ReplaceAll(clean, "\n", " ")
	clean = strings.ReplaceAll(clean, "\t", " ")

	runes := []rune(clean)
	if len(runes) <= maxLen {
		return string(runes)
	}
	return string(runes[:maxLen]) + "..."
}

func utf16PtrToString(p *uint16) string {
	if p == nil {
		return ""
	}
	var runes []uint16
	for *p != 0 {
		runes = append(runes, *p)
		p = (*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + 2))
	}
	return string(utf16.Decode(runes))
}

func itoa(val int) string {
	if val == 0 {
		return "0"
	}
	var digits []byte
	v := val
	for v > 0 {
		d := byte(v % 10)
		digits = append([]byte{'0' + d}, digits...)
		v /= 10
	}
	return string(digits)
}
