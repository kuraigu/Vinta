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

  interface Props {
    onInsertFormatting: (prefix: string, suffix?: string) => void
  }

  let { onInsertFormatting }: Props = $props()
</script>

<div class="toolbar">
  <button
    class="tool-btn"
    title="Bold (**text**)"
    onclick={() => onInsertFormatting('**', '**')}
  >
    <Bold size={13} />
  </button>
  <button
    class="tool-btn"
    title="Italic (*text*)"
    onclick={() => onInsertFormatting('*', '*')}
  >
    <Italic size={13} />
  </button>
  <button
    class="tool-btn"
    title="Strikethrough (~~text~~)"
    onclick={() => onInsertFormatting('~~', '~~')}
  >
    <Strikethrough size={13} />
  </button>

  <div class="tool-divider"></div>

  <button
    class="tool-btn"
    title="Heading 1 (# H1)"
    onclick={() => onInsertFormatting('# ')}
  >
    <Heading1 size={13} />
  </button>
  <button
    class="tool-btn"
    title="Heading 2 (## H2)"
    onclick={() => onInsertFormatting('## ')}
  >
    <Heading2 size={13} />
  </button>
  <button
    class="tool-btn"
    title="Heading 3 (### H3)"
    onclick={() => onInsertFormatting('### ')}
  >
    <Heading3 size={13} />
  </button>

  <div class="tool-divider"></div>

  <button
    class="tool-btn"
    title="Bullet List (- Item)"
    onclick={() => onInsertFormatting('- ')}
  >
    <List size={13} />
  </button>
  <button
    class="tool-btn"
    title="Checklist (- [ ] Task)"
    onclick={() => onInsertFormatting('- [ ] ')}
  >
    <CheckSquare size={13} />
  </button>
  <button
    class="tool-btn"
    title="Link ([text](url))"
    onclick={() => onInsertFormatting('[', '](https://)')}
  >
    <Link size={13} />
  </button>
  <button
    class="tool-btn"
    title="Quote (> Quote)"
    onclick={() => onInsertFormatting('> ')}
  >
    <Quote size={13} />
  </button>
  <button
    class="tool-btn"
    title="Code Block (```lang\ncode\n```)"
    onclick={() => onInsertFormatting('```\n', '\n```')}
  >
    <Code size={13} />
  </button>
</div>

<style>
  .toolbar {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 6px;
    background: rgba(0, 0, 0, 0.2);
    border: 1px solid var(--c-border);
    border-radius: 8px;
    margin-bottom: 8px;
  }

  .tool-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    background: transparent;
    border: none;
    border-radius: 4px;
    color: var(--c-text-secondary);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .tool-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: var(--c-text-primary);
  }

  .tool-divider {
    width: 1px;
    height: 14px;
    background: var(--c-border);
    margin: 0 2px;
  }
</style>
