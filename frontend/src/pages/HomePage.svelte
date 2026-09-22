<script module lang="ts">
  export const config = {
    minimal: { w: 330, h: 35, r: 16 },
    expanded: { w: 440, h: 380, r: 30 },
    borderColor: 'var(--c-island-border-idle)',
    glowColor: 'var(--c-island-glow-idle)'
  }
</script>

<script lang="ts">
  import Clock from '../components/Clock.svelte'
  import Calendar from '../components/Calendar.svelte'
  import Weather from '../components/Weather.svelte'
  import DeviceStatusWidget from '../components/DeviceStatusWidget.svelte'
  import ComponentGallery from '../components/ComponentGallery.svelte'
  import { homeWidgetsVisibilityStore } from '../store'

  let { expanded = false } = $props<{ expanded?: boolean }>()

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
      <div class="minimal-slot">
        <DeviceStatusWidget />
      </div>
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

<style src="./styles/HomePage.css"></style>
