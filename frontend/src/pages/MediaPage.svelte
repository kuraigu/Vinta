<script module lang="ts">
  export const config = {
    minimal: { w: 368, h: 44, r: 16 },
    expanded: { w: 400, h: 225, r: 28 },
    borderColor: 'var(--vinta-island-border-media, #ff3366)',
    glowColor: 'var(--vinta-island-glow-media, #ff3366)'
  }
</script>

<script lang="ts">
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
