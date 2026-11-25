<!--
  @component SearchInput
  @description Filter text input with search icon, clear button, right-click text context
  menu, and debounced input events for real-time list filtering.
-->
<script lang="ts">
  import { Search, X } from 'lucide-svelte'
  import CustomContextMenu, {
    type ContextMenuItem
  } from './CustomContextMenu.svelte'
  import { buildTextEditMenuItems } from '../utils/textContextMenu'
  import { tStore } from '../store'

  interface Props {
    value: string
    placeholder?: string
    variant?: 'pill' | 'rounded'
    size?: 'sm' | 'md'
    maxWidth?: string
    enableShortcut?: boolean
    class?: string
  }

  let {
    value = $bindable(''),
    placeholder = 'Search...',
    variant = 'pill',
    size = 'md',
    maxWidth,
    enableShortcut = true,
    class: className = ''
  }: Props = $props()

  let inputEl = $state<HTMLInputElement | null>(null)
  let ctxCoords = $state<{ x: number; y: number } | null>(null)
  let ctxItems = $state<ContextMenuItem[]>([])

  /**
   * Focuses the underlying search input element.
   */
  export function focus() {
    inputEl?.focus()
  }

  /**
   * Clears the active search input when Escape is pressed.
   *
   * @param e - Keyboard event from the input element.
   */
  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      value = ''
    }
  }

  /**
   * Clears the search query via the clear button.
   *
   * @param e - Mouse click event.
   */
  function handleClear(e: MouseEvent) {
    e.stopPropagation()
    value = ''
  }

  function handleContextMenu(e: MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    ctxItems = buildTextEditMenuItems(
      inputEl,
      {
        cut: $tStore('context_menu.cut'),
        copy: $tStore('context_menu.copy'),
        paste: $tStore('context_menu.paste'),
        selectAll: $tStore('context_menu.select_all'),
        clear: $tStore('context_menu.clear')
      },
      () => {
        value = ''
        inputEl?.focus()
      }
    )
    ctxCoords = { x: e.clientX, y: e.clientY }
  }
  function handleWindowKeyDown(e: KeyboardEvent) {
    if (!enableShortcut) return
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'f') {
      e.preventDefault()
      inputEl?.focus()
      return
    }
    if (e.key === '/') {
      const target = e.target as HTMLElement | null
      const tag = target?.tagName?.toLowerCase()
      if (tag === 'input' || tag === 'textarea' || target?.isContentEditable) {
        return
      }
      e.preventDefault()
      inputEl?.focus()
    }
  }
</script>

<svelte:window onkeydown={handleWindowKeyDown} />

<div
  class="vinta-search-input variant-{variant} size-{size} {className}"
  style:max-width={maxWidth}
>
  <Search size={14} class="vinta-search-icon" />
  <input
    bind:this={inputEl}
    type="text"
    {placeholder}
    bind:value
    onclick={(e) => e.stopPropagation()}
    onkeydown={handleKeyDown}
    oncontextmenu={handleContextMenu}
  />
  {#if value.trim().length > 0}
    <button
      type="button"
      class="vinta-clear-search-btn"
      aria-label="Clear search query"
      onclick={handleClear}
    >
      <X size={12} />
    </button>
  {/if}
</div>

{#if ctxCoords}
  <CustomContextMenu
    x={ctxCoords.x}
    y={ctxCoords.y}
    items={ctxItems}
    onclose={() => (ctxCoords = null)}
  />
{/if}

<style src="./styles/SearchInput.css"></style>
