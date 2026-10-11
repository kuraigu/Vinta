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

    background: linear-gradient(
        145deg,
        rgba(255, 255, 255, 0.065) 0%,
        rgba(255, 255, 255, 0.02) 100%
      ),
      #15161c;
    border-radius: 20px;
    border: 1px solid rgba(255, 255, 255, 0.09);
    box-sizing: border-box;
    box-shadow:
      inset 0 1px 0 0 rgba(255, 255, 255, 0.12),
      0 4px 18px rgba(0, 0, 0, 0.18);
    position: relative;
    overflow: hidden;
    transition:
      transform 0.2s ease,
      border-color 0.2s ease,
      box-shadow 0.2s ease;
  }

  .bento-item:hover {
    transform: translateY(-1px);
    border-color: rgba(255, 255, 255, 0.16);
    box-shadow:
      inset 0 1px 0 0 rgba(255, 255, 255, 0.16),
      0 6px 22px rgba(0, 0, 0, 0.24);
  }

  /* Svelte scopes CSS, but we can target the direct child to ensure it fills the item */
  .bento-item > :global(*) {
    flex: 1;
    min-width: 0;
    min-height: 0;
  }
</style>
