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
    <div class="header">
      <div class="header-side left">
        <div class="title">Now Playing</div>
      </div>
      <div class="header-side right">
        <div class="clock-wrapper">
          <Clock expanded={false} />
        </div>
      </div>
    </div>
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

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px 0 16px;
    margin-bottom: 6px;
    z-index: 2;
    pointer-events: none;
    width: 100%;
    box-sizing: border-box;
    flex-shrink: 0;
  }

  .header-side {
    flex: 1 1 0;
    display: flex;
    align-items: center;
    min-width: 0;
  }

  .header-side.right {
    justify-content: flex-end;
  }

  .title {
    font-size: 15px;
    font-weight: 700;
    letter-spacing: -0.01em;
    color: var(--c-text-primary);
    text-align: left;
  }

  .clock-wrapper {
    font-size: 12px;
    font-weight: 500;
    opacity: 0.65;
    font-feature-settings: 'tnum';
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
    padding: 0 0 4px 0;
    box-sizing: border-box;
  }
</style>
