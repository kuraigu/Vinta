<script lang="ts">
  import { swipe } from '../utils/useSwipe'
  import type { Snippet } from 'svelte'

  interface Props {
    children?: Snippet
    class?: string
    enableSwipe?: boolean
    ignoreSelector?: string
  }

  let {
    children,
    class: className = '',
    enableSwipe = true,
    ignoreSelector = 'input, select, textarea, range, .custom-select, .no-swipe, .no-drag'
  }: Props = $props()

  let swipeOptions = $derived({
    axis: 'y' as const,
    enableSwipe,
    ignoreSelector
  })
</script>

<div class="scroll-view {className}" use:swipe={swipeOptions}>
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
    width: 3px;
  }
  .scroll-view::-webkit-scrollbar-track {
    background: transparent;
  }
  .scroll-view::-webkit-scrollbar-thumb {
    background: var(--c-scrollbar-thumb, rgba(255, 255, 255, 0.18));
    border-radius: 3px;
    transition: background 0.2s ease;
  }
  .scroll-view:hover::-webkit-scrollbar-thumb {
    background: var(--c-scrollbar-thumb-hover, rgba(255, 255, 255, 0.35));
  }
</style>
