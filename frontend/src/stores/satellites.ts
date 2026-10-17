import { writable, get } from 'svelte/store'
import {
  SetSatelliteBounds,
  RemoveSatellite
} from '../../bindings/Vinta/backend/app.js'
import { SatelliteBounds } from '../../bindings/Vinta/backend/models.js'

export type SatelliteAnchor = 'left' | 'right' | 'bottom' | 'top'
export type TooltipSide = 'bottom' | 'right' | 'left' | 'top'
export type FourPartRadii = [number, number, number, number]
export type FourPartGaps = [number, number, number, number]

export interface SatelliteConfig {
  id: string
  anchor: SatelliteAnchor
  x?: number
  y?: number
  width: number
  height: number
  radius: number | FourPartRadii
  gap: number | FourPartGaps
  offsetX?: number
  offsetY?: number
  tooltipPad?: number
  tooltipSide?: TooltipSide
  active: boolean
}

export const satellitesStore = writable<Record<string, SatelliteConfig>>({})

/**
 * Checks whether two satellite configurations are geometrically and functionally identical.
 */
function isConfigEqual(a: SatelliteConfig, b: SatelliteConfig): boolean {
  if (
    a.id !== b.id ||
    a.anchor !== b.anchor ||
    a.active !== b.active ||
    a.width !== b.width ||
    a.height !== b.height ||
    a.x !== b.x ||
    a.y !== b.y ||
    a.offsetX !== b.offsetX ||
    a.offsetY !== b.offsetY ||
    a.tooltipPad !== b.tooltipPad ||
    a.tooltipSide !== b.tooltipSide
  ) {
    return false
  }

  const radA = JSON.stringify(a.radius)
  const radB = JSON.stringify(b.radius)
  if (radA !== radB) return false

  const gapA = JSON.stringify(a.gap)
  const gapB = JSON.stringify(b.gap)
  if (gapA !== gapB) return false

  return true
}

/**
 * Registers or updates a floating satellite dock in the local store
 * and synchronizes the region configuration with the Win32 backend.
 */
export function registerSatellite(config: SatelliteConfig) {
  let isChanged = false
  satellitesStore.update((store) => {
    const existing = store[config.id]
    if (existing && isConfigEqual(existing, config)) {
      return store
    }
    isChanged = true
    store[config.id] = { ...config }
    return store
  })
  if (isChanged) {
    syncSatelliteWithBackend(config)
  }
}

/**
 * Updates properties of an existing satellite dock.
 */
export function updateSatellite(
  id: string,
  updates: Partial<Omit<SatelliteConfig, 'id'>>
) {
  let updatedConfig: SatelliteConfig | null = null
  satellitesStore.update((store) => {
    if (store[id]) {
      store[id] = { ...store[id], ...updates }
      updatedConfig = store[id]
    }
    return store
  })
  if (updatedConfig) {
    syncSatelliteWithBackend(updatedConfig)
  }
}

/**
 * Adjusts the tooltip padding on a satellite's Win32 hit-test region.
 * Call this when a custom tooltip becomes visible or hides so the OS region
 * expands just enough to cover the tooltip without bleeding onto the island.
 */
export function setSatelliteTooltipPadding(
  id: string,
  pad: number,
  side: TooltipSide
) {
  updateSatellite(id, { tooltipPad: pad, tooltipSide: side })
}

/**
 * Unregisters a satellite dock and disengages its Win32 hit-test region.
 */
export function unregisterSatellite(id: string) {
  satellitesStore.update((store) => {
    delete store[id]
    return store
  })
  if (typeof RemoveSatellite === 'function') {
    RemoveSatellite(id).catch(() => {})
  }
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('vinta-resync-bounds'))
  }
}

/**
 * Computes the maximum horizontal footprint (width plus gap and outer offset)
 * for all currently active satellites anchored to the specified flank.
 * This is used for dynamic screen boundary clamping.
 */
export function getActiveSatelliteFootprint(
  anchor: 'left' | 'right',
  store: Record<string, SatelliteConfig> = get(satellitesStore)
): number {
  let maxFootprint = 0

  for (const key of Object.keys(store)) {
    const sat = store[key]
    if (sat && sat.active && sat.anchor === anchor && sat.width > 0) {
      const horizontalOffset = Math.max(0, sat.offsetX || 0)
      const gapVal = Array.isArray(sat.gap)
        ? (anchor === 'left' ? sat.gap[1] : sat.gap[3])
        : sat.gap
      const footprint = sat.width + gapVal + horizontalOffset
      if (footprint > maxFootprint) {
        maxFootprint = footprint
      }
    }
  }

  return maxFootprint
}

/**
 * Forcibly resynchronizes all active satellites with the native Win32 backend.
 * Clears desynchronizations or missed IPC calls after drag or lag events.
 */
export function resyncAllSatellites() {
  const store = get(satellitesStore)
  for (const id of Object.keys(store)) {
    const config = store[id]
    if (config && config.active) {
      syncSatelliteWithBackend(config)
    }
  }
}

function syncSatelliteWithBackend(config: SatelliteConfig) {
  if (typeof SetSatelliteBounds === 'function') {
    const singleRadius = typeof config.radius === 'number' ? config.radius : 0
    const radiiArray = Array.isArray(config.radius)
      ? config.radius
      : [config.radius, config.radius, config.radius, config.radius]

    const singleGap = typeof config.gap === 'number' ? config.gap : 0
    const gapsArray = Array.isArray(config.gap)
      ? config.gap
      : [config.gap, config.gap, config.gap, config.gap]

    const model = new SatelliteBounds({
      id: config.id,
      anchor: config.anchor,
      x: config.x !== undefined ? Math.round(config.x) : 0,
      y: config.y !== undefined ? Math.round(config.y) : 0,
      width: config.width,
      height: config.height,
      radius: singleRadius,
      radii: radiiArray,
      gap: singleGap,
      gaps: gapsArray,
      offsetX: config.offsetX || 0,
      offsetY: config.offsetY || 0,
      tooltipPad: config.tooltipPad || 0,
      tooltipSide: config.tooltipSide || '',
      active: config.active
    } as unknown as Partial<SatelliteBounds>)
    SetSatelliteBounds(model).catch(() => {})
  }
}
