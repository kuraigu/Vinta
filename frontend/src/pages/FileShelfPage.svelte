<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 220, h: 36, r: 16 },
    expanded: { w: 460, h: 360, r: 28 }
  }
</script>

<script lang="ts">
  /**
   * @fileoverview Drag and drop file holding shelf.
   *
   * Serves as a temporary desktop holding area for files and folders dragged from
   * Windows Explorer or other apps, supporting batch removal and opening files in shell.
   */

  import { SvelteSet } from 'svelte/reactivity'
  import { Call } from '@wailsio/runtime'
  import {
    FolderDown,
    Plus,
    Trash2,
    Copy,
    X,
    Check,
    FileArchive,
    ShieldCheck,
    Image,
    FileText,
    Music,
    Film,
    FileCode,
    Folder,
    FolderOpen,
    File
  } from 'lucide-svelte'
  import BackButton from '../components/BackButton.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import { hideTooltip } from '../stores/tooltip'
  import { tStore } from '../store'
  import {
    fileShelfStore,
    removeFileFromShelf,
    removeFilesFromShelf,
    promptChooseShelfFiles,
    clearFileShelf,
    dragFilesOut,
    copyShelfPaths,
    bundleShelfToZip,
    stripShelfExif,
    revealShelfFileInExplorer,
    formatFileSize,
    type ShelfFileItem
  } from '../stores/fileShelf'

  let isDropActive = $state(false)
  let dragDepth = 0
  let selectedPaths = new SvelteSet<string>()

  /**
   * Synchronizes selected file paths set with the active shelf item list.
   */
  function syncSelectedPaths(): void {
    const valid = new Set($fileShelfStore.items.map((i) => i.path))
    for (const p of selectedPaths) {
      if (!valid.has(p)) {
        selectedPaths.delete(p)
      }
    }
  }

  $effect(syncSelectedPaths)

  /**
   * Extracts target filesystem paths from currently selected shelf tiles.
   *
   * @returns Array of absolute file path strings.
   */
  function computeTargetPaths(): string[] {
    if (selectedPaths.size > 0) {
      return $fileShelfStore.items
        .filter((i) => selectedPaths.has(i.path))
        .map((i) => i.path)
    }
    return $fileShelfStore.items.map((i) => i.path)
  }

  let targetPaths = $derived.by(computeTargetPaths)

  /**
   * Evaluates whether any selected files are image formats that support metadata stripping.
   *
   * @returns True if strippable image files are selected.
   */
  function computeHasStrippableImages(): boolean {
    const targetSet = new Set(targetPaths)
    return $fileShelfStore.items.some(
      (item) =>
        targetSet.has(item.path) &&
        (item.extension === '.jpg' ||
          item.extension === '.jpeg' ||
          item.extension === '.png')
    )
  }

  let hasStrippableImages = $derived.by(computeHasStrippableImages)

  /**
   * Toggles selection state for an individual file tile.
   *
   * @param path - Filesystem path of the item.
   */
  function toggleItemSelection(path: string): void {
    if (selectedPaths.has(path)) {
      selectedPaths.delete(path)
    } else {
      selectedPaths.add(path)
    }
  }

  /**
   * Selects all files currently on the shelf, or deselects all if already fully selected.
   */
  function toggleSelectAll(): void {
    if (selectedPaths.size === $fileShelfStore.items.length) {
      selectedPaths.clear()
    } else {
      for (const item of $fileShelfStore.items) {
        selectedPaths.add(item.path)
      }
    }
  }

  /**
   * Removes all currently selected items from the shelf.
   */
  function removeSelectedItems(): void {
    if (selectedPaths.size === 0) return
    const toRemove = Array.from(selectedPaths)
    selectedPaths.clear()
    removeFilesFromShelf(toRemove)
  }

  /**
   * Handles drag enter visual indicators over the drop target zone.
   *
   * @param e - Drag event.
   */
  function handleDragEnter(e: DragEvent): void {
    if (!e.dataTransfer?.types?.includes('Files')) return
    dragDepth += 1
    isDropActive = true
  }

  /**
   * Handles drag over events, verifying payload type and enabling copy drop effect.
   *
   * @param e - Drag event.
   */
  function handleDragOver(e: DragEvent): void {
    if (!e.dataTransfer?.types?.includes('Files')) return
    isDropActive = true
  }

  /**
   * Clears drop highlight state once dragged files leave the shelf region.
   *
   * @param e - Drag event.
   */
  function handleDragLeave(e: DragEvent): void {
    if (!e.dataTransfer?.types?.includes('Files')) return
    dragDepth = Math.max(0, dragDepth - 1)
    if (dragDepth === 0) {
      isDropActive = false
    }
  }

  /**
   * Processes dropped filesystem files, adding new unique paths to the shelf store.
   */
  function handleDrop(): void {
    dragDepth = 0
    isDropActive = false
  }

  /**
   * Resets active drag-and-drop state indicators.
   */
  function cancelActiveDrag(): void {
    fileShelfStore.update((s) => ({
      ...s,
      isDraggingOut: false,
      isOverShelf: false,
      draggingPaths: []
    }))
    void Call.ByName('Vinta/backend.App.CancelNativeFileDrag')
  }

  /**
   * Handles keyboard shortcuts on the file shelf (Delete, Ctrl+A).
   *
   * @param e - Keyboard event.
   */
  function handleWindowKeyDown(e: KeyboardEvent): void {
    if (e.key === 'Escape') {
      if ($fileShelfStore.isDraggingOut) {
        cancelActiveDrag()
      } else if (selectedPaths.size > 0) {
        selectedPaths.clear()
      }
    }
  }

  /**
   * Handles mouse down on a shelf tile:
   * Clicking toggles selection, while dragging starts authentic Windows Shell OLE drag
   * that drops directly into Windows File Explorer, Desktop, and conversation/chat boxes.
   */
  function handleTileMouseDown(e: MouseEvent, item: ShelfFileItem): void {
    if (e.button !== 0) return
    const targetEl = e.target as HTMLElement | null
    if (targetEl?.closest('button.tile-remove-badge')) return
    e.preventDefault()

    const dragPaths =
      selectedPaths.has(item.path) && selectedPaths.size > 0
        ? targetPaths
        : [item.path]

    const startX = e.clientX
    const startY = e.clientY
    let dragInitiated = false

    function onMouseMove(moveEv: MouseEvent) {
      if (dragInitiated) return
      const dist = Math.hypot(moveEv.clientX - startX, moveEv.clientY - startY)
      if (dist > 4) {
        dragInitiated = true
        cleanup()
        hideTooltip()
        void dragFilesOut(dragPaths)
      }
    }

    function onMouseUp() {
      if (!dragInitiated) {
        toggleItemSelection(item.path)
      }
      cleanup()
    }

    function cleanup() {
      window.removeEventListener('mousemove', onMouseMove)
      window.removeEventListener('mouseup', onMouseUp)
    }

    window.addEventListener('mousemove', onMouseMove)
    window.addEventListener('mouseup', onMouseUp)
  }

  /**
   * Formats hover tooltip text displaying filename, size, and system path.
   *
   * @param item - Shelf file record.
   * @returns Formatted tooltip text.
   */
  function formatTileTooltip(item: ShelfFileItem): string {
    const filename = item.name || 'Unknown'
    const modDate = item.modifiedTime
      ? new Date(item.modifiedTime)
      : new Date(item.addedAt)
    const dateFormatted = modDate.toLocaleString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })

    // Compute parent directory path for Location
    let location = item.path
    const lastSlash = Math.max(
      location.lastIndexOf('\\'),
      location.lastIndexOf('/')
    )
    if (lastSlash > 0) {
      location = location.slice(0, lastSlash)
    }

    return `${filename}\n${dateFormatted}\n${location}`
  }

  /**
   * Derives file extension badge text for tile previews (e.g. "PNG", "PDF").
   *
   * @param item - Shelf file record.
   * @returns Uppercase extension string.
   */
  function formatExtensionBadge(item: ShelfFileItem): string {
    if (item.isDir) return 'DIR'
    if (!item.extension) return 'FILE'
    return item.extension.replace(/^\./, '').toUpperCase()
  }
</script>

<svelte:window onkeydown={handleWindowKeyDown} />

<div
  class="file-shelf-page"
  class:is-drop-active={isDropActive}
  data-vinta="file-shelf-page"
  role="region"
  aria-label="File Holding Shelf"
  ondragenter={handleDragEnter}
  ondragover={handleDragOver}
  ondragleave={handleDragLeave}
  ondrop={handleDrop}
>
  {#if $fileShelfStore.items.length > 0}
    <div
      class="shelf-inline-bar"
      class:is-dragging-banner={$fileShelfStore.isDraggingOut}
      class:is-cancel-hover={$fileShelfStore.isDraggingOut &&
        $fileShelfStore.isOverShelf}
    >
      {#if $fileShelfStore.isDraggingOut}
        <button
          type="button"
          class="dragging-banner-btn no-swipe no-drag"
          class:is-cancel-hover={$fileShelfStore.isOverShelf}
          onclick={cancelActiveDrag}
          use:islandTooltip={$fileShelfStore.isOverShelf
            ? 'Release mouse to cancel drag'
            : 'Click or Esc to cancel drag'}
        >
          {#if $fileShelfStore.isOverShelf}
            <span class="cancel-banner-icon">
              <X size={11} strokeWidth={3} />
            </span>
            <span class="dragging-live-text cancel-text">
              Release to cancel drag
            </span>
          {:else}
            <span class="dragging-live-text">
              Dragging {$fileShelfStore.draggingPaths.length === 1
                ? '1 file'
                : `${$fileShelfStore.draggingPaths.length} files`} · Drag back here
              or Esc to cancel
            </span>
          {/if}
        </button>
      {:else}
        <button
          type="button"
          class="select-toggle-pill no-swipe no-drag"
          class:has-selection={selectedPaths.size > 0}
          use:islandTooltip={selectedPaths.size === $fileShelfStore.items.length
            ? 'Click to clear selection'
            : 'Click to select all'}
          onclick={toggleSelectAll}
        >
          <span class="select-check-dot">
            {#if selectedPaths.size > 0}
              <Check size={10} strokeWidth={3} />
            {/if}
          </span>
          <span>
            {selectedPaths.size > 0
              ? `${selectedPaths.size} chosen`
              : 'Select All'}
          </span>
        </button>

        <div class="inline-actions-row">
          {#if selectedPaths.size > 0}
            <button
              type="button"
              class="inline-pill-btn"
              disabled={$fileShelfStore.busyAction !== null}
              use:islandTooltip={'Copy chosen files to clipboard'}
              onclick={() => void copyShelfPaths(Array.from(selectedPaths))}
            >
              <Copy size={11} />
              <span>Copy</span>
            </button>

            {#if targetPaths.length > 0}
              <button
                type="button"
                class="inline-icon-btn"
                disabled={$fileShelfStore.busyAction !== null}
                use:islandTooltip={'Reveal chosen file in File Explorer'}
                onclick={() => {
                  if (targetPaths.length > 0) {
                    void revealShelfFileInExplorer(targetPaths[0])
                  }
                }}
              >
                <FolderOpen size={12} />
              </button>
            {/if}

            <button
              type="button"
              class="inline-icon-btn"
              disabled={$fileShelfStore.busyAction !== null}
              use:islandTooltip={'Bundle chosen files into a .zip'}
              onclick={() => void bundleShelfToZip(targetPaths)}
            >
              <FileArchive size={12} />
            </button>

            {#if hasStrippableImages}
              <button
                type="button"
                class="inline-icon-btn"
                disabled={$fileShelfStore.busyAction !== null}
                use:islandTooltip={'Strip EXIF metadata from JPEG/PNG images'}
                onclick={() => void stripShelfExif(targetPaths)}
              >
                <ShieldCheck size={12} />
              </button>
            {/if}

            <button
              type="button"
              class="inline-icon-btn danger-pill"
              use:islandTooltip={'Remove chosen files from shelf'}
              onclick={removeSelectedItems}
            >
              <X size={12} />
            </button>
          {:else}
            <button
              type="button"
              class="inline-pill-btn"
              disabled={$fileShelfStore.busyAction !== null}
              use:islandTooltip={'Bundle all files into a .zip'}
              onclick={() => void bundleShelfToZip(targetPaths)}
            >
              <FileArchive size={11} />
              <span>
                {$fileShelfStore.busyAction === 'zip'
                  ? 'Bundling...'
                  : 'Bundle ZIP'}
              </span>
            </button>

            <button
              type="button"
              class="inline-icon-btn danger-pill"
              use:islandTooltip={$tStore('fileshelf.clear_shelf')}
              onclick={clearFileShelf}
            >
              <Trash2 size={12} />
            </button>
          {/if}
        </div>
      {/if}
    </div>
  {/if}

  <ScrollView class="shelf-scroll-area" enableSwipe={false}>
    {#if $fileShelfStore.isDraggingOut && $fileShelfStore.isOverShelf}
      <div class="shelf-cancel-overlay" aria-live="polite">
        <div class="shelf-cancel-card">
          <div class="cancel-icon-ring">
            <X size={20} strokeWidth={2.5} />
          </div>
          <span class="cancel-card-title">Release to Cancel</span>
        </div>
      </div>
    {/if}

    {#if $fileShelfStore.items.length === 0}
      <div class="shelf-empty-state">
        <div class="empty-icon-wrap">
          <FolderDown size={22} />
        </div>
        <p class="empty-title">{$tStore('fileshelf.drop_prompt')}</p>
        <p class="empty-subtitle">
          Stage files from File Explorer, Desktop, or apps, then drag them out
          whenever you need them.
        </p>
        <button
          type="button"
          class="empty-choose-btn"
          onclick={() => void promptChooseShelfFiles()}
        >
          <Plus size={13} />
          <span>Choose Files</span>
        </button>
      </div>
    {:else}
      <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
      <div class="shelf-grid no-swipe no-drag" role="list">
        {#each $fileShelfStore.items as item (item.path)}
          {@const isSelected = selectedPaths.has(item.path)}
          {@const isBeingDragged =
            $fileShelfStore.isDraggingOut &&
            $fileShelfStore.draggingPaths.includes(item.path)}
          <div
            class="shelf-tile no-swipe no-drag category-{item.category}"
            class:is-selected={isSelected}
            class:is-dragging={isBeingDragged}
            role="listitem"
            use:islandTooltip={{ text: formatTileTooltip(item), align: 'left' }}
            onmousedown={(e) => handleTileMouseDown(e, item)}
            ondragstart={(e) => e.preventDefault()}
            ondblclick={() => void revealShelfFileInExplorer(item.path)}
          >
            <div class="tile-stage">
              {#if item.previewData}
                <img
                  class="tile-thumb-img"
                  class:is-native-icon={item.category !== 'image'}
                  src={item.previewData}
                  alt={item.name}
                  draggable="false"
                />
              {:else if item.category === 'image'}
                <Image size={24} />
              {:else if item.category === 'archive'}
                <FileArchive size={24} />
              {:else if item.category === 'document'}
                <FileText size={24} />
              {:else if item.category === 'audio'}
                <Music size={24} />
              {:else if item.category === 'video'}
                <Film size={24} />
              {:else if item.category === 'code'}
                <FileCode size={24} />
              {:else if item.category === 'folder'}
                <Folder size={24} />
              {:else}
                <File size={24} />
              {/if}

              {#if !item.previewData || item.category !== 'image'}
                <span class="tile-ext-tag">{formatExtensionBadge(item)}</span>
              {/if}

              <button
                type="button"
                class="tile-select-badge no-swipe no-drag"
                class:checked={isSelected}
                tabindex="-1"
                aria-hidden="true"
              >
                {#if isSelected}
                  <Check size={10} strokeWidth={3} />
                {/if}
              </button>

              <button
                type="button"
                class="tile-remove-badge no-swipe no-drag"
                use:islandTooltip={'Remove from shelf'}
                onpointerdown={(e) => e.stopPropagation()}
                onclick={(e) => {
                  e.stopPropagation()
                  selectedPaths.delete(item.path)
                  removeFileFromShelf(item.path)
                }}
              >
                <X size={10} strokeWidth={2.5} />
              </button>
            </div>

            <span class="tile-name">{item.name}</span>
            <span class="tile-meta"
              >{formatFileSize(item.size, item.isDir)}</span
            >
          </div>
        {/each}
      </div>
    {/if}
  </ScrollView>

  <div class="header">
    <div class="header-left">
      <BackButton />
      <span class="title">File Shelf</span>
      {#if $fileShelfStore.statusMessage}
        <span class="shelf-status-pill">{$fileShelfStore.statusMessage}</span>
      {:else if $fileShelfStore.items.length > 0}
        <span class="shelf-count-pill">{$fileShelfStore.items.length}</span>
      {/if}
    </div>

    <div class="header-right">
      <button
        type="button"
        class="header-icon-btn"
        use:islandTooltip={'Choose files or folders to add'}
        onclick={() => void promptChooseShelfFiles()}
      >
        <Plus size={14} />
      </button>

      {#if $fileShelfStore.items.length > 0}
        <button
          type="button"
          class="header-icon-btn clear-btn"
          use:islandTooltip={'Clear all staged files'}
          onclick={() => {
            selectedPaths.clear()
            clearFileShelf()
          }}
        >
          <Trash2 size={13} />
        </button>
      {/if}
    </div>
  </div>
</div>

<style src="./styles/FileShelfPage.css"></style>
