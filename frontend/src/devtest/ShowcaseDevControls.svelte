<!--
  @component ShowcaseDevControls
  @description DevTest panel for controlling the in-app Virtual Cursor and automated showcase sequence.
  Implements the choreographed 13-Scene Narrative Adaptive Island showcase with pre-warming and clean-up.

  Used for promotional purposes.
-->
<script lang="ts">
  import { onMount } from 'svelte'
  import { Events } from '$lib/events'
  import {
    Play,
    Square,
    MousePointer,
    RotateCcw,
    Clock,
    CloudSun,
    Calculator,
    Palette,
    Film,
    Sparkles,
    Calendar,
    Timer,
    Bell,
    FolderDown,
    FileText,
    Grid,
    Monitor,
    Wind,
    MessageSquare,
    Zap,
    Trash2
  } from 'lucide-svelte'
  import {
    themeStore,
    activeSettingsTabStore,
    localeStore,
    calendarStore,
    displayLayoutStore,
    hiddenPillOffsetXStore,
    timerAppStore,
    fileShelfStore,
    clearFileShelf,
    notifData,
    dndStore,
    notifAutoWakeStore,
    hasCompletedOnboardingStore,
    devPinnedStore,
    globalMediaStore,
    dockPinnedAppsStore,
    defaultDockPinnedApps,
    simulatedClockStore
  } from '../store'
  import type { MonitorInfo } from '../../bindings/Vinta/backend/models.js'
  import {
    setActiveTimerTab,
    dismissRingingAlarm,
    cleanShowcaseAlarms
  } from '../stores/timerApp'
  import { suppressExternalNotifications } from '../stores/system'
  import { stopAlarmRingingChime } from '../services/audioService'
  import {
    ListNotes,
    DeleteNote,
    SendMediaControl
  } from '../../bindings/Vinta/backend/app.js'

  const ALEX_AVATAR =
    'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><defs><linearGradient id="g" x1="0%" y1="0%" x2="100%" y2="100%"><stop offset="0%" stop-color="%235865F2"/><stop offset="100%" stop-color="%23EB459E"/></linearGradient></defs><rect width="100" height="100" rx="50" fill="url(%23g)"/><circle cx="50" cy="38" r="18" fill="%23ffffff"/><path d="M22 84c0-15.5 12.5-28 28-28s28 12.5 28 28z" fill="%23ffffff"/></svg>'

  const ACT_SPEEDS_STORAGE_KEY = 'vinta_showcase_act_speeds'
  const GLOBAL_SPEED_STORAGE_KEY = 'vinta_showcase_global_speed'

  function loadInitialGlobalSpeed(): number {
    try {
      const saved = localStorage.getItem(GLOBAL_SPEED_STORAGE_KEY)
      if (saved) {
        const val = parseFloat(saved)
        if (!isNaN(val) && val > 0) return val
      }
    } catch {
      // ignore
    }
    return 1
  }

  function loadInitialActSpeeds(): Record<number, number> {
    const defaults: Record<number, number> = {
      1: 1,
      2: 1,
      3: 1,
      4: 1,
      5: 1,
      6: 1,
      7: 1,
      8: 1,
      9: 1,
      10: 1,
      11: 1,
      12: 1
    }
    try {
      const saved = localStorage.getItem(ACT_SPEEDS_STORAGE_KEY)
      if (saved) {
        const parsed = JSON.parse(saved)
        if (parsed && typeof parsed === 'object') {
          for (let i = 1; i <= 12; i++) {
            if (
              typeof parsed[i] === 'number' &&
              !isNaN(parsed[i]) &&
              parsed[i] > 0
            ) {
              defaults[i] = parsed[i]
            }
          }
        }
      }
    } catch {
      // ignore
    }
    return defaults
  }

  const CURSOR_VISIBLE_STORAGE_KEY = 'vinta_showcase_cursor_visible'
  const CURSOR_TRAIL_STORAGE_KEY = 'vinta_showcase_cursor_trail'

  function loadInitialCursorVisible(): boolean {
    try {
      const saved = localStorage.getItem(CURSOR_VISIBLE_STORAGE_KEY)
      if (saved !== null) return JSON.parse(saved)
    } catch {}
    return true
  }

  function loadInitialCursorTrail(): boolean {
    try {
      const saved = localStorage.getItem(CURSOR_TRAIL_STORAGE_KEY)
      if (saved !== null) return JSON.parse(saved)
    } catch {}
    return true
  }

  const NOTIF_SENDER_STORAGE_KEY = 'vinta_showcase_notif_sender'
  const NOTIF_MESSAGE_STORAGE_KEY = 'vinta_showcase_notif_message'

  function loadInitialNotifSender(): string {
    try {
      const saved = localStorage.getItem(NOTIF_SENDER_STORAGE_KEY)
      if (saved && saved.trim()) return saved
    } catch {}
    return 'Alex'
  }

  function loadInitialNotifMessage(): string {
    try {
      const saved = localStorage.getItem(NOTIF_MESSAGE_STORAGE_KEY)
      if (saved && saved.trim()) return saved
    } catch {}
    return 'Hey! Have you heard this new track? Definitely check it out!'
  }

  let isSequenceRunning = $state(false)
  let statusText = $state('Ready')
  let isCursorVisible = $state(loadInitialCursorVisible())
  let isMotionTrailEnabled = $state(loadInitialCursorTrail())
  let notifSenderName = $state(loadInitialNotifSender())
  let notifMessageText = $state(loadInitialNotifMessage())
  let activeStepIdx = $state(0)
  let speedMultiplier = $state(loadInitialGlobalSpeed())
  let activeActNumber = $state<number>(0)
  let prewarmDurationMs = $state(800)
  let sequenceAbortController: { abort: boolean } | null = null

  // --- ACTS & PRE-WARMER TOGGLES ---
  let enabledActs = $state<Record<number, boolean>>({
    1: true,
    2: true,
    3: true,
    4: true,
    5: true,
    6: true,
    7: true,
    8: true,
    9: true,
    10: true,
    11: true,
    12: true
  })

  let actSpeedMultipliers = $state<Record<number, number>>(
    loadInitialActSpeeds()
  )

  $effect(function persistShowcaseSpeeds() {
    const actMap: Record<number, number> = {}
    for (let i = 1; i <= 12; i++) {
      actMap[i] = actSpeedMultipliers[i]
    }
    try {
      localStorage.setItem(
        GLOBAL_SPEED_STORAGE_KEY,
        JSON.stringify(speedMultiplier)
      )
      localStorage.setItem(ACT_SPEEDS_STORAGE_KEY, JSON.stringify(actMap))
    } catch {
      // ignore storage quota / sandbox restrictions
    }
  })

  $effect(function persistShowcaseCursor() {
    try {
      localStorage.setItem(
        CURSOR_VISIBLE_STORAGE_KEY,
        JSON.stringify(isCursorVisible)
      )
      localStorage.setItem(
        CURSOR_TRAIL_STORAGE_KEY,
        JSON.stringify(isMotionTrailEnabled)
      )
    } catch {
      // ignore storage quota / sandbox restrictions
    }
  })

  $effect(function persistShowcaseNotif() {
    try {
      localStorage.setItem(NOTIF_SENDER_STORAGE_KEY, notifSenderName)
      localStorage.setItem(NOTIF_MESSAGE_STORAGE_KEY, notifMessageText)
    } catch {
      // ignore storage quota / sandbox restrictions
    }
  })

  /**
   * Computes compound speed multiplier combining per-act multiplier and global sequence speed.
   */
  function getEffectiveMultiplier(actNum?: number): number {
    const targetAct = actNum || activeActNumber
    const actMult = (targetAct > 0 ? actSpeedMultipliers[targetAct] : 1) ?? 1
    const globalMult = speedMultiplier || 1
    return Math.max(0.1, actMult * globalMult)
  }

  function resetAllActSpeeds() {
    for (let i = 1; i <= 12; i++) {
      actSpeedMultipliers[i] = 1
    }
    try {
      localStorage.setItem(
        ACT_SPEEDS_STORAGE_KEY,
        JSON.stringify(actSpeedMultipliers)
      )
    } catch {
      // ignore
    }
  }

  let enabledPrewarmModes = $state<Record<string, boolean>>({
    home: true,
    weather: true,
    media: true,
    notifications: true,
    apps: true,
    notes: true,
    fileshelf: true,
    calculator: true,
    calendar: true,
    timer: true
  })

  const actLabels: Record<number, string> = {
    1: '1. Notification & Avatar',
    2: '2. Media Playback',
    3: '3. Apps Launcher',
    4: '4. Notes Scratchpad',
    5: '5. Multi-Source Shelf',
    6: '6. Tactile Keypad',
    7: '7. Home & Weather',
    8: '8. Calendar Agenda',
    9: '9. Background Alarm',
    10: '10. Settings & Themes',
    11: '11. Live Ringing Alarm',
    12: '12. Outro Desktop Settle'
  }

  function toggleAllActs(enable: boolean) {
    for (let i = 1; i <= 12; i++) {
      enabledActs[i] = enable
    }
  }

  function toggleAllPrewarm(enable: boolean) {
    for (const k of Object.keys(enabledPrewarmModes)) {
      enabledPrewarmModes[k] = enable
    }
  }

  function sleep(ms: number): Promise<void> {
    const mult = getEffectiveMultiplier()
    return new Promise((resolve) => setTimeout(resolve, ms / mult))
  }

  function emitCursorAction(payload: {
    action:
      | 'show'
      | 'hide'
      | 'move'
      | 'click'
      | 'keypad-sequence'
      | 'scroll'
      | 'type'
      | 'reset'
      | 'drag-file'
    selector?: string
    containsText?: string
    keys?: string[]
    xRatio?: number
    yRatio?: number
    duration?: number
    deltaY?: number
    text?: string
    delay?: number
    keyDelay?: number
    holdDurationMs?: number
    holdProgress?: boolean
    fileBadge?: {
      name: string
      source: string
      icon: string
    }
    fromXRatio?: number
    fromYRatio?: number
    toXRatio?: number
    toYRatio?: number
    fromSelector?: string
    toSelector?: string
    pickRandom?: boolean
  }) {
    Events.Emit('devtest-cursor-action', payload)
  }

  /**
   * Dispatches cursor action and awaits completion via devtest-cursor-done event.
   */
  function runCursorAction(payload: {
    action:
      | 'show'
      | 'hide'
      | 'move'
      | 'click'
      | 'keypad-sequence'
      | 'scroll'
      | 'type'
      | 'reset'
      | 'drag-file'
    selector?: string
    containsText?: string
    keys?: string[]
    xRatio?: number
    yRatio?: number
    duration?: number
    deltaY?: number
    text?: string
    delay?: number
    keyDelay?: number
    holdDurationMs?: number
    holdProgress?: boolean
    fileBadge?: {
      name: string
      source: string
      icon: string
    }
    fromXRatio?: number
    fromYRatio?: number
    toXRatio?: number
    toYRatio?: number
    fromSelector?: string
    toSelector?: string
    pickRandom?: boolean
  }): Promise<void> {
    const mult = getEffectiveMultiplier()
    return new Promise((resolve) => {
      let resolved = false
      let unsub: (() => void) | null = null

      let durMs = payload.duration ? payload.duration * 1000 : 700
      if (payload.action === 'keypad-sequence' && payload.keys) {
        durMs = payload.keys.length * 450
      }
      if (payload.action === 'drag-file' && payload.holdProgress) {
        durMs += 1100
      }
      const timeoutMs = (durMs + 3000) / mult
      const timeout = setTimeout(() => {
        if (!resolved) {
          resolved = true
          if (unsub) unsub()
          resolve()
        }
      }, timeoutMs)

      unsub = Events.On('devtest-cursor-done', () => {
        if (!resolved) {
          resolved = true
          clearTimeout(timeout)
          if (unsub) unsub()
          resolve()
        }
      })

      emitCursorAction({
        ...payload,
        duration: payload.duration ? payload.duration / mult : undefined,
        delay: payload.delay ? payload.delay / mult : undefined,
        keyDelay: Math.round(420 / mult),
        holdDurationMs: Math.round(950 / mult)
      })
    })
  }

  function toggleCursor() {
    isCursorVisible = !isCursorVisible
    emitCursorAction({ action: isCursorVisible ? 'show' : 'hide' })
  }

  function resetCursor() {
    isCursorVisible = true
    emitCursorAction({ action: 'reset' })
    if (isMotionTrailEnabled) {
      Events.Emit('devtest-cursor-trail', { enabled: true })
    }
  }

  function toggleMotionTrail() {
    isMotionTrailEnabled = !isMotionTrailEnabled
    Events.Emit('devtest-cursor-trail', { enabled: isMotionTrailEnabled })
  }

  function handleSwitchMonitor(mon: MonitorInfo) {
    const baseCenter =
      $displayLayoutStore.virtualLeft + $displayLayoutStore.primaryCenter
    const monCenter = mon.left + mon.width / 2
    const targetOffset = Math.round(monCenter - baseCenter)
    $hiddenPillOffsetXStore = targetOffset
    $displayLayoutStore = {
      ...$displayLayoutStore,
      currentMonitor: mon
    }
  }

  // --- PRE-WARMER & CLEAN-UP HELPERS ---

  /**
   * Pre-mounts selected pages into memory with settle time so first-render compilation does not drop frames.
   */
  async function runPrewarmer() {
    statusText = 'Pre-warming selected pages into browser memory...'
    $hasCompletedOnboardingStore = true
    hasCompletedOnboardingStore.set(true)
    try {
      localStorage.setItem(
        'vinta_has_completed_onboarding',
        JSON.stringify(true)
      )
    } catch {}

    const allModes = [
      'home',
      'weather',
      'media',
      'notifications',
      'apps',
      'notes',
      'fileshelf',
      'calculator',
      'calendar',
      'timer'
    ]
    const modes = allModes.filter((m) => enabledPrewarmModes[m])

    if (modes.length === 0) {
      statusText = 'No pages selected for pre-warming.'
      return
    }

    for (const m of modes) {
      statusText = `Pre-warming: ${m}...`
      Events.Emit('devtest-navigate', { mode: m, expand: true })
      await sleep(prewarmDurationMs)
    }

    // Return cleanly to minimal home while keeping dev pin active
    statusText = 'Returning to minimal home...'
    Events.Emit('devtest-pin', { pinned: true })
    Events.Emit('devtest-navigate', { mode: 'home', expand: false })
    await sleep(prewarmDurationMs)
    if (isCursorVisible) {
      emitCursorAction({ action: 'show' })
    }
    if (isMotionTrailEnabled) {
      Events.Emit('devtest-cursor-trail', { enabled: true })
    }
    statusText = 'Pre-warming complete. Pages primed in memory.'
  }

  onMount(() => {
    $hasCompletedOnboardingStore = true
    $devPinnedStore = true
    Events.Emit('devtest-pin', { pinned: true })
    suppressExternalNotifications.set(true)
    Events.Emit('devtest-suppress-notifications', { suppress: true })
    if (isCursorVisible) {
      emitCursorAction({ action: 'show' })
    }
    if (isMotionTrailEnabled) {
      Events.Emit('devtest-cursor-trail', { enabled: true })
    }
  })

  /**
   * Cleans up mock calendar events, alarms, notes, and staged files created by showcase.
   */
  async function cleanupShowcaseData() {
    statusText = 'Cleaning up showcase data...'
    suppressExternalNotifications.set(true)
    Events.Emit('devtest-suppress-notifications', { suppress: true })

    // 1. Calendar: filter out demo launch event
    calendarStore.update((s) => ({
      ...s,
      events: s.events.filter(
        (e) =>
          !e.title.toLowerCase().includes('showcase') &&
          !e.title.toLowerCase().includes('launch') &&
          !e.title.toLowerCase().includes('demo')
      )
    }))

    // 2. Alarms & Timers: clear showcase alarm and reset ringing/chimes & active tab
    cleanShowcaseAlarms()
    setActiveTimerTab('timer')
    timerAppStore.update((s) => ({
      ...s,
      activeTab: 'timer',
      ringingAlarm: null,
      focus: {
        ...s.focus,
        isRunning: false,
        timeRemaining: 0
      }
    }))
    Events.Emit('devtest-timer-state', { activeTab: 'timer' })
    try {
      stopAlarmRingingChime()
    } catch {}

    // 3. File Shelf: clear staged files and local storage
    clearFileShelf()
    Events.Emit('devtest-clear-fileshelf', {})
    try {
      localStorage.removeItem('vinta_file_shelf_paths')
    } catch {}

    // 4. Notes: purge generated demo notes from disk
    try {
      const files = await ListNotes().catch(() => [])
      for (const f of files) {
        const fileName =
          typeof f === 'string'
            ? f
            : (f as { fileName?: string })?.fileName || ''
        const title =
          typeof f === 'object' && f && 'title' in f ? String(f.title) : ''
        const lower = (fileName + ' ' + title).toLowerCase()
        if (
          lower.includes('untitled') ||
          lower.includes('vinta') ||
          lower.includes('showcase') ||
          lower.includes('ideas')
        ) {
          if (fileName) await DeleteNote(fileName).catch(() => {})
        }
      }
    } catch {}

    // 5. Notifications: clear all notifications completely
    notifData.set([])
    try {
      localStorage.removeItem('vinta_notifications')
    } catch {}

    // Cross-window sync with overlay window
    resetSimulatedShowcaseTime()
    Events.Emit('devtest-cleanup', {})
    Events.Emit('devtest-clear-notifications', {})
    Events.Emit('devtest-showcase-folders', { visible: false })

    // 6. Reset language, theme, and settings tab
    $localeStore = 'en'
    $themeStore = 'dark-gold'
    $activeSettingsTabStore = 'appearance'
    Events.Emit('devtest-theme', { theme: 'dark-gold' })

    // 7. Reset dock pinned apps to defaults
    $dockPinnedAppsStore = { ...defaultDockPinnedApps }

    // 8. Reset media across all windows
    globalMediaStore.update((m) => ({
      ...m,
      isPlaying: false
    }))
    Events.Emit('devtest-media-state', { isPlaying: false })

    // 9. Ensure DND is off and island is compact on Home
    $dndStore = false
    $notifAutoWakeStore = true
    Events.Emit('devtest-navigate', { mode: 'home', expand: false })
    if (isCursorVisible) {
      emitCursorAction({ action: 'show' })
    }
    if (isMotionTrailEnabled) {
      Events.Emit('devtest-cursor-trail', { enabled: true })
    }

    statusText = 'Showcase data cleanly wiped & reset.'
  }

  /**
   * Overrides digital clocks across all open windows to a fixed showcase time.
   *
   * @param hour - 24-hour hour value (0-23).
   * @param minute - Minute value (0-59).
   * @param second - Second value (default: 0).
   */
  function setSimulatedShowcaseTime(
    hour: number,
    minute: number,
    second: number = 0
  ) {
    const now = new Date()
    const d = new Date(
      now.getFullYear(),
      now.getMonth(),
      now.getDate(),
      hour,
      minute,
      second,
      0
    )
    simulatedClockStore.set(d)
    Events.Emit('devtest-clock', { time: d.getTime() })
  }

  /**
   * Restores digital clocks across all open windows back to live system time.
   */
  function resetSimulatedShowcaseTime() {
    simulatedClockStore.set(null)
    Events.Emit('devtest-clock', { time: null })
  }

  // --- INDIVIDUAL SCENE ACTIONS ---

  /**
   * Smoothly glides the virtual cursor to the Control Dock satellite's minimize button
   * on the right flank of the island, and clicks it to collapse the expanded page.
   */
  async function minimizeViaControlDock(contextLabel: string) {
    statusText = `[${contextLabel}] Gliding to Control Dock minimize button...`
    await sleep(350)
    await runCursorAction({
      action: 'click',
      selector:
        '.floating-controls button:has(svg.lucide-minimize-2), .floating-controls .button-wrapper:first-child button, [data-satellite-host="controls"] button:first-child, .floating-controls button',
      duration: 0.4
    })
    await sleep(150)
    const targetMode = $globalMediaStore.isPlaying ? 'media' : 'home'
    Events.Emit('devtest-navigate', { mode: targetMode, expand: false })
    await sleep(250)
  }

  /**
   * Act 1: Discord Notification -> Minimize to Media Pill
   */
  async function triggerAct1Notification() {
    activeActNumber = 1
    $dndStore = false
    $notifAutoWakeStore = true
    $hasCompletedOnboardingStore = true

    // Clear prior notifications so inbox is completely pristine
    notifData.set([])
    try {
      localStorage.removeItem('vinta_notifications')
    } catch {}
    Events.Emit('devtest-clear-notifications', {})
    await sleep(150)

    // Dispatch the notification with customizable Discord sender and copy
    const notifId = Date.now()
    const sender = notifSenderName.trim() || 'Alex'
    const message =
      notifMessageText.trim() ||
      'Hey! Have you heard this new track? Definitely check it out!'
    Events.Emit('notification', {
      id: notifId,
      appName: 'Discord',
      title: sender,
      message: message,
      icon: ALEX_AVATAR,
      heroImage: ALEX_AVATAR,
      isShowcase: true
    })

    // Set island to minimal notification pill
    Events.Emit('devtest-navigate', { mode: 'notifications', expand: false })
    statusText = '[Act 1] Discord notification pill arrives...'
    await sleep(550)

    // Glide virtual cursor to click the notification pill
    statusText = '[Act 1] Gliding to click notification pill...'
    await runCursorAction({
      action: 'click',
      selector:
        '.notification-layout.minimal, .banner-view, .banner-view-container, .island-wrapper',
      duration: 0.42
    })
    await sleep(200)

    // Expand into Notification Page
    Events.Emit('devtest-navigate', { mode: 'notifications', expand: true })
    await sleep(400)

    // Click the expand chevron button on the notification card
    statusText = '[Act 1] Expanding notification card details...'
    await runCursorAction({
      action: 'click',
      selector: '.expand-indicator-btn, .notif-item',
      duration: 0.38
    })
    await sleep(200)

    // Linger on expanded card with subtle cursor inspection of the message text
    statusText = '[Act 1] Viewing expanded notification...'
    await runCursorAction({
      action: 'move',
      selector: '.notif-item-message, .hero-image-wrapper, .notif-item',
      duration: 0.38
    })
    await sleep(350)

    // Click the Control Dock / Floating Controls satellite minimize button to collapse
    await minimizeViaControlDock('Act 1')

    // Transition cleanly into the compact media player pill preserving active track
    globalMediaStore.update((m) => ({
      ...m,
      isAvailable: true
    }))
    Events.Emit('devtest-navigate', { mode: 'media', expand: false })
    statusText = '[Act 1] Minimized into media player pill...'
    await sleep(350)
    statusText = 'Act 1 complete.'
  }

  /**
   * Act 2: Click Minimized Media Pill to Expand & Play
   */
  async function triggerAct2MediaPlay() {
    activeActNumber = 2
    // Context-awareness: ensure media state is compact before expanding
    globalMediaStore.update((m) => ({
      ...m,
      isAvailable: true
    }))
    Events.Emit('devtest-navigate', { mode: 'media', expand: false })
    await sleep(100)

    statusText = '[Act 2] Clicking minimized media pill to expand...'
    await runCursorAction({
      action: 'click',
      selector: '.media-container.minimal, .minimal-layout, .island-wrapper',
      duration: 0.38
    })
    await sleep(180)

    Events.Emit('devtest-navigate', { mode: 'media', expand: true })
    await sleep(400)

    // Rewind active track to start (0:00) before clicking play
    try {
      const activeSession =
        $globalMediaStore.activeSessionId || $globalMediaStore.sourceApp || ''
      SendMediaControl(`seek:0:${activeSession}`)
      SendMediaControl('seek:0')
    } catch {}
    globalMediaStore.update((m) => ({
      ...m,
      positionMs: 0
    }))
    Events.Emit('devtest-media-state', {
      positionMs: 0
    })
    await sleep(100)

    statusText = '[Act 2] Gliding to Play Button...'
    await runCursorAction({
      action: 'click',
      selector: '.hero-play-btn',
      duration: 0.38
    })

    // Send native play command to start playback in Spotify / system media
    try {
      SendMediaControl('play')
    } catch {}

    // Ensure playback active state and rewind to start (0:00) across all windows
    try {
      const activeSession =
        $globalMediaStore.activeSessionId || $globalMediaStore.sourceApp || ''
      SendMediaControl(`seek:0:${activeSession}`)
      SendMediaControl('seek:0')
    } catch {}
    globalMediaStore.update((m) => ({
      ...m,
      isPlaying: true,
      positionMs: 0
    }))
    Events.Emit('devtest-media-state', {
      isPlaying: true,
      positionMs: 0
    })

    statusText = '[Act 2] Hovering cursor over active media player...'
    await sleep(250)
    await runCursorAction({
      action: 'move',
      selector:
        '.waveform-scrub-wrapper, .spectrum-bars-container, .track-title-expanded',
      duration: 0.4
    })
    await sleep(400)
    statusText = 'Act 2 complete. Music playing.'
  }

  /**
   * Act 3: Apps Launcher via Dock & Quick App Search
   */
  async function triggerAct3AppsLauncher() {
    activeActNumber = 3
    statusText = '[Act 3] Gliding directly to Apps Launcher icon on Dock...'
    await runCursorAction({
      action: 'click',
      selector:
        '.dock-btn:has(svg.lucide-grid), .core-group .dock-btn:first-child',
      duration: 0.38
    })
    await sleep(150)

    Events.Emit('devtest-navigate', { mode: 'apps', expand: true })
    await sleep(350)

    statusText = '[Act 3] Searching for "Note"...'
    await runCursorAction({
      action: 'click',
      selector:
        '[data-vinta="apps-page"] input, .vinta-search-input input, .search-input input',
      duration: 0.35
    })
    await sleep(100)
    await runCursorAction({
      action: 'type',
      selector:
        '[data-vinta="apps-page"] input, .vinta-search-input input, .search-input input',
      text: 'Note',
      duration: 0.3,
      delay: 45
    })
    await sleep(250)

    statusText = '[Act 3] Launching Notes App from search results...'
    await runCursorAction({
      action: 'click',
      selector:
        '.app-tile.app-notes, [data-tooltip*="Notes"], .apps-grid .app-tile',
      duration: 0.35
    })
    await sleep(250)
    Events.Emit('devtest-navigate', { mode: 'notes', expand: true })
    await sleep(350)
    statusText = 'Act 3 complete.'
  }

  /**
   * Act 4: Notes Scratchpad & Live Typing
   */
  async function triggerAct4Notes() {
    activeActNumber = 4
    // Context-awareness: ensure Notes page is mounted
    Events.Emit('devtest-navigate', { mode: 'notes', expand: true })
    await sleep(350)

    statusText = '[Act 4] Notes Page opened. Clicking + to create new note...'
    await runCursorAction({
      action: 'click',
      selector: 'button.new-note-action, .board-create-btn, .new-note-action',
      duration: 0.38
    })
    await sleep(300)

    statusText = '[Act 4] Setting note title...'
    await runCursorAction({
      action: 'click',
      selector: '.banner-title-input',
      duration: 0.3
    })
    await sleep(100)
    await runCursorAction({
      action: 'type',
      selector: '.banner-title-input',
      text: 'Vinta Ideas',
      duration: 0.3,
      delay: 38
    })
    await sleep(200)

    statusText = '[Act 4] Live typing note content...'
    await runCursorAction({
      action: 'click',
      selector: '.editor-surface',
      duration: 0.3
    })
    await sleep(100)
    await runCursorAction({
      action: 'type',
      selector: '.editor-surface',
      text: 'Adaptive island for Windows 10/11',
      duration: 0.3,
      delay: 32
    })
    await sleep(350)

    // Click back to display the list of notes with our new note
    statusText = '[Act 4] Clicking back button to return to notes list...'
    await runCursorAction({
      action: 'click',
      selector:
        '.banner-left-group .back-btn, .banner-header .back-btn, button.back-btn',
      duration: 0.35
    })
    await sleep(300)

    statusText = '[Act 4] Reviewing newly created note...'
    await runCursorAction({
      action: 'move',
      selector: '.note-card:first-child, .notes-board-grid .note-card',
      duration: 0.35
    })
    await sleep(250)

    // Click Control Dock minimize button to collapse Notes cleanly
    await minimizeViaControlDock('Act 4')
    statusText = 'Act 4 complete.'
  }

  /**
   * Act 5: File Shelf Multi-Source Drag & Staging
   */
  async function triggerAct5FileShelf() {
    activeActNumber = 5
    // Show simulated Windows desktop folder windows with smooth transition
    Events.Emit('devtest-showcase-folders', {
      visible: true,
      destinationCount: 0
    })
    await sleep(350)

    statusText =
      '[Act 5] Dragging proposal_final.pdf from Downloads window into Vinta Island...'
    await runCursorAction({
      action: 'drag-file',
      holdProgress: true,
      fileBadge: {
        name: 'proposal_final.pdf',
        source: 'from Downloads',
        icon: '📄'
      },
      fromSelector:
        '[data-showcase-folder="downloads"] .folder-file-item.is-primary, [data-showcase-folder="downloads"]',
      fromXRatio: 0.09,
      fromYRatio: 0.45,
      toSelector: '.main-island, .island-wrapper, .dock-container',
      toXRatio: 0.5,
      toYRatio: 0.04,
      duration: 0.75
    })
    await sleep(150)

    // Stage first item and expand open File Shelf page
    Events.Emit('devtest-mock-fileshelf', { singleFile: true })
    statusText = '[Act 5] File dropped into Island. Opening File Shelf page...'
    Events.Emit('devtest-navigate', { mode: 'fileshelf', expand: true })
    await sleep(400)

    statusText =
      '[Act 5] Dragging hero_render_4k.png from Design Assets window into open File Shelf...'
    await runCursorAction({
      action: 'drag-file',
      holdProgress: false,
      fileBadge: {
        name: 'hero_render_4k.png',
        source: 'from Design Assets',
        icon: '🖼️'
      },
      fromSelector:
        '[data-showcase-folder="design"] .folder-file-item.is-primary, [data-showcase-folder="design"]',
      fromXRatio: 0.91,
      fromYRatio: 0.42,
      toSelector: '.file-shelf-page, .main-island',
      toXRatio: 0.5,
      toYRatio: 0.16,
      duration: 0.7
    })
    await sleep(150)

    // Stage both items in store
    fileShelfStore.update((s) => ({
      ...s,
      items: [
        {
          path: 'C:\\Projects\\proposal_final.pdf',
          name: 'proposal_final.pdf',
          size: 2450000,
          extension: 'pdf',
          isDir: false,
          category: 'document',
          addedAt: Date.now()
        },
        {
          path: 'C:\\Projects\\hero_render_4k.png',
          name: 'hero_render_4k.png',
          size: 8900000,
          extension: 'png',
          isDir: false,
          category: 'image',
          addedAt: Date.now()
        }
      ]
    }))
    Events.Emit('devtest-mock-fileshelf', { singleFile: false })
    await sleep(200)

    statusText = '[Act 5] Clicking "Select All" to select all staged files...'
    await runCursorAction({
      action: 'click',
      selector: '[data-action="select-all"], .select-toggle-pill',
      duration: 0.35
    })
    Events.Emit('devtest-fileshelf-select-all', {})
    await sleep(250)

    statusText =
      '[Act 5] Dragging all selected files from File Shelf into Client Deliverables folder...'
    await runCursorAction({
      action: 'drag-file',
      fileBadge: {
        name: '2 files selected',
        source: 'to Client Deliverables',
        icon: '📁'
      },
      fromSelector:
        '.shelf-tile.is-selected, .shelf-tile:first-child, .shelf-tile',
      fromXRatio: 0.5,
      fromYRatio: 0.25,
      toSelector:
        '[data-showcase-folder="destination"] .destination-dropzone-prompt, [data-showcase-folder="destination"]',
      toXRatio: 0.72,
      toYRatio: 0.28,
      duration: 0.75
    })
    Events.Emit('devtest-showcase-folders', {
      visible: true,
      destinationCount: 2
    })
    clearFileShelf()
    Events.Emit('devtest-clear-fileshelf', {})
    statusText = '[Act 5] Files transferred to Client Deliverables...'
    await sleep(400)

    // Hide desktop folders with smooth graceful exit
    Events.Emit('devtest-showcase-folders', { visible: false })
    await sleep(350)
    statusText = 'Act 5 complete.'
  }

  /**
   * Act 6: Tactile Calculator Keypad via Dock
   */
  async function triggerAct6Calculator() {
    activeActNumber = 6
    statusText = '[Act 6] Gliding directly to Calculator icon on Dock...'
    await runCursorAction({
      action: 'click',
      selector: '.dock-btn:has(svg.lucide-calculator)',
      duration: 0.38
    })
    await sleep(150)

    Events.Emit('devtest-navigate', { mode: 'calculator', expand: true })
    await sleep(400)

    statusText = '[Act 6] Physically clicking keypad: 1250 * 1.15 =...'
    await runCursorAction({
      action: 'keypad-sequence',
      keys: ['1', '2', '5', '0', '*', '1', '.', '1', '5', '=']
    })
    await sleep(250)

    statusText = '[Act 6] Holding on resolved calculation (1,437.5)...'
    await runCursorAction({
      action: 'move',
      selector: '.calc-display, .calc-screen, .display-main',
      duration: 0.4
    })
    await sleep(350)

    statusText = 'Act 6 complete.'
  }

  /**
   * Act 7: Home Bento & Mascot Weather Jump via Dock
   */
  async function triggerAct7HomeWeather() {
    activeActNumber = 7
    setSimulatedShowcaseTime(15, 58)
    statusText = '[Act 7] Gliding directly to Home icon on Dock...'
    await runCursorAction({
      action: 'click',
      selector:
        '.dock-btn:has(svg.lucide-house), .core-group .dock-btn:nth-child(2)',
      duration: 0.4
    })
    await sleep(180)

    Events.Emit('devtest-navigate', { mode: 'home', expand: true })
    await sleep(400)

    statusText = '[Act 7] Gliding through spotlight bento card...'
    await runCursorAction({
      action: 'move',
      selector: '.spotlight-card-feed, .social-feed-container > :nth-child(2)',
      duration: 0.4
    })
    await sleep(200)

    statusText = '[Act 7] Hovering weather mascot for reactive bounce...'
    await runCursorAction({
      action: 'move',
      selector: '.weather-art-link, [data-weather-mascot], .home-hero-art',
      duration: 0.35
    })
    await sleep(250)

    statusText = '[Act 7] Clicking weather mascot to open forecast...'
    await runCursorAction({
      action: 'click',
      selector: '.weather-art-link, [data-weather-mascot]',
      duration: 0.35
    })
    await sleep(200)

    Events.Emit('devtest-navigate', { mode: 'weather', expand: true })
    await sleep(400)

    statusText = '[Act 7] Tracing hourly forecast trajectory...'
    await runCursorAction({
      action: 'move',
      selector: '.hourly-item:nth-child(3), .hourly-scroll-container',
      duration: 0.45
    })
    await sleep(300)

    statusText = 'Act 7 complete.'
  }

  /**
   * Act 8: Calendar Agenda via Dock
   */
  async function triggerAct8Calendar() {
    activeActNumber = 8
    setSimulatedShowcaseTime(15, 58)
    // Clean any prior showcase launch events to prevent duplicates across multiple test runs
    calendarStore.update((s) => ({
      ...s,
      events: s.events.filter((e) => !e.title.toLowerCase().includes('launch'))
    }))

    statusText = '[Act 8] Gliding directly to Calendar icon on Dock...'
    await runCursorAction({
      action: 'click',
      selector: '.dock-btn:has(svg.lucide-calendar)',
      duration: 0.38
    })
    await sleep(150)

    Events.Emit('devtest-navigate', { mode: 'calendar', expand: true })
    await sleep(450)

    statusText = '[Act 8] Gliding to explore schedule on a different date...'
    await runCursorAction({
      action: 'click',
      selector: '.days-grid .day-cell:not(.today):not(.empty)',
      pickRandom: true,
      duration: 0.55
    })
    await sleep(900)

    statusText = '[Act 8] Returning cursor to today on calendar grid...'
    await runCursorAction({
      action: 'click',
      selector: '.days-grid .day-cell.today, .day-cell.today',
      duration: 0.5
    })
    await sleep(750)

    statusText = '[Act 8] Clicking Add Event for today...'
    await runCursorAction({
      action: 'click',
      selector: '.tray-add-btn',
      duration: 0.4
    })
    await sleep(350)

    statusText = '[Act 8] Typing event title: Launch Review...'
    await runCursorAction({
      action: 'click',
      selector: '#cal-event-title.form-input, .modal-form input.form-input',
      duration: 0.3
    })
    await sleep(100)
    await runCursorAction({
      action: 'type',
      selector: '#cal-event-title.form-input, .modal-form input.form-input',
      text: 'Launch Review',
      duration: 0.3,
      delay: 38
    })
    await sleep(250)

    statusText = '[Act 8] Submitting new calendar event...'
    await runCursorAction({
      action: 'click',
      selector: 'button.modal-btn-submit',
      duration: 0.35
    })
    await sleep(350)

    statusText = '[Act 8] Reviewing scheduled launch event in agenda...'
    await runCursorAction({
      action: 'move',
      selector: '.agenda-event-card, .calendar-agenda-tray',
      duration: 0.4
    })
    await sleep(300)

    statusText = 'Act 8 complete.'
  }

  /**
   * Act 9: Background Alarm Setup via Apps Page & Timer
   */
  async function triggerAct9TimerAlarm() {
    activeActNumber = 9
    setSimulatedShowcaseTime(15, 59)
    // Clean any prior showcase alarms to prevent duplicates across multiple test runs
    cleanShowcaseAlarms()
    setActiveTimerTab('timer')
    timerAppStore.update((s) => ({
      ...s,
      activeTab: 'timer',
      alarms: s.alarms.filter(
        (a) =>
          !a.label.toLowerCase().includes('review') &&
          !a.label.toLowerCase().includes('launch') &&
          a.time !== '16:00'
      )
    }))
    Events.Emit('devtest-timer-state', { activeTab: 'timer' })

    // Timer is not pinned to the dock by default, so we open the Apps Page to search & launch it
    statusText = '[Act 9] Gliding directly to Apps Page launcher on Dock...'
    await runCursorAction({
      action: 'click',
      selector:
        '.dock-btn:has(svg.lucide-grid), .core-group .dock-btn:first-child',
      duration: 0.38
    })
    await sleep(150)

    Events.Emit('devtest-navigate', { mode: 'apps', expand: true })
    await sleep(350)

    statusText = '[Act 9] Searching for "Timer" in Apps Page...'
    await runCursorAction({
      action: 'click',
      selector:
        '[data-vinta="apps-page"] input, .vinta-search-input input, .search-input input',
      duration: 0.35
    })
    await sleep(100)
    await runCursorAction({
      action: 'type',
      selector:
        '[data-vinta="apps-page"] input, .vinta-search-input input, .search-input input',
      text: 'Timer',
      duration: 0.3,
      delay: 45
    })
    await sleep(250)

    statusText = '[Act 9] Launching Timer from search results...'
    await runCursorAction({
      action: 'click',
      selector:
        '.app-tile.app-timer, [data-tooltip*="Timer"], .apps-grid button.app-timer, .app-tile:has(svg.lucide-timer)',
      duration: 0.35
    })
    await sleep(200)

    Events.Emit('devtest-navigate', { mode: 'timer', expand: true })
    await sleep(450)

    statusText = '[Act 9] Switching from Timer to Alarms tab...'
    await runCursorAction({
      action: 'click',
      selector:
        '.tabs-nav button.tab-btn:has(svg.lucide-bell), .tabs-nav button.tab-btn:nth-child(4), button.tab-btn:has(svg.lucide-bell)',
      duration: 0.38
    })
    setActiveTimerTab('alarms')
    $timerAppStore = { ...$timerAppStore, activeTab: 'alarms' }
    Events.Emit('devtest-timer-state', { activeTab: 'alarms' })
    await sleep(250)

    statusText = '[Act 9] Clicking + New Alarm...'
    await runCursorAction({
      action: 'click',
      selector:
        '[data-action="create-alarm"], .app-header button.header-tool-btn, .empty-state-action',
      duration: 0.35
    })
    Events.Emit('devtest-open-new-alarm', { time: '16:00' })
    await sleep(250)

    statusText = '[Act 9] Entering alarm label: 16:00 Launch Review...'
    await runCursorAction({
      action: 'click',
      selector: '.alarm-label-input',
      duration: 0.3
    })
    await sleep(100)
    await runCursorAction({
      action: 'type',
      selector: '.alarm-label-input',
      text: '16:00 Launch Review',
      duration: 0.3,
      delay: 35
    })
    await sleep(200)

    statusText = '[Act 9] Saving and arming new alarm...'
    await runCursorAction({
      action: 'click',
      selector: 'button.drawer-confirm-btn',
      duration: 0.35
    })
    await sleep(300)

    statusText = '[Act 9] 16:00 Launch Review alarm armed and ready...'
    await runCursorAction({
      action: 'move',
      selector: '.alarm-item-card, .alarms-list-container, .alarms-panel',
      duration: 0.4
    })
    await sleep(300)

    statusText = 'Act 9 complete.'
  }

  /**
   * Act 10: Settings via Dock, Multi-Language & Built-in Themes
   */
  async function triggerAct10SettingsThemes() {
    activeActNumber = 10
    setSimulatedShowcaseTime(15, 59)
    statusText = '[Act 10] Gliding directly to Settings icon on Dock...'
    await runCursorAction({
      action: 'click',
      selector:
        '.dock-group.settings-group .dock-btn, .dock-btn:has(svg.lucide-settings)',
      duration: 0.38
    })
    await sleep(150)

    Events.Emit('devtest-navigate', { mode: 'settings', expand: true })
    await sleep(400)

    // 1. Language localization (Deutsch -> Filipino -> English)
    statusText = '[Act 10] Gliding to Language tab in Settings sidebar...'
    await runCursorAction({
      action: 'click',
      selector:
        '.sidebar-nav button:nth-child(8), button.nav-item:has(svg.lucide-languages)',
      duration: 0.35
    })
    $activeSettingsTabStore = 'language'
    await sleep(250)

    // Language: German (Deutsch)
    statusText = '[Act 10] Opening language dropdown for Deutsch...'
    await runCursorAction({
      action: 'click',
      selector: '.setting-group .select-trigger',
      duration: 0.35
    })
    await sleep(200)

    statusText = '[Act 10] Selecting German (Deutsch)...'
    await runCursorAction({
      action: 'click',
      selector: '.select-dropdown.portaled .option-item',
      containsText: 'Deutsch',
      duration: 0.35
    })
    $localeStore = 'de'
    statusText = '[Act 10] UI localized to German (Deutsch)...'
    await sleep(2000)

    // Language: Filipino
    statusText = '[Act 10] Opening language dropdown for Filipino...'
    await runCursorAction({
      action: 'click',
      selector: '.setting-group .select-trigger',
      duration: 0.35
    })
    await sleep(200)

    statusText = '[Act 10] Selecting Filipino...'
    await runCursorAction({
      action: 'click',
      selector: '.select-dropdown.portaled .option-item',
      containsText: 'Filipino',
      duration: 0.35
    })
    $localeStore = 'fil'
    statusText = '[Act 10] UI localized to Filipino...'
    await sleep(2000)

    // Restore English
    statusText = '[Act 10] Opening language dropdown to restore English...'
    await runCursorAction({
      action: 'click',
      selector: '.setting-group .select-trigger',
      duration: 0.35
    })
    await sleep(200)

    statusText = '[Act 10] Restoring English (US)...'
    await runCursorAction({
      action: 'click',
      selector: '.select-dropdown.portaled .option-item',
      containsText: 'English',
      duration: 0.35
    })
    $localeStore = 'en'
    statusText = '[Act 10] Restored language to English...'
    await sleep(2000)

    // 2. Appearance & Built-in Themes (Nord Frost -> Colorblind Accessible -> Dark Gold)
    statusText = '[Act 10] Switching to Appearance tab in sidebar...'
    await runCursorAction({
      action: 'click',
      selector:
        '.sidebar-nav button:first-child, button.nav-item:has(svg.lucide-palette)',
      duration: 0.35
    })
    $activeSettingsTabStore = 'appearance'
    await sleep(250)

    // Theme: Nord Frost
    statusText = '[Act 10] Selecting Nord Frost theme...'
    await runCursorAction({
      action: 'click',
      selector: '.theme-setting-content .select-trigger',
      duration: 0.35
    })
    await sleep(200)

    await runCursorAction({
      action: 'click',
      selector: '.select-dropdown.portaled .option-item',
      containsText: 'Nord Frost',
      duration: 0.35
    })
    $themeStore = 'nord-frost'
    Events.Emit('devtest-theme', { theme: 'nord-frost' })
    statusText = '[Act 10] Theme changed: Nord Frost...'
    await sleep(2000)

    // Theme: Colorblind Accessible
    statusText = '[Act 10] Selecting Colorblind Accessible theme...'
    await runCursorAction({
      action: 'click',
      selector: '.theme-setting-content .select-trigger',
      duration: 0.35
    })
    await sleep(200)

    await runCursorAction({
      action: 'click',
      selector: '.select-dropdown.portaled .option-item',
      containsText: 'Colorblind',
      duration: 0.35
    })
    $themeStore = 'colorblind-safe'
    Events.Emit('devtest-theme', { theme: 'colorblind-safe' })
    statusText = '[Act 10] Theme changed: Colorblind Accessible...'
    await sleep(2000)

    // Restore Dark Gold
    statusText = '[Act 10] Selecting Dark Gold theme...'
    await runCursorAction({
      action: 'click',
      selector: '.theme-setting-content .select-trigger',
      duration: 0.35
    })
    await sleep(200)

    await runCursorAction({
      action: 'click',
      selector: '.select-dropdown.portaled .option-item',
      containsText: 'Dark Gold',
      duration: 0.35
    })
    $themeStore = 'dark-gold'
    Events.Emit('devtest-theme', { theme: 'dark-gold' })
    statusText = '[Act 10] Restored to Dark Gold luxury amber...'
    await sleep(2000)

    // Click Control Dock minimize button to collapse Settings
    await minimizeViaControlDock('Act 10')
    statusText = 'Act 10 complete.'
  }

  /**
   * Act 11: Live Ringing Alarm Payoff with Sound & Dismiss to Minimized Home
   */
  async function triggerAct11AlarmRinging() {
    activeActNumber = 11
    setSimulatedShowcaseTime(16, 0)

    statusText =
      '[Act 11] 16:00 arrives: background playback fades gracefully...'
    // Duck and pause background media with a smooth musical breath before the alarm sounds
    try {
      SendMediaControl('pause')
    } catch {}
    globalMediaStore.update((m) => ({
      ...m,
      isPlaying: false
    }))
    Events.Emit('devtest-media-state', { isPlaying: false })

    // 400ms musical breath: allows the audio pause and waveforms to settle smoothly
    await sleep(400)

    statusText = '[Act 11] Scheduled 16:00 alarm rings with chime!'
    Events.Emit('devtest-trigger-ringing-alarm', {
      label: '16:00 Launch Review',
      time: '16:00'
    })
    // Rings for 5.0 seconds with sound and rhythmic pulsing
    await sleep(5000)

    statusText = '[Act 11] Gliding to Dismiss (X) button...'
    await runCursorAction({
      action: 'click',
      selector:
        '.banner-action-btn.stop, .mini-btn.ringing-dismiss-btn, button:has(svg.lucide-x)',
      duration: 0.38
    })
    try {
      stopAlarmRingingChime()
      dismissRingingAlarm()
    } catch {}
    try {
      SendMediaControl('pause')
    } catch {}
    Events.Emit('devtest-dismiss-alarm', {})
    Events.Emit('devtest-media-state', { isPlaying: false })
    await sleep(250)

    // Settle cleanly into minimized Home
    Events.Emit('devtest-navigate', { mode: 'home', expand: false })
    statusText =
      '[Act 11] Dismissed alarm. Island safely settled into minimized Home...'
    await sleep(400)
    statusText = 'Act 11 complete.'
  }

  /**
   * Act 12: Outro Snap & Desktop Settle
   */
  async function triggerAct12Outro() {
    activeActNumber = 12
    statusText = '[Act 12] Adaptive Island is resting in minimized Home...'
    await sleep(200)

    statusText = '[Act 12] Virtual cursor glides smoothly to desktop center...'
    await runCursorAction({
      action: 'move',
      xRatio: 0.5,
      yRatio: 0.5,
      duration: 0.45
    })
    await sleep(350)

    await runCursorAction({
      action: 'move',
      xRatio: 0.5,
      yRatio: 0.85,
      duration: 0.5
    })
    await sleep(350)
    if (!isCursorVisible) {
      await runCursorAction({ action: 'hide' })
    }
    statusText = 'Act 12 complete. Master sequence finished.'
  }

  /**
   * MASTER: Runs the choreographed 12-Scene Sequence
   */
  async function runFullMasterSequence() {
    if (isSequenceRunning) return
    isSequenceRunning = true
    suppressExternalNotifications.set(true)
    Events.Emit('devtest-suppress-notifications', { suppress: true })
    const abortCtrl = { abort: false }
    sequenceAbortController = abortCtrl

    try {
      // 1. Wipe all notifications and prior demo data cleanly before proceeding
      await cleanupShowcaseData()
      suppressExternalNotifications.set(true)
      Events.Emit('devtest-suppress-notifications', { suppress: true })
      await sleep(300)

      // 2. Ensure dock has pinned icons visible for realistic dock clicking
      $dockPinnedAppsStore = {
        calendar: true,
        notes: true,
        calculator: true,
        timer: true,
        weather: false,
        fileshelf: true
      }

      // 3. Pre-warm all pages first so Svelte chunks, SVG icons & styles are primed
      await runPrewarmer()
      if (abortCtrl.abort) return

      isCursorVisible = true
      await runCursorAction({ action: 'show' })
      if (isMotionTrailEnabled) {
        Events.Emit('devtest-cursor-trail', { enabled: true })
      }

      // 4. Set initial promotional clock time (15:58 / 3:58 PM) leading into 16:00 alarm
      setSimulatedShowcaseTime(15, 58)

      // 5. 2-second countdown for OBS recording sync
      for (let c = 2; c > 0; c--) {
        if (abortCtrl.abort) return
        statusText = `Starting in ${c}s... Ready OBS Recording!`
        await sleep(1000)
      }

      if (abortCtrl.abort) return
      if (enabledActs[1]) {
        activeStepIdx = 1
        await triggerAct1Notification()
        await sleep(300)
      }

      if (abortCtrl.abort) return
      if (enabledActs[2]) {
        activeStepIdx = 2
        await triggerAct2MediaPlay()
        await sleep(300)
      }

      if (abortCtrl.abort) return
      if (enabledActs[3]) {
        activeStepIdx = 3
        await triggerAct3AppsLauncher()
        await sleep(300)
      }

      if (abortCtrl.abort) return
      if (enabledActs[4]) {
        activeStepIdx = 4
        await triggerAct4Notes()
        await sleep(300)
      }

      if (abortCtrl.abort) return
      if (enabledActs[5]) {
        activeStepIdx = 5
        await triggerAct5FileShelf()
        await sleep(300)
      }

      if (abortCtrl.abort) return
      if (enabledActs[6]) {
        activeStepIdx = 6
        await triggerAct6Calculator()
        await sleep(300)
      }

      if (abortCtrl.abort) return
      if (enabledActs[7]) {
        activeStepIdx = 7
        await triggerAct7HomeWeather()
        await sleep(300)
      }

      if (abortCtrl.abort) return
      if (enabledActs[8]) {
        activeStepIdx = 8
        await triggerAct8Calendar()
        await sleep(300)
      }

      if (abortCtrl.abort) return
      if (enabledActs[9]) {
        activeStepIdx = 9
        await triggerAct9TimerAlarm()
        await sleep(300)
      }

      if (abortCtrl.abort) return
      if (enabledActs[10]) {
        activeStepIdx = 10
        await triggerAct10SettingsThemes()
        await sleep(300)
      }

      if (abortCtrl.abort) return
      if (enabledActs[11]) {
        activeStepIdx = 11
        await triggerAct11AlarmRinging()
        await sleep(300)
      }

      if (abortCtrl.abort) return
      if (enabledActs[12]) {
        activeStepIdx = 12
        await triggerAct12Outro()
      }
    } finally {
      isSequenceRunning = false
      activeActNumber = 0
      suppressExternalNotifications.set(false)
      Events.Emit('devtest-suppress-notifications', { suppress: false })
      sequenceAbortController = null
      statusText = 'Showcase sequence finished.'
    }
  }

  function stopSequence() {
    if (sequenceAbortController) {
      sequenceAbortController.abort = true
    }
    isSequenceRunning = false
    suppressExternalNotifications.set(false)
    Events.Emit('devtest-suppress-notifications', { suppress: false })
    resetSimulatedShowcaseTime()
    statusText = 'Sequence stopped.'
  }
</script>

<div class="showcase-controls">
  <!-- Status Header -->
  <div class="showcase-status-card" class:is-running={isSequenceRunning}>
    <div class="status-top-row">
      <div class="status-badge">
        <Sparkles size={13} />
        <span
          >{isSequenceRunning ? 'Showcase Running' : 'Director Standby'}</span
        >
      </div>
      {#if isSequenceRunning}
        <span class="step-counter">Act {activeStepIdx} / 12</span>
      {/if}
    </div>
    <div class="status-message">{statusText}</div>
  </div>

  <!-- Master Automation Action -->
  <div class="master-action-row">
    {#if !isSequenceRunning}
      <button class="master-btn primary" onclick={runFullMasterSequence}>
        <Film size={15} />
        <span>Start Full 12-Act Narrative Sequence</span>
      </button>
    {:else}
      <button class="master-btn danger" onclick={stopSequence}>
        <Square size={14} />
        <span>Stop Sequence</span>
      </button>
    {/if}
  </div>

  <!-- Pre-Warmer & Clean-up Utility Bar -->
  <div class="maintenance-bar">
    <button
      class="maint-btn"
      disabled={isSequenceRunning}
      onclick={runPrewarmer}
      title="Pre-mount selected pages into memory so screen recording has 0 frame drops"
    >
      <Zap size={13} />
      <span>Pre-Warm Selected</span>
    </button>
    <button
      class="maint-btn"
      class:active={Boolean($simulatedClockStore)}
      disabled={isSequenceRunning}
      onclick={() =>
        $simulatedClockStore
          ? resetSimulatedShowcaseTime()
          : setSimulatedShowcaseTime(16, 0)}
      title="Toggle promotional 16:00 (4:00 PM) clock vs live system clock"
    >
      <Clock size={13} />
      <span
        >{$simulatedClockStore ? 'Clock: 16:00 (Active)' : 'Clock: Live'}</span
      >
    </button>
    <button
      class="maint-btn warning"
      disabled={isSequenceRunning}
      onclick={cleanupShowcaseData}
      title="Accurately remove all mock notes, alarms, calendar agendas and staged files"
    >
      <Trash2 size={13} />
      <span>Clean Up Showcase Data</span>
    </button>
  </div>

  <!-- Sequence & Pre-warm Configuration Section -->
  <div class="config-panel">
    <div class="config-header">
      <span class="config-title">Sequence & Pre-warm Toggles</span>
    </div>

    <!-- Pre-warm Pages Toggle Grid -->
    <div class="toggle-group-title">
      <span>Pre-warm Pages</span>
      <div class="toggle-actions">
        <button class="toggle-action-btn" onclick={() => toggleAllPrewarm(true)}
          >All</button
        >
        <button
          class="toggle-action-btn"
          onclick={() => toggleAllPrewarm(false)}>None</button
        >
      </div>
    </div>

    <!-- Pre-warm Duration Slider & Numeric Input -->
    <div class="prewarm-speed-row">
      <span class="speed-label">Duration:</span>
      <input
        type="range"
        class="prewarm-range-slider"
        min="200"
        max="2500"
        step="50"
        bind:value={prewarmDurationMs}
      />
      <div class="prewarm-input-wrap">
        <input
          type="number"
          class="prewarm-number-input"
          min="200"
          max="2500"
          step="50"
          bind:value={prewarmDurationMs}
        />
        <span class="unit-text">ms</span>
      </div>
    </div>

    <div class="chips-grid">
      {#each Object.keys(enabledPrewarmModes) as mode (mode)}
        <button
          type="button"
          class="chip-toggle"
          class:active={enabledPrewarmModes[mode]}
          onclick={() =>
            (enabledPrewarmModes[mode] = !enabledPrewarmModes[mode])}
        >
          <span class="chip-dot"></span>
          <span class="chip-text">{mode}</span>
        </button>
      {/each}
    </div>

    <!-- Narrative Acts Toggle Grid & Per-Act Speeds -->
    <div class="toggle-group-title" style="margin-top: 10px;">
      <span>Narrative Acts & Speeds</span>
      <div class="toggle-actions">
        <button class="toggle-action-btn" onclick={() => toggleAllActs(true)}
          >All</button
        >
        <button class="toggle-action-btn" onclick={() => toggleAllActs(false)}
          >None</button
        >
        <button
          class="toggle-action-btn"
          onclick={resetAllActSpeeds}
          title="Reset all act speeds to 1.0x">Reset 1.0x</button
        >
      </div>
    </div>
    <div class="acts-checklist">
      {#each [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12] as actNum (actNum)}
        <div class="act-card-row" class:is-active={activeActNumber === actNum}>
          <div class="act-card-top">
            <label class="act-checkbox-row">
              <input type="checkbox" bind:checked={enabledActs[actNum]} />
              <span class="act-label">{actLabels[actNum]}</span>
            </label>
            <button
              type="button"
              class="act-speed-badge"
              class:modified={actSpeedMultipliers[actNum] !== 1}
              title="Click to reset to 1.0x"
              onclick={() => (actSpeedMultipliers[actNum] = 1)}
            >
              {actSpeedMultipliers[actNum].toFixed(1)}x
            </button>
          </div>
          <div class="act-speed-slider-wrap">
            <input
              type="range"
              min="0.3"
              max="2.5"
              step="0.1"
              class="act-range-slider"
              bind:value={actSpeedMultipliers[actNum]}
              title={`Speed multiplier for ${actLabels[actNum]}: ${actSpeedMultipliers[actNum]}x`}
            />
          </div>
        </div>
      {/each}
    </div>
  </div>

  <!-- Display & Monitor Targeting -->
  {#if $displayLayoutStore.monitors && $displayLayoutStore.monitors.length > 1}
    <div class="monitor-target-bar">
      <div class="bar-label">
        <Monitor size={12} />
        <span>Island Target Display:</span>
      </div>
      <div class="monitor-btn-row">
        {#each $displayLayoutStore.monitors as mon (mon.id)}
          <button
            class="monitor-btn"
            class:active={$displayLayoutStore.currentMonitor?.id === mon.id}
            onclick={() => handleSwitchMonitor(mon)}
          >
            {mon.name || `Display ${mon.id}`}
          </button>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Notification Customizer (Act 1) -->
  <div class="customizer-card">
    <div class="customizer-header">
      <MessageSquare size={13} />
      <span>Act 1 Discord Notification Setup</span>
    </div>
    <div class="customizer-grid">
      <div class="customizer-field sender-field">
        <span class="field-label">Sender:</span>
        <input
          type="text"
          class="customizer-input"
          placeholder="e.g. Alex"
          bind:value={notifSenderName}
        />
      </div>
      <div class="customizer-field message-field">
        <span class="field-label">Message:</span>
        <input
          type="text"
          class="customizer-input wide"
          placeholder="e.g. Hey! Have you heard this new track? Definitely check it out!"
          bind:value={notifMessageText}
        />
      </div>
    </div>
  </div>

  <!-- Virtual Cursor, Motion Blur Trail & Speed Controls -->
  <div class="cursor-bar">
    <button
      class="tool-btn"
      class:active={isCursorVisible}
      onclick={toggleCursor}
    >
      <MousePointer size={13} />
      <span>{isCursorVisible ? 'Hide Cursor' : 'Show Cursor'}</span>
    </button>
    <button class="tool-btn" onclick={resetCursor}>
      <RotateCcw size={13} />
      <span>Center</span>
    </button>
    <button
      class="tool-btn"
      class:active={isMotionTrailEnabled}
      onclick={toggleMotionTrail}
      title="Toggle subtle ghost motion trail for high-speed recordings"
    >
      <Wind size={13} />
      <span>Trail {isMotionTrailEnabled ? 'ON' : 'OFF'}</span>
    </button>
    <button
      class="tool-btn"
      class:active={speedMultiplier !== 1}
      onclick={() =>
        (speedMultiplier =
          speedMultiplier === 1
            ? 0.8
            : speedMultiplier === 0.8
              ? 1.25
              : speedMultiplier === 1.25
                ? 1.5
                : speedMultiplier === 1.5
                  ? 0.6
                  : 1)}
      title="Global speed multiplier (scales all acts proportionally)"
    >
      <Zap size={13} />
      <span>Global {speedMultiplier}x</span>
    </button>
  </div>

  <!-- Individual Scene Triggers -->
  <div class="step-triggers-title">Individual Scene Triggers</div>
  <div class="step-grid">
    <button
      class="step-btn"
      disabled={isSequenceRunning}
      onclick={triggerAct1Notification}
    >
      <MessageSquare size={13} />
      <span>1. Notification & Avatar</span>
    </button>

    <button
      class="step-btn"
      disabled={isSequenceRunning}
      onclick={triggerAct2MediaPlay}
    >
      <Play size={13} />
      <span>2. Media Playback</span>
    </button>

    <button
      class="step-btn"
      disabled={isSequenceRunning}
      onclick={triggerAct3AppsLauncher}
    >
      <Grid size={13} />
      <span>3. Apps Launcher</span>
    </button>

    <button
      class="step-btn"
      disabled={isSequenceRunning}
      onclick={triggerAct4Notes}
    >
      <FileText size={13} />
      <span>4. Notes Scratchpad</span>
    </button>

    <button
      class="step-btn"
      disabled={isSequenceRunning}
      onclick={triggerAct5FileShelf}
    >
      <FolderDown size={13} />
      <span>5. Multi-Source Shelf</span>
    </button>

    <button
      class="step-btn"
      disabled={isSequenceRunning}
      onclick={triggerAct6Calculator}
    >
      <Calculator size={13} />
      <span>6. Tactile Keypad</span>
    </button>

    <button
      class="step-btn"
      disabled={isSequenceRunning}
      onclick={triggerAct7HomeWeather}
    >
      <CloudSun size={13} />
      <span>7. Home & Weather</span>
    </button>

    <button
      class="step-btn"
      disabled={isSequenceRunning}
      onclick={triggerAct8Calendar}
    >
      <Calendar size={13} />
      <span>8. Calendar Agenda</span>
    </button>

    <button
      class="step-btn"
      disabled={isSequenceRunning}
      onclick={triggerAct9TimerAlarm}
    >
      <Timer size={13} />
      <span>9. Background Alarm</span>
    </button>

    <button
      class="step-btn"
      disabled={isSequenceRunning}
      onclick={triggerAct10SettingsThemes}
    >
      <Palette size={13} />
      <span>10. Settings & Themes</span>
    </button>

    <button
      class="step-btn"
      disabled={isSequenceRunning}
      onclick={triggerAct11AlarmRinging}
    >
      <Bell size={13} />
      <span>11. Live Ringing Alarm</span>
    </button>

    <button
      class="step-btn"
      disabled={isSequenceRunning}
      onclick={triggerAct12Outro}
    >
      <Sparkles size={13} />
      <span>12. Outro Desktop Settle</span>
    </button>
  </div>
</div>

<style>
  .showcase-controls {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 6px 0;
  }

  .showcase-status-card {
    background: #090b10;
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    padding: 10px 12px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .showcase-status-card.is-running {
    border-color: #e2a03f;
    background: rgba(226, 160, 63, 0.06);
  }

  .status-top-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .status-badge {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    font-weight: 700;
    color: #e2a03f;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .step-counter {
    font-size: 11px;
    font-weight: 600;
    color: #8b949e;
  }

  .status-message {
    font-size: 12px;
    color: #f0f3f6;
    font-weight: 500;
  }

  .master-action-row {
    display: flex;
  }

  .master-btn {
    width: 100%;
    padding: 10px 14px;
    border-radius: 8px;
    border: none;
    font-size: 13px;
    font-weight: 700;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .master-btn.primary {
    background: #e2a03f;
    color: #0c0e14;
  }

  .master-btn.primary:hover {
    background: #f3b14e;
    transform: translateY(-1px);
  }

  .master-btn.danger {
    background: #ef4444;
    color: #ffffff;
  }

  .maintenance-bar {
    display: flex;
    gap: 6px;
  }

  .maint-btn {
    flex: 1;
    background: #141822;
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #f0f3f6;
    border-radius: 6px;
    padding: 7px 10px;
    font-size: 11px;
    font-weight: 600;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .maint-btn:hover:not(:disabled) {
    background: #1d2331;
    border-color: rgba(255, 255, 255, 0.18);
  }

  .maint-btn.warning:hover:not(:disabled) {
    border-color: #f59e0b;
    color: #f59e0b;
  }

  .maint-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .config-panel {
    background: #0d1017;
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    padding: 10px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .config-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .config-title {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #e2a03f;
  }

  .toggle-group-title {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 11px;
    font-weight: 600;
    color: #8b949e;
  }

  .toggle-actions {
    display: flex;
    gap: 4px;
  }

  .toggle-action-btn {
    background: #141822;
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #8b949e;
    border-radius: 4px;
    padding: 2px 6px;
    font-size: 10px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .toggle-action-btn:hover {
    color: #f0f3f6;
    border-color: rgba(255, 255, 255, 0.2);
  }

  .prewarm-speed-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
    background: #141822;
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 4px;
    padding: 4px 8px;
  }

  .speed-label {
    font-size: 11px;
    font-weight: 600;
    color: #8b949e;
  }

  .prewarm-range-slider {
    flex: 1;
    accent-color: #e2a03f;
    cursor: pointer;
    height: 4px;
  }

  .prewarm-input-wrap {
    display: flex;
    align-items: center;
    gap: 3px;
  }

  .prewarm-number-input {
    width: 48px;
    background: #0d1017;
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 3px;
    color: #f0f3f6;
    font-size: 11px;
    font-weight: 600;
    padding: 2px 4px;
    text-align: right;
  }

  .prewarm-number-input:focus {
    border-color: #e2a03f;
    outline: none;
  }

  .unit-text {
    font-size: 10px;
    color: #8b949e;
  }

  .chips-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
    gap: 5px;
  }

  .chip-toggle {
    display: flex;
    align-items: center;
    gap: 5px;
    background: #141822;
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #8b949e;
    border-radius: 4px;
    padding: 4px 6px;
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
    text-transform: capitalize;
  }

  .chip-toggle:hover {
    border-color: rgba(255, 255, 255, 0.18);
    color: #f0f3f6;
  }

  .chip-toggle.active {
    background: rgba(226, 160, 63, 0.12);
    border-color: #e2a03f;
    color: #f0f3f6;
  }

  .chip-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #484f58;
    transition: background 0.15s ease;
  }

  .chip-toggle.active .chip-dot {
    background: #e2a03f;
  }

  .acts-checklist {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 6px;
  }

  .act-card-row {
    display: flex;
    flex-direction: column;
    gap: 4px;
    background: #141822;
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 4px;
    padding: 5px 6px 4px 6px;
    transition: all 0.15s ease;
  }

  .act-card-row:hover {
    border-color: rgba(255, 255, 255, 0.16);
  }

  .act-card-row.is-active {
    border-color: #e2a03f;
    background: rgba(226, 160, 63, 0.08);
  }

  .act-card-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 4px;
    min-width: 0;
  }

  .act-checkbox-row {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    font-weight: 500;
    color: #f0f3f6;
    cursor: pointer;
    user-select: none;
    min-width: 0;
    flex: 1;
  }

  .act-checkbox-row input[type='checkbox'] {
    accent-color: #e2a03f;
    cursor: pointer;
    flex-shrink: 0;
  }

  .act-label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 11px;
  }

  .act-speed-badge {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #8b949e;
    font-family: monospace;
    font-size: 10px;
    font-weight: 600;
    padding: 1px 4px;
    border-radius: 3px;
    cursor: pointer;
    user-select: none;
    flex-shrink: 0;
    transition: all 0.15s ease;
  }

  .act-speed-badge:hover {
    border-color: rgba(255, 255, 255, 0.2);
    color: #f0f3f6;
  }

  .act-speed-badge.modified {
    color: #e2a03f;
    background: rgba(226, 160, 63, 0.15);
    border-color: rgba(226, 160, 63, 0.35);
  }

  .act-speed-slider-wrap {
    display: flex;
    align-items: center;
    width: 100%;
  }

  .act-range-slider {
    width: 100%;
    height: 4px;
    accent-color: #e2a03f;
    cursor: pointer;
    background: rgba(255, 255, 255, 0.08);
    border-radius: 2px;
  }

  .monitor-target-bar {
    display: flex;
    flex-direction: column;
    gap: 6px;
    background: #0d1017;
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 6px;
    padding: 8px 10px;
  }

  .bar-label {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    font-weight: 600;
    color: #8b949e;
  }

  .monitor-btn-row {
    display: flex;
    gap: 6px;
  }

  .monitor-btn {
    flex: 1;
    background: #141822;
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #f0f3f6;
    border-radius: 4px;
    padding: 5px 8px;
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .monitor-btn:hover {
    background: #1d2331;
    border-color: rgba(255, 255, 255, 0.16);
  }

  .monitor-btn.active {
    background: rgba(226, 160, 63, 0.15);
    border-color: #e2a03f;
    color: #e2a03f;
  }

  .cursor-bar {
    display: flex;
    gap: 6px;
  }

  .tool-btn {
    flex: 1;
    background: #141822;
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #f0f3f6;
    border-radius: 6px;
    padding: 6px 8px;
    font-size: 11px;
    font-weight: 600;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 5px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .tool-btn:hover {
    background: #1d2331;
    border-color: rgba(255, 255, 255, 0.16);
  }

  .tool-btn.active {
    border-color: #e2a03f;
    color: #e2a03f;
  }

  .step-triggers-title {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #8b949e;
    margin-top: 4px;
  }

  .step-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 6px;
  }

  .step-btn {
    background: #141822;
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #f0f3f6;
    border-radius: 6px;
    padding: 8px 10px;
    font-size: 11px;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 6px;
    cursor: pointer;
    text-align: left;
    transition: all 0.15s ease;
  }

  .step-btn:hover:not(:disabled) {
    background: #1d2331;
    border-color: #e2a03f;
  }

  .step-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Act 1 Notification Customizer */
  .customizer-card {
    background: #0d1017;
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    padding: 10px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .customizer-header {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    font-weight: 700;
    color: #e2a03f;
    letter-spacing: 0.3px;
  }

  .customizer-grid {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .customizer-field {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .field-label {
    font-size: 11px;
    color: #8b949e;
    font-weight: 600;
    min-width: 52px;
  }

  .customizer-input {
    flex: 1;
    background: #141822;
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 4px;
    color: #f0f3f6;
    font-size: 11px;
    padding: 5px 8px;
    outline: none;
    transition: border-color 0.15s ease;
  }

  .customizer-input:focus {
    border-color: #e2a03f;
  }
</style>
