<script lang="ts">
  let {
    columns = '1fr',
    rows = 'auto',
    gap = '1rem',
    height = '100%',
    children
  } = $props<{
    columns?: string
    rows?: string
    gap?: string
    height?: string
    children?: import('svelte').Snippet
  }>()
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
  {@render children?.()}
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
