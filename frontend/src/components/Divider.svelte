<script lang="ts">
  // Export properties so the bar is completely reusable!
  let {
    width = '2px',
    height = '60%',
    color = 'var(--c-item-bg)',
    radius = '999px'
  } = $props<{
    width?: string
    height?: string
    color?: string
    radius?: string
  }>()
</script>

<div
  class="divider"
  style="
    width: {width}; 
    height: {height}; 
    background-color: {color}; 
    border-radius: {radius};
  "
></div>

<style>
  .divider {
    /* flex-shrink: 0 prevents the bar from getting squished if the page gets crowded */
    flex-shrink: 0;
  }
</style>
