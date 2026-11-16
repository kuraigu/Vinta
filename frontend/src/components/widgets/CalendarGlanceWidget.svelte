<!--
  @component CalendarGlanceWidget
  @description Modular Home calendar widget with `data-vinta-widget="calendar"` scoping.
-->
<script lang="ts">
  import WidgetFrame from './WidgetFrame.svelte'
  import CalendarGlance from '../CalendarGlance.svelte'
  import type { WidgetSizePreset } from '../../stores/settings'

  interface Props {
    size?: WidgetSizePreset
    showMenuButton?: boolean
    onOpenCalendar?: () => void
    oncontextmenu?: (e: MouseEvent) => void
  }

  let {
    size = '2x1',
    showMenuButton = false,
    onOpenCalendar = undefined,
    oncontextmenu = undefined
  }: Props = $props()
</script>

<WidgetFrame id="calendar" {size} {showMenuButton} {oncontextmenu}>
  <CalendarGlance {onOpenCalendar} />
</WidgetFrame>
