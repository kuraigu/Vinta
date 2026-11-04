// Package services provides native platform integrations and hardware monitoring bridges.
package services

import (
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/sys/windows"
)

var (
	procGetSystemTimes       = kernel32.NewProc("GetSystemTimes")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")

	pdh                              = syscall.NewLazyDLL("pdh.dll")
	procPdhOpenQueryW                = pdh.NewProc("PdhOpenQueryW")
	procPdhCloseQuery               = pdh.NewProc("PdhCloseQuery")
	procPdhAddEnglishCounterW        = pdh.NewProc("PdhAddEnglishCounterW")
	procPdhCollectQueryData          = pdh.NewProc("PdhCollectQueryData")
	procPdhGetFormattedCounterArrayW = pdh.NewProc("PdhGetFormattedCounterArrayW")
)

type filetime struct {
	dwLowDateTime  uint32
	dwHighDateTime uint32
}

type memoryStatusEx struct {
	cbSize                  uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

type pdhFmtCounterValueItem struct {
	szName      *uint16
	cStatus     uint32
	_pad        uint32
	doubleValue float64
}

// SystemTelemetry models snapshot metrics for hardware load and memory allocation.
type SystemTelemetry struct {
	CPUPercent  float64 `json:"cpuPercent"`
	RAMPercent  float64 `json:"ramPercent"`
	RAMUsedMB   uint64  `json:"ramUsedMB"`
	RAMTotalMB  uint64  `json:"ramTotalMB"`
	GPUPercent  float64 `json:"gpuPercent"`
	TimestampMs int64   `json:"timestampMs"`
}

type telemetrySampler struct {
	mu          sync.Mutex
	prevIdle    uint64
	prevKernel  uint64
	prevUser    uint64
	hasPrevTime bool

	pdhQuery   uintptr
	pdhCounter uintptr
	pdhReady   bool
	gpuBuf     []byte

	activeListeners atomic.Int32
	stopCh          chan struct{}
	isRunning       bool
}

var sampler = &telemetrySampler{}

func filetimeToUint64(ft filetime) uint64 {
	return (uint64(ft.dwHighDateTime) << 32) | uint64(ft.dwLowDateTime)
}

func (s *telemetrySampler) sampleCPU() float64 {
	var idle, kernel, user filetime
	r, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idle)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)
	if r == 0 {
		return 0
	}

	idleVal := filetimeToUint64(idle)
	kernelVal := filetimeToUint64(kernel)
	userVal := filetimeToUint64(user)

	if !s.hasPrevTime {
		s.prevIdle = idleVal
		s.prevKernel = kernelVal
		s.prevUser = userVal
		s.hasPrevTime = true
		return 0
	}

	idleDelta := idleVal - s.prevIdle
	kernelDelta := kernelVal - s.prevKernel
	userDelta := userVal - s.prevUser

	s.prevIdle = idleVal
	s.prevKernel = kernelVal
	s.prevUser = userVal

	totalDelta := kernelDelta + userDelta
	if totalDelta == 0 {
		return 0
	}

	if totalDelta < idleDelta {
		return 0
	}

	cpuPercent := float64(totalDelta-idleDelta) / float64(totalDelta) * 100.0
	if cpuPercent < 0 {
		return 0
	}
	if cpuPercent > 100 {
		return 100
	}
	return cpuPercent
}

func (s *telemetrySampler) sampleRAM() (float64, uint64, uint64) {
	var mem memoryStatusEx
	mem.cbSize = uint32(unsafe.Sizeof(mem))

	r, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&mem)))
	if r == 0 {
		return 0, 0, 0
	}

	totalMB := mem.ullTotalPhys / (1024 * 1024)
	availMB := mem.ullAvailPhys / (1024 * 1024)
	usedMB := totalMB - availMB

	pct := float64(mem.dwMemoryLoad)
	if totalMB > 0 {
		pct = (float64(usedMB) / float64(totalMB)) * 100.0
	}

	return pct, usedMB, totalMB
}

func (s *telemetrySampler) initPDH() {
	if s.pdhReady {
		return
	}
	if procPdhOpenQueryW.Find() != nil || procPdhAddEnglishCounterW.Find() != nil {
		return
	}

	var q uintptr
	r, _, _ := procPdhOpenQueryW.Call(0, 0, uintptr(unsafe.Pointer(&q)))
	if r != 0 {
		return
	}

	counterPath, err := windows.UTF16PtrFromString(`\GPU Engine(*)\Utilization Percentage`)
	if err != nil {
		_, _, _ = procPdhCloseQuery.Call(q)
		return
	}

	var c uintptr
	r, _, _ = procPdhAddEnglishCounterW.Call(q, uintptr(unsafe.Pointer(counterPath)), 0, uintptr(unsafe.Pointer(&c)))
	if r != 0 {
		_, _, _ = procPdhCloseQuery.Call(q)
		return
	}

	s.pdhQuery = q
	s.pdhCounter = c
	s.pdhReady = true

	_, _, _ = procPdhCollectQueryData.Call(s.pdhQuery)
}

func (s *telemetrySampler) closePDH() {
	if s.pdhReady && s.pdhQuery != 0 {
		_, _, _ = procPdhCloseQuery.Call(s.pdhQuery)
		s.pdhQuery = 0
		s.pdhCounter = 0
		s.pdhReady = false
		s.gpuBuf = nil
	}
}

func (s *telemetrySampler) sampleGPU() float64 {
	if !s.pdhReady {
		s.initPDH()
		if !s.pdhReady {
			return 0
		}
	}

	r, _, _ := procPdhCollectQueryData.Call(s.pdhQuery)
	if r != 0 {
		return 0
	}

	var bufSize uint32
	var itemCount uint32
	const pdhFmtDouble = 0x00000200

	r, _, _ = procPdhGetFormattedCounterArrayW.Call(
		s.pdhCounter,
		uintptr(pdhFmtDouble),
		uintptr(unsafe.Pointer(&bufSize)),
		uintptr(unsafe.Pointer(&itemCount)),
		0,
	)

	const pdhMoreData = 0x800007D2
	if r != pdhMoreData || bufSize == 0 || itemCount == 0 {
		return 0
	}

	if uint32(len(s.gpuBuf)) < bufSize {
		s.gpuBuf = make([]byte, bufSize)
	}

	r, _, _ = procPdhGetFormattedCounterArrayW.Call(
		s.pdhCounter,
		uintptr(pdhFmtDouble),
		uintptr(unsafe.Pointer(&bufSize)),
		uintptr(unsafe.Pointer(&itemCount)),
		uintptr(unsafe.Pointer(&s.gpuBuf[0])),
	)
	if r != 0 {
		return 0
	}

	items := (*[1 << 16]pdhFmtCounterValueItem)(unsafe.Pointer(&s.gpuBuf[0]))[:itemCount:itemCount]
	var totalGPU float64
	for i := range items {
		if items[i].cStatus == 0 && items[i].doubleValue > 0 {
			totalGPU += items[i].doubleValue
		}
	}

	if totalGPU > 100 {
		totalGPU = 100
	}
	return totalGPU
}

// GetCurrentTelemetry returns an instantaneous hardware utilization snapshot.
func GetCurrentTelemetry() SystemTelemetry {
	sampler.mu.Lock()
	defer sampler.mu.Unlock()

	cpu := sampler.sampleCPU()
	ramPct, ramUsed, ramTotal := sampler.sampleRAM()
	gpu := sampler.sampleGPU()

	return SystemTelemetry{
		CPUPercent:  cpu,
		RAMPercent:  ramPct,
		RAMUsedMB:   ramUsed,
		RAMTotalMB:  ramTotal,
		GPUPercent:  gpu,
		TimestampMs: time.Now().UnixMilli(),
	}
}

// StartTelemetryStream starts periodic 1 Hz telemetry broadcasts to the frontend.
func StartTelemetryStream() {
	sampler.mu.Lock()
	defer sampler.mu.Unlock()

	sampler.activeListeners.Add(1)
	if sampler.isRunning {
		return
	}

	sampler.isRunning = true
	sampler.stopCh = make(chan struct{})

	go func() {
		ticker := time.NewTicker(1000 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-sampler.stopCh:
				return
			case <-ticker.C:
				if sampler.activeListeners.Load() <= 0 {
					continue
				}

				hiddenMu.RLock()
				hidden := isAppHidden
				hiddenMu.RUnlock()

				if hidden {
					continue
				}

				data := GetCurrentTelemetry()
				if app := application.Get(); app != nil {
					app.Event.Emit("system:telemetry", data)
				}
			}
		}
	}()
}

// StopTelemetryStream decrements the active subscription count and halts background sampling when idle.
func StopTelemetryStream() {
	sampler.mu.Lock()
	defer sampler.mu.Unlock()

	remaining := sampler.activeListeners.Add(-1)
	if remaining <= 0 {
		sampler.activeListeners.Store(0)
		if sampler.isRunning {
			sampler.isRunning = false
			if sampler.stopCh != nil {
				close(sampler.stopCh)
				sampler.stopCh = nil
			}
			sampler.closePDH()
		}
	}
}

