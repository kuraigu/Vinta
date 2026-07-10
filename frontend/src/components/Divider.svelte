<script lang="ts">
  // Export properties so the bar is completely reusable!
  export let width: string = '2px'
  export let height: string = '60%'
  export let color: string = 'rgba(255, 255, 255, 0.05)'
  export let radius: string = '999px'
</script>

<div
  class="divider"
  style="
    width: {width}; 
    height: {height}; 
    background-color: {color}; 
    border-radius: {radius};
  "
/>

<style>
  .divider {
    /* flex-shrink: 0 prevents the bar from getting squished if the page gets crowded */
    flex-shrink: 0;
  }
</style>
