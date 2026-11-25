<!--
  @component Dock
  @description Top launcher bar with fixed Core/Settings anchors, a scrollable middle track
  with edge fades, active capsule indicators, and right-click quick actions.
-->
<script lang="ts">
  import { getContext, onMount } from 'svelte'
  import {
    ArrowLeft,
    Grid,
    House,
    Calendar,
    FileText,
    Calculator,
    Timer,
    Music,
    Bell,
    BellOff,
    Settings,
    FolderDown,
    Activity,
    Pin,
    PinOff,
    EyeOff,
    Play,
    Pause,
    SlidersHorizontal,
    Trash2
  } from 'lucide-svelte'
  import { Events } from '$lib/events'
  import { SendMediaControl } from '../../bindings/Vinta/backend/app.js'
  import { AppControls } from '../utils/AppControls.svelte'
  import DockButton from './DockButton.svelte'
  import CustomContextMenu, {
    type ContextMenuItem
  } from './CustomContextMenu.svelte'
  import HomePage from '../pages/HomePage.svelte'
  import CalendarPage from '../pages/CalendarPage.svelte'
  import NotesPage from '../pages/NotesPage.svelte'
  import CalculatorPage from '../pages/CalculatorPage.svelte'
  import TimerPage from '../pages/TimerPage.svelte'
  import MediaPage from '../pages/MediaPage.svelte'
  import Notification from './Notification.svelte'
  import SettingsPage from '../pages/SettingsPage.svelte'
  import PerformancePage from '../pages/PerformancePage.svelte'
  import FileShelfPage from '../pages/FileShelfPage.svelte'
  import { fileShelfStore } from '../stores/fileShelf'
  import {
    notifData,
    dndStore,
    activeOverlayStore,
    pageVisibilityStore,
    dockPinnedAppsStore,
    globalMediaStore,
    timerAppStore,
    isAnyTimerRunning,
    pauseCountdown,
    startCountdown,
    pauseStopwatch,
    startStopwatch,
    pauseFocus,
    startFocus,
    tStore,
    type IslandConfig,
    type DockPinnedApps,
    type PageVisibility
  } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  const openOverlay =
    getContext<(overlay: 'notifications' | 'apps' | null) => void>(
      'openOverlay'
    )
  const closeOverlay = getContext<() => void>('closeOverlay')

  let {
    activePage,
    canGoBack = false,
    onGoBack
  }: {
    activePage: unknown
    islandConfig?: IslandConfig
    canGoBack?: boolean
    onGoBack?: () => void
  } = $props()

  interface DockItem {
    id: string
    tooltip: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon: any
    active: boolean
    badge?: number
    group: 'pinned' | 'dynamic'
    pinKey?: keyof DockPinnedApps
    visibilityKey?: keyof PageVisibility
    onclick: (e: MouseEvent) => void
  }

  let dockEl = $state<HTMLDivElement | null>(null)
  let dockWidth = $state(480)
  let overflowOpen = $state(false)
  let overflowWrapEl = $state<HTMLDivElement | null>(null)
  let ctxOpen = $state(false)
  let ctxCoords = $state({ x: 0, y: 0 })
  let ctxItem = $state<DockItem | null>(null)

  const checkTimerRunning = () => isAnyTimerRunning($timerAppStore)
  const isTimerRunning = $derived.by(checkTimerRunning)

  const showTimerDynamic = $derived(
    Boolean(
      !$dockPinnedAppsStore.timer &&
      $pageVisibilityStore.timer !== false &&
      (isTimerRunning || activePage === TimerPage)
    )
  )

  const showMedia = $derived(
    Boolean(
      $pageVisibilityStore.media !== false &&
      (Boolean($globalMediaStore?.isPlaying) || activePage === MediaPage)
    )
  )

  const showNotifs = $derived(
    Boolean(
      $pageVisibilityStore.notifications !== false &&
      ($notifData.length > 0 ||
        activePage === Notification ||
        $activeOverlayStore === 'notifications')
    )
  )

  const showFileShelfDynamic = $derived(
    Boolean(
      !$dockPinnedAppsStore.fileshelf &&
      $pageVisibilityStore.fileshelf !== false &&
      ($fileShelfStore.items.length > 0 || activePage === FileShelfPage)
    )
  )

  const showPerformanceDynamic = $derived(
    Boolean(
      $pageVisibilityStore.performance !== false &&
      activePage === PerformancePage
    )
  )

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
        pinKey: 'calendar',
        visibilityKey: 'calendar',
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
        pinKey: 'notes',
        visibilityKey: 'notes',
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
        pinKey: 'calculator',
        visibilityKey: 'calculator',
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
        pinKey: 'timer',
        visibilityKey: 'timer',
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
        pinKey: 'fileshelf',
        visibilityKey: 'fileshelf',
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
        pinKey: 'timer',
        visibilityKey: 'timer',
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
        pinKey: 'fileshelf',
        visibilityKey: 'fileshelf',
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
        visibilityKey: 'media',
        onclick: (e: MouseEvent) => AppControls.goMedia(navigate, e)
      })
    }

    if (showNotifs) {
      items.push({
        id: 'notifications',
        tooltip: $tStore('dock.notifications'),
        icon: Bell,
        active:
          activePage === Notification ||
          $activeOverlayStore === 'notifications',
        badge: $notifData.length > 0 ? $notifData.length : undefined,
        group: 'dynamic',
        visibilityKey: 'notifications',
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
        visibilityKey: 'performance',
        onclick: (e: MouseEvent) => AppControls.goPerformance(navigate, e)
      })
    }

    return items
  }

  const middleItems = $derived.by(computeMiddleItems)

  const partitionedMiddle = $derived.by(() => {
    const fixedButtons = (canGoBack ? 3 : 2) + 1
    const fixedW = fixedButtons * 38 + 32
    const availableMiddleW = Math.max(0, dockWidth - fixedW)
    const rawMaxMiddle = Math.max(1, Math.floor((availableMiddleW + 4) / 38))

    if (middleItems.length <= rawMaxMiddle) {
      return { visible: middleItems, overflow: [] as DockItem[] }
    }

    const maxVisible = Math.max(1, Math.floor((availableMiddleW - 44 + 4) / 38))
    const chosenIds: Record<string, boolean> = {}
    let chosenCount = 0

    const pickItem = (id: string) => {
      if (chosenIds[id]) return
      chosenIds[id] = true
      chosenCount++
    }

    for (const item of middleItems) {
      if (chosenCount >= maxVisible) break
      if (item.active) pickItem(item.id)
    }
    for (const item of middleItems) {
      if (chosenCount >= maxVisible) break
      if (item.group === 'dynamic') pickItem(item.id)
    }
    for (const item of middleItems) {
      if (chosenCount >= maxVisible) break
      pickItem(item.id)
    }

    const visible: DockItem[] = []
    const overflow: DockItem[] = []
    for (const item of middleItems) {
      if (chosenIds[item.id]) {
        visible.push(item)
      } else {
        overflow.push(item)
      }
    }
    return { visible, overflow }
  })

  const visiblePinnedItems = $derived(
    partitionedMiddle.visible.filter((i) => i.group === 'pinned')
  )
  const visibleDynamicItems = $derived(
    partitionedMiddle.visible.filter((i) => i.group === 'dynamic')
  )
  const overflowItems = $derived(partitionedMiddle.overflow)
  const overflowHasActive = $derived(overflowItems.some((i) => i.active))

  $effect(function closeOverflowWhenEmpty() {
    if (overflowItems.length === 0 && overflowOpen) {
      overflowOpen = false
    }
  })

  function handleDockItemContextMenu(e: MouseEvent, item: DockItem) {
    e.preventDefault()
    e.stopPropagation()
    ctxItem = item
    ctxCoords = { x: e.clientX, y: e.clientY }
    ctxOpen = true
  }

  function toggleTimerQuickState() {
    const state = $timerAppStore
    if (state.focus.isRunning) {
      pauseFocus()
      return
    }
    if (state.timer.isRunning) {
      pauseCountdown()
      return
    }
    if (state.stopwatch.isRunning) {
      pauseStopwatch()
      return
    }
    if (state.activeTab === 'focus') {
      startFocus()
    } else if (state.activeTab === 'stopwatch') {
      startStopwatch()
    } else {
      startCountdown()
    }
  }

  function getDockContextMenuItems(): ContextMenuItem[] {
    const item = ctxItem
    if (!item) return []
    const menu: ContextMenuItem[] = []

    if (item.id === 'media') {
      const playing = Boolean($globalMediaStore?.isPlaying)
      menu.push(
        {
          label: playing ? $tStore('media.pause') : $tStore('media.play'),
          icon: playing ? Pause : Play,
          action: () => {
            SendMediaControl(playing ? 'pause' : 'play').catch(() => {})
          }
        },
        {
          label: $tStore('dock.open_audio_mixer'),
          icon: SlidersHorizontal,
          action: () => {
            AppControls.goMedia(navigate)
            setTimeout(() => Events.Emit('vinta:open-audio-mixer'), 60)
          }
        }
      )
    } else if (item.id === 'timer' || item.id === 'timer-dynamic') {
      menu.push({
        label: isTimerRunning
          ? $tStore('dock.pause_timer')
          : $tStore('dock.resume_timer'),
        icon: isTimerRunning ? Pause : Play,
        action: toggleTimerQuickState
      })
    } else if (item.id === 'notifications') {
      menu.push({
        label: $dndStore
          ? $tStore('notification.dnd_on')
          : $tStore('notification.dnd_off'),
        icon: $dndStore ? BellOff : Bell,
        action: () => dndStore.update((v) => !v)
      })
      if ($notifData.length > 0) {
        menu.push({
          label: $tStore('notification.clear_all'),
          icon: Trash2,
          danger: true,
          action: () => notifData.set([])
        })
      }
    }

    if (item.pinKey) {
      if (menu.length > 0) {
        menu.push({ label: '', action: () => {}, divider: true })
      }
      const key = item.pinKey
      const isPinned = Boolean($dockPinnedAppsStore[key])
      menu.push({
        label: isPinned
          ? $tStore('dock.unpin_from_dock')
          : $tStore('dock.pin_to_dock'),
        icon: isPinned ? PinOff : Pin,
        action: () => {
          dockPinnedAppsStore.update((prev) => ({
            ...prev,
            [key]: !isPinned
          }))
        }
      })
    }

    if (item.visibilityKey) {
      const visKey = item.visibilityKey
      menu.push({
        label: $tStore('dock.hide_module'),
        icon: EyeOff,
        action: () => {
          pageVisibilityStore.update((prev) => ({
            ...prev,
            [visKey]: false
          }))
        }
      })
    }

    return menu
  }

  onMount(() => {
    let ro: ResizeObserver | null = null
    if (dockEl && typeof ResizeObserver !== 'undefined') {
      dockWidth = dockEl.clientWidth || 480
      ro = new ResizeObserver((entries) => {
        for (const entry of entries) {
          dockWidth = entry.contentRect.width
        }
      })
      ro.observe(dockEl)
    }

    const handleDocPointerDown = (e: PointerEvent) => {
      if (!overflowOpen) return
      const target = e.target as Node | null
      if (target && overflowWrapEl?.contains(target)) return
      if (target instanceof HTMLElement && target.closest('.ctx-menu')) return
      overflowOpen = false
    }

    const handleDocKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && overflowOpen) {
        e.preventDefault()
        overflowOpen = false
      }
    }

    document.addEventListener('pointerdown', handleDocPointerDown)
    document.addEventListener('keydown', handleDocKeyDown)

    const unsubJump = Events.OnGlobal('vinta:dock-jump', (ev: unknown) => {
      const data = ((ev as { data?: unknown })?.data ?? ev) as
        { index?: number } | undefined
      const idx = data?.index
      if (typeof idx !== 'number' || idx < 0) return

      const targets: Array<() => void> = [
        () =>
          $activeOverlayStore === 'apps'
            ? closeOverlay?.()
            : openOverlay?.('apps'),
        () => AppControls.goHome(navigate),
        ...middleItems.map(
          (item) => () => item.onclick(new MouseEvent('click'))
        ),
        () => AppControls.goSettings(navigate)
      ]
      targets[idx]?.()
    })
    return () => {
      ro?.disconnect()
      document.removeEventListener('pointerdown', handleDocPointerDown)
      document.removeEventListener('keydown', handleDocKeyDown)
      unsubJump()
    }
  })
</script>

<div class="dock-container" class:is-media={activePage === MediaPage}>
  <div
    bind:this={dockEl}
    class="dock"
    role="toolbar"
    aria-label={$tStore('dock.dock_aria')}
  >
    <!-- Core Anchors -->
    <div class="dock-group core-group">
      {#if canGoBack}
        <DockButton
          tooltip={$tStore('dock.back')}
          tooltipPosition="bottom"
          onclick={(e: MouseEvent) => {
            e.stopPropagation()
            onGoBack?.()
          }}
        >
          <ArrowLeft size={18} />
        </DockButton>
      {/if}

      <DockButton
        tooltip={$tStore('dock.apps')}
        tooltipPosition="bottom"
        active={$activeOverlayStore === 'apps'}
        onclick={(e: MouseEvent) =>
          AppControls.goApps(
            () =>
              $activeOverlayStore === 'apps'
                ? closeOverlay?.()
                : openOverlay?.('apps'),
            e
          )}
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

    <!-- Middle Items + Labeled Overflow Pill -->
    {#if middleItems.length > 0}
      <div class="dock-middle">
        {#if visiblePinnedItems.length > 0}
          <div class="dock-group pinned-group">
            {#each visiblePinnedItems as item (item.id)}
              {@const Icon = item.icon}
              <DockButton
                tooltip={item.tooltip}
                tooltipPosition="bottom"
                active={item.active}
                badge={item.badge}
                onclick={item.onclick}
                oncontextmenu={(e) => handleDockItemContextMenu(e, item)}
              >
                <Icon size={18} />
              </DockButton>
            {/each}
          </div>
        {/if}

        {#if visiblePinnedItems.length > 0 && visibleDynamicItems.length > 0}
          <div class="dock-divider" aria-hidden="true"></div>
        {/if}

        {#if visibleDynamicItems.length > 0}
          <div class="dock-group dynamic-group">
            {#each visibleDynamicItems as item (item.id)}
              {@const Icon = item.icon}
              <DockButton
                tooltip={item.tooltip}
                tooltipPosition="bottom"
                active={item.active}
                badge={item.badge}
                onclick={item.onclick}
                oncontextmenu={(e) => handleDockItemContextMenu(e, item)}
              >
                <Icon size={18} />
              </DockButton>
            {/each}
          </div>
        {/if}

        {#if overflowItems.length > 0}
          <div class="dock-overflow-wrap" bind:this={overflowWrapEl}>
            <button
              type="button"
              class="dock-overflow-pill"
              class:active={overflowOpen || overflowHasActive}
              aria-expanded={overflowOpen}
              aria-label={$tStore('dock.more')}
              onclick={(e: MouseEvent) => {
                e.stopPropagation()
                overflowOpen = !overflowOpen
              }}
            >
              <span>+{overflowItems.length}</span>
            </button>

            {#if overflowOpen}
              <div class="dock-overflow-popover" role="menu">
                {#each overflowItems as item (item.id)}
                  {@const Icon = item.icon}
                  <button
                    type="button"
                    class="dock-overflow-item"
                    class:active={item.active}
                    role="menuitem"
                    onclick={(e: MouseEvent) => {
                      overflowOpen = false
                      item.onclick(e)
                    }}
                    oncontextmenu={(e) => handleDockItemContextMenu(e, item)}
                  >
                    <Icon size={16} />
                    <span class="dock-overflow-label">{item.tooltip}</span>
                    {#if item.badge && item.badge > 0}
                      <span class="dock-overflow-badge">
                        {item.badge > 99 ? '99+' : item.badge}
                      </span>
                    {/if}
                  </button>
                {/each}
              </div>
            {/if}
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

{#if ctxOpen}
  <CustomContextMenu
    x={ctxCoords.x}
    y={ctxCoords.y}
    items={getDockContextMenuItems()}
    onclose={() => {
      ctxOpen = false
    }}
  />
{/if}

<style src="./styles/Dock.css"></style>
