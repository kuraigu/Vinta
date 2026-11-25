<!--
  @component WidgetFrame
  @description Shared Home widget container providing Bento grid size presets (`1x1`, `2x1`, `2x2`),
  `data-vinta-widget` theme hooks, CSS layout/style containment, and optional hover menu trigger.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import { MoreHorizontal, GripVertical } from 'lucide-svelte'
  import { islandTooltip } from '../../services/islandTooltip'
  import { tStore, type WidgetSizePreset } from '../../store'

  interface Props {
    id: string
    size?: WidgetSizePreset
    showMenuButton?: boolean
    reorderable?: boolean
    onreorder?: (sourceId: string, targetId: string) => void
    oncontextmenu?: (e: MouseEvent) => void
    children: Snippet
  }

  let {
    id,
    size = '2x1',
    showMenuButton = false,
    reorderable = true,
    onreorder = undefined,
    oncontextmenu = undefined,
    children
  }: Props = $props()

  let frameEl = $state<HTMLDivElement | null>(null)
  let isDragging = $state(false)
  let isDropTarget = $state(false)

  function handleDragStart(e: DragEvent) {
    if (!reorderable || !onreorder) return
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move'
      e.dataTransfer.setData('text/vinta-widget', id)
      if (frameEl) {
        e.dataTransfer.setDragImage(
          frameEl,
          Math.round(frameEl.clientWidth / 2),
          20
        )
      }
    }
    isDragging = true
  }

  function handleDragOver(e: DragEvent) {
    if (!reorderable || !onreorder) return
    if (e.dataTransfer?.types?.includes('text/vinta-widget')) {
      e.preventDefault()
      if (e.dataTransfer) {
        e.dataTransfer.dropEffect = 'move'
      }
      isDropTarget = true
    }
  }

  function handleDragLeave() {
    isDropTarget = false
  }

  function handleDrop(e: DragEvent) {
    isDropTarget = false
    if (!reorderable || !onreorder) return
    const sourceId = e.dataTransfer?.getData('text/vinta-widget')
    if (sourceId && sourceId !== id) {
      e.preventDefault()
      onreorder(sourceId, id)
    }
  }

  function handleDragEnd() {
    isDragging = false
    isDropTarget = false
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  bind:this={frameEl}
  class="vinta-widget-frame span-{size}"
  class:has-drag-handle={reorderable && Boolean(onreorder)}
  class:is-dragging={isDragging}
  class:is-drop-target={isDropTarget}
  data-vinta-widget={id}
  data-widget-size={size}
  ondragover={handleDragOver}
  ondragleave={handleDragLeave}
  ondrop={handleDrop}
  {oncontextmenu}
>
  {@render children()}

  {#if reorderable && onreorder}
    <span
      class="widget-drag-trigger no-swipe no-drag"
      draggable="true"
      aria-label={$tStore('home.widget_move_up')}
      ondragstart={handleDragStart}
      ondragend={handleDragEnd}
    >
      <GripVertical size={14} strokeWidth={2.2} />
    </span>
  {/if}

  {#if showMenuButton && oncontextmenu}
    <button
      type="button"
      class="widget-menu-trigger no-swipe no-drag"
      use:islandTooltip={{
        text: $tStore('home.widget_options'),
        position: 'top'
      }}
      aria-label={$tStore('home.widget_options')}
      onmousedown={(e) => e.stopPropagation()}
      onclick={(e) => {
        e.stopPropagation()
        oncontextmenu(e)
      }}
    >
      <MoreHorizontal size={14} />
    </button>
  {/if}
</div>

<style src="../styles/WidgetFrame.css"></style>
