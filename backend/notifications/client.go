// Package notifications provides the interface for launching and interacting with the C# notification interceptor.
package notifications

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// StartNotificationListener spawns the C# notification interceptor
// and streams its output.
func StartNotificationListener(ctx context.Context) {
	exePath := filepath.Join("backend", "notifications", "bin", "Release", "net8.0-windows10.0.19041.0", "win-x64", "publish", "VintaNotifications.exe")
	absExePath, _ := filepath.Abs(exePath)

	cmd := exec.CommandContext(ctx, absExePath)
	cmd.Dir = filepath.Dir(absExePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logMsg("Failed to pipe stdout: " + err.Error())
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logMsg("Failed to pipe stderr: " + err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		logMsg("Failed to start VintaNotifications.exe: " + err.Error())
		return
	}

	go func() {
		<-ctx.Done()
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	logMsg("C# Notification Interceptor started successfully")

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			logMsg("[C# Interceptor ERROR]: " + scanner.Text())
		}
	}()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		logMsg(fmt.Sprintf("[C# Interceptor]: %s", line))

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			logMsg("Failed to parse JSON from interceptor: " + err.Error())
			continue
		}

		// Handle errors or status messages from C#
		if _, hasError := data["error"]; hasError {
			if data["error"] == "Access denied" {
				runtime.EventsEmit(ctx, "notification_permission_denied", true)
			}
			continue
		}
		if _, hasStatus := data["status"]; hasStatus {
			continue
		}

		// Otherwise, it is a notification payload
		payload := NotificationPayload{
			AppName:   getString(data, "appName"),
			Title:     getString(data, "title"),
			Message:   getString(data, "message"),
			HeroImage: getString(data, "heroImage"),
			Icon:      getString(data, "icon"),
		}

		runtime.EventsEmit(ctx, "notification", payload)
	}

	if err := cmd.Wait(); err != nil {
		logMsg("C# Notification Interceptor exited with error: " + err.Error())
	} else {
		logMsg("C# Notification Interceptor exited cleanly")
	}
}

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
