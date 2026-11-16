<!--
  @component NotificationWidget
  @description Modular Home notifications glance widget with `data-vinta-widget="notifications"` scoping.
-->
<script lang="ts">
  import WidgetFrame from './WidgetFrame.svelte'
  import NotificationGlance from '../NotificationGlance.svelte'
  import type { WidgetSizePreset } from '../../stores/settings'

  interface Props {
    size?: WidgetSizePreset
    showMenuButton?: boolean
    oncontextmenu?: (e: MouseEvent) => void
  }

  let {
    size = '2x1',
    showMenuButton = false,
    oncontextmenu = undefined
  }: Props = $props()
</script>

<WidgetFrame id="notifications" {size} {showMenuButton} {oncontextmenu}>
  <NotificationGlance />
</WidgetFrame>
