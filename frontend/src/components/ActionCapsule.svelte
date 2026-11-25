<!--
  @component ActionCapsule
  @description Compact clickable pill button used for inline quick actions and filter tags.
-->
<script lang="ts">
  import type { Component } from 'svelte'
  import { ChevronRight } from 'lucide-svelte'
  import { islandTooltip } from '../services/islandTooltip'

  interface Props {
    label: string
    icon?: Component
    tooltip?: string
    variant?: 'default' | 'accent' | 'frosted'
    onclick?: (e: MouseEvent) => void
    class?: string
  }

  let {
    label,
    icon: IconComponent,
    tooltip,
    variant = 'default',
    onclick,
    class: className = ''
  }: Props = $props()
</script>

<button
  type="button"
  class="action-capsule-btn variant-{variant} {className}"
  {onclick}
  use:islandTooltip={tooltip || ''}
>
  {#if IconComponent}
    <IconComponent size={12} class="capsule-icon" />
  {/if}
  <span class="capsule-label">{label}</span>
  <ChevronRight size={14} class="capsule-arrow" />
</button>

<style src="./styles/ActionCapsule.css"></style>
