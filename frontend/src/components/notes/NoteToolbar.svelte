<script lang="ts">
  import {
    Bold,
    Italic,
    Strikethrough,
    Heading1,
    Heading2,
    Heading3,
    List,
    CheckSquare,
    Link,
    Code,
    Quote
  } from 'lucide-svelte'

  export type FormatType = 'inline' | 'line' | 'link' | 'codeblock'

  interface Props {
    onInsertFormatting: (
      type: FormatType,
      prefix: string,
      suffix?: string
    ) => void
  }

  let { onInsertFormatting }: Props = $props()
</script>

<div class="toolbar">
  <button
    class="tool-btn"
    title="Bold (**text**)"
    onmousedown={(e) => e.preventDefault()}
    onclick={() => onInsertFormatting('inline', '**', '**')}
  >
    <Bold size={13} />
  </button>
  <button
    class="tool-btn"
    title="Italic (*text*)"
    onmousedown={(e) => e.preventDefault()}
    onclick={() => onInsertFormatting('inline', '*', '*')}
  >
    <Italic size={13} />
  </button>
  <button
    class="tool-btn"
    title="Strikethrough (~~text~~)"
    onmousedown={(e) => e.preventDefault()}
    onclick={() => onInsertFormatting('inline', '~~', '~~')}
  >
    <Strikethrough size={13} />
  </button>

  <div class="tool-divider"></div>

  <button
    class="tool-btn"
    title="Heading 1 (# H1)"
    onmousedown={(e) => e.preventDefault()}
    onclick={() => onInsertFormatting('line', '# ')}
  >
    <Heading1 size={13} />
  </button>
  <button
    class="tool-btn"
    title="Heading 2 (## H2)"
    onmousedown={(e) => e.preventDefault()}
    onclick={() => onInsertFormatting('line', '## ')}
  >
    <Heading2 size={13} />
  </button>
  <button
    class="tool-btn"
    title="Heading 3 (### H3)"
    onmousedown={(e) => e.preventDefault()}
    onclick={() => onInsertFormatting('line', '### ')}
  >
    <Heading3 size={13} />
  </button>

  <div class="tool-divider"></div>

  <button
    class="tool-btn"
    title="Bullet List (- Item)"
    onmousedown={(e) => e.preventDefault()}
    onclick={() => onInsertFormatting('line', '- ')}
  >
    <List size={13} />
  </button>
  <button
    class="tool-btn"
    title="Checklist (- [ ] Task)"
    onmousedown={(e) => e.preventDefault()}
    onclick={() => onInsertFormatting('line', '- [ ] ')}
  >
    <CheckSquare size={13} />
  </button>
  <button
    class="tool-btn"
    title="Quote (> Quote)"
    onmousedown={(e) => e.preventDefault()}
    onclick={() => onInsertFormatting('line', '> ')}
  >
    <Quote size={13} />
  </button>
  <button
    class="tool-btn"
    title="Link ([text](url))"
    onmousedown={(e) => e.preventDefault()}
    onclick={() => onInsertFormatting('link', '[', '](https://)')}
  >
    <Link size={13} />
  </button>
  <button
    class="tool-btn"
    title="Code Block (```lang\ncode\n```)"
    onmousedown={(e) => e.preventDefault()}
    onclick={() => onInsertFormatting('codeblock', '```\n', '\n```')}
  >
    <Code size={13} />
  </button>
</div>

<style src="../styles/notes/NoteToolbar.css"></style>
