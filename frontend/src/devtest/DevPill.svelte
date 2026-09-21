<script lang="ts">
  import { getContext, onMount, untrack } from 'svelte'
  import { fade } from 'svelte/transition'
  import { SetDevPillBounds } from '../../bindings/Vinta/backend/app.js'

  import {
    Pin,
    Maximize2,
    Minimize2,
    Crosshair,
    ScrollText,
    Sun,
    Moon,
    ChevronDown,
    ChevronUp,
    RefreshCw,
    X
  } from 'lucide-svelte'

  import HomePage, { config as homeConfig } from '../pages/HomePage.svelte'
  import MediaPage, { config as mediaConfig } from '../pages/MediaPage.svelte'
  import NotificationPage, {
    config as notifConfig
  } from '../pages/NotificationPage.svelte'
  import NotesPage, { config as notesConfig } from '../pages/NotesPage.svelte'
  import AppsPage, { config as appsConfig } from '../pages/AppsPage.svelte'
  import SettingsPage, {
    config as settingsConfig
  } from '../pages/SettingsPage.svelte'

  import NotificationDevControls from './NotificationDevControls.svelte'
  import HomeDevControls from './HomeDevControls.svelte'
  import CustomSelect from '../components/CustomSelect.svelte'
  import type { Component } from 'svelte'
  import type { IslandConfig } from '../store'
  import {
    themeStore,
    devTestingStore,
    devPinnedStore,
    activePageIdStore
  } from '../store'
  import { debugBoundsOverlay, debugLogs } from './debugConfig'

  let devW = $state(0)
  let devH = $state(0)
  let extraH = $state(0)
  let isExpandedMode = $state(false)
  let isMinimized = $state(false)
  let isDevPillDismissed = $state(false)

  const navigate =
    getContext<
      (page: Component, config: IslandConfig, expand?: boolean) => void
    >('navigate')

  function syncDevPillBounds() {
    const scale = window.devicePixelRatio || 1
    if (isDevPillDismissed) {
      SetDevPillBounds(0, 0, 0, scale)
      return
    }
    if (devW > 0 && devH > 0) {
      const activeExtra = isMinimized ? 0 : extraH
      SetDevPillBounds(devW, devH, activeExtra, scale)
    } else {
      SetDevPillBounds(0, 0, 0, scale)
    }
  }

  function handleDevBoundsSync() {
    syncDevPillBounds()
  }

  $effect(handleDevBoundsSync)

  // Activate DevTest non-persistent mode and listen to dropdown events
  onMount(() => {
    devTestingStore.set(true)
    const handleToggle = (e: Event) => {
      const target = e.target as HTMLElement | null
      // Only react to dropdown overlays originating from within DevPill
      if (target && !target.closest?.('.dev-pill')) return
      const detail = (e as CustomEvent).detail
      extraH = detail?.open ? 250 : 0
      syncDevPillBounds()
    }
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'F12' || (e.ctrlKey && e.shiftKey && e.key === 'D')) {
        isDevPillDismissed = !isDevPillDismissed
        const scale = window.devicePixelRatio || 1
        if (isDevPillDismissed) {
          devTestingStore.set(false)
          devPinnedStore.set(false)
          SetDevPillBounds(0, 0, 0, scale)
        } else {
          devTestingStore.set(true)
          syncDevPillBounds()
        }
      }
    }
    window.addEventListener('vinta-overlay', handleToggle)
    window.addEventListener('resize', syncDevPillBounds)
    window.addEventListener('keydown', handleKeyDown)

    return () => {
      devTestingStore.set(false)
      devPinnedStore.set(false)
      const scale = window.devicePixelRatio || 1
      SetDevPillBounds(0, 0, 0, scale)
      window.removeEventListener('vinta-overlay', handleToggle)
      window.removeEventListener('resize', syncDevPillBounds)
      window.removeEventListener('keydown', handleKeyDown)
    }
  })

  // Registry of all application pages
  const pages: Record<
    string,
    {
      label: string
      page: Component
      config: IslandConfig
      controls: Component | null
    }
  > = {
    home: {
      label: 'Home',
      page: HomePage,
      config: homeConfig,
      controls: HomeDevControls
    },
    media: {
      label: 'Media Player',
      page: MediaPage,
      config: mediaConfig,
      controls: null
    },
    notification: {
      label: 'Notification',
      page: NotificationPage,
      config: notifConfig,
      controls: NotificationDevControls
    },
    notes: {
      label: 'Notes',
      page: NotesPage,
      config: notesConfig,
      controls: null
    },
    apps: {
      label: 'Apps & Tools',
      page: AppsPage,
      config: appsConfig,
      controls: null
    },
    settings: {
      label: 'Settings',
      page: SettingsPage,
      config: settingsConfig,
      controls: null
    }
  }

  let pageOptions = $derived(
    Object.entries(pages).map(([key, data]) => ({
      value: key,
      label: data.label
    }))
  )

  let targetMode = $state($activePageIdStore || 'home')

  // Bidirectional sync: when page changes on the Island, update targetMode without reacting to targetMode itself
  function handleExternalPageSync() {
    const activeId = $activePageIdStore
    if (activeId && pages[activeId]) {
      untrack(() => {
        if (targetMode !== activeId) {
          targetMode = activeId
        }
      })
    }
  }

  $effect(handleExternalPageSync)

  // Direct forced navigation when user picks a page from the DevPill dropdown
  function handlePageSelect(newMode: string) {
    targetMode = newMode
    const pageData = pages[newMode]
    if (pageData) {
      $devPinnedStore = true
      navigate(pageData.page, pageData.config, isExpandedMode)
    }
  }

  function toggleExpandedMode() {
    isExpandedMode = !isExpandedMode
    const pageData = pages[targetMode]
    if (pageData) {
      $devPinnedStore = true
      navigate(pageData.page, pageData.config, isExpandedMode)
    }
  }

  /**
   * Forces a full browser reload of the webview to reload Vite modules and styles.
   */
  function handleReload() {
    window.location.reload()
  }
</script>

{#if !isDevPillDismissed}
  <div
    class="dev-pill"
    in:fade={{ duration: 300 }}
    bind:clientWidth={devW}
    bind:clientHeight={devH}
  >
    <div class="top-toolbar">
      <div class="toolbar-left">
        <span class="label">DEV TEST</span>
        <CustomSelect
          options={pageOptions}
          bind:value={targetMode}
          onChange={handlePageSelect}
        />
      </div>

      <div class="quick-toggles">
        <button
          class="icon-btn {$devPinnedStore ? 'active' : ''}"
          onclick={() => ($devPinnedStore = !$devPinnedStore)}
          title={$devPinnedStore
            ? 'Unpin Page (Auto-hide Active)'
            : 'Pin Page (Freeze Auto-hide & Routing)'}
        >
          <Pin size={13} strokeWidth={2.2} />
        </button>
        <button
          class="icon-btn {isExpandedMode ? 'active' : ''}"
          onclick={toggleExpandedMode}
          title={isExpandedMode
            ? 'Switch to Minimal Island'
            : 'Switch to Expanded Card'}
        >
          {#if isExpandedMode}
            <Minimize2 size={13} strokeWidth={2.2} />
          {:else}
            <Maximize2 size={13} strokeWidth={2.2} />
          {/if}
        </button>
        <button
          class="icon-btn {$debugBoundsOverlay ? 'active' : ''}"
          onclick={() => ($debugBoundsOverlay = !$debugBoundsOverlay)}
          title={$debugBoundsOverlay
            ? 'Hide Bounds Overlay'
            : 'Show Bounds Overlay'}
        >
          <Crosshair size={13} strokeWidth={2.2} />
        </button>
        <button
          class="icon-btn {$debugLogs ? 'active' : ''}"
          onclick={() => ($debugLogs = !$debugLogs)}
          title={$debugLogs ? 'Disable Verbose Logs' : 'Enable Verbose Logs'}
        >
          <ScrollText size={13} strokeWidth={2.2} />
        </button>
        <button
          class="icon-btn"
          onclick={() =>
            ($themeStore = $themeStore === 'dark' ? 'light' : 'dark')}
          title="Toggle Theme"
        >
          {#if $themeStore === 'dark'}
            <Sun size={13} strokeWidth={2.2} />
          {:else}
            <Moon size={13} strokeWidth={2.2} />
          {/if}
        </button>
        <button
          class="icon-btn"
          onclick={handleReload}
          title="Reload Webview (Refresh Page)"
        >
          <RefreshCw size={13} strokeWidth={2.2} />
        </button>
        <button
          class="icon-btn minimize-btn"
          onclick={() => (isMinimized = !isMinimized)}
          title={isMinimized
            ? 'Expand DevTest Panel'
            : 'Minimize DevTest Panel'}
        >
          {#if isMinimized}
            <ChevronDown size={13} strokeWidth={2.2} />
          {:else}
            <ChevronUp size={13} strokeWidth={2.2} />
          {/if}
        </button>
        <button
          class="icon-btn"
          onclick={() => {
            isDevPillDismissed = true
            devTestingStore.set(false)
            devPinnedStore.set(false)
            SetDevPillBounds(0, 0, 0, window.devicePixelRatio || 1)
          }}
          title="Dismiss DevTest Panel (Press F12 to restore)"
        >
          <X size={13} strokeWidth={2.2} />
        </button>
      </div>
    </div>

    {#if !isMinimized}
      <div class="divider"></div>
      <div class="actions-scroll">
        {#if pages[targetMode]?.controls}
          {@const Controls = pages[targetMode].controls}
          <Controls />
        {:else}
          <div class="empty-controls">
            <span class="empty-title"
              >{pages[targetMode]?.label || 'Active Page'}</span
            >
            <span class="empty-desc"
              >No specialized simulation controls for this view.</span
            >
          </div>
        {/if}
      </div>
    {/if}
  </div>
{/if}

<style>
  .dev-pill {
    background: var(--c-bg);
    color: var(--c-text-primary);
    border-radius: 14px;
    border: 1px solid var(--c-border);
    display: flex;
    flex-direction: column;
    width: 430px;
    max-height: 280px;
    padding: 10px 12px;
    box-sizing: border-box;
    z-index: 9999;
    pointer-events: auto;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
    gap: 8px;
  }

  .top-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    flex-shrink: 0;
  }

  .toolbar-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .divider {
    width: 100%;
    height: 1px;
    background: var(--c-border);
    flex-shrink: 0;
  }

  .actions-scroll {
    width: 100%;
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    padding-right: 2px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .actions-scroll::-webkit-scrollbar {
    width: 4px;
  }

  .actions-scroll::-webkit-scrollbar-thumb {
    background: var(--c-border);
    border-radius: 4px;
  }

  .label {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--c-text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .quick-toggles {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-left: 2px;
  }

  .icon-btn {
    background: transparent;
    border: 1px solid transparent;
    color: var(--c-text-secondary);
    padding: 5px 6px;
    border-radius: 6px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s ease;
  }

  .icon-btn:hover {
    background: var(--c-hover);
    color: var(--c-text-primary);
  }

  .icon-btn.active {
    background: rgba(0, 122, 255, 0.15);
    color: var(--c-primary, #007aff);
    border-color: rgba(0, 122, 255, 0.35);
  }

  .empty-controls {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 24px 12px;
    text-align: center;
    gap: 4px;
  }

  .empty-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--c-text-primary);
  }

  .empty-desc {
    font-size: 11px;
    color: var(--c-text-secondary);
  }
</style>
