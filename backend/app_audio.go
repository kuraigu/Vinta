// Package backend provides core application services and state management for Vinta.
package backend

import (
	"Vinta/backend/services"
)

// GetAppAudioSessions returns active audio application sessions.
func (a *App) GetAppAudioSessions() []services.AppAudioSession {
	sessions, _ := services.GetNativeAudioSessions()
	if sessions == nil {
		return []services.AppAudioSession{}
	}
	return sessions
}

// SetAppAudioVolume adjusts the volume for an application by its process ID.
func (a *App) SetAppAudioVolume(pid uint32, volume float64) bool {
	return services.SetNativeSessionVolume(pid, float32(volume))
}

// SetAppAudioMute toggles mute state for an application by its process ID.
func (a *App) SetAppAudioMute(pid uint32, mute bool) bool {
	return services.SetNativeSessionMute(pid, mute)
}

// GetAudioEndpoints returns active audio output devices.
func (a *App) GetAudioEndpoints() []services.AudioEndpoint {
	endpoints, _ := services.GetNativeAudioEndpoints()
	if endpoints == nil {
		return []services.AudioEndpoint{}
	}
	return endpoints
}

// SetDefaultAudioEndpoint switches the active default playback device.
func (a *App) SetDefaultAudioEndpoint(endpointID string) bool {
	return services.SetNativeDefaultAudioEndpoint(endpointID)
}

// SetMasterAudioVolume sets the system master volume level (0-100).
func (a *App) SetMasterAudioVolume(percent float64) bool {
	return services.SetNativeMasterVolume(float32(percent))
}

// ToggleMasterAudioMute toggles the master audio mute state.
func (a *App) ToggleMasterAudioMute() bool {
	return services.ToggleNativeMasterMute()
}

