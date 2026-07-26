// Package services provides management and IPC communication for VintaServices.exe.
package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	stdinWriter io.WriteCloser
	cancelCtx   context.CancelFunc

	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procCreateJobObject          = kernel32.NewProc("CreateJobObjectW")
	procSetInformationJobObject  = kernel32.NewProc("SetInformationJobObject")
	procAssignProcessToJobObject = kernel32.NewProc("AssignProcessToJobObject")
	jobObjectHandle              uintptr
)

const (
	jobObjectExtendedLimitInformationClass = 9
	jobObjectLimitKillOnJobClose           = 0x2000
)

//nolint:revive // Windows API C structs
type JOBOBJECT_BASIC_LIMIT_INFORMATION struct {
	PerProcessUserTimeLimit int64
	PerJobUserTimeLimit     int64
	LimitFlags              uint32
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

//nolint:revive // Windows API C structs
type IO_COUNTERS struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

//nolint:revive // Windows API C structs
type JOBOBJECT_EXTENDED_LIMIT_INFORMATION struct {
	BasicLimitInformation JOBOBJECT_BASIC_LIMIT_INFORMATION
	IoInfo                IO_COUNTERS
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

func initJobObject() {
	job, _, _ := procCreateJobObject.Call(0, 0)
	if job != 0 {
		var info JOBOBJECT_EXTENDED_LIMIT_INFORMATION
		info.BasicLimitInformation.LimitFlags = jobObjectLimitKillOnJobClose
		_, _, _ = procSetInformationJobObject.Call(
			job,
			uintptr(jobObjectExtendedLimitInformationClass),
			uintptr(unsafe.Pointer(&info)),
			uintptr(unsafe.Sizeof(info)),
		)
		jobObjectHandle = job
	}
}

// StartServicesListener launches VintaServices.exe as a child process attached to a Windows Job Object.
func StartServicesListener(ctx context.Context) {
	initJobObject()

	runCtx, cancel := context.WithCancel(ctx)
	cancelCtx = cancel

	exePath, err := os.Executable()
	if err != nil {
		logMsg("Failed to get executable path: " + err.Error())
		return
	}

	dir := filepath.Dir(exePath)
	absExePath := filepath.Join(dir, "VintaServices.exe")

	if _, err := os.Stat(absExePath); os.IsNotExist(err) {
		absExePath = filepath.Join(dir, "..", "..", "backend", "services", "bin", "Release", "net8.0-windows10.0.19041.0", "win-x64", "publish", "VintaServices.exe")
	}

	if _, err := os.Stat(absExePath); os.IsNotExist(err) {
		absExePath = filepath.Join(dir, "backend", "services", "bin", "Release", "net8.0-windows10.0.19041.0", "win-x64", "publish", "VintaServices.exe")
	}

	if _, err := os.Stat(absExePath); os.IsNotExist(err) {
		logMsg("VintaServices.exe not found at: " + absExePath)
		return
	}

	go monitorAndRunServices(runCtx, absExePath)
}

// StopServicesListener forcefully terminates any running VintaServices.exe process tree.
func StopServicesListener() {
	if cancelCtx != nil {
		cancelCtx()
	}
	cmd := exec.Command("taskkill", "/F", "/IM", "VintaServices.exe", "/T")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Run()
}

// SendMediaControl writes a control command to VintaServices.exe stdin pipe
func SendMediaControl(action string) {
	if stdinWriter != nil {
		_, _ = fmt.Fprintf(stdinWriter, "COMMAND:%s\n", action)
	}
}

func monitorAndRunServices(ctx context.Context, absExePath string) {
	restartDelay := 1 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		runServiceProcess(ctx, absExePath)

		select {
		case <-ctx.Done():
			return
		case <-time.After(restartDelay):
			logMsg(fmt.Sprintf("Restarting VintaServices.exe (delay: %v)...", restartDelay))
			if restartDelay < 16*time.Second {
				restartDelay *= 2
			}
		}
	}
}

func runServiceProcess(ctx context.Context, absExePath string) {
	cmd := exec.CommandContext(ctx, absExePath)
	cmd.Dir = filepath.Dir(absExePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	var err error
	stdinWriter, err = cmd.StdinPipe()
	if err != nil {
		logMsg("Failed to pipe stdin: " + err.Error())
		return
	}

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
		logMsg("Failed to start VintaServices.exe: " + err.Error())
		return
	}

	// Assign child process to Windows Job Object so OS auto-kills child if parent dies
	if jobObjectHandle != 0 && cmd.Process != nil {
		_, _, _ = procAssignProcessToJobObject.Call(jobObjectHandle, uintptr(cmd.Process.Pid))
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			logMsg("[VintaServices Stderr] " + scanner.Text())
		}
	}()

	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err == nil {
			msgType, _ := raw["type"].(string)
			switch msgType {
			case "notification":
				runtime.EventsEmit(ctx, "notification", raw)
			case "media":
				runtime.EventsEmit(ctx, "media-update", raw)
			}
		}
	}

	_ = cmd.Wait()
}

func logMsg(msg string) {
	fmt.Printf("[VintaServices] %s\n", msg)
}
