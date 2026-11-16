<script module lang="ts">
  export const config = {
    minimal: { w: 368, h: 44, r: 16 },
    expanded: { w: 400, h: 225, r: 28 },
    borderColor: 'var(--vinta-island-border-media)',
    glowColor: 'var(--vinta-island-glow-media)'
  }
</script>

<script lang="ts">
  /**
   * @fileoverview Page container hosting the full-page Media Player and Soundscapes view.
   *
   * Sets up island sizing and mounts the comprehensive MediaWidget component.
   */

  import MediaWidget from '../components/MediaWidget.svelte'

  let { expanded = false }: { expanded?: boolean } = $props()
</script>

<div
  data-vinta="media-page"
  class="media-page-layout"
  class:expanded
  class:minimal={!expanded}
>
  <MediaWidget {expanded} />
</div>

<style src="./styles/MediaPage.css"></style>
