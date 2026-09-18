import { writable, get } from 'svelte/store'
import {
  CheckForUpdates,
  DownloadUpdate,
  PauseUpdateDownload,
  ResumeUpdateDownload,
  CancelUpdateDownload,
  ApplyUpdate,
  IsPortableMode,
  SimulateUpdateAvailable,
  SimulateDownloadCycle
} from '../../bindings/Vinta/backend/app.js'

export interface ReleaseInfo {
  version: string
  tagName: string
  name: string
  body: string
  publishedAt: string
  htmlUrl: string
  installerUrl: string
  portableUrl: string
  checksumsUrl?: string
  isPrerelease: boolean
  isPortable?: boolean
}

export interface UpdateCheckResult {
  status: 'up-to-date' | 'update-available' | 'dev-mode' | 'error' | string
  updateAvailable: boolean
  currentVersion: string
  latestVersion: string
  release: ReleaseInfo
  checkedAt: string
  message: string
  isPortable?: boolean
  isCustomRepo?: boolean
}

export interface DownloadProgress {
  status: 'downloading' | 'paused' | 'verifying' | 'ready' | 'error' | 'idle'
  downloadedBytes: number
  totalBytes: number
  percentage: number
  speedBytesPerSec: number
  error?: string
}

export type UpdateLifecycleState =
  | 'idle'
  | 'checking'
  | 'available'
  | 'downloading'
  | 'paused'
  | 'verifying'
  | 'ready'
  | 'error'

export const devTestingStore = writable<boolean>(false)
export const devPinnedStore = writable<boolean>(false)
export const updateInfoStore = writable<UpdateCheckResult | null>(null)
export const updateCheckingStore = writable<boolean>(false)
export const isPortableModeStore = writable<boolean>(false)

export const downloadProgressStore = writable<DownloadProgress>({
  status: 'idle',
  downloadedBytes: 0,
  totalBytes: 0,
  percentage: 0,
  speedBytesPerSec: 0
})

export const updateLifecycleStatusStore =
  writable<UpdateLifecycleState>('idle')

/**
 * Initializes runtime portable mode detection.
 */
export async function detectPortableMode(): Promise<boolean> {
  try {
    if (typeof IsPortableMode === 'function') {
      const isPortable = await IsPortableMode()
      isPortableModeStore.set(Boolean(isPortable))
      return Boolean(isPortable)
    }
  } catch {
    isPortableModeStore.set(false)
  }
  return false
}

/**
 * Initiates an update check.
 */
export async function triggerUpdateCheck(
  force = true
): Promise<UpdateCheckResult | null> {
  updateCheckingStore.set(true)
  try {
    if (typeof CheckForUpdates === 'function') {
      const res = await CheckForUpdates(force)
      if (res) {
        updateInfoStore.set(res as UpdateCheckResult)
        if (res.updateAvailable) {
          updateLifecycleStatusStore.set('available')
        } else {
          updateLifecycleStatusStore.set('idle')
        }
        return res as UpdateCheckResult
      }
    }
  } catch (err) {
    updateInfoStore.set({
      status: 'error',
      updateAvailable: false,
      currentVersion: '',
      latestVersion: '',
      release: {} as ReleaseInfo,
      checkedAt: new Date().toLocaleTimeString(),
      message: String(err)
    })
    updateLifecycleStatusStore.set('error')
  } finally {
    updateCheckingStore.set(false)
  }
  return null
}

/**
 * Starts or restarts downloading an update.
 */
export async function triggerDownload(
  downloadUrl: string,
  checksumsUrl = ''
): Promise<void> {
  if (!downloadUrl) return
  updateLifecycleStatusStore.set('downloading')
  downloadProgressStore.update((p) => ({
    ...p,
    status: 'downloading',
    error: undefined
  }))

  try {
    if (typeof DownloadUpdate === 'function') {
      await DownloadUpdate(downloadUrl, checksumsUrl)
    }
  } catch (err) {
    downloadProgressStore.update((p) => ({
      ...p,
      status: 'error',
      error: String(err)
    }))
    updateLifecycleStatusStore.set('error')
  }
}

/**
 * Pauses an active download stream.
 */
export async function triggerPauseDownload(): Promise<void> {
  try {
    if (typeof PauseUpdateDownload === 'function') {
      await PauseUpdateDownload()
    }
    updateLifecycleStatusStore.set('paused')
    downloadProgressStore.update((p) => ({ ...p, status: 'paused' }))
  } catch (err) {
    console.error('Failed to pause download:', err)
  }
}

/**
 * Resumes a paused download stream.
 */
export async function triggerResumeDownload(): Promise<void> {
  try {
    updateLifecycleStatusStore.set('downloading')
    downloadProgressStore.update((p) => ({ ...p, status: 'downloading' }))
    if (typeof ResumeUpdateDownload === 'function') {
      await ResumeUpdateDownload()
    }
  } catch (err) {
    console.error('Failed to resume download:', err)
  }
}

/**
 * Cancels download and clears temporary files.
 */
export async function triggerCancelDownload(): Promise<void> {
  try {
    if (typeof CancelUpdateDownload === 'function') {
      await CancelUpdateDownload()
    }
  } catch (err) {
    console.error('Failed to cancel download:', err)
  } finally {
    updateLifecycleStatusStore.set(
      get(updateInfoStore)?.updateAvailable ? 'available' : 'idle'
    )
    downloadProgressStore.set({
      status: 'idle',
      downloadedBytes: 0,
      totalBytes: 0,
      percentage: 0,
      speedBytesPerSec: 0
    })
  }
}

/**
 * Applies the validated update and restarts the application.
 */
export async function triggerApplyUpdate(): Promise<void> {
  if (typeof ApplyUpdate === 'function') {
    await ApplyUpdate()
  }
}

/**
 * Triggers dev mock update events.
 */
export async function triggerDevSimulation(
  version = 'v0.4.0-alpha',
  notes = ''
): Promise<void> {
  if (typeof SimulateUpdateAvailable === 'function') {
    await SimulateUpdateAvailable(version, notes)
  }
}

/**
 * Triggers dev download simulation.
 */
export async function triggerDevDownloadCycle(
  failChecksum = false
): Promise<void> {
  updateLifecycleStatusStore.set('downloading')
  if (typeof SimulateDownloadCycle === 'function') {
    await SimulateDownloadCycle(failChecksum)
  }
}
