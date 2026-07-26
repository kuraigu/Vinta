<script lang="ts">
  import { getContext, onMount } from 'svelte'
  import { EventsOn } from '../../wailsjs/runtime/runtime.js'
  import {
    Grid,
    Music,
    Settings,
    Power,
    Minimize2,
    EyeOff,
    House,
    Bell
  } from 'lucide-svelte'
  import { AppControls } from '../utils/AppControls.svelte'
  import DockButton from './DockButton.svelte'
  import AppsPage from '../pages/AppsPage.svelte'
  import HomePage from '../pages/HomePage.svelte'
  import NotificationPage from '../pages/NotificationPage.svelte'
  import MediaPage, { config as mediaConfig } from '../pages/MediaPage.svelte'
  import SettingsPage from '../pages/SettingsPage.svelte'
  import type { IslandConfig } from '../store'
  import { notifData } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  const hidePill = getContext<() => void>('hidePill')

  let {
    activePage,
    islandConfig
  }: { activePage: unknown; islandConfig?: IslandConfig } = $props()

  let isMediaPlaying = $state(false)

  onMount(() => {
    const unlisten = EventsOn(
      'media-update',
      (data: { isPlaying?: boolean }) => {
        if (data && typeof data.isPlaying === 'boolean') {
          isMediaPlaying = data.isPlaying
        }
      }
    )
    return () => {
      if (unlisten) unlisten()
    }
  })
</script>

<div class="dock-container">
  <div class="dock">
    <DockButton
      tooltip="Apps"
      active={activePage === AppsPage}
      onclick={() => AppControls.goApps(navigate)}
    >
      <Grid size={18} />
    </DockButton>

    <div class="divider"></div>

    <DockButton
      tooltip="Home"
      active={activePage === HomePage}
      onclick={(e: MouseEvent) => AppControls.goHome(navigate, e)}
    >
      <House size={18} />
    </DockButton>

    <div class="divider"></div>

    <DockButton
      tooltip="Now Playing"
      active={activePage === MediaPage}
      onclick={(e: MouseEvent) => AppControls.goMedia(navigate, e)}
    >
      <Music size={18} />
    </DockButton>

    <div class="divider"></div>

    <DockButton
      tooltip="Notifications"
      active={activePage === NotificationPage}
      badge={$notifData.length > 0 ? $notifData.length : undefined}
      onclick={(e: MouseEvent) => AppControls.goNotif(navigate, e)}
    >
      <Bell size={18} />
    </DockButton>

    <div class="divider"></div>

    <DockButton
      tooltip="Settings"
      active={activePage === SettingsPage}
      onclick={(e: MouseEvent) => AppControls.goSettings(navigate, e)}
    >
      <Settings size={18} />
    </DockButton>

    <div class="divider"></div>

    <DockButton
      variant="success"
      tooltip="Hide Widget"
      onclick={(e: MouseEvent) => AppControls.doHide(hidePill, e)}
    >
      <EyeOff size={18} />
    </DockButton>

    <div class="divider"></div>

    <DockButton
      variant="warning"
      tooltip="Collapse Island"
      onclick={(e: MouseEvent) =>
        AppControls.doMinimize(
          navigate,
          e,
          activePage,
          islandConfig || mediaConfig,
          isMediaPlaying
        )}
    >
      <Minimize2 size={18} />
    </DockButton>

    <div class="divider"></div>

    <DockButton
      variant="danger"
      tooltip="Quit Vinta"
      onclick={(e: MouseEvent) => AppControls.doQuit(e)}
    >
      <Power size={18} />
    </DockButton>
  </div>
</div>

<style>
  .dock-container {
    position: absolute;
    bottom: 8px;
    left: 0;
    right: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 100;
    pointer-events: none;
  }

  .dock {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    border-radius: 20px;
    background: var(--c-surface-elevated, rgba(30, 30, 40, 0.75));
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    border: 1px solid var(--c-border-subtle, rgba(255, 255, 255, 0.1));
    box-shadow:
      0 4px 16px rgba(0, 0, 0, 0.3),
      0 0 0 1px rgba(255, 255, 255, 0.05);
    pointer-events: auto;
  }

  .divider {
    width: 1px;
    height: 16px;
    background: var(--c-border-subtle, rgba(255, 255, 255, 0.1));
    margin: 0 2px;
  }
</style>
