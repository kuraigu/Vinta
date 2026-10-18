// Package backend provides core application services and state management for Vinta.
package backend

import (
	"fmt"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	globalHotkeyMutex sync.Mutex
	registeredHotkeys = make(map[string]string)
)

// RegisterNamedGlobalShortcut registers an application-wide hotkey accelerator for a specific action.
func (a *App) RegisterNamedGlobalShortcut(name string, accelerator string) (bool, error) {
	wailsApp := application.Get()
	if wailsApp == nil {
		return false, fmt.Errorf("application instance not available")
	}

	globalHotkeyMutex.Lock()
	defer globalHotkeyMutex.Unlock()

	trimmed := strings.TrimSpace(accelerator)
	oldAccel := registeredHotkeys[name]

	if trimmed == "" {
		if oldAccel != "" {
			_ = wailsApp.GlobalShortcut.Unregister(oldAccel)
			delete(registeredHotkeys, name)
		}
		return true, nil
	}

	if oldAccel != "" {
		_ = wailsApp.GlobalShortcut.Unregister(oldAccel)
		delete(registeredHotkeys, name)
	}

	eventName := "hotkey-toggle-island"
	switch name {
	case "quick_notes":
		eventName = "hotkey-quick-notes"
	case "mode_switch":
		eventName = "hotkey-mode-switch"
	default:
		eventName = "hotkey-toggle-island"
	}

	err := wailsApp.GlobalShortcut.Register(trimmed, func() {
		wailsApp.Event.Emit(eventName)
	})
	if err != nil {
		return false, err
	}

	registeredHotkeys[name] = trimmed
	return true, nil
}

// UnregisterNamedGlobalShortcut removes a registered global shortcut by action name.
func (a *App) UnregisterNamedGlobalShortcut(name string) error {
	wailsApp := application.Get()
	if wailsApp == nil {
		return fmt.Errorf("application instance not available")
	}

	globalHotkeyMutex.Lock()
	defer globalHotkeyMutex.Unlock()

	if oldAccel := registeredHotkeys[name]; oldAccel != "" {
		_ = wailsApp.GlobalShortcut.Unregister(oldAccel)
		delete(registeredHotkeys, name)
	}
	return nil
}

// GetNamedGlobalShortcut returns the currently registered accelerator string for a specific action.
func (a *App) GetNamedGlobalShortcut(name string) string {
	globalHotkeyMutex.Lock()
	defer globalHotkeyMutex.Unlock()
	return registeredHotkeys[name]
}

// RegisterGlobalShortcut registers the main toggle shortcut for backward compatibility.
func (a *App) RegisterGlobalShortcut(accelerator string) (bool, error) {
	return a.RegisterNamedGlobalShortcut("toggle", accelerator)
}

// UnregisterGlobalShortcut removes the main toggle shortcut for backward compatibility.
func (a *App) UnregisterGlobalShortcut() error {
	return a.UnregisterNamedGlobalShortcut("toggle")
}

// GetGlobalShortcut returns the main toggle shortcut accelerator string.
func (a *App) GetGlobalShortcut() string {
	return a.GetNamedGlobalShortcut("toggle")
}
