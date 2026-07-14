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
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WpnNotification matches the structure of the Notification table in wpndatabase.db
type WpnNotification struct {
	PrimaryId   int
	Payload     []byte
	ArrivalTime uint64
}

// ToastText represents the text nodes in a Toast XML
type ToastText struct {
	ID   int    `xml:"id,attr"`
	Text string `xml:",chardata"`
}

// ToastImage represents the image nodes in a Toast XML
type ToastImage struct {
	Src       string `xml:"src,attr"`
	Placement string `xml:"placement,attr"`
}

// ToastBinding represents the binding node in a Toast XML
type ToastBinding struct {
	Texts  []ToastText  `xml:"text"`
	Images []ToastImage `xml:"image"`
}

// ToastVisual represents the visual node in a Toast XML
type ToastVisual struct {
	Binding ToastBinding `xml:"binding"`
}

// ToastXML represents the root Toast XML structure
type ToastXML struct {
	Visual ToastVisual `xml:"visual"`
}

// NotificationPayload is what we send to the frontend
type NotificationPayload struct {
	AppName string `json:"appName"`
	Icon    string `json:"icon"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

func parseToastPayload(payload []byte) (NotificationPayload, error) {
	startIdx := bytes.Index(payload, []byte("<toast"))
	endIdx := bytes.Index(payload, []byte("</toast>"))
	if startIdx == -1 || endIdx == -1 {
		return NotificationPayload{}, fmt.Errorf("no toast xml found")
	}

	xmlData := payload[startIdx : endIdx+8]
	var toast ToastXML
	if err := xml.Unmarshal(xmlData, &toast); err != nil {
		return NotificationPayload{}, err
	}

	var title, message string
	for _, t := range toast.Visual.Binding.Texts {
		if t.ID == 1 {
			title = t.Text
		} else if t.ID == 2 {
			message = t.Text
		} else if t.ID == 3 && message == "" { // sometimes ID=2 and 3 are message lines
			message = t.Text
		}
	}

	var icon string
	for _, img := range toast.Visual.Binding.Images {
		if img.Placement == "appLogoOverride" {
			icon = img.Src
			break
		}
		if icon == "" {
			icon = img.Src
		}
	}

	if icon != "" {
		if strings.HasPrefix(icon, "file:///") {
			path := strings.TrimPrefix(icon, "file:///")
			path = filepath.Clean(path)
			if b, err := os.ReadFile(path); err == nil {
				icon = "data:image/png;base64," + base64.StdEncoding.EncodeToString(b)
			} else {
				icon = ""
			}
		} else if filepath.IsAbs(icon) {
			if b, err := os.ReadFile(icon); err == nil {
				icon = "data:image/png;base64," + base64.StdEncoding.EncodeToString(b)
			} else {
				icon = ""
			}
		}
	}
	if icon == "" {
		icon = ""
	}

	return NotificationPayload{
		AppName: "Windows", // Temporarily assigned here, will be overridden by the DB query if available
		Icon:    icon,
		Title:   title,
		Message: message,
	}, nil
}

func logMsg(msg string) {
	f, _ := os.OpenFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "vinta_debug.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		f.WriteString(time.Now().Format(time.RFC3339) + " - " + msg + "\n")
		f.Close()
	}
}

// StartNotificationListener runs in a background goroutine and polls the wpndatabase
func StartNotificationListener(ctx context.Context) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return
	}

	dbPath := filepath.Join(localAppData, "Microsoft", "Windows", "Notifications", "wpndatabase.db")

	if _, err := os.Stat(dbPath); err != nil {
		return
	}

	db, err := sql.Open("sqlite3", "file:"+dbPath+"?mode=ro")
	if err != nil {
		logMsg("Failed to open notification DB: " + err.Error())
		return
	}
	defer db.Close()

	logMsg("Started Windows Notification Interceptor. Listening for toasts...")

	var lastArrivalTime uint64

	row := db.QueryRow("SELECT ArrivalTime FROM Notification WHERE Type='toast' ORDER BY ArrivalTime DESC LIMIT 1")
	row.Scan(&lastArrivalTime)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rows, err := db.Query(`
				SELECT n.Id, n.Payload, n.ArrivalTime, h.PrimaryId 
				FROM Notification n 
				LEFT JOIN NotificationHandler h ON n.HandlerId = h.RecordId 
				WHERE n.Type='toast' AND n.ArrivalTime > ? 
				ORDER BY n.ArrivalTime ASC`, lastArrivalTime)
			if err != nil {
				logMsg("Error querying notifications: " + err.Error())
				continue
			}

			var newMaxTime uint64 = lastArrivalTime
			for rows.Next() {
				var notif WpnNotification
				var rawAppName sql.NullString
				if err := rows.Scan(&notif.PrimaryId, &notif.Payload, &notif.ArrivalTime, &rawAppName); err != nil {
					logMsg("Error scanning notification: " + err.Error())
					continue
				}

				if notif.ArrivalTime > newMaxTime {
					newMaxTime = notif.ArrivalTime
				}

				payload, err := parseToastPayload(notif.Payload)

				if rawAppName.Valid && rawAppName.String != "" {
					parts := strings.Split(rawAppName.String, ".")
					payload.AppName = parts[len(parts)-1] // e.g. "Microsoft.Windows.Explorer" -> "Explorer"
				}

				if err != nil {
					logMsg(fmt.Sprintf("Error parsing payload: %v | Raw: %s", err, string(notif.Payload)))
				} else if payload.Title != "" {
					logMsg("Intercepted notification: " + payload.Title)
					runtime.EventsEmit(ctx, "notification", payload)
				} else {
					logMsg("Empty title in notification payload: " + string(notif.Payload))
				}
			}
			rows.Close()

			lastArrivalTime = newMaxTime
		}
	}
}
