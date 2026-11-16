import { Scissors, Copy, ClipboardPaste, CheckSquare, X } from 'lucide-svelte'
import type { ContextMenuItem } from '../components/CustomContextMenu.svelte'

export interface TextEditMenuLabels {
  cut: string
  copy: string
  paste: string
  selectAll: string
  clear?: string
}

/**
 * Builds standard Cut / Copy / Paste / Select All context menu items for an editable
 * input, textarea, or contenteditable element without dropping Svelte input bindings.
 */
export function buildTextEditMenuItems(
  target: HTMLElement | null,
  labels: TextEditMenuLabels,
  onClear?: () => void
): ContextMenuItem[] {
  if (!target) return []

  const isInput =
    target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement

  let selectedText = ''
  let hasContent = false

  if (isInput) {
    const start = target.selectionStart ?? 0
    const end = target.selectionEnd ?? 0
    selectedText = target.value.slice(start, end)
    hasContent = target.value.length > 0
  } else {
    const sel = window.getSelection()
    selectedText = sel ? sel.toString() : ''
    hasContent = (target.textContent || '').trim().length > 0 || target.innerHTML.includes('<img')
  }

  const hasSelection = selectedText.length > 0

  const items: ContextMenuItem[] = [
    {
      label: labels.cut,
      icon: Scissors,
      disabled: !hasSelection,
      action: () => {
        if (!hasSelection) return
        void navigator.clipboard?.writeText(selectedText).catch(() => {})
        target.focus()
        if (isInput) {
          const start = target.selectionStart ?? 0
          const end = target.selectionEnd ?? 0
          target.setRangeText('', start, end, 'start')
          target.dispatchEvent(new Event('input', { bubbles: true }))
        } else {
          document.execCommand('delete')
          target.dispatchEvent(new Event('input', { bubbles: true }))
        }
      }
    },
    {
      label: labels.copy,
      icon: Copy,
      disabled: !hasSelection,
      action: () => {
        if (!hasSelection) return
        void navigator.clipboard?.writeText(selectedText).catch(() => {})
      }
    },
    {
      label: labels.paste,
      icon: ClipboardPaste,
      action: () => {
        void (async () => {
          let clipText = ''
          try {
            clipText = (await navigator.clipboard?.readText()) || ''
          } catch {
            return
          }
          if (!clipText) return
          target.focus()
          if (isInput) {
            const start = target.selectionStart ?? target.value.length
            const end = target.selectionEnd ?? target.value.length
            target.setRangeText(clipText, start, end, 'end')
            target.dispatchEvent(new Event('input', { bubbles: true }))
          } else {
            document.execCommand('insertText', false, clipText)
            target.dispatchEvent(new Event('input', { bubbles: true }))
          }
        })()
      }
    },
    { label: '', action: () => {}, divider: true },
    {
      label: labels.selectAll,
      icon: CheckSquare,
      disabled: !hasContent,
      action: () => {
        target.focus()
        if (isInput) {
          target.select()
        } else {
          const range = document.createRange()
          range.selectNodeContents(target)
          const sel = window.getSelection()
          sel?.removeAllRanges()
          sel?.addRange(range)
        }
      }
    }
  ]

  if (onClear && labels.clear && hasContent) {
    items.push({
      label: labels.clear,
      icon: X,
      action: onClear
    })
  }

  return items
}
