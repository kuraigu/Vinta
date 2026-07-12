<script module lang="ts">
  // This page dictates its own dimensions to the parent App!
  export const config = {
    minimal: { w: 330, h: 35, r: 35 },
    expanded: { w: 580, h: 300, r: 35 },
    borderColor: 'var(--c-success)'
  }
</script>

<script lang="ts">
  import Clock from '../components/Clock.svelte'
  import Divider from '../components/Divider.svelte'
  import Calendar from '../components/Calendar.svelte'
  import Weather from '../components/Weather.svelte'
  import BentoGrid from '../components/BentoGrid.svelte'
  import BentoItem from '../components/BentoItem.svelte'

  let { expanded = false } = $props<{ expanded?: boolean }>()
</script>

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
      <BentoGrid columns="200px 1fr" gap="1rem">
        <BentoItem>
          <Calendar expanded={true} />
        </BentoItem>
        <BentoGrid columns="1fr" rows="1fr 1fr" gap="1rem">
          <BentoItem>
            <Clock expanded={true} />
          </BentoItem>
          <BentoItem>
            <Weather expanded={true} />
          </BentoItem>
        </BentoGrid>
      </BentoGrid>
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
    min-width: 0; /* Allows text truncation if it ever gets squished */
  }

  .expanded-view {
    padding: 1rem;
    width: 100%;
    height: 100%;
    box-sizing: border-box;
  }
</style>
