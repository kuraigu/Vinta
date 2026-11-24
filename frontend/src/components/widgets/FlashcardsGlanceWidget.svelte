<!--
  @component FlashcardsGlanceWidget
  @description Compact Home Bento card displaying due flashcard count and 1-click study launch.
-->
<script lang="ts">
  import { getContext, onMount } from 'svelte'
  import { GraduationCap, ChevronRight } from 'lucide-svelte'
  import WidgetFrame from './WidgetFrame.svelte'
  import BentoCard from '../BentoCard.svelte'
  import IconBadge from '../IconBadge.svelte'
  import FlashcardsPage, {
    config as flashcardsConfig
  } from '../../pages/FlashcardsPage.svelte'
  import {
    dueFlashcardsCount,
    flashcardsStore,
    initFlashcardsStore
  } from '../../stores/flashcardsStore'
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
    void initFlashcardsStore()
  })

  const canOpen = $derived($pageVisibilityStore.flashcards !== false)
  const activeDeck = $derived(
    $flashcardsStore.decks.find(
      (d) => d.id === $flashcardsStore.activeDeckId
    ) ??
      $flashcardsStore.decks[0] ??
      null
  )
</script>

<WidgetFrame
  id="flashcards"
  {size}
  {showMenuButton}
  {onreorder}
  {oncontextmenu}
>
  <BentoCard
    variant="bento"
    class="fc-glance-card"
    disabled={!canOpen}
    onclick={() => {
      if (canOpen) navigate(FlashcardsPage, flashcardsConfig, true)
    }}
  >
    <div class="glance-header">
      <div class="glance-title-group">
        <IconBadge icon={GraduationCap} color="surface" size="sm" />
        <span class="glance-title-text">{$tStore('flashcards.title')}</span>
      </div>
      {#if canOpen}
        <ChevronRight size={14} class="glance-nav-chevron" />
      {/if}
    </div>
    <div class="fc-glance-body">
      <span class="fc-glance-count">
        {$dueFlashcardsCount > 0
          ? $tStore('flashcards.due_count', { count: $dueFlashcardsCount })
          : $tStore('flashcards.minimal_done')}
      </span>
      {#if activeDeck}
        <span class="fc-glance-deck" title={activeDeck.name}
          >{activeDeck.name}</span
        >
      {/if}
    </div>
  </BentoCard>
</WidgetFrame>

<style src="./styles/FlashcardsGlanceWidget.css"></style>
