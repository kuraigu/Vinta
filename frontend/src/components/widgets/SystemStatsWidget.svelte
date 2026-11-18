<!--
  @component SystemStatsWidget
  @description Modular Home system telemetry widget with `data-vinta-widget="systemStats"` scoping.
-->
<script lang="ts">
  import WidgetFrame from './WidgetFrame.svelte'
  import SystemStats from '../SystemStats.svelte'
  import type { WidgetSizePreset } from '../../stores/settings'

  interface Props {
    size?: WidgetSizePreset
    active?: boolean
    showMenuButton?: boolean
    oncontextmenu?: (e: MouseEvent) => void
  }

  let {
    size = '2x1',
    active = true,
    showMenuButton = false,
    oncontextmenu = undefined
  }: Props = $props()
</script>

<WidgetFrame id="systemStats" {size} {showMenuButton} {oncontextmenu}>
  <SystemStats expanded={active} />
</WidgetFrame>
