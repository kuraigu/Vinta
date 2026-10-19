<script lang="ts">
  /**
   * Component properties for ExampleComponent.
   */
  interface Props {
    expanded?: boolean
  }

  let { expanded = false }: Props = $props()
</script>

<div data-vinta="example-component" class="example-container">
  {#if !expanded}
    <div class="minimal-view">
      <span>Minimal Widget</span>
    </div>
  {:else}
    <div class="expanded-view">
      <h3 class="title">Expanded Feature</h3>
      <p class="description">Starter component using standard Vinta tokens.</p>
    </div>
  {/if}
</div>

<style>
  .example-container {
    color: var(--vinta-text-primary, #ffffff);
  }

  .minimal-view {
    font-size: 12px;
    font-weight: 600;
  }

  .expanded-view {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px;
    border-radius: var(--vinta-radius-card, 16px);
    background: var(--vinta-surface-card, rgba(255, 255, 255, 0.05));
    border: 1px solid var(--vinta-border, rgba(255, 255, 255, 0.1));
  }

  .title {
    margin: 0;
    font-size: 13px;
    font-weight: 700;
  }

  .description {
    margin: 0;
    font-size: 11px;
    color: var(--vinta-text-secondary, rgba(255, 255, 255, 0.7));
  }
</style>
