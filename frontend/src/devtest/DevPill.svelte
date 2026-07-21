<script lang="ts">
  import { getContext, onMount, untrack } from 'svelte'
  import { fade } from 'svelte/transition'
  import { SetDevPillBounds } from '../../wailsjs/go/backend/App.js'

  import NotificationPage, {
    config as notifConfig
  } from '../pages/NotificationPage.svelte'
  import ClockPage, { config as clockConfig } from '../pages/ClockPage.svelte'
  import NotificationDevControls from './NotificationDevControls.svelte'
  import CustomSelect from './CustomSelect.svelte'
  import type { Component } from 'svelte'
  import type { IslandConfig } from '../store'
  import { themeStore, shapeStore } from '../store'

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
      SetDevPillBounds(devW, devH + extraH, scale)
    }
  })

  // Listen to CustomSelect toggle events to expand the native OS click bounds
  onMount(() => {
    const handleToggle = (e: Event) => {
      const detail = (e as CustomEvent).detail
      extraH = detail ? 250 : 0
    }
    window.addEventListener('dev-dropdown-toggle', handleToggle)
    return () => window.removeEventListener('dev-dropdown-toggle', handleToggle)
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
    clock: {
      label: 'Clock (Idle)',
      page: ClockPage,
      config: clockConfig,
      controls: null
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

  let targetMode = $state('clock')

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
  <div class="left-section">
    <span class="label">Dev</span>
    <CustomSelect options={pageOptions} bind:value={targetMode} />

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
    </div>
  </div>

  <div class="actions">
    {#if pages[targetMode].controls}
      {@const Controls = pages[targetMode].controls}
      <Controls />
    {/if}
  </div>
</div>

<style>
  .dev-pill {
    /* No absolute positioning! We are in a flex column now */
    background: var(--c-bg);
    color: var(--c-text-primary);
    border-radius: 12px;
    border: 1px solid var(--c-border);
    display: flex;
    align-items: center; /* Back to centered! */
    justify-content: space-between;
    padding: 8px 15px;
    box-sizing: border-box;
    z-index: 9999;
    pointer-events: auto; /* Ensures it receives clicks */
    width: max-content;
    height: max-content;
  }

  .left-section {
    display: flex;
    align-items: center;
    gap: 10px;
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
    margin-left: 4px;
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
    /* Removed transition to avoid WebView2 compositing overlap bugs */
  }

  .icon-btn:hover {
    background: var(--c-hover);
  }

  .actions {
    display: flex;
    gap: 8px;
  }
</style>
