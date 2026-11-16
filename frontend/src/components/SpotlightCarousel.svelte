<!--
  @component SpotlightCarousel
  @description Pure reusable peeking carousel with O(1) 3-node slide windowing,
  elastic horizontal swipe gestures, side-card focus-first clicks, and dot indicators.
-->
<script module lang="ts">
  import type { ComponentType } from 'svelte'
  import type { IconBadgeColor } from './IconBadge.svelte'

  export interface SpotlightAmbientLayer {
    fps?: number
    onBake: (
      ctx: CanvasRenderingContext2D,
      w: number,
      h: number,
      dpr: number
    ) => void | Promise<void>
    onDraw: (
      ctx: CanvasRenderingContext2D,
      w: number,
      h: number,
      t: number,
      baked: HTMLCanvasElement | null
    ) => void
  }

  export interface SpotlightItem {
    id: string
    title: string
    description: string
    iconType: 'component' | 'image'
    iconComponent?: ComponentType<any>
    iconSrc?: string
    badgeColorName?: IconBadgeColor
    badgeBg?: string
    badgeColor?: string
    url?: string
    internalTab?: string
    ambient?: SpotlightAmbientLayer
    action?: () => void | Promise<void>
  }
</script>

<script lang="ts">
  import { ChevronRight, ChevronLeft } from 'lucide-svelte'
  import IconBadge from './IconBadge.svelte'
  import BakedCanvasLayer from './BakedCanvasLayer.svelte'
  import { swipe } from '../utils/useSwipe'

  interface Props {
    items: SpotlightItem[]
    initialIndex?: number
    onSelect?: (item: SpotlightItem, index: number) => void
  }

  let {
    items = [],
    initialIndex = 0,
    onSelect = undefined
  }: Props = $props()

  function clampIndex(idx: number, len: number): number {
    if (len <= 0) return 0
    return Math.max(0, Math.min(len - 1, idx))
  }

  let currentIndex = $state(0)
  let initializedFromProp = false
  let dragOffset = $state(0)
  let isDragging = $state(false)
  let swipedThisGesture = false

  $effect(() => {
    const len = items.length
    if (len === 0) {
      currentIndex = 0
      return
    }
    if (!initializedFromProp) {
      initializedFromProp = true
      currentIndex = clampIndex(initialIndex, len)
      return
    }
    if (currentIndex >= len) {
      currentIndex = len - 1
    }
  })

  // O(1) 3-node window: only prev, active, and next cards are mounted at any time.
  let visibleWindow = $derived.by(() => {
    if (items.length === 0) return []
    const start = Math.max(0, currentIndex - 1)
    const end = Math.min(items.length - 1, currentIndex + 1)
    const result: Array<{ item: SpotlightItem; index: number; delta: number }> = []
    for (let i = start; i <= end; i++) {
      result.push({
        item: items[i],
        index: i,
        delta: i - currentIndex
      })
    }
    return result
  })

  function goToIndex(index: number) {
    currentIndex = clampIndex(index, items.length)
  }

  function handlePrev(e?: MouseEvent) {
    if (e) e.stopPropagation()
    if (currentIndex <= 0) return
    goToIndex(currentIndex - 1)
  }

  function handleNext(e?: MouseEvent) {
    if (e) e.stopPropagation()
    if (currentIndex >= items.length - 1) return
    goToIndex(currentIndex + 1)
  }

  function handleSlideClick(item: SpotlightItem, index: number) {
    if (swipedThisGesture) {
      swipedThisGesture = false
      return
    }
    if (index !== currentIndex) {
      goToIndex(index)
      return
    }
    if (onSelect) {
      onSelect(item, index)
    } else if (item.action) {
      void item.action()
    }
  }

  function handleSwipeNext() {
    swipedThisGesture = true
    if (currentIndex < items.length - 1) {
      goToIndex(currentIndex + 1)
    }
  }

  function handleSwipePrev() {
    swipedThisGesture = true
    if (currentIndex > 0) {
      goToIndex(currentIndex - 1)
    }
  }

  function handleDragStart() {
    isDragging = true
    swipedThisGesture = false
    dragOffset = 0
  }

  function handleDragMove(offset: { x: number; y: number }) {
    if (!isDragging) return
    let dx = offset.x
    if (
      (currentIndex === 0 && dx > 0) ||
      (currentIndex === items.length - 1 && dx < 0)
    ) {
      dx *= 0.35
    }
    dragOffset = dx
  }

  function handleDragEnd() {
    isDragging = false
    dragOffset = 0
  }
</script>

{#if items.length > 0}
  <div class="spotlight-carousel-wrap">
    <div
      class="spotlight-scroll-track"
      role="region"
      aria-label="Spotlight carousel"
      use:swipe={{
        axis: 'x',
        threshold: 30,
        dragThreshold: 8,
        onDragStart: handleDragStart,
        onDragMove: handleDragMove,
        onSwipeLeft: handleSwipeNext,
        onSwipeRight: handleSwipePrev,
        onDragEnd: handleDragEnd
      }}
    >
      <div class="spotlight-stage">
        {#each visibleWindow as entry (entry.item.id)}
          {@const item = entry.item}
          <button
            type="button"
            class="spotlight-card"
            class:is-side={entry.delta !== 0}
            style="
              transform: translateX(calc({entry.delta * 100}% + {entry.delta * 10}px + {dragOffset}px));
              transition: {isDragging
                ? 'none'
                : 'transform 0.48s cubic-bezier(0.2, 1.15, 0.35, 1), background 0.18s ease, border-color 0.18s ease, opacity 0.2s ease'};
            "
            onclick={() => handleSlideClick(item, entry.index)}
            aria-label="{item.title}: {item.description}"
          >
            {#if item.ambient}
              <BakedCanvasLayer
                fps={item.ambient.fps ?? 14}
                paused={entry.delta !== 0}
                onBake={item.ambient.onBake}
                onDraw={item.ambient.onDraw}
              />
            {/if}
            <div
              class="spotlight-card-body"
              style="position: relative; z-index: 1;"
            >
              {#if item.iconType === 'image' && item.iconSrc}
                <IconBadge
                  size="lg"
                  color={item.badgeColorName ?? 'blue'}
                  customBg={item.badgeColorName ? undefined : item.badgeBg}
                  customColor={item.badgeColorName ? undefined : item.badgeColor}
                >
                  <img
                    src={item.iconSrc}
                    alt=""
                    draggable="false"
                    loading="lazy"
                  />
                </IconBadge>
              {:else}
                <IconBadge
                  size="lg"
                  color={item.badgeColorName ?? 'blue'}
                  icon={item.iconComponent}
                  customBg={item.badgeColorName ? undefined : item.badgeBg}
                  customColor={item.badgeColorName ? undefined : item.badgeColor}
                />
              {/if}

              <div class="spotlight-text-block">
                <span class="spotlight-title">{item.title}</span>
                <span class="spotlight-desc">{item.description}</span>
              </div>
            </div>

            <div class="spotlight-chevron" style="position: relative; z-index: 1;">
              <ChevronRight size={15} strokeWidth={2.4} />
            </div>
          </button>
        {/each}
      </div>
    </div>

    {#if items.length > 1}
      <div class="spotlight-footer">
        <button
          type="button"
          class="spotlight-arrow-btn"
          disabled={currentIndex <= 0}
          onclick={handlePrev}
          aria-label="Previous card"
        >
          <ChevronLeft size={13} strokeWidth={2.4} />
        </button>

        <div class="spotlight-dots">
          {#each items as it, idx (it.id)}
            <button
              type="button"
              class="spotlight-dot"
              class:active={idx === currentIndex}
              onclick={(e) => {
                e.stopPropagation()
                goToIndex(idx)
              }}
              aria-label="Go to card {idx + 1}"
            ></button>
          {/each}
        </div>

        <button
          type="button"
          class="spotlight-arrow-btn"
          disabled={currentIndex >= items.length - 1}
          onclick={handleNext}
          aria-label="Next card"
        >
          <ChevronRight size={13} strokeWidth={2.4} />
        </button>
      </div>
    {/if}
  </div>
{/if}

<style src="./styles/SpotlightCarousel.css"></style>
