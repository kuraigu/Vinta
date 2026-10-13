<script lang="ts">
  import { getContext } from 'svelte'
  import {
    Grid,
    House,
    Calendar,
    FileText,
    Calculator,
    Timer,
    Music,
    Bell,
    Settings
  } from 'lucide-svelte'
  import { AppControls } from '../utils/AppControls.svelte'
  import DockButton from './DockButton.svelte'
  import AppsPage from '../pages/AppsPage.svelte'
  import HomePage from '../pages/HomePage.svelte'
  import CalendarPage from '../pages/CalendarPage.svelte'
  import NotesPage from '../pages/NotesPage.svelte'
  import CalculatorPage from '../pages/CalculatorPage.svelte'
  import TimerPage from '../pages/TimerPage.svelte'
  import MediaPage from '../pages/MediaPage.svelte'
  import NotificationPage from '../pages/NotificationPage.svelte'
  import SettingsPage from '../pages/SettingsPage.svelte'
  import type { IslandConfig } from '../store'
  import {
    notifData,
    pageVisibilityStore,
    dockPinnedAppsStore,
    globalMediaStore,
    timerAppStore,
    isAnyTimerRunning
  } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let { activePage }: { activePage: unknown; islandConfig?: IslandConfig } =
    $props()

  /**
   * Evaluates if any timer, stopwatch, or focus session is currently running.
   */
  const checkTimerRunning = () => isAnyTimerRunning($timerAppStore)
  const isTimerRunning = $derived.by(checkTimerRunning)

  /**
   * Evaluates whether the timer widget should appear dynamically when unpinned.
   */
  const checkShowTimerDynamic = () =>
    Boolean(
      !$dockPinnedAppsStore.timer &&
      $pageVisibilityStore.timer !== false &&
      (isTimerRunning || activePage === TimerPage)
    )
  const showTimerDynamic = $derived.by(checkShowTimerDynamic)

  /**
   * Evaluates whether the now-playing media widget should appear dynamically.
   */
  const checkShowMedia = () =>
    Boolean(
      $pageVisibilityStore.media !== false &&
      (Boolean($globalMediaStore?.isPlaying) || activePage === MediaPage)
    )
  const showMedia = $derived.by(checkShowMedia)

  /**
   * Evaluates whether the notifications widget should appear dynamically.
   */
  const checkShowNotifs = () =>
    Boolean(
      $pageVisibilityStore.notifications !== false &&
      ($notifData.length > 0 || activePage === NotificationPage)
    )
  const showNotifs = $derived.by(checkShowNotifs)

  /**
   * Checks if any dynamic task items are active.
   */
  const checkHasDynamicTasks = () =>
    Boolean(showTimerDynamic || showMedia || showNotifs)
  const hasDynamicTasks = $derived.by(checkHasDynamicTasks)
</script>

<div class="dock-container">
  <div class="dock" role="toolbar" aria-label="Quick Access Dock">
    <!-- Core Anchors -->
    <div class="dock-group core-group">
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
    </div>

    <!-- Pinned Favorites -->
    <div class="dock-group pinned-group">
      {#if $dockPinnedAppsStore.calendar && $pageVisibilityStore.calendar !== false}
        <DockButton
          tooltip="Calendar"
          active={activePage === CalendarPage}
          onclick={(e: MouseEvent) => AppControls.goCalendar(navigate, e)}
        >
          <Calendar size={18} />
        </DockButton>
      {/if}

      {#if $dockPinnedAppsStore.notes && $pageVisibilityStore.notes !== false}
        <DockButton
          tooltip="Notes"
          active={activePage === NotesPage}
          onclick={(e: MouseEvent) => AppControls.goNotes(navigate, e)}
        >
          <FileText size={18} />
        </DockButton>
      {/if}

      {#if $dockPinnedAppsStore.calculator && $pageVisibilityStore.calculator !== false}
        <DockButton
          tooltip="Calculator"
          active={activePage === CalculatorPage}
          onclick={(e: MouseEvent) => AppControls.goCalculator(navigate, e)}
        >
          <Calculator size={18} />
        </DockButton>
      {/if}

      {#if $dockPinnedAppsStore.timer && $pageVisibilityStore.timer !== false}
        <DockButton
          tooltip="Timer"
          active={activePage === TimerPage}
          onclick={(e: MouseEvent) => AppControls.goTimer(navigate, e)}
        >
          <Timer size={18} />
        </DockButton>
      {/if}
    </div>

    <!-- Dynamic Activity Section (Surfaces running Timer, Playing Media, Notifications) -->
    {#if hasDynamicTasks}
      <div class="dock-divider" aria-hidden="true"></div>

      <div class="dock-group dynamic-group">
        {#if showTimerDynamic}
          <DockButton
            tooltip="Timer"
            active={activePage === TimerPage}
            onclick={(e: MouseEvent) => AppControls.goTimer(navigate, e)}
          >
            <Timer size={18} />
          </DockButton>
        {/if}

        {#if showMedia}
          <DockButton
            tooltip="Now Playing"
            active={activePage === MediaPage}
            onclick={(e: MouseEvent) => AppControls.goMedia(navigate, e)}
          >
            <Music size={18} />
          </DockButton>
        {/if}

        {#if showNotifs}
          <DockButton
            tooltip="Notifications"
            active={activePage === NotificationPage}
            badge={$notifData.length > 0 ? $notifData.length : undefined}
            onclick={(e: MouseEvent) => AppControls.goNotif(navigate, e)}
          >
            <Bell size={18} />
          </DockButton>
        {/if}
      </div>

      <div class="dock-divider" aria-hidden="true"></div>
    {/if}

    <!-- Trailing Settings Anchor -->
    <div class="dock-group settings-group">
      <DockButton
        tooltip="Settings"
        active={activePage === SettingsPage}
        onclick={(e: MouseEvent) => AppControls.goSettings(navigate, e)}
      >
        <Settings size={18} />
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
    gap: 6px;
    padding: 2px 6px;
    pointer-events: auto;
  }

  .dock-group {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .dock-divider {
    width: 1px;
    height: 18px;
    background: var(--c-border, rgba(255, 255, 255, 0.12));
    margin: 0 2px;
    flex-shrink: 0;
    opacity: 0.8;
  }
</style>
