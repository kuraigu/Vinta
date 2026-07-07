<script lang="ts">
  export let columns: string = '1fr'
  export let rows: string = 'auto'
  export let gap: string = '1rem'
  export let height: string = '100%'
</script>

<div
  class="bento-grid"
  style="
    --bento-columns: {columns};
    --bento-rows: {rows};
    --bento-gap: {gap};
    --bento-height: {height};
  "
>
  <slot />
</div>

<style>
  .bento-grid {
    display: grid;
    grid-template-columns: var(--bento-columns);
    grid-template-rows: var(--bento-rows);
    gap: var(--bento-gap);
    width: 100%;
    height: var(--bento-height);
    box-sizing: border-box;
  }
</style>
