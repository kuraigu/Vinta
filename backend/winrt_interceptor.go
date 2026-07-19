package backend

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"Vinta/backend/winrt/windows/ui/notifications"
	"Vinta/backend/winrt/windows/ui/notifications/management"
	"github.com/go-ole/go-ole"
	_ "github.com/mattn/go-sqlite3"
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

	// Open the WPN database once and reuse the connection for all notifications.
	// This avoids the overhead of opening/closing SQLite on every poll.
	var wpnDb *sql.DB
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		dbPath := filepath.Join(localAppData, "Microsoft", "Windows", "Notifications", "wpndatabase.db")
		if _, statErr := os.Stat(dbPath); statErr == nil {
			wpnDb, err = sql.Open("sqlite3", "file:"+dbPath+"?mode=ro")
			if err != nil {
				logMsg("Failed to open wpndatabase.db: " + err.Error())
				wpnDb = nil
			} else {
				wpnDb.SetMaxOpenConns(1)
				defer wpnDb.Close()
				logMsg("Opened wpndatabase.db for persistent queries")
			}
		}
	}

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

					// --- Extract notification text + image content + appName from DB ---
					title, body, heroImage, dbAppName := extractNotifContent(un, id, wpnDb)

					// --- Extract AppUserModelId from WinRT (better fallback) ---
					// "App" is the generic WPN PrimaryId for all unpackaged Win32 apps — not useful.
					// Always try extractAppId(un) which may return a real package AUMID
					// (e.g. Google Play Games AUMID for mobile games running on Windows).
					winrtId := extractAppId(un)
					appName := dbAppName
					if appName == "" || strings.EqualFold(appName, "App") {
						if winrtId != "" {
							appName = winrtId
						}
					}
					logMsg(fmt.Sprintf("Notification id=%d: dbAppName=%q winrtId=%q → appName=%q title=%q", id, dbAppName, winrtId, appName, title))

					payload := NotificationPayload{
						AppName:   appName,
						Title:     title,
						Message:   body,
						HeroImage: heroImage,
					}

					logMsg(fmt.Sprintf("Emitting notification: appName=%q title=%q body=%q heroImage_len=%d", appName, title, body, len(heroImage)))
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

// extractNotifContent extracts (title, body, heroImageURL, appName) from a UserNotification.
func extractNotifContent(un *notifications.UserNotification, id uint32, wpnDb *sql.DB) (string, string, string, string) {
	notif, err := un.GetNotification()
	if err != nil || notif == nil {
		return "", "", "", ""
	}

	visual, err := notif.GetVisual()
	if err != nil || visual == nil {
		return "", "", "", ""
	}

	// Try "ToastGeneric" first
	binding, err := visual.GetBinding("ToastGeneric")
	if err != nil || binding == nil {
		return "", "", "", ""
	}

	textElementsPtr, err := binding.GetTextElements()
	if err != nil || textElementsPtr == nil {
		return "", "", "", ""
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

	// --- Image elements & AppName (fetched via on-demand SQLite payload query) ---
	heroImage, dbAppName := extractDbDetails(id, title, wpnDb)

	return title, body, heroImage, dbAppName
}

// WpnToastImage represents the image nodes in a Toast XML
type WpnToastImage struct {
	Src       string `xml:"src,attr"`
	Placement string `xml:"placement,attr"`
}

// WpnToastBinding represents the binding node in a Toast XML
type WpnToastBinding struct {
	Images []WpnToastImage `xml:"image"`
}

// WpnToastVisual represents the visual node in a Toast XML
type WpnToastVisual struct {
	Binding WpnToastBinding `xml:"binding"`
}

// WpnToastXML represents the root Toast XML structure
type WpnToastXML struct {
	Visual WpnToastVisual `xml:"visual"`
}

// extractDbDetails queries the wpndatabase.db for the XML payload and AUMID of a notification.
// It retries a few times to handle the case where Windows writes the DB row asynchronously
// after the WinRT notification event fires (which causes ErrNoRows on first query).
// If the ID lookup fails entirely, it falls back to searching by notification title content
// (handles apps like Rise of Kingdoms that send via two channels: WPN and a non-WPN bridge).
func extractDbDetails(id uint32, titleHint string, db *sql.DB) (string, string) {
	if db == nil {
		return "", ""
	}

	const maxRetries = 10
	const retryDelay = 200 * time.Millisecond

	var payload []byte
	var primaryId sql.NullString

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := db.QueryRow(`
			SELECT n.Payload, h.PrimaryId
			FROM Notification n
			LEFT JOIN NotificationHandler h ON n.HandlerId = h.RecordId
			WHERE n.Id = ?`, id).Scan(&payload, &primaryId)
		if err == sql.ErrNoRows {
			// Notification row not yet written — wait and retry
			logMsg(fmt.Sprintf("DB miss for id=%d attempt %d/%d, retrying...", id, attempt+1, maxRetries))
			if attempt < maxRetries-1 {
				time.Sleep(retryDelay)
				continue
			}
			logMsg(fmt.Sprintf("DB: gave up waiting for id=%d after %d attempts", id, maxRetries))
			return "", ""
		}
		if err != nil {
			logMsg(fmt.Sprintf("DB error for id=%d: %v", id, err))
			return "", ""
		}
		// Row found — but PrimaryId may be NULL if NotificationHandler isn't linked yet.
		// Retry until we get a non-empty PrimaryId (or we have the payload).
		if !primaryId.Valid || primaryId.String == "" {
			logMsg(fmt.Sprintf("DB hit for id=%d attempt %d but PrimaryId is NULL, retrying...", id, attempt+1))
			if attempt < maxRetries-1 {
				time.Sleep(retryDelay)
				continue
			}
			logMsg(fmt.Sprintf("DB: PrimaryId never populated for id=%d after %d attempts", id, maxRetries))
			// Still use payload for hero image even without an appName
			break
		}
		logMsg(fmt.Sprintf("DB hit for id=%d on attempt %d, primaryId=%q", id, attempt, primaryId.String))
		break
	}

	// If ID lookup gave up and we have no payload, try a content-based fallback.
	// Some apps (e.g. Rise of Kingdoms via Google Play Games) send one transient notification
	// that never reaches wpndatabase.db, and a second WPN-persisted copy with the correct PrimaryId.
	// Searching by title content finds the WPN copy.
	if len(payload) == 0 && titleHint != "" {
		like := "%" + strings.ReplaceAll(titleHint, "%", "\\%") + "%"
		var fallbackPrimary sql.NullString
		var fallbackPayload []byte
		err := db.QueryRow(`
			SELECT n.Payload, h.PrimaryId
			FROM Notification n
			LEFT JOIN NotificationHandler h ON n.HandlerId = h.RecordId
			WHERE CAST(n.Payload AS TEXT) LIKE ? ESCAPE '\'
			  AND h.PrimaryId IS NOT NULL
			  AND h.PrimaryId != ''
			  AND h.PrimaryId != 'App'
			ORDER BY n.Id DESC
			LIMIT 1`, like).Scan(&fallbackPayload, &fallbackPrimary)
		if err == nil && fallbackPrimary.Valid && fallbackPrimary.String != "" {
			logMsg(fmt.Sprintf("DB content fallback for id=%d: found primaryId=%q via title %q", id, fallbackPrimary.String, titleHint))
			payload = fallbackPayload
			primaryId = fallbackPrimary
		}
	}

	appName := ""
	if primaryId.Valid && primaryId.String != "" {
		appName = primaryId.String
	}

	heroImage := ""
	if len(payload) > 0 {
		startIdx := bytes.Index(payload, []byte("<toast"))
		endIdx := bytes.Index(payload, []byte("</toast>"))
		if startIdx != -1 && endIdx != -1 {
			xmlData := payload[startIdx : endIdx+8]
			var toast WpnToastXML
			if err := xml.Unmarshal(xmlData, &toast); err != nil {
				return "", appName
			}

			for _, img := range toast.Visual.Binding.Images {
				if img.Placement == "hero" {
					heroImage = img.Src
					break
				}
			}
		}
	}

	if heroImage != "" {
		if strings.HasPrefix(heroImage, "file:///") {
			path := strings.TrimPrefix(heroImage, "file:///")
			path = filepath.Clean(path)
			if b, err := os.ReadFile(path); err == nil {
				heroImage = "data:image/png;base64," + base64.StdEncoding.EncodeToString(b)
			} else {
				heroImage = ""
			}
		} else if filepath.IsAbs(heroImage) {
			if b, err := os.ReadFile(heroImage); err == nil {
				heroImage = "data:image/png;base64," + base64.StdEncoding.EncodeToString(b)
			} else {
				heroImage = ""
			}
		}
	}

	return heroImage, appName
}
