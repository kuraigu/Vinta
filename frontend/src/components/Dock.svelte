<script lang="ts">
  import { getContext } from 'svelte'
  import {
    Grid,
    Music,
    Settings,
    House,
    Bell,
    FileText,
    Calculator,
    Info
  } from 'lucide-svelte'
  import { AppControls } from '../utils/AppControls.svelte'
  import DockButton from './DockButton.svelte'
  import AppsPage from '../pages/AppsPage.svelte'
  import HomePage from '../pages/HomePage.svelte'
  import NotificationPage from '../pages/NotificationPage.svelte'
  import MediaPage from '../pages/MediaPage.svelte'
  import SettingsPage from '../pages/SettingsPage.svelte'
  import NotesPage from '../pages/NotesPage.svelte'
  import CalculatorPage from '../pages/CalculatorPage.svelte'
  import type { IslandConfig } from '../store'
  import {
    notifData,
    pageVisibilityStore,
    activeSettingsTabStore
  } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let { activePage }: { activePage: unknown; islandConfig?: IslandConfig } =
    $props()
</script>

<div class="dock-container">
  <div class="dock">
    <div class="dock-group">
      <DockButton
        tooltip="Apps"
        active={activePage === AppsPage}
        onclick={(e: MouseEvent) => AppControls.goApps(navigate, e)}
      >
        <Grid size={18} />
      </DockButton>

      <DockButton
        tooltip="Home"
        active={activePage === HomePage}
        onclick={(e: MouseEvent) => AppControls.goHome(navigate, e)}
      >
        <House size={18} />
      </DockButton>

      {#if $pageVisibilityStore.notes !== false}
        <DockButton
          tooltip="Notes"
          active={activePage === NotesPage}
          onclick={(e: MouseEvent) => AppControls.goNotes(navigate, e)}
        >
          <FileText size={18} />
        </DockButton>
      {/if}

      {#if $pageVisibilityStore.calculator !== false}
        <DockButton
          tooltip="Calculator"
          active={activePage === CalculatorPage}
          onclick={(e: MouseEvent) => AppControls.goCalculator(navigate, e)}
        >
          <Calculator size={18} />
        </DockButton>
      {/if}

      {#if $pageVisibilityStore.media}
        <DockButton
          tooltip="Now Playing"
          active={activePage === MediaPage}
          onclick={(e: MouseEvent) => AppControls.goMedia(navigate, e)}
        >
          <Music size={18} />
        </DockButton>
      {/if}

      {#if $pageVisibilityStore.notifications}
        <DockButton
          tooltip="Notifications"
          active={activePage === NotificationPage}
          badge={$notifData.length > 0 ? $notifData.length : undefined}
          onclick={(e: MouseEvent) => AppControls.goNotif(navigate, e)}
        >
          <Bell size={18} />
        </DockButton>
      {/if}

      <DockButton
        tooltip="Settings"
        active={activePage === SettingsPage &&
          $activeSettingsTabStore !== 'about'}
        onclick={(e: MouseEvent) => AppControls.goSettings(navigate, e)}
      >
        <Settings size={18} />
      </DockButton>

      <DockButton
        tooltip="About"
        active={activePage === SettingsPage &&
          $activeSettingsTabStore === 'about'}
        onclick={(e: MouseEvent) => AppControls.goAbout(navigate, e)}
      >
        <Info size={18} />
      </DockButton>
    </div>
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
    gap: 8px;
    padding: 2px 6px;
    pointer-events: auto;
  }

  .dock-group {
    display: flex;
    align-items: center;
    gap: 6px;
  }
</style>
