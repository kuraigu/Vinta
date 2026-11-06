<!--
  @component SpotlightCarousel
  @description Horizontal carousel slider with momentum swipe gestures and dot indicators.
-->
<script lang="ts">
  import { onMount, getContext, type ComponentType } from 'svelte'
  import {
    Heart,
    Star,
    Palette,
    ChevronRight,
    ChevronLeft
  } from 'lucide-svelte'
  import { Browser } from '@wailsio/runtime'
  import { activeSettingsTabStore, tStore, type IslandConfig } from '../store'
  import SettingsPage, {
    config as settingsConfig
  } from '../pages/SettingsPage.svelte'
  import { OpenURL } from '../../bindings/Vinta/backend/app.js'
  import { swipe } from '../utils/useSwipe'

  /**
   * Data representation for a spotlight promotion banner.
   */
  export interface SpotlightItem {
    id: string
    title: string
    description: string
    iconType: 'component' | 'image'
    iconComponent?: ComponentType<any>
    iconSrc?: string
    badgeBg: string
    badgeColor: string
    url?: string
    internalTab?: string
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
      badgeBg: 'var(--vinta-warning, #f59e0b)',
      badgeColor: '#ffffff',
      url: 'https://github.com/kuraigu/Vinta'
    },
    {
      id: 'appearance',
      title: $tStore('spotlight.appearance_title'),
      description: $tStore('spotlight.appearance_desc'),
      iconType: 'component',
      iconComponent: Palette,
      badgeBg: 'var(--vinta-accent-primary, #38bdf8)',
      badgeColor: '#0b1120',
      internalTab: 'appearance'
    }
  ])

  let { items }: Props = $props()
  let computedItems = $derived(items ?? defaultSpotlightItems)

  let currentIndex = $state(0)
  let trackEl = $state<HTMLElement | null>(null)

  /**
   * Scrolls the track to the card at the target index.
   */
  function scrollToIndex(index: number, smooth = true) {
    if (!trackEl) return
    const cards = trackEl.querySelectorAll<HTMLElement>('.spotlight-card')
    const targetCard = cards[index]
    if (targetCard) {
      const trackCenter = trackEl.clientWidth / 2
      const cardCenter = targetCard.offsetLeft + targetCard.offsetWidth / 2
      const targetScrollLeft = cardCenter - trackCenter
      trackEl.scrollTo({
        left: targetScrollLeft,
        behavior: smooth ? 'smooth' : 'auto'
      })
      currentIndex = index
    }
  }

  /**
   * Scrolls to the previous spotlight card without wrapping.
   */
  function handlePrev(e?: MouseEvent) {
    if (e) e.stopPropagation()
    if (currentIndex <= 0) return
    scrollToIndex(currentIndex - 1, true)
  }

  /**
   * Scrolls to the next spotlight card without wrapping.
   */
  function handleNext(e?: MouseEvent) {
    if (e) e.stopPropagation()
    if (currentIndex >= computedItems.length - 1) return
    scrollToIndex(currentIndex + 1, true)
  }

  /**
   * Initializes active card with a random index and scrolls without animation.
   */
  function initializeRandomIndex() {
    if (computedItems.length > 1) {
      currentIndex = Math.floor(Math.random() * computedItems.length)
      requestAnimationFrame(() => {
        scrollToIndex(currentIndex, false)
      })
    }
  }

  onMount(initializeRandomIndex)

  /**
   * Synchronizes active index with the centered card on scroll.
   */
  function handleScroll() {
    if (!trackEl) return
    const center = trackEl.scrollLeft + trackEl.clientWidth / 2
    const cards = trackEl.querySelectorAll<HTMLElement>('.spotlight-card')

    let closestIdx = 0
    let closestDist = Infinity

    cards.forEach((card, idx) => {
      const cardCenter = card.offsetLeft + card.offsetWidth / 2
      const dist = Math.abs(center - cardCenter)
      if (dist < closestDist) {
        closestDist = dist
        closestIdx = idx
      }
    })

    if (
      closestIdx !== currentIndex &&
      closestIdx >= 0 &&
      closestIdx < computedItems.length
    ) {
      currentIndex = closestIdx
    }
  }

  /**
   * Handles user click on a spotlight card.
   */
  async function handleCardClick(item: SpotlightItem) {
    if (item.url) {
      try {
        await Browser.OpenURL(item.url)
        return
      } catch {
        // Fallback to binding or window.open
      }

      try {
        if (typeof OpenURL === 'function') {
          await OpenURL(item.url)
          return
        }
      } catch {
        // Fallback to window.open
      }

      window.open(item.url, '_blank')
      return
    }

    if (item.internalTab && navigate) {
      $activeSettingsTabStore = item.internalTab
      navigate(SettingsPage, settingsConfig, true)
    }
  }
</script>

<div class="spotlight-carousel-wrap">
  <div
    bind:this={trackEl}
    class="spotlight-scroll-track"
    role="region"
    aria-label="Spotlight carousel"
    use:swipe={{
      axis: 'x',
      threshold: 30,
      dragThreshold: 8,
      onSwipeLeft: () => handleNext(),
      onSwipeRight: () => handlePrev(),
      onDragEnd: () => scrollToIndex(currentIndex, true)
    }}
    onscroll={handleScroll}
  >
    {#each computedItems as item (item.id)}
      <button
        type="button"
        class="spotlight-card"
        onclick={() => void handleCardClick(item)}
        aria-label="{item.title}: {item.description}"
      >
        <div class="spotlight-card-body">
          <div
            class="spotlight-icon-badge"
            style="background: {item.badgeBg}; color: {item.badgeColor};"
          >
            {#if item.iconType === 'image' && item.iconSrc}
              <img src={item.iconSrc} alt="" loading="lazy" />
            {:else if item.iconComponent}
              {@const Icon = item.iconComponent}
              <Icon size={13} strokeWidth={2.2} />
            {/if}
          </div>

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
              scrollToIndex(idx)
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
