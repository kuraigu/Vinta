package services

import (
	"encoding/json"
	"os/exec"
	"syscall"
	"unsafe"
)

var (
	bthprops                     = syscall.NewLazyDLL("bthprops.cpl")
	procBluetoothFindFirstRadio  = bthprops.NewProc("BluetoothFindFirstRadio")
	procBluetoothFindRadioClose  = bthprops.NewProc("BluetoothFindRadioClose")
	procBluetoothFindFirstDevice = bthprops.NewProc("BluetoothFindFirstDevice")
	procBluetoothFindDeviceClose = bthprops.NewProc("BluetoothFindDeviceClose")

	kernel32Dll     = syscall.NewLazyDLL("kernel32.dll")
	procCloseHandle = kernel32Dll.NewProc("CloseHandle")
)

type bluetoothFindRadioParams struct {
	dwSize uint32
}

type bluetoothDeviceSearchParams struct {
	dwSize               uint32
	fReturnAuthenticated int32
	fReturnRemembered    int32
	fReturnUnknown       int32
	fReturnConnected     int32
	fIssueInquiry        int32
	cTimeoutMultiplier   byte
	hRadio               syscall.Handle
}

type bluetoothDeviceInfo struct {
	dwSize         uint32
	address        uint64
	ulClass        uint32
	fConnected     int32
	fRemembered    int32
	fAuthenticated int32
	stLastSeen     [16]byte
	stLastUsed     [16]byte
	szName         [248]uint16
}

// checkBluetooth queries Windows Bluetooth radios and connected devices.
func checkBluetooth() (hasBT bool, enabled bool, connected bool, deviceName string) {
	params := bluetoothFindRadioParams{dwSize: 4}
	var hRadio syscall.Handle
	hFind, _, _ := procBluetoothFindFirstRadio.Call(
		uintptr(unsafe.Pointer(&params)),
		uintptr(unsafe.Pointer(&hRadio)),
	)
	if hFind == 0 {
		return false, false, false, ""
	}
	defer func() {
		_, _, _ = procBluetoothFindRadioClose.Call(hFind)
	}()
	defer func() {
		_, _, _ = procCloseHandle.Call(uintptr(hRadio))
	}()

	hasBT = true
	enabled = true

	// Check if any connected devices exist
	searchParams := bluetoothDeviceSearchParams{
		dwSize:           uint32(unsafe.Sizeof(bluetoothDeviceSearchParams{})),
		fReturnConnected: 1,
		hRadio:           hRadio,
	}
	var devInfo bluetoothDeviceInfo
	devInfo.dwSize = uint32(unsafe.Sizeof(devInfo))

	hDevFind, _, _ := procBluetoothFindFirstDevice.Call(
		uintptr(unsafe.Pointer(&searchParams)),
		uintptr(unsafe.Pointer(&devInfo)),
	)
	if hDevFind != 0 {
		defer func() {
			_, _, _ = procBluetoothFindDeviceClose.Call(hDevFind)
		}()
		connected = true
		deviceName = syscall.UTF16ToString(devInfo.szName[:])
	}

	return hasBT, enabled, connected, deviceName
}

// WirelessStatus represents the structured wireless and connectivity status.
type WirelessStatus struct {
	HasWifi             bool   `json:"hasWifi"`
	WifiEnabled         bool   `json:"wifiEnabled"`
	WifiConnected       bool   `json:"wifiConnected"`
	WifiSsid            string `json:"wifiSsid"`
	WifiSignal          int    `json:"wifiSignal"`
	HasBluetooth        bool   `json:"hasBluetooth"`
	BluetoothEnabled    bool   `json:"bluetoothEnabled"`
	BluetoothConnected  bool   `json:"bluetoothConnected"`
	BluetoothDeviceName string `json:"bluetoothDeviceName"`
	AirplaneMode        bool   `json:"airplaneMode"`
}

type nativeRadioStatus struct {
	HasWifi          bool   `json:"hasWifi"`
	WifiEnabled      bool   `json:"wifiEnabled"`
	WifiConnected    bool   `json:"wifiConnected"`
	WifiSsid         string `json:"wifiSsid"`
	WifiSignal       int    `json:"wifiSignal"`
	HasBluetooth     bool   `json:"hasBluetooth"`
	BluetoothEnabled bool   `json:"bluetoothEnabled"`
	AirplaneMode     bool   `json:"airplaneMode"`
}

// GetWirelessStatus queries Windows native Wi-Fi, Bluetooth, and Radio state.
func GetWirelessStatus() WirelessStatus {
	hasBT, btEnabled, btConnected, btDevice := checkBluetooth()

	res := WirelessStatus{
		HasWifi:             true,
		WifiEnabled:         true,
		WifiConnected:       false,
		WifiSignal:          0,
		HasBluetooth:        hasBT,
		BluetoothEnabled:    btEnabled,
		BluetoothConnected:  btConnected,
		BluetoothDeviceName: btDevice,
		AirplaneMode:        false,
	}

	// Query native WinRT radio and network manager in-process
	if radioJSON, ok := GetNativeRadiosStatus(); ok && radioJSON != "" {
		var nStatus nativeRadioStatus
		if err := json.Unmarshal([]byte(radioJSON), &nStatus); err == nil {
			res.HasWifi = nStatus.HasWifi
			res.WifiEnabled = nStatus.WifiEnabled
			res.WifiConnected = nStatus.WifiConnected
			res.WifiSsid = nStatus.WifiSsid
			res.WifiSignal = nStatus.WifiSignal
			res.HasBluetooth = nStatus.HasBluetooth || hasBT
			res.BluetoothEnabled = nStatus.BluetoothEnabled
			res.AirplaneMode = nStatus.AirplaneMode
		}
	}

	if !res.WifiEnabled || res.AirplaneMode {
		res.WifiConnected = false
	}

	return res
}

// SetWifiEnabled toggles the Wi-Fi radio on or off.
func SetWifiEnabled(enabled bool) bool {
	targetState := 2 // Off
	if enabled {
		targetState = 1 // On
	}
	if SetNativeRadioState(1, targetState) {
		return true
	}

	// Fallback via powershell if needed
	var action string
	if enabled {
		action = "enable"
	} else {
		action = "disable"
	}
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", "Get-NetAdapter | Where-Object { $_.InterfaceDescription -match 'Wireless|Wi-Fi|802.11|WLAN' -or $_.Name -match 'Wi-Fi|Wireless' } | "+action+"-NetAdapter -Confirm:$false")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Run()
	return true
}

// SetBluetoothEnabled toggles the Bluetooth radio on or off.
func SetBluetoothEnabled(enabled bool) bool {
	targetState := 2 // Off
	if enabled {
		targetState = 1 // On
	}
	return SetNativeRadioState(3, targetState)
}

// SetAirplaneMode toggles Airplane mode (all wireless radios).
func SetAirplaneMode(enabled bool) bool {
	targetState := 1 // Restore radios (Airplane mode off)
	if enabled {
		targetState = 2 // Disable all radios (Airplane mode on)
	}
	return SetNativeRadioState(255, targetState)
}
