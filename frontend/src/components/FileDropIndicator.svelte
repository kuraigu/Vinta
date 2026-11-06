<!--
  @component FileDropIndicator
  @description Animated dropzone overlay that appears over the island when dragging files from Explorer.
-->
<script lang="ts">
  import { FolderDown } from 'lucide-svelte'

  /**
   * Props for the hold-to-trigger file drag indicator overlay.
   */
  interface Props {
    active: boolean
    progress: number
    isCompact?: boolean
  }

  let { active, progress, isCompact = false }: Props = $props()
</script>

{#if active}
  <div
    class="file-drop-indicator"
    class:is-compact={isCompact}
    data-vinta="file-drop-indicator"
  >
    <div class="indicator-icon-badge">
      <FolderDown size={isCompact ? 13 : 18} />
    </div>
    <div class="indicator-text-group">
      <span class="indicator-title">
        {isCompact ? 'Hold for File Shelf' : 'Hold to open File Shelf'}
      </span>
      {#if !isCompact}
        <span class="indicator-subtitle">
          Drop now to stage files immediately
        </span>
      {/if}
    </div>
    <div class="indicator-progress-track">
      <div
        class="indicator-progress-fill"
        style="width: {Math.min(100, Math.max(0, progress))}%;"
      ></div>
    </div>
  </div>
{/if}

<style src="./styles/FileDropIndicator.css"></style>
