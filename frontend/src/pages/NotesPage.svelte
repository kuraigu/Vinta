<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 },
    expanded: { w: 540, h: 440, r: 32 },
    borderColor: 'var(--c-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { onMount } from 'svelte'
  import { Check, Lock, Unlock, FileText, Plus, Save } from 'lucide-svelte'
  import SidebarLayout from '../components/SidebarLayout.svelte'
  import NoteSidebar from '../components/notes/NoteSidebar.svelte'
  import NoteToolbar from '../components/notes/NoteToolbar.svelte'
  import NoteLive from '../components/notes/NoteLive.svelte'
  import {
    ListNotes,
    ReadNote,
    SaveNote,
    DeleteNote,
    OpenNotesFolder,
    RenameNote
  } from '../../wailsjs/go/backend/App.js'

  interface NoteItem {
    id: string
    title: string
    fileName: string
    content: string
    modTime: number
  }

  let notes = $state<NoteItem[]>([])
  let selectedNoteId = $state<string>('')
  let searchQuery = $state<string>('')
  let isLocked = $state<boolean>(false)
  let isDirty = $state<boolean>(false)
  let saveStatus = $state<string>('')
  let saveTimeout: ReturnType<typeof setTimeout> | null = null
  let noteLiveApi = $state<{
    insertFormatting: (
      type: FormatType,
      prefix: string,
      suffix?: string
    ) => void
    selectAll: () => void
  } | null>(null)

  // Active note binding helpers
  let activeIndex = $derived(notes.findIndex((n) => n.id === selectedNoteId))
  let activeNote = $derived(activeIndex !== -1 ? notes[activeIndex] : null)

  let filteredNotes = $derived(
    notes.filter((n) => {
      const q = searchQuery.toLowerCase().trim()
      if (!q) return true
      return (
        n.title.toLowerCase().includes(q) || n.content.toLowerCase().includes(q)
      )
    })
  )

  async function loadNotesFromBackend() {
    try {
      const fileList = await ListNotes()
      if (fileList && Array.isArray(fileList) && fileList.length > 0) {
        const loadedNotes: NoteItem[] = []
        for (const f of fileList) {
          let text = ''
          try {
            text = await ReadNote(f.fileName)
          } catch {
            text = ''
          }
          loadedNotes.push({
            id: f.id,
            title: f.title,
            fileName: f.fileName,
            content: text,
            modTime: f.modTime
          })
        }
        loadedNotes.sort((a, b) => b.modTime - a.modTime)
        notes = loadedNotes
        if (notes.length > 0) {
          selectedNoteId = notes[0].id
        }
        return
      }
    } catch {
      // Fallback for dev mode / web browser
    }

    notes = []
    selectedNoteId = ''

    if (notes.length === 0) {
      createNewNote()
    }
  }

  onMount(() => {
    loadNotesFromBackend()
  })

  function createNewNote() {
    // Find the next available "Untitled" name (Untitled, Untitled (2), Untitled (3), ...)
    const existingTitles = new Set(notes.map((n) => n.title))
    let newTitle = 'Untitled'
    if (existingTitles.has('Untitled') || existingTitles.has('untitled')) {
      let n = 2
      while (
        existingTitles.has(`Untitled (${n})`) ||
        existingTitles.has(`untitled (${n})`)
      )
        n++
      newTitle = `Untitled (${n})`
    }
    const newFileName = `${newTitle}.md`
    const newNote: NoteItem = {
      id: newTitle,
      title: newTitle,
      fileName: newFileName,
      content: '',
      modTime: Date.now()
    }

    notes = [newNote, ...notes]
    selectedNoteId = newTitle

    persistNote(newNote)
  }

  function persistNote(note: NoteItem) {
    note.modTime = Date.now()
    try {
      SaveNote(note.fileName, note.content).catch(() => {})
    } catch {
      // Fallback
    }
  }

  function saveActiveNote() {
    if (activeIndex === -1 || !notes[activeIndex]) return
    persistNote(notes[activeIndex])
    isDirty = false
    saveStatus = 'saved'
    if (saveTimeout) clearTimeout(saveTimeout)
    saveTimeout = setTimeout(() => {
      if (saveStatus === 'saved') saveStatus = ''
    }, 1500)
  }

  function handleContentChange() {
    if (activeIndex === -1) return
    isDirty = true
  }

  function handleSelectNote(id: string) {
    if (id === selectedNoteId) return
    if (isDirty && activeNote) {
      persistNote(activeNote)
      isDirty = false
    }
    selectedNoteId = id
  }

  function toggleLock() {
    if (isDirty && activeNote) {
      persistNote(activeNote)
      isDirty = false
    }
    isLocked = !isLocked
  }

  function handleTitleChange() {
    if (activeIndex === -1) return
    saveStatus = 'saving'
    if (saveTimeout) clearTimeout(saveTimeout)
    saveTimeout = setTimeout(async () => {
      if (activeIndex === -1 || !notes[activeIndex]) return
      const current = notes[activeIndex]
      const originalTitle = current.title.trim() || 'Untitled'
      try {
        const newFileName = await RenameNote(current.fileName, originalTitle)
        notes[activeIndex] = {
          ...current,
          title: originalTitle,
          fileName: newFileName,
          id: newFileName.replace(/\.md$/i, '')
        }
        selectedNoteId = notes[activeIndex].id
      } catch {
        // Fallback: file rename failed, just save content
      }
      persistNote(notes[activeIndex])
    }, 600)
  }

  async function deleteSelectedNote(e: Event, id: string) {
    e.stopPropagation()

    if (saveTimeout) {
      clearTimeout(saveTimeout)
      saveTimeout = null
    }

    const target = notes.find((n) => n.id === id)
    notes = notes.filter((n) => n.id !== id)

    if (selectedNoteId === id) {
      selectedNoteId = notes.length > 0 ? notes[0].id : ''
    }

    if (target) {
      try {
        await DeleteNote(target.fileName)
      } catch {
        // Fallback
      }
    }

    if (notes.length === 0) {
      createNewNote()
    }
  }

  function handleOpenFolder() {
    try {
      OpenNotesFolder().catch(() => {})
    } catch {
      // Fallback
    }
  }

  import type { FormatType } from '../components/notes/NoteToolbar.svelte'

  function insertFormatting(
    type: FormatType,
    prefix: string,
    suffix: string = ''
  ) {
    if (activeIndex === -1 || !notes[activeIndex]) return
    if (noteLiveApi) {
      noteLiveApi.insertFormatting(type, prefix, suffix)
    }
  }

  function formatTimeAgo(timestamp: number): string {
    const diff = Math.floor((Date.now() - timestamp) / 1000)
    if (diff < 60) return 'Just now'
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
    return `${Math.floor(diff / 86400)}d ago`
  }
</script>

<SidebarLayout title="Quick Notes">
  {#snippet sidebar()}
    <NoteSidebar
      notes={filteredNotes}
      {selectedNoteId}
      bind:searchQuery
      onSelectNote={handleSelectNote}
      onNewNote={createNewNote}
      onDeleteNote={deleteSelectedNote}
      onOpenFolder={handleOpenFolder}
      {formatTimeAgo}
    />
  {/snippet}

  {#snippet content()}
    {#if activeNote && activeIndex !== -1}
      <div class="editor-header">
        <input
          type="text"
          class="title-input"
          bind:value={notes[activeIndex].title}
          oninput={handleTitleChange}
          disabled={isLocked}
          placeholder="Note Title..."
        />
        <div class="header-right">
          <div class="save-badge-container">
            {#if isDirty}
              <span class="save-badge unsaved" title="Unsaved changes">
                <span class="dot-unsaved">●</span> Unsaved
              </span>
            {:else if saveStatus === 'saved'}
              <span class="save-badge saved">
                <Check size={11} /> Saved
              </span>
            {/if}
          </div>
          <button
            class="save-btn"
            class:dirty={isDirty}
            disabled={!isDirty}
            onclick={saveActiveNote}
            title="Save Note (Ctrl+S)"
          >
            <Save size={13} />
            <span>Save</span>
          </button>
          <button
            class="view-toggle-btn"
            class:active={isLocked}
            onclick={toggleLock}
            title={isLocked ? 'Unlock Note to Edit' : 'Lock Note (Read Only)'}
          >
            {#if isLocked}
              <Lock size={13} />
              <span>Locked</span>
            {:else}
              <Unlock size={13} />
              <span>Unlocked</span>
            {/if}
          </button>
        </div>
      </div>

      {#if !isLocked}
        <NoteToolbar onInsertFormatting={insertFormatting} />
      {/if}

      <div class="editor-body">
        <NoteLive
          content={notes[activeIndex].content}
          {isLocked}
          onchange={(newContent) => {
            if (activeIndex !== -1 && notes[activeIndex]) {
              notes[activeIndex].content = newContent
              handleContentChange()
            }
          }}
          onSave={saveActiveNote}
          onReady={(api) => {
            noteLiveApi = api
          }}
        />
      </div>
    {:else}
      <div class="empty-editor">
        <FileText size={32} class="empty-icon" />
        <span>No note selected</span>
        <button class="new-note-btn small" onclick={createNewNote}>
          <Plus size={13} /> Create Note
        </button>
      </div>
    {/if}
  {/snippet}
</SidebarLayout>

<style src="./styles/NotesPage.css"></style>
