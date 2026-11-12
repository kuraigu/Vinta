/**
 * @fileoverview Temporary file drop shelf state.
 *
 * Holds files dragged in from Windows Explorer, provides previews, handles dragging
 * files out to other apps, and supports batch zipping and EXIF metadata stripping.
 */

import { writable, get } from 'svelte/store'
import { Events } from '$lib/events'
import {
  BundleShelfZip,
  CopyShelfPathsToClipboard,
  InspectShelfFiles,
  PromptChooseShelfFiles,
  RevealShelfFileInExplorer,
  StartNativeFileDrag,
  StripShelfImageMetadata
} from '../../bindings/Vinta/backend/app'

const STORAGE_KEY = 'vinta_file_shelf_paths'

/**
 * Metadata and preview payload for a file or folder staged in the File Shelf.
 */
export interface ShelfFileItem {
  path: string
  name: string
  size: number
  extension: string
  isDir: boolean
  category: string
  previewData?: string
  modifiedTime?: number
  addedAt: number
}

/**
 * Reactive state for the File Holding Shelf store.
 */
export interface FileShelfState {
  items: ShelfFileItem[]
  busyAction: 'zip' | 'exif' | 'copy' | null
  statusMessage: string | null
  isDraggingOut: boolean
  isOverShelf: boolean
  draggingPaths: string[]
}

export const fileShelfStore = writable<FileShelfState>({
  items: [],
  busyAction: null,
  statusMessage: null,
  isDraggingOut: false,
  isOverShelf: false,
  draggingPaths: []
})

let dragEventsBound = false

export function ensureDragEventsBound(): void {
  if (dragEventsBound) return
  dragEventsBound = true

  Events.On('fileshelf:drag-over-shelf', (rawEv: unknown) => {
    const isOver = Boolean(
      typeof rawEv === 'object' && rawEv !== null && 'data' in rawEv
        ? (rawEv as { data: unknown }).data
        : rawEv
    )
    fileShelfStore.update((s) => ({
      ...s,
      isOverShelf: isOver
    }))
  })

  Events.On('fileshelf:drag-ended', (rawEv: unknown) => {
    fileShelfStore.update((s) => ({
      ...s,
      isDraggingOut: false,
      isOverShelf: false,
      draggingPaths: []
    }))

    const evObj = rawEv as { data?: unknown; paths?: string[]; dropped?: boolean } | undefined
    const payload = (evObj?.data ?? evObj) as { paths?: string[]; dropped?: boolean } | undefined
    if (payload && payload.dropped && Array.isArray(payload.paths) && payload.paths.length > 0) {
      removeFilesFromShelf(payload.paths)
      const count = payload.paths.length
      setShelfStatus(count === 1 ? 'Moved 1 file from shelf' : `Moved ${count} files from shelf`)
    } else {
      setShelfStatus('Drag cancelled')
    }
  })
}

let statusTimer: ReturnType<typeof setTimeout> | null = null

/**
 * Displays a temporary status badge message in the File Shelf footer.
 */
export function setShelfStatus(message: string | null, durationMs = 2400): void {
  if (statusTimer) {
    clearTimeout(statusTimer)
    statusTimer = null
  }
  fileShelfStore.update((s) => ({ ...s, statusMessage: message }))
  if (message && durationMs > 0) {
    statusTimer = setTimeout(() => {
      fileShelfStore.update((s) => ({ ...s, statusMessage: null }))
      statusTimer = null
    }, durationMs)
  }
}

function persistShelfPaths(items: ShelfFileItem[]): void {
  try {
    const paths = items.map((item) => item.path)
    localStorage.setItem(STORAGE_KEY, JSON.stringify(paths))
  } catch {
    // Ignore storage quota errors
  }
}

/**
 * Queries the Go backend to stat local file paths and generate image thumbnails.
 */
export async function inspectShelfFiles(
  paths: string[]
): Promise<ShelfFileItem[]> {
  if (!paths || paths.length === 0) return []
  try {
    const res = await InspectShelfFiles(paths)
    return Array.isArray(res) ? (res as ShelfFileItem[]) : []
  } catch {
    return []
  }
}

/**
 * Rehydrates previously staged shelf files from localStorage on startup.
 */
export async function initFileShelfStore(): Promise<void> {
  ensureDragEventsBound()
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return
    const savedPaths = JSON.parse(raw)
    if (!Array.isArray(savedPaths) || savedPaths.length === 0) return
    const validItems = await inspectShelfFiles(savedPaths)
    fileShelfStore.update((s) => ({ ...s, items: validItems }))
    persistShelfPaths(validItems)
  } catch {
    // Ignore malformed saved paths
  }
}

/**
 * Adds one or more local file paths to the shelf, deduplicating by case-insensitive path.
 */
export async function addFilesToShelf(
  paths: string[]
): Promise<ShelfFileItem[]> {
  const inspected = await inspectShelfFiles(paths)
  if (inspected.length === 0) return []

  let nextItems: ShelfFileItem[] = []
  fileShelfStore.update((state) => {
    const existingMap = new Map<string, ShelfFileItem>()
    for (const item of state.items) {
      existingMap.set(item.path.toLowerCase(), item)
    }
    for (const fresh of inspected) {
      existingMap.set(fresh.path.toLowerCase(), fresh)
    }
    nextItems = Array.from(existingMap.values()).sort(
      (a, b) => b.addedAt - a.addedAt
    )
    return {
      ...state,
      items: nextItems
    }
  })

  persistShelfPaths(nextItems)
  setShelfStatus(
    inspected.length === 1
      ? `Added ${inspected[0].name}`
      : `Added ${inspected.length} items`
  )
  return inspected
}

/**
 * Removes a single staged file from the shelf by path.
 */
export function removeFileFromShelf(path: string): void {
  removeFilesFromShelf([path])
}

/**
 * Removes multiple staged files from the shelf by path.
 */
export function removeFilesFromShelf(paths: string[]): void {
  if (!paths || paths.length === 0) return
  const targetKeys = new Set(paths.map((p) => p.toLowerCase()))
  let nextItems: ShelfFileItem[] = []
  fileShelfStore.update((state) => {
    nextItems = state.items.filter(
      (item) => !targetKeys.has(item.path.toLowerCase())
    )
    return {
      ...state,
      items: nextItems
    }
  })
  persistShelfPaths(nextItems)
}

/**
 * Opens the native file chooser dialog and adds the selected files to the shelf.
 */
export async function promptChooseShelfFiles(): Promise<ShelfFileItem[]> {
  try {
    const res = await PromptChooseShelfFiles()
    const picked = Array.isArray(res) ? (res as ShelfFileItem[]) : []
    if (picked.length === 0) return []
    return await addFilesToShelf(picked.map((p) => p.path))
  } catch {
    return []
  }
}

/**
 * Clears all staged files from the shelf.
 */
export function clearFileShelf(): void {
  fileShelfStore.update((state) => ({
    ...state,
    items: [],
    statusMessage: null,
    isDraggingOut: false,
    isOverShelf: false,
    draggingPaths: []
  }))
  persistShelfPaths([])
}

/**
 * Initiates a native Windows Shell OLE drag-and-drop operation for the specified paths.
 */
export async function dragFilesOut(paths: string[]): Promise<void> {
  if (!paths || paths.length === 0) return
  ensureDragEventsBound()
  fileShelfStore.update((s) => ({
    ...s,
    isDraggingOut: true,
    isOverShelf: false,
    draggingPaths: paths
  }))
  try {
    await StartNativeFileDrag(paths)
  } catch {
    fileShelfStore.update((s) => ({
      ...s,
      isDraggingOut: false,
      isOverShelf: false,
      draggingPaths: []
    }))
    setShelfStatus('Unable to start file drag')
  }
}

/**
 * Copies the file paths of chosen shelf items to the system clipboard.
 * If no paths are provided, the clipboard is left untouched.
 */
export async function copyShelfPaths(paths?: string[]): Promise<boolean> {
  const targetPaths = paths && paths.length > 0 ? paths : []
  if (targetPaths.length === 0) return false

  fileShelfStore.update((s) => ({ ...s, busyAction: 'copy' }))
  try {
    await CopyShelfPathsToClipboard(targetPaths)
    setShelfStatus(
      targetPaths.length === 1
        ? 'Copied 1 path'
        : `Copied ${targetPaths.length} paths`
    )
    return true
  } catch {
    setShelfStatus('Failed to copy paths')
    return false
  } finally {
    fileShelfStore.update((s) => ({ ...s, busyAction: null }))
  }
}

/**
 * Compresses all (or selected) shelf items into a .zip archive and adds it to the shelf.
 */
export async function bundleShelfToZip(
  paths?: string[]
): Promise<ShelfFileItem | null> {
  const targetPaths =
    paths && paths.length > 0
      ? paths
      : get(fileShelfStore).items.map((i) => i.path)
  if (targetPaths.length === 0) return null

  fileShelfStore.update((s) => ({ ...s, busyAction: 'zip' }))
  try {
    const created = (await BundleShelfZip(
      targetPaths
    )) as ShelfFileItem | null
    if (created && created.path) {
      await addFilesToShelf([created.path])
      setShelfStatus(`Created ${created.name}`)
      return created
    }
    setShelfStatus('Failed to create ZIP')
    return null
  } catch {
    setShelfStatus('Failed to create ZIP')
    return null
  } finally {
    fileShelfStore.update((s) => ({ ...s, busyAction: null }))
  }
}

/**
 * Strips EXIF and ancillary metadata from staged JPEG/PNG images and adds cleaned copies to the shelf.
 */
export async function stripShelfExif(
  paths?: string[]
): Promise<ShelfFileItem[]> {
  const targetPaths =
    paths && paths.length > 0
      ? paths
      : get(fileShelfStore)
          .items.filter(
            (i) =>
              i.extension === '.jpg' ||
              i.extension === '.jpeg' ||
              i.extension === '.png'
          )
          .map((i) => i.path)
  if (targetPaths.length === 0) {
    setShelfStatus('No JPEG or PNG images on shelf')
    return []
  }

  fileShelfStore.update((s) => ({ ...s, busyAction: 'exif' }))
  try {
    const cleaned = (await StripShelfImageMetadata(
      targetPaths
    )) as ShelfFileItem[] | null
    if (Array.isArray(cleaned) && cleaned.length > 0) {
      await addFilesToShelf(cleaned.map((c) => c.path))
      setShelfStatus(
        cleaned.length === 1
          ? `Cleaned ${cleaned[0].name}`
          : `Cleaned ${cleaned.length} images`
      )
      return cleaned
    }
    setShelfStatus('No supported images to clean')
    return []
  } catch {
    setShelfStatus('Failed to strip image metadata')
    return []
  } finally {
    fileShelfStore.update((s) => ({ ...s, busyAction: null }))
  }
}

/**
 * Formats a byte count into a concise human-readable file size string.
 */
export function formatFileSize(bytes: number, isDir = false): string {
  if (isDir) return 'Folder'
  if (bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.min(
    Math.floor(Math.log(bytes) / Math.log(1024)),
    units.length - 1
  )
  const val = bytes / Math.pow(1024, i)
  return `${val < 10 && i > 0 ? val.toFixed(1) : Math.round(val)} ${units[i]}`
}

/**
 * Reveals a staged file or folder in Windows Explorer with selection.
 */
export async function revealShelfFileInExplorer(path: string): Promise<void> {
  try {
    await RevealShelfFileInExplorer(path)
  } catch {
    setShelfStatus('Failed to reveal file in Explorer')
  }
}


