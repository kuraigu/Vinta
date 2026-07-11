<script module lang="ts">
  export const config = {
    minimal: { w: 360, h: 65, r: 24 },
    expanded: { w: 420, h: 300, r: 35 }
  }
</script>

<script lang="ts">
  import Notification from '../components/Notification.svelte'
  import { notifData } from '../store'

  let { expanded = false } = $props<{ expanded?: boolean }>()
</script>

<Notification {expanded} notifications={$notifData} />
