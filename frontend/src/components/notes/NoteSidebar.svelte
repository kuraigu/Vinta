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

<style>
  .notes-sidebar {
    display: flex;
    flex-direction: column;
    gap: 8px;
    height: 100%;
    min-height: 0;
  }

  .sidebar-header {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .new-note-btn {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    background: var(--c-primary, #6366f1);
    color: #ffffff;
    border: none;
    border-radius: 8px;
    padding: 6px 10px;
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .new-note-btn:hover {
    filter: brightness(1.1);
    transform: translateY(-1px);
  }

  .folder-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid var(--c-border);
    border-radius: 8px;
    color: var(--c-text-secondary);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .folder-btn:hover {
    background: rgba(255, 255, 255, 0.12);
    color: var(--c-text-primary);
  }

  .search-box {
    position: relative;
    display: flex;
    align-items: center;
  }

  .search-box :global(.search-icon) {
    position: absolute;
    left: 8px;
    color: var(--c-text-secondary);
    pointer-events: none;
  }

  .search-box input {
    width: 100%;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--c-border);
    border-radius: 8px;
    padding: 5px 8px 5px 26px;
    font-size: 11px;
    color: var(--c-text-primary);
    outline: none;
    box-sizing: border-box;
  }

  .search-box input:focus {
    border-color: var(--c-primary, #6366f1);
  }

  :global(.notes-list-scroll) {
    flex: 1;
    min-height: 0;
  }

  .notes-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding-right: 4px;
  }

  .empty-list {
    font-size: 11px;
    color: var(--c-text-secondary);
    text-align: center;
    padding: 20px 0;
  }

  .note-card {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 8px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid transparent;
    border-radius: 10px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .note-card:hover {
    background: rgba(255, 255, 255, 0.06);
  }

  .note-card.active {
    background: rgba(255, 255, 255, 0.09);
    border-color: var(--c-border);
  }

  .note-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .note-card-title {
    font-size: 12px;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--c-text-primary);
  }

  .delete-btn {
    opacity: 0;
    background: none;
    border: none;
    color: var(--c-text-secondary);
    cursor: pointer;
    padding: 2px;
    border-radius: 4px;
    transition: all 0.15s ease;
  }

  .note-card:hover .delete-btn {
    opacity: 0.7;
  }

  .delete-btn:hover {
    opacity: 1 !important;
    color: #ef4444;
  }

  .note-card-snippet {
    font-size: 10px;
    color: var(--c-text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .note-card-time {
    font-size: 9px;
    color: var(--c-text-secondary);
    opacity: 0.6;
    margin-top: 2px;
  }
</style>
