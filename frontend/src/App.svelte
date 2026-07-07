<script lang="ts">
  import { onMount } from 'svelte'
  import { SetWidgetBounds } from '../wailsjs/go/backend/App.js'
  import ClockPage from './pages/ClockPage.svelte'

  // Dynamic Routing: The active app that the island is displaying
  let activePage = ClockPage

  // App-specific dimensions (dynamically pulled from the active page via binding!)
  let islandConfig = {
    minimal: { w: 300, h: 35, r: 10 },
    expanded: { w: 600, h: 150, r: 35 }
  }

  let expanded = false
  let islandNode: HTMLElement

  onMount(() => {
    // Force OS clipping to match our minimal config immediately on startup
    SetWidgetBounds(
      islandConfig.minimal.w,
      islandConfig.minimal.h,
      islandConfig.minimal.r
    )
  })

  function handleExpand() {
    if (expanded) return
    // Update click region first (instant, no OS resize = no flicker), then let CSS animate
    SetWidgetBounds(
      islandConfig.expanded.w,
      islandConfig.expanded.h,
      islandConfig.expanded.r
    )
    expanded = true
  }

  function handleShrink() {
    if (!expanded) return
    expanded = false

    // Wait for the CSS animation to finish before shrinking the click region
    const onTransitionEnd = (e: TransitionEvent) => {
      if (e.target !== islandNode) return
      islandNode.removeEventListener('transitionend', onTransitionEnd)
      if (!expanded) {
        SetWidgetBounds(
          islandConfig.minimal.w,
          islandConfig.minimal.h,
          islandConfig.minimal.r
        )
      }
    }
    islandNode.addEventListener('transitionend', onTransitionEnd)
  }

  function toggle() {
    if (expanded) {
      handleShrink()
    } else {
      handleExpand()
    }
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<div
  bind:this={islandNode}
  class="island {expanded ? 'expanded' : ''}"
  style="
    --island-width: {expanded
    ? islandConfig.expanded.w
    : islandConfig.minimal.w}px;
    --island-height: {expanded
    ? islandConfig.expanded.h
    : islandConfig.minimal.h}px;
    --island-radius: {expanded
    ? islandConfig.expanded.r
    : islandConfig.minimal.r}px;
  "
  on:click={toggle}
>
  <div class="content">
    <svelte:component this={activePage} {expanded} bind:config={islandConfig} />
  </div>
</div>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    height: 100vh;
    overflow: hidden;
  }

  .island {
    background-color: #000000;
    border-radius: var(--island-radius);
    cursor: pointer;
    overflow: hidden;

    width: var(--island-width);
    height: var(--island-height);

    /* Smooth spring-like transition */
    transition: all 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);

    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    pointer-events: auto; /* Re-enable clicks on the island itself */
  }

  .content {
    /* Smooth fade in for content */
    animation: fadeIn 0.4s ease-in-out;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
  }

  @keyframes fadeIn {
    0% {
      opacity: 0;
      transform: scale(0.9);
    }
    50% {
      opacity: 0;
    }
    100% {
      opacity: 1;
      transform: scale(1);
    }
  }
</style>
