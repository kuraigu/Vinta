<script lang="ts">
  import { onMount } from 'svelte'
  import { fade } from 'svelte/transition'
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
    // Set initial OS click region to match the minimal pill
    SetWidgetBounds(
      islandConfig.minimal.w,
      islandConfig.minimal.h,
      islandConfig.minimal.r
    )
  })

  function handleExpand() {
    if (expanded) return
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

    // Wait for CSS animation to finish before shrinking the OS click region
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
    {#if !expanded}
      <div
        class="view"
        in:fade={{ duration: 200, delay: 350 }}
        out:fade={{ duration: 100 }}
        style="width: {islandConfig.minimal.w}px; height: {islandConfig.minimal
          .h}px;"
      >
        <svelte:component
          this={activePage}
          expanded={false}
          bind:config={islandConfig}
        />
      </div>
    {:else}
      <div
        class="view"
        in:fade={{ duration: 200, delay: 350 }}
        out:fade={{ duration: 100 }}
        style="width: {islandConfig.expanded.w}px; height: {islandConfig
          .expanded.h}px;"
      >
        <svelte:component
          this={activePage}
          expanded={true}
          bind:config={islandConfig}
        />
      </div>
    {/if}
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
    margin-top: 5px;
    background-color: #000000;
    border-radius: var(--island-radius);
    cursor: pointer;

    width: var(--island-width);
    height: var(--island-height);

    box-sizing: border-box;
    border: 1px solid rgba(255, 255, 255, 0.08);
    box-shadow: 0 0 0 transparent;

    /* Smooth Apple-like ease-out transition without the spring bounce */
    transition: width 0.4s cubic-bezier(0.22, 1, 0.36, 1),
      height 0.4s cubic-bezier(0.22, 1, 0.36, 1),
      border-radius 0.4s cubic-bezier(0.22, 1, 0.36, 1), border-color 0.3s ease,
      box-shadow 0.3s ease;

    /* Hardware acceleration to prevent frame drops */
    will-change: width, height, border-radius;

    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    pointer-events: auto; /* Re-enable clicks on the island itself */
  }

  .island:hover {
    border-color: rgba(255, 255, 255, 0.25);
    box-shadow: 0 0 12px rgba(255, 255, 255, 0.18);
  }

  .content {
    display: grid;
    place-items: center;
    width: 100%;
    height: 100%;
    border-radius: inherit;
    overflow: hidden;
  }

  .view {
    grid-area: 1 / 1;
    display: flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
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
