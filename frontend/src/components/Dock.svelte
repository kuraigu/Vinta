<!--
  @component Dock
  @description Bottom or top launcher bar. Displays pinned application icons with
  smooth hover magnification, active dot indicators for running background tasks,
  and quick navigation clicks to switch island pages.
-->
<script lang="ts">
  import { getContext, onMount } from 'svelte'
  import {
    Grid,
    House,
    Calendar,
    FileText,
    Calculator,
    Timer,
    Music,
    Bell,
    Settings,
    Folder,
    FolderDown,
    Activity
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
  import PerformancePage from '../pages/PerformancePage.svelte'
  import FileShelfPage from '../pages/FileShelfPage.svelte'
  import { fileShelfStore } from '../stores/fileShelf'
  import type { IslandConfig } from '../store'
  import {
    notifData,
    pageVisibilityStore,
    dockPinnedAppsStore,
    globalMediaStore,
    timerAppStore,
    isAnyTimerRunning,
    tStore
  } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let {
    activePage,
    islandConfig
  }: { activePage: unknown; islandConfig?: IslandConfig } = $props()

  interface DockItem {
    id: string
    tooltip: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon: any
    active: boolean
    badge?: number
    group: 'pinned' | 'dynamic'
    onclick: (e: MouseEvent) => void
  }

  let isFolderOpen = $state(false)
  let folderWrapperRef = $state<HTMLDivElement | null>(null)

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
   * Evaluates whether the File Shelf should appear dynamically when unpinned and holding items.
   */
  const checkShowFileShelfDynamic = () =>
    Boolean(
      !$dockPinnedAppsStore.fileshelf &&
      $pageVisibilityStore.fileshelf !== false &&
      ($fileShelfStore.items.length > 0 || activePage === FileShelfPage)
    )
  const showFileShelfDynamic = $derived.by(checkShowFileShelfDynamic)

  /**
   * Evaluates whether the Performance monitor should appear dynamically when active.
   */
  const checkShowPerformanceDynamic = () =>
    Boolean(
      $pageVisibilityStore.performance !== false &&
      activePage === PerformancePage
    )
  const showPerformanceDynamic = $derived.by(checkShowPerformanceDynamic)

  /**
   * Builds the list of middle dock items (pinned favorites and dynamic tasks).
   */
  function computeMiddleItems(): DockItem[] {
    const items: DockItem[] = []

    if (
      $dockPinnedAppsStore.calendar &&
      $pageVisibilityStore.calendar !== false
    ) {
      items.push({
        id: 'calendar',
        tooltip: $tStore('dock.calendar'),
        icon: Calendar,
        active: activePage === CalendarPage,
        group: 'pinned',
        onclick: (e: MouseEvent) => AppControls.goCalendar(navigate, e)
      })
    }

    if ($dockPinnedAppsStore.notes && $pageVisibilityStore.notes !== false) {
      items.push({
        id: 'notes',
        tooltip: $tStore('dock.notes'),
        icon: FileText,
        active: activePage === NotesPage,
        group: 'pinned',
        onclick: (e: MouseEvent) => AppControls.goNotes(navigate, e)
      })
    }

    if (
      $dockPinnedAppsStore.calculator &&
      $pageVisibilityStore.calculator !== false
    ) {
      items.push({
        id: 'calculator',
        tooltip: $tStore('dock.calculator'),
        icon: Calculator,
        active: activePage === CalculatorPage,
        group: 'pinned',
        onclick: (e: MouseEvent) => AppControls.goCalculator(navigate, e)
      })
    }

    if ($dockPinnedAppsStore.timer && $pageVisibilityStore.timer !== false) {
      items.push({
        id: 'timer',
        tooltip: $tStore('dock.timer'),
        icon: Timer,
        active: activePage === TimerPage,
        group: 'pinned',
        onclick: (e: MouseEvent) => AppControls.goTimer(navigate, e)
      })
    }

    if (
      $dockPinnedAppsStore.fileshelf &&
      $pageVisibilityStore.fileshelf !== false
    ) {
      items.push({
        id: 'fileshelf',
        tooltip: $tStore('dock.file_shelf'),
        icon: FolderDown,
        active: activePage === FileShelfPage,
        badge:
          $fileShelfStore.items.length > 0
            ? $fileShelfStore.items.length
            : undefined,
        group: 'pinned',
        onclick: (e: MouseEvent) => AppControls.goFileShelf(navigate, e)
      })
    }

    if (showTimerDynamic) {
      items.push({
        id: 'timer-dynamic',
        tooltip: $tStore('dock.timer'),
        icon: Timer,
        active: activePage === TimerPage,
        group: 'dynamic',
        onclick: (e: MouseEvent) => AppControls.goTimer(navigate, e)
      })
    }

    if (showFileShelfDynamic) {
      items.push({
        id: 'fileshelf-dynamic',
        tooltip: $tStore('dock.file_shelf'),
        icon: FolderDown,
        active: activePage === FileShelfPage,
        badge:
          $fileShelfStore.items.length > 0
            ? $fileShelfStore.items.length
            : undefined,
        group: 'dynamic',
        onclick: (e: MouseEvent) => AppControls.goFileShelf(navigate, e)
      })
    }

    if (showMedia) {
      items.push({
        id: 'media',
        tooltip: $tStore('dock.now_playing'),
        icon: Music,
        active: activePage === MediaPage,
        group: 'dynamic',
        onclick: (e: MouseEvent) => AppControls.goMedia(navigate, e)
      })
    }

    if (showNotifs) {
      items.push({
        id: 'notifications',
        tooltip: $tStore('dock.notifications'),
        icon: Bell,
        active: activePage === NotificationPage,
        badge: $notifData.length > 0 ? $notifData.length : undefined,
        group: 'dynamic',
        onclick: (e: MouseEvent) => AppControls.goNotif(navigate, e)
      })
    }

    if (showPerformanceDynamic) {
      items.push({
        id: 'performance-dynamic',
        tooltip: $tStore('dock.system_stats'),
        icon: Activity,
        active: activePage === PerformancePage,
        group: 'dynamic',
        onclick: (e: MouseEvent) => AppControls.goPerformance(navigate, e)
      })
    }

    return items
  }
  const middleItems = $derived.by(computeMiddleItems)

  /**
   * Computes which items fit inline on the dock and which fold into the Dock Folder
   * when the current page's expanded width is smaller than the total dock width.
   */
  function computeDockLayout(): {
    visibleItems: DockItem[]
    folderItems: DockItem[]
  } {
    const pageW = islandConfig?.expanded?.w ?? 420
    const availableW = pageW - 36
    const totalCount = 3 + middleItems.length
    const hasDyn = middleItems.some((i) => i.group === 'dynamic')
    const dividerAllowance = hasDyn ? 22 : 0
    const requiredFullW = totalCount * 36 + 16 + dividerAllowance

    if (requiredFullW <= availableW) {
      return { visibleItems: middleItems, folderItems: [] }
    }

    // Reserve space for 3 anchor buttons (Apps, Home, Settings) + 1 Folder button = 4 slots
    const maxTotalButtons = Math.max(4, Math.floor((availableW - 16) / 36))
    const maxVisibleMiddle = Math.max(0, maxTotalButtons - 4)

    // Prioritize keeping the currently active page visible on the dock bar if possible
    const activeIdx = middleItems.findIndex((i) => i.active)
    if (activeIdx >= maxVisibleMiddle && maxVisibleMiddle > 0) {
      const reordered = [...middleItems]
      const [activeItem] = reordered.splice(activeIdx, 1)
      reordered.splice(maxVisibleMiddle - 1, 0, activeItem)
      return {
        visibleItems: reordered.slice(0, maxVisibleMiddle),
        folderItems: reordered.slice(maxVisibleMiddle)
      }
    }

    return {
      visibleItems: middleItems.slice(0, maxVisibleMiddle),
      folderItems: middleItems.slice(maxVisibleMiddle)
    }
  }
  const dockLayout = $derived.by(computeDockLayout)
  const visiblePinned = $derived.by(() =>
    dockLayout.visibleItems.filter((i) => i.group === 'pinned')
  )
  const visibleDynamic = $derived.by(() =>
    dockLayout.visibleItems.filter((i) => i.group === 'dynamic')
  )
  const folderItems = $derived.by(() => dockLayout.folderItems)
  const hasFolderActive = $derived.by(() =>
    folderItems.some((item) => item.active)
  )
  const folderBadgeTotal = $derived.by(() =>
    folderItems.reduce((sum, item) => sum + (item.badge || 0), 0)
  )

  function syncFolderState(): void {
    if (folderItems.length === 0 && isFolderOpen) {
      isFolderOpen = false
    }
  }
  $effect(syncFolderState)

  function toggleFolder(e: MouseEvent): void {
    e.stopPropagation()
    isFolderOpen = !isFolderOpen
  }

  function handleFolderItemClick(item: DockItem, e: MouseEvent): void {
    isFolderOpen = false
    item.onclick(e)
  }

  onMount(() => {
    function handleOutsideClick(e: MouseEvent) {
      if (
        isFolderOpen &&
        folderWrapperRef &&
        !folderWrapperRef.contains(e.target as Node)
      ) {
        isFolderOpen = false
      }
    }
    document.addEventListener('mousedown', handleOutsideClick)
    return () => {
      document.removeEventListener('mousedown', handleOutsideClick)
    }
  })
</script>

<div class="dock-container">
  <div class="dock" role="toolbar" aria-label={$tStore('dock.dock_aria')}>
    <!-- Core Anchors -->
    <div class="dock-group core-group">
      <DockButton
        tooltip={$tStore('dock.apps')}
        tooltipPosition="bottom"
        active={activePage === AppsPage}
        onclick={(e: MouseEvent) => AppControls.goApps(navigate, e)}
      >
        <Grid size={18} />
      </DockButton>

      <DockButton
        tooltip={$tStore('dock.home')}
        tooltipPosition="bottom"
        active={activePage === HomePage}
        onclick={(e: MouseEvent) => AppControls.goHome(navigate, e)}
      >
        <House size={18} />
      </DockButton>
    </div>

    <!-- Pinned Favorites -->
    {#if visiblePinned.length > 0}
      <div class="dock-group pinned-group">
        {#each visiblePinned as item (item.id)}
          {@const Icon = item.icon}
          <DockButton
            tooltip={item.tooltip}
            tooltipPosition="bottom"
            active={item.active}
            badge={item.badge}
            onclick={item.onclick}
          >
            <Icon size={18} />
          </DockButton>
        {/each}
      </div>
    {/if}

    <!-- Dynamic Activity Section -->
    {#if visibleDynamic.length > 0}
      <div class="dock-divider" aria-hidden="true"></div>

      <div class="dock-group dynamic-group">
        {#each visibleDynamic as item (item.id)}
          {@const Icon = item.icon}
          <DockButton
            tooltip={item.tooltip}
            tooltipPosition="bottom"
            active={item.active}
            badge={item.badge}
            onclick={item.onclick}
          >
            <Icon size={18} />
          </DockButton>
        {/each}
      </div>

      <div class="dock-divider" aria-hidden="true"></div>
    {/if}

    <!-- Overflow Dock Folder -->
    {#if folderItems.length > 0}
      <div class="dock-folder-wrapper" bind:this={folderWrapperRef}>
        <DockButton
          tooltip={$tStore('dock.more')}
          tooltipPosition="bottom"
          active={isFolderOpen || hasFolderActive}
          badge={!isFolderOpen && folderBadgeTotal > 0
            ? folderBadgeTotal
            : undefined}
          onclick={toggleFolder}
        >
          <Folder size={17} />
        </DockButton>

        {#if isFolderOpen}
          <div
            class="dock-folder-popover"
            role="menu"
            aria-label="Overflow dock items"
          >
            {#each folderItems as item (item.id)}
              {@const Icon = item.icon}
              <DockButton
                tooltip={item.tooltip}
                tooltipPosition="bottom"
                active={item.active}
                badge={item.badge}
                onclick={(e: MouseEvent) => handleFolderItemClick(item, e)}
              >
                <Icon size={18} />
              </DockButton>
            {/each}
          </div>
        {/if}
      </div>
    {/if}

    <!-- Trailing Settings Anchor -->
    <div class="dock-group settings-group">
      <DockButton
        tooltip={$tStore('dock.settings')}
        tooltipPosition="bottom"
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
    top: 12px;
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
    background: rgba(255, 255, 255, 0.07);
    margin: 0 2px;
    flex-shrink: 0;
    opacity: 1;
    box-shadow: none;
  }

  .dock-folder-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }

  .dock-folder-popover {
    position: absolute;
    top: calc(100% + 8px);
    right: -12px;
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 8px;
    border-radius: 14px;
    background: var(--vinta-dock-popover-bg, #141d2c);
    border: 1px solid
      var(--vinta-dock-popover-border, rgba(255, 255, 255, 0.14));
    box-shadow: var(
      --vinta-dock-popover-shadow,
      0 10px 30px rgba(0, 0, 0, 0.65)
    );
    z-index: 220;
    animation: folder-popover-in 0.16s cubic-bezier(0.2, 0.8, 0.2, 1);
  }

  @keyframes folder-popover-in {
    from {
      opacity: 0;
      transform: translateY(-4px) scale(0.96);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
</style>
