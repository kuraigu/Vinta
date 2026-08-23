<script lang="ts">
  import ScrollView from '../ScrollView.svelte'
  import ContextMenu, { type ContextMenuItem } from '../ContextMenu.svelte'

  import type { FormatType } from './NoteToolbar.svelte'

  interface EditorApi {
    insertFormatting: (
      type: FormatType,
      prefix: string,
      suffix?: string
    ) => void
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

  export interface CodeBlockInfo {
    start: number
    end: number
    lang: string
    code: string
  }

  export function findCodeBlocks(docLines: string[]): CodeBlockInfo[] {
    const blocks: CodeBlockInfo[] = []
    let currentStart = -1
    let currentLang = ''

    for (let i = 0; i < docLines.length; i++) {
      const line = docLines[i] ?? ''
      const openMatch = line.match(/^[ ]{0,3}```(\S*)/)
      if (openMatch) {
        if (currentStart === -1) {
          currentStart = i
          currentLang = openMatch[1] || ''
        } else {
          if (/^[ ]{0,3}```\s*$/.test(line)) {
            const innerCode = docLines.slice(currentStart + 1, i).join('\n')
            blocks.push({
              start: currentStart,
              end: i,
              lang: currentLang,
              code: innerCode
            })
            currentStart = -1
            currentLang = ''
          }
        }
      }
    }

    return blocks
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
  // (instead of a nested <div>) plus the inner HTML.
  interface LineRender {
    cls: string
    html: string
  }

  function renderLine(
    line: string,
    i: number,
    activeLine: number,
    blocks: CodeBlockInfo[]
  ): LineRender {
    const block = blocks.find((b) => i >= b.start && i <= b.end)

    if (block) {
      const isBlockActive = activeLine >= block.start && activeLine <= block.end

      if (isBlockActive) {
        // While typing cursor is inside this code block, look completely normal!
        return {
          cls: '',
          html: `<span>${escapeHtml(line) || '<br/>'}</span>`
        }
      } else {
        // Formatted code block card with Header & Copy button when cursor is outside
        if (i === block.start) {
          const langDisplay = block.lang.trim().toUpperCase() || 'CODE'
          return {
            cls: 'nl-cb-header-line',
            html: `<div class="nl-cb-header" contenteditable="false"><div class="nl-cb-lang"><svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"></polyline><polyline points="8 6 2 12 8 18"></polyline></svg><span>${escapeHtml(langDisplay)}</span></div><button class="nl-cb-copy-btn" data-cb-start="${block.start}" contenteditable="false" type="button"><svg class="nl-cb-copy-icon" viewBox="0 0 24 24" width="11" height="11" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg><span class="nl-cb-copy-label">Copy</span></button></div>`
          }
        }
        if (i === block.end) {
          return {
            cls: 'nl-cb-footer-line',
            html: `<div class="nl-cb-footer" contenteditable="false"></div>`
          }
        }
        return {
          cls: 'nl-code-line',
          html: `<span class="nl-code-text">${escapeHtml(line) || '<br/>'}</span>`
        }
      }
    }

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

    interface EditorSnapshot {
      content: string
      line: number
      offset: number
    }

    const undoStack: EditorSnapshot[] = [
      { content: lines.join('\n'), line: 0, offset: 0 }
    ]
    const redoStack: EditorSnapshot[] = []

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
      let textChild: Node | null = walker.nextNode()

      while (textChild) {
        if (textChild === range.startContainer) {
          charCount += range.startOffset
          break
        }
        charCount += textChild.textContent?.length ?? 0
        textChild = walker.nextNode()
      }

      return charCount
    }

    function placeCaret(el: HTMLDivElement, offset: number) {
      if (!el) return
      node.focus()
      const sel = window.getSelection()
      if (!sel) return
      const range = document.createRange()

      const walk = document.createTreeWalker(el, NodeFilter.SHOW_TEXT, {
        acceptNode(n) {
          if (n.parentElement?.closest('[contenteditable="false"]')) {
            return NodeFilter.FILTER_REJECT
          }
          return NodeFilter.FILTER_ACCEPT
        }
      })
      const textNode = walk.nextNode()

      if (textNode) {
        const max = textNode.textContent?.length ?? 0
        const pos = Math.min(Math.max(0, offset), max)
        range.setStart(textNode, pos)
        range.collapse(true)
      } else {
        range.setStart(el, 0)
        range.collapse(true)
      }

      try {
        sel.removeAllRanges()
        sel.addRange(range)
      } catch {
        try {
          el.focus()
        } catch (err) {
          void err
        }
      }
    }

    // ── Rendering ──

    function setLineRaw(i: number) {
      const el = getLineEl(i)
      if (el) {
        el.className = 'nl-line'
        const text = lines[i] ?? ''
        if (text === '') {
          el.innerHTML = '<br>'
        } else {
          el.textContent = text
        }
      }
    }

    function formatLine(i: number) {
      const el = getLineEl(i)
      if (el) {
        if (i === activeLine) {
          setLineRaw(i)
          return
        }
        const blocks = findCodeBlocks(lines)
        const { cls, html } = renderLine(lines[i] ?? '', i, activeLine, blocks)
        el.className = cls ? `nl-line ${cls}` : 'nl-line'
        el.innerHTML = html
      }
    }

    function buildDOM() {
      node.innerHTML = ''
      const blocks = findCodeBlocks(lines)
      for (let i = 0; i < lines.length; i++) {
        const div = document.createElement('div')
        div.dataset.nlLine = String(i)
        if (i === activeLine) {
          div.className = 'nl-line'
          const text = lines[i] ?? ''
          if (text === '') {
            div.innerHTML = '<br>'
          } else {
            div.textContent = text
          }
        } else {
          const { cls, html } = renderLine(lines[i], i, activeLine, blocks)
          div.className = cls ? `nl-line ${cls}` : 'nl-line'
          div.innerHTML = html
        }
        node.appendChild(div)
      }
    }

    function reindexLines() {
      node.querySelectorAll('.nl-line').forEach((el, i) => {
        ;(el as HTMLElement).dataset.nlLine = String(i)
      })
    }

    // ── Undo / Redo ──

    function applySnapshot(snap: EditorSnapshot) {
      lines = snap.content.split('\n')
      activeLine = Math.max(0, Math.min(snap.line, lines.length - 1))
      buildDOM()
      const el = getLineEl(activeLine)
      if (el) {
        const off = Math.min(snap.offset, lines[activeLine]?.length ?? 0)
        placeCaret(el, off)
      }
      internalChange = true
      onchange(snap.content)
    }

    function emitChange() {
      const snapshot = lines.join('\n')
      const el = getLineEl(activeLine)
      const off = caretOffset(el)

      if (undoStack[undoStack.length - 1]?.content !== snapshot) {
        undoStack.push({ content: snapshot, line: activeLine, offset: off })
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

      const domLineCount = node.querySelectorAll('.nl-line').length
      if (domLineCount !== lines.length) {
        buildDOM()
      } else if (oldActive !== newLine) {
        formatLine(oldActive)
        setLineRaw(newLine)
      }

      const el = getLineEl(newLine)
      if (el) {
        const pos = caret ?? lines[newLine]?.length ?? 0
        placeCaret(el, pos)
      }
    }

    // ── Event handlers ──

    function onInput() {
      if (isComposing) return

      if (node.children.length === 0 || !node.querySelector('.nl-line')) {
        node.innerHTML = '<div class="nl-line" data-nl-line="0"><br></div>'
        lines = ['']
        activeLine = 0
        const first = getLineEl(0)
        if (first) {
          placeCaret(first, 0)
        }
        emitChange()
        return
      }

      const el = getLineEl(activeLine)
      if (el) {
        lines[activeLine] = el.textContent ?? ''
      }

      const newLine = caretLineIndex()
      if (newLine !== activeLine) {
        switchToLine(newLine, caretOffset())
      }

      emitChange()
    }

    function wrapFormatting(prefix: string, suffix: string) {
      insertFormatting('inline', prefix, suffix)
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
          shouldPair = after === '' || /^[\s)\]},.;:]/.test(after)
        } else if (openChar === '*' || openChar === '_' || openChar === '~') {
          shouldPair = before === '' || /\s$/.test(before)
        } else if (openChar === '"' || openChar === "'") {
          shouldPair = before === '' || !/\w$/.test(before)
        } else if (openChar === '`') {
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

          lines[activeLine] = rawLine.slice(0, Math.max(7, off))
          lines.splice(activeLine + 1, 0, after)
          switchToLine(activeLine + 1, 0)
          emitChange()
          return
        }

        if (prefixInfo.type === 'bullet') {
          if (prefixInfo.isMarkerOnly || prefixInfo.content.trim() === '') {
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
          const range = sel.getRangeAt(0)
          const startEl = lineElFromNode(range.startContainer)
          const endEl = lineElFromNode(range.endContainer)
          if (
            (startEl && endEl && startEl !== endEl) ||
            range.commonAncestorContainer === node ||
            sel.toString().length >= lines.join('\n').length - 1
          ) {
            e.preventDefault()
            lines = ['']
            activeLine = 0
            buildDOM()
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

      // ── Ctrl shortcuts ──
      if (ctrl && (e.key === 'b' || e.key === 'B')) {
        e.preventDefault()
        wrapFormatting('**', '**')
        return
      }
      if (ctrl && (e.key === 'i' || e.key === 'I')) {
        e.preventDefault()
        wrapFormatting('*', '*')
        return
      }
      if (ctrl && (e.key === 'k' || e.key === 'K')) {
        e.preventDefault()
        wrapFormatting('[', '](url)')
        return
      }
      if (ctrl && (e.key === '`' || e.key === 'e' || e.key === 'E')) {
        e.preventDefault()
        wrapFormatting('`', '`')
        return
      }
      if (
        ctrl &&
        (e.key === 'd' ||
          e.key === 'D' ||
          (e.shiftKey && (e.key === 'X' || e.key === 'x')))
      ) {
        e.preventDefault()
        wrapFormatting('~~', '~~')
        return
      }

      if (ctrl && (e.key === 's' || e.key === 'S')) {
        e.preventDefault()
        onSave?.()
        return
      }

      if (ctrl && (e.key === 'a' || e.key === 'A')) {
        e.preventDefault()
        selectAll()
        return
      }

      if (ctrl && e.key === 'z') {
        e.preventDefault()
        if (undoStack.length > 1) {
          const current = undoStack.pop()!
          redoStack.push(current)
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
          const el = getLineEl(activeLine)
          if (el) {
            document.execCommand('delete')
            lines[activeLine] = el.textContent ?? ''
          }
          emitChange()
        }
        return
      }

      // ── Arrow navigation across lines (Obsidian column-preservation behavior) ──
      if (e.key === 'ArrowDown' && !e.shiftKey && !e.altKey && !ctrl) {
        if (activeLine < lines.length - 1) {
          e.preventDefault()
          const el = getLineEl(activeLine)
          const off = caretOffset(el)
          const nextIdx = activeLine + 1
          const nextLen = lines[nextIdx]?.length ?? 0
          const targetOffset = Math.min(off, nextLen)
          switchToLine(nextIdx, targetOffset)
          return
        }
      }

      if (e.key === 'ArrowUp' && !e.shiftKey && !e.altKey && !ctrl) {
        if (activeLine > 0) {
          e.preventDefault()
          const el = getLineEl(activeLine)
          const off = caretOffset(el)
          const prevIdx = activeLine - 1
          const prevLen = lines[prevIdx]?.length ?? 0
          const targetOffset = Math.min(off, prevLen)
          switchToLine(prevIdx, targetOffset)
          return
        }
      }
    }

    function onSelectionChange() {
      if (document.activeElement !== node) return
      const sel = window.getSelection()
      if (sel && !sel.isCollapsed && sel.toString().length > 0) return

      const newLine = caretLineIndex()
      if (newLine !== activeLine) {
        switchToLine(newLine, caretOffset())
      }
    }

    // ── Click: checkbox toggle, copy button, link click, line selection & virtual space ──
    function onClick(e: MouseEvent) {
      const target = e.target as HTMLElement

      const copyBtn = target.closest(
        '.nl-cb-copy-btn'
      ) as HTMLButtonElement | null
      if (copyBtn) {
        e.preventDefault()
        e.stopPropagation()
        const startIdx = parseInt(copyBtn.dataset.cbStart ?? '-1', 10)
        const blocks = findCodeBlocks(lines)
        const block = blocks.find((b) => b.start === startIdx)
        if (block) {
          navigator.clipboard.writeText(block.code).catch(() => {})
          const label = copyBtn.querySelector('.nl-cb-copy-label')
          if (label) {
            label.textContent = 'Copied!'
            copyBtn.classList.add('copied')
            setTimeout(() => {
              label.textContent = 'Copy'
              copyBtn.classList.remove('copied')
            }, 2000)
          }
        }
        return
      }

      const link = target.closest('.nl-link') as HTMLAnchorElement | null
      if (link) {
        e.preventDefault()
        const href = link.getAttribute('href')
        if (href) {
          try {
            window.open(href, '_blank', 'noopener,noreferrer')
          } catch (err) {
            void err
          }
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

      const sel = window.getSelection()
      if (sel && !sel.isCollapsed) {
        // User has an active text selection/highlight. Do not collapse it!
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

    function getNodeCharOffset(
      container: HTMLElement,
      targetNode: Node,
      targetOffset: number
    ): number {
      let charCount = 0
      const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT)
      let textChild: Node | null = walker.nextNode()

      while (textChild) {
        if (textChild === targetNode) {
          charCount += targetOffset
          return charCount
        }
        charCount += textChild.textContent?.length ?? 0
        textChild = walker.nextNode()
      }
      return charCount
    }

    function getSelectionRange(): {
      startLine: number
      startOffset: number
      endLine: number
      endOffset: number
      isCollapsed: boolean
    } {
      const sel = window.getSelection()
      if (!sel || !sel.rangeCount) {
        return {
          startLine: activeLine,
          startOffset: lines[activeLine]?.length ?? 0,
          endLine: activeLine,
          endOffset: lines[activeLine]?.length ?? 0,
          isCollapsed: true
        }
      }

      const range = sel.getRangeAt(0)
      if (range.collapsed) {
        const lineIdx = caretLineIndex()
        const lineEl = getLineEl(lineIdx)
        const off = lineEl ? caretOffset(lineEl) : (lines[lineIdx]?.length ?? 0)
        return {
          startLine: lineIdx,
          startOffset: off,
          endLine: lineIdx,
          endOffset: off,
          isCollapsed: true
        }
      }

      const startEl = lineElFromNode(range.startContainer)
      const endEl = lineElFromNode(range.endContainer)

      let startLineIdx =
        startEl && startEl.dataset.nlLine !== undefined
          ? parseInt(startEl.dataset.nlLine, 10)
          : activeLine
      let endLineIdx =
        endEl && endEl.dataset.nlLine !== undefined
          ? parseInt(endEl.dataset.nlLine, 10)
          : startLineIdx

      let startOffset = startEl
        ? getNodeCharOffset(startEl, range.startContainer, range.startOffset)
        : range.startOffset

      let endOffset = endEl
        ? getNodeCharOffset(endEl, range.endContainer, range.endOffset)
        : range.endOffset

      // Normalize if selection was made backwards
      if (
        startLineIdx > endLineIdx ||
        (startLineIdx === endLineIdx && startOffset > endOffset)
      ) {
        const tmpL = startLineIdx
        const tmpO = startOffset
        startLineIdx = endLineIdx
        startOffset = endOffset
        endLineIdx = tmpL
        endOffset = tmpO
      }

      return {
        startLine: startLineIdx,
        startOffset,
        endLine: endLineIdx,
        endOffset,
        isCollapsed: false
      }
    }

    // ── Toolbar API ──
    function insertFormatting(
      type: FormatType,
      prefix: string,
      suffix: string = ''
    ) {
      node.focus()
      const selRange = getSelectionRange()
      const { startLine, startOffset, endLine, endOffset, isCollapsed } =
        selRange

      if (type === 'codeblock') {
        if (!isCollapsed) {
          if (startLine === endLine) {
            const lineText = lines[startLine] ?? ''
            const before = lineText.slice(0, startOffset)
            const selected = lineText.slice(startOffset, endOffset)
            const after = lineText.slice(endOffset)
            const newSlice: string[] = []
            if (before) newSlice.push(before)
            newSlice.push('```', selected, '```')
            if (after) newSlice.push(after)
            lines.splice(startLine, 1, ...newSlice)
            activeLine = startLine + (before ? 2 : 1)
          } else {
            const firstLine = lines[startLine] ?? ''
            const lastLine = lines[endLine] ?? ''
            const before = firstLine.slice(0, startOffset)
            const selectedFirst = firstLine.slice(startOffset)
            const middleLines = lines.slice(startLine + 1, endLine)
            const selectedLast = lastLine.slice(0, endOffset)
            const after = lastLine.slice(endOffset)

            const selectedBlock = [selectedFirst, ...middleLines, selectedLast]
            const newSlice: string[] = []
            if (before) newSlice.push(before)
            newSlice.push('```', ...selectedBlock, '```')
            if (after) newSlice.push(after)

            lines.splice(startLine, endLine - startLine + 1, ...newSlice)
            activeLine = startLine + (before ? 2 : 1)
          }
          buildDOM()
          const targetEl = getLineEl(activeLine)
          if (targetEl) {
            placeCaret(targetEl, 0)
          }
        } else {
          // No selection
          const lineText = lines[startLine] ?? ''
          if (lineText.trim() === '') {
            lines.splice(startLine, 1, '```', '', '```')
            activeLine = startLine + 1
          } else {
            const before = lineText.slice(0, startOffset)
            const after = lineText.slice(startOffset)
            const newSlice: string[] = []
            if (before) newSlice.push(before)
            newSlice.push('```', '', '```')
            if (after) newSlice.push(after)
            lines.splice(startLine, 1, ...newSlice)
            activeLine = startLine + (before ? 2 : 1)
          }
          buildDOM()
          const targetEl = getLineEl(activeLine)
          if (targetEl) {
            placeCaret(targetEl, 0)
          }
        }
      } else if (type === 'line') {
        // Apply prefix across all selected lines
        for (let l = startLine; l <= endLine; l++) {
          const currentLineText = lines[l] ?? ''
          const existingPrefixMatch = currentLineText.match(
            /^(\s*)(#{1,6}\s+|- \[[ xX]\]\s+|- \s+|\*\s+|>\s*)/
          )
          const rawContent = existingPrefixMatch
            ? currentLineText.slice(existingPrefixMatch[0].length)
            : currentLineText.trimStart()
          const existingPrefix = existingPrefixMatch
            ? existingPrefixMatch[2]
            : ''

          const newLine =
            existingPrefix === prefix ||
            (prefix === '- ' &&
              (existingPrefix === '- ' || existingPrefix === '* '))
              ? rawContent
              : `${prefix}${rawContent}`
          lines[l] = newLine
        }
        buildDOM()
        activeLine = endLine
        const targetEl = getLineEl(activeLine)
        if (targetEl) {
          placeCaret(targetEl, lines[activeLine]?.length ?? 0)
        }
      } else if (type === 'link') {
        if (startLine === endLine) {
          const lineText = lines[startLine] ?? ''
          const selected = isCollapsed
            ? ''
            : lineText.slice(startOffset, endOffset)
          const replacement = selected
            ? `[${selected}](https://)`
            : `[](https://)`
          const newText =
            lineText.slice(0, startOffset) +
            replacement +
            lineText.slice(endOffset)
          lines[startLine] = newText
          activeLine = startLine
          setLineRaw(startLine)
          const el = getLineEl(startLine)
          if (el) {
            const newCaret = selected
              ? startOffset + selected.length + 3
              : startOffset + 1
            placeCaret(el, newCaret)
          }
        }
      } else {
        // Inline: bold (**), italic (*), strikethrough (~~)
        if (startLine === endLine) {
          const lineText = lines[startLine] ?? ''
          if (!isCollapsed) {
            const selected = lineText.slice(startOffset, endOffset)
            const replacement = `${prefix}${selected}${suffix}`
            const newText =
              lineText.slice(0, startOffset) +
              replacement +
              lineText.slice(endOffset)
            lines[startLine] = newText
            activeLine = startLine
            setLineRaw(startLine)
            const el = getLineEl(startLine)
            if (el) {
              placeCaret(el, startOffset + replacement.length)
            }
          } else {
            // No selection: insert tokens, place caret right between prefix & suffix
            const replacement = `${prefix}${suffix}`
            const newText =
              lineText.slice(0, startOffset) +
              replacement +
              lineText.slice(startOffset)
            lines[startLine] = newText
            activeLine = startLine
            setLineRaw(startLine)
            const el = getLineEl(startLine)
            if (el) {
              placeCaret(el, startOffset + prefix.length)
            }
          }
        } else {
          // Multi-line selection: wrap each line's selected portion with the tokens
          for (let l = startLine; l <= endLine; l++) {
            const lineText = lines[l] ?? ''
            if (l === startLine) {
              const before = lineText.slice(0, startOffset)
              const selText = lineText.slice(startOffset)
              lines[l] =
                before + (selText ? `${prefix}${selText}${suffix}` : '')
            } else if (l === endLine) {
              const selText = lineText.slice(0, endOffset)
              const after = lineText.slice(endOffset)
              lines[l] = (selText ? `${prefix}${selText}${suffix}` : '') + after
            } else {
              lines[l] = lineText ? `${prefix}${lineText}${suffix}` : ''
            }
          }
          buildDOM()
          activeLine = endLine
          const el = getLineEl(endLine)
          if (el) {
            placeCaret(el, endOffset + prefix.length + suffix.length)
          }
        }
      }
      emitChange()
    }

    function selectAll() {
      node.focus()
      const sel = window.getSelection()
      if (!sel) return
      const range = document.createRange()
      range.selectNodeContents(node)
      sel.removeAllRanges()
      sel.addRange(range)
    }

    const api: EditorApi = { insertFormatting, selectAll }

    function notifyReady() {
      editorApi = api
      onReady?.(api)
    }

    function onPaste(e: ClipboardEvent) {
      e.preventDefault()
      const text = e.clipboardData?.getData('text/plain') ?? ''
      if (!text) return

      const el = getLineEl(activeLine)
      const currentLineText = el
        ? (el.textContent ?? '')
        : (lines[activeLine] ?? '')
      const off = caretOffset(el)
      const before = currentLineText.slice(0, off)
      const after = currentLineText.slice(off)

      const pastedLines = text
        .replace(/\r\n/g, '\n')
        .replace(/\r/g, '\n')
        .split('\n')

      if (pastedLines.length === 1) {
        const merged = before + pastedLines[0] + after
        lines[activeLine] = merged
        if (el) {
          el.textContent = merged
          placeCaret(el, off + pastedLines[0].length)
        }
      } else {
        const firstLine = before + pastedLines[0]
        const lastLine = pastedLines[pastedLines.length - 1] + after
        const middleLines = pastedLines.slice(1, -1)
        const newSlice = [firstLine, ...middleLines, lastLine]

        lines.splice(activeLine, 1, ...newSlice)
        activeLine = activeLine + newSlice.length - 1
        buildDOM()
        const targetEl = getLineEl(activeLine)
        if (targetEl) {
          placeCaret(targetEl, pastedLines[pastedLines.length - 1].length)
        }
      }
      emitChange()
    }

    function onCopy(e: ClipboardEvent) {
      const sel = window.getSelection()
      if (!sel || sel.isCollapsed) return
      const range = sel.getRangeAt(0)
      if (
        range.commonAncestorContainer === node ||
        sel.toString().length >= lines.join('\n').length - 2
      ) {
        e.preventDefault()
        e.clipboardData?.setData('text/plain', lines.join('\n'))
      }
    }

    // ── Mount ──
    buildDOM()
    notifyReady()

    node.addEventListener('input', onInput)
    node.addEventListener('keydown', onKeydown)
    node.addEventListener('click', onClick)
    node.addEventListener('paste', onPaste)
    node.addEventListener('copy', onCopy)
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
        node.removeEventListener('paste', onPaste)
        node.removeEventListener('copy', onCopy)
        document.removeEventListener('selectionchange', onSelectionChange)
      }
    }
  }
</script>

{#if isLocked}
  {@const lockedLines = content.split('\n')}
  {@const lockedBlocks = findCodeBlocks(lockedLines)}
  <ScrollView class="nl-scroll">
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="nl-doc locked"
      onclick={(e) => {
        const copyBtn = (e.target as HTMLElement)?.closest(
          '.nl-cb-copy-btn'
        ) as HTMLButtonElement | null
        if (copyBtn) {
          e.preventDefault()
          e.stopPropagation()
          const startIdx = parseInt(copyBtn.dataset.cbStart ?? '-1', 10)
          const block = lockedBlocks.find((b) => b.start === startIdx)
          if (block) {
            navigator.clipboard.writeText(block.code).catch(() => {})
            const label = copyBtn.querySelector('.nl-cb-copy-label')
            if (label) {
              label.textContent = 'Copied!'
              copyBtn.classList.add('copied')
              setTimeout(() => {
                label.textContent = 'Copy'
                copyBtn.classList.remove('copied')
              }, 2000)
            }
          }
        }
      }}
    >
      {#each lockedLines as line, i (i)}
        {@const r = renderLine(line, i, -1, lockedBlocks)}
        <div class="nl-line {r.cls}">
          <!-- eslint-disable-next-line svelte/no-at-html-tags -->
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

<style src="../styles/notes/NoteLive.css"></style>
