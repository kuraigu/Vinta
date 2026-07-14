<script lang="ts">
  import { getContext } from 'svelte'
  import { Grid, Settings, Power, Minus, EyeOff } from 'lucide-svelte'
  import { Quit } from '../../wailsjs/go/backend/App.js'
  import ClockPage, { config as clockConfig } from '../pages/ClockPage.svelte'
  import AppsPage, { config as appsConfig } from '../pages/AppsPage.svelte'
  import SettingsPage, {
    config as settingsConfig
  } from '../pages/SettingsPage.svelte'
  import type { IslandConfig } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  const hidePill = getContext<() => void>('hidePill')

  export let activePage: unknown

  function goApps() {
    navigate(AppsPage, appsConfig, true)
  }

  function goSettings(e: Event) {
    e.stopPropagation()
    navigate(SettingsPage, settingsConfig, true)
  }

  function doMinimize(e: Event) {
    e.stopPropagation()
    navigate(ClockPage, clockConfig, false)
  }

  function doHide(e: Event) {
    e.stopPropagation()
    hidePill()
  }

  function doQuit() {
    Quit()
  }
</script>

<div class="dock-container">
  <div class="dock">
    <button
      class="dock-btn"
      class:active={activePage === AppsPage}
      on:click={goApps}
      title="Apps"
    >
      <Grid size={18} />
    </button>
    <div class="divider"></div>
    <button
      class="dock-btn"
      class:active={activePage === SettingsPage}
      on:click={goSettings}
      title="Settings"
    >
      <Settings size={18} />
    </button>
    <div class="divider"></div>
    <button class="dock-btn danger" on:click={doQuit} title="Quit Vinta">
      <Power size={18} />
    </button>
    <div class="divider"></div>
    <button class="dock-btn" on:click={doMinimize} title="Minimize">
      <Minus size={18} />
    </button>
    <div class="divider"></div>
    <button class="dock-btn" on:click={doHide} title="Hide Pill">
      <EyeOff size={18} />
    </button>
  </div>
</div>

<style>
  .dock-container {
    position: absolute;
    bottom: 8px; /* Moved lower */
    left: 0;
    width: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
    pointer-events: none; /* Let clicks pass through container */
  }

  .dock {
    pointer-events: auto; /* Re-enable clicks for the dock itself */
    display: flex;
    align-items: center;
    padding: 6px 12px;
    gap: 12px;
  }

  .divider {
    width: 1px;
    height: 16px;
    background: var(--c-border);
    opacity: 0.3;
  }

  .dock-btn {
    background: transparent;
    border: none;
    color: var(--c-text-secondary);
    width: 32px;
    height: 32px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
  }

  .dock-btn:hover {
    color: var(--c-text-primary);
    background: rgba(255, 255, 255, 0.1);
    transform: scale(1.1);
  }

  :global([data-theme='light']) .dock-btn:hover {
    background: rgba(0, 0, 0, 0.05);
  }

  .dock-btn.active {
    color: var(--c-text-primary);
    background: var(--c-item-bg);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .dock-btn.danger:hover {
    color: var(--c-danger);
    background: rgba(215, 0, 21, 0.1);
  }
</style>
