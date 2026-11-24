<!--
  @component HabitsGlanceWidget
  @description Compact Home Bento widget displaying daily habits progress and 1-click check-in.
-->
<script lang="ts">
  import { getContext, onMount } from 'svelte'
  import { ListChecks, ChevronRight, CheckCircle2, Circle } from 'lucide-svelte'
  import WidgetFrame from './WidgetFrame.svelte'
  import BentoCard from '../BentoCard.svelte'
  import IconBadge from '../IconBadge.svelte'
  import HabitsPage, {
    config as habitsConfig
  } from '../../pages/HabitsPage.svelte'
  import {
    habitsStore,
    habitsProgress,
    getTodayDateKey,
    initHabitsStore,
    toggleHabitToday
  } from '../../stores/habitsStore'
  import {
    pageVisibilityStore,
    tStore,
    type IslandConfig,
    type WidgetSizePreset
  } from '../../store'

  interface Props {
    size?: WidgetSizePreset
    showMenuButton?: boolean
    onreorder?: (sourceId: string, targetId: string) => void
    oncontextmenu?: (e: MouseEvent) => void
  }

  let {
    size = '1x1',
    showMenuButton = false,
    onreorder = undefined,
    oncontextmenu = undefined
  }: Props = $props()

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  onMount(() => {
    void initHabitsStore()
  })

  const canOpen = $derived($pageVisibilityStore.habits !== false)
  const today = $derived(getTodayDateKey())
  const nextHabit = $derived(
    $habitsStore.habits.find((h) => h.lastCompletedDate !== today) ??
      $habitsStore.habits[0] ??
      null
  )
</script>

<WidgetFrame id="habits" {size} {showMenuButton} {onreorder} {oncontextmenu}>
  <BentoCard
    variant="bento"
    class="hb-glance-card"
    disabled={!canOpen}
    onclick={() => {
      if (canOpen) navigate(HabitsPage, habitsConfig, true)
    }}
  >
    <div class="glance-header">
      <div class="glance-title-group">
        <IconBadge icon={ListChecks} color="surface" size="sm" />
        <span class="glance-title-text">{$tStore('habits.title')}</span>
      </div>
      {#if canOpen}
        <ChevronRight size={14} class="glance-nav-chevron" />
      {/if}
    </div>

    {#if nextHabit}
      {@const isDone = nextHabit.lastCompletedDate === today}
      <button
        type="button"
        class="hb-glance-quick"
        onclick={(e) => {
          e.stopPropagation()
          toggleHabitToday(nextHabit.id)
        }}
      >
        {#if isDone}
          <CheckCircle2 size={14} class="hb-glance-check done" />
        {:else}
          <Circle size={14} class="hb-glance-check" />
        {/if}
        <span class="hb-glance-habit" title={nextHabit.title}
          >{nextHabit.title}</span
        >
        <span class="hb-glance-count"
          >{$habitsProgress.done}/{$habitsProgress.total}</span
        >
      </button>
    {/if}
  </BentoCard>
</WidgetFrame>

<style src="./styles/HabitsGlanceWidget.css"></style>
