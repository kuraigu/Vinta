<script lang="ts">
  import ScrollView from '../ScrollView.svelte'
  import { Check, Copy } from 'lucide-svelte'

  interface Props {
    content: string
  }

  let { content }: Props = $props()
  let copiedIndex = $state<number | null>(null)

  function copyCode(code: string, index: number) {
    if (navigator?.clipboard?.writeText) {
      navigator.clipboard.writeText(code).then(() => {
        copiedIndex = index
        setTimeout(() => {
          if (copiedIndex === index) copiedIndex = null
        }, 2000)
      })
    }
  }

  interface CodeBlockData {
    lang: string
    code: string
    highlightedHtml: string
  }

  function highlightSyntax(code: string): string {
    if (!code) return ''

    // Tokenize comments, strings, directives, keywords, numbers, and plain text in one single-pass regex
    const tokenRegex =
      /(\/\/[^\n]*|\/\*[\s\S]*?\*\/)|("[^"\\]*(?:\\.[^"\\]*)*"|'[^'\\]*(?:\\.[^'\\]*)*')|(#\w+)|(\b(?:using|namespace|class|struct|public|private|protected|template|typename|void|int|float|double|char|bool|auto|const|static|if|else|for|while|do|return|switch|case|break|continue|new|delete|try|catch|throw|function|let|var|import|from|export|default|async|await)\b)|(\b\d+\b)|([^"'/#a-zA-Z0-9_\s]+|[a-zA-Z_]\w*|\s+)/g

    return code.replace(
      tokenRegex,
      (match, comment, string, directive, keyword, number) => {
        if (comment)
          return `<span class="tok-comment">${escapeHtml(comment)}</span>`
        if (string)
          return `<span class="tok-string">${escapeHtml(string)}</span>`
        if (directive)
          return `<span class="tok-directive">${escapeHtml(directive)}</span>`
        if (keyword)
          return `<span class="tok-keyword">${escapeHtml(keyword)}</span>`
        if (number)
          return `<span class="tok-number">${escapeHtml(number)}</span>`
        return escapeHtml(match)
      }
    )
  }

  function escapeHtml(str: string): string {
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
  }

  function formatInline(text: string): string {
    let html = escapeHtml(text)
    html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    html = html.replace(/\*(.*?)\*/g, '<em>$1</em>')
    html = html.replace(/~~(.*?)~~/g, '<del>$1</del>')
    html = html.replace(/`(.*?)`/g, '<code class="inline-code">$1</code>')
    return html
  }

  // Parse markdown into rendered segments and codeblock references
  function parseMarkdown(md: string): {
    html: string
    codeBlocks: CodeBlockData[]
  } {
    if (!md) {
      return {
        html: '<span class="empty-preview">Write your markdown note here...</span>',
        codeBlocks: []
      }
    }

    const codeBlocks: CodeBlockData[] = []
    const lines = md.split('\n')
    const resultLines: string[] = []

    let inCodeBlock = false
    let currentCodeLang = ''
    let currentCodeLines: string[] = []

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      const trimmed = line.trim()

      if (trimmed.startsWith('```')) {
        if (!inCodeBlock) {
          inCodeBlock = true
          currentCodeLang = trimmed.slice(3).trim()
          currentCodeLines = []
        } else {
          inCodeBlock = false
          const idx = codeBlocks.length
          const rawCode = currentCodeLines.join('\n')
          codeBlocks.push({
            lang: currentCodeLang.toLowerCase() || 'code',
            code: rawCode,
            highlightedHtml: highlightSyntax(rawCode)
          })
          resultLines.push(`<!--CODEBLOCK_${idx}-->`)
        }
        continue
      }

      if (inCodeBlock) {
        currentCodeLines.push(line)
        continue
      }

      if (trimmed.startsWith('###### ')) {
        resultLines.push(`<h6>${formatInline(trimmed.slice(7))}</h6>`)
      } else if (trimmed.startsWith('##### ')) {
        resultLines.push(`<h5>${formatInline(trimmed.slice(6))}</h5>`)
      } else if (trimmed.startsWith('#### ')) {
        resultLines.push(`<h4>${formatInline(trimmed.slice(5))}</h4>`)
      } else if (trimmed.startsWith('### ')) {
        resultLines.push(`<h3>${formatInline(trimmed.slice(4))}</h3>`)
      } else if (trimmed.startsWith('## ')) {
        resultLines.push(`<h2>${formatInline(trimmed.slice(3))}</h2>`)
      } else if (trimmed.startsWith('# ')) {
        resultLines.push(`<h1>${formatInline(trimmed.slice(2))}</h1>`)
      } else if (trimmed === '---' || trimmed === '***') {
        resultLines.push('<hr/>')
      } else if (trimmed.startsWith('> ')) {
        resultLines.push(
          `<blockquote>${formatInline(trimmed.slice(2))}</blockquote>`
        )
      } else if (trimmed.startsWith('>')) {
        resultLines.push(
          `<blockquote>${formatInline(trimmed.slice(1))}</blockquote>`
        )
      } else if (trimmed.startsWith('- [x] ') || trimmed.startsWith('- [X] ')) {
        resultLines.push(
          `<div class="task-item checked">✓ <span>${formatInline(trimmed.slice(6))}</span></div>`
        )
      } else if (trimmed.startsWith('- [ ] ')) {
        resultLines.push(
          `<div class="task-item">☐ <span>${formatInline(trimmed.slice(6))}</span></div>`
        )
      } else if (trimmed.startsWith('- ') || trimmed.startsWith('* ')) {
        resultLines.push(`<li>${formatInline(trimmed.slice(2))}</li>`)
      } else {
        resultLines.push(formatInline(line))
      }
    }

    if (inCodeBlock) {
      const idx = codeBlocks.length
      const rawCode = currentCodeLines.join('\n')
      codeBlocks.push({
        lang: currentCodeLang.toLowerCase() || 'code',
        code: rawCode,
        highlightedHtml: highlightSyntax(rawCode)
      })
      resultLines.push(`<!--CODEBLOCK_${idx}-->`)
    }

    return {
      html: resultLines.join('<br/>'),
      codeBlocks
    }
  }

  let parsed = $derived(parseMarkdown(content))
</script>

<ScrollView class="preview-scroll">
  <div class="markdown-preview">
    {#each parsed.html.split(/(<!--CODEBLOCK_\d+-->)/) as chunk, i (i)}
      {#if chunk.startsWith('<!--CODEBLOCK_') && chunk.endsWith('-->')}
        {@const blockIdx = parseInt(
          chunk.replace('<!--CODEBLOCK_', '').replace('-->', ''),
          10
        )}
        {@const block = parsed.codeBlocks[blockIdx]}
        {#if block}
          <div class="code-block-container">
            <div class="code-block-header">
              <span class="code-lang">{block.lang}</span>
              <button
                class="code-copy-btn"
                onclick={() => copyCode(block.code, blockIdx)}
              >
                {#if copiedIndex === blockIdx}
                  <Check size={11} />
                  <span>Copied!</span>
                {:else}
                  <Copy size={11} />
                  <span>Copy</span>
                {/if}
              </button>
            </div>
            <pre class="code-content"><code
                ><!-- eslint-disable-next-line svelte/no-at-html-tags -->
{@html block.highlightedHtml}</code
              ></pre>
          </div>
        {/if}
      {:else}
        <!-- eslint-disable-next-line svelte/no-at-html-tags -->
        {@html chunk}
      {/if}
    {/each}
  </div>
</ScrollView>

<style>
  :global(.preview-scroll) {
    flex: 1;
    min-height: 0;
    width: 100%;
  }

  .markdown-preview {
    font-size: 12px;
    line-height: 1.6;
    color: var(--c-text-primary);
    text-align: left;
    padding: 0;
    margin: 0;
    width: 100%;
    box-sizing: border-box;
  }

  .markdown-preview :global(h1) {
    font-size: 18px;
    font-weight: 700;
    margin: 0 0 8px 0;
    text-align: left;
  }

  .markdown-preview :global(h2) {
    font-size: 15px;
    font-weight: 700;
    margin: 10px 0 6px 0;
    text-align: left;
  }

  .markdown-preview :global(h3) {
    font-size: 13px;
    font-weight: 600;
    margin: 8px 0 4px 0;
    text-align: left;
  }

  .markdown-preview :global(h4) {
    font-size: 12px;
    font-weight: 600;
    margin: 6px 0 4px 0;
    text-align: left;
  }

  .markdown-preview :global(h5),
  .markdown-preview :global(h6) {
    font-size: 11px;
    font-weight: 600;
    margin: 4px 0 2px 0;
    text-align: left;
  }

  .markdown-preview :global(hr) {
    border: none;
    border-top: 1px solid var(--c-border);
    margin: 12px 0;
  }

  .markdown-preview :global(.inline-code) {
    background: rgba(255, 255, 255, 0.1);
    padding: 2px 5px;
    border-radius: 4px;
    font-family: monospace;
    font-size: 11px;
  }

  .code-block-container {
    background: rgba(0, 0, 0, 0.4);
    border: 1px solid var(--c-border);
    border-radius: 8px;
    margin: 8px 0;
    overflow: hidden;
    text-align: left;
  }

  .code-block-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4px 10px;
    background: rgba(255, 255, 255, 0.05);
    border-bottom: 1px solid var(--c-border);
    font-family: monospace;
    font-size: 10px;
    color: var(--c-text-secondary);
  }

  .code-lang {
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: 600;
  }

  .code-copy-btn {
    display: flex;
    align-items: center;
    gap: 3px;
    background: transparent;
    border: none;
    color: var(--c-text-secondary);
    font-size: 10px;
    cursor: pointer;
    padding: 2px 6px;
    border-radius: 4px;
    transition: all 0.15s ease;
  }

  .code-copy-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: var(--c-text-primary);
  }

  .code-content {
    padding: 8px 10px;
    margin: 0;
    font-family: Consolas, 'Courier New', monospace;
    font-size: 11px;
    line-height: 1.5;
    overflow-x: auto;
    white-space: pre-wrap;
    color: var(--c-text-primary);
  }

  .markdown-preview :global(.tok-keyword) {
    color: #c084fc;
    font-weight: 600;
  }

  .markdown-preview :global(.tok-directive) {
    color: #f472b6;
    font-weight: 600;
  }

  .markdown-preview :global(.tok-string) {
    color: #a3e635;
  }

  .markdown-preview :global(.tok-comment) {
    color: rgba(255, 255, 255, 0.4);
    font-style: italic;
  }

  .markdown-preview :global(.tok-number) {
    color: #fb923c;
  }

  .markdown-preview :global(blockquote) {
    border-left: 3px solid var(--c-primary, #6366f1);
    background: rgba(255, 255, 255, 0.04);
    margin: 6px 0;
    padding: 6px 10px;
    border-radius: 0 6px 6px 0;
    color: var(--c-text-secondary);
    font-style: italic;
    text-align: left;
  }

  .markdown-preview :global(.task-item) {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 4px 0;
    text-align: left;
  }

  .markdown-preview :global(.task-item.checked) {
    color: var(--c-text-secondary);
    text-decoration: line-through;
  }

  .markdown-preview :global(.empty-preview) {
    color: var(--c-text-secondary);
    opacity: 0.6;
    font-style: italic;
  }
</style>
