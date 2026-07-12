<script lang="ts">
  import { onMount, setContext } from 'svelte'
  import { fade } from 'svelte/transition'
  import { SetWidgetBounds, IsTestMode } from '../wailsjs/go/backend/App.js'
  import ClockPage, { config as clockConfig } from './pages/ClockPage.svelte'
  import DevPill from './devtest/DevPill.svelte'
  import Colors from './Colors.svelte'
  import { themeStore, shapeStore } from './store'

  $effect(() => {
    document.documentElement.setAttribute('data-theme', $themeStore)
  })

  import type { Component } from 'svelte'
  import type { IslandConfig } from './store'

  // Dynamic Routing: The active app that the island is displaying
  let activePage = $state<Component>(ClockPage)

  let islandConfig = $state<IslandConfig>(clockConfig)

  let expanded = $state(false)
  let islandNode = $state<HTMLElement>()
  let isTestMode = $state(false)

  let isTransitionGlow = $state(true)
  let glowTimeout: number

  $effect(() => {
    // Re-run this effect whenever expansion or page changes
    const _p = activePage

    isTransitionGlow = true
    clearTimeout(glowTimeout)
    glowTimeout = window.setTimeout(() => {
      isTransitionGlow = false
    }, 2500) // Glow plays for 2.5 seconds during transition
  })

  function updateBounds(w: number, h: number, r: number) {
    const computedR = $shapeStore === 'box' ? 12 : r
    const scale = window.devicePixelRatio || 1
    SetWidgetBounds(w, h, computedR, scale)
  }

  // React to shape changes immediately, but don't react to 'expanded' changes
  // to avoid instantly shrinking the OS window region and clipping the animation.
  shapeStore.subscribe(() => {
    if (!islandNode) return
    const currentW = expanded ? islandConfig.expanded.w : islandConfig.minimal.w
    const currentH = expanded ? islandConfig.expanded.h : islandConfig.minimal.h
    const currentR = expanded ? islandConfig.expanded.r : islandConfig.minimal.r
    updateBounds(currentW, currentH, currentR)
  })

  // Global navigate for dev pill / pages
  setContext(
    'navigate',
    (page: Component, pageConfig: IslandConfig, expand: boolean = false) => {
      // Save old config values before updating
      const oldW = expanded ? islandConfig.expanded.w : islandConfig.minimal.w
      const oldH = expanded ? islandConfig.expanded.h : islandConfig.minimal.h

      activePage = page
      islandConfig = pageConfig
      expanded = expand

      const newW = expanded ? islandConfig.expanded.w : islandConfig.minimal.w
      const newH = expanded ? islandConfig.expanded.h : islandConfig.minimal.h
      const newR = expanded ? islandConfig.expanded.r : islandConfig.minimal.r

      if (newW >= oldW && newH >= oldH) {
        // Growing in all dimensions (or same). Safe to update OS region immediately.
        updateBounds(newW, newH, newR)
      } else {
        // Shrinking in at least one dimension.
        // Create a temporary bounding box so the CSS animation isn't clipped!
        const maxW = Math.max(oldW, newW)
        const maxH = Math.max(oldH, newH)
        updateBounds(maxW, maxH, newR)

        const onTransitionEnd = (e: TransitionEvent) => {
          if (e.target !== islandNode) return
          islandNode.removeEventListener('transitionend', onTransitionEnd)
          // Ensure we haven't navigated again while waiting
          if (activePage === page && expanded === expand) {
            updateBounds(newW, newH, newR)
          }
        }
        islandNode.addEventListener('transitionend', onTransitionEnd)
      }
    }
  )

  onMount(async () => {
    // Check if launched via `scripts.bat test`
    isTestMode = await IsTestMode()
    updateBounds(
      islandConfig.minimal.w,
      islandConfig.minimal.h,
      islandConfig.minimal.r
    )
  })

  function handleExpand() {
    if (expanded) return
    expanded = true
    updateBounds(
      islandConfig.expanded.w,
      islandConfig.expanded.h,
      islandConfig.expanded.r
    )
  }

  function handleShrink() {
    if (!expanded) return
    expanded = false

    // Wait for CSS animation to finish before shrinking the OS click region
    const onTransitionEnd = (e: TransitionEvent) => {
      if (e.target !== islandNode) return
      islandNode.removeEventListener('transitionend', onTransitionEnd)
      if (!expanded) {
        updateBounds(
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

<div class="island-container">
  <Colors />
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    bind:this={islandNode}
    class="island {expanded ? 'expanded' : ''} {isTransitionGlow
      ? 'is-transition-glow'
      : ''}"
    style="
      width: {expanded ? islandConfig.expanded.w : islandConfig.minimal.w}px;
      height: {expanded ? islandConfig.expanded.h : islandConfig.minimal.h}px;
      --island-r: {$shapeStore === 'box'
      ? 12
      : expanded
        ? islandConfig.expanded.r
        : islandConfig.minimal.r}px;
      border-radius: var(--island-r);
      --island-border-color: {islandConfig.borderColor || 'transparent'};
    "
    onclick={toggle}
  >
    <div class="content">
      {#if !expanded}
        {@const ActivePage = activePage}
        <div
          class="view"
          in:fade={{ duration: 200, delay: 350 }}
          out:fade={{ duration: 100 }}
          style="width: {islandConfig.minimal.w}px; height: {islandConfig
            .minimal.h}px;"
        >
          <ActivePage expanded={false} />
        </div>
      {:else}
        {@const ActivePage = activePage}
        <div
          class="view"
          in:fade={{ duration: 200, delay: 350 }}
          out:fade={{ duration: 100 }}
          style="width: {islandConfig.expanded.w}px; height: {islandConfig
            .expanded.h}px;"
        >
          <ActivePage expanded={true} />
        </div>
      {/if}
    </div>
  </div>

  {#if isTestMode}
    <DevPill />
  {/if}
</div>

<style>
  /* Remove translateZ from root to prevent transparent viewport memory bloat */
  :global(html) {
    box-sizing: border-box;
  }

  :global(body) {
    margin: 0;
    padding: 0;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    height: 100vh;
    overflow: hidden;
    font-family:
      'Inter',
      system-ui,
      -apple-system,
      sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }

  .island-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 15px; /* Matches the gap in backend/app.go */
  }

  .island {
    --ambient-glow-ring:
      inset 0 0 0 1px var(--island-border-color),
      inset 0 0 0 1px var(--island-border-color);

    margin-top: 5px; /* Matches visualTop in backend/app.go */
    cursor: pointer;
    position: relative;
    background: var(--c-bg);

    background-clip: padding-box;
    -webkit-background-clip: padding-box;

    border: none;
    box-shadow:
      inset 0 1px 0 0 rgba(255, 255, 255, 0.15),
      /* Top highlight */ inset 0 0 0 1px rgba(255, 255, 255, 0.08),
      /* Subtle border */ 0 5px 5px rgba(0, 0, 0, 0.5); /* Drop shadow */

    transition:
      filter 0.4s ease,
      box-shadow 0.4s ease,
      --island-border-color 0.3s ease,
      width 0.4s cubic-bezier(0.22, 1, 0.36, 1),
      height 0.4s cubic-bezier(0.22, 1, 0.36, 1),
      border-radius 0.4s cubic-bezier(0.22, 1, 0.36, 1);

    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--c-text-primary);
    pointer-events: auto; /* Re-enable clicks on the island itself */

    transform: translateZ(0);
    -webkit-transform: translateZ(0);

    will-change: width, height, box-shadow;

    contain: layout style;

    font-variant-numeric: tabular-nums;
  }

  @property --angle {
    syntax: '<angle>';
    initial-value: 0deg;
    inherits: false;
  }

  @keyframes orbit {
    to {
      --angle: 360deg;
    }
  }

  .island::before {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: inherit;
    box-shadow: var(--ambient-glow-ring);

    /* 
     * NOTE: We must inline the mask instead of using a CSS variable. 
     * Browsers have a bug where `@property` animations fail to update 
     * if nested inside an unregistered CSS variable applied to a mask.
     */
    -webkit-mask: conic-gradient(
      from var(--angle),
      transparent 0%,
      black 50%,
      transparent 100%
    );
    mask: conic-gradient(
      from var(--angle),
      transparent 0%,
      black 50%,
      transparent 100%
    );
    pointer-events: none;
    animation: orbit 3s linear infinite;
    opacity: 0;
    transition: opacity 0.3s ease;
  }

  .island.is-transition-glow::before {
    opacity: 1;
  }

  .island:hover::before {
    opacity: 0; /* Hide glow when the solid hover ring activates */
  }

  .island:hover {
    box-shadow:
      inset 0 0 0 1px var(--island-border-color),
      inset 0 0 0 1px var(--island-border-color),
      0 0 8px var(--island-border-color),
      /* Neon glow effect */ 0 8px 16px rgba(0, 0, 0, 0.6);
  }

  .content {
    display: grid;
    place-items: center;
    width: 100%;
    height: 100%;
    border-radius: inherit;
    overflow: hidden;
    position: relative;
  }

  .view {
    grid-area: 1 / 1;
    display: flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
  }
</style>
