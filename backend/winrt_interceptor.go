package backend

import (
	"context"
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"Vinta/backend/winrt/windows/ui/notifications"
	"Vinta/backend/winrt/windows/ui/notifications/management"
	"github.com/go-ole/go-ole"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// StartNotificationListener initializes the WinRT COM apartment
// and listens to UserNotificationListener events.
func StartNotificationListener(ctx context.Context) {
	// Initialize COM for the current thread
	err := ole.RoInitialize(1) // 1 = RO_INIT_MULTITHREADED
	if err != nil {
		logMsg("Failed to initialize WinRT COM: " + err.Error())
		return
	}

	// Get the UserNotificationListener singleton
	listener, err := management.UserNotificationListenerGetCurrent()
	if err != nil {
		logMsg("Failed to get UserNotificationListener: " + err.Error())
		return
	}

	// Request access
	asyncOp, err := listener.RequestAccessAsync()
	if err != nil {
		logMsg("Failed to call RequestAccessAsync: " + err.Error())
		return
	}

	// Wait for the async operation to complete by polling GetResults
	var status management.UserNotificationListenerAccessStatus
	for {
		res, err := asyncOp.GetResults()
		if err == nil {
			status = management.UserNotificationListenerAccessStatus(uintptr(res))
			break
		}
		if oleErr, ok := err.(*ole.OleError); ok && oleErr.Code() != 0x8000000E {
			logMsg("RequestAccessAsync failed")
			runtime.EventsEmit(ctx, "notification_permission_denied", true)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	if status != management.UserNotificationListenerAccessStatusAllowed {
		logMsg(fmt.Sprintf("Notification access denied or unspecified: %d", status))
		runtime.EventsEmit(ctx, "notification_permission_denied", true)
		return
	}

	logMsg("WinRT Notification access granted! Polling notifications...")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Keep track of processed IDs
	processedIds := make(map[uint32]bool)
	isFirstPoll := true

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			asyncNotifs, err := listener.GetNotificationsAsync(notifications.NotificationKindsToast)
			if err != nil {
				continue
			}

			// Wait for async
			var notifsView unsafe.Pointer
			for {
				res, err := asyncNotifs.GetResults()
				if err == nil {
					notifsView = res
					break
				}
				if oleErr, ok := err.(*ole.OleError); ok && oleErr.Code() != 0x8000000E {
					logMsg(fmt.Sprintf("GetResults failed: %v", oleErr))
					break // error
				}
				time.Sleep(50 * time.Millisecond)
			}
			if notifsView == nil {
				continue
			}

			// notifsView is a raw COM pointer to IVectorView<UserNotification>
			// VTable for IVectorView:
			// 0..2 IUnknown
			// 3..5 IInspectable
			// 6 GetAt(index, &out)
			// 7 GetSize(&size)

			type iVectorViewVtbl struct {
				ole.IInspectableVtbl
				GetAt   uintptr
				GetSize uintptr
			}

			v := (*ole.IInspectable)(notifsView)
			vtbl := (*iVectorViewVtbl)(unsafe.Pointer(v.RawVTable))

			var size uint32
			hr, _, _ := syscall.SyscallN(
				vtbl.GetSize,
				uintptr(unsafe.Pointer(v)),
				uintptr(unsafe.Pointer(&size)),
			)
			if hr != 0 {
				logMsg(fmt.Sprintf("GetSize failed with hr: %x", hr))
				continue
			}

			for i := uint32(0); i < size; i++ {
				var un *notifications.UserNotification
				hr, _, _ = syscall.SyscallN(
					vtbl.GetAt,
					uintptr(unsafe.Pointer(v)),
					uintptr(i),
					uintptr(unsafe.Pointer(&un)),
				)
				if hr != 0 || un == nil {
					continue
				}

				id, err := un.GetId()
				if err != nil {
					continue
				}
				if processedIds[id] {
					continue
				}

				processedIds[id] = true

				// On the first poll, just seed existing IDs so we don't replay
				// all notifications that were already there before Vinta started.
				if isFirstPoll {
					continue
				}

				// Process in a safe wrapper so a single bad notification never
				// crashes the whole listener goroutine.
				func() {
					defer func() {
						if r := recover(); r != nil {
							logMsg(fmt.Sprintf("PANIC processing notification id=%d: %v", id, r))
						}
					}()

					// --- Extract notification text + image content ---
					title, body, heroImage := extractNotifContent(un)
					// --- Extract AppUserModelId ---
					appName := extractAppId(un)
					if appName == "" {
						appName = title
					}
					if appName == "" {
						appName = "Notification"
					}

					payload := NotificationPayload{
						AppName:   appName,
						Title:     title,
						Message:   body,
						HeroImage: heroImage,
					}

					logMsg(fmt.Sprintf("Emitting notification: title=%q body=%q heroImage=%q", title, body, heroImage))
					runtime.EventsEmit(ctx, "notification", payload)
				}()
			}

			if isFirstPoll {
				logMsg(fmt.Sprintf("Seeded %d existing notification IDs (not emitting)", size))
				isFirstPoll = false
			}
		}
	}
}

// extractAppId reads the AppUserModelId from UserNotification.AppInfo.Id.
func extractAppId(un *notifications.UserNotification) string {
	itf := un.MustQueryInterface(ole.NewGUID("adf7e52f-4e53-42d5-9c33-eb5ea515b23e"))
	if itf == nil {
		return ""
	}
	defer itf.Release()

	type iUserNotificationVtbl struct {
		ole.IInspectableVtbl
		GetNotification uintptr
		GetAppInfo      uintptr
		GetId           uintptr
		GetCreationTime uintptr
	}
	type iUserNotification struct {
		ole.IInspectable
	}
	v := (*iUserNotification)(unsafe.Pointer(itf))
	vtbl := (*iUserNotificationVtbl)(unsafe.Pointer(v.RawVTable))

	var appInfoRaw uintptr
	hr, _, _ := syscall.SyscallN(
		vtbl.GetAppInfo,
		uintptr(unsafe.Pointer(v)),
		uintptr(unsafe.Pointer(&appInfoRaw)),
	)
	if hr != 0 || appInfoRaw == 0 {
		return ""
	}
	type comVtbl struct {
		QueryInterface uintptr
		AddRef         uintptr
		Release        uintptr
	}
	defer func() {
		p := *(*uintptr)(unsafe.Pointer(appInfoRaw))
		cv := (*comVtbl)(unsafe.Pointer(p))
		syscall.SyscallN(cv.Release, appInfoRaw)
	}()

	iidAppInfo := ole.NewGUID("cf7f59b3-6a09-4de8-a6c0-5792d56880d1")
	var appInfoItfRaw uintptr
	appInfoVtblPtr := *(*uintptr)(unsafe.Pointer(appInfoRaw))
	appInfoVtbl := (*comVtbl)(unsafe.Pointer(appInfoVtblPtr))
	hr, _, _ = syscall.SyscallN(
		appInfoVtbl.QueryInterface,
		appInfoRaw,
		uintptr(unsafe.Pointer(iidAppInfo)),
		uintptr(unsafe.Pointer(&appInfoItfRaw)),
	)
	if hr != 0 || appInfoItfRaw == 0 {
		return ""
	}
	defer func() {
		p := *(*uintptr)(unsafe.Pointer(appInfoItfRaw))
		cv := (*comVtbl)(unsafe.Pointer(p))
		syscall.SyscallN(cv.Release, appInfoItfRaw)
	}()

	type iAppInfoVtbl struct {
		ole.IInspectableVtbl
		GetId uintptr
	}
	aiVtblPtr := *(*uintptr)(unsafe.Pointer(appInfoItfRaw))
	aiVtbl := (*iAppInfoVtbl)(unsafe.Pointer(aiVtblPtr))

	var idHStr uintptr
	hr, _, _ = syscall.SyscallN(
		aiVtbl.GetId,
		appInfoItfRaw,
		uintptr(unsafe.Pointer(&idHStr)),
	)
	if hr == 0 && idHStr != 0 {
		hs := ole.HString(idHStr)
		id := hs.String()
		ole.DeleteHString(hs)
		return id
	}
	return ""
}

// extractNotifContent extracts (title, body, heroImageURL) from a UserNotification.
func extractNotifContent(un *notifications.UserNotification) (string, string, string) {
	notif, err := un.GetNotification()
	if err != nil || notif == nil {
		return "", "", ""
	}

	visual, err := notif.GetVisual()
	if err != nil || visual == nil {
		return "", "", ""
	}

	// Try "ToastGeneric" first
	binding, err := visual.GetBinding("ToastGeneric")
	if err != nil || binding == nil {
		return "", "", ""
	}

	textElementsPtr, err := binding.GetTextElements()
	if err != nil || textElementsPtr == nil {
		return "", "", ""
	}

	type iVectorViewVtbl struct {
		ole.IInspectableVtbl
		GetAt   uintptr
		GetSize uintptr
	}
	v := (*ole.IInspectable)(textElementsPtr)
	vtbl := (*iVectorViewVtbl)(unsafe.Pointer(v.RawVTable))

	var size uint32
	hr, _, _ := syscall.SyscallN(
		vtbl.GetSize,
		uintptr(unsafe.Pointer(v)),
		uintptr(unsafe.Pointer(&size)),
	)

	texts := make([]string, 0, size)
	if hr == 0 && size > 0 {
		for i := uint32(0); i < size; i++ {
			var elem *notifications.AdaptiveNotificationText
			hr, _, _ = syscall.SyscallN(
				vtbl.GetAt,
				uintptr(unsafe.Pointer(v)),
				uintptr(i),
				uintptr(unsafe.Pointer(&elem)),
			)
			if hr != 0 || elem == nil {
				continue
			}
			t, err := elem.GetText()
			if err == nil {
				t = strings.TrimSpace(t)
				if t != "" {
					texts = append(texts, t)
				}
			}
		}
	}

	title := ""
	body := ""
	if len(texts) > 0 {
		title = texts[0]
	}
	if len(texts) > 1 {
		body = texts[1]
	}

	// --- Image elements (disabled to prevent crashes) ---
	heroImage := ""

	return title, body, heroImage
}
