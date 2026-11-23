<!--
  @fileoverview Overlay page hosting the searchable Apps launcher on top of the active workspace
  without resetting underlying workspace state.
-->
<script lang="ts">
  import { getContext, onMount } from 'svelte'
  import { Events } from '$lib/events'
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
    ScrollText,
    ArrowUpRight,
    AppWindow,
    Grid,
    Terminal,
    Pin,
    PinOff,
    Shield,
    BellOff,
    Database,
    EyeOff,
    RefreshCw,
    FolderOpen,
    Copy,
    Trash2,
    ExternalLink
  } from 'lucide-svelte'
  import OverlayPage from '../../components/OverlayPage.svelte'
  import HorizontalTabSelection, {
    type TabItem
  } from '../../components/HorizontalTabSelection.svelte'
  import CustomContextMenu, {
    type ContextMenuItem
  } from '../../components/CustomContextMenu.svelte'
  import SearchInput from '../../components/SearchInput.svelte'
  import ScrollView from '../../components/ScrollView.svelte'
  import { islandTooltip } from '../../services/islandTooltip'
  import {
    GetInstalledWindowsApps,
    LaunchWindowsApp,
    CopyShelfTextToClipboard,
    RevealShelfFileInExplorer
  } from '../../../bindings/Vinta/backend/app.js'
  import {
    nativeIconsStore,
    requestNativeAppIcon
  } from '../../services/iconService'
  import {
    evaluateExpression,
    isCalculationError
  } from '../../utils/calculatorEngine'
  import HomePage, { config as homeConfig } from '../HomePage.svelte'
  import MediaPage, { config as mediaConfig } from '../MediaPage.svelte'
  import SettingsPage, {
    config as settingsConfig
  } from '../SettingsPage.svelte'
  import NotesPage, { config as notesConfig } from '../NotesPage.svelte'
  import CalculatorPage, {
    config as calculatorConfig
  } from '../CalculatorPage.svelte'
  import TimerPage, { config as timerConfig } from '../TimerPage.svelte'
  import CalendarPage, {
    config as calendarConfig
  } from '../CalendarPage.svelte'
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
    activeOverlayStore,
    appsOverlayTabStore,
    activeSettingsTabStore,
    pageVisibilityStore,
    updateInfoStore,
    notificationsOverlayTabStore,
    dndStore,
    privacyModeStore,
    launcherUsageStore,
    launcherPinnedIdsStore,
    tStore,
    type IslandConfig,
    type PageVisibility,
    type OverlayPageId,
    type AppsOverlayTab,
    type NotificationsOverlayTab
  } from '../../store'
  import { isEditableEventTarget } from '../../utils/keyboardGuard'

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

  interface DesktopAppItem {
    id: string
    name: string
    target: string
    appId: string
    sourceApp: string
    kind: string
  }

  interface QuickCommandDef {
    id: string
    keywords: string
    titleKey: string
    descKey: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon: any
    action: () => void
  }

  let searchQuery = $state('')
  let selectedIndex = $state(0)
  let desktopApps = $state<DesktopAppItem[]>([])
  let desktopLoading = $state(false)
  let copiedMathFeedback = $state(false)
  let gridHostEl = $state<HTMLDivElement | null>(null)
  let searchInputRef = $state<{ focus: () => void } | null>(null)
  let ctxOpen = $state(false)
  let ctxCoords = $state({ x: 0, y: 0 })
  let ctxTileKey = $state<string | null>(null)

  const launcherTabs = $derived<TabItem<AppsOverlayTab>[]>([
    { id: 'island', label: $tStore('apps.tab_island'), icon: Grid },
    { id: 'desktop', label: $tStore('apps.tab_desktop'), icon: AppWindow }
  ])

  function recordLaunchUsage(tileKey: string) {
    launcherUsageStore.update((prev) => {
      const existing = prev[tileKey]
      return {
        ...prev,
        [tileKey]: {
          count: (existing?.count ?? 0) + 1,
          lastUsedAt: Date.now()
        }
      }
    })
  }

  async function loadDesktopApps(forceRefresh = false) {
    if (desktopLoading) return
    desktopLoading = true
    try {
      const list = await GetInstalledWindowsApps(forceRefresh)
      if (Array.isArray(list)) {
        desktopApps = list
      }
    } catch {
      // ignore
    } finally {
      desktopLoading = false
    }
  }

  onMount(() => {
    const focusSearch = () => {
      selectedIndex = 0
      requestAnimationFrame(() => {
        searchInputRef?.focus()
      })
    }
    focusSearch()

    void loadDesktopApps(false)

    const unsubFocus = Events.OnGlobal('vinta:focus-apps-search', focusSearch)
    return () => {
      unsubFocus()
    }
  })

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
      id: 'audio-mixer',
      nameKey: 'apps.audio_mixer_name',
      descKey: 'apps.audio_mixer_desc',
      icon: SlidersHorizontal,
      visibilityKey: 'media',
      action: () => {
        openWorkspace(MediaPage, mediaConfig)
        setTimeout(() => Events.Emit('vinta:open-audio-mixer'), 60)
      }
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
      nameKey: 'apps.notifs_name',
      descKey: 'apps.notifs_desc',
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

  const quickCommands: QuickCommandDef[] = [
    {
      id: 'dnd',
      keywords: 'dnd do not disturb mute notifications',
      titleKey: 'apps.cmd_dnd',
      descKey: 'apps.cmd_dnd_desc',
      icon: BellOff,
      action: () => {
        dndStore.update((v) => !v)
        onClose()
      }
    },
    {
      id: 'privacy',
      keywords: 'privacy mode screen share hide',
      titleKey: 'apps.cmd_privacy',
      descKey: 'apps.cmd_privacy_desc',
      icon: Shield,
      action: () => {
        privacyModeStore.update((v) => !v)
        onClose()
      }
    },
    {
      id: 'mixer',
      keywords: 'mixer audio volume soundscapes',
      titleKey: 'apps.audio_mixer_name',
      descKey: 'apps.audio_mixer_desc',
      icon: SlidersHorizontal,
      action: () => {
        openWorkspace(MediaPage, mediaConfig)
        setTimeout(() => Events.Emit('vinta:open-audio-mixer'), 60)
      }
    },
    {
      id: 'backup',
      keywords: 'backup restore settings system',
      titleKey: 'apps.cmd_backup',
      descKey: 'apps.cmd_backup_desc',
      icon: Database,
      action: () => {
        activeSettingsTabStore.set('system')
        openWorkspace(SettingsPage, settingsConfig)
      }
    },
    {
      id: 'settings',
      keywords: 'settings preferences config',
      titleKey: 'apps.settings_name',
      descKey: 'apps.settings_desc',
      icon: Settings,
      action: () => openWorkspace(SettingsPage, settingsConfig)
    },
    {
      id: 'hide',
      keywords: 'hide collapse minimize island',
      titleKey: 'apps.cmd_hide',
      descKey: 'apps.cmd_hide_desc',
      icon: EyeOff,
      action: () => {
        onClose()
        Events.Emit('vinta:collapse-island')
      }
    }
  ]

  let resolvedApps = $derived(
    allApps
      .map((app) => ({
        ...app,
        tileKey: `island:${app.id}`,
        name: $tStore(app.nameKey),
        desc: $tStore(app.descKey)
      }))
      .sort((a, b) => a.name.localeCompare(b.name))
  )

  const trimmedQuery = $derived(searchQuery.trim())
  const isCommandMode = $derived(trimmedQuery.startsWith('>'))
  const isWebSearchMode = $derived(trimmedQuery.startsWith('?'))
  const webSearchTerm = $derived(
    isWebSearchMode ? trimmedQuery.slice(1).trim() : ''
  )

  const mathEvaluation = $derived.by(() => {
    if (!trimmedQuery || isCommandMode || isWebSearchMode) return null
    const rawExpr = trimmedQuery.startsWith('=')
      ? trimmedQuery.slice(1).trim()
      : trimmedQuery
    if (!rawExpr) return null
    const looksLikeMath =
      trimmedQuery.startsWith('=') ||
      (/^[0-9(\s]/.test(rawExpr) && /[+\-*/%^()]/.test(rawExpr))
    if (!looksLikeMath) return null
    const res = evaluateExpression(rawExpr)
    if (!res || isCalculationError(res)) return null
    return { expr: rawExpr, result: res }
  })

  const filteredCommands = $derived.by(() => {
    if (!isCommandMode) return []
    const q = trimmedQuery.slice(1).trim().toLowerCase()
    const resolved = quickCommands.map((c) => ({
      ...c,
      title: $tStore(c.titleKey),
      desc: $tStore(c.descKey)
    }))
    if (!q) return resolved
    return resolved.filter(
      (c) =>
        c.title.toLowerCase().includes(q) ||
        c.desc.toLowerCase().includes(q) ||
        c.keywords.toLowerCase().includes(q)
    )
  })

  let filteredApps = $derived.by(() => {
    if (isCommandMode || isWebSearchMode) return []
    const q = trimmedQuery.toLowerCase()
    if (!q) return resolvedApps
    return resolvedApps.filter(
      (a) =>
        a.name.toLowerCase().includes(q) || a.desc.toLowerCase().includes(q)
    )
  })

  let filteredDesktopApps = $derived.by(() => {
    if (isCommandMode || isWebSearchMode) return []
    const q = trimmedQuery.toLowerCase()
    if (!q) return desktopApps
    return desktopApps.filter((a) => a.name.toLowerCase().includes(q))
  })

  const pinnedSet = $derived(new Set($launcherPinnedIdsStore))

  function sortByPinAndUsage<T extends { id: string; name: string }>(
    list: T[],
    prefix: 'island:' | 'desktop:'
  ): T[] {
    return [...list].sort((a, b) => {
      const ka = `${prefix}${a.id}`
      const kb = `${prefix}${b.id}`
      const pa = pinnedSet.has(ka) ? 1 : 0
      const pb = pinnedSet.has(kb) ? 1 : 0
      if (pa !== pb) return pb - pa
      const ua = $launcherUsageStore[ka]?.count ?? 0
      const ub = $launcherUsageStore[kb]?.count ?? 0
      if (ua !== ub) return ub - ua
      const la = $launcherUsageStore[ka]?.lastUsedAt ?? 0
      const lb = $launcherUsageStore[kb]?.lastUsedAt ?? 0
      if (la !== lb) return lb - la
      return a.name.localeCompare(b.name)
    })
  }

  const pinnedIsland = $derived.by(() => {
    if (trimmedQuery) return []
    return sortByPinAndUsage(
      resolvedApps.filter((a) => pinnedSet.has(a.tileKey)),
      'island:'
    )
  })

  const recentIsland = $derived.by(() => {
    if (trimmedQuery) return []
    return sortByPinAndUsage(
      resolvedApps.filter(
        (a) =>
          !pinnedSet.has(a.tileKey) &&
          ($launcherUsageStore[a.tileKey]?.count ?? 0) > 0
      ),
      'island:'
    ).slice(0, 4)
  })

  const pinnedDesktop = $derived.by(() => {
    if (trimmedQuery) return []
    return sortByPinAndUsage(
      desktopApps.filter((a) => pinnedSet.has(`desktop:${a.id}`)),
      'desktop:'
    )
  })

  const recentDesktop = $derived.by(() => {
    if (trimmedQuery) return []
    return sortByPinAndUsage(
      desktopApps.filter((a) => {
        const k = `desktop:${a.id}`
        return !pinnedSet.has(k) && ($launcherUsageStore[k]?.count ?? 0) > 0
      }),
      'desktop:'
    ).slice(0, 4)
  })

  function launchDesktopApp(app: DesktopAppItem) {
    recordLaunchUsage(`desktop:${app.id}`)
    void LaunchWindowsApp(app.target)
    onClose()
    Events.Emit('vinta:collapse-island')
  }

  function launchIslandApp(app: (typeof resolvedApps)[number]) {
    recordLaunchUsage(app.tileKey)
    app.action()
  }

  function getCachedDesktopAppIcon(app: DesktopAppItem): string {
    const k = `${app.sourceApp.toLowerCase()}|${app.appId.toLowerCase()}`
    return $nativeIconsStore[k] || ''
  }

  $effect(function prefetchDesktopIcons() {
    if ($appsOverlayTabStore !== 'desktop') return
    const candidates = trimmedQuery
      ? filteredDesktopApps.slice(0, 24)
      : [...pinnedDesktop, ...recentDesktop, ...desktopApps].slice(0, 32)
    const missing = candidates.filter((app) => !getCachedDesktopAppIcon(app))
    if (missing.length === 0) return

    let cancelled = false
    let idx = 0
    const batchSize = 6
    const step = () => {
      if (cancelled || idx >= missing.length) return
      const end = Math.min(idx + batchSize, missing.length)
      for (let i = idx; i < end; i++) {
        const app = missing[i]
        requestNativeAppIcon(app.sourceApp, app.appId)
      }
      idx = end
      if (idx < missing.length) {
        setTimeout(step, 50)
      }
    }
    step()
    return () => {
      cancelled = true
    }
  })

  function copyMathResult(val: string) {
    void CopyShelfTextToClipboard(val).catch(() =>
      navigator.clipboard?.writeText(val)
    )
    copiedMathFeedback = true
    setTimeout(() => {
      copiedMathFeedback = false
    }, 1200)
  }

  function openWebSearch(term: string) {
    if (!term) return
    window.open(
      `https://www.google.com/search?q=${encodeURIComponent(term)}`,
      '_blank',
      'noopener,noreferrer'
    )
    onClose()
  }

  $effect(function resetAppsSelectionOnFilterChange() {
    void trimmedQuery
    void $appsOverlayTabStore
    selectedIndex = 0
  })

  function scrollSelectedTileIntoView() {
    if (!gridHostEl) return
    const tiles = gridHostEl.querySelectorAll<HTMLElement>('[data-omni-item]')
    tiles[selectedIndex]?.scrollIntoView({ block: 'nearest' })
  }

  function handleTileContextMenu(e: MouseEvent, tileKey: string) {
    e.preventDefault()
    e.stopPropagation()
    ctxTileKey = tileKey
    ctxCoords = { x: e.clientX, y: e.clientY }
    ctxOpen = true
  }

  function getTileContextMenuItems(): ContextMenuItem[] {
    if (!ctxTileKey) return []
    const key = ctxTileKey
    const isPinned = pinnedSet.has(key)
    const hasUsage = ($launcherUsageStore[key]?.count ?? 0) > 0
    const isDesktop = key.startsWith('desktop:')
    const rawId = isDesktop
      ? key.slice('desktop:'.length)
      : key.slice('island:'.length)
    const desktopItem = isDesktop
      ? desktopApps.find((d) => d.id === rawId)
      : undefined
    const islandItem = !isDesktop
      ? resolvedApps.find((a) => a.id === rawId)
      : undefined

    const items: ContextMenuItem[] = []

    if (desktopItem || islandItem) {
      items.push({
        label: $tStore('apps.ctx_open'),
        icon: ExternalLink,
        action: () => {
          if (desktopItem) launchDesktopApp(desktopItem)
          else if (islandItem) launchIslandApp(islandItem)
        }
      })
    }

    items.push({
      label: isPinned ? $tStore('apps.unpin') : $tStore('apps.pin'),
      icon: isPinned ? PinOff : Pin,
      action: () => {
        launcherPinnedIdsStore.update((prev) =>
          isPinned ? prev.filter((k) => k !== key) : [key, ...prev]
        )
      }
    })

    if (hasUsage) {
      items.push({
        label: $tStore('apps.ctx_remove_recent'),
        icon: Trash2,
        action: () => {
          launcherUsageStore.update((prev) => {
            const next = { ...prev }
            delete next[key]
            return next
          })
        }
      })
    }

    if (desktopItem) {
      if (desktopItem.kind === 'shortcut' && desktopItem.target) {
        items.push({
          label: $tStore('apps.ctx_open_location'),
          icon: FolderOpen,
          action: () => {
            void RevealShelfFileInExplorer(desktopItem.target)
          }
        })
      }
      items.push({
        label: $tStore('apps.ctx_copy_name'),
        icon: Copy,
        action: () => {
          void CopyShelfTextToClipboard(desktopItem.name).catch(() =>
            navigator.clipboard?.writeText(desktopItem.name)
          )
        }
      })
    }

    return items
  }

  function handleAppsKeyDown(e: KeyboardEvent) {
    if ($activeOverlayStore !== 'apps') return
    if (e.ctrlKey && !e.altKey && e.key === 'Tab') {
      e.preventDefault()
      $appsOverlayTabStore =
        $appsOverlayTabStore === 'island' ? 'desktop' : 'island'
      return
    }
    if (
      isEditableEventTarget(e.target) &&
      !(
        e.target instanceof HTMLElement &&
        e.target.closest('[data-vinta="apps-overlay-page"]')
      )
    ) {
      return
    }
    const itemEls = gridHostEl
      ? Array.from(gridHostEl.querySelectorAll<HTMLElement>('[data-omni-item]'))
      : []
    const len = itemEls.length

    if (e.key === 'Enter') {
      if (mathEvaluation && !trimmedQuery.startsWith('>')) {
        e.preventDefault()
        copyMathResult(mathEvaluation.result.display)
        return
      }
      if (isWebSearchMode && webSearchTerm) {
        e.preventDefault()
        openWebSearch(webSearchTerm)
        return
      }
      const target = itemEls[selectedIndex] ?? itemEls[0]
      if (target) {
        e.preventDefault()
        target.click()
      }
      return
    }

    if (len === 0) return

    if (e.key === 'ArrowRight') {
      e.preventDefault()
      selectedIndex = (selectedIndex + 1) % len
      requestAnimationFrame(scrollSelectedTileIntoView)
      return
    }
    if (e.key === 'ArrowLeft') {
      e.preventDefault()
      selectedIndex = (selectedIndex - 1 + len) % len
      requestAnimationFrame(scrollSelectedTileIntoView)
      return
    }
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      selectedIndex = Math.min(len - 1, selectedIndex + 4)
      requestAnimationFrame(scrollSelectedTileIntoView)
      return
    }
    if (e.key === 'ArrowUp') {
      e.preventDefault()
      selectedIndex = Math.max(0, selectedIndex - 4)
      requestAnimationFrame(scrollSelectedTileIntoView)
    }
  }
</script>

<svelte:window onkeydown={handleAppsKeyDown} />

<OverlayPage
  dataVinta="apps-overlay-page"
  title={$tStore('apps.title')}
  {onClose}
>
  {#snippet headerLeft()}
    <HorizontalTabSelection
      tabs={launcherTabs}
      bind:value={$appsOverlayTabStore}
      colorVariant="subtle"
      ariaLabel={$tStore('apps.title')}
    />
  {/snippet}

  {#snippet headerActions()}
    <SearchInput
      bind:this={searchInputRef}
      bind:value={searchQuery}
      placeholder={$tStore('apps.search_placeholder')}
      maxWidth="148px"
    />
    {#if $appsOverlayTabStore === 'desktop'}
      <button
        type="button"
        class="launcher-rescan-btn"
        class:is-spinning={desktopLoading}
        disabled={desktopLoading}
        use:islandTooltip={{
          text: $tStore('apps.rescan_desktop'),
          position: 'bottom'
        }}
        aria-label={$tStore('apps.rescan_desktop')}
        onclick={(e) => {
          e.stopPropagation()
          void loadDesktopApps(true)
        }}
      >
        <RefreshCw size={14} strokeWidth={2.1} />
      </button>
    {/if}
  {/snippet}

  <div class="apps-page in-overlay" bind:this={gridHostEl}>
    <ScrollView class="apps-scroll-area">
      {#if mathEvaluation}
        <button
          type="button"
          class="omni-action-card"
          data-omni-item
          onclick={() => copyMathResult(mathEvaluation.result.display)}
        >
          <div class="omni-action-icon">
            <Calculator size={18} />
          </div>
          <div class="omni-action-copy">
            <span class="omni-action-title"
              >= {mathEvaluation.result.display}</span
            >
            <span class="omni-action-sub">
              {copiedMathFeedback
                ? $tStore('common.copied')
                : `${mathEvaluation.expr} · ${$tStore('apps.omni_copy_hint')}`}
            </span>
          </div>
        </button>
      {/if}

      {#if isWebSearchMode}
        {#if webSearchTerm}
          <button
            type="button"
            class="omni-action-card"
            data-omni-item
            onclick={() => openWebSearch(webSearchTerm)}
          >
            <div class="omni-action-icon">
              <Search size={18} />
            </div>
            <div class="omni-action-copy">
              <span class="omni-action-title">
                {$tStore('apps.omni_web_search', { query: webSearchTerm })}
              </span>
              <span class="omni-action-sub">Google</span>
            </div>
            <ArrowUpRight size={15} />
          </button>
        {:else}
          <div class="empty-state">
            <Search size={22} class="empty-icon" />
            <span class="empty-title">{$tStore('apps.omni_web_hint')}</span>
          </div>
        {/if}
      {:else if isCommandMode}
        {#if filteredCommands.length === 0}
          <div class="empty-state">
            <Terminal size={22} class="empty-icon" />
            <span class="empty-title">{$tStore('apps.no_apps_found')}</span>
          </div>
        {:else}
          <div class="omni-command-list">
            {#each filteredCommands as cmd, idx (cmd.id)}
              {@const CmdIcon = cmd.icon}
              <button
                type="button"
                class="omni-action-card"
                class:is-selected={idx === selectedIndex}
                data-omni-item
                onclick={() => cmd.action()}
              >
                <div class="omni-action-icon">
                  <CmdIcon size={17} />
                </div>
                <div class="omni-action-copy">
                  <span class="omni-action-title">{cmd.title}</span>
                  <span class="omni-action-sub">{cmd.desc}</span>
                </div>
              </button>
            {/each}
          </div>
        {/if}
      {:else if trimmedQuery}
        {@const activeFilteredCount =
          $appsOverlayTabStore === 'island'
            ? filteredApps.length
            : filteredDesktopApps.length}
        {#if activeFilteredCount === 0 && !mathEvaluation}
          <div class="empty-state">
            <Search size={22} class="empty-icon" />
            <span class="empty-title">{$tStore('apps.no_apps_found')}</span>
            <span class="empty-desc"
              >{$tStore('apps.no_apps_match', { query: searchQuery })}</span
            >
          </div>
        {:else if $appsOverlayTabStore === 'island' && filteredApps.length > 0}
          <div class="apps-grid">
            {#each filteredApps as app, i (app.id)}
              {@const isEnabled =
                !app.visibilityKey || $pageVisibilityStore[app.visibilityKey]}
              <button
                class="app-tile app-{app.id}"
                class:is-disabled={!isEnabled}
                class:is-selected={i === selectedIndex}
                data-omni-item
                oncontextmenu={(e) => handleTileContextMenu(e, app.tileKey)}
                onclick={(e) => {
                  e.stopPropagation()
                  if (isEnabled) launchIslandApp(app)
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
        {:else if $appsOverlayTabStore === 'desktop' && filteredDesktopApps.length > 0}
          <div class="apps-grid">
            {#each filteredDesktopApps.slice(0, 24) as dApp, j (dApp.id)}
              {@const iconUri = getCachedDesktopAppIcon(dApp)}
              <button
                class="app-tile is-desktop-app"
                class:is-selected={j === selectedIndex}
                data-omni-item
                use:islandTooltip={{ text: dApp.name, position: 'bottom' }}
                oncontextmenu={(e) =>
                  handleTileContextMenu(e, `desktop:${dApp.id}`)}
                onclick={(e) => {
                  e.stopPropagation()
                  launchDesktopApp(dApp)
                }}
              >
                <div class="tile-icon-box desktop-icon-box">
                  {#if iconUri}
                    <img src={iconUri} alt="" class="desktop-app-img" />
                  {:else}
                    <AppWindow size={22} class="desktop-app-fallback" />
                  {/if}
                </div>
                <span class="tile-title">{dApp.name}</span>
              </button>
            {/each}
          </div>
        {/if}
      {:else if $appsOverlayTabStore === 'island'}
        {#if pinnedIsland.length > 0}
          <div class="launcher-section-label">
            {$tStore('apps.section_pinned')}
          </div>
          <div class="apps-grid">
            {#each pinnedIsland as app (app.id)}
              {@const isEnabled =
                !app.visibilityKey || $pageVisibilityStore[app.visibilityKey]}
              <button
                class="app-tile app-{app.id}"
                class:is-disabled={!isEnabled}
                oncontextmenu={(e) => handleTileContextMenu(e, app.tileKey)}
                onclick={(e) => {
                  e.stopPropagation()
                  if (isEnabled) launchIslandApp(app)
                }}
              >
                <div class="tile-icon-box">
                  <app.icon size={24} strokeWidth={2} class="tile-icon" />
                </div>
                <span class="tile-title">{app.name}</span>
              </button>
            {/each}
          </div>
        {/if}

        {#if recentIsland.length > 0}
          <div class="launcher-section-label">
            {$tStore('apps.section_recent')}
          </div>
          <div class="apps-grid">
            {#each recentIsland as app (app.id)}
              {@const isEnabled =
                !app.visibilityKey || $pageVisibilityStore[app.visibilityKey]}
              <button
                class="app-tile app-{app.id}"
                class:is-disabled={!isEnabled}
                oncontextmenu={(e) => handleTileContextMenu(e, app.tileKey)}
                onclick={(e) => {
                  e.stopPropagation()
                  if (isEnabled) launchIslandApp(app)
                }}
              >
                <div class="tile-icon-box">
                  <app.icon size={24} strokeWidth={2} class="tile-icon" />
                </div>
                <span class="tile-title">{app.name}</span>
              </button>
            {/each}
          </div>
        {/if}

        <div class="launcher-section-label">
          {$tStore('apps.section_island_tools')}
        </div>
        <div class="apps-grid">
          {#each resolvedApps as app, i (app.id)}
            {@const isEnabled =
              !app.visibilityKey || $pageVisibilityStore[app.visibilityKey]}
            <button
              class="app-tile app-{app.id}"
              class:is-disabled={!isEnabled}
              class:is-selected={i === selectedIndex}
              data-omni-item
              use:islandTooltip={{
                text: isEnabled
                  ? ''
                  : `${app.desc} (${$tStore('common.disabled')})`,
                position: 'bottom'
              }}
              oncontextmenu={(e) => handleTileContextMenu(e, app.tileKey)}
              onclick={(e) => {
                e.stopPropagation()
                if (isEnabled) {
                  launchIslandApp(app)
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
      {:else}
        {#if desktopLoading && desktopApps.length === 0}
          <div class="empty-state">
            <AppWindow size={22} class="empty-icon" />
            <span class="empty-title">{$tStore('common.loading')}</span>
          </div>
        {:else if desktopApps.length === 0}
          <div class="empty-state">
            <AppWindow size={22} class="empty-icon" />
            <span class="empty-title">{$tStore('apps.no_apps_found')}</span>
          </div>
        {:else}
          {#if pinnedDesktop.length > 0}
            <div class="launcher-section-label">
              {$tStore('apps.section_pinned')}
            </div>
            <div class="apps-grid">
              {#each pinnedDesktop as dApp (dApp.id)}
                {@const iconUri = getCachedDesktopAppIcon(dApp)}
                <button
                  class="app-tile is-desktop-app"
                  use:islandTooltip={{ text: dApp.name, position: 'bottom' }}
                  oncontextmenu={(e) =>
                    handleTileContextMenu(e, `desktop:${dApp.id}`)}
                  onclick={(e) => {
                    e.stopPropagation()
                    launchDesktopApp(dApp)
                  }}
                >
                  <div class="tile-icon-box desktop-icon-box">
                    {#if iconUri}
                      <img src={iconUri} alt="" class="desktop-app-img" />
                    {:else}
                      <AppWindow size={22} class="desktop-app-fallback" />
                    {/if}
                  </div>
                  <span class="tile-title">{dApp.name}</span>
                </button>
              {/each}
            </div>
          {/if}

          {#if recentDesktop.length > 0}
            <div class="launcher-section-label">
              {$tStore('apps.section_recent')}
            </div>
            <div class="apps-grid">
              {#each recentDesktop as dApp (dApp.id)}
                {@const iconUri = getCachedDesktopAppIcon(dApp)}
                <button
                  class="app-tile is-desktop-app"
                  use:islandTooltip={{ text: dApp.name, position: 'bottom' }}
                  oncontextmenu={(e) =>
                    handleTileContextMenu(e, `desktop:${dApp.id}`)}
                  onclick={(e) => {
                    e.stopPropagation()
                    launchDesktopApp(dApp)
                  }}
                >
                  <div class="tile-icon-box desktop-icon-box">
                    {#if iconUri}
                      <img src={iconUri} alt="" class="desktop-app-img" />
                    {:else}
                      <AppWindow size={22} class="desktop-app-fallback" />
                    {/if}
                  </div>
                  <span class="tile-title">{dApp.name}</span>
                </button>
              {/each}
            </div>
          {/if}

          <div class="launcher-section-label">
            {$tStore('apps.section_desktop_apps')}
          </div>
          <div class="apps-grid">
            {#each desktopApps as dApp, i (dApp.id)}
              {@const iconUri = getCachedDesktopAppIcon(dApp)}
              <button
                class="app-tile is-desktop-app"
                class:is-selected={i === selectedIndex}
                data-omni-item
                use:islandTooltip={{ text: dApp.name, position: 'bottom' }}
                oncontextmenu={(e) =>
                  handleTileContextMenu(e, `desktop:${dApp.id}`)}
                onclick={(e) => {
                  e.stopPropagation()
                  launchDesktopApp(dApp)
                }}
              >
                <div class="tile-icon-box desktop-icon-box">
                  {#if iconUri}
                    <img src={iconUri} alt="" class="desktop-app-img" />
                  {:else}
                    <AppWindow size={22} class="desktop-app-fallback" />
                  {/if}
                </div>
                <span class="tile-title">{dApp.name}</span>
              </button>
            {/each}
          </div>
        {/if}
      {/if}
    </ScrollView>
  </div>
</OverlayPage>

{#if ctxOpen}
  <CustomContextMenu
    x={ctxCoords.x}
    y={ctxCoords.y}
    items={getTileContextMenuItems()}
    onclose={() => {
      ctxOpen = false
    }}
  />
{/if}

<style src="../styles/AppsOverlayPage.css"></style>
