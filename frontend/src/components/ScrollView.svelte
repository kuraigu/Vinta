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
    hideIndicatorOnScroll?: boolean
    scrollIndicatorLabel?: string
    scrollIndicatorAriaLabel?: string
    scrollIndicatorCompact?: boolean
    scrollStep?: number
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
    hideIndicatorOnScroll = false,
    scrollIndicatorLabel,
    scrollIndicatorAriaLabel,
    scrollIndicatorCompact = false,
    scrollStep = 110,
    ignoreSelector = 'input, select, textarea, range, .custom-select, .no-swipe, .no-drag',
    onscroll,
    element = $bindable()
  }: Props = $props()

  let canScrollUp = $state(false)
  let canScrollDown = $state(false)
  let isScrolledDown = $state(false)
  let resizeObserver: ResizeObserver | null = null
  let mutationObserver: MutationObserver | null = null
  let scrollStateRaf: number | null = null

  let swipeOptions = $derived({
    axis: 'y' as const,
    enableSwipe,
    ignoreSelector
  })

  function updateScrollState() {
    if (!element || !showScrollIndicators) {
      canScrollUp = false
      canScrollDown = false
      isScrolledDown = false
      return
    }
    const { scrollTop, scrollHeight, clientHeight } = element
    canScrollUp = scrollTop > 4
    isScrolledDown = scrollTop > 20
    canScrollDown = scrollHeight - scrollTop - clientHeight > 8
  }

  function handleInternalScroll(e: Event) {
    if (showScrollIndicators && scrollStateRaf === null) {
      scrollStateRaf = requestAnimationFrame(() => {
        scrollStateRaf = null
        updateScrollState()
      })
    }
    onscroll?.(e)
  }

  function scrollDownStep(e: MouseEvent) {
    e.stopPropagation()
    element?.scrollBy({ top: scrollStep, behavior: 'smooth' })
  }

  onMount(() => {
    if (showScrollIndicators && element) {
      const container = element
      updateScrollState()
      const observeChildren = () => {
        if (!resizeObserver) return
        for (const child of Array.from(container.children)) {
          resizeObserver.observe(child)
        }
      }
      if (typeof ResizeObserver !== 'undefined') {
        resizeObserver = new ResizeObserver(() => updateScrollState())
        resizeObserver.observe(container)
        observeChildren()
      }
      if (typeof MutationObserver !== 'undefined') {
        mutationObserver = new MutationObserver(() => {
          observeChildren()
          updateScrollState()
        })
        mutationObserver.observe(container, { childList: true })
      }
    }
  })

  onDestroy(() => {
    if (scrollStateRaf !== null) {
      cancelAnimationFrame(scrollStateRaf)
      scrollStateRaf = null
    }
    mutationObserver?.disconnect()
    mutationObserver = null
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
    <ScrollIndicator
      visible={canScrollDown && (!hideIndicatorOnScroll || !isScrolledDown)}
      onclick={scrollDownStep}
      label={scrollIndicatorLabel}
      ariaLabel={scrollIndicatorAriaLabel}
      compact={scrollIndicatorCompact}
    />
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
