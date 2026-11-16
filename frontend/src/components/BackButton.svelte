<!--
  @component BackButton
  @description Standard chevron button for navigating back to the home or previous page.
-->
<script lang="ts">
  import { getContext } from 'svelte'
  import { ArrowLeft } from 'lucide-svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import { tStore } from '../store'
  import type { IslandConfig } from '../store'

  interface Props {
    onback?: () => void
    title?: string
  }

  let { onback, title }: Props = $props()

  const goBack = getContext<(() => void) | undefined>('goBack')
  const openOverlay = getContext<
    ((overlay: 'notifications' | 'apps' | null) => void) | undefined
  >('openOverlay')

  let resolvedTitle = $derived(title || $tStore('common.back') || 'Back')

  function handleClick() {
    if (onback) {
      onback()
    } else if (typeof goBack === 'function') {
      goBack()
    } else if (typeof openOverlay === 'function') {
      openOverlay('apps')
    }
  }
</script>

<button
  class="back-btn"
  onclick={handleClick}
  use:islandTooltip={resolvedTitle}
>
  <ArrowLeft size={14} />
</button>

<style src="./styles/BackButton.css"></style>
