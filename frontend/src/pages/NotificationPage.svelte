<!--
  @fileoverview Page container hosting the Notification history and management view.

  Configures the expanded island viewport dimensions and mounts the interactive Notification component.
-->
<script module lang="ts">
  export const config = {
    minimal: { w: 360, h: 65, r: 18 },
    expanded: { w: 440, h: 380, r: 30 },
    borderColor: 'var(--vinta-island-border-notification)',
    glowColor: 'var(--vinta-island-glow-notification)'
  }
</script>

<script lang="ts">
  import { getContext } from 'svelte'
  import { ChevronUp } from 'lucide-svelte'
  import Notification from '../components/Notification.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import { notifData, tStore } from '../store'
  import type { IslandConfig } from '../store'
  import HomePage, { config as homeConfig } from './HomePage.svelte'

  let { expanded = false }: { expanded?: boolean } = $props()

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')
</script>

<div
  data-vinta="notification-page"
  class="notification-layout"
  class:minimal={!expanded}
>
  <div class="notif-content">
    <Notification {expanded} notifications={$notifData} />
  </div>

  {#if !expanded}
    <div class="shrink-action">
      <button
        class="collapse-btn"
        use:islandTooltip={$tStore('notification.dismiss_tooltip')}
        onclick={(e: MouseEvent) => {
          if (e) e.stopPropagation()
          navigate(HomePage, homeConfig, false)
        }}
      >
        <ChevronUp size={14} />
      </button>
    </div>
  {/if}
</div>

<style src="./styles/NotificationPage.css"></style>
