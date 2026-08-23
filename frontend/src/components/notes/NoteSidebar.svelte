<script lang="ts">
  import { Plus, Search, Trash2, FolderOpen } from 'lucide-svelte'
  import ScrollView from '../ScrollView.svelte'

  interface NoteItem {
    id: string
    title: string
    fileName: string
    content: string
    modTime: number
  }

  interface Props {
    notes: NoteItem[]
    selectedNoteId: string
    searchQuery: string
    onSelectNote: (id: string) => void
    onNewNote: () => void
    onDeleteNote: (e: Event, id: string) => void
    onOpenFolder: () => void
    formatTimeAgo: (timestamp: number) => string
  }

  let {
    notes,
    selectedNoteId,
    searchQuery = $bindable(''),
    onSelectNote,
    onNewNote,
    onDeleteNote,
    onOpenFolder,
    formatTimeAgo
  }: Props = $props()
</script>

<div class="notes-sidebar">
  <div class="sidebar-header">
    <button class="new-note-btn" onclick={onNewNote}>
      <Plus size={14} />
      <span>New Note</span>
    </button>
    <button
      class="folder-btn"
      title="Open Notes Folder in Explorer"
      onclick={onOpenFolder}
    >
      <FolderOpen size={14} />
    </button>
  </div>

  <div class="search-box">
    <Search size={13} class="search-icon" />
    <input type="text" placeholder="Search notes..." bind:value={searchQuery} />
  </div>

  <ScrollView class="notes-list-scroll">
    <div class="notes-list">
      {#if notes.length === 0}
        <div class="empty-list">No notes found</div>
      {:else}
        {#each notes as note (note.id)}
          <div
            class="note-card"
            class:active={selectedNoteId === note.id}
            onclick={() => onSelectNote(note.id)}
            role="button"
            tabindex="0"
            onkeydown={(e) => e.key === 'Enter' && onSelectNote(note.id)}
          >
            <div class="note-card-header">
              <span class="note-card-title"
                >{note.title || 'Untitled Note'}</span
              >
              <button
                class="delete-btn"
                title="Delete Note"
                onclick={(e) => onDeleteNote(e, note.id)}
              >
                <Trash2 size={12} />
              </button>
            </div>
            <div class="note-card-snippet">
              {(function cleanSnippet(str: string) {
                if (!str) return 'Empty note'
                const cleaned = str
                  .replace(/```[\s\S]*?```/g, '')
                  .replace(/^- \[[ xX]\] /gm, '')
                  .replace(/^[-*>] /gm, '')
                  .replace(/#{1,6}\s+/g, '')
                  .replace(/[*_~`]/g, '')
                  .replace(/\s+/g, ' ')
                  .trim()
                return cleaned.slice(0, 50) || 'Empty note'
              })(note.content)}
            </div>
            <div class="note-card-time">
              {formatTimeAgo(note.modTime)}
            </div>
          </div>
        {/each}
      {/if}
    </div>
  </ScrollView>
</div>

<style src="../styles/notes/NoteSidebar.css"></style>
