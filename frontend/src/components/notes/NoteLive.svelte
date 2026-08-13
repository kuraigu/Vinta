<script lang="ts">
  import { slide } from 'svelte/transition'
  import ScrollView from '../ScrollView.svelte'
  import ContextMenu, { type ContextMenuItem } from '../ContextMenu.svelte'

  interface EditorApi {
    insertAtCursor: (prefix: string, suffix: string) => void
    selectAll: () => void
  }

  interface Props {
    content: string
    isLocked?: boolean
    onchange: (newContent: string) => void
    onSave?: () => void
    onReady?: (api: EditorApi) => void
  }

  let {
    content = '',
    isLocked = false,
    onchange,
    onSave,
    onReady
  }: Props = $props()

  // Context menu state (lives in Svelte layer, not the action)
  let ctxOpen = $state(false)
  let ctxX = $state(0)
  let ctxY = $state(0)

  // Reference to the editor div so context menu items can call execCommand
  let editorDivRef = $state<HTMLDivElement | null>(null)

  // Local reference to the editor's imperative API, so the context menu
  // (defined in this same component) can trigger a proper raw-text select-all
  // instead of falling back to execCommand('selectAll').
  let editorApi: EditorApi | null = null

  function handleContextMenu(e: MouseEvent) {
    e.preventDefault()
    ctxX = e.clientX
    ctxY = e.clientY
    ctxOpen = true
  }

  const ctxItems: ContextMenuItem[] = [
    {
      label: 'Select All',
      action: () => {
        editorApi?.selectAll()
      }
    },
    {
      label: 'Copy',
      action: () => {
        const text = window.getSelection()?.toString() ?? ''
        if (text) navigator.clipboard.writeText(text).catch(() => {})
      }
    },
    {
      label: 'Cut',
      action: () => {
        const sel = window.getSelection()
        const text = sel?.toString() ?? ''
        if (text) {
          navigator.clipboard.writeText(text).catch(() => {})
          document.execCommand('delete')
        }
      }
    },
    { divider: true, label: '', action: () => {} },
    {
      label: 'Paste',
      action: () => {
        navigator.clipboard
          .readText()
          .then((text) => {
            if (!text) return
            editorDivRef?.focus()
            document.execCommand('insertText', false, text)
          })
          .catch(() => {})
      }
    }
  ]

  // ─── Utilities ────────────────────────────────────────────────────────────

  function escapeHtml(s: string): string {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
  }

  function formatInline(text: string): string {
    let h = escapeHtml(text)
    // Links before bold/italic so [text] isn't consumed
    h = h.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_m, label, url) => {
      return `<a class="nl-link" href="${escapeHtml(url)}" target="_blank" rel="noopener noreferrer">${escapeHtml(label)}</a>`
    })
    h = h.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    h = h.replace(/\*(.*?)\*/g, '<em>$1</em>')
    h = h.replace(/~~(.*?)~~/g, '<del>$1</del>')
    h = h.replace(/`(.*?)`/g, '<code class="nl-code">$1</code>')
    return h
  }



  interface LinePrefixInfo {
    type: 'checkbox' | 'bullet' | 'none'
    prefix: string
    isMarkerOnly: boolean
    content: string
  }

  function getLinePrefix(rawLine: string): LinePrefixInfo {
    const trimmed = rawLine.trim()
    if (/^-\s*\[[ xX]\]/.test(trimmed)) {
      const isMarkerOnly =
        trimmed === '- [ ]' || trimmed === '- [x]' || trimmed === '- [X]'
      const content = rawLine.replace(/^\s*-\s*\[[ xX]\]\s*/, '')
      return { type: 'checkbox', prefix: '- [ ] ', isMarkerOnly, content }
    }
    if (/^[-*](\s|$)/.test(trimmed) && !trimmed.startsWith('- [')) {
      const isMarkerOnly = trimmed === '-' || trimmed === '*'
      const content = rawLine.replace(/^\s*[-*]\s*/, '')
      return { type: 'bullet', prefix: '- ', isMarkerOnly, content }
    }
    return { type: 'none', prefix: '', isMarkerOnly: false, content: rawLine }
  }

  // renderLine returns the class to put on the .nl-line container itself
  // (instead of a nested <div>) plus the inner HTML. Keeping everything at
  // ONE div-per-line (no nested block divs) matters: nested block elements
  // cause browsers to insert extra newlines when converting a selection to
  // plain text on copy, which is what was breaking Select All.
  interface LineRender {
    cls: string
    html: string
  }

  function renderLine(line: string): LineRender {

    const t = line.trim()
    if (t.startsWith('# '))
      return { cls: '', html: `<h1>${formatInline(line.slice(2))}</h1>` }
    if (t.startsWith('## '))
      return { cls: '', html: `<h2>${formatInline(line.slice(3))}</h2>` }
    if (t.startsWith('### '))
      return { cls: '', html: `<h3>${formatInline(line.slice(4))}</h3>` }
    if (t.startsWith('#### '))
      return { cls: '', html: `<h4>${formatInline(line.slice(5))}</h4>` }
    if (t.startsWith('> '))
      return {
        cls: '',
        html: `<blockquote>${formatInline(line.slice(2))}</blockquote>`
      }
    if (
      t.startsWith('- [x] ') ||
      t.startsWith('- [X] ') ||
      t === '- [x]' ||
      t === '- [X]'
    )
      return {
        cls: 'nl-task checked',
        html: `<span class="nl-checkbox checked" contenteditable="false">✓</span><span class="nl-task-text">${formatInline(line.replace(/^-\s*\[[xX]\]\s*/, '')) || '<br>'}</span>`
      }
    if (t.startsWith('- [ ] ') || t === '- [ ]')
      return {
        cls: 'nl-task',
        html: `<span class="nl-checkbox" contenteditable="false"></span><span class="nl-task-text">${formatInline(line.replace(/^-\s*\[[ ]\]\s*/, '')) || '<br>'}</span>`
      }
    if (t.startsWith('- ') || t.startsWith('* ') || t === '-' || t === '*')
      return {
        cls: 'nl-bullet',
        html: `<span class="nl-dot" contenteditable="false">•</span><span class="nl-bullet-text">${formatInline(line.replace(/^[-*]\s*/, '')) || '<br>'}</span>`
      }
    if (t === '---' || t === '***') return { cls: '', html: `<hr/>` }
    if (!t) return { cls: '', html: `<br/>` }
    return { cls: '', html: `<span>${formatInline(line)}</span>` }
  }

  // ─── Imperative editor action ──────────────────────────────────────────────
  // Single contenteditable container. Each .nl-line is a plain (non-editable)
  // child div. The container itself is contenteditable so the browser handles
  // all native selection (Ctrl+A, drag, Ctrl+C, etc.) naturally.
  // We only update innerHTML of lines the caret is NOT currently on.

  function noteLiveAction(node: HTMLDivElement, params: { content: string }) {
    let lines: string[] = params.content.split('\n')
    let activeLine = 0
    let internalChange = false
    let isComposing = false

    const undoStack: string[] = [lines.join('\n')]
    const redoStack: string[] = []

    // ── DOM helpers ──

    function getLineEl(i: number): HTMLDivElement | null {
      return node.querySelector<HTMLDivElement>(`[data-nl-line="${i}"]`)
    }

    function lineElFromNode(n: Node | null): HTMLDivElement | null {
      let cur: Node | null = n
      while (cur && cur !== node) {
        if ((cur as HTMLElement).dataset?.nlLine !== undefined)
          return cur as HTMLDivElement
        cur = cur.parentNode
      }
      return null
    }

    function caretLineIndex(): number {
      const sel = window.getSelection()
      if (!sel || !sel.anchorNode) return activeLine
      const el = lineElFromNode(sel.anchorNode)
      if (el && el.dataset.nlLine !== undefined)
        return parseInt(el.dataset.nlLine, 10)
      return activeLine
    }

    function caretOffset(lineEl?: HTMLDivElement | null): number {
      const sel = window.getSelection()
      if (!sel || !sel.rangeCount) return 0
      const range = sel.getRangeAt(0)
      const container = lineEl || getLineEl(activeLine)
      if (!container) return range.startOffset

      let charCount = 0
      const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT)
      let node: Node | null = walker.nextNode()

      while (node) {
        if (node === range.startContainer) {
          charCount += range.startOffset
          break
        }
        charCount += node.textContent?.length ?? 0
        node = walker.nextNode()
      }

      if (container.classList.contains('nl-task')) {
        charCount += 6
      } else if (container.classList.contains('nl-bullet')) {
        charCount += 2
      }

      return charCount
    }

    function placeCaret(el: HTMLDivElement, offset: number) {
      let target: Node = el
      const contentHolder = el.querySelector('.nl-task-text, .nl-bullet-text')
      const root = contentHolder || el
      const walk = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, {
        acceptNode(n) {
          if (n.parentElement?.closest('[contenteditable="false"]')) {
            return NodeFilter.FILTER_REJECT
          }
          return NodeFilter.FILTER_ACCEPT
        }
      })
      const first = walk.nextNode()
      if (first) target = first
      const max = target.textContent?.length ?? 0
      const range = document.createRange()
      try {
        range.setStart(target, Math.min(offset, max))
        range.collapse(true)
        const sel = window.getSelection()
        sel?.removeAllRanges()
        sel?.addRange(range)
      } catch {
        try {
          el.focus()
        } catch {}
      }
    }

    // ── Rendering ──

    function buildDOM() {
      node.innerHTML = ''
      for (let i = 0; i < lines.length; i++) {
        const div = document.createElement('div')
        div.dataset.nlLine = String(i)
        const { cls, html } = renderLine(lines[i])
        div.className = cls ? `nl-line ${cls}` : 'nl-line'
        div.innerHTML = html
        node.appendChild(div)
      }
    }

    function reindexLines() {
      node.querySelectorAll('.nl-line').forEach((el, i) => {
        ;(el as HTMLElement).dataset.nlLine = String(i)
      })
    }

    // ── Undo / Redo ──

    function applySnapshot(snapshot: string) {
      lines = snapshot.split('\n')
      activeLine = Math.min(activeLine, lines.length - 1)
      buildDOM()
      internalChange = true
      onchange(snapshot)
    }

    function emitChange() {
      const snapshot = lines.join('\n')
      if (undoStack[undoStack.length - 1] !== snapshot) {
        undoStack.push(snapshot)
        if (undoStack.length > 200) undoStack.shift()
        redoStack.length = 0
      }
      internalChange = true
      onchange(snapshot)
    }

    // ── Active line management ──

    function switchToLine(newLine: number, caret?: number) {
      if (newLine < 0 || newLine >= lines.length) return
      const oldActive = activeLine
      activeLine = newLine

      if (oldActive !== newLine) {
        buildDOM()
      }

      const el = getLineEl(newLine)
      if (el) {
        const pos = caret ?? lines[newLine]?.length ?? 0
        node.focus()
        placeCaret(el, pos)
      }
    }

    // ── Event handlers ──

    function syncLinesFromDOM() {
      const lineEls = Array.from(
        node.querySelectorAll<HTMLDivElement>('.nl-line')
      )
      if (lineEls.length === 0) {
        // If user selected all and deleted (or container was cleared)
        const emptyDiv = document.createElement('div')
        emptyDiv.className = 'nl-line'
        emptyDiv.dataset.nlLine = '0'
        emptyDiv.innerHTML = '<br>'
        node.appendChild(emptyDiv)
        lines = ['']
        activeLine = 0
      } else {
        lines = lineEls.map((el) => {
          if (el.classList.contains('nl-task')) {
            const isChecked = el.classList.contains('checked')
            const textEl = el.querySelector('.nl-task-text')
            const text = textEl ? (textEl.textContent ?? '') : ''
            return (isChecked ? '- [x] ' : '- [ ] ') + text
          }
          if (el.classList.contains('nl-bullet')) {
            const textEl = el.querySelector('span:not(.nl-dot)')
            const text = textEl ? (textEl.textContent ?? '') : ''
            return '- ' + text
          }
          return el.textContent ?? ''
        })
        reindexLines()
        if (activeLine >= lines.length) {
          activeLine = lines.length - 1
        }
      }
    }

    function onInput() {
      if (isComposing) return
      syncLinesFromDOM()
      const newLine = caretLineIndex()

      if (newLine !== activeLine) {
        switchToLine(newLine, caretOffset())
      }

      emitChange()
    }

    function onKeydown(e: KeyboardEvent) {
      const ctrl = e.ctrlKey || e.metaKey

      // ── Tab ──
      if (e.key === 'Tab') {
        e.preventDefault()
        const el = getLineEl(activeLine)
        if (!el) return
        const text = el.textContent ?? ''
        const off = caretOffset()
        if (!e.shiftKey) {
          el.textContent = text.slice(0, off) + '  ' + text.slice(off)
          lines[activeLine] = el.textContent
          placeCaret(el, off + 2)
        } else {
          const stripped = text.replace(/^ {1,2}/, '')
          const removed = text.length - stripped.length
          if (removed > 0) {
            el.textContent = stripped
            lines[activeLine] = stripped
            placeCaret(el, Math.max(0, off - removed))
          }
        }
        emitChange()
        return
      }

      // ── Auto-closing pairs and Triple Backtick ──
      const pairs: Record<string, string> = {
        '`': '`',
        '*': '*',
        _: '_',
        '~': '~',
        '"': '"',
        "'": "'"
      }

      if (!ctrl && !e.altKey && e.key in pairs) {
        const el = getLineEl(activeLine)
        const text = el ? (el.textContent ?? '') : (lines[activeLine] ?? '')
        const off = caretOffset(el)
        const before = text.slice(0, off)
        const after = text.slice(off)
        const openChar = e.key
        const closeChar = pairs[openChar]

        const sel = window.getSelection()

        // 1. Selection wrapping
        if (sel && !sel.isCollapsed && sel.toString().length > 0) {
          e.preventDefault()
          const selectedText = sel.toString()
          const replacement = `${openChar}${selectedText}${closeChar}`
          if (el) {
            el.textContent = before + replacement + after
            lines[activeLine] = el.textContent
            placeCaret(el, off + replacement.length)
          }
          emitChange()
          return
        }

        // 2. Triple backtick code block expansion
        if (openChar === '`' && before.endsWith('``')) {
          e.preventDefault()
          lines[activeLine] = before + '`'
          lines.splice(activeLine + 1, 0, '', '```')
          buildDOM()
          activeLine = activeLine + 1
          setLineRaw(activeLine)
          const target = getLineEl(activeLine)
          if (target) {
            node.focus()
            placeCaret(target, 0)
          }
          emitChange()
          return
        }

        // 3. Skip over closing quote/bracket if caret is before it
        if (
          after.startsWith(closeChar) &&
          openChar === closeChar &&
          openChar !== '*' &&
          openChar !== '_' &&
          openChar !== '`'
        ) {
          e.preventDefault()
          if (el) placeCaret(el, off + 1)
          return
        }

        // 4. Pairing conditions per key type
        let shouldPair = false

        if (openChar === '(' || openChar === '[' || openChar === '{') {
          // Brackets: pair if after is empty, space, or closing bracket/punctuation
          shouldPair = after === '' || /^[\s\)\]\},.;:]/.test(after)
        } else if (openChar === '*' || openChar === '_' || openChar === '~') {
          // Formatting markers: pair if before is empty or ends with whitespace
          shouldPair = before === '' || /\s$/.test(before)
        } else if (openChar === '"' || openChar === "'") {
          // Quotes: pair if before does not end with word char
          shouldPair = before === '' || !/\w$/.test(before)
        } else if (openChar === '`') {
          // Backtick: pair if before does not end with word char
          shouldPair = before === '' || !/\w$/.test(before)
        }

        if (shouldPair) {
          e.preventDefault()
          const newText = before + openChar + closeChar + after
          lines[activeLine] = newText
          if (el) {
            el.textContent = newText
            placeCaret(el, off + 1)
          }
          emitChange()
          return
        }
      }

      // ── Enter ──
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        const el = getLineEl(activeLine)
        const text = el ? (el.textContent ?? '') : (lines[activeLine] ?? '')
        const rawLine = lines[activeLine] ?? ''
        const off = caretOffset(el)
        const before = text.slice(0, off)
        const after = text.slice(off)

        const prefixInfo = getLinePrefix(rawLine)
          if (prefixInfo.type === 'checkbox') {
            if (prefixInfo.isMarkerOnly || prefixInfo.content.trim() === '') {
              // Enter on empty checkbox exits the list completely
              lines[activeLine] = ''
              buildDOM()
              const target = getLineEl(activeLine)
              if (target) {
                node.focus()
                placeCaret(target, 0)
              }
              emitChange()
              return
            }

            // Non-empty checkbox: create a normal line break below (do NOT auto-spawn another checkbox)
            lines[activeLine] = rawLine.slice(0, Math.max(7, off))
            lines.splice(activeLine + 1, 0, after)
            switchToLine(activeLine + 1, 0)
            emitChange()
            return
          }

          if (prefixInfo.type === 'bullet') {
            if (prefixInfo.isMarkerOnly || prefixInfo.content.trim() === '') {
              // Enter on empty bullet exits the list completely
              lines[activeLine] = ''
              buildDOM()
              const target = getLineEl(activeLine)
              if (target) {
                node.focus()
                placeCaret(target, 0)
              }
              emitChange()
              return
            }

            const prefix = prefixInfo.prefix
            lines[activeLine] = rawLine.slice(0, Math.max(prefix.length, off))
            lines.splice(activeLine + 1, 0, prefix + after)
            switchToLine(activeLine + 1, prefix.length)
            emitChange()
            return
          }
        }

        lines[activeLine] = before
        lines.splice(activeLine + 1, 0, after)
        switchToLine(activeLine + 1, 0)
        emitChange()
        return
      }

      // ── Backspace / Delete with active selection or at line boundary ──
      if (e.key === 'Backspace' || e.key === 'Delete') {
        const sel = window.getSelection()
        if (sel && !sel.isCollapsed && sel.toString().length > 0) {
          const text = sel.toString()
          // Check if selection covers the full document or multiple lines
          if (
            text.replace(/\s+/g, '') === lines.join('').replace(/\s+/g, '') ||
            text.includes('\n')
          ) {
            e.preventDefault()
            lines = ['']
            activeLine = 0
            buildDOM()
            node.focus()
            const first = getLineEl(0)
            if (first) placeCaret(first, 0)
            emitChange()
            return
          }
        }
      }

      if (e.key === 'Backspace') {
        const el = getLineEl(activeLine)
        if (!el) return
        const off = caretOffset(el)
        const text = el.textContent ?? ''

        // Delete empty pair when caret is between them (e.g. *|*, (|), [|])
        if (off > 0 && off < text.length) {
          const charBefore = text[off - 1]
          const charAfter = text[off]
          const pairMap: Record<string, string> = {
            '`': '`',
            '*': '*',
            _: '_',
            '~': '~',
            '"': '"',
            "'": "'",
            '(': ')',
            '[': ']',
            '{': '}'
          }
          if (pairMap[charBefore] && pairMap[charBefore] === charAfter) {
            e.preventDefault()
            const updated = text.slice(0, off - 1) + text.slice(off + 1)
            el.textContent = updated
            lines[activeLine] = updated
            placeCaret(el, off - 1)
            emitChange()
            return
          }
        }

        if (off === 0 && activeLine > 0) {
          e.preventDefault()
          const curr = el.textContent ?? ''
          const prev = lines[activeLine - 1] ?? ''
          const merged = prev + curr
          lines[activeLine - 1] = merged
          lines.splice(activeLine, 1)
          const targetLine = activeLine - 1
          el.remove()
          reindexLines()
          activeLine = targetLine
          setLineRaw(targetLine)
          const prevEl = getLineEl(targetLine)
          if (prevEl) {
            prevEl.textContent = merged
            node.focus()
            placeCaret(prevEl, prev.length)
          }
          emitChange()
          return
        }
      }

      // ── Arrow keys ──
      if (e.key === 'ArrowUp' && activeLine > 0) {
        const off = caretOffset()
        if (off === 0) {
          e.preventDefault()
          switchToLine(activeLine - 1, lines[activeLine - 1]?.length ?? 0)
        }
      }
      if (e.key === 'ArrowDown' && activeLine < lines.length - 1) {
        const el = getLineEl(activeLine)
        const off = caretOffset()
        if (el && off >= (el.textContent?.length ?? 0)) {
          e.preventDefault()
          const nextLine = activeLine + 1
          const codeMapNow = computeCodeMap(lines)
          const isCloseFence = codeMapNow[nextLine] === 'close'
          const caretPos = isCloseFence ? (lines[nextLine]?.length ?? 3) : 0
          switchToLine(nextLine, caretPos)
        }
      }

      // ── Ctrl shortcuts ──
      if (ctrl && (e.key === 's' || e.key === 'S')) {
        e.preventDefault()
        onSave?.()
        return
      }

      if (ctrl && (e.key === 'a' || e.key === 'A')) {
        e.preventDefault()
        selectAllRaw()
        return
      }

      if (ctrl && e.key === 'z') {
        e.preventDefault()
        if (undoStack.length > 1) {
          redoStack.push(undoStack.pop()!)
          applySnapshot(undoStack[undoStack.length - 1])
        }
        return
      }
      if (ctrl && (e.key === 'y' || (e.shiftKey && e.key === 'z'))) {
        e.preventDefault()
        if (redoStack.length > 0) {
          const next = redoStack.pop()!
          undoStack.push(next)
          applySnapshot(next)
        }
        return
      }
      if (ctrl && (e.key === 'x' || e.key === 'X')) {
        const sel = window.getSelection()
        if (sel && !sel.isCollapsed && sel.toString().length > 0) {
          e.preventDefault()
          const text = sel.toString()
          navigator.clipboard.writeText(text).catch(() => {})
          try {
            document.execCommand('cut')
          } catch {
            document.execCommand('delete')
          }
          syncLinesFromDOM()
          emitChange()
        }
        return
      }
    }

    function onSelectionChange() {
      if (document.activeElement !== node) return
      const sel = window.getSelection()
      // Skip line switching if user has highlighted text range across elements
      if (sel && !sel.isCollapsed && sel.toString().length > 0) return

      const newLine = caretLineIndex()
      if (newLine !== activeLine) {
        switchToLine(newLine, caretOffset())
      }
    }

    // ── Click: checkbox toggle, link click, line selection & virtual space ──
    function onClick(e: MouseEvent) {
      const target = e.target as HTMLElement
      const link = target.closest('.nl-link') as HTMLAnchorElement | null
      if (link) {
        e.preventDefault()
        const href = link.getAttribute('href')
        if (href) {
          try {
            window.open(href, '_blank', 'noopener,noreferrer')
          } catch {}
        }
        return
      }

      const checkbox = target.closest('.nl-checkbox') as HTMLElement | null
      if (checkbox) {
        e.preventDefault()
        const lineEl = lineElFromNode(checkbox)
        if (lineEl?.dataset.nlLine !== undefined) {
          const i = parseInt(lineEl.dataset.nlLine, 10)
          const line = lines[i] ?? ''
          if (line.startsWith('- [ ] '))
            lines[i] = line.replace('- [ ] ', '- [x] ')
          else if (/^- \[[xX]\] /.test(line))
            lines[i] = line.replace(/^- \[[xX]\] /, '- [ ] ')
          formatLine(i)
          emitChange()
        }
        return
      }

      const lineEl = lineElFromNode(target)
      if (lineEl && lineEl.dataset.nlLine !== undefined) {
        const clickedIdx = parseInt(lineEl.dataset.nlLine, 10)
        switchToLine(clickedIdx, caretOffset())
        return
      }

      // Clicking empty virtual space anywhere in container
      const lastIdx = Math.max(0, lines.length - 1)
      switchToLine(lastIdx, lines[lastIdx]?.length ?? 0)
    }

    // ── Toolbar API ──
    function insertAtCursor(prefix: string, suffix: string) {
      node.focus()
      const el = getLineEl(activeLine)
      if (!el) return
      const sel = window.getSelection()
      if (!sel || !sel.rangeCount) return
      const range = sel.getRangeAt(0)
      const text = el.textContent ?? ''
      const start = range.startOffset
      const end = range.endOffset
      const selected = text.slice(start, end)
      const replacement = `${prefix}${selected || (suffix ? 'text' : '')}${suffix}`
      const fullText = text.slice(0, start) + replacement + text.slice(end)

      if (fullText.includes('\n')) {
        const newLines = fullText.split('\n')
        lines.splice(activeLine, 1, ...newLines)

        const prefixLines = prefix.split('\n')
        const targetLine = activeLine + prefixLines.length - 1
        const lastPrefixLine = prefixLines[prefixLines.length - 1]
        const caretPos = selected
          ? lastPrefixLine.length + selected.length
          : suffix
            ? lastPrefixLine.length + (suffix.startsWith('\n') ? 0 : 4)
            : lastPrefixLine.length

        activeLine = targetLine
        buildDOM()
        const targetEl = getLineEl(targetLine)
        if (targetEl) {
          node.focus()
          placeCaret(targetEl, caretPos)
        }
      } else {
        el.textContent = fullText
        lines[activeLine] = fullText
        const newCaret =
          start + prefix.length + (selected ? selected.length : suffix ? 4 : 0)
        placeCaret(el, newCaret)
      }
      emitChange()
    }

    // Select-all that operates on the raw markdown source (not the rendered
    // HTML), so what gets copied is real text with real linebreaks — used
    // by both Ctrl+A and the context menu's "Select All".
    function selectAllRaw() {
      for (let i = 0; i < lines.length; i++) {
        setLineRaw(i)
      }
      node.focus()
      setTimeout(() => {
        const sel = window.getSelection()
        const range = document.createRange()
        range.selectNodeContents(node)
        sel?.removeAllRanges()
        sel?.addRange(range)
      }, 0)
    }

    const api: EditorApi = { insertAtCursor, selectAll: selectAllRaw }

    function notifyReady() {
      editorApi = api
      onReady?.(api)
    }

    // ── Mount ──
    buildDOM()
    notifyReady()

    node.addEventListener('input', onInput)
    node.addEventListener('keydown', onKeydown)
    node.addEventListener('click', onClick)
    document.addEventListener('selectionchange', onSelectionChange)
    node.addEventListener('compositionstart', () => {
      isComposing = true
    })
    node.addEventListener('compositionend', () => {
      isComposing = false
      onInput()
    })

    return {
      update(newParams: { content: string }) {
        if (internalChange) {
          internalChange = false
          return
        }
        if (newParams.content !== lines.join('\n')) {
          lines = newParams.content.split('\n')
          activeLine = 0
          buildDOM()
          notifyReady()
        }
      },
      destroy() {
        node.removeEventListener('input', onInput)
        node.removeEventListener('keydown', onKeydown)
        node.removeEventListener('click', onClick)
        document.removeEventListener('selectionchange', onSelectionChange)
      }
    }
  }
</script>

{#if isLocked}
  {@const lockedLines = content.split('\n')}
  {@const lockedCodeMap = computeCodeMap(lockedLines)}
  <ScrollView class="nl-scroll">
    <div class="nl-doc locked">
      {#each lockedLines as line, i (i)}
        {@const r = renderLine(line, lockedCodeMap[i])}
        <div
          class="nl-line {r.cls}"
          data-code-ctx={lockedCodeMap[i] ?? undefined}
        >
          {@html r.html}
        </div>
      {/each}
    </div>
  </ScrollView>
{:else}
  <ScrollView class="nl-scroll">
    <div
      bind:this={editorDivRef}
      use:noteLiveAction={{ content }}
      class="nl-doc editable"
      contenteditable="true"
      role="textbox"
      tabindex="0"
      aria-multiline="true"
      oncontextmenu={handleContextMenu}
    ></div>
  </ScrollView>
{/if}

{#if ctxOpen}
  <ContextMenu
    x={ctxX}
    y={ctxY}
    items={ctxItems}
    onclose={() => {
      ctxOpen = false
    }}
  />
{/if}

<style>
  :global(.nl-scroll) {
    width: 100%;
    height: 100%;
  }

  .nl-doc {
    width: 100%;
    min-height: 100%;
    padding: 0 0 80px 0; /* virtual space at bottom */
    font-family: inherit;
    font-size: 13px;
    line-height: 1.6;
    color: var(--c-text-primary);
    text-align: left;
    outline: none;
    box-sizing: border-box;
    cursor: text;
    caret-color: var(--c-text-primary);
    white-space: pre-wrap;
    word-break: break-word;
  }

  .nl-doc :global(.nl-line) {
    width: 100%;
    min-height: 1.4em;
    position: relative;
    outline: none;
  }

  .nl-doc :global(h1) {
    font-size: 18px;
    font-weight: 700;
    margin: 4px 0 2px 0;
    line-height: 1.3;
  }
  .nl-doc :global(h2) {
    font-size: 15px;
    font-weight: 700;
    margin: 4px 0 2px 0;
  }
  .nl-doc :global(h3) {
    font-size: 13.5px;
    font-weight: 600;
    margin: 3px 0 2px 0;
  }
  .nl-doc :global(h4) {
    font-size: 13px;
    font-weight: 600;
    margin: 2px 0;
  }

  .nl-doc :global(blockquote) {
    border-left: 3px solid var(--c-primary, #6366f1);
    padding-left: 10px;
    margin: 2px 0;
    font-style: italic;
    color: var(--c-text-secondary);
  }

  .nl-doc :global(.nl-bullet) {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .nl-doc :global(.nl-dot) {
    color: var(--c-primary, #6366f1);
    font-weight: bold;
    flex-shrink: 0;
    user-select: none;
  }

  .nl-doc :global(.nl-task) {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .nl-doc :global(.nl-checkbox) {
    width: 15px;
    height: 15px;
    border-radius: 4px;
    border: 1.5px solid var(--c-border);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 10px;
    font-weight: bold;
    color: #fff;
    cursor: pointer;
    flex-shrink: 0;
    user-select: none;
    transition: all 0.15s ease;
  }
  .nl-doc :global(.nl-checkbox.checked) {
    background: var(--c-primary, #6366f1);
    border-color: var(--c-primary, #6366f1);
  }
  .nl-doc :global(.nl-task.checked .nl-task-text) {
    text-decoration: line-through;
    opacity: 0.5;
  }

  .nl-doc :global(hr) {
    border: none;
    border-top: 1px solid var(--c-border);
    margin: 8px 0;
  }

  .nl-doc :global(strong) {
    font-weight: 700;
  }
  .nl-doc :global(em) {
    font-style: italic;
  }
  .nl-doc :global(del) {
    text-decoration: line-through;
    opacity: 0.6;
  }

  .nl-doc :global(.nl-code) {
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 4px;
    padding: 1px 5px;
    font-family: monospace;
    font-size: 11.5px;
    color: #f43f5e;
  }

  .nl-doc :global(.nl-link) {
    color: var(--c-primary, #6366f1);
    text-decoration: underline;
    text-underline-offset: 2px;
    cursor: pointer;
    transition: opacity 0.15s;
  }
  .nl-doc :global(.nl-link:hover) {
    opacity: 0.75;
  }


  .nl-doc :global(.nl-line.nl-bullet-editing),
  .nl-doc :global(.nl-line.nl-task-editing) {
    padding-left: 12px;
  }
  .nl-doc :global(.nl-link) {
    color: var(--c-primary, #6366f1);
    text-decoration: underline;
    cursor: pointer;
  }
</style>
