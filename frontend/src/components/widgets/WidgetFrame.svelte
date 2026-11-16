<!--
  @component WidgetFrame
  @description Shared Home widget container providing Bento grid size presets (`1x1`, `2x1`, `2x2`),
  `data-vinta-widget` theme hooks, CSS layout/style containment, and optional hover menu trigger.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import { MoreHorizontal } from 'lucide-svelte'
  import { islandTooltip } from '../../services/islandTooltip'
  import { tStore, type WidgetSizePreset } from '../../store'

  interface Props {
    id: string
    size?: WidgetSizePreset
    showMenuButton?: boolean
    oncontextmenu?: (e: MouseEvent) => void
    children: Snippet
  }

  let {
    id,
    size = '2x1',
    showMenuButton = false,
    oncontextmenu = undefined,
    children
  }: Props = $props()
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="vinta-widget-frame span-{size}"
  data-vinta-widget={id}
  data-widget-size={size}
  {oncontextmenu}
>
  {@render children()}

  {#if showMenuButton && oncontextmenu}
    <button
      type="button"
      class="widget-menu-trigger"
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
      <MoreHorizontal size={13} />
    </button>
  {/if}
</div>

<style src="../styles/WidgetFrame.css"></style>
