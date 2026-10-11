<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 },
    expanded: { w: 600, h: 480, r: 28 },
    borderColor: 'var(--c-border)',
    glowColor: 'transparent'
  }

  export type StickyColor =
    'yellow' | 'green' | 'pink' | 'purple' | 'blue' | 'charcoal'

  export interface VNote {
    id: string
    title: string
    fileName: string
    color: StickyColor
    content: string
    pinned?: boolean
    createdAt: number
    modTime: number
  }
</script>

<script lang="ts">
  import { onMount, tick } from 'svelte'
  import {
    Plus,
    Search,
    FolderOpen,
    Trash2,
    List,
    Bold,
    Italic,
    Underline,
    Strikethrough,
    ListOrdered,
    Image as ImageIcon
  } from 'lucide-svelte'
  import BackButton from '../components/BackButton.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import {
    ListNotes,
    ReadNote,
    SaveNote,
    DeleteNote,
    OpenNotesFolder,
    RenameNote
  } from '../../bindings/Vinta/backend/app.js'

  let notes = $state<VNote[]>([])
  let activeIndex = $state<number>(0)
  let viewMode = $state<'board' | 'editor'>('board')
  let isDrawerOpen = $state(false)
  let drawerSearchQuery = $state('')
  let isDraggingOver = $state(false)
  let saveStatus = $state<'saved' | 'saving' | ''>('')
  let saveTimer: ReturnType<typeof setTimeout> | null = null
  let renameTimer: ReturnType<typeof setTimeout> | null = null

  let editorRef = $state<HTMLDivElement | null>(null)
  let fileInputRef = $state<HTMLInputElement | null>(null)

  const colorPaletteCycle: StickyColor[] = [
    'yellow',
    'green',
    'pink',
    'purple',
    'blue',
    'charcoal'
  ]

  let activeNote = $derived(
    notes.length > 0 && activeIndex >= 0 && activeIndex < notes.length
      ? notes[activeIndex]
      : null
  )

  let filteredDrawerNotes = $derived(
    notes.filter((n) => {
      const q = drawerSearchQuery.toLowerCase().trim()
      if (!q) return true
      const snippet = n.content.replace(/<[^>]*>/g, '').toLowerCase()
      return n.title.toLowerCase().includes(q) || snippet.includes(q)
    })
  )

  /**
   * Loads all .vnote files from disk and populates the in-memory notes store.
   */
  async function loadNotesFromBackend() {
    try {
      const fileList = await ListNotes()
      if (fileList && Array.isArray(fileList) && fileList.length > 0) {
        const loaded: VNote[] = []
        for (const f of fileList) {
          let raw = ''
          try {
            raw = await ReadNote(f.fileName)
          } catch {
            raw = ''
          }

          let noteObj: VNote | null = null
          if (raw) {
            try {
              const parsed = JSON.parse(raw)
              if (parsed && typeof parsed === 'object') {
                noteObj = {
                  id: f.id,
                  title: parsed.title || f.title || 'Untitled',
                  fileName: f.fileName,
                  color: parsed.color || 'yellow',
                  content: parsed.content || '',
                  pinned: Boolean(parsed.pinned),
                  createdAt: parsed.createdAt || f.modTime || Date.now(),
                  modTime: parsed.modTime || f.modTime || Date.now()
                }
              }
            } catch {
              // Plaintext fallback
            }
          }

          if (!noteObj) {
            noteObj = {
              id: f.id,
              title: f.title || 'Untitled',
              fileName: f.fileName,
              color: 'yellow',
              content: raw || '',
              pinned: false,
              createdAt: f.modTime || Date.now(),
              modTime: f.modTime || Date.now()
            }
          }
          loaded.push(noteObj)
        }

        loaded.sort(
          (a, b) => (b.modTime || b.createdAt) - (a.modTime || a.createdAt)
        )
        notes = loaded
        activeIndex = 0
        syncEditorContent()
        return
      }
    } catch {
      // Fallback
    }

    notes = []
  }

  /**
   * Populates the contenteditable node with the current active note's HTML.
   */
  function syncEditorContent() {
    void tick().then(() => {
      if (editorRef && activeNote) {
        editorRef.innerHTML = activeNote.content || ''
      }
    })
  }

  /**
   * Creates a new note and focuses it immediately in editor mode.
   */
  function createNewNote() {
    const existingTitles = new Set(notes.map((n) => n.title.toLowerCase()))
    let newTitle = 'Untitled'
    if (existingTitles.has('untitled')) {
      let counter = 2
      while (existingTitles.has(`untitled (${counter})`)) {
        counter++
      }
      newTitle = `Untitled (${counter})`
    }

    const nextColor = colorPaletteCycle[notes.length % colorPaletteCycle.length]
    const newFileName = `${newTitle}.vnote`
    const now = Date.now()

    const newNote: VNote = {
      id: newTitle,
      title: newTitle,
      fileName: newFileName,
      color: nextColor,
      content: '',
      pinned: false,
      createdAt: now,
      modTime: now
    }

    notes = [newNote, ...notes]
    activeIndex = 0
    isDrawerOpen = false
    viewMode = 'editor'
    persistNote(newNote)
    syncEditorContent()
  }

  /**
   * Persists the given note document to disk as formatted JSON.
   *
   * @param note - Target note object.
   */
  function persistNote(note: VNote) {
    note.modTime = Date.now()
    const payload = {
      id: note.id,
      title: note.title,
      fileName: note.fileName,
      color: note.color || 'yellow',
      content: note.content || '',
      pinned: Boolean(note.pinned),
      createdAt: note.createdAt || Date.now(),
      modTime: note.modTime
    }

    try {
      SaveNote(note.fileName, JSON.stringify(payload, null, 2)).catch(() => {})
    } catch {
      // Fallback
    }
  }

  /**
   * Schedules a debounced auto-save operation.
   */
  function triggerAutoSave() {
    saveStatus = 'saving'
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(() => {
      if (editorRef && activeNote) {
        activeNote.content = editorRef.innerHTML
        persistNote(activeNote)
        saveStatus = 'saved'
        setTimeout(() => {
          if (saveStatus === 'saved') saveStatus = ''
        }, 1500)
      }
    }, 400)
  }

  /**
   * Handles note title change and debounced rename on disk.
   */
  function handleTitleChange() {
    triggerAutoSave()
    if (renameTimer) clearTimeout(renameTimer)
    renameTimer = setTimeout(async () => {
      if (!activeNote) return
      const cleanTitle = activeNote.title.trim() || 'Untitled'
      try {
        const updatedFileName = await RenameNote(
          activeNote.fileName,
          cleanTitle
        )
        activeNote.fileName = updatedFileName
        activeNote.id = updatedFileName.replace(/\.vnote$/i, '')
        persistNote(activeNote)
      } catch {
        persistNote(activeNote)
      }
    }, 600)
  }

  /**
   * Changes the color theme of the active note.
   *
   * @param color - The chosen sticky note color theme.
   */
  function handleSelectColor(color: StickyColor) {
    if (!activeNote) return
    activeNote.color = color
    triggerAutoSave()
  }

  /**
   * Deletes the active note from disk.
   */
  async function handleDeleteActiveNote() {
    if (!activeNote) return
    const target = activeNote
    const currentIdx = activeIndex

    notes = notes.filter((_, idx) => idx !== currentIdx)
    activeIndex = Math.max(0, currentIdx - 1)

    try {
      await DeleteNote(target.fileName)
    } catch {
      // Fallback
    }

    viewMode = 'board'
  }

  /**
   * Executes rich text formatting commands on selection.
   *
   * @param command - The action identifier.
   */
  function applyFormat(command: string) {
    if (!editorRef) return
    editorRef.focus()
    document.execCommand(command, false)
    triggerAutoSave()
  }

  /**
   * Inserts an image directly into the text editor at the cursor position.
   *
   * @param dataUrl - Base64 image data URL.
   * @param name - Image filename.
   */
  function insertInlineImage(dataUrl: string, name?: string) {
    if (!editorRef) return
    editorRef.focus()
    const imgHtml = `<p><img src="${dataUrl}" class="note-attached-image" alt="${name || 'attachment'}" /></p><p><br></p>`
    const success = document.execCommand('insertHTML', false, imgHtml)
    if (!success) {
      editorRef.innerHTML += imgHtml
    }
    triggerAutoSave()
  }

  /**
   * Reads an image file as base64 data URL and attaches it inline.
   *
   * @param file - Target image file.
   */
  function processImageFile(file: File) {
    if (!file.type.startsWith('image/')) return
    const reader = new FileReader()
    reader.onload = (e) => {
      const dataUrl = e.target?.result as string
      if (dataUrl) {
        insertInlineImage(dataUrl, file.name)
      }
    }
    reader.readAsDataURL(file)
  }

  /**
   * Handles native file picker change.
   *
   * @param e - Input change event.
   */
  function handleFileInputChange(e: Event) {
    const target = e.target as HTMLInputElement
    if (target.files && target.files.length > 0) {
      for (let i = 0; i < target.files.length; i++) {
        processImageFile(target.files[i])
      }
      target.value = ''
    }
  }

  /**
   * Handles clipboard paste events to embed copied images.
   *
   * @param e - Clipboard paste event.
   */
  function handlePaste(e: ClipboardEvent) {
    const items = e.clipboardData?.items
    if (items) {
      for (let i = 0; i < items.length; i++) {
        if (items[i].type.startsWith('image/')) {
          const file = items[i].getAsFile()
          if (file) {
            e.preventDefault()
            processImageFile(file)
            return
          }
        }
      }
    }
    triggerAutoSave()
  }

  /**
   * Handles keyboard shortcuts inside the editor surface.
   *
   * @param e - Keyboard event.
   */
  function handleKeyDown(e: KeyboardEvent) {
    if (e.ctrlKey || e.metaKey) {
      if (e.key === 'b' || e.key === 'B') {
        e.preventDefault()
        applyFormat('bold')
      } else if (e.key === 'i' || e.key === 'I') {
        e.preventDefault()
        applyFormat('italic')
      } else if (e.key === 'u' || e.key === 'U') {
        e.preventDefault()
        applyFormat('underline')
      } else if (e.key === 's' || e.key === 'S') {
        e.preventDefault()
        if (editorRef && activeNote) {
          activeNote.content = editorRef.innerHTML
          persistNote(activeNote)
          saveStatus = 'saved'
          setTimeout(() => {
            if (saveStatus === 'saved') saveStatus = ''
          }, 1500)
        }
      }
    }
  }

  /**
   * Handles dragover for dropping image files.
   *
   * @param e - Drag event.
   */
  function handleDragOver(e: DragEvent) {
    e.preventDefault()
    isDraggingOver = true
  }

  /**
   * Handles dragleave.
   */
  function handleDragLeave() {
    isDraggingOver = false
  }

  /**
   * Handles drop of image files.
   *
   * @param e - Drag event.
   */
  function handleDrop(e: DragEvent) {
    e.preventDefault()
    isDraggingOver = false
    if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
      for (let i = 0; i < e.dataTransfer.files.length; i++) {
        processImageFile(e.dataTransfer.files[i])
      }
    }
  }

  /**
   * Opens the Vinta notes folder in Windows File Explorer.
   */
  function handleOpenFolder() {
    try {
      OpenNotesFolder().catch(() => {})
    } catch {
      // Fallback
    }
  }

  /**
   * Opens the chosen note in the editor view.
   *
   * @param idx - Index of note to open.
   */
  function openNoteEditor(idx: number) {
    activeIndex = idx
    viewMode = 'editor'
    syncEditorContent()
  }

  /**
   * Returns from editor view back to the main note cards screen.
   */
  function handleBackToBoard() {
    if (editorRef && activeNote) {
      activeNote.content = editorRef.innerHTML
      persistNote(activeNote)
    }
    viewMode = 'board'
  }

  /**
   * Deletes a note directly from the card board.
   *
   * @param e - Mouse event.
   * @param idx - Index of note to remove.
   */
  async function deleteNoteByIndex(e: MouseEvent, idx: number) {
    e.stopPropagation()
    const target = notes[idx]
    if (!target) return
    notes = notes.filter((_, i) => i !== idx)
    if (activeIndex >= notes.length) {
      activeIndex = Math.max(0, notes.length - 1)
    }
    try {
      await DeleteNote(target.fileName)
    } catch {
      // Fallback
    }
  }

  /**
   * Extracts a plaintext snippet from HTML content.
   *
   * @param html - Note content string.
   * @returns Clean preview excerpt.
   */
  function extractSnippet(html: string): string {
    if (!html) return 'Empty note'
    const text = html
      .replace(/<[^>]*>/g, ' ')
      .replace(/&nbsp;/g, ' ')
      .replace(/\s+/g, ' ')
      .trim()
    return text || 'Empty note'
  }

  /**
   * Checks whether the note contains any inline attached images.
   *
   * @param html - Note content string.
   * @returns True if images are embedded.
   */
  function hasImage(html: string): boolean {
    return Boolean(html && html.includes('<img'))
  }

  /**
   * Relative time formatter.
   *
   * @param timestamp - Millisecond epoch timestamp.
   * @returns Formatted relative time string.
   */
  function formatTimeAgo(timestamp: number): string {
    const diff = Math.floor((Date.now() - timestamp) / 1000)
    if (diff < 60) return 'Just now'
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
    return `${Math.floor(diff / 86400)}d ago`
  }

  onMount(() => {
    loadNotesFromBackend()
  })
</script>

<div
  class="notes-page-container {viewMode === 'editor'
    ? `color-${activeNote?.color || 'yellow'}`
    : ''}"
  ondragover={handleDragOver}
  ondragleave={handleDragLeave}
  ondrop={handleDrop}
  role="region"
  aria-label="Quick Notes"
>
  {#if isDraggingOver}
    <div class="drag-drop-overlay">
      <span>Drop image to insert into note</span>
    </div>
  {/if}

  <input
    type="file"
    accept="image/*"
    multiple
    bind:this={fileInputRef}
    onchange={handleFileInputChange}
    style="display: none;"
  />

  {#if viewMode === 'board'}
    <!-- Main Notes Screen (Displays all note cards) -->
    <div class="notes-board-view">
      <div class="board-header">
        <div class="board-header-left">
          <BackButton title="Back to Apps" />
          <span class="board-title">Quick Notes</span>
          {#if notes.length > 0}
            <span class="board-count-pill">{notes.length}</span>
          {/if}
        </div>

        <div class="board-search-box">
          <Search size={13} class="board-search-icon" />
          <input
            type="text"
            placeholder="Search notes..."
            class="board-search-input"
            bind:value={drawerSearchQuery}
          />
        </div>

        <div class="board-header-right">
          <button
            class="board-action-btn"
            title="Open Notes Folder in Explorer"
            onclick={handleOpenFolder}
          >
            <FolderOpen size={14} />
          </button>
          <button
            class="board-action-btn new-note-action"
            title="Create Note"
            onclick={createNewNote}
          >
            <Plus size={16} />
          </button>
        </div>
      </div>

      <!-- Notes Grid inside ScrollView with dock clearance -->
      <div class="board-grid-wrapper">
        <ScrollView>
          {#if filteredDrawerNotes.length === 0}
            <div class="empty-board-state">
              {#if drawerSearchQuery}
                <p>No notes matching "{drawerSearchQuery}"</p>
              {:else}
                <p>No notes yet</p>
                <button class="board-create-btn" onclick={createNewNote}>
                  <Plus size={14} /> Create your first note
                </button>
              {/if}
            </div>
          {:else}
            <div class="notes-cards-grid">
              {#each filteredDrawerNotes as n, idx (n.id)}
                {@const realIndex = notes.findIndex((item) => item.id === n.id)}
                <div
                  class="note-card card-{n.color || 'yellow'}"
                  onclick={() =>
                    openNoteEditor(realIndex >= 0 ? realIndex : idx)}
                  role="button"
                  tabindex="0"
                  onkeydown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      openNoteEditor(realIndex >= 0 ? realIndex : idx)
                    }
                  }}
                >
                  <div class="note-card-band">
                    <button
                      class="note-card-delete-btn"
                      title="Delete note"
                      onclick={(e) =>
                        deleteNoteByIndex(e, realIndex >= 0 ? realIndex : idx)}
                    >
                      <Trash2 size={12} />
                    </button>
                  </div>

                  <div class="note-card-body">
                    <span class="note-card-title">{n.title || 'Untitled'}</span>
                    <p class="note-card-snippet">{extractSnippet(n.content)}</p>
                  </div>

                  <div class="note-card-footer">
                    <span class="note-card-date"
                      >{formatTimeAgo(n.modTime || n.createdAt)}</span
                    >
                    {#if hasImage(n.content)}
                      <span class="note-card-has-img" title="Contains image">
                        <ImageIcon size={11} />
                      </span>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </ScrollView>
      </div>
    </div>
  {:else}
    <!-- Active Note Editor (User approved sticky canvas) -->
    <!-- Subtle Banner Header -->
    <div class="banner-header">
      <div class="banner-left-group">
        <BackButton onback={handleBackToBoard} title="Back to Notes" />

        <button
          class="banner-btn"
          title="All notes"
          onclick={handleBackToBoard}
        >
          <List size={15} />
        </button>

        <button class="banner-btn" title="New note" onclick={createNewNote}>
          <Plus size={16} />
        </button>
      </div>

      <!-- Center: Note Title right in the subtle banner -->
      <div class="banner-center-title">
        {#if activeNote}
          <input
            type="text"
            class="banner-title-input"
            bind:value={activeNote.title}
            oninput={handleTitleChange}
            placeholder="Note Title..."
          />
        {/if}
      </div>

      <!-- Right: Colors, Folder, Delete -->
      <div class="banner-right-group">
        <div class="banner-color-picker">
          {#each colorPaletteCycle as c (c)}
            <button
              class="swatch-dot-btn dot-{c} {activeNote?.color === c
                ? 'selected'
                : ''}"
              title="{c} color"
              onclick={() => handleSelectColor(c)}
            ></button>
          {/each}
        </div>

        <button
          class="banner-btn"
          title="Open Notes Folder in Explorer"
          onclick={handleOpenFolder}
        >
          <FolderOpen size={14} />
        </button>

        <button
          class="banner-btn"
          title="Delete note"
          onclick={handleDeleteActiveNote}
        >
          <Trash2 size={14} />
        </button>
      </div>
    </div>

    <!-- Compact Formatting Toolbar -->
    <div class="notes-toolbar">
      <div class="toolbar-actions-group">
        <button
          class="tool-action-btn"
          title="Bold (Ctrl+B)"
          onclick={() => applyFormat('bold')}
        >
          <Bold size={13} />
        </button>

        <button
          class="tool-action-btn"
          title="Italic (Ctrl+I)"
          onclick={() => applyFormat('italic')}
        >
          <Italic size={13} />
        </button>

        <button
          class="tool-action-btn"
          title="Underline (Ctrl+U)"
          onclick={() => applyFormat('underline')}
        >
          <Underline size={13} />
        </button>

        <button
          class="tool-action-btn"
          title="Strikethrough"
          onclick={() => applyFormat('strikeThrough')}
        >
          <Strikethrough size={13} />
        </button>

        <button
          class="tool-action-btn"
          title="Bullet List"
          onclick={() => applyFormat('insertUnorderedList')}
        >
          <ListOrdered size={13} />
        </button>

        <button
          class="tool-action-btn"
          title="Attach image"
          onclick={() => fileInputRef?.click()}
        >
          <ImageIcon size={13} />
        </button>
      </div>

      <div class="save-state-tag">
        {#if saveStatus === 'saving'}
          <span>Saving...</span>
        {:else if saveStatus === 'saved'}
          <span class="state-saved-text">Saved</span>
        {/if}
      </div>
    </div>

    <!-- Main Text Area (Using ScrollView and clearing bottom dock) -->
    <div class="notes-main-canvas">
      <ScrollView>
        <div
          bind:this={editorRef}
          class="editor-surface"
          contenteditable="true"
          oninput={triggerAutoSave}
          onpaste={handlePaste}
          onkeydown={handleKeyDown}
          role="textbox"
          tabindex="0"
        ></div>
      </ScrollView>
    </div>
  {/if}
</div>

<style src="./styles/NotesPage.css"></style>
