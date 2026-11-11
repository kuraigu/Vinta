<!--
  @component ScrollIndicator
  @description Floating scroll cue pill indicating more content is available below.
-->
<script lang="ts">
  import { ChevronDown } from 'lucide-svelte'
  import { tStore } from '../store'

  interface Props {
    onclick?: (e: MouseEvent) => void
    label?: string
    ariaLabel?: string
    visible?: boolean
    compact?: boolean
    class?: string
  }

  let {
    onclick,
    label,
    ariaLabel,
    visible = true,
    compact = false,
    class: className = ''
  }: Props = $props()

  let isCompact = $derived(compact || !label)
  let resolvedAriaLabel = $derived(
    ariaLabel || label || $tStore('home.scroll_cue_aria') || 'Scroll down'
  )
</script>

<button
  type="button"
  class="vinta-scroll-indicator {className}"
  class:compact={isCompact}
  class:hidden={!visible}
  {onclick}
  aria-label={resolvedAriaLabel}
>
  {#if label && !compact}
    <span class="scroll-indicator-text">{label}</span>
  {/if}
  <ChevronDown
    size={isCompact ? 12 : 13}
    class="scroll-indicator-icon"
    strokeWidth={2.4}
  />
</button>

<style src="./styles/ScrollIndicator.css"></style>
