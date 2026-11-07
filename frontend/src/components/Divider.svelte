<!--
  @component Divider
  @description Subtle separator line supporting horizontal or vertical orientation.
-->
<script lang="ts">
  interface Props {
    width?: string
    height?: string
    color?: string
    radius?: string
  }

  // Export properties so the bar is completely reusable!
  let {
    width = '2px',
    height = '60%',
    color = 'var(--vinta-item-bg)',
    radius = '999px'
  }: Props = $props()
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
