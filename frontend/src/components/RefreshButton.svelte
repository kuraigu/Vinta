<!--
  @component RefreshButton
  @description Icon button that spins during network refresh operations (weather, calendars, feeds).
-->
<script lang="ts">
  import { RefreshCw } from 'lucide-svelte'
  import { islandTooltip } from '../services/islandTooltip'

  interface Props {
    onclick?: (e: MouseEvent) => void
    loading?: boolean
    title?: string
    size?: number
  }

  let {
    onclick,
    loading = false,
    title = 'Refresh',
    size = 14
  }: Props = $props()

  function handleClick(e: MouseEvent) {
    if (onclick) {
      onclick(e)
    }
  }
</script>

<button
  type="button"
  class="refresh-btn"
  class:spinning={loading}
  onclick={handleClick}
  use:islandTooltip={title}
  aria-label={title}
>
  <RefreshCw {size} />
</button>

<style src="./styles/RefreshButton.css"></style>
