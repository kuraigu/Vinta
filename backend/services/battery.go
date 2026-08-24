package services

import (
	"os/exec"
	"syscall"
	"unsafe"
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemPowerStatus = kernel32.NewProc("GetSystemPowerStatus")
)

// systemPowerStatus matches the Win32 SYSTEM_POWER_STATUS struct.
type systemPowerStatus struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	SystemStatusFlag    byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

// BatteryStatus represents the structured battery and power status for the frontend.
type BatteryStatus struct {
	HasBattery      bool   `json:"hasBattery"`
	IsCharging      bool   `json:"isCharging"`
	Percentage      int    `json:"percentage"`
	BatterySaver    bool   `json:"batterySaver"`
	LifeTimeSeconds int    `json:"lifeTimeSeconds"`
	State           string `json:"state"` // "charging", "discharging", "full", "desktop", "unknown"
}

// GetBatteryStatus queries Win32 GetSystemPowerStatus and returns the current battery & power state.
func GetBatteryStatus() BatteryStatus {
	var sps systemPowerStatus
	r, _, _ := procGetSystemPowerStatus.Call(uintptr(unsafe.Pointer(&sps)))
	if r == 0 {
		return BatteryStatus{
			HasBattery: false,
			Percentage: 0,
			State:      "unknown",
		}
	}

	// BatteryFlag == 128 means no system battery installed (Desktop PC)
	// BatteryLifePercent == 255 means battery status is unknown / not present
	if sps.BatteryFlag&128 != 0 || sps.BatteryLifePercent == 255 {
		return BatteryStatus{
			HasBattery:   false,
			IsCharging:   sps.ACLineStatus == 1,
			Percentage:   100,
			BatterySaver: sps.SystemStatusFlag == 1,
			State:        "desktop",
		}
	}

	hasBattery := true
	pct := int(sps.BatteryLifePercent)
	if pct > 100 {
		pct = 100
	}
	if pct < 0 {
		pct = 0
	}

	isCharging := sps.BatteryFlag&8 != 0 || (sps.ACLineStatus == 1 && pct < 100)
	isFull := pct == 100 && sps.ACLineStatus == 1
	batterySaver := sps.SystemStatusFlag == 1

	state := "discharging"
	if isFull {
		state = "full"
	} else if isCharging {
		state = "charging"
	}

	lifeTime := int(sps.BatteryLifeTime)
	if sps.BatteryLifeTime == 0xFFFFFFFF {
		lifeTime = -1
	}

	return BatteryStatus{
		HasBattery:      hasBattery,
		IsCharging:      isCharging,
		Percentage:      pct,
		BatterySaver:    batterySaver,
		LifeTimeSeconds: lifeTime,
		State:           state,
	}
}

// SetBatterySaver toggles Windows Battery Saver threshold (100% to force ON, 20% default to restore).
func SetBatterySaver(enabled bool) bool {
	val := "20"
	if enabled {
		val = "100"
	}
	cmd1 := exec.Command("powercfg", "/setdcvalueindex", "SCHEME_CURRENT", "SUB_ENERGYSAVER", "ESBATTTHRESHOLD", val)
	cmd1.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd1.Run()

	cmd2 := exec.Command("powercfg", "/setactive", "SCHEME_CURRENT")
	cmd2.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd2.Run()

	return true
}
