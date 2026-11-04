<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 },
    expanded: { w: 440, h: 380, r: 30 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { getContext } from 'svelte'
  import {
    House,
    Bell,
    Music,
    Settings,
    Info,
    Search,
    FileText,
    SlidersHorizontal,
    Download,
    Calculator,
    Timer,
    Calendar as CalendarIcon,
    CloudSun,
    FolderDown,
    Activity
  } from 'lucide-svelte'
  import SearchInput from '../components/SearchInput.svelte'
  import HomePage, { config as homeConfig } from './HomePage.svelte'
  import NotificationPage, {
    config as notifConfig
  } from './NotificationPage.svelte'
  import MediaPage, { config as mediaConfig } from './MediaPage.svelte'
  import SettingsPage, { config as settingsConfig } from './SettingsPage.svelte'
  import NotesPage, { config as notesConfig } from './NotesPage.svelte'
  import CalculatorPage, {
    config as calculatorConfig
  } from './CalculatorPage.svelte'
  import TimerPage, { config as timerConfig } from './TimerPage.svelte'
  import CalendarPage, { config as calendarConfig } from './CalendarPage.svelte'
  import WeatherPage, { config as weatherConfig } from './WeatherPage.svelte'
  import FileShelfPage, {
    config as fileShelfConfig
  } from './FileShelfPage.svelte'
  import ActionCenterPage, {
    config as actionCenterConfig
  } from './ActionCenterPage.svelte'
  import PerformancePage, {
    config as performanceConfig
  } from './PerformancePage.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import {
    notifData,
    activeSettingsTabStore,
    pageVisibilityStore,
    updateInfoStore,
    type PageVisibility
  } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  /**
   * Opens the targeted page in expanded mode.
   *
   * @param page - Component to navigate to.
   * @param pageConfig - Island geometry config for the target page.
   */
  function openApp(page: unknown, pageConfig: IslandConfig) {
    navigate(page, pageConfig, true)
  }

  let searchQuery = $state('')

  interface AppItem {
    id: string
    name: string
    desc: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon: any
    badge?: boolean
    visibilityKey?: keyof PageVisibility
    action: () => void
  }

  const allApps: AppItem[] = [
    {
      id: 'about',
      name: 'About',
      desc: 'About Vinta',
      icon: Info,
      action: () => {
        activeSettingsTabStore.set('about')
        openApp(SettingsPage, settingsConfig)
      }
    },
    {
      id: 'actioncenter',
      name: 'Action Center',
      desc: 'Quick settings',
      icon: SlidersHorizontal,
      visibilityKey: 'actioncenter',
      action: () => openApp(ActionCenterPage, actionCenterConfig)
    },
    {
      id: 'calculator',
      name: 'Calculator',
      desc: 'Calculator',
      icon: Calculator,
      visibilityKey: 'calculator',
      action: () => openApp(CalculatorPage, calculatorConfig)
    },
    {
      id: 'calendar',
      name: 'Calendar',
      desc: 'Schedule & events',
      icon: CalendarIcon,
      visibilityKey: 'calendar',
      action: () => openApp(CalendarPage, calendarConfig)
    },
    {
      id: 'fileshelf',
      name: 'File Shelf',
      desc: 'Drag & drop file shelf',
      icon: FolderDown,
      visibilityKey: 'fileshelf',
      action: () => openApp(FileShelfPage, fileShelfConfig)
    },
    {
      id: 'home',
      name: 'Home',
      desc: 'Widgets & clock',
      icon: House,
      action: () => openApp(HomePage, homeConfig)
    },
    {
      id: 'media',
      name: 'Media',
      desc: 'Now playing',
      icon: Music,
      visibilityKey: 'media',
      action: () => openApp(MediaPage, mediaConfig)
    },
    {
      id: 'notifs',
      name: 'Notifications',
      desc: 'Notifications',
      icon: Bell,
      badge: true,
      visibilityKey: 'notifications',
      action: () => openApp(NotificationPage, notifConfig)
    },
    {
      id: 'notes',
      name: 'Quick Notes',
      desc: 'Quick notes',
      icon: FileText,
      visibilityKey: 'notes',
      action: () => openApp(NotesPage, notesConfig)
    },
    {
      id: 'settings',
      name: 'Settings',
      desc: 'Settings',
      icon: Settings,
      action: () => {
        activeSettingsTabStore.set('appearance')
        openApp(SettingsPage, settingsConfig)
      }
    },
    {
      id: 'performance',
      name: 'System Stats',
      desc: 'CPU, RAM, and GPU stats',
      icon: Activity,
      visibilityKey: 'performance',
      action: () => openApp(PerformancePage, performanceConfig)
    },
    {
      id: 'timer',
      name: 'Timer',
      desc: 'Timer, stopwatch & focus',
      icon: Timer,
      visibilityKey: 'timer',
      action: () => openApp(TimerPage, timerConfig)
    },
    {
      id: 'weather',
      name: 'Weather',
      desc: 'Forecast & metrics',
      icon: CloudSun,
      visibilityKey: 'weather',
      action: () => openApp(WeatherPage, weatherConfig)
    },
    {
      id: 'updates',
      name: 'Updates',
      desc: 'Check for updates',
      icon: Download,
      badge: $updateInfoStore?.updateAvailable,
      action: () => {
        activeSettingsTabStore.set('updates')
        openApp(SettingsPage, settingsConfig)
      }
    }
  ]

  /**
   * Computes the list of apps sorted alphabetically and filtered by the active search query.
   *
   * @returns Filtered and alphabetically sorted array of applications.
   */
  function getFilteredApps(): AppItem[] {
    const q = searchQuery.trim().toLowerCase()
    const sorted = [...allApps].sort((a, b) => a.name.localeCompare(b.name))
    if (!q) return sorted
    return sorted.filter(
      (app) =>
        app.name.toLowerCase().includes(q) || app.desc.toLowerCase().includes(q)
    )
  }

  let filteredApps = $derived.by(getFilteredApps)
</script>

<div data-vinta="apps-page" class="apps-page">
  <ScrollView class="apps-scroll-area">
    {#if filteredApps.length === 0}
      <div class="empty-state">
        <Search size={22} class="empty-icon" />
        <span class="empty-title">No apps found</span>
        <span class="empty-desc">No applications match "{searchQuery}"</span>
      </div>
    {:else}
      <div class="apps-grid">
        {#each filteredApps as app (app.id)}
          {@const isEnabled =
            !app.visibilityKey || $pageVisibilityStore[app.visibilityKey]}
          <button
            class="app-tile app-{app.id}"
            class:is-disabled={!isEnabled}
            use:islandTooltip={{
              text: isEnabled ? app.desc : `${app.desc} (Disabled)`,
              position: 'bottom'
            }}
            onclick={(e) => {
              e.stopPropagation()
              if (isEnabled) {
                app.action()
              }
            }}
          >
            <div class="tile-icon-box">
              <app.icon size={24} strokeWidth={2} class="tile-icon" />
              {#if isEnabled && app.badge && $notifData.length > 0}
                <span class="tile-badge">
                  {$notifData.length > 9 ? '9+' : $notifData.length}
                </span>
              {:else if isEnabled && app.id === 'updates' && $updateInfoStore?.updateAvailable}
                <span class="tile-badge-dot"></span>
              {/if}
            </div>
            <span class="tile-title">{app.name}</span>
          </button>
        {/each}
      </div>
    {/if}
  </ScrollView>

  <div class="header">
    <div class="header-left">
      <span class="title">Apps</span>
    </div>
    <SearchInput
      bind:value={searchQuery}
      placeholder="Search apps..."
      maxWidth="175px"
    />
  </div>
</div>

<style src="./styles/AppsPage.css"></style>
