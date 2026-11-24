<!--
  @component PinnedNoteWidget
  @description Compact Home Bento widget displaying the note pinned to Home (with quick launch to Notes).
-->
<script lang="ts">
  import { getContext, onMount } from 'svelte'
  import { FileText, Pin, ChevronRight } from 'lucide-svelte'
  import WidgetFrame from './WidgetFrame.svelte'
  import BentoCard from '../BentoCard.svelte'
  import IconBadge from '../IconBadge.svelte'
  import NotesPage, {
    config as notesConfig
  } from '../../pages/NotesPage.svelte'
  import { ReadNote } from '../../../bindings/Vinta/backend/app.js'
  import {
    pinnedHomeNoteFileStore,
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
    size = '2x1',
    showMenuButton = false,
    onreorder = undefined,
    oncontextmenu = undefined
  }: Props = $props()

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let noteTitle = $state('')
  let notePreview = $state('')

  function stripHtml(html: string): string {
    return html
      .replace(/<br\s*\/?>/gi, ' ')
      .replace(/<[^>]*>/g, ' ')
      .replace(/&nbsp;/g, ' ')
      .replace(/\s+/g, ' ')
      .trim()
  }

  async function refreshPinnedNote(fileName: string | null) {
    if (!fileName) {
      noteTitle = ''
      notePreview = ''
      return
    }
    try {
      let raw = ''
      if (typeof ReadNote === 'function') {
        raw = await ReadNote(fileName)
      }
      if (!raw) {
        noteTitle = ''
        notePreview = ''
        return
      }
      const parsed = JSON.parse(raw) as { title?: string; content?: string }
      noteTitle = parsed.title?.trim() || $tStore('notes.untitled')
      notePreview = stripHtml(parsed.content || '')
    } catch {
      noteTitle = ''
      notePreview = ''
    }
  }

  $effect(function loadPinnedHomeNoteEffect() {
    void refreshPinnedNote($pinnedHomeNoteFileStore)
  })

  onMount(() => {
    void refreshPinnedNote($pinnedHomeNoteFileStore)
  })

  const canOpen = $derived($pageVisibilityStore.notes !== false)
</script>

<WidgetFrame
  id="pinnedNote"
  {size}
  {showMenuButton}
  {onreorder}
  {oncontextmenu}
>
  <BentoCard
    variant="bento"
    class="pn-glance-card"
    disabled={!canOpen}
    onclick={() => {
      if (canOpen) navigate(NotesPage, notesConfig, true)
    }}
  >
    <div class="glance-header">
      <div class="glance-title-group">
        <IconBadge
          icon={$pinnedHomeNoteFileStore ? Pin : FileText}
          color="surface"
          size="sm"
        />
        <span
          class="glance-title-text"
          title={noteTitle || $tStore('notes.pinned_home_title')}
        >
          {noteTitle || $tStore('notes.pinned_home_title')}
        </span>
      </div>
      {#if canOpen}
        <ChevronRight size={14} class="glance-nav-chevron" />
      {/if}
    </div>
    <p class="pn-preview">
      {notePreview || $tStore('notes.pinned_home_empty')}
    </p>
  </BentoCard>
</WidgetFrame>

<style src="./styles/PinnedNoteWidget.css"></style>
