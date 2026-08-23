<script module lang="ts">
  export const config = {
    minimal: { w: 368, h: 44, r: 22 },
    expanded: { w: 440, h: 270, r: 28 },
    borderColor: 'rgba(255, 60, 100, 0.4)',
    glowColor: 'rgba(255, 60, 100, 0.35)'
  }
</script>

<script lang="ts">
  import MediaWidget from '../components/MediaWidget.svelte'

  let { expanded = false }: { expanded?: boolean } = $props()
</script>

<div class="media-page-layout {expanded ? 'expanded' : 'minimal'}">
  {#if !expanded}
    <MediaWidget expanded={false} />
  {:else}
    <div class="expanded-content">
      <MediaWidget expanded={true} />
    </div>
  {/if}
</div>

<style src="./styles/MediaPage.css"></style>
