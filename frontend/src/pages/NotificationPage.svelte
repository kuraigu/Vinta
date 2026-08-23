<script module lang="ts">
  export const config = {
    minimal: { w: 360, h: 65, r: 24 },
    expanded: { w: 440, h: 380, r: 30 },
    borderColor: 'var(--c-island-border-notification)',
    glowColor: 'var(--c-island-glow-notification)'
  }
</script>

<script lang="ts">
  import { getContext } from 'svelte'
  import { ChevronUp } from 'lucide-svelte'
  import { AppControls } from '../utils/AppControls.svelte'
  import Notification from '../components/Notification.svelte'
  import { notifData } from '../store'
  import type { IslandConfig } from '../store'

  let { expanded = false }: { expanded?: boolean } = $props()

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')
</script>

<div class="notification-layout" class:minimal={!expanded}>
  <div class="notif-content">
    <Notification {expanded} notifications={$notifData} />
  </div>

  {#if !expanded}
    <div class="shrink-action">
      <button
        class="collapse-btn"
        title="Collapse notification"
        onclick={(e: MouseEvent) => AppControls.doMinimize(navigate, e)}
      >
        <ChevronUp size={14} />
      </button>
    </div>
  {/if}
</div>

<style src="./styles/NotificationPage.css"></style>
