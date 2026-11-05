<script lang="ts">
  import { Browser } from '@wailsio/runtime'

  let {
    content = ''
  }: {
    content?: string
  } = $props()

  /**
   * Sanitizes and parses markdown text into safe HTML with image and link support.
   */
  function renderMarkdown(raw: string): string {
    if (!raw) return ''

    let text = raw
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')

    // Allow and decode safe inline images: <img src="..." alt="..." />
    text = text.replace(
      /&lt;img\s+([^&]*?src=["'](https?:\/\/[^"']+)["'][^&]*?)&gt;/gi,
      '<img $1 loading="lazy" />'
    )

    // Markdown Images: ![alt](url)
    text = text.replace(
      /!\[([^\]]*)\]\((https?:\/\/[^\s)]+)\)/g,
      '<img src="$2" alt="$1" class="md-image" loading="lazy" />'
    )

    // Code Blocks: ```code```
    text = text.replace(/```([\s\S]*?)```/g, '<pre class="md-code-block"><code>$1</code></pre>')

    // Inline Code: `code`
    text = text.replace(/`([^`]+)`/g, '<code class="md-inline-code">$1</code>')

    // Headings: ### H3, ## H2, # H1
    text = text.replace(/^### (.*$)/gim, '<h3 class="md-h3">$1</h3>')
    text = text.replace(/^## (.*$)/gim, '<h2 class="md-h2">$1</h2>')
    text = text.replace(/^# (.*$)/gim, '<h1 class="md-h1">$1</h1>')

    // Bold: **text**
    text = text.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')

    // Italic: *text*
    text = text.replace(/\*([^*]+)\*/g, '<em>$1</em>')

    // Markdown Links: [text](url)
    text = text.replace(
      /\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g,
      '<a href="$2" target="_blank" rel="noreferrer" class="md-link">$1</a>'
    )

    // Unordered Lists: - item or * item
    text = text.replace(/^\s*[-*]\s+(.*$)/gim, '<li class="md-list-item">$1</li>')

    // Paragraph line breaks (double newline to paragraph gap)
    const paragraphs = text.split(/\n{2,}/)
    return paragraphs
      .map((p) => {
        const trimmed = p.trim()
        if (
          trimmed.startsWith('<h1') ||
          trimmed.startsWith('<h2') ||
          trimmed.startsWith('<h3') ||
          trimmed.startsWith('<pre') ||
          trimmed.startsWith('<li')
        ) {
          return trimmed
        }
        return `<p class="md-paragraph">${trimmed.replace(/\n/g, '<br/>')}</p>`
      })
      .join('\n')
  }

  let htmlContent = $derived(renderMarkdown(content))

  function handleContainerClick(e: MouseEvent) {
    const target = (e.target as HTMLElement).closest('a.md-link')
    if (target) {
      e.preventDefault()
      const href = target.getAttribute('href')
      if (href) {
        Browser.OpenURL(href).catch(() => window.open(href, '_blank'))
      }
    }
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="markdown-viewer-container" onclick={handleContainerClick}>
  {@html htmlContent}
</div>

<style>
  .markdown-viewer-container {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;
    color: var(--vinta-text-secondary);
    font-size: 0.78rem;
    line-height: 1.5;
    text-align: left;
    word-break: break-word;
  }

  :global(.markdown-viewer-container .md-h1) {
    margin: 8px 0 4px;
    font-size: 1.05rem;
    font-weight: 700;
    color: var(--vinta-text-primary);
    text-align: left;
  }

  :global(.markdown-viewer-container .md-h2) {
    margin: 8px 0 4px;
    font-size: 0.92rem;
    font-weight: 700;
    color: var(--vinta-text-primary);
    text-align: left;
  }

  :global(.markdown-viewer-container .md-h3) {
    margin: 6px 0 2px;
    font-size: 0.82rem;
    font-weight: 700;
    color: var(--vinta-text-primary);
    text-align: left;
  }

  :global(.markdown-viewer-container .md-paragraph) {
    margin: 0 0 6px;
    line-height: 1.5;
    text-align: left;
  }

  :global(.markdown-viewer-container .md-image) {
    max-width: 100%;
    height: auto;
    border-radius: 8px;
    margin: 8px 0;
    object-fit: contain;
    border: 1px solid var(--vinta-border);
    background: rgba(0, 0, 0, 0.2);
  }

  :global(.markdown-viewer-container .md-link) {
    color: var(--vinta-accent-primary, #38bdf8);
    text-decoration: underline;
    text-underline-offset: 2px;
    cursor: pointer;
  }

  :global(.markdown-viewer-container .md-inline-code) {
    font-family: inherit;
    font-size: 0.74rem;
    padding: 1px 5px;
    border-radius: 4px;
    background: var(--vinta-surface-elevated);
    border: 1px solid var(--vinta-border);
    color: var(--vinta-text-primary);
  }

  :global(.markdown-viewer-container .md-code-block) {
    font-family: inherit;
    font-size: 0.74rem;
    padding: 10px 12px;
    border-radius: 8px;
    background: rgba(0, 0, 0, 0.35);
    border: 1px solid var(--vinta-border);
    color: var(--vinta-text-primary);
    overflow-x: auto;
    margin: 6px 0;
  }

  :global(.markdown-viewer-container .md-list-item) {
    margin-left: 18px;
    list-style-type: disc;
    text-align: left;
  }
</style>

