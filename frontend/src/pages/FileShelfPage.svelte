<script lang="ts" module>
  import { ISLAND_WIDTH_TIERS, type IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 },
    expanded: { w: ISLAND_WIDTH_TIERS.standard, h: 360, r: 28 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  /**
   * @fileoverview Drag and drop file holding shelf.
   *
   * Serves as a temporary desktop holding area for files and folders dragged from
   * Windows Explorer or other apps, supporting batch removal and opening files in shell.
   */

  import { onMount } from 'svelte'
  import { Events } from '$lib/events'
  import { CancelNativeFileDrag } from '../../bindings/Vinta/backend/app'
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
  import PageHeader from '../components/PageHeader.svelte'
  import EmptyState from '../components/EmptyState.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import { hideTooltip } from '../stores/tooltip'
  import { activeOverlayStore, tStore } from '../store'
  import { isEditableEventTarget } from '../utils/keyboardGuard'
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

  let { expanded = false }: { expanded?: boolean } = $props()

  let isDropActive = $state(false)
  let dragDepth = 0
  let rawSelectedPaths = $state<string[]>([])

  /**
   * Derives active selected file paths by filtering against current shelf items.
   */
  function computeSelectedPaths(): string[] {
    const items = $fileShelfStore?.items ?? []
    const valid = new Set(items.map((i) => i.path))
    return rawSelectedPaths.filter((p) => valid.has(p))
  }

  let selectedPaths = $derived.by(computeSelectedPaths)

  Events.On('devtest-fileshelf-select-all', () => {
    const items = $fileShelfStore?.items ?? []
    rawSelectedPaths = items.map((i) => i.path)
  })

  /**
   * Extracts target filesystem paths from currently selected shelf tiles.
   *
   * @returns Array of absolute file path strings.
   */
  function computeTargetPaths(): string[] {
    const items = $fileShelfStore?.items ?? []
    if (selectedPaths.length > 0) {
      const selected = items.filter((i) => selectedPaths.includes(i.path))
      if (selected.length > 0) {
        return selected.map((i) => i.path)
      }
    }
    return items.map((i) => i.path)
  }

  let targetPaths = $derived.by(computeTargetPaths)

  /**
   * Evaluates whether any selected files are image formats that support metadata stripping.
   *
   * @returns True if strippable image files are selected.
   */
  function computeHasStrippableImages(): boolean {
    const targetSet = new Set(targetPaths)
    const items = $fileShelfStore?.items ?? []
    return items.some(
      (item) =>
        targetSet.has(item.path) &&
        (item.extension?.toLowerCase().replace(/^\./, '') === 'jpg' ||
          item.extension?.toLowerCase().replace(/^\./, '') === 'jpeg' ||
          item.extension?.toLowerCase().replace(/^\./, '') === 'png')
    )
  }

  let hasStrippableImages = $derived.by(computeHasStrippableImages)

  /**
   * Toggles selection state for an individual file tile.
   *
   * @param path - Filesystem path of the item.
   */
  function toggleItemSelection(path: string): void {
    if (selectedPaths.includes(path)) {
      rawSelectedPaths = selectedPaths.filter((p) => p !== path)
    } else {
      rawSelectedPaths = [...selectedPaths, path]
    }
  }

  /**
   * Selects all files currently on the shelf, or deselects all if already fully selected.
   */
  function toggleSelectAll(): void {
    if (selectedPaths.length === $fileShelfStore.items.length) {
      rawSelectedPaths = []
    } else {
      rawSelectedPaths = $fileShelfStore.items.map((item) => item.path)
    }
  }

  /**
   * Removes all currently selected items from the shelf.
   */
  function removeSelectedItems(): void {
    if (selectedPaths.length === 0) return
    const toRemove = [...selectedPaths]
    rawSelectedPaths = []
    removeFilesFromShelf(toRemove)
  }

  /**
   * Handles drag enter visual indicators over the drop target zone.
   *
   * @param e - Drag event.
   */
  function handleDragEnter(e: DragEvent): void {
    if (!e.dataTransfer?.types?.includes('Files')) return
    e.preventDefault()
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
    e.preventDefault()
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'copy'
    }
    isDropActive = true
  }

  /**
   * Clears drop highlight state once dragged files leave the shelf region.
   *
   * @param e - Drag event.
   */
  function handleDragLeave(e: DragEvent): void {
    if (!e.dataTransfer?.types?.includes('Files')) return
    e.preventDefault()
    dragDepth = Math.max(0, dragDepth - 1)
    if (dragDepth === 0) {
      isDropActive = false
    }
  }

  /**
   * Resets drop highlight state when files are dropped onto the shelf.
   * Wails native WindowFilesDropped handles the actual filesystem paths via `vinta-files-dropped`.
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
    void CancelNativeFileDrag()
  }

  /**
   * Handles keyboard shortcuts on the file shelf (Delete, Ctrl+A).
   *
   * @param e - Keyboard event.
   */
  function handleWindowKeyDown(e: KeyboardEvent): void {
    if (!expanded || $activeOverlayStore !== null) return
    if (isEditableEventTarget(e.target)) return
    if (e.key === 'Escape') {
      if ($fileShelfStore.isDraggingOut) {
        cancelActiveDrag()
      } else if (selectedPaths.length > 0) {
        rawSelectedPaths = []
      }
      return
    }
    if ((e.ctrlKey || e.metaKey) && (e.key === 'a' || e.key === 'A')) {
      if ($fileShelfStore.items.length > 0) {
        e.preventDefault()
        rawSelectedPaths = $fileShelfStore.items.map((item) => item.path)
      }
      return
    }
    if (e.key === 'Delete' || e.key === 'Backspace') {
      if (selectedPaths.length > 0) {
        e.preventDefault()
        removeSelectedItems()
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
      selectedPaths.includes(item.path) && selectedPaths.length > 0
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
          data-action="select-all"
          class:has-selection={selectedPaths.length > 0}
          onclick={toggleSelectAll}
        >
          <span class="select-check-dot">
            {#if selectedPaths.length > 0}
              <Check size={10} strokeWidth={3} />
            {/if}
          </span>
          <span>
            {selectedPaths.length > 0
              ? `${selectedPaths.length} chosen`
              : 'Select All'}
          </span>
        </button>

        <div class="inline-actions-row">
          {#if selectedPaths.length > 0}
            <button
              type="button"
              class="inline-pill-btn"
              disabled={$fileShelfStore.busyAction !== null}
              onclick={() => void copyShelfPaths([...selectedPaths])}
            >
              <Copy size={11} />
              <span>Copy</span>
            </button>

            {#if targetPaths.length > 0}
              <button
                type="button"
                class="inline-icon-btn"
                disabled={$fileShelfStore.busyAction !== null}
                use:islandTooltip={'Reveal in Explorer'}
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
              use:islandTooltip={'Bundle to ZIP'}
              onclick={() => void bundleShelfToZip(targetPaths)}
            >
              <FileArchive size={12} />
            </button>

            {#if hasStrippableImages}
              <button
                type="button"
                class="inline-icon-btn"
                disabled={$fileShelfStore.busyAction !== null}
                use:islandTooltip={'Strip EXIF metadata'}
                onclick={() => void stripShelfExif(targetPaths)}
              >
                <ShieldCheck size={12} />
              </button>
            {/if}

            <button
              type="button"
              class="inline-icon-btn danger-pill"
              use:islandTooltip={'Remove selected'}
              onclick={removeSelectedItems}
            >
              <X size={12} />
            </button>
          {:else}
            <button
              type="button"
              class="inline-pill-btn"
              disabled={$fileShelfStore.busyAction !== null}
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
      <EmptyState
        seamless={true}
        icon={FolderDown}
        title={$tStore('fileshelf.drop_prompt')}
        description={$tStore('fileshelf.empty_subtitle')}
        actionLabel={$tStore('fileshelf.choose_files')}
        onAction={() => void promptChooseShelfFiles()}
      />
    {:else}
      <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
      <div class="shelf-grid no-swipe no-drag" role="list">
        {#each $fileShelfStore.items as item (item.path)}
          {@const isSelected = selectedPaths.includes(item.path)}
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
                  rawSelectedPaths = selectedPaths.filter(
                    (p) => p !== item.path
                  )
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

  <PageHeader position="bottom" title={$tStore('fileshelf.title')}>
    {#snippet center()}
      {#if $fileShelfStore.statusMessage}
        <span class="shelf-status-pill">{$fileShelfStore.statusMessage}</span>
      {/if}
    {/snippet}

    {#snippet actions()}
      <button
        type="button"
        class="header-icon-btn"
        use:islandTooltip={$tStore('fileshelf.add_tooltip')}
        onclick={() => void promptChooseShelfFiles()}
        aria-label={$tStore('fileshelf.add_tooltip')}
      >
        <Plus size={14} />
      </button>

      {#if $fileShelfStore.items.length > 0}
        <button
          type="button"
          class="header-icon-btn clear-btn"
          use:islandTooltip={$tStore('fileshelf.clear_tooltip')}
          onclick={() => {
            rawSelectedPaths = []
            clearFileShelf()
          }}
          aria-label={$tStore('fileshelf.clear_tooltip')}
        >
          <Trash2 size={13} />
        </button>
      {/if}
    {/snippet}
  </PageHeader>
</div>

<style src="./styles/FileShelfPage.css"></style>
