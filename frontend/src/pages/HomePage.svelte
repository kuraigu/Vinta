<script module lang="ts">
  export const config = {
    minimal: { w: 330, h: 35, r: 35 },
    expanded: { w: 440, h: 380, r: 30 },
    borderColor: 'var(--c-island-border-idle)',
    glowColor: 'var(--c-island-glow-idle)'
  }
</script>

<script lang="ts">
  import Clock from '../components/Clock.svelte'
  import Calendar from '../components/Calendar.svelte'
  import Weather from '../components/Weather.svelte'
  import BatteryWidget from '../components/BatteryWidget.svelte'
  import ComponentGallery from '../components/ComponentGallery.svelte'
  import { homeWidgetsVisibilityStore, batteryDataStore } from '../store'

  let { expanded = false } = $props<{ expanded?: boolean }>()

  let hasBattery = $derived($batteryDataStore.hasBattery)

  let availableSlides = $derived.by(() => {
    const list = []
    if ($homeWidgetsVisibilityStore.clock) {
      list.push({ id: 'clock', snippet: clockSlide, title: 'Clock' })
    }
    if ($homeWidgetsVisibilityStore.calendar) {
      list.push({ id: 'calendar', snippet: calendarSlide, title: 'Calendar' })
    }
    if ($homeWidgetsVisibilityStore.weather) {
      list.push({ id: 'weather', snippet: weatherSlide, title: 'Weather' })
    }
    return list
  })

  let activeSlideTitle = $state('Home')

  function handleActiveSlideEffect() {
    if (
      availableSlides.length > 0 &&
      !availableSlides.some((s) => s.title === activeSlideTitle)
    ) {
      activeSlideTitle = availableSlides[0].title
    }
  }

  $effect(handleActiveSlideEffect)

  function handleSlideChange(index: number) {
    activeSlideTitle = availableSlides[index]?.title || 'Home'
  }
</script>

{#snippet clockSlide()}
  <div class="slide-bleed-card">
    <Clock expanded={true} />
  </div>
{/snippet}

{#snippet calendarSlide()}
  <div class="slide-bleed-card">
    <Calendar expanded={true} />
  </div>
{/snippet}

{#snippet weatherSlide()}
  <div class="slide-bleed-card">
    <Weather expanded={true} />
  </div>
{/snippet}

<div class="page-container">
  {#if !expanded}
    <div class="minimal-view">
      {#if $homeWidgetsVisibilityStore.clock || ($homeWidgetsVisibilityStore.battery && hasBattery)}
        <div class="minimal-slot status-cluster">
          {#if $homeWidgetsVisibilityStore.battery && hasBattery}
            <BatteryWidget />
          {/if}
          {#if $homeWidgetsVisibilityStore.clock}
            <Clock expanded={false} />
          {/if}
        </div>
      {/if}
      {#if $homeWidgetsVisibilityStore.weather}
        <div class="minimal-slot">
          <Weather expanded={false} />
        </div>
      {/if}
      {#if $homeWidgetsVisibilityStore.calendar}
        <div class="minimal-slot">
          <Calendar expanded={false} />
        </div>
      {/if}
    </div>
  {:else}
    <div class="expanded-view">
      <div class="header">
        <div class="header-side left"></div>
        <div class="header-center">
          <div class="title">{activeSlideTitle}</div>
        </div>
        <div class="header-side right"></div>
      </div>
      <div class="gallery-wrapper">
        <ComponentGallery
          slides={availableSlides}
          onSlideChange={handleSlideChange}
        />
      </div>
    </div>
  {/if}
</div>

<style>
  .page-container {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    position: relative;
    box-sizing: border-box;
  }

  /* Minimal View - Dynamic flexbox */
  .minimal-view {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 18px;
    width: 100%;
    height: 100%;
    padding: 0 18px;
    box-sizing: border-box;
  }

  .minimal-slot {
    display: flex;
    align-items: center;
    justify-content: center;
    white-space: nowrap;
    flex-shrink: 0;
  }

  .minimal-slot.status-cluster {
    gap: 6px;
  }

  /* Expanded View - Integrated footer padding for Dock */
  .expanded-view {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    padding-bottom: 58px;
    box-sizing: border-box;
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px 0 16px;
    margin-bottom: 10px;
    width: 100%;
    box-sizing: border-box;
    flex-shrink: 0;
  }

  .header-side {
    flex: 1 1 0;
    display: flex;
    align-items: center;
    min-width: 0;
  }

  .header-side.left {
    pointer-events: auto;
  }

  .header-side.right {
    justify-content: flex-end;
  }

  .header-center {
    flex: 1 1 0;
    display: flex;
    justify-content: center;
    align-items: center;
    white-space: nowrap;
  }

  .title {
    font-weight: 700;
    font-size: 15px;
    letter-spacing: -0.01em;
    text-align: center;
    color: var(--c-text-primary);
  }

  /* Gallery fills edge-to-edge — viewport clips at its own border-radius */
  .gallery-wrapper {
    flex: 1;
    min-height: 0;
    width: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
    position: relative;
    box-sizing: border-box;
  }

  /*
   * Each slide card provides its own background so it bleeds cleanly
   * to the outer viewport's border-radius: 16px clip boundary.
   */
  .slide-bleed-card {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    box-sizing: border-box;
    background: var(--c-item-bg);
    box-shadow: inset 0 1px 0 var(--c-border);
    overflow: hidden;
  }
</style>
