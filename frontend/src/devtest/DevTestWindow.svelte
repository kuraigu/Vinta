<script lang="ts">
  import { onMount } from 'svelte'
  import { Events } from '@wailsio/runtime'
  import {
    Pin,
    Maximize2,
    Minimize2,
    Crosshair,
    ScrollText,
    Sun,
    Moon,
    RefreshCw
  } from 'lucide-svelte'
  import CustomSelect from '../components/CustomSelect.svelte'
  import NotificationDevControls from './NotificationDevControls.svelte'
  import HomeDevControls from './HomeDevControls.svelte'
  import {
    themeStore,
    devPinnedStore,
    addCalendarEvent,
    syncAllCalendarFeeds
  } from '../store'
  import { debugBoundsOverlay, debugLogs } from './debugConfig'

  let isExpandedMode = $state(false)
  let targetMode = $state('home')

  const pageOptions = [
    { value: 'home', label: 'Home Page' },
    { value: 'calendar', label: 'Calendar' },
    { value: 'media', label: 'Media Player' },
    { value: 'notifications', label: 'Notifications' },
    { value: 'notes', label: 'Scratchpad' },
    { value: 'apps', label: 'App Launcher' },
    { value: 'settings', label: 'Settings' }
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
    emitNavigate(newMode, isExpandedMode)
  }

  function toggleExpandedMode() {
    isExpandedMode = !isExpandedMode
    $devPinnedStore = true
    emitNavigate(targetMode, isExpandedMode)
  }

  function handleReload() {
    window.location.reload()
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

  function toggleTheme() {
    $themeStore = $themeStore === 'dark' ? 'light' : 'dark'
    document.documentElement.setAttribute('data-theme', $themeStore)
    try {
      Events.Emit('devtest-theme', { theme: $themeStore })
    } catch {}
  }

  onMount(() => {
    document.documentElement.setAttribute('data-theme', 'dark')

    const unsubSync = Events.On(
      'devtest-sync-page',
      (ev: { data: unknown }) => {
        const pageId = (ev?.data as { pageId?: string })?.pageId
        if (pageId && pageId !== targetMode) {
          targetMode = pageId
        }
      }
    )

    return () => {
      unsubSync()
    }
  })
</script>

<div class="devtest-window">
  <header class="window-header">
    <div class="header-title">
      <span class="badge">DEV</span>
      <h2>Vinta Developer Tools</h2>
    </div>

    <div class="header-actions">
      <button
        class="action-btn {$devPinnedStore ? 'active' : ''}"
        onclick={togglePin}
        title={$devPinnedStore ? 'Unpin Page' : 'Pin Page'}
      >
        <Pin size={14} strokeWidth={2.2} />
      </button>

      <button
        class="action-btn {isExpandedMode ? 'active' : ''}"
        onclick={toggleExpandedMode}
        title={isExpandedMode
          ? 'Switch to Minimal Island'
          : 'Switch to Expanded Card'}
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
        title={$debugBoundsOverlay
          ? 'Hide Bounds Overlay'
          : 'Show Bounds Overlay'}
      >
        <Crosshair size={14} strokeWidth={2.2} />
      </button>

      <button
        class="action-btn {$debugLogs ? 'active' : ''}"
        onclick={toggleLogs}
        title={$debugLogs ? 'Disable Verbose Logs' : 'Enable Verbose Logs'}
      >
        <ScrollText size={14} strokeWidth={2.2} />
      </button>

      <button class="action-btn" onclick={toggleTheme} title="Toggle Theme">
        {#if $themeStore === 'dark'}
          <Sun size={14} strokeWidth={2.2} />
        {:else}
          <Moon size={14} strokeWidth={2.2} />
        {/if}
      </button>

      <button
        class="action-btn"
        onclick={handleReload}
        title="Reload Dev Tools"
      >
        <RefreshCw size={14} strokeWidth={2.2} />
      </button>
    </div>
  </header>

  <div class="window-body">
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

    <section class="control-section simulation-section">
      <span class="section-label">Simulated Hardware & Feeds</span>
      <div class="simulation-content">
        {#if targetMode === 'notifications'}
          <NotificationDevControls />
        {:else if targetMode === 'home'}
          <HomeDevControls />
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
  </div>
</div>

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
</style>
