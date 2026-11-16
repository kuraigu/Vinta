// Package backend provides core application services and state management for Vinta.
package backend

import (
	"time"
	"unsafe"
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
				shouldBeTransparent := !a.isForceInteractive && !a.isCurrentlyInteractive
				a.hitRectsVersion++
				a.pillHitMu.Unlock()
				if shouldBeTransparent {
					gwl := int32(GWL_EXSTYLE)
					style, _, _ := procGetWindowLong.Call(h, uintptr(gwl))
					if (style & WS_EX_TRANSPARENT) == 0 {
						style |= WS_EX_LAYERED | WS_EX_TRANSPARENT
						_, _, _ = procSetWindowLong.Call(h, uintptr(gwl), style)
					}
				}
				_, _, _ = procSetLayeredWindowAttrs.Call(h, 0, 255, 2)
			}
		})
	}
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
		gwl := int32(GWL_EXSTYLE)
		style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
		style &^= WS_EX_TRANSPARENT
		_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)
		_, _, _ = procSetLayeredWindowAttrs.Call(hwnd, 0, 255, 2)
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
	gwl := int32(GWL_EXSTYLE)
	style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
	if active {
		style |= WS_EX_LAYERED | WS_EX_TRANSPARENT
		_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)
		_, _, _ = procSetLayeredWindowAttrs.Call(hwnd, 0, 255, 2)
	} else {
		style &^= WS_EX_TRANSPARENT
		_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)
		_, _, _ = procSetLayeredWindowAttrs.Call(hwnd, 0, 255, 2)
	}
}

// startHitTestingLoop runs a 60Hz background loop to dynamically toggle WS_EX_TRANSPARENT
// based on cursor proximity to the active pill bounding box, active satellites, or force-interactive mode.
func (a *App) startHitTestingLoop(hwnd uintptr) {
	go func() {
		defer SetupPanicRecovery()
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()

		missCount := 0
		hitCount := 0
		forceOutsideCount := 0
		var rects []pillHitBounds
		var lastPt point32
		hasLastPt := false
		lastRectsVersion := ^uint64(0)
		verifiedStyle := false

		for range ticker.C {
			if hwnd == 0 {
				continue
			}
			isWin, _, _ := procIsWindow.Call(hwnd)
			if isWin == 0 {
				return
			}

			a.pillHitMu.RLock()
			outbound := a.isOutboundDragging
			force := a.isForceInteractive
			rectsVer := a.hitRectsVersion
			nRects := len(a.pillHitRects)
			if cap(rects) < nRects {
				rects = make([]pillHitBounds, nRects)
			} else {
				rects = rects[:nRects]
			}
			copy(rects, a.pillHitRects)
			a.pillHitMu.RUnlock()

			isHit := false
			rectsChanged := false
			if !outbound {
				cursorInAnyRect := false
				cursorHidden := false
				if len(rects) > 0 || force {
					const cursorShowing = 0x00000001
					var ci struct {
						cbSize      uint32
						flags       uint32
						hCursor     uintptr
						ptScreenPos point32
					}
					ci.cbSize = uint32(unsafe.Sizeof(ci))
					ret, _, _ := procGetCursorInfo.Call(uintptr(unsafe.Pointer(&ci)))
					if ret != 0 {
						if (ci.flags & cursorShowing) == 0 {
							// Hidden cursors (such as locked FPS game pointers) must never click Vinta.
							cursorHidden = true
							hasLastPt = false
						} else {
							pt := ci.ptScreenPos
							// Only short-circuit stationary cursors after the 8-frame miss grace window
							// or after force mode settles. Skipping earlier freezes missCount below 8
							// and leaves WS_EX_TRANSPARENT off until the mouse moves again.
							rectsChanged = rectsVer != lastRectsVersion
							if rectsChanged {
								verifiedStyle = false
							} else if hasLastPt && pt.X == lastPt.X && pt.Y == lastPt.Y && verifiedStyle && !force && (missCount >= 8 || hitCount > 4) {
								continue
							}
							lastPt = pt
							hasLastPt = true
							lastRectsVersion = rectsVer

							retScreen, _, _ := procScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&pt)))
							if retScreen != 0 {
								for _, rect := range rects {
									if pt.X >= rect.left && pt.X <= rect.right &&
										pt.Y >= rect.top && pt.Y <= rect.bottom {
										cursorInAnyRect = true
										break
									}
								}
							}
						}
					}
				}

				if cursorHidden {
					isHit = false
				} else if force {
					if cursorInAnyRect {
						forceOutsideCount = 0
						isHit = true
					} else {
						forceOutsideCount++
						// Auto-expire force interactive if cursor has been outside all rects for >1s and LMB is up
						if forceOutsideCount > 60 {
							if lmbState, _, _ := procGetAsyncKeyState.Call(0x01); (lmbState & 0x8000) == 0 {
								a.pillHitMu.Lock()
								a.isForceInteractive = false
								force = false
								a.pillHitMu.Unlock()
								forceOutsideCount = 0
							}
						}
						if force {
							isHit = true
						}
					}
				} else if cursorInAnyRect {
					isHit = true
				}
			}

			a.pillHitMu.RLock()
			currInteractive := a.isCurrentlyInteractive
			a.pillHitMu.RUnlock()

			if isHit {
				missCount = 0
				hitCount++
				if currInteractive && hitCount > 1 && verifiedStyle {
					continue
				}
			} else {
				hitCount = 0
				missCount++
				// Hold interactive state for an ~120ms grace window (8 frames at 60Hz)
				// only when multiple stationary hit rects (such as satellites) are active
				// so swift cursor motion across gaps does not drop clicks. When rects move
				// (gliding) or only a single rect is active, clear immediately so vacated
				// glide paths remain clickable across all island states.
				if len(rects) > 1 && !rectsChanged && missCount < 8 {
					continue
				}
				if !currInteractive && missCount > 8 && verifiedStyle {
					continue
				}
			}

			gwl := int32(GWL_EXSTYLE)
			style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwl))
			hasTransparent := (style & WS_EX_TRANSPARENT) != 0
			verifiedStyle = true

			needsUpdate := false
			if isHit {
				if hasTransparent {
					style &^= WS_EX_TRANSPARENT
					needsUpdate = true
				}
			} else if !hasTransparent {
				// Never switch to WS_EX_TRANSPARENT while the left mouse button is held down during an inbound drag,
				// because changing window styles mid-drag causes OLE to lose DragLeave/Drop synchronization and lock up File Explorer.
				if !outbound {
					if lmbState, _, _ := procGetAsyncKeyState.Call(0x01); (lmbState & 0x8000) != 0 {
						continue
					}
				}
				style |= WS_EX_TRANSPARENT
				needsUpdate = true
			}

			a.pillHitMu.Lock()
			a.isCurrentlyInteractive = isHit
			a.pillHitMu.Unlock()
			if needsUpdate {
				_, _, _ = procSetWindowLong.Call(hwnd, uintptr(gwl), style)
				if isHit {
					_, _, _ = procSetLayeredWindowAttrs.Call(hwnd, 0, 255, 2)
				}
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
