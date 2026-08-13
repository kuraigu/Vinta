package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"
)

var (
	appDataDirOnce sync.Once
	baseAppDataDir string
)

// GetAppDataDir returns the persistent root directory for Vinta under APPDATA.
// Uses %APPDATA%/FreeMatrix Studios/Vinta without hardcoding user paths.
func GetAppDataDir() string {
	appDataDirOnce.Do(func() {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = os.Getenv("LOCALAPPDATA")
		}
		if appData == "" {
			appData = os.TempDir()
		}
		baseAppDataDir = filepath.Join(appData, "FreeMatrix Studios", "Vinta")
		_ = os.MkdirAll(baseAppDataDir, 0755)
	})
	return baseAppDataDir
}

// GetDebugDir returns the debug directory under %APPDATA%/FreeMatrix Studios/Vinta/debug.
func GetDebugDir() string {
	debugDir := filepath.Join(GetAppDataDir(), "debug")
	_ = os.MkdirAll(debugDir, 0755)
	return debugDir
}

// GetCrashLogPath returns the path to crash.log inside the debug folder.
func GetCrashLogPath() string {
	return filepath.Join(GetDebugDir(), "crash.log")
}

// GetDebugLogPath returns the path to vinta_debug.log inside the debug folder.
func GetDebugLogPath() string {
	return filepath.Join(GetDebugDir(), "vinta_debug.log")
}

// LogCrash records any runtime error or panic stack trace to crash.log regardless of debug settings.
func LogCrash(err interface{}, stackTrace string) {
	logPath := GetCrashLogPath()
	f, fileErr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		return
	}
	defer func() { _ = f.Close() }()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("[%s] CRASH / UNHANDLED ERROR: %v\nStack Trace:\n%s\n----------------------------------------\n\n", timestamp, err, stackTrace)
	_, _ = f.WriteString(entry)
}

// SetupPanicRecovery registers a global defer handler to catch panics and write crash logs.
func SetupPanicRecovery() {
	if r := recover(); r != nil {
		stackTrace := string(debug.Stack())
		LogCrash(r, stackTrace)
	}
}

// LogDebug appends debug info to vinta_debug.log inside the debug directory.
func LogDebug(msg string) {
	logPath := GetDebugLogPath()
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("[%s] %s\n", timestamp, msg)
	_, _ = f.WriteString(entry)
}
