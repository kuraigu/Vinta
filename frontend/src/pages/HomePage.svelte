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
    flex-direction: column;
    justify-content: center;
    align-items: center;
    position: relative;
    box-sizing: border-box;
  }

  /* Minimal View - 100% Symmetrical Left & Right margins */
  .minimal-view {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    max-width: 100%;
    height: 100%;
    padding: 0 16px;
    box-sizing: border-box;
  }

  .side.left {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    flex: 1 1 0;
    min-width: 0;
    white-space: nowrap;
  }

  .center {
    display: flex;
    align-items: center;
    justify-content: center;
    flex: 1 1 0;
    min-width: 0;
    white-space: nowrap;
  }

  .side.right {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    flex: 1 1 0;
    min-width: 0;
    white-space: nowrap;
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
    justify-content: center;
    align-items: center;
    padding: 12px 16px 4px 16px;
  }

  .title {
    font-size: 0.85rem;
    font-weight: 700;
    letter-spacing: 0.03em;
    opacity: 0.85;
    color: var(--c-text-primary);
    text-align: center;
  }

  .gallery-wrapper {
    flex: 1;
    min-height: 0;
    width: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
    position: relative;
  }

  .single-card-container {
    width: 100%;
    height: 100%;
    padding: 0 16px;
    box-sizing: border-box;
    display: flex;
    justify-content: center;
    align-items: center;
  }
</style>
