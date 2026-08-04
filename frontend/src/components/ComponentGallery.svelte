<script lang="ts">
  import type { Snippet, Component } from 'svelte'
  import { ChevronLeft, ChevronRight } from 'lucide-svelte'
  import { swipe } from '../utils/useSwipe'

  export interface SlideItem {
    id?: string
    snippet?: Snippet
    component?: Component<Record<string, unknown>>
    props?: Record<string, unknown>
  }

  let {
    slides = [],
    autoPlay = false,
    autoPlayInterval = 4000,
    loop = true,
    showControls = true,
    showDots = true,
    initialIndex = 0,
    enableSwipe = true,
    onSlideChange = () => {}
  } = $props<{
    slides: SlideItem[]
    autoPlay?: boolean
    autoPlayInterval?: number
    loop?: boolean
    showControls?: boolean
    showDots?: boolean
    initialIndex?: number
    enableSwipe?: boolean
    onSlideChange?: (index: number) => void
  }>()

  let isLoopingable = $derived(loop && slides.length > 1)

  let clonedSlides = $derived.by(() => {
    if (!isLoopingable || slides.length === 0) return slides
    const first = slides[0]
    const last = slides[slides.length - 1]
    return [last, ...slides, first]
  })

  // Internal 1-based index when looping to account for left clone
  let displayIndex = $state(initialIndex + (loop && slides.length > 1 ? 1 : 0))
  let isInstant = $state(false)
  let isDragging = $state(false)
  let isAnimating = $state(false)
  let dragOffset = $state(0)
  let isHovered = $state(false)
  let autoPlayTimer: number

  // Synchronize when slides array changes dynamically
  $effect(() => {
    if (isLoopingable) {
      if (displayIndex < 0 || displayIndex >= clonedSlides.length) {
        displayIndex = 1
      }
    } else {
      if (displayIndex < 0 || displayIndex >= slides.length) {
        displayIndex = 0
      }
    }
  })

  // Calculated real 0-based index for external consumers & pagination dots
  let activeRealIndex = $derived.by(() => {
    if (!isLoopingable || slides.length === 0) return Math.max(0, displayIndex)
    if (displayIndex <= 0) return slides.length - 1
    if (displayIndex >= clonedSlides.length - 1) return 0
    return displayIndex - 1
  })

  // Notify parent component on real slide change
  $effect(() => {
    onSlideChange(activeRealIndex)
  })

  function goToSlide(targetRealIndex: number) {
    if (slides.length === 0) return
    isInstant = false
    isAnimating = true
    displayIndex = isLoopingable ? targetRealIndex + 1 : targetRealIndex
  }

  function nextSlide() {
    if (slides.length === 0 || isAnimating) return
    if (!isLoopingable) {
      if (displayIndex < slides.length - 1) {
        isAnimating = true
        displayIndex++
      }
      return
    }
    if (displayIndex >= clonedSlides.length - 1) return
    isInstant = false
    isAnimating = true
    displayIndex++
  }

  function prevSlide() {
    if (slides.length === 0 || isAnimating) return
    if (!isLoopingable) {
      if (displayIndex > 0) {
        isAnimating = true
        displayIndex--
      }
      return
    }
    if (displayIndex <= 0) return
    isInstant = false
    isAnimating = true
    displayIndex--
  }

  function handleTransitionEnd(e?: TransitionEvent) {
    if (e && (e.target !== e.currentTarget || e.propertyName !== 'transform'))
      return
    if (!isLoopingable) {
      isAnimating = false
      return
    }

    if (displayIndex <= 0) {
      isInstant = true
      displayIndex = clonedSlides.length - 2
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          isInstant = false
          isAnimating = false
        })
      })
    } else if (displayIndex >= clonedSlides.length - 1) {
      isInstant = true
      displayIndex = 1
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          isInstant = false
          isAnimating = false
        })
      })
    } else {
      isAnimating = false
    }
  }

  const swipeOptions = $derived({
    axis: 'x' as const,
    enableSwipe: enableSwipe && slides.length > 1,
    ignoreSelector: '.nav-btn, .dot-btn, select, option, .no-swipe',
    onDragStart: () => {
      isDragging = true
      isInstant = false
      dragOffset = 0
    },
    onDragMove: (offset: { x: number; y: number }) => {
      dragOffset = offset.x
    },
    onDragEnd: () => {
      isDragging = false
      dragOffset = 0
    },
    onSwipeLeft: () => {
      nextSlide()
    },
    onSwipeRight: () => {
      prevSlide()
    }
  })

  // Auto-play logic
  $effect(() => {
    if (autoPlay && !isHovered && !isDragging && slides.length > 1) {
      autoPlayTimer = window.setInterval(() => {
        nextSlide()
      }, autoPlayInterval)
    } else {
      clearInterval(autoPlayTimer)
    }

    return () => {
      clearInterval(autoPlayTimer)
    }
  })
</script>

<div
  class="gallery-container"
  class:is-dragging={isDragging}
  class:is-swiping={isDragging}
  class:is-instant={isInstant}
  use:swipe={swipeOptions}
  onmouseenter={() => (isHovered = true)}
  onmouseleave={() => (isHovered = false)}
  role="region"
  aria-label="Component Slideshow"
>
  <div class="gallery-viewport">
    <div
      class="gallery-track"
      ontransitionend={handleTransitionEnd}
      style="
        transform: translateX(calc(-{displayIndex * 100}% + {dragOffset}px));
        transition: {isDragging || isInstant
        ? 'none'
        : 'transform 0.38s cubic-bezier(0.22, 1, 0.36, 1)'};
      "
    >
      {#each clonedSlides as slide, i (i)}
        <div
          class="gallery-slide"
          class:active={i === displayIndex ||
            (!isLoopingable && i === displayIndex)}
        >
          {#if slide.snippet}
            {@render slide.snippet()}
          {:else if slide.component}
            <slide.component {...slide.props || {}} />
          {/if}
        </div>
      {/each}
    </div>

    {#if showControls && slides.length > 1}
      <button
        class="nav-btn prev"
        type="button"
        onclick={(e) => {
          e.stopPropagation()
          prevSlide()
        }}
        disabled={!loop && activeRealIndex === 0}
        aria-label="Previous Slide"
      >
        <ChevronLeft size={16} />
      </button>

      <button
        class="nav-btn next"
        type="button"
        onclick={(e) => {
          e.stopPropagation()
          nextSlide()
        }}
        disabled={!loop && activeRealIndex === slides.length - 1}
        aria-label="Next Slide"
      >
        <ChevronRight size={16} />
      </button>
    {/if}
  </div>

  {#if showDots && slides.length > 1}
    <div class="pagination-dots">
      {#each slides as slide, i (slide.id || i)}
        <button
          class="dot-btn"
          type="button"
          class:active={i === activeRealIndex}
          onclick={(e) => {
            e.stopPropagation()
            goToSlide(i)
          }}
          aria-label="Go to slide {i + 1}"
        ></button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .gallery-container {
    position: relative;
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    user-select: none;
    touch-action: pan-y;
    cursor: grab;
  }

  .gallery-container.is-dragging {
    cursor: grabbing;
  }

  .gallery-container.is-instant .gallery-slide {
    transition: none !important;
  }

  .gallery-viewport {
    position: relative;
    flex: 1;
    min-height: 0;
    width: 100%;
    overflow: hidden;
    border-radius: 16px;
  }

  .gallery-track {
    display: flex;
    width: 100%;
    height: 100%;
    will-change: transform;
  }

  .gallery-slide {
    flex: 0 0 100%;
    width: 100%;
    height: 100%;
    padding: 0;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    opacity: 0.4;
    transform: scale(0.98);
    transition:
      opacity 0.35s ease,
      transform 0.35s ease;
    pointer-events: none;
  }

  .gallery-slide.active {
    opacity: 1;
    transform: scale(1);
    pointer-events: auto;
  }

  /* Navigation buttons - overlayed inside viewport */
  .nav-btn {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: var(--c-surface-elevated, rgba(30, 30, 40, 0.85));
    color: var(--c-text-muted, rgba(255, 255, 255, 0.7));
    border: 1px solid var(--c-border, rgba(255, 255, 255, 0.12));
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    z-index: 10;
    opacity: 0;
    pointer-events: none;
    transition:
      opacity 0.2s ease,
      background-color 0.2s ease,
      color 0.2s ease,
      transform 0.2s cubic-bezier(0.22, 1, 0.36, 1);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.18);
  }

  .gallery-container:hover .nav-btn {
    opacity: 0.9;
    pointer-events: auto;
  }

  .nav-btn:hover {
    opacity: 1 !important;
    background: var(--c-hover, rgba(50, 50, 65, 0.95));
    color: var(--c-text-primary, #ffffff);
    transform: translateY(-50%) scale(1.1);
  }

  .nav-btn:active {
    transform: translateY(-50%) scale(0.95);
  }

  .nav-btn:disabled {
    opacity: 0.15 !important;
    cursor: not-allowed;
    transform: translateY(-50%) !important;
    pointer-events: none;
  }

  .nav-btn.prev {
    left: 8px;
  }

  .nav-btn.next {
    right: 8px;
  }

  /* Pagination dots bar */
  .pagination-dots {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding-top: 6px;
    padding-bottom: 2px;
    flex-shrink: 0;
    z-index: 25;
  }

  .dot-btn {
    width: 6px;
    height: 6px;
    border-radius: 3px;
    background: var(--c-text-muted, rgba(255, 255, 255, 0.25));
    opacity: 0.4;
    border: none;
    padding: 0;
    cursor: pointer;
    pointer-events: auto;
    transition:
      width 0.3s cubic-bezier(0.22, 1, 0.36, 1),
      background-color 0.3s ease,
      opacity 0.3s ease;
  }

  .dot-btn:hover {
    opacity: 0.8;
    background: var(--c-text-primary, rgba(255, 255, 255, 0.6));
  }

  .dot-btn.active {
    width: 18px;
    opacity: 1;
    background: var(--c-primary, #6366f1);
  }
</style>
