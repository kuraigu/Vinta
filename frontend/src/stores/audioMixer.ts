import { writable } from 'svelte/store'
import {
  GetAppAudioSessions,
  SetAppAudioVolume,
  SetAppAudioMute,
  GetAudioEndpoints,
  SetDefaultAudioEndpoint,
  SetMasterAudioVolume,
  ToggleMasterAudioMute
} from '../../bindings/Vinta/backend/app.js'
import type { AppAudioSession, AudioEndpoint } from '../../bindings/Vinta/backend/services/models.js'

export const audioSessionsStore = writable<AppAudioSession[]>([])
export const audioEndpointsStore = writable<AudioEndpoint[]>([])
export const masterVolumeStore = writable<number>(100)
export const masterMuteStore = writable<boolean>(false)

let pollTimer: ReturnType<typeof setInterval> | null = null
let pollSubscriberCount = 0

/**
 * Fetches the latest list of active application audio sessions.
 */
export async function refreshAudioSessions(): Promise<void> {
  if (typeof GetAppAudioSessions !== 'function') return
  try {
    const sessions = await GetAppAudioSessions()
    if (sessions) {
      audioSessionsStore.set(sessions)
    }
  } catch {
    // Ignore native bridge query errors
  }
}

/**
 * Fetches the list of active audio output endpoints.
 */
export async function refreshAudioEndpoints(): Promise<void> {
  if (typeof GetAudioEndpoints !== 'function') return
  try {
    const endpoints = await GetAudioEndpoints()
    if (endpoints) {
      audioEndpointsStore.set(endpoints)
    }
  } catch {
    // Ignore native bridge query errors
  }
}

/**
 * Adjusts the volume for a specific application session.
 */
export async function setAppVolume(pid: number, volume: number): Promise<void> {
  audioSessionsStore.update((list) =>
    list.map((item) => (item.pid === pid ? { ...item, volume } : item))
  )
  if (typeof SetAppAudioVolume === 'function') {
    try {
      await SetAppAudioVolume(pid, volume)
    } catch {
      // Ignore IPC error
    }
  }
}

/**
 * Toggles the mute state for a specific application session.
 */
export async function toggleAppMute(pid: number, isCurrentlyMuted: boolean): Promise<void> {
  const nextMute = !isCurrentlyMuted
  audioSessionsStore.update((list) =>
    list.map((item) => (item.pid === pid ? { ...item, isMuted: nextMute } : item))
  )
  if (typeof SetAppAudioMute === 'function') {
    try {
      await SetAppAudioMute(pid, nextMute)
    } catch {
      // Ignore IPC error
    }
  }
}

/**
 * Switches the system default playback endpoint.
 */
export async function switchDefaultEndpoint(endpointId: string): Promise<void> {
  audioEndpointsStore.update((list) =>
    list.map((ep) => ({
      ...ep,
      isDefault: ep.id === endpointId
    }))
  )
  if (typeof SetDefaultAudioEndpoint === 'function') {
    try {
      await SetDefaultAudioEndpoint(endpointId)
      setTimeout(refreshAudioEndpoints, 300)
    } catch {
      // Ignore IPC error
    }
  }
}

/**
 * Adjusts the master audio playback volume.
 */
export async function setMasterVolume(volume: number): Promise<void> {
  const clamped = Math.max(0, Math.min(100, volume))
  masterVolumeStore.set(clamped)
  if (typeof SetMasterAudioVolume === 'function') {
    try {
      await SetMasterAudioVolume(clamped)
    } catch {
      // Ignore IPC error
    }
  }
}

/**
 * Toggles master audio mute.
 */
export async function toggleMasterMute(): Promise<void> {
  masterMuteStore.update((m) => !m)
  if (typeof ToggleMasterAudioMute === 'function') {
    try {
      await ToggleMasterAudioMute()
    } catch {
      // Ignore IPC error
    }
  }
}

/**
 * Starts periodic polling when an audio mixer UI is opened.
 */
export function startMixerPolling(): () => void {
  pollSubscriberCount++
  if (pollSubscriberCount === 1) {
    void refreshAudioSessions()
    void refreshAudioEndpoints()
    pollTimer = setInterval(() => {
      void refreshAudioSessions()
    }, 1200)
  }

  return () => {
    pollSubscriberCount = Math.max(0, pollSubscriberCount - 1)
    if (pollSubscriberCount === 0 && pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }
}

