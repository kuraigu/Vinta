<!--
  @component SettingsSidebar
  @description Left-hand navigation list of settings categories. Uses shared IconBadge
  primitives and SidebarLayout navigation classes for zero CSS duplication.
-->
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
    Cloud,
    Download,
    Info
  } from 'lucide-svelte'
  import IconBadge, { type IconBadgeColor } from '../IconBadge.svelte'
  import { activeSettingsTabStore, tStore } from '../../store'

  const categories: {
    id: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon: any
    badge: IconBadgeColor
  }[] = [
    { id: 'appearance', icon: Palette, badge: 'rose' },
    { id: 'modules', icon: Layers, badge: 'emerald' },
    { id: 'behavior', icon: Sliders, badge: 'cyan' },
    { id: 'shortcuts', icon: Keyboard, badge: 'purple' },
    { id: 'notifications', icon: Bell, badge: 'blue' },
    { id: 'clock', icon: Clock, badge: 'amber' },
    { id: 'weather', icon: CloudSun, badge: 'cyan' },
    { id: 'language', icon: Languages, badge: 'purple' },
    { id: 'account', icon: Cloud, badge: 'blue' },
    { id: 'system', icon: Cpu, badge: 'emerald' },
    { id: 'about', icon: Info, badge: 'blue' },
    { id: 'updates', icon: Download, badge: 'amber' }
  ]
</script>

<div class="sidebar-nav-list" role="tablist">
  {#each categories as cat (cat.id)}
    {@const isActive = $activeSettingsTabStore === cat.id}
    <button
      type="button"
      role="tab"
      aria-selected={isActive}
      class="sidebar-nav-btn"
      class:active={isActive}
      onclick={() => ($activeSettingsTabStore = cat.id)}
    >
      <IconBadge icon={cat.icon} color={cat.badge} size="sm" />
      <span class="sidebar-nav-label">{$tStore('settings.tabs.' + cat.id)}</span
      >
    </button>
  {/each}
</div>

<style src="../styles/SidebarLayout.css"></style>
