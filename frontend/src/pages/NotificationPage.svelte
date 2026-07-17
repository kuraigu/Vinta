<script module lang="ts">
  export const config = {
    minimal: { w: 360, h: 65, r: 24 },
    expanded: { w: 420, h: 400, r: 35 },
    borderColor: 'var(--c-island-border-notification)',
    glowColor: 'var(--c-island-glow-notification)'
  }
</script>

<script lang="ts">
  import Notification from '../components/Notification.svelte'
  import { notifData } from '../store'

  let { expanded = false } = $props<{ expanded?: boolean }>()
</script>

<Notification {expanded} notifications={$notifData} />
