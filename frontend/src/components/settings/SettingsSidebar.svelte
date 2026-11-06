<script lang="ts">
  import {
    Palette,
    Layers,
    Sliders,
    Keyboard,
    Bell,
    Clock,
    CloudSun,
    Languages,
    Cpu,
    Download,
    Info
  } from 'lucide-svelte'
  import { activeSettingsTabStore, tStore } from '../../store'

  const categories = [
    { id: 'appearance', name: 'Appearance', icon: Palette },
    { id: 'modules', name: 'Apps & Features', icon: Layers },
    { id: 'behavior', name: 'Behavior', icon: Sliders },
    { id: 'shortcuts', name: 'Shortcuts', icon: Keyboard },
    { id: 'notifications', name: 'Notifications', icon: Bell },
    { id: 'clock', name: 'Clock & Date', icon: Clock },
    { id: 'weather', name: 'Weather', icon: CloudSun },
    { id: 'language', name: 'Language', icon: Languages },
    { id: 'system', name: 'System', icon: Cpu },
    { id: 'about', name: 'About', icon: Info },
    { id: 'updates', name: 'Updates', icon: Download }
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
      <span>{$tStore('settings.tabs.' + cat.id)}</span>
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
    color: var(--vinta-text-secondary);
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
    text-align: left;
  }

  .nav-item:hover {
    background: rgba(255, 255, 255, 0.06);
    color: var(--vinta-text-primary);
  }

  .nav-item.active {
    background: rgba(255, 255, 255, 0.1);
    color: var(--vinta-text-primary);
    font-weight: 600;
  }

  .nav-item :global(.nav-icon) {
    color: var(--vinta-text-secondary);
  }

  .nav-item.active :global(.nav-icon) {
    color: var(--vinta-accent-primary, var(--vinta-primary, #38bdf8));
  }
</style>
