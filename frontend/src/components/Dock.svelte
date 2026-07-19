<script lang="ts">
  import { getContext } from 'svelte'
  import { Grid, Settings, Power, Shrink, EyeOff } from 'lucide-svelte'
  import { AppControls } from '../utils/AppControls.svelte'
  import DockButton from './DockButton.svelte'
  import AppsPage from '../pages/AppsPage.svelte'
  import SettingsPage from '../pages/SettingsPage.svelte'
  import type { IslandConfig } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  const hidePill = getContext<() => void>('hidePill')

  let { activePage }: { activePage: unknown } = $props()
</script>

<div class="dock-container">
  <div class="dock">
    <DockButton
      tooltip="Apps"
      active={activePage === AppsPage}
      onclick={() => AppControls.goApps(navigate)}
    >
      <Grid size={18} />
    </DockButton>

    <div class="divider"></div>

    <DockButton
      tooltip="Settings"
      active={activePage === SettingsPage}
      onclick={(e: MouseEvent) => AppControls.goSettings(navigate, e)}
    >
      <Settings size={18} />
    </DockButton>

    <div class="divider"></div>

    <DockButton
      variant="success"
      tooltip="Hide Pill"
      onclick={(e: MouseEvent) => AppControls.doHide(hidePill, e)}
    >
      <EyeOff size={18} />
    </DockButton>

    <div class="divider"></div>

    <DockButton
      variant="warning"
      tooltip="Minimize"
      onclick={(e: MouseEvent) => AppControls.doMinimize(navigate, e)}
    >
      <Shrink size={18} />
    </DockButton>

    <div class="divider"></div>

    <DockButton
      variant="danger"
      tooltip="Quit Vinta"
      onclick={AppControls.doQuit}
    >
      <Power size={18} />
    </DockButton>
  </div>
</div>

<style>
  .dock-container {
    position: absolute;
    bottom: 8px;
    left: 0;
    width: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
    pointer-events: none;
  }

  .dock {
    pointer-events: auto;
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
</style>
