<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 }, // Apps page is never minimal
    expanded: { w: 420, h: 285, r: 28 }, // Comfortable width and height for Bento Grid + Dock
    borderColor: 'var(--c-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { getContext } from 'svelte'
  import { House, Bell, Music, Settings, Info, Search } from 'lucide-svelte'
  import HomePage, { config as homeConfig } from './HomePage.svelte'
  import NotificationPage, {
    config as notifConfig
  } from './NotificationPage.svelte'
  import MediaPage, { config as mediaConfig } from './MediaPage.svelte'
  import SettingsPage, { config as settingsConfig } from './SettingsPage.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import { notifData } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  function openApp(page: unknown, config: IslandConfig) {
    navigate(page, config, true)
  }

  let searchQuery = $state('')

  const allApps = [
    {
      id: 'home',
      name: 'Home',
      desc: 'Widgets & Clock',
      icon: House,
      color: 'gradient-orange',
      action: () => openApp(HomePage, homeConfig)
    },
    {
      id: 'media',
      name: 'Media',
      desc: 'Now Playing',
      icon: Music,
      color: 'gradient-pink',
      action: () => openApp(MediaPage, mediaConfig)
    },
    {
      id: 'notifs',
      name: 'Notifications',
      desc: 'System Alerts',
      icon: Bell,
      color: 'gradient-blue',
      badge: true,
      action: () => openApp(NotificationPage, notifConfig)
    },
    {
      id: 'settings',
      name: 'Settings',
      desc: 'Customization',
      icon: Settings,
      color: 'gradient-purple',
      action: () => openApp(SettingsPage, settingsConfig)
    },
    {
      id: 'about',
      name: 'About',
      desc: 'Vinta Island v1.0',
      icon: Info,
      color: 'gradient-dark',
      action: () => {}
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
          <button
            class="bento-app-card {app.color}"
            onclick={(e) => {
              e.stopPropagation()
              app.action()
            }}
          >
            <div class="card-icon-wrapper">
              <app.icon size={20} />
              {#if app.badge && $notifData.length > 0}
                <span class="card-badge">
                  {$notifData.length > 9 ? '9+' : $notifData.length}
                </span>
              {/if}
            </div>
            <div class="card-text">
              <span class="card-title">{app.name}</span>
              <span class="card-desc">{app.desc}</span>
            </div>
          </button>
        {/each}
      </div>
    {/if}
  </ScrollView>
</div>

<style>
  .apps-page {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 14px 16px 58px 16px;
    box-sizing: border-box;
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
    flex-shrink: 0;
  }

  .title {
    font-weight: 700;
    font-size: 15px;
    letter-spacing: -0.01em;
    color: var(--c-text-primary);
  }

  .search-box {
    display: flex;
    align-items: center;
    gap: 6px;
    background: var(--c-item-bg, rgba(255, 255, 255, 0.08));
    border: 1px solid var(--c-border, rgba(255, 255, 255, 0.1));
    border-radius: 12px;
    padding: 4px 10px;
    flex: 1;
    max-width: 190px;
  }

  :global(.search-icon) {
    color: var(--c-text-muted);
    flex-shrink: 0;
  }

  .search-box input {
    background: transparent;
    border: none;
    outline: none;
    color: var(--c-text-primary);
    font-size: 11px;
    width: 100%;
  }

  .search-box input::placeholder {
    color: var(--c-text-muted);
  }

  .bento-container {
    flex: 1;
    min-height: 0;
    padding-right: 2px;
  }

  .bento-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
  }

  .bento-app-card {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    border-radius: 14px;
    background: var(--c-item-bg, rgba(255, 255, 255, 0.06));
    border: 1px solid var(--c-border, rgba(255, 255, 255, 0.1));
    color: var(--c-text-primary);
    cursor: pointer;
    text-align: left;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .bento-app-card:hover {
    background: var(--c-hover, rgba(255, 255, 255, 0.12));
    transform: translateY(-1px);
    border-color: var(--c-primary, #6366f1);
  }

  .card-icon-wrapper {
    position: relative;
    width: 34px;
    height: 34px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    color: var(--c-text-primary);
  }

  .bento-app-card.gradient-orange .card-icon-wrapper {
    background: linear-gradient(135deg, var(--c-warning), var(--c-amber));
  }

  .bento-app-card.gradient-pink .card-icon-wrapper {
    background: linear-gradient(135deg, var(--c-danger), var(--c-purple));
  }

  .bento-app-card.gradient-blue .card-icon-wrapper {
    background: linear-gradient(
      135deg,
      var(--c-primary),
      var(--c-surface-elevated, #3a6073)
    );
  }

  .bento-app-card.gradient-purple .card-icon-wrapper {
    background: linear-gradient(135deg, var(--c-purple), var(--c-primary));
  }

  .bento-app-card.gradient-dark .card-icon-wrapper {
    background: linear-gradient(135deg, var(--c-contrast-bg), var(--c-item-bg));
  }

  .card-badge {
    position: absolute;
    top: -3px;
    right: -3px;
    background: var(--c-danger);
    color: var(--c-text-primary);
    font-size: 8px;
    font-weight: 700;
    padding: 1px 4px;
    border-radius: 6px;
    line-height: 1;
    border: 1px solid rgba(0, 0, 0, 0.4);
  }

  .card-text {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .card-title {
    font-size: 12px;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .card-desc {
    font-size: 9px;
    color: var(--c-text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .empty-state {
    font-size: 11px;
    color: var(--c-text-muted);
    text-align: center;
    padding: 24px 0;
  }
</style>
