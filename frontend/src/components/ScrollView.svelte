<!--
  @component ScrollView
  @description Touch and mouse drag scroll container. Provides smooth momentum scrolling
  with hidden scrollbars and customizable fade indicators at top and bottom edges.
-->
<script lang="ts">
  import { onMount, onDestroy, type Snippet } from 'svelte'
  import { swipe } from '../utils/useSwipe'
  import ScrollIndicator from './ScrollIndicator.svelte'

  interface Props {
    children?: Snippet
    class?: string
    enableSwipe?: boolean
    hideScrollbar?: boolean
    showScrollIndicators?: boolean
    scrollIndicatorLabel?: string
    scrollIndicatorCompact?: boolean
    ignoreSelector?: string
    onscroll?: (e: Event) => void
    element?: HTMLElement
  }

  let {
    children,
    class: className = '',
    enableSwipe = true,
    hideScrollbar = false,
    showScrollIndicators = false,
    scrollIndicatorLabel,
    scrollIndicatorCompact = false,
    ignoreSelector = 'input, select, textarea, range, .custom-select, .no-swipe, .no-drag',
    onscroll,
    element = $bindable()
  }: Props = $props()

  let canScrollUp = $state(false)
  let canScrollDown = $state(false)
  let resizeObserver: ResizeObserver | null = null

  let swipeOptions = $derived({
    axis: 'y' as const,
    enableSwipe,
    ignoreSelector
  })

  function updateScrollState() {
    if (!element || !showScrollIndicators) {
      canScrollUp = false
      canScrollDown = false
      return
    }
    const { scrollTop, scrollHeight, clientHeight } = element
    canScrollUp = scrollTop > 4
    canScrollDown = scrollHeight - scrollTop - clientHeight > 4
  }

  function handleInternalScroll(e: Event) {
    if (showScrollIndicators) {
      updateScrollState()
    }
    onscroll?.(e)
  }

  function scrollDownStep(e: MouseEvent) {
    e.stopPropagation()
    element?.scrollBy({ top: 110, behavior: 'smooth' })
  }

  onMount(() => {
    if (showScrollIndicators && element) {
      updateScrollState()
      if (typeof ResizeObserver !== 'undefined') {
        resizeObserver = new ResizeObserver(() => updateScrollState())
        resizeObserver.observe(element)
        if (element.firstElementChild) {
          resizeObserver.observe(element.firstElementChild)
        }
      }
    }
  })

  onDestroy(() => {
    resizeObserver?.disconnect()
    resizeObserver = null
  })
</script>

{#if showScrollIndicators}
  <div
    class="scroll-view-host"
    class:has-top-overflow={canScrollUp}
    class:has-bottom-overflow={canScrollDown}
  >
    <div
      bind:this={element}
      class="scroll-view {className}"
      class:hide-scrollbar={hideScrollbar}
      use:swipe={swipeOptions}
      onscroll={handleInternalScroll}
    >
      {#if children}
        {@render children()}
      {/if}
    </div>
    {#if canScrollDown}
      <ScrollIndicator
        onclick={scrollDownStep}
        label={scrollIndicatorLabel}
        compact={scrollIndicatorCompact}
      />
    {/if}
  </div>
{:else}
  <div
    bind:this={element}
    class="scroll-view {className}"
    class:hide-scrollbar={hideScrollbar}
    use:swipe={swipeOptions}
    onscroll={handleInternalScroll}
  >
    {#if children}
      {@render children()}
    {/if}
  </div>
{/if}

<style src="./styles/ScrollView.css"></style>
