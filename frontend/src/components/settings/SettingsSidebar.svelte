<script lang="ts">
  import {
    Palette,
    Layers,
    Sliders,
    Bell,
    Clock,
    CloudSun,
    Cpu,
    RefreshCw,
    Info
  } from '@lucide/svelte'
  import { activeSettingsTabStore } from '../../store'

  const categories = [
    { id: 'appearance', name: 'Appearance', icon: Palette },
    { id: 'modules', name: 'Apps & Modules', icon: Layers },
    { id: 'behavior', name: 'Behavior', icon: Sliders },
    { id: 'notifications', name: 'Notifications', icon: Bell },
    { id: 'clock', name: 'Clock & Date', icon: Clock },
    { id: 'weather', name: 'Weather', icon: CloudSun },
    { id: 'system', name: 'System', icon: Cpu },
    { id: 'about', name: 'About', icon: Info },
    { id: 'updates', name: 'Updates', icon: RefreshCw }
  ]
</script>

<div class="sidebar-nav">
  {#each categories as cat (cat.id)}
    <button
      class="nav-item"
      class:active={$activeSettingsTabStore === cat.id}
      onclick={() => ($activeSettingsTabStore = cat.id)}
    >
      <cat.icon size={14} class="nav-icon" />
      <span>{cat.name}</span>
    </button>
  {/each}
</div>

<style>
  .sidebar-nav {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding-right: 4px;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 8px 10px;
    border-radius: 8px;
    background: transparent;
    border: none;
    color: var(--c-text-secondary);
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
    text-align: left;
  }

  .nav-item:hover {
    background: rgba(255, 255, 255, 0.06);
    color: var(--c-text-primary);
  }

  .nav-item.active {
    background: rgba(255, 255, 255, 0.1);
    color: var(--c-text-primary);
    font-weight: 600;
  }

  .nav-item :global(.nav-icon) {
    color: var(--c-text-secondary);
  }

  .nav-item.active :global(.nav-icon) {
    color: var(--c-primary, #6366f1);
  }
</style>
