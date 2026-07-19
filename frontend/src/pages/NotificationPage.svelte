<script module lang="ts">
  export const config = {
    minimal: { w: 360, h: 65, r: 24 },
    expanded: { w: 420, h: 400, r: 35 },
    borderColor: 'var(--c-island-border-notification)',
    glowColor: 'var(--c-island-glow-notification)'
  }
</script>

<script lang="ts">
  import { getContext } from 'svelte'
  import { Shrink } from 'lucide-svelte'
  import { AppControls } from '../utils/AppControls.svelte'
  import DockButton from '../components/DockButton.svelte' // Note: Default import (removed curly braces)
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
      <DockButton
        variant="warning"
        tooltip=""
        onclick={(e: MouseEvent) => AppControls.doMinimize(navigate, e)}
      >
        <Shrink size={18} />
      </DockButton>
    </div>
  {/if}
</div>

<style>
  .notification-layout {
    width: 100%;
    height: 100%;
    box-sizing: border-box;
  }

  /* When minimal (!expanded), display horizontally */
  .notification-layout.minimal {
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    padding: 0 12px; /* Adds breathing room from the island borders */
  }

  /* Let the notification take up all available space on the left */
  .notification-layout.minimal .notif-content {
    flex: 1;
    min-width: 0; /* Prevents text overflow issues in flexbox */
    margin-right: 8px;
  }

  /* Keep the dock button neatly centered on the right */
  .shrink-action {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0; /* Prevents the button from getting squished */
  }
</style>
