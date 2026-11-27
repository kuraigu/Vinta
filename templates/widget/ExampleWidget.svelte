<!--
  @component ExampleWidget
  @description Starter Home Bento widget wrapped in WidgetFrame with size presets (1x1, 2x1, 2x2).
-->
<script lang="ts">
  import { getContext } from 'svelte'
  import { LayoutGrid, ChevronRight } from 'lucide-svelte'
  import WidgetFrame from './WidgetFrame.svelte'
  import BentoCard from '../BentoCard.svelte'
  import IconBadge from '../IconBadge.svelte'
  import type { IslandConfig, WidgetSizePreset } from '../../store'

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

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  function handleOpen() {
    // Call navigate(TargetPage, targetConfig, true) when wired to a workspace page.
    void navigate
  }
</script>

<WidgetFrame id="example" {size} {showMenuButton} {onreorder} {oncontextmenu}>
  <BentoCard variant="bento" class="example-widget-card" onclick={handleOpen}>
    <div class="glance-header">
      <div class="glance-title-group">
        <IconBadge icon={LayoutGrid} color="surface" size="sm" />
        <span class="glance-title-text">Example Widget</span>
      </div>
      <ChevronRight size={14} class="glance-nav-chevron" />
    </div>

    <div class="example-widget-body">
      <span class="example-widget-value">Ready</span>
      <span class="example-widget-sub">Compact {size} widget</span>
    </div>
  </BentoCard>
</WidgetFrame>

<style>
  .example-widget-body {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-inline-size: 0;
  }

  .example-widget-value {
    font-size: 13px;
    font-weight: 700;
    color: var(--vinta-text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .example-widget-sub {
    font-size: 11px;
    color: var(--vinta-text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
</style>
