<script lang="ts">
  import BentoCard from './BentoCard.svelte'

  interface Props {
    expanded?: boolean
  }

  let { expanded = false }: Props = $props()
</script>

<div class="example-container">
  {#if !expanded}
    <div class="minimal-view">
      <span>Minimal Widget</span>
    </div>
  {:else}
    <BentoCard variant="bento">
      <h3 class="title">Expanded Feature</h3>
      <p class="description">Starter component built with standard Vinta primitives.</p>
    </BentoCard>
  {/if}
</div>

<style>
  .example-container {
    color: var(--vinta-text-primary);
  }

  .minimal-view {
    font-size: 12px;
    font-weight: 600;
  }

  .title {
    margin: 0 0 4px 0;
    font-size: 13px;
    font-weight: 700;
    color: var(--vinta-text-primary);
  }

  .description {
    margin: 0;
    font-size: 11px;
    color: var(--vinta-text-secondary);
  }
</style>
