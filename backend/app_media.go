// Package backend provides core application services and state management for Vinta.
package backend

import (
	"Vinta/backend/services"
)

// SendMediaControl forwards playback commands to the native media services bridge.
func (a *App) SendMediaControl(action string) {
	services.SendMediaControl(action)
}

// SelectMediaSession asks the native media service to switch active playback to sessionID.
func (a *App) SelectMediaSession(sessionID string) bool {
	return services.SelectMediaSession(sessionID)
}

// RequestMediaUpdate asks the backend and native WinRT bridge to emit a fresh media state snapshot to the frontend.
func (a *App) RequestMediaUpdate() {
	services.RequestMediaUpdate()
}

// RequeryMediaArt requests the native WinRT service to re-extract the thumbnail stream for the current track.
func (a *App) RequeryMediaArt() {
	services.RequeryMediaArt()
}

// FocusMediaApp brings the top-level window of the specified media application to the foreground.
func (a *App) FocusMediaApp(sourceApp string) bool {
	return services.FocusMediaApp(sourceApp)
}

// FocusNotificationApp brings the source window of a notification to the foreground.
func (a *App) FocusNotificationApp(sourceApp string, aumid string) bool {
	return services.FocusNotificationApp(sourceApp, aumid)
}
