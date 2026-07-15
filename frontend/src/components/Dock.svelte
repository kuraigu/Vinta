<script lang="ts">
  import { getContext } from 'svelte'
  import { Grid, Settings, Power, Shrink, EyeOff } from 'lucide-svelte'
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
      data-tooltip="Apps"
    >
      <Grid size={18} />
    </button>
    <div class="divider"></div>
    <button
      class="dock-btn"
      class:active={activePage === SettingsPage}
      on:click={goSettings}
      data-tooltip="Settings"
    >
      <Settings size={18} />
    </button>
    <div class="divider"></div>
    <button class="dock-btn success" on:click={doHide} data-tooltip="Hide Pill">
      <EyeOff size={18} />
    </button>
    <div class="divider"></div>
    <button
      class="dock-btn warning"
      on:click={doMinimize}
      data-tooltip="Minimize"
    >
      <Shrink size={18} />
    </button>
    <div class="divider"></div>
    <button class="dock-btn danger" on:click={doQuit} data-tooltip="Quit Vinta">
      <Power size={18} />
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
    position: relative; /* For custom tooltip */
  }

  /* Custom Tooltip */
  .dock-btn::after {
    content: attr(data-tooltip);
    position: absolute;
    bottom: calc(100% + 12px);
    left: 50%;
    transform: translateX(-50%) scale(0.9);
    background: rgba(20, 20, 20, 0.95);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    color: #fff;
    padding: 6px 10px;
    border-radius: 8px;
    font-size: 0.75rem;
    font-weight: 500;
    white-space: nowrap;
    opacity: 0;
    pointer-events: none;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
    border: 1px solid rgba(255, 255, 255, 0.1);
    z-index: 100;
  }

  :global([data-theme='light']) .dock-btn::after {
    background: rgba(255, 255, 255, 0.95);
    color: #000;
    border: 1px solid rgba(0, 0, 0, 0.1);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .dock-btn:hover::after {
    opacity: 1;
    transform: translateX(-50%) scale(1);
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

  /* Traffic light hover colors for actions */
  .dock-btn.danger:hover {
    color: #ff3b30;
    background: rgba(255, 59, 48, 0.1);
  }

  .dock-btn.warning:hover {
    color: #ffcc00;
    background: rgba(255, 204, 0, 0.1);
  }

  .dock-btn.success:hover {
    color: #34c759;
    background: rgba(52, 199, 89, 0.1);
  }
</style>
