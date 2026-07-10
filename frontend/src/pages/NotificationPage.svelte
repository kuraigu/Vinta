<script context="module" lang="ts">
  export const config = {
    minimal: { w: 320, h: 65, r: 24 },
    expanded: { w: 400, h: 250, r: 35 }
  }
</script>

<script lang="ts">
  import Notification from '../components/Notification.svelte'
  import { notifData } from '../store'

  export let expanded: boolean = false
</script>


<Notification 
  {expanded} 
  notifications={$notifData} 
/>
