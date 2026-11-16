<!--
  @fileoverview Overlay page hosting the searchable Apps launcher on top of the active workspace
  without resetting underlying workspace state.
-->
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
    Activity,
    Megaphone,
    ScrollText
  } from 'lucide-svelte'
  import OverlayPage from '../../components/OverlayPage.svelte'
  import SearchInput from '../../components/SearchInput.svelte'
  import ScrollView from '../../components/ScrollView.svelte'
  import { islandTooltip } from '../../services/islandTooltip'
  import HomePage, { config as homeConfig } from '../HomePage.svelte'
  import MediaPage, { config as mediaConfig } from '../MediaPage.svelte'
  import SettingsPage, { config as settingsConfig } from '../SettingsPage.svelte'
  import NotesPage, { config as notesConfig } from '../NotesPage.svelte'
  import CalculatorPage, {
    config as calculatorConfig
  } from '../CalculatorPage.svelte'
  import TimerPage, { config as timerConfig } from '../TimerPage.svelte'
  import CalendarPage, { config as calendarConfig } from '../CalendarPage.svelte'
  import WeatherPage, { config as weatherConfig } from '../WeatherPage.svelte'
  import FileShelfPage, {
    config as fileShelfConfig
  } from '../FileShelfPage.svelte'
  import ActionCenterPage, {
    config as actionCenterConfig
  } from '../ActionCenterPage.svelte'
  import PerformancePage, {
    config as performanceConfig
  } from '../PerformancePage.svelte'
  import {
    notifData,
    activeSettingsTabStore,
    pageVisibilityStore,
    updateInfoStore,
    notificationsOverlayTabStore,
    tStore,
    type IslandConfig,
    type PageVisibility,
    type OverlayPageId,
    type NotificationsOverlayTab
  } from '../../store'

  interface Props {
    onClose: () => void
  }

  let { onClose }: Props = $props()

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  const openOverlay =
    getContext<
      (
        overlay: OverlayPageId,
        options?: { tab?: NotificationsOverlayTab }
      ) => void
    >('openOverlay')

  function openWorkspace(page: unknown, pageConfig: IslandConfig) {
    onClose()
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

  const allApps: AppItemDef[] = [
    {
      id: 'about',
      nameKey: 'apps.about_name',
      descKey: 'apps.about_desc',
      icon: Info,
      action: () => {
        activeSettingsTabStore.set('about')
        openWorkspace(SettingsPage, settingsConfig)
      }
    },
    {
      id: 'announcements',
      nameKey: 'apps.announcements_name',
      descKey: 'apps.announcements_desc',
      icon: Megaphone,
      action: () => openOverlay('notifications', { tab: 'announcements' })
    },
    {
      id: 'releasenotes',
      nameKey: 'apps.releasenotes_name',
      descKey: 'apps.releasenotes_desc',
      icon: ScrollText,
      action: () => openOverlay('notifications', { tab: 'releases' })
    },
    {
      id: 'actioncenter',
      nameKey: 'apps.actioncenter_name',
      descKey: 'apps.actioncenter_desc',
      icon: SlidersHorizontal,
      visibilityKey: 'actioncenter',
      action: () => openWorkspace(ActionCenterPage, actionCenterConfig)
    },
    {
      id: 'calculator',
      nameKey: 'apps.calculator_name',
      descKey: 'apps.calculator_desc',
      icon: Calculator,
      visibilityKey: 'calculator',
      action: () => openWorkspace(CalculatorPage, calculatorConfig)
    },
    {
      id: 'calendar',
      nameKey: 'apps.calendar_name',
      descKey: 'apps.calendar_desc',
      icon: CalendarIcon,
      visibilityKey: 'calendar',
      action: () => openWorkspace(CalendarPage, calendarConfig)
    },
    {
      id: 'fileshelf',
      nameKey: 'apps.fileshelf_name',
      descKey: 'apps.fileshelf_desc',
      icon: FolderDown,
      visibilityKey: 'fileshelf',
      action: () => openWorkspace(FileShelfPage, fileShelfConfig)
    },
    {
      id: 'home',
      nameKey: 'apps.home_name',
      descKey: 'apps.home_desc',
      icon: House,
      action: () => openWorkspace(HomePage, homeConfig)
    },
    {
      id: 'media',
      nameKey: 'apps.media_name',
      descKey: 'apps.media_desc',
      icon: Music,
      visibilityKey: 'media',
      action: () => openWorkspace(MediaPage, mediaConfig)
    },
    {
      id: 'notes',
      nameKey: 'apps.notes_name',
      descKey: 'apps.notes_desc',
      icon: FileText,
      visibilityKey: 'notes',
      action: () => openWorkspace(NotesPage, notesConfig)
    },
    {
      id: 'notifications',
      nameKey: 'apps.notifications_name',
      descKey: 'apps.notifications_desc',
      icon: Bell,
      badge: true,
      visibilityKey: 'notifications',
      action: () => {
        notificationsOverlayTabStore.set('notifications')
        openOverlay('notifications', { tab: 'notifications' })
      }
    },
    {
      id: 'performance',
      nameKey: 'apps.performance_name',
      descKey: 'apps.performance_desc',
      icon: Activity,
      visibilityKey: 'performance',
      action: () => openWorkspace(PerformancePage, performanceConfig)
    },
    {
      id: 'settings',
      nameKey: 'apps.settings_name',
      descKey: 'apps.settings_desc',
      icon: Settings,
      action: () => openWorkspace(SettingsPage, settingsConfig)
    },
    {
      id: 'timer',
      nameKey: 'apps.timer_name',
      descKey: 'apps.timer_desc',
      icon: Timer,
      visibilityKey: 'timer',
      action: () => openWorkspace(TimerPage, timerConfig)
    },
    {
      id: 'updates',
      nameKey: 'apps.updates_name',
      descKey: 'apps.updates_desc',
      icon: Download,
      action: () => {
        activeSettingsTabStore.set('updates')
        openWorkspace(SettingsPage, settingsConfig)
      }
    },
    {
      id: 'weather',
      nameKey: 'apps.weather_name',
      descKey: 'apps.weather_desc',
      icon: CloudSun,
      visibilityKey: 'weather',
      action: () => openWorkspace(WeatherPage, weatherConfig)
    }
  ]

  let resolvedApps = $derived(
    allApps.map((app) => ({
      ...app,
      name: $tStore(app.nameKey),
      desc: $tStore(app.descKey)
    }))
  )

  let filteredApps = $derived.by(() => {
    const q = searchQuery.trim().toLowerCase()
    if (!q) return resolvedApps
    return resolvedApps.filter(
      (a) =>
        a.name.toLowerCase().includes(q) || a.desc.toLowerCase().includes(q)
    )
  })
</script>

<OverlayPage
  dataVinta="apps-overlay-page"
  title={$tStore('apps.title')}
  {onClose}
>
  {#snippet headerActions()}
    <SearchInput
      bind:value={searchQuery}
      placeholder={$tStore('apps.search_placeholder')}
      maxWidth="175px"
    />
  {/snippet}

  <div class="apps-page">
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
  </div>
</OverlayPage>

<style src="../styles/AppsPage.css"></style>
