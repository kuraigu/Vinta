<!--
  @component SpotlightCarousel
  @description Horizontal carousel slider with elastic momentum swipe gestures and dot indicators.
-->
<script lang="ts">
  import { onMount, onDestroy, getContext, type ComponentType } from 'svelte'
  import {
    Heart,
    Star,
    Palette,
    MessageSquareHeart,
    ChevronRight,
    ChevronLeft
  } from 'lucide-svelte'
  import { Browser } from '@wailsio/runtime'
  import { activeSettingsTabStore, tStore, type IslandConfig } from '../store'
  import SettingsPage, {
    config as settingsConfig
  } from '../pages/SettingsPage.svelte'
  import IconBadge, { type IconBadgeColor } from './IconBadge.svelte'
  import { OpenURL } from '../../bindings/Vinta/backend/app.js'
  import { swipe } from '../utils/useSwipe'
  import { reportFeedback } from '../services/feedbackService'

  // Correct asset imports matching your folder structure
  import enohotelLogoSvg from '../illustrations/enohotel-daniel-belda/simple_logo.svg'
  import enohotelBgSvg from '../illustrations/enohotel-daniel-belda/enohotel-card-bg.svg'

  export interface SpotlightItem {
    id: string
    title: string
    description: string
    iconType: 'component' | 'image'
    iconComponent?: ComponentType<any>
    iconSrc?: string
    badgeColorName?: IconBadgeColor
    badgeBg: string
    badgeColor: string
    url?: string
    internalTab?: string
    action?: () => void | Promise<void>
  }

  interface Props {
    items?: SpotlightItem[]
  }

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let defaultSpotlightItems: SpotlightItem[] = $derived([
    {
      id: 'support',
      title: $tStore('spotlight.support_title'),
      description: $tStore('spotlight.support_desc'),
      iconType: 'component',
      iconComponent: Heart,
      badgeColorName: 'rose',
      badgeBg: 'var(--vinta-danger, #ef4444)',
      badgeColor: '#ffffff',
      url: 'https://ko-fi.com/dippedinsoju'
    },
    {
      id: 'github',
      title: $tStore('spotlight.github_title'),
      description: $tStore('spotlight.github_desc'),
      iconType: 'component',
      iconComponent: Star,
      badgeColorName: 'amber',
      badgeBg: 'var(--vinta-warning, #f59e0b)',
      badgeColor: '#ffffff',
      url: 'https://github.com/kuraigu/Vinta'
    },
    {
      id: 'enohotel',
      title: $tStore('spotlight.enohotel_title'),
      description: $tStore('spotlight.enohotel_desc'),
      iconType: 'image',
      iconSrc: enohotelLogoSvg,
      badgeColorName: 'amber',
      badgeBg: `url("${enohotelBgSvg}")`,
      badgeColor: '#936B19',
      url: 'https://enohoteldanielbelda.com/'
    },
    {
      id: 'appearance',
      title: $tStore('spotlight.appearance_title'),
      description: $tStore('spotlight.appearance_desc'),
      iconType: 'component',
      iconComponent: Palette,
      badgeColorName: 'cyan',
      badgeBg: 'var(--vinta-accent-primary, #38bdf8)',
      badgeColor: '#ffffff',
      internalTab: 'appearance'
    },
    {
      id: 'feedback',
      title: $tStore('spotlight.feedback_title'),
      description: $tStore('spotlight.feedback_desc'),
      iconType: 'component',
      iconComponent: MessageSquareHeart,
      badgeColorName: 'emerald',
      badgeBg: 'var(--vinta-success, #10b981)',
      badgeColor: '#ffffff',
      action: () => void reportFeedback()
    }
  ])

  let { items }: Props = $props()
  let computedItems = $derived(items ?? defaultSpotlightItems)

  let currentIndex = $state(0)
  let wrapEl = $state<HTMLElement | null>(null)
  let trackEl = $state<HTMLElement | null>(null)
  let currentTranslateX = $state(0)
  let dragOffset = $state(0)
  let isDragging = $state(false)
  let isInstant = $state(true)
  let resizeObserver: ResizeObserver | null = null
  let resizeRafId: number | undefined
  let swipedThisGesture = false

  function goToIndex(index: number, smooth = true) {
    if (!trackEl || !wrapEl) return
    const cards = trackEl.querySelectorAll<HTMLElement>('.spotlight-card')
    const targetCard = cards[index]
    const wrapWidth = wrapEl.clientWidth

    if (targetCard && wrapWidth > 0 && targetCard.offsetWidth > 0) {
      const wrapCenter = wrapWidth / 2
      const cardCenter = targetCard.offsetLeft + targetCard.offsetWidth / 2
      const targetX = Math.round(wrapCenter - cardCenter)

      if (!smooth) {
        isInstant = true
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            isInstant = false
          })
        })
      } else {
        isInstant = false
      }

      currentTranslateX = targetX
      currentIndex = index
    }
  }

  function handlePrev(e?: MouseEvent) {
    if (e) e.stopPropagation()
    if (currentIndex <= 0) return
    goToIndex(currentIndex - 1, true)
  }

  function handleNext(e?: MouseEvent) {
    if (e) e.stopPropagation()
    if (currentIndex >= computedItems.length - 1) return
    goToIndex(currentIndex + 1, true)
  }

  let lastMeasuredWidth = 0

  function handleTrackResize(entries: ResizeObserverEntry[]) {
    const entry = entries[0]
    if (!entry || !wrapEl || isDragging) return
    const newWidth = Math.round(entry.contentRect.width)
    if (newWidth <= 0 || newWidth === lastMeasuredWidth) return
    lastMeasuredWidth = newWidth

    if (resizeRafId !== undefined) {
      cancelAnimationFrame(resizeRafId)
    }
    resizeRafId = requestAnimationFrame(() => {
      if (!wrapEl || isDragging) return
      goToIndex(currentIndex, false)
    })
  }

  function initializeCarousel() {
    if (computedItems.length > 1) {
      currentIndex = Math.floor(Math.random() * computedItems.length)
    }

    if (wrapEl && typeof ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(handleTrackResize)
      resizeObserver.observe(wrapEl)
    }

    requestAnimationFrame(() => {
      goToIndex(currentIndex, false)
    })
  }

  onMount(initializeCarousel)

  onDestroy(() => {
    if (resizeObserver) {
      resizeObserver.disconnect()
      resizeObserver = null
    }
    if (resizeRafId !== undefined) {
      cancelAnimationFrame(resizeRafId)
    }
  })

  let lastItemsLength = 0
  function handleItemsChangedEffect() {
    const len = computedItems.length
    if (len > 0 && len !== lastItemsLength && wrapEl) {
      lastItemsLength = len
      if (currentIndex >= len) {
        currentIndex = 0
      }
      requestAnimationFrame(() => {
        if (!isDragging) {
          goToIndex(currentIndex, false)
        }
      })
    }
  }
  $effect(handleItemsChangedEffect)

  async function handleCardClick(item: SpotlightItem) {
    if (item.action) {
      await item.action()
      return
    }

    if (item.url) {
      try {
        await Browser.OpenURL(item.url)
        return
      } catch {
        // Fallback
      }

      try {
        if (typeof OpenURL === 'function') {
          await OpenURL(item.url)
          return
        }
      } catch {
        // Fallback
      }

      window.open(item.url, '_blank')
      return
    }

    if (item.internalTab && navigate) {
      $activeSettingsTabStore = item.internalTab
      navigate(SettingsPage, settingsConfig, true)
    }
  }

  function handleSwipeNext() {
    swipedThisGesture = true
    if (currentIndex < computedItems.length - 1) {
      goToIndex(currentIndex + 1, true)
    } else {
      goToIndex(currentIndex, true)
    }
  }

  function handleSwipePrev() {
    swipedThisGesture = true
    if (currentIndex > 0) {
      goToIndex(currentIndex - 1, true)
    } else {
      goToIndex(currentIndex, true)
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
      (currentIndex === computedItems.length - 1 && dx < 0)
    ) {
      dx *= 0.35
    }
    dragOffset = dx
  }

  function handleDragEnd() {
    isDragging = false
    dragOffset = 0
    if (!swipedThisGesture) {
      goToIndex(currentIndex, true)
    }
  }
</script>

<div bind:this={wrapEl} class="spotlight-carousel-wrap">
  <div
    bind:this={trackEl}
    class="spotlight-scroll-track"
    role="region"
    aria-label="Spotlight carousel"
    style="
      transform: translateX({currentTranslateX + dragOffset}px);
      transition: {isDragging || isInstant
      ? 'none'
      : 'transform 0.48s cubic-bezier(0.2, 1.15, 0.35, 1)'};
    "
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
    {#each computedItems as item (item.id)}
      <button
        type="button"
        class="spotlight-card"
        style={item.badgeBg && item.badgeBg.startsWith('url(')
          ? `background: ${item.badgeBg}; background-size: cover;`
          : ''}
        onclick={() => void handleCardClick(item)}
        aria-label="{item.title}: {item.description}"
      >
        <div class="spotlight-card-body">
          {#if item.iconType === 'image' && item.iconSrc}
            <IconBadge
              size="lg"
              color={item.badgeColorName ?? 'blue'}
              customBg={item.badgeColorName ? undefined : item.badgeBg}
              customColor={item.badgeColorName ? undefined : item.badgeColor}
            >
              <img src={item.iconSrc} alt="" loading="lazy" />
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

        <div class="spotlight-chevron">
          <ChevronRight size={15} strokeWidth={2.4} />
        </div>
      </button>
    {/each}
  </div>

  {#if computedItems.length > 1}
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
        {#each computedItems as it, idx (it.id)}
          <button
            type="button"
            class="spotlight-dot"
            class:active={idx === currentIndex}
            onclick={(e) => {
              e.stopPropagation()
              goToIndex(idx, true)
            }}
            aria-label="Go to card {idx + 1}"
          ></button>
        {/each}
      </div>

      <button
        type="button"
        class="spotlight-arrow-btn"
        disabled={currentIndex >= computedItems.length - 1}
        onclick={handleNext}
        aria-label="Next card"
      >
        <ChevronRight size={13} strokeWidth={2.4} />
      </button>
    </div>
  {/if}
</div>

<style src="./styles/SpotlightCarousel.css"></style>
