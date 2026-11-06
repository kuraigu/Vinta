/**
 * @fileoverview Stores app version info and manages background updates.
 *
 * Keeps track of version numbers and git commits, checks GitHub for new releases,
 * downloads installers, and triggers restarts when an update is ready.
 */

import { writable, get } from 'svelte/store'
import {
  CheckForUpdates,
  DownloadUpdate,
  PauseUpdateDownload,
  ResumeUpdateDownload,
  CancelUpdateDownload,
  ApplyUpdate,
  SimulateUpdateAvailable,
  SimulateDownloadCycle,
  GetAppInfo
} from '../../bindings/Vinta/backend/app.js'

export interface ReleaseInfo {
  version: string
  tagName: string
  name: string
  body: string
  publishedAt: string
  htmlUrl: string
  installerUrl: string
  checksumsUrl?: string
  isPrerelease: boolean
}

export interface UpdateCheckResult {
  status: 'up-to-date' | 'update-available' | 'dev-mode' | 'error' | string
  updateAvailable: boolean
  currentVersion: string
  latestVersion: string
  release: ReleaseInfo
  checkedAt: string
  message: string
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

export interface AppInfoData {
  name: string
  version: string
  gitCommit: string
  buildType: string
  buildTime: string
  isOfficialBuild: boolean
  goVersion: string
  arch: string
  os: string
  wailsVersion: string
  repoUrl: string
}

export const appInfoStore = writable<AppInfoData>({
  name: 'Vinta',
  version: '0.1.0',
  gitCommit: 'dev',
  buildType: 'Development',
  buildTime: '',
  isOfficialBuild: false,
  goVersion: 'go1.23',
  arch: 'amd64',
  os: 'windows',
  wailsVersion: 'v2.13.0',
  repoUrl: 'https://github.com/kuraigu/Vinta'
})

// Hydrate application metadata from Wails backend on initialization
if (typeof GetAppInfo === 'function') {
  GetAppInfo()
    .then((info) => {
      if (info) {
        appInfoStore.set(info as AppInfoData)
      }
    })
    .catch(() => {})
}

export const devTestingStore = writable<boolean>(false)
export const devPinnedStore = writable<boolean>(false)
export const updateInfoStore = writable<UpdateCheckResult | null>(null)
export const updateCheckingStore = writable<boolean>(false)

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
