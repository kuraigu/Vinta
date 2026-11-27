// Package backend provides core application services and state management for Vinta.
package backend

import (
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"

	"Vinta/backend/services"
)

// HitRect defines a physical client rectangle for hit testing.
type HitRect struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type pillHitBounds struct {
	left, top, right, bottom int32
}

// TooltipBounds holds the screen-relative bounding box and corner radius for an active tooltip.
type TooltipBounds struct {
	ID     string  `json:"id"`
	X      int     `json:"x"`
	Y      int     `json:"y"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Radius int     `json:"radius"`
	Scale  float64 `json:"scale"`
	Active bool    `json:"active"`
}

// SatelliteBounds represents the position, dimensions, radius, and anchor configuration for a floating satellite dock.
type SatelliteBounds struct {
	ID          string `json:"id"`
	Anchor      string `json:"anchor"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Radius      int    `json:"radius"`
	Radii       [4]int `json:"radii,omitempty"`
	Gap         int    `json:"gap"`
	Gaps        [4]int `json:"gaps,omitempty"`
	OffsetX     int    `json:"offsetX"`
	OffsetY     int    `json:"offsetY"`
	TooltipPad  int    `json:"tooltipPad"`
	TooltipSide string `json:"tooltipSide"`
	Active      bool   `json:"active"`
}

// SetHitRects updates the list of physical pixel rectangles that receive mouse input.
func (a *App) SetHitRects(rects []HitRect) {
	nextRects := make([]pillHitBounds, 0, len(rects))
	for _, r := range rects {
		if r.Width > 0 && r.Height > 0 {
			nextRects = append(nextRects, pillHitBounds{
				left:   int32(r.Left),
				top:    int32(r.Top),
				right:  int32(r.Left + r.Width),
				bottom: int32(r.Top + r.Height),
			})
		}
	}

	a.pillHitMu.Lock()
	changed := len(nextRects) != len(a.pillHitRects)
	if !changed {
		for i := range nextRects {
			if nextRects[i] != a.pillHitRects[i] {
				changed = true
				break
			}
		}
	}
	if changed {
		a.pillHitRects = nextRects
		a.hitRectsVersion++
	}
	a.pillHitMu.Unlock()

	if changed {
		activeIslandMu.Lock()
		if activeIslandHwnd != 0 {
			activeIslandRects = make([]pillHitBounds, len(nextRects))
			copy(activeIslandRects, nextRects)
		}
		activeIslandMu.Unlock()

		if a.getHwnd() != 0 {
			a.requestHitTestReconcile()
		}
	}

	if len(rects) > 0 {
		var maxBottom int32
		for _, r := range nextRects {
			if r.bottom > maxBottom {
				maxBottom = r.bottom
			}
		}
		if maxBottom > 0 {
			a.pillHitMu.RLock()
			currHeight := a.overlayHeight
			a.pillHitMu.RUnlock()
			if currHeight > 0 && int(maxBottom)+16 > currHeight {
				go a.ensureOverlayHeight(int(maxBottom) + 16)
			}
		}

		a.initialShowOnce.Do(func() {
			if h := a.getHwnd(); h != 0 {
				_, _, _ = procSetLayeredWindowAttrs.Call(h, 0, 255, 2)
			}
			if a.window != nil {
				a.window.Show()
			}
			if h := a.getHwnd(); h != 0 {
				a.pillHitMu.Lock()
				a.hitRectsVersion++
				a.pillHitMu.Unlock()
				_, _, _ = procSetLayeredWindowAttrs.Call(h, 0, 255, 2)
			}
		})
	}
}

// setClickThrough is the only place that mutates WS_EX_TRANSPARENT. It reads the live style
// so cached state can never disagree with the window, and reports whether it changed anything.
func (a *App) setClickThrough(hwnd uintptr, transparent bool) bool {
	gwl := int32(GWL_EXSTYLE)
	style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
	if style == 0 {
		return false
	}
	has := (style & WS_EX_TRANSPARENT) != 0
	if has == transparent {
		return false
	}
	if transparent {
		style |= WS_EX_LAYERED | WS_EX_TRANSPARENT
	} else {
		style &^= WS_EX_TRANSPARENT
	}
	_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)
	return true
}

// SetPillBounds updates the physical pixel bounding box of the active island for hit testing.
func (a *App) SetPillBounds(left, top, width, height int) {
	if width <= 0 || height <= 0 {
		a.SetHitRects(nil)
	} else {
		a.SetHitRects([]HitRect{{Left: left, Top: top, Width: width, Height: height}})
	}
}

// UpdateIslandRect updates the physical pixel bounding box of the active island directly from live CSS client geometry.
func (a *App) UpdateIslandRect(x, y, w, h float64) {
	if w <= 0 || h <= 0 {
		a.SetHitRects(nil)
	} else {
		a.SetHitRects([]HitRect{{Left: int(x), Top: int(y), Width: int(w), Height: int(h)}})
	}
}

// SetForceInteractive forces the window to be interactive during dragging or modal flows.
func (a *App) SetForceInteractive(active bool) {
	a.pillHitMu.Lock()
	a.isForceInteractive = active
	a.hitRectsVersion++
	outbound := a.isOutboundDragging
	if !outbound {
		a.isCurrentlyInteractive = active
	}
	a.pillHitMu.Unlock()

	if outbound {
		return
	}
	hwnd := a.getHwnd()
	if hwnd == 0 {
		return
	}
	if active {
		a.setClickThrough(hwnd, false)
	}
}

// IsOutboundDragging reports whether Vinta is actively performing an outbound OLE file drag.
func (a *App) IsOutboundDragging() bool {
	a.pillHitMu.RLock()
	defer a.pillHitMu.RUnlock()
	return a.isOutboundDragging
}

// setOutboundDragging toggles click-through transparency on the Vinta window during outbound OLE file drags.
func (a *App) setOutboundDragging(active bool) {
	a.pillHitMu.Lock()
	a.isOutboundDragging = active
	a.isCurrentlyInteractive = !active
	a.hitRectsVersion++
	a.pillHitMu.Unlock()

	hwnd := a.getHwnd()
	if hwnd == 0 {
		return
	}
	a.setClickThrough(hwnd, active)
}

// requestHitTestReconcile signals the hit-test worker goroutine without blocking the caller.
// Safe to call from high-frequency WM_INPUT handlers (1000Hz-8000Hz gaming mice) or WebView2 IPC.
func (a *App) requestHitTestReconcile() {
	if a.hitTestSignal == nil || services.IsAppHidden() {
		return
	}
	select {
	case a.hitTestSignal <- struct{}{}:
	default:
	}
}

// reconcileHitTest evaluates the live cursor position against active hit rects and updates
// WS_EX_TRANSPARENT immediately. Must run on the hit-test worker goroutine, never inside wndProc.
func (a *App) reconcileHitTest(hwnd uintptr) {
	if hwnd == 0 || services.IsAppHidden() {
		return
	}
	a.hitTestEvalMu.Lock()
	defer a.hitTestEvalMu.Unlock()

	const (
		cursorShowing    = 0x00000001
		cursorSuppressed = 0x00000002
	)

	a.pillHitMu.RLock()
	outbound := a.isOutboundDragging
	force := a.isForceInteractive
	rectsVer := a.hitRectsVersion
	nRects := len(a.pillHitRects)
	if cap(a.hitTestRectsScratch) < nRects {
		a.hitTestRectsScratch = make([]pillHitBounds, nRects)
	} else {
		a.hitTestRectsScratch = a.hitTestRectsScratch[:nRects]
	}
	copy(a.hitTestRectsScratch, a.pillHitRects)
	a.pillHitMu.RUnlock()

	rects := a.hitTestRectsScratch
	rectsChanged := rectsVer != a.hitTestLastRectsVer
	a.hitTestLastRectsVer = rectsVer

	isHit := false
	cursorInAnyRect := false
	cursorHidden := false
	now := time.Now()

	if !outbound {
		if len(rects) > 0 || force {
			var ci struct {
				cbSize      uint32
				flags       uint32
				hCursor     uintptr
				ptScreenPos point32
			}
			ci.cbSize = uint32(unsafe.Sizeof(ci))
			ret, _, _ := procGetCursorInfo.Call(uintptr(unsafe.Pointer(&ci)))
			if ret != 0 {
				pt := ci.ptScreenPos
				if retScreen, _, _ := procScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&pt))); retScreen != 0 {
					for _, rect := range rects {
						if pt.X >= rect.left && pt.X <= rect.right &&
							pt.Y >= rect.top && pt.Y <= rect.bottom {
							cursorInAnyRect = true
							break
						}
					}
				}

				// When the OS cursor is visible (or touch/pen suppressed), interact immediately.
				// When the OS cursor is hidden (ci.flags == 0), distinguish locked FPS mouselook
				// from games that render a custom software cursor in menus (e.g. Valorant):
				// require active cursor movement, no tiny ClipCursor lock box, and 60ms continuous
				// presence inside a hit rect so brief 1-frame FPS flicks never trigger clicks.
				if (ci.flags & (cursorShowing | cursorSuppressed)) != 0 {
					a.hitTestHiddenHitStart = time.Time{}
					a.hitTestHiddenInit = false
					a.hitTestHiddenMoved = false
				} else {
					if !a.hitTestHiddenInit {
						a.hitTestHiddenLastPt = ci.ptScreenPos
						a.hitTestHiddenInit = true
					} else if ci.ptScreenPos != a.hitTestHiddenLastPt {
						a.hitTestHiddenMoved = true
						a.hitTestHiddenLastPt = ci.ptScreenPos
					}

					allowSoftwareCursor := false
					if cursorInAnyRect && a.hitTestHiddenMoved {
						microClipped := false
						var clipRc rect32
						if clipRet, _, _ := procGetClipCursor.Call(uintptr(unsafe.Pointer(&clipRc))); clipRet != 0 {
							if (clipRc.Right-clipRc.Left) <= 16 && (clipRc.Bottom-clipRc.Top) <= 16 {
								microClipped = true
							}
						}
						if microClipped {
							a.hitTestHiddenHitStart = time.Time{}
						} else {
							if a.hitTestHiddenHitStart.IsZero() {
								a.hitTestHiddenHitStart = now
							}
							if now.Sub(a.hitTestHiddenHitStart) >= 60*time.Millisecond {
								allowSoftwareCursor = true
							}
						}
					} else {
						a.hitTestHiddenHitStart = time.Time{}
					}

					if !allowSoftwareCursor {
						cursorHidden = true
					}
				}
			}
		}

		switch {
		case cursorHidden:
			isHit = false
		case force:
			if cursorInAnyRect {
				a.hitTestForceOutStart = time.Time{}
			} else {
				if a.hitTestForceOutStart.IsZero() {
					a.hitTestForceOutStart = now
				}
				// Auto-expire force interactive if the cursor has been outside all rects for >1s and LMB is up.
				if now.Sub(a.hitTestForceOutStart) > time.Second {
					if lmbState, _, _ := procGetAsyncKeyState.Call(0x01); (lmbState & 0x8000) == 0 {
						a.pillHitMu.Lock()
						a.isForceInteractive = false
						a.pillHitMu.Unlock()
						force = false
						a.hitTestForceOutStart = time.Time{}
					}
				}
			}
			isHit = force || cursorInAnyRect
		default:
			isHit = cursorInAnyRect
		}
	}

	if isHit {
		a.hitTestMissStart = time.Time{}
	} else {
		if a.hitTestMissStart.IsZero() {
			a.hitTestMissStart = now
		}
		// Hold interactive state for 120ms when multiple stationary rects (satellites)
		// are active so swift motion across gaps does not drop clicks.
		if len(rects) > 1 && !rectsChanged && now.Sub(a.hitTestMissStart) < 120*time.Millisecond {
			return
		}
		// Never switch to WS_EX_TRANSPARENT while the left mouse button is held during an
		// inbound drag: changing styles mid-drag desyncs OLE DragLeave/Drop.
		if !outbound {
			if lmbState, _, _ := procGetAsyncKeyState.Call(0x01); (lmbState & 0x8000) != 0 {
				gwl := int32(GWL_EXSTYLE)
				style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
				if style != 0 && (style&WS_EX_TRANSPARENT) == 0 {
					return
				}
			}
		}
	}

	a.setClickThrough(hwnd, !isHit)

	if !outbound && !cursorHidden && len(rects) > 0 {
		lmbState, _, _ := procGetAsyncKeyState.Call(0x01)
		rmbState, _, _ := procGetAsyncKeyState.Call(0x02)
		isButtonDown := (lmbState&0x8000) != 0 || (rmbState&0x8000) != 0
		if isButtonDown && !a.hitTestWasButtonDown && !force && !cursorInAnyRect {
			if wailsApp := application.Get(); wailsApp != nil {
				go wailsApp.Event.Emit("vinta:click-outside")
			}
		}
		a.hitTestWasButtonDown = isButtonDown
	} else {
		a.hitTestWasButtonDown = false
	}

	a.pillHitMu.Lock()
	a.isCurrentlyInteractive = isHit
	a.pillHitMu.Unlock()
}

// startHitTestingLoop runs the dedicated hit-test worker goroutine that reconciles WS_EX_TRANSPARENT
// on coalesced WM_INPUT / SetHitRects signals (capped at ~120Hz so 1000Hz-8000Hz gaming mice never
// flood Win32 syscalls or block the message loop) and backs them up with a 150ms housekeeping tick
// (falling back to 16ms polling if Raw Input registration is unavailable).
func (a *App) startHitTestingLoop(hwnd uintptr) {
	go func() {
		defer SetupPanicRecovery()
		const minSignalInterval = 8 * time.Millisecond
		interval := 16 * time.Millisecond
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		var (
			lastReconcile time.Time
			coalesceTimer *time.Timer
			coalesceC     <-chan time.Time
		)
		defer func() {
			if coalesceTimer != nil {
				coalesceTimer.Stop()
			}
		}()

		runReconcile := func() {
			if services.IsAppHidden() {
				return
			}
			if hwnd == 0 {
				hwnd = a.getHwnd()
				if hwnd == 0 {
					return
				}
			}
			if isWin, _, _ := procIsWindow.Call(hwnd); isWin == 0 {
				hwnd = a.getHwnd()
				if hwnd == 0 {
					return
				}
			}
			lastReconcile = time.Now()
			a.reconcileHitTest(hwnd)
		}

		stopCtx := a.lifecycleContext()
		for {
			select {
			case <-stopCtx.Done():
				return
			case <-a.hitTestSignal:
				if services.IsAppHidden() {
					continue
				}
				elapsed := time.Since(lastReconcile)
				if elapsed >= minSignalInterval {
					if coalesceTimer != nil {
						coalesceTimer.Stop()
						coalesceC = nil
					}
					runReconcile()
				} else if coalesceC == nil {
					wait := minSignalInterval - elapsed
					if coalesceTimer == nil {
						coalesceTimer = time.NewTimer(wait)
					} else {
						coalesceTimer.Reset(wait)
					}
					coalesceC = coalesceTimer.C
				}
			case <-coalesceC:
				coalesceC = nil
				runReconcile()
			case <-ticker.C:
				a.pillHitMu.RLock()
				rawActive := a.rawInputRegistered
				a.pillHitMu.RUnlock()
				desiredInterval := 16 * time.Millisecond
				if rawActive {
					desiredInterval = 150 * time.Millisecond
				}
				if desiredInterval != interval {
					interval = desiredInterval
					ticker.Reset(interval)
				}
				runReconcile()
			}
		}
	}()
}

// SetSatelliteBounds registers or updates a floating satellite dock in the window region manager.
func (a *App) SetSatelliteBounds(bounds SatelliteBounds) {
	a.satellitesMu.Lock()
	if a.satellites == nil {
		a.satellites = make(map[string]SatelliteBounds)
	}
	a.satellites[bounds.ID] = bounds
	a.satellitesMu.Unlock()
	a.updateWindowRegion()
}

// RemoveSatellite removes a floating satellite dock by identifier.
func (a *App) RemoveSatellite(id string) {
	a.satellitesMu.Lock()
	if a.satellites != nil {
		delete(a.satellites, id)
	}
	a.satellitesMu.Unlock()
	a.updateWindowRegion()
}

// SetFloatingControlsBounds adapts legacy calls to the generalized satellite dock registry.
func (a *App) SetFloatingControlsBounds(width int, height int, radius int, gap int) {
	a.SetSatelliteBounds(SatelliteBounds{
		ID:     "controls",
		Anchor: "right",
		Width:  width,
		Height: height,
		Radius: radius,
		Gap:    gap,
		Active: width > 0 && height > 0,
	})
}

// SetTooltipBounds registers an active floating tooltip region.
func (a *App) SetTooltipBounds(bounds TooltipBounds) {
	a.tooltipsMu.Lock()
	if a.tooltips == nil {
		a.tooltips = make(map[string]TooltipBounds)
	}
	if bounds.Active && bounds.Width > 0 && bounds.Height > 0 {
		a.tooltips[bounds.ID] = bounds
	} else {
		delete(a.tooltips, bounds.ID)
	}
	a.tooltipsMu.Unlock()
	a.updateWindowRegion()
}

// ClearTooltipBounds removes an active floating tooltip region.
func (a *App) ClearTooltipBounds(id string) {
	a.tooltipsMu.Lock()
	if a.tooltips != nil {
		if id == "*" || id == "all" || id == "" {
			a.tooltips = make(map[string]TooltipBounds)
		} else {
			delete(a.tooltips, id)
		}
	}
	a.tooltipsMu.Unlock()
	a.updateWindowRegion()
}

// ClearAllTooltips removes all active floating tooltip regions.
func (a *App) ClearAllTooltips() {
	a.tooltipsMu.Lock()
	a.tooltips = make(map[string]TooltipBounds)
	a.tooltipsMu.Unlock()
	a.updateWindowRegion()
}
