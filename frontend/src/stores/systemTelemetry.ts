/**
 * @fileoverview Live CPU, RAM, and GPU performance metrics.
 *
 * Receives hardware load numbers from the Go backend and maintains rolling history
 * buffers for rendering sparkline charts on the performance page and home feed.
 */

import { writable } from 'svelte/store'
import { Events } from '@wailsio/runtime'
import {
  GetSystemTelemetry,
  StartSystemTelemetry,
  StopSystemTelemetry
} from '../../bindings/Vinta/backend/app.js'
import type { SystemTelemetry } from '../../bindings/Vinta/backend/services/models.js'

export interface TelemetryData {
  cpuPercent: number
  ramPercent: number
  ramUsedMB: number
  ramTotalMB: number
  gpuPercent: number
  timestampMs: number
}

const HISTORY_LIMIT = 24

export const systemTelemetryStore = writable<TelemetryData>({
  cpuPercent: 0,
  ramPercent: 0,
  ramUsedMB: 0,
  ramTotalMB: 0,
  gpuPercent: 0,
  timestampMs: 0
})

export const cpuHistoryStore = writable<number[]>(new Array(HISTORY_LIMIT).fill(0))
export const ramHistoryStore = writable<number[]>(new Array(HISTORY_LIMIT).fill(0))
export const gpuHistoryStore = writable<number[]>(new Array(HISTORY_LIMIT).fill(0))

let subscriberCount = 0
let unbindEvent: (() => void) | null = null

/**
 * Pushes a new reading into a fixed-length sparkline history array.
 *
 * @param store - The target writable store containing numeric values.
 * @param value - The latest scalar metric reading.
 */
function appendHistory(
  store: typeof cpuHistoryStore,
  value: number
) {
  store.update((arr) => {
    const next = [...arr.slice(1), Math.round(value)]
    return next
  })
}

/**
 * Updates all stores when a new hardware telemetry event arrives.
 *
 * @param data - Raw snapshot payload received over IPC.
 */
function handleTelemetryUpdate(data: Partial<TelemetryData>) {
  if (!data) return

  const reading: TelemetryData = {
    cpuPercent: Math.max(0, Math.min(100, Math.round(data.cpuPercent ?? 0))),
    ramPercent: Math.max(0, Math.min(100, Math.round(data.ramPercent ?? 0))),
    ramUsedMB: data.ramUsedMB ?? 0,
    ramTotalMB: data.ramTotalMB ?? 0,
    gpuPercent: Math.max(0, Math.min(100, Math.round(data.gpuPercent ?? 0))),
    timestampMs: data.timestampMs ?? Date.now()
  }

  systemTelemetryStore.set(reading)
  appendHistory(cpuHistoryStore, reading.cpuPercent)
  appendHistory(ramHistoryStore, reading.ramPercent)
  appendHistory(gpuHistoryStore, reading.gpuPercent)
}

/**
 * Subscribes a component to live system telemetry updates and starts backend polling.
 *
 * @returns Cleanup function that decrements subscription count and unsubscribes when idle.
 */
export function subscribeToTelemetry(): () => void {
  subscriberCount++

  if (subscriberCount === 1) {
    void StartSystemTelemetry()

    // Fetch initial reading immediately
    void GetSystemTelemetry()
      .then((initial: SystemTelemetry) => {
        if (initial) {
          handleTelemetryUpdate(initial)
        }
      })
      .catch(() => {})

    unbindEvent = Events.On('system:telemetry', (payload: any) => {
      const data = payload?.data?.[0] ?? payload?.data ?? payload
      handleTelemetryUpdate(data)
    })
  }

  return () => {
    subscriberCount = Math.max(0, subscriberCount - 1)
    if (subscriberCount === 0) {
      void StopSystemTelemetry()
      if (unbindEvent) {
        unbindEvent()
        unbindEvent = null
      }
    }
  }
}

