<script lang="ts">
  import { swipe } from '../utils/useSwipe'
  import type { Snippet } from 'svelte'

  let {
    children,
    class: className = '',
    enableSwipe = true,
    ignoreSelector = 'input, select, textarea, range, .custom-select, .no-swipe, .no-drag'
  } = $props<{
    children?: Snippet
    class?: string
    enableSwipe?: boolean
    ignoreSelector?: string
  }>()

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
  }

  .scroll-view::-webkit-scrollbar {
    width: 4px;
  }
  .scroll-view::-webkit-scrollbar-track {
    background: transparent;
  }
  .scroll-view::-webkit-scrollbar-thumb {
    background: var(--c-scrollbar-thumb, rgba(100, 100, 100, 0.4));
    border-radius: 4px;
  }
</style>
