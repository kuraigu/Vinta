<script lang="ts">
  import { getContext, onMount, untrack } from 'svelte'
  import { fade } from 'svelte/transition'
  import { SetDevPillBounds } from '../../wailsjs/go/backend/App.js'

  import NotificationPage, {
    config as notifConfig
  } from '../pages/NotificationPage.svelte'
  import HomePage, { config as homeConfig } from '../pages/HomePage.svelte'
  import NotificationDevControls from './NotificationDevControls.svelte'
  import WeatherDevControls from './WeatherDevControls.svelte'
  import CustomSelect from '../components/CustomSelect.svelte'
  import type { Component } from 'svelte'
  import type { IslandConfig } from '../store'
  import { themeStore, shapeStore, devTestingStore } from '../store'

  let devW = $state(0)
  let devH = $state(0)
  let extraH = $state(0)

  const navigate =
    getContext<
      (page: Component, config: IslandConfig, expand?: boolean) => void
    >('navigate')

  $effect(() => {
    if (devW > 0 && devH > 0) {
      const scale = window.devicePixelRatio || 1
      SetDevPillBounds(devW, devH, extraH, scale)
    }
  })

  // Activate DevTest non-persistent mode and listen to dropdown events
  onMount(() => {
    devTestingStore.set(true)
    const handleToggle = (e: Event) => {
      const detail = (e as CustomEvent).detail
      extraH = detail ? 250 : 0
    }
    window.addEventListener('dev-dropdown-toggle', handleToggle)
    return () => {
      devTestingStore.set(false)
      const scale = window.devicePixelRatio || 1
      SetDevPillBounds(0, 0, 0, scale)
      window.removeEventListener('dev-dropdown-toggle', handleToggle)
    }
  })

  // Registry pattern: Each page provides its own config and optional dev controls
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
      label: 'Home (Idle)',
      page: HomePage,
      config: homeConfig,
      controls: WeatherDevControls
    },
    notification: {
      label: 'Notification',
      page: NotificationPage,
      config: notifConfig,
      controls: NotificationDevControls
    }
  }

  let pageOptions = $derived(
    Object.entries(pages).map(([key, data]) => ({
      value: key,
      label: data.label
    }))
  )

  let targetMode = $state('home')
  let isMinimized = $state(false)

  $effect(() => {
    const pageData = pages[targetMode]
    if (pageData) {
      untrack(() => {
        navigate(pageData.page, pageData.config, false)
      })
    }
  })
</script>

<div
  class="dev-pill"
  in:fade={{ duration: 300 }}
  bind:clientWidth={devW}
  bind:clientHeight={devH}
>
  <div class="top-toolbar">
    <div class="toolbar-left">
      <span class="label">DEV TEST</span>
      <CustomSelect options={pageOptions} bind:value={targetMode} />
    </div>

    <div class="quick-toggles">
      <button
        class="icon-btn"
        onclick={() =>
          ($themeStore = $themeStore === 'dark' ? 'light' : 'dark')}
        title="Toggle Theme"
      >
        {$themeStore === 'dark' ? '☀️' : '🌙'}
      </button>
      <button
        class="icon-btn"
        onclick={() => ($shapeStore = $shapeStore === 'pill' ? 'box' : 'pill')}
        title="Toggle Shape"
      >
        {$shapeStore === 'pill' ? '🔲' : '💊'}
      </button>
      <button
        class="icon-btn minimize-btn"
        onclick={() => (isMinimized = !isMinimized)}
        title={isMinimized ? 'Expand DevTest Panel' : 'Minimize DevTest Panel'}
      >
        {isMinimized ? '🔽' : '🔼'}
      </button>
    </div>
  </div>

  {#if !isMinimized && pages[targetMode]?.controls}
    {@const Controls = pages[targetMode].controls}
    <div class="divider"></div>
    <div class="actions-scroll">
      <Controls />
    </div>
  {/if}
</div>

<style>
  .dev-pill {
    /* Vertical panel layout instead of wide horizontal row */
    background: var(--c-bg);
    color: var(--c-text-primary);
    border-radius: 14px;
    border: 1px solid var(--c-border);
    display: flex;
    flex-direction: column;
    width: 380px;
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
    gap: 4px;
    margin-left: 2px;
  }

  .icon-btn {
    background: transparent;
    border: 1px solid transparent;
    color: var(--c-text-primary);
    padding: 4px 6px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 1rem;
    position: relative;
    z-index: 1;
  }

  .icon-btn:hover {
    background: var(--c-hover);
  }
</style>
