<script lang="ts">
  import Clock from '../components/Clock.svelte'
  import HorizontalBar from '../components/HorizontalBar.svelte'
  import Calendar from '../components/Calendar.svelte'
  import SystemStats from '../components/SystemStats.svelte'
  import BentoGrid from '../components/BentoGrid.svelte'
  import BentoItem from '../components/BentoItem.svelte'
  import DotIndicator from '../components/DotIndicator.svelte'

  export let expanded: boolean = false

  // This page dictates its own dimensions to the parent App!
  export const config = {
    minimal: { w: 250, h: 35, r: 35 },
    expanded: { w: 580, h: 280, r: 35 }
  }
</script>

<div class="page-container {expanded ? 'expanded' : 'minimal'}">
  {#if !expanded}
    <DotIndicator />
    <Calendar {expanded} />
    <HorizontalBar />
    <Clock {expanded} />
  {:else}
    <BentoGrid columns="200px 1fr" gap="1rem">
      <BentoItem>
        <Calendar {expanded} />
      </BentoItem>
      <BentoGrid columns="1fr" rows="1fr 1fr" gap="1rem">
        <BentoItem>
          <Clock {expanded} />
        </BentoItem>
        <BentoItem>
          <SystemStats {expanded} />
        </BentoItem>
      </BentoGrid>
    </BentoGrid>
  {/if}
</div>

<style>
  .page-container {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    box-sizing: border-box;
  }

  .page-container.expanded {
    padding: 1rem;
  }
</style>
