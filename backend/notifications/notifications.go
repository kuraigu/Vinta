package notifications

import (
	"os"
	"path/filepath"
	"time"
)

// NotificationPayload is what we send to the frontend
type NotificationPayload struct {
	AppName   string `json:"appName"`
	RawAppID  string `json:"rawAppId"`
	Icon      string `json:"icon"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	HeroImage string `json:"heroImage"`
}

func logMsg(msg string) {
	f, _ := os.OpenFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "vinta_debug.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		_, _ = f.WriteString(time.Now().Format(time.RFC3339) + " - " + msg + "\n")
		_ = f.Close()
	}
}
