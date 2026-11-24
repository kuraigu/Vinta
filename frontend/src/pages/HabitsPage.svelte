<!--
  @component HabitsPage
  @description Daily habits and break/hydration micro-routines page with streak tracking.
-->
<script module lang="ts">
  import { ISLAND_WIDTH_TIERS, type IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 160, h: 36, r: 18 },
    expanded: { w: ISLAND_WIDTH_TIERS.standard, h: 420, r: 28 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { onMount } from 'svelte'
  import {
    CheckCircle2,
    Circle,
    Flame,
    Plus,
    Trash2,
    Bell,
    ListChecks
  } from 'lucide-svelte'
  import IslandPage from '../components/IslandPage.svelte'
  import PageHeader from '../components/PageHeader.svelte'
  import BentoCard from '../components/BentoCard.svelte'
  import IconBadge from '../components/IconBadge.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import CustomSelect from '../components/CustomSelect.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import {
    habitsStore,
    habitsProgress,
    getTodayDateKey,
    initHabitsStore,
    addHabit,
    deleteHabit,
    toggleHabitToday
  } from '../stores/habitsStore'
  import { tStore } from '../store'

  interface Props {
    expanded?: boolean
  }

  let { expanded = false }: Props = $props()

  let showAddForm = $state(false)
  let newTitle = $state('')
  let newReminder = $state('0')

  const reminderOptions = $derived([
    { value: '0', label: $tStore('habits.reminder_off') },
    { value: '30', label: $tStore('habits.reminder_30m') },
    { value: '60', label: $tStore('habits.reminder_60m') },
    { value: '120', label: $tStore('habits.reminder_120m') }
  ])

  onMount(() => {
    void initHabitsStore()
  })

  const today = $derived(getTodayDateKey())

  function handleAddHabit() {
    if (!newTitle.trim()) return
    addHabit(newTitle, Number(newReminder) || 0)
    newTitle = ''
    newReminder = '0'
    showAddForm = false
  }
</script>

<IslandPage {expanded}>
  {#snippet minimal()}
    <div class="habits-minimal">
      <ListChecks size={14} />
      <span>
        {$tStore('habits.minimal_progress', {
          done: $habitsProgress.done,
          total: $habitsProgress.total
        })}
      </span>
    </div>
  {/snippet}

  {#snippet expandedContent()}
    <div class="habits-page" data-vinta="habits-page">
      <PageHeader title={$tStore('habits.title')}>
        {#snippet actions()}
          <span class="hb-progress-pill">
            {$habitsProgress.done}/{$habitsProgress.total}
          </span>
          <button
            type="button"
            class="hb-header-pill"
            onclick={() => (showAddForm = !showAddForm)}
          >
            <Plus size={12} />
            <span>{$tStore('habits.add_habit')}</span>
          </button>
        {/snippet}
      </PageHeader>

      <ScrollView class="hb-scroll-area">
        {#if showAddForm}
          <BentoCard variant="bento" class="hb-form-card">
            <input
              type="text"
              class="hb-input"
              placeholder={$tStore('habits.title_placeholder')}
              bind:value={newTitle}
              onkeydown={(e) => {
                if (e.key === 'Enter') handleAddHabit()
              }}
            />
            <div class="hb-form-row">
              <CustomSelect
                options={reminderOptions}
                bind:value={newReminder}
              />
              <div class="hb-form-actions">
                <button
                  type="button"
                  class="hb-btn secondary"
                  onclick={() => (showAddForm = false)}
                >
                  {$tStore('common.cancel')}
                </button>
                <button
                  type="button"
                  class="hb-btn primary"
                  disabled={!newTitle.trim()}
                  onclick={handleAddHabit}
                >
                  {$tStore('common.save')}
                </button>
              </div>
            </div>
          </BentoCard>
        {/if}

        {#if $habitsStore.habits.length === 0}
          <BentoCard variant="bento" class="hb-empty-card">
            <ListChecks size={20} />
            <span>{$tStore('habits.empty')}</span>
          </BentoCard>
        {:else}
          <div class="hb-list">
            {#each $habitsStore.habits as habit (habit.id)}
              {@const isDone = habit.lastCompletedDate === today}
              <BentoCard
                variant="bento"
                class="hb-bento-row {isDone ? 'is-done' : ''}"
                onclick={() => toggleHabitToday(habit.id)}
              >
                <div class="hb-row-left">
                  <IconBadge
                    icon={isDone ? CheckCircle2 : Circle}
                    color={isDone ? 'emerald' : 'slate'}
                    size="lg"
                  />
                  <div class="hb-row-info">
                    <span class="hb-row-title" title={habit.title}>
                      {habit.title}
                    </span>
                    <span class="hb-row-sub">
                      {#if habit.streak > 0}
                        <span class="hb-inline-meta streak">
                          <Flame size={10} />
                          {habit.streak}d
                        </span>
                      {/if}
                      {#if habit.reminderMinutes > 0}
                        <span class="hb-inline-meta">
                          <Bell size={10} />
                          {$tStore('habits.reminder_every', {
                            mins: habit.reminderMinutes
                          })}
                        </span>
                      {:else if habit.streak === 0}
                        {$tStore('habits.desc')}
                      {/if}
                    </span>
                  </div>
                </div>

                <button
                  type="button"
                  class="hb-delete-btn"
                  use:islandTooltip={{
                    text: $tStore('common.delete'),
                    position: 'left'
                  }}
                  aria-label={$tStore('common.delete')}
                  onclick={(e) => {
                    e.stopPropagation()
                    deleteHabit(habit.id)
                  }}
                >
                  <Trash2 size={13} />
                </button>
              </BentoCard>
            {/each}
          </div>
        {/if}
      </ScrollView>
    </div>
  {/snippet}
</IslandPage>

<style src="./styles/HabitsPage.css"></style>
