<script module lang="ts">
  export const config = {
    minimal: { w: 340, h: 44, r: 22 },
    expanded: { w: 440, h: 270, r: 28 },
    borderColor: 'rgba(255, 60, 100, 0.4)',
    glowColor: 'rgba(255, 60, 100, 0.35)'
  }
</script>

<script lang="ts">
  import MediaWidget from '../components/MediaWidget.svelte'
  import Clock from '../components/Clock.svelte'

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

<style>
  .media-page-layout {
    width: 100%;
    height: 100%;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    position: relative;
    overflow: hidden;
  }

  .media-page-layout.expanded {
    padding-bottom: 54px; /* Integrated space above Dock */
  }

  .expanded-content {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
    width: 100%;
    height: 100%;
    min-height: 0;
    padding: 0;
    box-sizing: border-box;
  }
</style>
