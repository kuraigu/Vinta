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
  import Notification from '../components/Notification.svelte'
  import { notifData } from '../store'

  let { expanded = false }: { expanded?: boolean } = $props()
</script>

<div
  data-vinta="notification-page"
  class="notification-layout"
  class:minimal={!expanded}
>
  <div class="notif-content">
    <Notification {expanded} notifications={$notifData} />
  </div>
</div>

<style src="./styles/NotificationPage.css"></style>
