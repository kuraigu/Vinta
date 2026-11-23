<!--
  @component ActionCenterWidget
  @description Modular Home quick controls widget with `data-vinta-widget="actionCenter"` scoping.
-->
<script lang="ts">
  import WidgetFrame from './WidgetFrame.svelte'
  import ActionCenterGlance from '../ActionCenterGlance.svelte'
  import type { WidgetSizePreset } from '../../stores/settings'

  interface Props {
    size?: WidgetSizePreset
    showMenuButton?: boolean
    onreorder?: (sourceId: string, targetId: string) => void
    oncontextmenu?: (e: MouseEvent) => void
  }

  let {
    size = '2x1',
    showMenuButton = false,
    onreorder = undefined,
    oncontextmenu = undefined
  }: Props = $props()
</script>

<WidgetFrame
  id="actionCenter"
  {size}
  {showMenuButton}
  {onreorder}
  {oncontextmenu}
>
  <ActionCenterGlance />
</WidgetFrame>
