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
    ignoreSelector = 'input, select, textarea, range, .custom-select, .no-swipe, .no-drag, [draggable="true"]',
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

  let dragScrollRaf: number | null = null
  let dragScrollDelta = 0

  function stopDragAutoScroll() {
    dragScrollDelta = 0
    if (dragScrollRaf !== null) {
      cancelAnimationFrame(dragScrollRaf)
      dragScrollRaf = null
    }
  }

  function stepDragAutoScroll() {
    if (!element || dragScrollDelta === 0) {
      dragScrollRaf = null
      return
    }
    const prevTop = element.scrollTop
    element.scrollTop += dragScrollDelta
    if (element.scrollTop === prevTop) {
      dragScrollRaf = null
      return
    }
    dragScrollRaf = requestAnimationFrame(stepDragAutoScroll)
  }

  function handleContainerDragOver(e: DragEvent) {
    if (!element || element.scrollHeight <= element.clientHeight) {
      stopDragAutoScroll()
      return
    }
    if (!e.dataTransfer?.types?.includes('text/vinta-widget')) return
    e.preventDefault()

    const rect = element.getBoundingClientRect()
    const edgePx = 64
    let nextDelta = 0

    if (e.clientY < rect.top + edgePx && element.scrollTop > 0) {
      const ratio = Math.min(1, (rect.top + edgePx - e.clientY) / edgePx)
      nextDelta = -(4 + ratio * 10)
    } else if (
      e.clientY > rect.bottom - edgePx &&
      element.scrollTop + element.clientHeight < element.scrollHeight
    ) {
      const ratio = Math.min(1, (e.clientY - (rect.bottom - edgePx)) / edgePx)
      nextDelta = 4 + ratio * 10
    }

    if (nextDelta === 0) {
      stopDragAutoScroll()
      return
    }

    dragScrollDelta = nextDelta
    if (dragScrollRaf === null) {
      dragScrollRaf = requestAnimationFrame(stepDragAutoScroll)
    }
  }

  function handleContainerDragLeave(e: DragEvent) {
    const related = e.relatedTarget as Node | null
    if (related && element?.contains(related)) return
    stopDragAutoScroll()
  }

  onDestroy(() => {
    stopDragAutoScroll()
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
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      bind:this={element}
      class="scroll-view {className}"
      class:hide-scrollbar={hideScrollbar}
      use:swipe={swipeOptions}
      onscroll={handleInternalScroll}
      ondragover={handleContainerDragOver}
      ondragleave={handleContainerDragLeave}
      ondrop={stopDragAutoScroll}
      ondragend={stopDragAutoScroll}
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
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    bind:this={element}
    class="scroll-view {className}"
    class:hide-scrollbar={hideScrollbar}
    use:swipe={swipeOptions}
    onscroll={handleInternalScroll}
    ondragover={handleContainerDragOver}
    ondragleave={handleContainerDragLeave}
    ondrop={stopDragAutoScroll}
    ondragend={stopDragAutoScroll}
  >
    {#if children}
      {@render children()}
    {/if}
  </div>
{/if}

<style src="./styles/ScrollView.css"></style>
