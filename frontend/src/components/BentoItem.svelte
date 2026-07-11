<script lang="ts">
  let { children } = $props<{ children?: import('svelte').Snippet }>()
</script>

<div class="bento-item">
  {@render children?.()}
</div>

<style>
  .bento-item {
    display: flex;
    align-items: stretch;
    justify-content: stretch;
    width: 100%;
    height: 100%;
    box-sizing: border-box;

    background: var(--c-item-bg);
    border-radius: 16px;
    border: 1px solid var(--c-border);
    box-sizing: border-box;
    box-shadow:
      inset 0 1px 0 var(--c-border),
      0 2px 8px rgba(0, 0, 0, 0.1);
    position: relative;
    overflow: hidden;
  }

  /* Svelte scopes CSS, but we can target the direct child to ensure it fills the item */
  .bento-item > :global(*) {
    flex: 1;
    min-width: 0;
    min-height: 0;
  }
</style>
