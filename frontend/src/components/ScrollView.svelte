<script lang="ts">
  import { swipe } from '../utils/useSwipe'
  import type { Snippet } from 'svelte'

  interface Props {
    children?: Snippet
    class?: string
    enableSwipe?: boolean
    ignoreSelector?: string
    onscroll?: (e: Event) => void
    element?: HTMLElement
  }

  let {
    children,
    class: className = '',
    enableSwipe = true,
    ignoreSelector = 'input, select, textarea, range, .custom-select, .no-swipe, .no-drag',
    onscroll,
    element = $bindable()
  }: Props = $props()

  let swipeOptions = $derived({
    axis: 'y' as const,
    enableSwipe,
    ignoreSelector
  })
</script>

<div
  bind:this={element}
  class="scroll-view {className}"
  use:swipe={swipeOptions}
  {onscroll}
>
  {#if children}
    {@render children()}
  {/if}
</div>

<style>
  .scroll-view {
    width: 100%;
    height: 100%;
    min-height: 0;
    overflow-y: auto;
    overflow-x: hidden;
    position: relative;
    overscroll-behavior-y: contain;
    touch-action: pan-y;
  }

  .scroll-view::-webkit-scrollbar {
    width: 5px;
  }
  .scroll-view::-webkit-scrollbar-track {
    background: transparent;
  }
  .scroll-view::-webkit-scrollbar-thumb {
    background: var(--vinta-scrollbar-thumb, rgba(255, 255, 255, 0.22));
    border-radius: 5px;
    transition:
      background 0.2s ease,
      width 0.2s ease;
  }
  .scroll-view:hover::-webkit-scrollbar-thumb {
    background: var(--vinta-scrollbar-thumb-hover, rgba(255, 255, 255, 0.45));
  }
</style>
