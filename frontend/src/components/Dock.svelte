<script lang="ts">
  import { getContext } from 'svelte'
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
  import MediaPage from '../pages/MediaPage.svelte'
  import SettingsPage from '../pages/SettingsPage.svelte'
  import type { IslandConfig } from '../store'
  import { notifData, pageVisibilityStore } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  const hidePill = getContext<() => void>('hidePill')

  let { activePage }: { activePage: unknown; islandConfig?: IslandConfig } =
    $props()
</script>

<div class="dock-container">
  <div class="dock">
    <DockButton
      tooltip="Apps"
      active={activePage === AppsPage}
      onclick={(e: MouseEvent) => AppControls.goApps(navigate, e)}
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

    {#if $pageVisibilityStore.media}
      <div class="divider"></div>

      <DockButton
        tooltip="Now Playing"
        active={activePage === MediaPage}
        onclick={(e: MouseEvent) => AppControls.goMedia(navigate, e)}
      >
        <Music size={18} />
      </DockButton>
    {/if}

    {#if $pageVisibilityStore.notifications}
      <div class="divider"></div>

      <DockButton
        tooltip="Notifications"
        active={activePage === NotificationPage}
        badge={$notifData.length > 0 ? $notifData.length : undefined}
        onclick={(e: MouseEvent) => AppControls.goNotif(navigate, e)}
      >
        <Bell size={18} />
      </DockButton>
    {/if}

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
      onclick={(e: MouseEvent) => AppControls.doHide(navigate, hidePill, e)}
    >
      <EyeOff size={18} />
    </DockButton>

    <div class="divider"></div>

    <DockButton
      variant="warning"
      tooltip="Collapse Island"
      onclick={(e: MouseEvent) =>
        AppControls.doMinimize(navigate, e, activePage)}
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
    bottom: 6px;
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
    background: var(--c-dock-popover-bg);
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    border: 1px solid var(--c-border);
    box-shadow: var(--c-dock-popover-shadow);
    pointer-events: auto;
  }

  .divider {
    width: 1px;
    height: 16px;
    background: var(--c-border);
    margin: 0 2px;
  }
</style>
