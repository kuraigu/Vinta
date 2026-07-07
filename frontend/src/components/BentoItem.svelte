<script lang="ts">
  // Wrapper to ensure children in a grid cell stretch properly.
</script>

<div class="bento-item">
  <slot />
</div>

<style>
  .bento-item {
    display: flex;
    align-items: stretch;
    justify-content: stretch;
    width: 100%;
    height: 100%;
    box-sizing: border-box;
  }

  /* Svelte scopes CSS, but we can target the direct child to ensure it fills the item */
  .bento-item > :global(*) {
    flex: 1;
    min-width: 0;
    min-height: 0;
  }
</style>
