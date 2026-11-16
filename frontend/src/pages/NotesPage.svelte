<!--
  @fileoverview Desktop scratchpad and quick notes manager.

  Provides a color-coded sticky note board and rich text editor with auto-save,
  markdown shortcuts, and inline image pasting.
-->
<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 },
    expanded: { w: 600, h: 480, r: 28 },
    borderColor: 'var(--vinta-border)',
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
    order?: number
    createdAt: number
    modTime: number
  }
</script>

<script lang="ts">
  import { onMount, tick } from 'svelte'
  import { SvelteSet } from 'svelte/reactivity'
  import {
    Plus,
    FolderOpen,
    Trash2,
    List,
    Bold,
    Italic,
    Underline,
    Strikethrough,
    Pin,
    Pencil,
    Copy,
    CopyPlus,
    Image as ImageIcon,
    Check,
    X
  } from 'lucide-svelte'
  import BackButton from '../components/BackButton.svelte'
  import SearchInput from '../components/SearchInput.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import CustomContextMenu, {
    type ContextMenuItem
  } from '../components/CustomContextMenu.svelte'
  import { buildTextEditMenuItems } from '../utils/textContextMenu'
  import { islandTooltip } from '../services/islandTooltip'
  import { tStore } from '../store'
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
  let drawerSearchQuery = $state('')
  let isDraggingOver = $state(false)
  let saveStatus = $state<'saved' | 'saving' | ''>('')
  let saveTimer: ReturnType<typeof setTimeout> | null = null
  let renameTimer: ReturnType<typeof setTimeout> | null = null

  let editorRef = $state<HTMLDivElement | null>(null)
  let titleInputRef = $state<HTMLInputElement | null>(null)
  let fileInputRef = $state<HTMLInputElement | null>(null)
  let noteCtxMenu = $state<{ x: number; y: number; index: number } | null>(null)
  let textCtxMenu = $state<{
    x: number
    y: number
    items: ContextMenuItem[]
  } | null>(null)
  let selectedNoteIds = new SvelteSet<string>()
  let draggedNoteId = $state<string | null>(null)
  let dragOverNoteId = $state<string | null>(null)

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

  function compareNotes(a: VNote, b: VNote): number {
    if (Boolean(a.pinned) !== Boolean(b.pinned)) {
      return a.pinned ? -1 : 1
    }
    const orderA =
      typeof a.order === 'number' ? a.order : Number.MAX_SAFE_INTEGER
    const orderB =
      typeof b.order === 'number' ? b.order : Number.MAX_SAFE_INTEGER
    if (orderA !== orderB) {
      return orderA - orderB
    }
    return (b.modTime || b.createdAt) - (a.modTime || a.createdAt)
  }

  let filteredDrawerNotes = $derived(
    notes
      .filter((n) => {
        const q = drawerSearchQuery.toLowerCase().trim()
        if (!q) return true
        const snippet = n.content.replace(/<[^>]*>/g, '').toLowerCase()
        return n.title.toLowerCase().includes(q) || snippet.includes(q)
      })
      .sort(compareNotes)
  )

  let allSelectedPinned = $derived.by(() => {
    if (selectedNoteIds.size === 0) return false
    const selected = notes.filter((n) => selectedNoteIds.has(n.id))
    return selected.length > 0 && selected.every((n) => Boolean(n.pinned))
  })

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
                  order:
                    typeof parsed.order === 'number' ? parsed.order : undefined,
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

        loaded.sort(compareNotes)
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
    const baseUntitled = $tStore('notes.untitled')
    let newTitle = baseUntitled
    if (existingTitles.has(baseUntitled.toLowerCase())) {
      let counter = 2
      while (existingTitles.has(`${baseUntitled.toLowerCase()} (${counter})`)) {
        counter++
      }
      newTitle = `${baseUntitled} (${counter})`
    }

    const nextColor = colorPaletteCycle[notes.length % colorPaletteCycle.length]
    const newFileName = `${newTitle}.vnote`
    const now = Date.now()

    const nextOrder =
      notes
        .filter((n) => typeof n.order === 'number')
        .reduce((min, n) => Math.min(min, n.order!), 0) - 1

    const newNote: VNote = {
      id: newTitle,
      title: newTitle,
      fileName: newFileName,
      color: nextColor,
      content: '',
      pinned: false,
      order: nextOrder,
      createdAt: now,
      modTime: now
    }

    notes = [newNote, ...notes]
    activeIndex = 0
    viewMode = 'editor'
    persistNote(newNote)
    syncEditorContent()
  }

  /**
   * Persists the given note document to disk as formatted JSON.
   *
   * @param note - Target note object.
   * @param touchModTime - Whether to bump modTime to current epoch.
   */
  function persistNote(note: VNote, touchModTime = true) {
    if (touchModTime) {
      note.modTime = Date.now()
    }
    const payload = {
      id: note.id,
      title: note.title,
      fileName: note.fileName,
      color: note.color || 'yellow',
      content: note.content || '',
      pinned: Boolean(note.pinned),
      ...(typeof note.order === 'number' ? { order: note.order } : {}),
      createdAt: note.createdAt || Date.now(),
      modTime: note.modTime || Date.now()
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
      const cleanTitle = activeNote.title.trim() || $tStore('notes.untitled')
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
    if (
      draggedNoteId ||
      viewMode !== 'editor' ||
      !e.dataTransfer?.types?.includes('Files')
    ) {
      return
    }
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
    if (draggedNoteId || viewMode !== 'editor') {
      isDraggingOver = false
      return
    }
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
   * Toggles the pinned status of a note directly from the card board.
   *
   * @param e - Mouse event.
   * @param idx - Index of note to pin or unpin.
   */
  function togglePinByIndex(e: MouseEvent, idx: number) {
    e.stopPropagation()
    const target = notes[idx]
    if (!target) return
    target.pinned = !target.pinned
    persistNote(target)
  }

  /**
   * Deletes a note directly from the card board.
   *
   * @param e - Mouse event.
   * @param idx - Index of note to remove.
   */
  async function deleteNoteByIndex(e: MouseEvent | null, idx: number) {
    if (e) e.stopPropagation()
    const target = notes[idx]
    if (!target) return
    selectedNoteIds.delete(target.id)
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
   * Toggles selection state of a note card.
   */
  function toggleNoteSelection(noteId: string) {
    if (selectedNoteIds.has(noteId)) {
      selectedNoteIds.delete(noteId)
    } else {
      selectedNoteIds.add(noteId)
    }
  }

  function clearSelection() {
    selectedNoteIds.clear()
  }

  function bulkTogglePin() {
    if (selectedNoteIds.size === 0) return
    const nextPinned = !allSelectedPinned
    for (const note of notes) {
      if (selectedNoteIds.has(note.id)) {
        note.pinned = nextPinned
        persistNote(note)
      }
    }
  }

  function bulkChangeColor(color: StickyColor) {
    if (selectedNoteIds.size === 0) return
    for (const note of notes) {
      if (selectedNoteIds.has(note.id)) {
        note.color = color
        persistNote(note)
      }
    }
  }

  async function bulkDeleteSelected() {
    if (selectedNoteIds.size === 0) return
    const targets = notes.filter((n) => selectedNoteIds.has(n.id))
    notes = notes.filter((n) => !selectedNoteIds.has(n.id))
    selectedNoteIds.clear()
    if (activeIndex >= notes.length) {
      activeIndex = Math.max(0, notes.length - 1)
    }
    await Promise.allSettled(targets.map((t) => DeleteNote(t.fileName)))
  }

  function handleCardClick(e: MouseEvent, noteId: string, idx: number) {
    if (e.ctrlKey || e.metaKey || selectedNoteIds.size > 0) {
      e.preventDefault()
      toggleNoteSelection(noteId)
      return
    }
    openNoteEditor(idx)
  }

  function handleNoteDragStart(e: DragEvent, noteId: string) {
    draggedNoteId = noteId
    dragOverNoteId = null
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move'
      e.dataTransfer.setData('text/plain', noteId)
    }
  }

  function handleNoteDragOver(e: DragEvent, targetId: string) {
    if (!draggedNoteId || draggedNoteId === targetId) return
    const src = notes.find((n) => n.id === draggedNoteId)
    const dst = notes.find((n) => n.id === targetId)
    if (!src || !dst || Boolean(src.pinned) !== Boolean(dst.pinned)) return
    e.preventDefault()
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move'
    }
    dragOverNoteId = targetId
  }

  function handleNoteDragLeave(targetId: string) {
    if (dragOverNoteId === targetId) {
      dragOverNoteId = null
    }
  }

  function handleNoteDrop(e: DragEvent, targetId: string) {
    e.preventDefault()
    e.stopPropagation()
    const sourceId = draggedNoteId
    draggedNoteId = null
    dragOverNoteId = null
    if (!sourceId || sourceId === targetId) return

    const sorted = [...notes].sort(compareNotes)
    const srcNote = sorted.find((n) => n.id === sourceId)
    const dstNote = sorted.find((n) => n.id === targetId)
    if (
      !srcNote ||
      !dstNote ||
      Boolean(srcNote.pinned) !== Boolean(dstNote.pinned)
    ) {
      return
    }

    const srcIdx = sorted.findIndex((n) => n.id === sourceId)
    const [moved] = sorted.splice(srcIdx, 1)
    const dstIdx = sorted.findIndex((n) => n.id === targetId)
    sorted.splice(dstIdx, 0, moved)

    sorted.forEach((n, index) => {
      if (n.order !== index) {
        n.order = index
        persistNote(n, false)
      }
    })
    notes = sorted
  }

  function handleNoteDragEnd() {
    draggedNoteId = null
    dragOverNoteId = null
  }

  /**
   * Opens a note in the editor and focuses its title field for renaming.
   */
  async function renameNoteByIndex(idx: number) {
    if (!notes[idx]) return
    openNoteEditor(idx)
    await tick()
    titleInputRef?.focus()
    titleInputRef?.select()
  }

  /**
   * Duplicates a note with a unique copy title and persists it to disk.
   */
  function duplicateNoteByIndex(idx: number) {
    const source = notes[idx]
    if (!source) return
    const existingTitles = new Set(notes.map((n) => n.title.toLowerCase()))
    const copyWord = $tStore('notes.copy_suffix')
    const baseName = `${source.title || $tStore('notes.untitled')} (${copyWord})`
    let copyTitle = baseName
    let counter = 2
    while (existingTitles.has(copyTitle.toLowerCase())) {
      copyTitle = `${source.title || $tStore('notes.untitled')} (${copyWord} ${counter})`
      counter++
    }
    const now = Date.now()
    const nextOrder =
      notes
        .filter((n) => typeof n.order === 'number')
        .reduce((min, n) => Math.min(min, n.order!), 0) - 1
    const duplicated: VNote = {
      id: copyTitle,
      title: copyTitle,
      fileName: `${copyTitle}.vnote`,
      color: source.color || 'yellow',
      content: source.content || '',
      pinned: false,
      order: nextOrder,
      createdAt: now,
      modTime: now
    }
    notes = [duplicated, ...notes]
    persistNote(duplicated)
  }

  /**
   * Copies plain text content of a note to the system clipboard.
   */
  async function copyNoteContentByIndex(idx: number) {
    const target = notes[idx]
    if (!target) return
    const tmp = document.createElement('div')
    tmp.innerHTML = target.content || ''
    const plain = (tmp.innerText || tmp.textContent || '').trim()
    if (!plain && !target.title) return
    try {
      await navigator.clipboard.writeText(plain || target.title)
    } catch {
      // Clipboard write unavailable
    }
  }

  function openNoteContextMenu(e: MouseEvent, idx: number) {
    e.preventDefault()
    e.stopPropagation()
    textCtxMenu = null
    noteCtxMenu = { x: e.clientX, y: e.clientY, index: idx }
  }

  function openTextContextMenu(e: MouseEvent, target: HTMLElement | null) {
    e.preventDefault()
    e.stopPropagation()
    noteCtxMenu = null
    textCtxMenu = {
      x: e.clientX,
      y: e.clientY,
      items: buildTextEditMenuItems(target, {
        cut: $tStore('context_menu.cut'),
        copy: $tStore('context_menu.copy'),
        paste: $tStore('context_menu.paste'),
        selectAll: $tStore('context_menu.select_all')
      })
    }
  }

  function getNoteContextMenuItems(): ContextMenuItem[] {
    if (!noteCtxMenu) return []
    const idx = noteCtxMenu.index
    const target = notes[idx]
    if (!target) return []
    return [
      {
        label: selectedNoteIds.has(target.id)
          ? $tStore('notes.deselect')
          : $tStore('notes.select'),
        icon: selectedNoteIds.has(target.id) ? X : Check,
        action: () => toggleNoteSelection(target.id)
      },
      {
        label: target.pinned ? $tStore('notes.unpin') : $tStore('notes.pin'),
        icon: Pin,
        action: () => {
          target.pinned = !target.pinned
          persistNote(target)
        }
      },
      {
        label: $tStore('notes.rename'),
        icon: Pencil,
        action: () => void renameNoteByIndex(idx)
      },
      {
        label: $tStore('notes.duplicate'),
        icon: CopyPlus,
        action: () => duplicateNoteByIndex(idx)
      },
      {
        label: $tStore('notes.copy'),
        icon: Copy,
        action: () => void copyNoteContentByIndex(idx)
      },
      { label: '', action: () => {}, divider: true },
      {
        label: $tStore('notes.delete'),
        icon: Trash2,
        danger: true,
        action: () => void deleteNoteByIndex(null, idx)
      }
    ]
  }

  /**
   * Extracts a plaintext snippet from HTML content.
   *
   * @param html - Note content string.
   * @returns Clean preview excerpt.
   */
  function extractSnippet(html: string): string {
    if (!html) return $tStore('notes.empty_note')
    const text = html
      .replace(/<[^>]*>/g, ' ')
      .replace(/&nbsp;/g, ' ')
      .replace(/\s+/g, ' ')
      .trim()
    return text || $tStore('notes.empty_note')
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
    if (diff < 60) return $tStore('notes.just_now')
    if (diff < 3600) {
      return $tStore('notes.mins_ago', { count: Math.floor(diff / 60) })
    }
    if (diff < 86400) {
      return $tStore('notes.hours_ago', { count: Math.floor(diff / 3600) })
    }
    return $tStore('notes.days_ago', { count: Math.floor(diff / 86400) })
  }

  onMount(() => {
    loadNotesFromBackend()
  })
</script>

<div
  data-vinta="notes-page"
  class="notes-page-container {viewMode === 'editor'
    ? `color-${activeNote?.color || 'yellow'}`
    : ''}"
  ondragover={handleDragOver}
  ondragleave={handleDragLeave}
  ondrop={handleDrop}
  role="region"
  aria-label={$tStore('notes.title')}
>
  {#if isDraggingOver}
    <div class="drag-drop-overlay">
      <span>{$tStore('notes.drop_image')}</span>
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
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="notes-board-view"
      onkeydown={(e) => {
        if (e.key === 'Escape' && selectedNoteIds.size > 0) {
          e.stopPropagation()
          clearSelection()
        }
      }}
    >
      <!-- Notes Grid inside ScrollView -->
      <div class="board-grid-wrapper">
        <ScrollView>
          {#if filteredDrawerNotes.length === 0}
            <div class="empty-board-state">
              {#if drawerSearchQuery}
                <p>
                  {$tStore('notes.no_matching_notes', {
                    query: drawerSearchQuery
                  })}
                </p>
              {:else}
                <p>{$tStore('notes.no_notes_yet')}</p>
                <button class="board-create-btn" onclick={createNewNote}>
                  <Plus size={14} />
                  {$tStore('notes.new_note')}
                </button>
              {/if}
            </div>
          {:else}
            <div class="notes-cards-grid">
              {#each filteredDrawerNotes as n, idx (n.id)}
                {@const realIndex = notes.findIndex((item) => item.id === n.id)}
                {@const targetIndex = realIndex >= 0 ? realIndex : idx}
                <div
                  class="note-card card-{n.color || 'yellow'}"
                  class:is-selected={selectedNoteIds.has(n.id)}
                  class:is-drag-target={dragOverNoteId === n.id}
                  class:is-dragging={draggedNoteId === n.id}
                  draggable="true"
                  onclick={(e) => handleCardClick(e, n.id, targetIndex)}
                  oncontextmenu={(e) => openNoteContextMenu(e, targetIndex)}
                  ondragstart={(e) => handleNoteDragStart(e, n.id)}
                  ondragover={(e) => handleNoteDragOver(e, n.id)}
                  ondragleave={() => handleNoteDragLeave(n.id)}
                  ondrop={(e) => handleNoteDrop(e, n.id)}
                  ondragend={handleNoteDragEnd}
                  role="button"
                  tabindex="0"
                  onkeydown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      if (selectedNoteIds.size > 0) {
                        e.preventDefault()
                        toggleNoteSelection(n.id)
                      } else {
                        openNoteEditor(targetIndex)
                      }
                    }
                  }}
                >
                  <div class="note-card-band"></div>
                  <button
                    type="button"
                    class="note-card-pin-btn"
                    class:is-pinned={n.pinned}
                    use:islandTooltip={n.pinned
                      ? $tStore('notes.unpin')
                      : $tStore('notes.pin')}
                    onclick={(e) => togglePinByIndex(e, targetIndex)}
                  >
                    <Pin size={14} />
                  </button>
                  <button
                    type="button"
                    class="note-card-delete-btn"
                    use:islandTooltip={$tStore('notes.delete')}
                    onclick={(e) => deleteNoteByIndex(e, targetIndex)}
                  >
                    <Trash2 size={14} />
                  </button>

                  <div class="note-card-body">
                    <span class="note-card-title"
                      >{n.title || $tStore('notes.untitled')}</span
                    >
                    <p class="note-card-snippet">{extractSnippet(n.content)}</p>
                  </div>

                  <div class="note-card-footer">
                    <span class="note-card-date"
                      >{formatTimeAgo(n.modTime || n.createdAt)}</span
                    >
                    {#if hasImage(n.content)}
                      <span
                        class="note-card-has-img"
                        use:islandTooltip={$tStore('notes.contains_image')}
                      >
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

      {#if selectedNoteIds.size > 0}
        <div class="notes-bulk-bar">
          <span class="bulk-count"
            >{$tStore('notes.selected_count', {
              count: selectedNoteIds.size
            })}</span
          >
          <div class="bulk-colors">
            {#each colorPaletteCycle as c (c)}
              <button
                type="button"
                class="bulk-color-dot card-{c}"
                onclick={() => bulkChangeColor(c)}
                use:islandTooltip={$tStore('notes.color_label', { color: c })}
                aria-label={$tStore('notes.color_label', { color: c })}
              ></button>
            {/each}
          </div>
          <div class="bulk-actions">
            <button
              type="button"
              class="board-action-btn"
              onclick={bulkTogglePin}
              use:islandTooltip={{
                text: allSelectedPinned
                  ? $tStore('notes.unpin')
                  : $tStore('notes.pin'),
                position: 'top'
              }}
              aria-label={allSelectedPinned
                ? $tStore('notes.unpin')
                : $tStore('notes.pin')}
            >
              <Pin size={14} />
            </button>
            <button
              type="button"
              class="board-action-btn bulk-delete-btn"
              onclick={bulkDeleteSelected}
              use:islandTooltip={{
                text: $tStore('notes.delete'),
                position: 'top'
              }}
              aria-label={$tStore('notes.delete')}
            >
              <Trash2 size={14} />
            </button>
            <button
              type="button"
              class="board-action-btn"
              onclick={clearSelection}
              use:islandTooltip={{
                text: $tStore('notes.clear_selection'),
                position: 'top'
              }}
              aria-label={$tStore('notes.clear_selection')}
            >
              <X size={14} />
            </button>
          </div>
        </div>
      {/if}

      <div class="board-header">
        <div class="board-header-left">
          <BackButton title={$tStore('notes.back_to_apps')} />
          <span class="board-title">{$tStore('notes.title')}</span>
          {#if notes.length > 0}
            <span class="board-count-pill">{notes.length}</span>
          {/if}
        </div>

        <SearchInput
          bind:value={drawerSearchQuery}
          placeholder={$tStore('notes.search_placeholder')}
          maxWidth="220px"
        />

        <div class="board-header-right">
          <button
            class="board-action-btn"
            use:islandTooltip={{
              text: $tStore('notes.open_folder'),
              position: 'top'
            }}
            onclick={handleOpenFolder}
          >
            <FolderOpen size={15} />
          </button>
          <button
            class="board-action-btn new-note-action"
            use:islandTooltip={{
              text: $tStore('notes.new_note'),
              position: 'top'
            }}
            onclick={createNewNote}
          >
            <Plus size={16} />
          </button>
        </div>
      </div>
    </div>
  {:else}
    <!-- Active Note Editor (Sticky canvas with bottom toolbar & banner header) -->
    <div class="notes-main-canvas">
      <ScrollView>
        <div
          bind:this={editorRef}
          class="editor-surface"
          data-placeholder={$tStore('notes.editor_placeholder')}
          contenteditable="true"
          oninput={triggerAutoSave}
          onpaste={handlePaste}
          onkeydown={handleKeyDown}
          oncontextmenu={(e) => openTextContextMenu(e, editorRef)}
          role="textbox"
          tabindex="0"
        ></div>
      </ScrollView>
    </div>

    <!-- Compact Formatting Toolbar -->
    <div class="notes-toolbar">
      <div class="toolbar-actions-group">
        <button
          class="tool-action-btn"
          use:islandTooltip={{
            text: $tStore('notes.format_bold'),
            position: 'top'
          }}
          onclick={() => applyFormat('bold')}
        >
          <Bold size={14} />
        </button>

        <button
          class="tool-action-btn"
          use:islandTooltip={{
            text: $tStore('notes.format_italic'),
            position: 'top'
          }}
          onclick={() => applyFormat('italic')}
        >
          <Italic size={14} />
        </button>

        <button
          class="tool-action-btn"
          use:islandTooltip={{
            text: $tStore('notes.format_underline'),
            position: 'top'
          }}
          onclick={() => applyFormat('underline')}
        >
          <Underline size={14} />
        </button>

        <button
          class="tool-action-btn"
          use:islandTooltip={{
            text: $tStore('notes.format_strike'),
            position: 'top'
          }}
          onclick={() => applyFormat('strikeThrough')}
        >
          <Strikethrough size={14} />
        </button>

        <button
          class="tool-action-btn"
          use:islandTooltip={{
            text: $tStore('notes.format_list'),
            position: 'top'
          }}
          onclick={() => applyFormat('insertUnorderedList')}
        >
          <List size={14} />
        </button>

        <button
          class="tool-action-btn"
          use:islandTooltip={{
            text: $tStore('notes.format_image'),
            position: 'top'
          }}
          onclick={() => fileInputRef?.click()}
        >
          <ImageIcon size={14} />
        </button>
      </div>

      <div class="save-state-tag">
        {#if saveStatus === 'saving'}
          <span>{$tStore('notes.saving')}</span>
        {:else if saveStatus === 'saved'}
          <span class="state-saved-text">{$tStore('notes.saved')}</span>
        {/if}
      </div>
    </div>

    <!-- Subtle Banner Header (Positioned at bottom, swapped with top Dock) -->
    <div class="banner-header">
      <div class="banner-left-group">
        <BackButton
          onback={handleBackToBoard}
          title={$tStore('notes.back_to_notes')}
        />

        <button
          class="banner-btn"
          use:islandTooltip={{
            text: $tStore('notes.all_notes'),
            position: 'top'
          }}
          onclick={handleBackToBoard}
        >
          <List size={15} />
        </button>

        <button
          class="banner-btn"
          use:islandTooltip={{
            text: $tStore('notes.new_note'),
            position: 'top'
          }}
          onclick={createNewNote}
        >
          <Plus size={16} />
        </button>
      </div>

      <!-- Center: Note Title right in the subtle banner -->
      <div class="banner-center-title">
        {#if activeNote}
          <input
            bind:this={titleInputRef}
            type="text"
            class="banner-title-input"
            bind:value={activeNote.title}
            oninput={handleTitleChange}
            oncontextmenu={(e) => openTextContextMenu(e, titleInputRef)}
            placeholder={$tStore('notes.title_placeholder')}
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
              use:islandTooltip={{
                text: $tStore('notes.color_label', { color: c }),
                position: 'top'
              }}
              aria-label={$tStore('notes.color_label', { color: c })}
              onclick={() => handleSelectColor(c)}
            ></button>
          {/each}
        </div>

        <button
          class="banner-btn"
          use:islandTooltip={{
            text: $tStore('notes.open_folder'),
            position: 'top'
          }}
          onclick={handleOpenFolder}
        >
          <FolderOpen size={15} />
        </button>

        <button
          class="banner-btn"
          use:islandTooltip={{ text: $tStore('notes.delete'), position: 'top' }}
          onclick={handleDeleteActiveNote}
        >
          <Trash2 size={15} />
        </button>
      </div>
    </div>
  {/if}

  {#if noteCtxMenu}
    <CustomContextMenu
      x={noteCtxMenu.x}
      y={noteCtxMenu.y}
      items={getNoteContextMenuItems()}
      onclose={() => (noteCtxMenu = null)}
    />
  {/if}

  {#if textCtxMenu}
    <CustomContextMenu
      x={textCtxMenu.x}
      y={textCtxMenu.y}
      items={textCtxMenu.items}
      onclose={() => (textCtxMenu = null)}
    />
  {/if}
</div>

<style src="./styles/NotesPage.css"></style>
