import HomePage, { config as homeConfig } from '../pages/HomePage.svelte'
import AppsPage, { config as appsConfig } from '../pages/AppsPage.svelte'
import SettingsPage, {
  config as settingsConfig
} from '../pages/SettingsPage.svelte'
import MediaPage, { config as mediaConfig } from '../pages/MediaPage.svelte'
import NotificationPage, {
  config as notifConfig
} from '../pages/NotificationPage.svelte'
import { globalMediaStore, type IslandConfig } from '../store'
import { get } from 'svelte/store'

export type NavFn = (
  page: unknown,
  pageConfig: IslandConfig,
  expand?: boolean
) => void

export function getFallbackPage(): {
  page: unknown
  config: IslandConfig
} {
  const mediaState = get(globalMediaStore)
  if (mediaState.isPlaying) {
    return { page: MediaPage, config: mediaConfig }
  }
  return { page: HomePage, config: homeConfig }
}

export const AppControls = {
  getFallbackPage,

  goHome: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(HomePage, homeConfig, true)
  },

  goApps: (navigate: NavFn) => {
    navigate(AppsPage, appsConfig, true)
  },

  goMedia: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(MediaPage, mediaConfig, true)
  },

  goNotif: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(NotificationPage, notifConfig, true)
  },

  goSettings: (navigate: NavFn, e?: Event) => {
    if (e) e.stopPropagation()
    navigate(SettingsPage, settingsConfig, true)
  },

  doMinimize: (navigate: NavFn, e?: Event) => {
    if (e) {
      e.stopPropagation()
      e.preventDefault()
    }
    const fallback = getFallbackPage()
    navigate(fallback.page, fallback.config, false)
  },

  doHide: (hidePill: () => void, e: Event) => {
    e.stopPropagation()
    hidePill()
  },

  doQuit: (e?: Event) => {
    if (e) e.stopPropagation()
    window.dispatchEvent(new CustomEvent('vinta-quit'))
  }
}
