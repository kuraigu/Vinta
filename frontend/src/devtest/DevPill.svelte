<script lang="ts">
  import { onMount, getContext } from 'svelte'
  import { fade } from 'svelte/transition'
  import { SetDevPillBounds } from '../../wailsjs/go/backend/App.js'

  import NotificationPage, {
    config as notifConfig
  } from '../pages/NotificationPage.svelte'
  import ClockPage, { config as clockConfig } from '../pages/ClockPage.svelte'
  import NotificationDevControls from './NotificationDevControls.svelte'
  import type { ComponentType } from 'svelte'
  import type { IslandConfig } from '../store'
  import { themeStore, shapeStore } from '../store'

  export let activePage: ComponentType

  let devW = 400
  let devH = 60

  const navigate =
    getContext<
      (page: ComponentType, config: IslandConfig, expand?: boolean) => void
    >('navigate')

  onMount(() => {
    const scale = window.devicePixelRatio || 1
    SetDevPillBounds(devW, devH, scale)
  })

  // Registry pattern: Each page provides its own config and optional dev controls
  const pages: Record<
    string,
    {
      label: string
      page: ComponentType
      config: IslandConfig
      controls: ComponentType | null
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

  let targetMode = 'clock'

  function handlePageChange(e: Event) {
    const target = e.target as HTMLSelectElement
    targetMode = target.value
    const pageData = pages[targetMode]

    // Always navigate to the selected page when changed from the dropdown
    navigate(pageData.page, pageData.config, false)
  }
</script>

<div
  class="dev-pill"
  in:fade={{ duration: 300 }}
  style="width: {devW}px; height: {devH}px;"
>
  <div class="left-section">
    <span class="label">Dev</span>
    <select class="page-select" on:change={handlePageChange} value={targetMode}>
      {#each Object.entries(pages) as [key, data] (key)}
        <option value={key}>{data.label}</option>
      {/each}
    </select>

    <div class="quick-toggles">
      <button
        class="icon-btn"
        on:click={() =>
          ($themeStore = $themeStore === 'dark' ? 'light' : 'dark')}
        title="Toggle Theme"
      >
        {$themeStore === 'dark' ? '☀️' : '🌙'}
      </button>
      <button
        class="icon-btn"
        on:click={() => ($shapeStore = $shapeStore === 'pill' ? 'box' : 'pill')}
        title="Toggle Shape"
      >
        {$shapeStore === 'pill' ? '🔲' : '💊'}
      </button>
    </div>
  </div>

  <div class="actions">
    {#if pages[targetMode].controls}
      <svelte:component this={pages[targetMode].controls} />
    {/if}
  </div>
</div>

<style>
  .dev-pill {
    /* No absolute positioning! We are in a flex column now */
    background: var(--c-bg);
    color: var(--c-text-primary);
    padding: 8px 12px;
    border-radius: 12px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
    border: 1px solid var(--c-border);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 15px;
    box-sizing: border-box;
    z-index: 9999;
    pointer-events: auto; /* Ensures it receives clicks */
  }

  .left-section {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .label {
    font-size: 0.8rem;
    font-weight: bold;
    color: #ffaa00;
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .page-select {
    background: transparent;
    color: var(--c-text-primary);
    border: 1px solid var(--c-border);
    padding: 4px 8px;
    border-radius: 6px;
    font-size: 0.85rem;
    outline: none;
    cursor: pointer;
  }

  .page-select:focus {
    border-color: var(--c-primary);
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
    transition: background 0.2s;
  }

  .icon-btn:hover {
    background: var(--c-hover);
  }

  .actions {
    display: flex;
    gap: 8px;
  }
</style>
