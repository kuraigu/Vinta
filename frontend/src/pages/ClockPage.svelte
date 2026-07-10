<script context="module" lang="ts">
  // This page dictates its own dimensions to the parent App!
  export const config = {
    minimal: { w: 320, h: 35, r: 35 },
    expanded: { w: 580, h: 300, r: 35 }
  }
</script>

<script lang="ts">
  import Clock from '../components/Clock.svelte'
  import Divider from '../components/Divider.svelte'
  import Calendar from '../components/Calendar.svelte'
  import Weather from '../components/Weather.svelte'
  import BentoGrid from '../components/BentoGrid.svelte'
  import BentoItem from '../components/BentoItem.svelte'

  export let expanded: boolean = false
</script>

<div class="page-container">
  {#if !expanded}
    <div class="minimal-view">
      <!--<DotIndicator />-->
      <Calendar expanded={false} />
      <Divider />
      <Clock expanded={false} />
      <Divider />
      <Weather expanded={false} />
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
    height: 100%;
    gap: 1rem;
    min-width: max-content; /* Prevents squishing */
    white-space: nowrap; /* Prevents text breaking */
  }

  .expanded-view {
    padding: 1rem;
    width: 100%;
    height: 100%;
    box-sizing: border-box;
  }
</style>
