<!--
  @component FlashcardsPage
  @description Spaced-repetition micro-study flashcards page with single-row PageHeader and BentoCard study view.
-->
<script module lang="ts">
  import { ISLAND_WIDTH_TIERS, type IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 170, h: 36, r: 18 },
    expanded: { w: ISLAND_WIDTH_TIERS.standard, h: 410, r: 28 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { onMount } from 'svelte'
  import {
    GraduationCap,
    Plus,
    Trash2,
    RotateCw,
    CheckCircle2,
    BookOpen,
    MoreVertical,
    FolderPlus,
    X
  } from 'lucide-svelte'
  import IslandPage from '../components/IslandPage.svelte'
  import PageHeader from '../components/PageHeader.svelte'
  import BentoCard from '../components/BentoCard.svelte'
  import IconBadge from '../components/IconBadge.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import CustomSelect from '../components/CustomSelect.svelte'
  import CustomContextMenu, {
    type ContextMenuItem
  } from '../components/CustomContextMenu.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import {
    flashcardsStore,
    dueFlashcardsCount,
    initFlashcardsStore,
    setActiveFlashcardDeck,
    createFlashcardDeck,
    deleteFlashcardDeck,
    addFlashcard,
    deleteFlashcard,
    reviewFlashcard,
    restoreStarterDeck,
    getNextIntervalLabel,
    type ReviewRating
  } from '../stores/flashcardsStore'
  import { tStore } from '../store'

  interface Props {
    expanded?: boolean
  }

  let { expanded = false }: Props = $props()

  let flipped = $state(false)
  let practiceAll = $state(false)
  let showAddCard = $state(false)
  let showNewDeck = $state(false)
  let newDeckName = $state('')
  let newFront = $state('')
  let newBack = $state('')
  let menuOpen = $state(false)
  let menuCoords = $state({ x: 0, y: 0 })

  onMount(() => {
    void initFlashcardsStore()
  })

  const activeDeck = $derived(
    $flashcardsStore.decks.find(
      (d) => d.id === $flashcardsStore.activeDeckId
    ) ??
      $flashcardsStore.decks[0] ??
      null
  )

  const deckOptions = $derived(
    $flashcardsStore.decks.map((d) => ({
      value: d.id,
      label: `${d.name} (${d.cards.length})`
    }))
  )

  const dueCards = $derived.by(() => {
    if (!activeDeck) return []
    const now = Date.now()
    return activeDeck.cards.filter((c) => c.dueAt <= now)
  })

  const studyQueue = $derived(
    practiceAll ? (activeDeck?.cards ?? []) : dueCards
  )

  const currentCard = $derived(studyQueue[0] ?? null)

  function handleRate(rating: ReviewRating) {
    if (!activeDeck || !currentCard) return
    flipped = false
    reviewFlashcard(activeDeck.id, currentCard.id, rating)
  }

  function handleCreateDeck() {
    if (!newDeckName.trim()) return
    createFlashcardDeck(newDeckName)
    newDeckName = ''
    showNewDeck = false
    flipped = false
  }

  function handleAddCard() {
    if (!activeDeck || !newFront.trim() || !newBack.trim()) return
    addFlashcard(activeDeck.id, newFront, newBack)
    newFront = ''
    newBack = ''
    showAddCard = false
    flipped = false
  }

  function openDeckMenu(e: MouseEvent) {
    e.stopPropagation()
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
    menuCoords = { x: rect.right, y: rect.bottom + 4 }
    menuOpen = !menuOpen
  }

  function getDeckMenuItems(): ContextMenuItem[] {
    const items: ContextMenuItem[] = [
      {
        label: $tStore('flashcards.new_deck'),
        icon: FolderPlus,
        action: () => {
          showNewDeck = true
          showAddCard = false
        }
      },
      {
        label: $tStore('flashcards.restore_starter'),
        icon: RotateCw,
        action: () => {
          restoreStarterDeck()
          flipped = false
          practiceAll = false
        }
      }
    ]
    if (activeDeck && $flashcardsStore.decks.length > 1) {
      items.push({
        label: $tStore('flashcards.delete_deck'),
        icon: Trash2,
        danger: true,
        action: () => {
          if (activeDeck) deleteFlashcardDeck(activeDeck.id)
        }
      })
    }
    return items
  }
</script>

<IslandPage {expanded}>
  {#snippet minimal()}
    <div class="flashcards-minimal">
      <GraduationCap size={14} />
      <span>
        {$dueFlashcardsCount > 0
          ? $tStore('flashcards.minimal_due', { count: $dueFlashcardsCount })
          : $tStore('flashcards.minimal_done')}
      </span>
    </div>
  {/snippet}

  {#snippet expandedContent()}
    <div class="flashcards-page" data-vinta="flashcards-page">
      <PageHeader title={$tStore('flashcards.title')}>
        {#snippet actions()}
          {#if deckOptions.length > 0}
            <div class="fc-deck-select">
              <CustomSelect
                options={deckOptions}
                value={activeDeck?.id ?? ''}
                onChange={(val) => {
                  setActiveFlashcardDeck(val)
                  flipped = false
                  practiceAll = false
                }}
              />
            </div>
          {/if}
          {#if activeDeck}
            <button
              type="button"
              class="fc-icon-btn"
              use:islandTooltip={{
                text: $tStore('flashcards.add_card'),
                position: 'bottom'
              }}
              aria-label={$tStore('flashcards.add_card')}
              onclick={() => {
                showAddCard = !showAddCard
                showNewDeck = false
              }}
            >
              <Plus size={14} />
            </button>
          {/if}
          <button
            type="button"
            class="fc-icon-btn"
            aria-label={$tStore('common.more')}
            aria-haspopup="menu"
            aria-expanded={menuOpen}
            onclick={openDeckMenu}
          >
            <MoreVertical size={14} />
          </button>
        {/snippet}
      </PageHeader>

      <ScrollView class="fc-scroll-area">
        {#if showNewDeck}
          <BentoCard variant="bento" class="fc-form-card">
            <input
              type="text"
              class="fc-input"
              placeholder={$tStore('flashcards.deck_name_placeholder')}
              bind:value={newDeckName}
              onkeydown={(e) => {
                if (e.key === 'Enter') handleCreateDeck()
              }}
            />
            <div class="fc-form-actions">
              <button
                type="button"
                class="fc-btn secondary"
                onclick={() => (showNewDeck = false)}
              >
                {$tStore('common.cancel')}
              </button>
              <button
                type="button"
                class="fc-btn primary"
                disabled={!newDeckName.trim()}
                onclick={handleCreateDeck}
              >
                {$tStore('flashcards.create_deck')}
              </button>
            </div>
          </BentoCard>
        {:else if showAddCard && activeDeck}
          <BentoCard variant="bento" class="fc-form-card">
            <textarea
              class="fc-textarea"
              rows="2"
              placeholder={$tStore('flashcards.front_placeholder')}
              bind:value={newFront}></textarea>
            <textarea
              class="fc-textarea"
              rows="2"
              placeholder={$tStore('flashcards.back_placeholder')}
              bind:value={newBack}></textarea>
            <div class="fc-form-actions">
              <button
                type="button"
                class="fc-btn secondary"
                onclick={() => (showAddCard = false)}
              >
                {$tStore('common.cancel')}
              </button>
              <button
                type="button"
                class="fc-btn primary"
                disabled={!newFront.trim() || !newBack.trim()}
                onclick={handleAddCard}
              >
                {$tStore('common.save')}
              </button>
            </div>
          </BentoCard>
        {:else if !activeDeck || activeDeck.cards.length === 0}
          <BentoCard variant="bento" class="fc-empty-card">
            <div class="fc-empty-left">
              <IconBadge icon={BookOpen} color="slate" size="lg" />
              <span class="fc-empty-title">
                {$tStore('flashcards.empty_deck')}
              </span>
            </div>
            <div class="fc-form-actions">
              <button
                type="button"
                class="fc-btn secondary"
                onclick={() => restoreStarterDeck()}
              >
                {$tStore('flashcards.restore_starter')}
              </button>
              <button
                type="button"
                class="fc-btn primary"
                onclick={() => (showAddCard = true)}
              >
                {$tStore('flashcards.add_first_card')}
              </button>
            </div>
          </BentoCard>
        {:else if currentCard}
          <BentoCard variant="bento" class="fc-study-card">
            <div class="fc-card-meta">
              <span class="fc-side-tag" class:is-answer={flipped}>
                {flipped
                  ? $tStore('flashcards.back')
                  : $tStore('flashcards.front')}
              </span>
              <div class="fc-meta-right">
                <span class="fc-queue-count">
                  1 / {studyQueue.length}
                </span>
                <button
                  type="button"
                  class="fc-card-delete"
                  use:islandTooltip={{
                    text: $tStore('flashcards.delete_card'),
                    position: 'left'
                  }}
                  aria-label={$tStore('flashcards.delete_card')}
                  onclick={() => {
                    if (activeDeck && currentCard) {
                      deleteFlashcard(activeDeck.id, currentCard.id)
                      flipped = false
                    }
                  }}
                >
                  <X size={13} />
                </button>
              </div>
            </div>

            <button
              type="button"
              class="fc-card-body"
              onclick={() => (flipped = !flipped)}
              aria-label={$tStore('flashcards.flip_card')}
            >
              <p class="fc-face-text">
                {flipped ? currentCard.back : currentCard.front}
              </p>
            </button>

            <div class="fc-card-footer">
              {#if flipped}
                <div class="fc-rate-grid">
                  <button
                    type="button"
                    class="fc-rate-btn again"
                    onclick={() => handleRate('again')}
                  >
                    <span>{$tStore('flashcards.rate_again')}</span>
                    <span class="fc-interval-hint">
                      {getNextIntervalLabel(currentCard, 'again')}
                    </span>
                  </button>
                  <button
                    type="button"
                    class="fc-rate-btn good"
                    onclick={() => handleRate('good')}
                  >
                    <span>{$tStore('flashcards.rate_good')}</span>
                    <span class="fc-interval-hint">
                      {getNextIntervalLabel(currentCard, 'good')}
                    </span>
                  </button>
                  <button
                    type="button"
                    class="fc-rate-btn easy"
                    onclick={() => handleRate('easy')}
                  >
                    <span>{$tStore('flashcards.rate_easy')}</span>
                    <span class="fc-interval-hint">
                      {getNextIntervalLabel(currentCard, 'easy')}
                    </span>
                  </button>
                </div>
              {:else}
                <button
                  type="button"
                  class="fc-btn primary fc-show-btn"
                  onclick={() => (flipped = true)}
                >
                  <RotateCw size={12} />
                  <span>{$tStore('flashcards.show_answer')}</span>
                </button>
              {/if}
            </div>
          </BentoCard>
        {:else}
          <BentoCard variant="bento" class="fc-empty-card">
            <div class="fc-empty-left">
              <IconBadge icon={CheckCircle2} color="emerald" size="lg" />
              <span class="fc-empty-title">
                {$tStore('flashcards.all_caught_up')}
              </span>
            </div>
            <div class="fc-form-actions">
              <button
                type="button"
                class="fc-btn secondary"
                onclick={() => {
                  practiceAll = true
                  flipped = false
                }}
              >
                <RotateCw size={12} />
                <span>{$tStore('flashcards.practice_all')}</span>
              </button>
            </div>
          </BentoCard>
        {/if}
      </ScrollView>
    </div>

    {#if menuOpen}
      <CustomContextMenu
        x={menuCoords.x}
        y={menuCoords.y}
        items={getDeckMenuItems()}
        onclose={() => {
          menuOpen = false
        }}
      />
    {/if}
  {/snippet}
</IslandPage>

<style src="./styles/FlashcardsPage.css"></style>
