<script module lang="ts">
  import { Quit } from '../../wailsjs/go/backend/App.js'
  import ClockPage, { config as clockConfig } from '../pages/ClockPage.svelte'
  import AppsPage, { config as appsConfig } from '../pages/AppsPage.svelte'
  import SettingsPage, {
    config as settingsConfig
  } from '../pages/SettingsPage.svelte'
  import type { IslandConfig } from '../store'

  type NavFn = (
    page: unknown,
    pageConfig: IslandConfig,
    expand?: boolean
  ) => void

  export const AppControls = {
    goApps: (navigate: NavFn) => {
      navigate(AppsPage, appsConfig, true)
    },

    goSettings: (navigate: NavFn, e: Event) => {
      e.stopPropagation()
      navigate(SettingsPage, settingsConfig, true)
    },

    doMinimize: (navigate: NavFn, e: Event) => {
      e.stopPropagation()
      navigate(ClockPage, clockConfig, false)
    },

    doHide: (hidePill: () => void, e: Event) => {
      e.stopPropagation()
      hidePill()
    },

    doQuit: () => {
      Quit()
    }
  }
</script>
