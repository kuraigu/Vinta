/**
 * @fileoverview Spaced-repetition flashcards store persisted in %APPDATA%/FreeMatrix Studios/Vinta/flashcards.json.
 */

import { writable, derived, get } from 'svelte/store'
import {
  GetFlashcardsData,
  SaveFlashcardsData
} from '../../bindings/Vinta/backend/app.js'

export type ReviewRating = 'again' | 'good' | 'easy'

export interface Flashcard {
  id: string
  front: string
  back: string
  intervalDays: number
  ease: number
  reps: number
  dueAt: number
}

export interface FlashcardDeck {
  id: string
  name: string
  cards: Flashcard[]
  createdAt: number
}

export interface FlashcardsState {
  decks: FlashcardDeck[]
  activeDeckId: string | null
  initialized: boolean
}

const STORAGE_FALLBACK_KEY = 'vinta_flashcards_fallback'

function createStarterDecks(): FlashcardDeck[] {
  const now = Date.now()
  return [
    {
      id: 'starter-quickstart',
      name: 'Quick Start',
      createdAt: now,
      cards: [
        {
          id: 'qs-1',
          front: 'What is the 20-20-20 eye rule?',
          back: 'Every 20 minutes, look 20 feet away for 20 seconds.',
          intervalDays: 0,
          ease: 2.5,
          reps: 0,
          dueAt: now
        },
        {
          id: 'qs-2',
          front: 'How do you open Desktop Apps from anywhere?',
          back: 'Press Alt+Shift+K from any window.',
          intervalDays: 0,
          ease: 2.5,
          reps: 0,
          dueAt: now
        },
        {
          id: 'qs-3',
          front: 'How long is a Pomodoro focus sprint?',
          back: '25 minutes of focus followed by a 5-minute break.',
          intervalDays: 0,
          ease: 2.5,
          reps: 0,
          dueAt: now
        }
      ]
    }
  ]
}

/**
 * Returns a short human-readable next interval label for a rating button (e.g. "10m", "1d", "3d").
 */
export function getNextIntervalLabel(
  card: Flashcard,
  rating: ReviewRating
): string {
  if (rating === 'again') return '10m'
  if (rating === 'good') {
    const days =
      card.reps === 0
        ? 1
        : card.reps === 1
          ? 3
          : Math.max(1, Math.round(card.intervalDays * card.ease))
    return `${days}d`
  }
  const nextEase = Math.min(3.0, card.ease + 0.15)
  const days =
    card.reps === 0
      ? 3
      : Math.max(2, Math.round(card.intervalDays * nextEase * 1.35))
  return `${days}d`
}

export const flashcardsStore = writable<FlashcardsState>({
  decks: [],
  activeDeckId: null,
  initialized: false
})

export const dueFlashcardsCount = derived(flashcardsStore, ($state) => {
  const now = Date.now()
  let count = 0
  for (const deck of $state.decks) {
    for (const card of deck.cards) {
      if (card.dueAt <= now) count++
    }
  }
  return count
})

let saveTimer: ReturnType<typeof setTimeout> | null = null

function persistFlashcards(state: FlashcardsState): void {
  if (!state.initialized) return
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    const payload = JSON.stringify({
      decks: state.decks,
      activeDeckId: state.activeDeckId
    })
    try {
      localStorage.setItem(STORAGE_FALLBACK_KEY, payload)
    } catch {
      // ignore storage quota errors
    }
    if (typeof SaveFlashcardsData === 'function') {
      void SaveFlashcardsData(payload).catch(() => {})
    }
  }, 250)
}

/**
 * Loads flashcard decks once from flashcards.json (with localStorage fallback).
 */
export async function initFlashcardsStore(): Promise<void> {
  if (get(flashcardsStore).initialized) return

  let raw = ''
  try {
    if (typeof GetFlashcardsData === 'function') {
      raw = await GetFlashcardsData()
    }
  } catch {
    // fallback below
  }

  if (!raw || raw === '{}') {
    try {
      raw = localStorage.getItem(STORAGE_FALLBACK_KEY) || ''
    } catch {
      raw = ''
    }
  }

  let decks: FlashcardDeck[] = []
  let activeDeckId: string | null = null

  if (raw && raw !== '{}') {
    try {
      const parsed = JSON.parse(raw) as {
        decks?: FlashcardDeck[]
        activeDeckId?: string | null
      }
      if (Array.isArray(parsed.decks)) {
        decks = parsed.decks
      }
      if (typeof parsed.activeDeckId === 'string') {
        activeDeckId = parsed.activeDeckId
      }
    } catch {
      // ignore malformed json
    }
  }

  if (
    decks.length === 0 ||
    (decks.length === 1 &&
      ((decks[0].id === 'starter-shortcuts' && decks[0].cards.length === 0) ||
        decks[0].id === 'starter-essentials'))
  ) {
    decks = createStarterDecks()
    activeDeckId = decks[0].id
  } else if (!activeDeckId || !decks.some((d) => d.id === activeDeckId)) {
    activeDeckId = decks[0]?.id ?? null
  }

  const nextState: FlashcardsState = {
    decks,
    activeDeckId,
    initialized: true
  }
  flashcardsStore.set(nextState)
  persistFlashcards(nextState)
}

function updateAndPersist(
  updater: (prev: FlashcardsState) => FlashcardsState
): void {
  flashcardsStore.update((prev) => {
    const next = updater(prev)
    persistFlashcards(next)
    return next
  })
}

export function setActiveFlashcardDeck(deckId: string): void {
  updateAndPersist((prev) => ({ ...prev, activeDeckId: deckId }))
}

export function createFlashcardDeck(name: string): string {
  const trimmed = name.trim()
  if (!trimmed) return ''
  const id = `deck-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`
  updateAndPersist((prev) => ({
    ...prev,
    activeDeckId: id,
    decks: [
      ...prev.decks,
      { id, name: trimmed, cards: [], createdAt: Date.now() }
    ]
  }))
  return id
}

export function deleteFlashcardDeck(deckId: string): void {
  updateAndPersist((prev) => {
    const decks = prev.decks.filter((d) => d.id !== deckId)
    const activeDeckId =
      prev.activeDeckId === deckId
        ? (decks[0]?.id ?? null)
        : prev.activeDeckId
    return { ...prev, decks, activeDeckId }
  })
}

/**
 * Restores the built-in starter deck (Focus & Study Essentials) and selects it.
 */
export function restoreStarterDeck(): void {
  const starter = createStarterDecks()[0]
  updateAndPersist((prev) => {
    const filtered = prev.decks.filter((d) => d.id !== starter.id)
    return {
      ...prev,
      activeDeckId: starter.id,
      decks: [starter, ...filtered]
    }
  })
}

export function addFlashcard(deckId: string, front: string, back: string): void {
  const f = front.trim()
  const b = back.trim()
  if (!f || !b) return
  const card: Flashcard = {
    id: `card-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
    front: f,
    back: b,
    intervalDays: 0,
    ease: 2.5,
    reps: 0,
    dueAt: Date.now()
  }
  updateAndPersist((prev) => ({
    ...prev,
    decks: prev.decks.map((d) =>
      d.id === deckId ? { ...d, cards: [...d.cards, card] } : d
    )
  }))
}

export function deleteFlashcard(deckId: string, cardId: string): void {
  updateAndPersist((prev) => ({
    ...prev,
    decks: prev.decks.map((d) =>
      d.id === deckId
        ? { ...d, cards: d.cards.filter((c) => c.id !== cardId) }
        : d
    )
  }))
}

/**
 * Applies spaced-repetition scheduling for a reviewed flashcard.
 */
export function reviewFlashcard(
  deckId: string,
  cardId: string,
  rating: ReviewRating
): void {
  const now = Date.now()
  const dayMs = 24 * 60 * 60 * 1000
  updateAndPersist((prev) => ({
    ...prev,
    decks: prev.decks.map((d) => {
      if (d.id !== deckId) return d
      return {
        ...d,
        cards: d.cards.map((c) => {
          if (c.id !== cardId) return c
          if (rating === 'again') {
            return {
              ...c,
              reps: 0,
              intervalDays: 0,
              ease: Math.max(1.3, c.ease - 0.2),
              dueAt: now + 10 * 60 * 1000
            }
          }
          const nextReps = c.reps + 1
          let nextInterval = 1
          let nextEase = c.ease
          if (rating === 'good') {
            nextInterval =
              c.reps === 0
                ? 1
                : c.reps === 1
                  ? 3
                  : Math.max(1, Math.round(c.intervalDays * c.ease))
          } else {
            nextEase = Math.min(3.0, c.ease + 0.15)
            nextInterval =
              c.reps === 0
                ? 3
                : Math.max(2, Math.round(c.intervalDays * nextEase * 1.35))
          }
          return {
            ...c,
            reps: nextReps,
            intervalDays: nextInterval,
            ease: nextEase,
            dueAt: now + nextInterval * dayMs
          }
        })
      }
    })
  }))
}
