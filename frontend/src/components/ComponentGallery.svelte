<script lang="ts">
  import { type Component, type Snippet } from 'svelte'
  import { ChevronLeft, ChevronRight } from 'lucide-svelte'
  import { swipe } from '../utils/useSwipe'

  export interface SlideItem {
    id: string
    title?: string
    snippet?: Snippet
    component?: Component<{ expanded?: boolean; [key: string]: unknown }>
    props?: Record<string, unknown>
  }

  let {
    slides = [],
    initialIndex = 0,
    autoPlay = false,
    autoPlayInterval = 5000,
    showControls = true,
    showDots = true,
    enableSwipe = true,
    loop = true,
    onSlideChange
  } = $props<{
    slides: SlideItem[]
    initialIndex?: number
    autoPlay?: boolean
    autoPlayInterval?: number
    showControls?: boolean
    showDots?: boolean
    enableSwipe?: boolean
    loop?: boolean
    onSlideChange?: (index: number) => void
  }>()

  // Build cloned array for seamless infinite looping if loop=true & slides > 1
  let isLoopingable = $derived(loop && slides.length > 1)

  let clonedSlides = $derived.by(() => {
    if (!isLoopingable) return slides
    const lastClone = {
      ...slides[slides.length - 1],
      id: `${slides[slides.length - 1].id}-clone-last`
    }
    const firstClone = { ...slides[0], id: `${slides[0].id}-clone-first` }
    return [lastClone, ...slides, firstClone]
  })

  // Internal index tracking (1-based when looping to account for left clone)
  let displayIndex = $state(initialIndex + (isLoopingable ? 1 : 0))
  let isInstant = $state(false)
  let isDragging = $state(false)
  let dragOffset = $state(0)
  let isHovered = $state(false)

  let autoPlayTimer: number | undefined

  // Active real index for dots & callbacks (0-based)
  let activeRealIndex = $derived.by(() => {
    if (slides.length === 0) return 0
    if (!isLoopingable) return displayIndex
    if (displayIndex === 0) return slides.length - 1
    if (displayIndex === clonedSlides.length - 1) return 0
    return displayIndex - 1
  })

  function goToSlide(targetRealIndex: number) {
    if (slides.length === 0) return
    const target = isLoopingable ? targetRealIndex + 1 : targetRealIndex
    if (displayIndex !== target) {
      isInstant = false
      displayIndex = target
      onSlideChange?.(targetRealIndex)
    }
  }

  function nextSlide() {
    if (slides.length <= 1) return
    isInstant = false
    displayIndex++
    onSlideChange?.(activeRealIndex)
  }

  function prevSlide() {
    if (slides.length <= 1) return
    isInstant = false
    displayIndex--
    onSlideChange?.(activeRealIndex)
  }

  function handleTransitionEnd() {
    if (!isLoopingable) return
    // If we landed on the right clone (first slide clone), jump to real first slide
    if (displayIndex >= clonedSlides.length - 1) {
      isInstant = true
      displayIndex = 1
    }
    // If we landed on the left clone (last slide clone), jump to real last slide
    else if (displayIndex <= 0) {
      isInstant = true
      displayIndex = slides.length
    }
  }

  let showArrowsTemp = $state(true)
  let hideArrowsTimeout: number

  // Show arrows briefly on mount / expand
  $effect(() => {
    showArrowsTemp = true
    clearTimeout(hideArrowsTimeout)
    hideArrowsTimeout = window.setTimeout(() => {
      showArrowsTemp = false
    }, 2500)
    return () => clearTimeout(hideArrowsTimeout)
  })

  // Swipe Action configuration
  let swipeOptions = $derived({
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
  class:show-temp-arrows={showArrowsTemp}
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
        : 'transform 0.4s cubic-bezier(0.22, 1, 0.36, 1)'};
      "
    >
      {#each clonedSlides as slide, i (slide.id || i)}
        <div
          class="gallery-slide"
          class:active={i === displayIndex ||
            (isLoopingable &&
              ((displayIndex === 0 && i === clonedSlides.length - 2) ||
                (displayIndex === clonedSlides.length - 1 && i === 1)))}
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
    padding: 0 6px;
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

  /* Navigation buttons - Theme aware styling */
  .nav-btn {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    width: 26px;
    height: 26px;
    border-radius: 50%;
    background: var(--c-dock-popover-bg, rgba(25, 25, 30, 0.8));
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid var(--c-border, rgba(255, 255, 255, 0.18));
    color: var(--c-text-primary, #ffffff);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    opacity: 0;
    transition:
      opacity 0.3s ease,
      background-color 0.2s ease,
      transform 0.2s ease;
    z-index: 25;
    pointer-events: auto;
    padding: 0;
    box-shadow: var(--c-dock-popover-shadow, 0 4px 12px rgba(0, 0, 0, 0.4));
  }

  /* Show arrows when expanded initially or when hovering container */
  .gallery-container.show-temp-arrows .nav-btn,
  .gallery-container:hover .nav-btn {
    opacity: 0.75;
  }

  /* Hide arrows completely when actively swiping/dragging */
  .gallery-container.is-swiping .nav-btn {
    opacity: 0 !important;
    pointer-events: none;
  }

  .nav-btn:hover:not(:disabled) {
    opacity: 1 !important;
    background: var(--c-hover, rgba(50, 50, 65, 0.95));
    color: var(--c-text-primary, #ffffff);
    transform: translateY(-50%) scale(1.1);
  }

  .nav-btn:disabled {
    opacity: 0.15 !important;
    cursor: not-allowed;
  }

  .nav-btn.prev {
    left: 2px;
  }

  .nav-btn.next {
    right: 2px;
  }

  /* Pagination dots bar - Theme aware styling */
  .pagination-dots {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding-top: 8px;
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
