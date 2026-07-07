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

    background: rgba(255, 255, 255, 0.05);
    border-radius: 24px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.1),
      0 10px 30px rgba(0, 0, 0, 0.2);
    overflow: hidden;
  }

  /* Svelte scopes CSS, but we can target the direct child to ensure it fills the item */
  .bento-item > :global(*) {
    flex: 1;
    min-width: 0;
    min-height: 0;
  }
</style>
