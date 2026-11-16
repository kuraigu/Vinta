<!--
  @component AnnouncementWidget
  @description Modular Home developer announcements widget with `data-vinta-widget="announcements"` scoping.
-->
<script lang="ts">
  import WidgetFrame from './WidgetFrame.svelte'
  import AnnouncementGlance from '../AnnouncementGlance.svelte'
  import type { WidgetSizePreset } from '../../stores/settings'

  interface Props {
    size?: WidgetSizePreset
    showMenuButton?: boolean
    onClick?: () => void
    oncontextmenu?: (e: MouseEvent) => void
  }

  let {
    size = '2x1',
    showMenuButton = false,
    onClick = undefined,
    oncontextmenu = undefined
  }: Props = $props()
</script>

<WidgetFrame id="announcements" {size} {showMenuButton} {oncontextmenu}>
  <AnnouncementGlance {onClick} />
</WidgetFrame>
