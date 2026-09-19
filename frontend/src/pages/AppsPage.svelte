<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 }, // Apps page is never minimal
    expanded: { w: 440, h: 380, r: 30 }, // Standardized dimensions
    borderColor: 'var(--c-border)',
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
    Calculator
  } from 'lucide-svelte'
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
  import ActionCenterPage, {
    config as actionCenterConfig
  } from './ActionCenterPage.svelte'
  import ScrollView from '../components/ScrollView.svelte'
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

  function openApp(page: unknown, config: IslandConfig) {
    navigate(page, config, true)
  }

  let searchQuery = $state('')

  interface AppItem {
    id: string
    name: string
    desc: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon: any
    color: string
    badge?: boolean
    visibilityKey?: keyof PageVisibility
    action: () => void
  }

  const allApps: AppItem[] = [
    {
      id: 'home',
      name: 'Home',
      desc: 'Widgets & Clock',
      icon: House,
      color: 'gradient-green',
      action: () => openApp(HomePage, homeConfig)
    },
    {
      id: 'actioncenter',
      name: 'Action Center',
      desc: 'Quick Controls',
      icon: SlidersHorizontal,
      color: 'gradient-indigo',
      visibilityKey: 'actioncenter',
      action: () => openApp(ActionCenterPage, actionCenterConfig)
    },
    {
      id: 'notes',
      name: 'Quick Notes',
      desc: 'Scratchpad & Notes',
      icon: FileText,
      color: 'gradient-amber',
      visibilityKey: 'notes',
      action: () => openApp(NotesPage, notesConfig)
    },
    {
      id: 'calculator',
      name: 'Calculator',
      desc: 'Math & Utilities',
      icon: Calculator,
      color: 'gradient-orange',
      action: () => openApp(CalculatorPage, calculatorConfig)
    },
    {
      id: 'media',
      name: 'Media',
      desc: 'Now Playing',
      icon: Music,
      color: 'gradient-pink',
      visibilityKey: 'media',
      action: () => openApp(MediaPage, mediaConfig)
    },
    {
      id: 'notifs',
      name: 'Notifications',
      desc: 'System Alerts',
      icon: Bell,
      color: 'gradient-blue',
      badge: true,
      visibilityKey: 'notifications',
      action: () => openApp(NotificationPage, notifConfig)
    },
    {
      id: 'settings',
      name: 'Settings',
      desc: 'Customization',
      icon: Settings,
      color: 'gradient-dark',
      action: () => {
        activeSettingsTabStore.set('appearance')
        openApp(SettingsPage, settingsConfig)
      }
    },
    {
      id: 'about',
      name: 'About',
      desc: 'Check Vinta Information',
      icon: Info,
      color: 'gradient-purple',
      action: () => {
        activeSettingsTabStore.set('about')
        openApp(SettingsPage, settingsConfig)
      }
    },
    {
      id: 'updates',
      name: 'Check for Updates',
      desc: 'Releases & Highlights',
      icon: Download,
      color: 'gradient-cyan',
      badge: $updateInfoStore?.updateAvailable,
      action: () => {
        activeSettingsTabStore.set('updates')
        openApp(SettingsPage, settingsConfig)
      }
    }
  ]

  let filteredApps = $derived.by(() => {
    const q = searchQuery.trim().toLowerCase()
    if (!q) return allApps
    return allApps.filter(
      (app) =>
        app.name.toLowerCase().includes(q) || app.desc.toLowerCase().includes(q)
    )
  })
</script>

<div class="apps-page">
  <div class="header">
    <div class="title">Apps</div>
    <div class="search-box">
      <Search size={13} class="search-icon" />
      <input
        type="text"
        placeholder="Search apps..."
        bind:value={searchQuery}
        onclick={(e) => e.stopPropagation()}
      />
    </div>
  </div>

  <ScrollView class="bento-container">
    {#if filteredApps.length === 0}
      <div class="empty-state">No apps matching "{searchQuery}"</div>
    {:else}
      <div class="bento-grid">
        {#each filteredApps as app (app.id)}
          {@const isEnabled =
            !app.visibilityKey || $pageVisibilityStore[app.visibilityKey]}
          <button
            class="bento-app-card {app.color}"
            class:is-disabled={!isEnabled}
            title={!isEnabled ? `${app.name} (Disabled in Settings)` : app.name}
            onclick={(e) => {
              e.stopPropagation()
              if (isEnabled) {
                app.action()
              }
            }}
          >
            <div class="card-icon-wrapper">
              <app.icon size={19} />
              {#if isEnabled && app.badge && $notifData.length > 0}
                <span class="card-badge">
                  {$notifData.length > 9 ? '9+' : $notifData.length}
                </span>
              {/if}
            </div>
            <div class="card-text">
              <span class="card-title">{app.name}</span>
              <span class="card-desc">{isEnabled ? app.desc : 'Disabled'}</span>
            </div>
          </button>
        {/each}
      </div>
    {/if}
  </ScrollView>
</div>

<style src="./styles/AppsPage.css"></style>
