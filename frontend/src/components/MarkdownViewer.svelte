<!--
  @component MarkdownViewer
  @description Renders markdown text into styled HTML for release notes, changelogs, and announcements.
-->
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

    // 1. Normalize line endings
    let text = raw.replace(/\r\n/g, '\n').replace(/\r/g, '\n')

    // 2. Escape basic HTML entities while preserving markdown
    text = text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')

    // 3. Restore and sanitize safe HTML elements (details, summary, b, i, img, br)
    text = text
      .replace(/&lt;details&gt;/gi, '<details class="md-details">')
      .replace(/&lt;\/details&gt;/gi, '</details>')
      .replace(/&lt;summary&gt;/gi, '<summary class="md-summary">')
      .replace(/&lt;\/summary&gt;/gi, '</summary>')
      .replace(/&lt;b&gt;/gi, '<strong>')
      .replace(/&lt;\/b&gt;/gi, '</strong>')
      .replace(/&lt;i&gt;/gi, '<em>')
      .replace(/&lt;\/i&gt;/gi, '</em>')
      .replace(/&lt;br\s*\/?&gt;/gi, '<br/>')
      .replace(
        /&lt;img\s+([^&]*?src=["'](https?:\/\/[^"']+)["'][^&]*?)&gt;/gi,
        '<img class="md-image" $1 loading="lazy" />'
      )

    // 4. Code Blocks (```code```)
    text = text.replace(
      /```([\s\S]*?)```/g,
      '<pre class="md-code-block"><code>$1</code></pre>'
    )

    // 5. Inline Code (`code`)
    text = text.replace(
      /`([^`\n]+)`/g,
      '<code class="md-inline-code">$1</code>'
    )

    // 6. Markdown Images & Badges: ![alt](url)
    text = text.replace(
      /!\[([^\]]*)\]\((https?:\/\/[^\s)]+)(?:\s+["'][^"']*["'])?\)/g,
      (_, alt, url) => {
        const isBadge =
          url.includes('img.shields.io') ||
          url.includes('badge') ||
          url.includes('shields')
        if (isBadge) {
          return `<img src="${url}" alt="${alt}" class="md-badge" loading="lazy" />`
        }
        return `<img src="${url}" alt="${alt}" class="md-image" loading="lazy" />`
      }
    )

    // 7. Headings: ### H3, ## H2, # H1
    text = text.replace(/^### (.*$)/gm, '<h3 class="md-h3">$1</h3>')
    text = text.replace(/^## (.*$)/gm, '<h2 class="md-h2">$1</h2>')
    text = text.replace(/^# (.*$)/gm, '<h1 class="md-h1">$1</h1>')

    // 8. Horizontal Rules
    text = text.replace(/^---+$/gm, '<hr class="md-hr" />')

    // 8b. Blockquotes: > quote
    text = text.replace(
      /^&gt;\s?(.*$)/gm,
      '<blockquote class="md-blockquote">$1</blockquote>'
    )
    text = text.replace(
      /((?:<blockquote class="md-blockquote">.*?<\/blockquote>\n*)+)/g,
      '<div class="md-blockquote-group">$1</div>'
    )

    // 9. Markdown Links: [text](url)
    text = text.replace(
      /\[([^\]\n]+)\]\((https?:\/\/[^\s)]+)\)/g,
      '<a href="$2" target="_blank" rel="noreferrer" class="md-link">$1</a>'
    )

    // 10. Bare HTTP/HTTPS URLs (not preceded by href=" or src=" or >)
    text = text.replace(
      /(?<!href=")(?<!src=")(?<!">)(https?:\/\/[^\s<)"']+)/g,
      '<a href="$1" target="_blank" rel="noreferrer" class="md-link">$1</a>'
    )

    // 11. Unordered Lists: * item or - item (MUST run before italic to avoid stealing leading asterisks)
    text = text.replace(
      /^\s*[-*]\s+(.*$)/gm,
      '<li class="md-list-item">$1</li>'
    )

    // 12. Group consecutive <li> into <ul>
    text = text.replace(
      /((?:<li class="md-list-item">.*<\/li>\n*)+)/g,
      '<ul class="md-list">$1</ul>'
    )

    // 13. Bold: **text**
    text = text.replace(/\*\*([^*\n]+)\*\*/g, '<strong>$1</strong>')

    // 14. Italic: *text* (word-bounded, single line only)
    text = text.replace(/(?<!\*)\*([^*\n]+)\*(?!\*)/g, '<em>$1</em>')

    // 15. Clean block paragraphs
    const blocks = text.split(/\n{2,}/)
    return blocks
      .map((b) => {
        const trimmed = b.trim()
        if (!trimmed) return ''
        if (
          trimmed.startsWith('<h1') ||
          trimmed.startsWith('<h2') ||
          trimmed.startsWith('<h3') ||
          trimmed.startsWith('<pre') ||
          trimmed.startsWith('<ul') ||
          trimmed.startsWith('<ol') ||
          trimmed.startsWith('<details') ||
          trimmed.startsWith('<div class="md-blockquote-group') ||
          trimmed.startsWith('<hr')
        ) {
          return trimmed
        }
        if (trimmed.startsWith('<img') && !trimmed.includes('\n')) {
          return `<div class="md-badge-row">${trimmed}</div>`
        }
        return `<p class="md-paragraph">${trimmed.replace(/\n/g, '<br/>')}</p>`
      })
      .filter(Boolean)
      .join('\n')
  }

  let htmlContent = $derived(renderMarkdown(content))

  function handleContainerClick(e: MouseEvent) {
    const linkTarget = (e.target as HTMLElement).closest('a.md-link')
    if (linkTarget) {
      e.preventDefault()
      const href = linkTarget.getAttribute('href')
      if (href) {
        Browser.OpenURL(href).catch(() => window.open(href, '_blank'))
      }
      return
    }

    const imgTarget = (e.target as HTMLElement).closest('img.md-image')
    if (imgTarget) {
      const src = imgTarget.getAttribute('src')
      if (src && !src.startsWith('data:')) {
        Browser.OpenURL(src).catch(() => window.open(src, '_blank'))
      }
    }
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div class="markdown-viewer-container" onclick={handleContainerClick}>
  <!-- eslint-disable-next-line svelte/no-at-html-tags -->
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
    cursor: pointer;
    transition:
      transform 0.15s ease,
      opacity 0.15s ease;
  }

  :global(.markdown-viewer-container .md-image:hover) {
    opacity: 0.95;
    transform: scale(1.008);
  }

  :global(.markdown-viewer-container .md-blockquote-group) {
    margin: 6px 0;
  }

  :global(.markdown-viewer-container .md-blockquote) {
    margin: 3px 0;
    padding: 4px 10px;
    border-left: 3px solid var(--vinta-accent-primary, #38bdf8);
    background: var(--vinta-surface-elevated, rgba(255, 255, 255, 0.03));
    border-radius: 0 6px 6px 0;
    font-style: italic;
    color: var(--vinta-text-secondary);
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

  :global(.markdown-viewer-container .md-list) {
    margin: 4px 0 8px 0;
    padding-left: 20px;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  :global(.markdown-viewer-container .md-list-item) {
    list-style-type: disc;
    line-height: 1.45;
    text-align: left;
  }

  :global(.markdown-viewer-container .md-badge-row) {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    align-items: center;
    margin: 2px 0 6px;
  }

  :global(.markdown-viewer-container .md-badge) {
    height: 20px;
    object-fit: contain;
    border: none;
    background: transparent;
    border-radius: 3px;
  }

  :global(.markdown-viewer-container .md-details) {
    margin: 8px 0;
    padding: 8px 12px;
    border-radius: 8px;
    background: var(--vinta-surface-elevated, rgba(255, 255, 255, 0.04));
    border: 1px solid var(--vinta-border);
  }

  :global(.markdown-viewer-container .md-summary) {
    cursor: pointer;
    font-weight: 600;
    color: var(--vinta-text-primary);
    user-select: none;
    margin-bottom: 6px;
  }

  :global(.markdown-viewer-container .md-hr) {
    border: none;
    border-top: 1px solid var(--vinta-border);
    margin: 12px 0;
  }
</style>
