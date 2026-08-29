import { writable } from 'svelte/store'
import { CheckForUpdates } from '../../wailsjs/go/backend/App.js'

export interface ReleaseInfo {
  version: string
  tagName: string
  name: string
  body: string
  publishedAt: string
  htmlUrl: string
  installerUrl: string
  portableUrl: string
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
}

export const devTestingStore = writable<boolean>(false)
export const devPinnedStore = writable<boolean>(false)
export const updateInfoStore = writable<UpdateCheckResult | null>(null)
export const updateCheckingStore = writable<boolean>(false)

export async function triggerUpdateCheck(
  force = true
): Promise<UpdateCheckResult | null> {
  updateCheckingStore.set(true)
  try {
    if (typeof CheckForUpdates === 'function') {
      const res = await CheckForUpdates(force)
      if (res) {
        updateInfoStore.set(res as UpdateCheckResult)
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
  } finally {
    updateCheckingStore.set(false)
  }
  return null
}
