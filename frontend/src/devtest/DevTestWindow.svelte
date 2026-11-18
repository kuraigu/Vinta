<script lang="ts">
  import { onMount } from 'svelte'
  import { Events } from '$lib/events'
  import {
    Pin,
    Maximize2,
    Minimize2,
    Crosshair,
    Inspect,
    ScrollText,
    Sun,
    Moon,
    RefreshCw,
    Copy,
    Check,
    Eye,
    Terminal
  } from 'lucide-svelte'
  import CustomSelect from '../components/CustomSelect.svelte'
  import CustomTooltip from '../components/CustomTooltip.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import NotificationDevControls from './NotificationDevControls.svelte'
  import HomeDevControls from './HomeDevControls.svelte'
  import WeatherDevControls from './WeatherDevControls.svelte'
  import ShowcaseDevControls from './ShowcaseDevControls.svelte'
  import {
    themeStore,
    devPinnedStore,
    addCalendarEvent,
    syncAllCalendarFeeds
  } from '../store'
  import { applyThemeToDOM, refreshStylesheetsAndTheme } from '../themes/engine'
  import {
    debugBoundsOverlay,
    debugLogs,
    debugClassInspector
  } from './debugConfig'
  import type { InspectorHoverPayload } from './ClassInspectorOverlay.svelte'

  let isExpandedMode = $state(false)
  let targetMode = $state('home')
  let inspectedElement = $state<InspectorHoverPayload | null>(null)
  let copiedSelector = $state(false)

  const urlParams =
    typeof window !== 'undefined'
      ? new URLSearchParams(window.location.search)
      : null
  const isShowcaseParam = urlParams?.get('showcase') === 'true'
  const isDesktopRuntime =
    typeof window !== 'undefined' &&
    Boolean(
      (
        window as unknown as {
          chrome?: { webview?: unknown }
          _wails?: unknown
        }
      ).chrome?.webview || (window as unknown as { _wails?: unknown })._wails
    )
  let activeTab = $state<'controls' | 'showcase'>(
    isShowcaseParam ? 'showcase' : 'controls'
  )

  const pageOptions = [
    { value: 'home', label: 'Home Page' },
    { value: 'weather', label: 'Weather' },
    { value: 'calendar', label: 'Calendar' },
    { value: 'media', label: 'Media Player' },
    { value: 'notifications', label: 'Notifications' },
    { value: 'notes', label: 'Scratchpad' },
    { value: 'apps', label: 'App Launcher' },
    { value: 'settings', label: 'Settings' },
    { value: 'fileshelf', label: 'File Shelf' },
    { value: 'calculator', label: 'Calculator' },
    { value: 'timer', label: 'Timer & Alarm' }
  ]

  function emitNavigate(mode: string, expand: boolean) {
    targetMode = mode
    try {
      Events.Emit('devtest-navigate', { mode, expand })
    } catch {
      // Ignore if runtime events not available
    }
  }

  function handlePageSelect(newMode: string) {
    $devPinnedStore = true
    if (
      newMode === 'fileshelf' ||
      newMode === 'notes' ||
      newMode === 'apps' ||
      newMode === 'settings'
    ) {
      isExpandedMode = true
    }
    emitNavigate(newMode, isExpandedMode)
  }

  function toggleExpandedMode() {
    isExpandedMode = !isExpandedMode
    $devPinnedStore = true
    emitNavigate(targetMode, isExpandedMode)
  }

  /**
   * Refreshes stylesheets and reloads both the main Vinta island and the devtest window.
   */
  async function handleReload() {
    const currentTheme = $themeStore || 'dark-gold'
    try {
      sessionStorage.setItem(
        'vinta_devtest_window_state',
        JSON.stringify({
          mode: targetMode,
          expand: isExpandedMode,
          pinned: $devPinnedStore
        })
      )
    } catch {
      // Ignore storage errors
    }
    await refreshStylesheetsAndTheme(currentTheme)
    try {
      Events.Emit('devtest-reload-css', {
        theme: currentTheme,
        mode: targetMode,
        expand: isExpandedMode,
        pinned: $devPinnedStore
      })
    } catch {
      // Ignore if runtime events not available
    }
    window.setTimeout(() => {
      window.location.reload()
    }, 60)
  }

  function togglePin() {
    $devPinnedStore = !$devPinnedStore
    try {
      Events.Emit('devtest-pin', { pinned: $devPinnedStore })
    } catch {}
  }

  function toggleBoundsOverlay() {
    $debugBoundsOverlay = !$debugBoundsOverlay
    try {
      Events.Emit('devtest-bounds-overlay', { show: $debugBoundsOverlay })
    } catch {}
  }

  function toggleLogs() {
    $debugLogs = !$debugLogs
    try {
      Events.Emit('devtest-logs', { enabled: $debugLogs })
    } catch {}
  }

  function toggleClassInspector() {
    $debugClassInspector = !$debugClassInspector
    if (!$debugClassInspector) {
      inspectedElement = null
    }
    try {
      Events.Emit('devtest-class-inspector', {
        enabled: $debugClassInspector
      })
    } catch {}
  }

  async function copySelector() {
    if (!inspectedElement?.selector) return
    try {
      await navigator.clipboard.writeText(inspectedElement.selector)
      copiedSelector = true
      setTimeout(() => {
        copiedSelector = false
      }, 1500)
    } catch {
      // Ignore clipboard failure
    }
  }

  function toggleTheme() {
    $themeStore = $themeStore === 'dark-gold' ? 'nord-frost' : 'dark-gold'
    try {
      Events.Emit('devtest-theme', { theme: $themeStore })
    } catch {}
  }

  Events.On('devtest-sync-page', (ev: { data: unknown }) => {
    const pageId = (ev?.data as { pageId?: string })?.pageId
    if (pageId && pageId !== targetMode) {
      targetMode = pageId
    }
  })

  Events.On('devtest-inspector-hover', (ev: { data: unknown }) => {
    const data = ev?.data as InspectorHoverPayload | undefined
    if (data && data.tag) {
      inspectedElement = data
    }
  })

  onMount(() => {
    try {
      const savedState = sessionStorage.getItem('vinta_devtest_window_state')
      if (savedState) {
        sessionStorage.removeItem('vinta_devtest_window_state')
        const parsed = JSON.parse(savedState) as {
          mode?: string
          expand?: boolean
          pinned?: boolean
        }
        if (parsed.mode) targetMode = parsed.mode
        if (typeof parsed.expand === 'boolean') isExpandedMode = parsed.expand
        if (typeof parsed.pinned === 'boolean') $devPinnedStore = parsed.pinned
      }
    } catch {
      // Ignore storage errors
    }

    // Emit initial navigation on mount so the main island spawns and syncs with devtest window
    emitNavigate(targetMode, isExpandedMode)
  })

  $effect(function handleDevTestThemeChange() {
    const currentTheme = $themeStore || 'dark-gold'
    applyThemeToDOM(currentTheme)
  })
</script>

<div class="devtest-window">
  <header class="window-header">
    <div class="header-title">
      <span class="badge">DEV</span>
      <h2>Vinta Developer Tools</h2>
      <span
        class="badge runtime-badge"
        class:desktop={isDesktopRuntime}
        use:islandTooltip={{
          text: isDesktopRuntime
            ? 'Running inside desktop WebView2 host'
            : 'Running in external web browser',
          position: 'bottom'
        }}
      >
        {isDesktopRuntime ? 'Desktop' : 'Web'}
      </span>
    </div>

    <div class="header-actions">
      <button
        class="action-btn"
        onclick={() => {
          $devPinnedStore = true
          emitNavigate(targetMode, isExpandedMode)
        }}
        use:islandTooltip={{
          text: 'Show / Wake Island',
          position: 'bottom'
        }}
      >
        <Eye size={14} strokeWidth={2.2} />
      </button>

      <button
        class="action-btn {$devPinnedStore ? 'active' : ''}"
        onclick={togglePin}
        use:islandTooltip={{
          text: $devPinnedStore ? 'Unpin Page' : 'Pin Page',
          position: 'bottom'
        }}
      >
        <Pin size={14} strokeWidth={2.2} />
      </button>

      <button
        class="action-btn {isExpandedMode ? 'active' : ''}"
        onclick={toggleExpandedMode}
        use:islandTooltip={{
          text: isExpandedMode
            ? 'Switch to Minimal Island'
            : 'Switch to Expanded Card',
          position: 'bottom'
        }}
      >
        {#if isExpandedMode}
          <Minimize2 size={14} strokeWidth={2.2} />
        {:else}
          <Maximize2 size={14} strokeWidth={2.2} />
        {/if}
      </button>

      <button
        class="action-btn {$debugBoundsOverlay ? 'active' : ''}"
        onclick={toggleBoundsOverlay}
        use:islandTooltip={{
          text: $debugBoundsOverlay
            ? 'Hide Bounds Overlay'
            : 'Show Bounds Overlay',
          position: 'bottom'
        }}
      >
        <Crosshair size={14} strokeWidth={2.2} />
      </button>

      <button
        class="action-btn {$debugClassInspector ? 'active' : ''}"
        onclick={toggleClassInspector}
        use:islandTooltip={{
          text: $debugClassInspector
            ? 'Disable Class Name Inspector (Ctrl+Shift+I)'
            : 'Enable Class Name Inspector (Ctrl+Shift+I)',
          position: 'bottom'
        }}
      >
        <Inspect size={14} strokeWidth={2.2} />
      </button>

      <button
        class="action-btn {$debugLogs ? 'active' : ''}"
        onclick={toggleLogs}
        use:islandTooltip={{
          text: $debugLogs ? 'Disable Verbose Logs' : 'Enable Verbose Logs',
          position: 'bottom'
        }}
      >
        <ScrollText size={14} strokeWidth={2.2} />
      </button>

      <button
        class="action-btn"
        onclick={toggleTheme}
        use:islandTooltip={{
          text: 'Toggle Theme',
          position: 'bottom',
          align: 'right'
        }}
      >
        {#if $themeStore === 'dark-gold'}
          <Sun size={14} strokeWidth={2.2} />
        {:else}
          <Moon size={14} strokeWidth={2.2} />
        {/if}
      </button>

      <button
        class="action-btn"
        onclick={() => Events.Emit('devtest:open-devtools')}
        use:islandTooltip={{
          text: 'Open Main Window DevTools (Performance Profiler)',
          position: 'bottom',
          align: 'right'
        }}
      >
        <Terminal size={14} strokeWidth={2.2} />
      </button>

      <button
        class="action-btn"
        onclick={handleReload}
        use:islandTooltip={{
          text: 'Refresh CSS & Reload Dev Tools',
          position: 'bottom',
          align: 'right'
        }}
      >
        <RefreshCw size={14} strokeWidth={2.2} />
      </button>
    </div>
  </header>

  <div class="window-nav-tabs">
    <button
      class="nav-tab-btn"
      class:active={activeTab === 'controls'}
      onclick={() => (activeTab = 'controls')}
    >
      Dev Controls
    </button>
    {#if isShowcaseParam}
      <button
        class="nav-tab-btn"
        class:active={activeTab === 'showcase'}
        onclick={() => (activeTab = 'showcase')}
      >
        Showcase Studio
      </button>
    {/if}
  </div>

  <div class="window-body">
    {#if isShowcaseParam && activeTab === 'showcase'}
      <section class="control-section showcase-section">
        <ShowcaseDevControls />
      </section>
    {:else}
      <section class="control-section">
        <span class="section-label">Active Island View</span>
        <div class="select-container">
          <CustomSelect
            options={pageOptions}
            bind:value={targetMode}
            onChange={handlePageSelect}
          />
        </div>
      </section>

      {#if $debugClassInspector}
        <section class="control-section inspector-section">
          <div class="inspector-header-row">
            <span class="section-label inspector-label">
              Class Name Inspector
            </span>
            {#if inspectedElement}
              <button
                class="copy-selector-btn"
                onclick={copySelector}
                use:islandTooltip={'Copy CSS selector'}
              >
                {#if copiedSelector}
                  <Check size={12} />
                  <span>Copied</span>
                {:else}
                  <Copy size={12} />
                  <span>Copy Selector</span>
                {/if}
              </button>
            {/if}
          </div>

          {#if inspectedElement}
            <div class="inspector-readout-card">
              <div class="inspector-meta-row">
                <span class="meta-badge tag"
                  >&lt;{inspectedElement.tag}&gt;</span
                >
                {#if inspectedElement.dataVinta}
                  <span class="meta-badge vinta"
                    >data-vinta="{inspectedElement.dataVinta}"</span
                  >
                {/if}
                <span class="meta-dims"
                  >{inspectedElement.width} × {inspectedElement.height}px</span
                >
              </div>
              <div class="inspector-classes-row">
                {#if inspectedElement.classes.length > 0}
                  {#each inspectedElement.classes as cls (cls)}
                    <span class="class-chip">.{cls}</span>
                  {/each}
                {:else}
                  <span class="no-classes-text">No semantic classes</span>
                {/if}
              </div>
              <code class="inspector-selector-code"
                >{inspectedElement.selector}</code
              >
            </div>
          {:else}
            <div class="inspector-empty">
              Hover over any element in the main Vinta island to inspect its
              classes and theme selector.
            </div>
          {/if}
        </section>
      {/if}

      <section class="control-section simulation-section">
        <span class="section-label">Simulated Hardware & Feeds</span>
        <div class="simulation-content">
          {#if targetMode === 'notifications'}
            <NotificationDevControls />
          {:else if targetMode === 'home'}
            <HomeDevControls />
          {:else if targetMode === 'weather'}
            <WeatherDevControls />
          {:else if targetMode === 'calendar'}
            <div class="mock-card">
              <span class="mock-title">Calendar & Event Feeds</span>
              <div class="mock-btn-row">
                <button
                  class="mock-btn primary"
                  onclick={() =>
                    addCalendarEvent({
                      title: 'Live Product Demo',
                      description: 'Simulated dev calendar event',
                      date: new Date().toISOString().split('T')[0],
                      startTime: '16:00',
                      endTime: '17:00',
                      location: 'Virtual Room',
                      color: 'indigo'
                    })}
                >
                  Add Today's Event
                </button>
                <button class="mock-btn" onclick={syncAllCalendarFeeds}>
                  Sync Remote Feeds
                </button>
              </div>
            </div>
          {:else}
            <div class="empty-state">
              <span class="empty-title">
                {pageOptions.find((p) => p.value === targetMode)?.label ||
                  'Active Page'}
              </span>
              <p class="empty-description">
                No specialized mock event controls are configured for this view.
              </p>
            </div>
          {/if}
        </div>
      </section>
    {/if}
  </div>
</div>

<CustomTooltip />

<style>
  :global(body.devtest-window-body) {
    margin: 0;
    padding: 0;
    background: #121214;
    color: #f4f4f5;
    font-family: 'Inter', system-ui, sans-serif;
    user-select: none;
    -webkit-user-select: none;
    overflow: hidden;
  }

  .devtest-window {
    display: flex;
    flex-direction: column;
    height: 100vh;
    box-sizing: border-box;
    background: #18181b;
    pointer-events: auto;
  }

  .window-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    background: #202024;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .header-title {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .badge {
    background: #3b82f6;
    color: #ffffff;
    font-size: 0.65rem;
    font-weight: 800;
    padding: 2px 6px;
    border-radius: 4px;
    letter-spacing: 0.5px;
  }

  .runtime-badge {
    background: #3f3f46;
    color: #e4e4e7;
    font-weight: 700;
  }

  .runtime-badge.desktop {
    background: #10b981;
    color: #052e16;
  }

  .header-title h2 {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 600;
    color: #ffffff;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .action-btn {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #a1a1aa;
    width: 28px;
    height: 28px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition:
      background 0.15s ease,
      color 0.15s ease,
      border-color 0.15s ease;
  }

  .action-btn:hover {
    background: rgba(255, 255, 255, 0.12);
    color: #ffffff;
    border-color: rgba(255, 255, 255, 0.18);
  }

  .action-btn.active {
    background: #3b82f6;
    color: #ffffff;
    border-color: #60a5fa;
  }

  .window-nav-tabs {
    display: flex;
    background: #18181b;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    padding: 6px 16px;
    gap: 8px;
  }

  .nav-tab-btn {
    background: transparent;
    border: none;
    color: #a1a1aa;
    font-size: 12px;
    font-weight: 600;
    padding: 6px 12px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .nav-tab-btn:hover {
    color: #ffffff;
    background: rgba(255, 255, 255, 0.06);
  }

  .nav-tab-btn.active {
    color: #e2a03f;
    background: rgba(226, 160, 63, 0.12);
  }

  .window-body {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .control-section {
    background: #202024;
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 8px;
    padding: 14px;
  }

  .section-label {
    display: block;
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    color: #71717a;
    letter-spacing: 0.5px;
    margin-bottom: 10px;
  }

  .select-container {
    width: 100%;
  }

  .simulation-content {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .empty-state {
    padding: 24px 16px;
    text-align: center;
    background: rgba(0, 0, 0, 0.15);
    border-radius: 6px;
  }

  .empty-title {
    display: block;
    font-size: 0.85rem;
    font-weight: 600;
    color: #d4d4d8;
    margin-bottom: 4px;
  }

  .empty-description {
    margin: 0;
    font-size: 0.75rem;
    color: #71717a;
  }

  .mock-card {
    display: flex;
    flex-direction: column;
    gap: 10px;
    background: rgba(0, 0, 0, 0.2);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    padding: 12px;
  }

  .mock-title {
    font-size: 0.8rem;
    font-weight: 700;
    color: #e4e4e7;
    letter-spacing: -0.01em;
  }

  .mock-btn-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .mock-btn {
    padding: 6px 12px;
    font-size: 0.75rem;
    font-weight: 600;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.06);
    color: #e4e4e7;
    border: 1px solid rgba(255, 255, 255, 0.1);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .mock-btn:hover {
    background: rgba(255, 255, 255, 0.12);
    color: #ffffff;
    border-color: rgba(255, 255, 255, 0.2);
  }

  .mock-btn.primary {
    background: #3b82f6;
    border-color: #60a5fa;
    color: #ffffff;
  }

  .mock-btn.primary:hover {
    background: #2563eb;
  }

  .inspector-section {
    border-color: rgba(56, 189, 248, 0.25);
  }

  .inspector-header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 10px;
  }

  .inspector-label {
    margin-bottom: 0;
    color: #38bdf8;
  }

  .copy-selector-btn {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 4px 8px;
    border-radius: 5px;
    background: rgba(56, 189, 248, 0.14);
    border: 1px solid rgba(56, 189, 248, 0.3);
    color: #38bdf8;
    font-size: 0.7rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .copy-selector-btn:hover {
    background: rgba(56, 189, 248, 0.25);
    color: #ffffff;
  }

  .inspector-readout-card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    background: rgba(0, 0, 0, 0.28);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 6px;
    padding: 10px;
    font-family: 'JetBrains Mono', 'Cascadia Code', Consolas, monospace;
  }

  .inspector-meta-row {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px;
  }

  .meta-badge {
    font-size: 0.7rem;
    font-weight: 700;
    padding: 2px 6px;
    border-radius: 4px;
  }

  .meta-badge.tag {
    background: rgba(56, 189, 248, 0.16);
    color: #38bdf8;
  }

  .meta-badge.vinta {
    background: rgba(192, 132, 252, 0.16);
    color: #c084fc;
  }

  .meta-dims {
    margin-left: auto;
    font-size: 0.68rem;
    color: #a1a1aa;
  }

  .inspector-classes-row {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  .class-chip {
    font-size: 0.7rem;
    font-weight: 600;
    color: #fbbf24;
    background: rgba(251, 191, 36, 0.12);
    border: 1px solid rgba(251, 191, 36, 0.25);
    padding: 1.5px 6px;
    border-radius: 4px;
  }

  .no-classes-text {
    font-size: 0.7rem;
    color: #71717a;
    font-style: italic;
  }

  .inspector-selector-code {
    display: block;
    font-size: 0.68rem;
    color: #e4e4e7;
    background: rgba(255, 255, 255, 0.04);
    padding: 5px 8px;
    border-radius: 4px;
    word-break: break-all;
    user-select: text;
    -webkit-user-select: text;
  }

  .inspector-empty {
    font-size: 0.74rem;
    color: #a1a1aa;
    padding: 10px;
    background: rgba(0, 0, 0, 0.2);
    border-radius: 6px;
    text-align: center;
  }
</style>
