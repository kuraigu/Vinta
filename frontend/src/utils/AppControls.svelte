<script module lang="ts">
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

  type NavFn = (
    page: unknown,
    pageConfig: IslandConfig,
    expand?: boolean
  ) => void

  export const AppControls = {
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

    goSettings: (navigate: NavFn, e: Event) => {
      e.stopPropagation()
      navigate(SettingsPage, settingsConfig, true)
    },

    doMinimize: (
      navigate: NavFn,
      e: Event,
      _currentPage?: unknown,
      _currentConfig?: IslandConfig,
      isMediaPlaying?: boolean
    ) => {
      e.stopPropagation()
      e.preventDefault()
      const mediaState = get(globalMediaStore)
      const isPlaying = isMediaPlaying || mediaState.isPlaying
      if (isPlaying) {
        navigate(MediaPage, mediaConfig, false)
      } else {
        navigate(HomePage, homeConfig, false)
      }
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
</script>
