<script module lang="ts">
  export const config = {
    minimal: { w: 330, h: 35, r: 35 },
    expanded: { w: 440, h: 370, r: 30 },
    borderColor: 'var(--c-island-border-idle)',
    glowColor: 'var(--c-island-glow-idle)'
  }
</script>

<script lang="ts">
  import Clock from '../components/Clock.svelte'
  import Calendar from '../components/Calendar.svelte'
  import Weather from '../components/Weather.svelte'
  import BentoItem from '../components/BentoItem.svelte'
  import ComponentGallery from '../components/ComponentGallery.svelte'

  let { expanded = false } = $props<{ expanded?: boolean }>()

  const slideTitles = ['Clock', 'Calendar', 'Weather']
  let activeSlideTitle = $state(slideTitles[0])

  function handleSlideChange(index: number) {
    activeSlideTitle = slideTitles[index] || 'Home'
  }
</script>

{#snippet clockSlide()}
  <div class="single-card-container">
    <BentoItem>
      <Clock expanded={true} />
    </BentoItem>
  </div>
{/snippet}

{#snippet calendarSlide()}
  <div class="single-card-container">
    <BentoItem>
      <Calendar expanded={true} />
    </BentoItem>
  </div>
{/snippet}

{#snippet weatherSlide()}
  <div class="single-card-container">
    <BentoItem>
      <Weather expanded={true} />
    </BentoItem>
  </div>
{/snippet}

<div class="page-container">
  {#if !expanded}
    <div class="minimal-view">
      <div class="side left">
        <Clock expanded={false} />
      </div>
      <div class="center">
        <Weather expanded={false} />
      </div>
      <div class="side right">
        <Calendar expanded={false} />
      </div>
    </div>
  {:else}
    <div class="expanded-view">
      <div class="header">
        <div class="title">{activeSlideTitle}</div>
      </div>
      <div class="gallery-wrapper">
        <ComponentGallery
          slides={[
            { id: 'clock', snippet: clockSlide },
            { id: 'calendar', snippet: calendarSlide },
            { id: 'weather', snippet: weatherSlide }
          ]}
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
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
  }

  .minimal-view {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1.5rem;
    width: 100%;
    height: 100%;
    margin-top: -2px; /* Visual optical adjustment for font baseline */
    padding: 0 1.2rem;
    box-sizing: border-box;
    white-space: nowrap; /* Prevents text breaking */
  }

  .side,
  .center {
    display: flex;
    align-items: center;
    overflow: hidden;
    min-width: 0;
  }

  .minimal-none-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--c-primary);
    opacity: 0.8;
  }

  .expanded-view {
    width: 100%;
    height: 100%;
    padding: 14px 16px 44px 16px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
  }

  .header {
    display: flex;
    justify-content: center;
    align-items: center;
    margin-bottom: 10px;
    flex-shrink: 0;
  }

  .title {
    font-size: 15px;
    font-weight: 600;
    letter-spacing: -0.01em;
    color: var(--c-text-primary);
  }

  .gallery-wrapper {
    flex: 1;
    min-height: 0;
    width: 100%;
  }

  .single-card-container {
    width: 100%;
    height: 100%;
    box-sizing: border-box;
  }
</style>
