<script module lang="ts">
  export const config = {
    minimal: { w: 368, h: 44, r: 16 },
    expanded: { w: 400, h: 225, r: 28 },
    borderColor: 'rgba(255, 60, 100, 0.4)',
    glowColor: 'rgba(255, 60, 100, 0.35)'
  }
</script>

<script lang="ts">
  import MediaWidget from '../components/MediaWidget.svelte'

  let { expanded = false }: { expanded?: boolean } = $props()
</script>

<div class="media-page-layout" class:expanded class:minimal={!expanded}>
  <MediaWidget {expanded} />
</div>

<style src="./styles/MediaPage.css"></style>
