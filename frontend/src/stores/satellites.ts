import { writable, get } from 'svelte/store'
import {
  SetSatelliteBounds,
  RemoveSatellite
} from '../../bindings/Vinta/backend/app.js'
import { SatelliteBounds } from '../../bindings/Vinta/backend/models.js'

export type SatelliteAnchor = 'left' | 'right' | 'bottom' | 'top'

export interface SatelliteConfig {
  id: string
  anchor: SatelliteAnchor
  width: number
  height: number
  radius: number
  gap: number
  offsetX?: number
  offsetY?: number
  active: boolean
}

export const satellitesStore = writable<Record<string, SatelliteConfig>>({})

/**
 * Registers or updates a floating satellite dock in the local store
 * and synchronizes the region configuration with the Win32 backend.
 */
export function registerSatellite(config: SatelliteConfig) {
  satellitesStore.update((store) => {
    store[config.id] = { ...config }
    return store
  })
  syncSatelliteWithBackend(config)
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
      const footprint = sat.width + sat.gap + horizontalOffset
      if (footprint > maxFootprint) {
        maxFootprint = footprint
      }
    }
  }

  return maxFootprint
}

function syncSatelliteWithBackend(config: SatelliteConfig) {
  if (typeof SetSatelliteBounds === 'function') {
    const model = new SatelliteBounds({
      id: config.id,
      anchor: config.anchor,
      width: config.width,
      height: config.height,
      radius: config.radius,
      gap: config.gap,
      offsetX: config.offsetX || 0,
      offsetY: config.offsetY || 0,
      active: config.active
    })
    SetSatelliteBounds(model).catch(() => {})
  }
}
