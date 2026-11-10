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
  /**
   * @fileoverview Application launcher and search grid.
   *
   * Lists all integrated Vinta tools and utilities in a searchable grid, letting users
   * filter tools by name or description and jump directly to any view.
   */

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
    Activity,
    Megaphone,
    ScrollText
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
  import AnnouncementsPage, {
    config as announcementsConfig
  } from './AnnouncementsPage.svelte'
  import ReleaseNotesPage, {
    config as releaseNotesConfig
  } from './ReleaseNotesPage.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import {
    notifData,
    activeSettingsTabStore,
    pageVisibilityStore,
    updateInfoStore,
    tStore,
    type PageVisibility
  } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  /**
   * Activates the selected application and reconfigures island dimensions according to its manifest.
   *
   * @param page - Target application identifier or page component.
   * @param pageConfig - Minimal and expanded dimensions registered for the application.
   */
  function openApp(page: unknown, pageConfig: IslandConfig) {
    navigate(page, pageConfig, true)
  }

  let searchQuery = $state('')

  interface AppItemDef {
    id: string
    nameKey: string
    descKey: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon: any
    badge?: boolean
    visibilityKey?: keyof PageVisibility
    action: () => void
  }

  interface ResolvedAppItem extends AppItemDef {
    name: string
    desc: string
  }

  const allApps: AppItemDef[] = [
    {
      id: 'about',
      nameKey: 'apps.about_name',
      descKey: 'apps.about_desc',
      icon: Info,
      action: () => {
        activeSettingsTabStore.set('about')
        openApp(SettingsPage, settingsConfig)
      }
    },
    {
      id: 'announcements',
      nameKey: 'apps.announcements_name',
      descKey: 'apps.announcements_desc',
      icon: Megaphone,
      action: () => openApp(AnnouncementsPage, announcementsConfig)
    },
    {
      id: 'releasenotes',
      nameKey: 'apps.releasenotes_name',
      descKey: 'apps.releasenotes_desc',
      icon: ScrollText,
      action: () => openApp(ReleaseNotesPage, releaseNotesConfig)
    },
    {
      id: 'actioncenter',
      nameKey: 'apps.actioncenter_name',
      descKey: 'apps.actioncenter_desc',
      icon: SlidersHorizontal,
      visibilityKey: 'actioncenter',
      action: () => openApp(ActionCenterPage, actionCenterConfig)
    },
    {
      id: 'calculator',
      nameKey: 'apps.calculator_name',
      descKey: 'apps.calculator_desc',
      icon: Calculator,
      visibilityKey: 'calculator',
      action: () => openApp(CalculatorPage, calculatorConfig)
    },
    {
      id: 'calendar',
      nameKey: 'apps.calendar_name',
      descKey: 'apps.calendar_desc',
      icon: CalendarIcon,
      visibilityKey: 'calendar',
      action: () => openApp(CalendarPage, calendarConfig)
    },
    {
      id: 'fileshelf',
      nameKey: 'apps.fileshelf_name',
      descKey: 'apps.fileshelf_desc',
      icon: FolderDown,
      visibilityKey: 'fileshelf',
      action: () => openApp(FileShelfPage, fileShelfConfig)
    },
    {
      id: 'home',
      nameKey: 'apps.home_name',
      descKey: 'apps.home_desc',
      icon: House,
      action: () => openApp(HomePage, homeConfig)
    },
    {
      id: 'media',
      nameKey: 'apps.media_name',
      descKey: 'apps.media_desc',
      icon: Music,
      visibilityKey: 'media',
      action: () => openApp(MediaPage, mediaConfig)
    },
    {
      id: 'notifs',
      nameKey: 'apps.notifs_name',
      descKey: 'apps.notifs_desc',
      icon: Bell,
      badge: true,
      visibilityKey: 'notifications',
      action: () => openApp(NotificationPage, notifConfig)
    },
    {
      id: 'notes',
      nameKey: 'apps.notes_name',
      descKey: 'apps.notes_desc',
      icon: FileText,
      visibilityKey: 'notes',
      action: () => openApp(NotesPage, notesConfig)
    },
    {
      id: 'settings',
      nameKey: 'apps.settings_name',
      descKey: 'apps.settings_desc',
      icon: Settings,
      action: () => {
        activeSettingsTabStore.set('appearance')
        openApp(SettingsPage, settingsConfig)
      }
    },
    {
      id: 'performance',
      nameKey: 'apps.performance_name',
      descKey: 'apps.performance_desc',
      icon: Activity,
      visibilityKey: 'performance',
      action: () => openApp(PerformancePage, performanceConfig)
    },
    {
      id: 'timer',
      nameKey: 'apps.timer_name',
      descKey: 'apps.timer_desc',
      icon: Timer,
      visibilityKey: 'timer',
      action: () => openApp(TimerPage, timerConfig)
    },
    {
      id: 'weather',
      nameKey: 'apps.weather_name',
      descKey: 'apps.weather_desc',
      icon: CloudSun,
      visibilityKey: 'weather',
      action: () => openApp(WeatherPage, weatherConfig)
    },
    {
      id: 'updates',
      nameKey: 'apps.updates_name',
      descKey: 'apps.updates_desc',
      icon: Download,
      badge: $updateInfoStore?.updateAvailable,
      action: () => {
        activeSettingsTabStore.set('updates')
        openApp(SettingsPage, settingsConfig)
      }
    }
  ]

  /**
   * Filters registered applications against user search input, matching localized names and descriptions.
   *
   * @returns Filtered and alphabetically sorted array of applications.
   */
  function getFilteredApps(): ResolvedAppItem[] {
    const q = searchQuery.trim().toLowerCase()
    const resolved: ResolvedAppItem[] = allApps.map((app) => ({
      ...app,
      name: $tStore(app.nameKey),
      desc: $tStore(app.descKey)
    }))
    const sorted = [...resolved].sort((a, b) => a.name.localeCompare(b.name))
    if (!q) return sorted
    return sorted.filter(
      (app) =>
        app.name.toLowerCase().includes(q) || app.desc.toLowerCase().includes(q)
    )
  }

  let filteredApps = $derived.by(getFilteredApps)

  /**
   * Launches the first matching enabled app when pressing Enter during search.
   *
   * @param e - Window keyboard event.
   */
  function handleAppsKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter' && searchQuery.trim() && filteredApps.length > 0) {
      const first = filteredApps.find(
        (app) => !app.visibilityKey || $pageVisibilityStore[app.visibilityKey]
      )
      if (first) {
        e.preventDefault()
        first.action()
      }
    }
  }
</script>

<svelte:window onkeydown={handleAppsKeyDown} />

<div data-vinta="apps-page" class="apps-page">
  <ScrollView class="apps-scroll-area">
    {#if filteredApps.length === 0}
      <div class="empty-state">
        <Search size={22} class="empty-icon" />
        <span class="empty-title">{$tStore('apps.no_apps_found')}</span>
        <span class="empty-desc"
          >{$tStore('apps.no_apps_match', { query: searchQuery })}</span
        >
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
              text: isEnabled
                ? ''
                : `${app.desc} (${$tStore('common.disabled')})`,
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
      <span class="title">{$tStore('apps.title')}</span>
    </div>
    <SearchInput
      bind:value={searchQuery}
      placeholder={$tStore('apps.search_placeholder')}
      maxWidth="175px"
    />
  </div>
</div>

<style src="./styles/AppsPage.css"></style>
