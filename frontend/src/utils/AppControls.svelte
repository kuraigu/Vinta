<script module lang="ts">
  import HomePage, { config as homeConfig } from '../pages/HomePage.svelte'
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
      navigate(HomePage, homeConfig, false)
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
