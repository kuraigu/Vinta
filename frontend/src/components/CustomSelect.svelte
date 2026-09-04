<script lang="ts">
  import { slide } from 'svelte/transition'
  import { ChevronDown, Search } from '@lucide/svelte'

  let {
    options,
    value = $bindable(),
    placeholder = 'Select...',
    fullWidth = false,
    direction = 'auto',
    searchable = false,
    searchPlaceholder = 'Search...',
    onSearch,
    onSelectPayload,
    onChange
  } = $props<{
    options: { value: string; label: string; payload?: unknown }[]
    value: string
    placeholder?: string
    fullWidth?: boolean
    direction?: 'auto' | 'down' | 'up'
    searchable?: boolean
    searchPlaceholder?: string
    onSearch?: (query: string) => Promise<void> | void
    onSelectPayload?: (payload: unknown) => void
    onChange?: (value: string) => void
  }>()

  let isOpen = $state(false)
  let searchTerm = $state('')
  let wrapperNode = $state<HTMLElement>()
  let dropdownRect = $state({ top: 0, left: 0, width: 0, dropUp: false })
  let searchInputNode = $state<HTMLInputElement>()

  let filteredOptions = $derived.by(() => {
    if (!searchable || onSearch || !searchTerm.trim()) return options
    const query = searchTerm.toLowerCase().trim()
    return options.filter((o) => o.label.toLowerCase().includes(query))
  })

  function handleSearchInput() {
    if (onSearch) {
      onSearch(searchTerm)
    }
  }

  function updateRect() {
    if (!wrapperNode) return
    const rect = wrapperNode.getBoundingClientRect()
    const spaceBelow = window.innerHeight - rect.bottom
    const isUp =
      direction === 'up' ||
      (direction === 'auto' && spaceBelow < 220 && rect.top > 150)

    dropdownRect = {
      top: isUp ? rect.top - 4 : rect.bottom + 4,
      left: rect.left,
      width: rect.width,
      dropUp: isUp
    }
  }

  function toggle(e: MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    if (!isOpen) {
      updateRect()
      searchTerm = ''
      setTimeout(() => {
        if (searchInputNode) searchInputNode.focus()
      }, 50)
    }
    isOpen = !isOpen
    if (!isOpen) {
      notifyOverlay(false)
    }
  }

  function select(
    opt: { value: string; label: string; payload?: unknown },
    e: MouseEvent
  ) {
    e.preventDefault()
    e.stopPropagation()
    value = opt.value
    if (onSelectPayload && opt.payload) {
      onSelectPayload(opt.payload)
    }
    if (onChange) {
      onChange(opt.value)
    }
    isOpen = false
    notifyOverlay(false)
  }

  function notifyOverlay(
    open: boolean,
    rect?: { x: number; y: number; w: number; h: number }
  ) {
    window.dispatchEvent(
      new CustomEvent('vinta-overlay', {
        detail: open ? { open, ...rect } : { open }
      })
    )
  }

  let selectedLabel = $derived(
    options.find((o) => o.value === value)?.label || value || placeholder
  )

  function portal(node: HTMLElement) {
    document.body.appendChild(node)
    updateRect()
    // Read the actual rendered position+size of the dropdown and tell App.svelte
    const r = node.getBoundingClientRect()
    notifyOverlay(true, { x: r.left, y: r.top, w: r.width, h: r.height })
    return {
      destroy() {
        notifyOverlay(false)
        if (node.parentElement) {
          node.parentElement.removeChild(node)
        }
      }
    }
  }

  function clickOutside(node: HTMLElement) {
    let listenerBound = false

    const handleClick = (e: MouseEvent) => {
      const target = e.target as Node
      const dropdownEl = document.querySelector('.select-dropdown.portaled')
      if (
        node.contains(target) ||
        (wrapperNode && wrapperNode.contains(target)) ||
        (dropdownEl && dropdownEl.contains(target))
      ) {
        return
      }
      if (isOpen) {
        isOpen = false
        notifyOverlay(false)
      }
    }

    const handleResize = () => {
      if (isOpen) {
        isOpen = false
        notifyOverlay(false)
      }
    }

    // Bind event listeners on next tick after opening so the opening click event doesn't immediately close it
    const timer = setTimeout(() => {
      document.addEventListener('pointerdown', handleClick, true)
      window.addEventListener('resize', handleResize, true)
      listenerBound = true
    }, 50)

    return {
      destroy() {
        clearTimeout(timer)
        if (listenerBound) {
          document.removeEventListener('pointerdown', handleClick, true)
          window.removeEventListener('resize', handleResize, true)
        }
      }
    }
  }

  function stopPropagation(e: Event) {
    e.stopPropagation()
  }
</script>

<div
  bind:this={wrapperNode}
  class="custom-select-wrapper"
  class:full-width={fullWidth}
>
  <button
    class="select-trigger"
    onclick={toggle}
    type="button"
    aria-expanded={isOpen}
  >
    <span class="label" title={selectedLabel}>{selectedLabel}</span>
    <ChevronDown size={14} class="chevron {isOpen ? 'open' : ''}" />
  </button>

  {#if isOpen}
    <div
      use:portal
      use:clickOutside
      class="select-dropdown portaled no-swipe"
      transition:slide={{ duration: 120 }}
      style="
        position: fixed;
        left: {dropdownRect.left}px;
        width: {dropdownRect.width}px;
        {dropdownRect.dropUp
        ? `bottom: ${window.innerHeight - dropdownRect.top}px;`
        : `top: ${dropdownRect.top}px;`}
        z-index: 999999;
      "
      onwheel={stopPropagation}
    >
      {#if searchable}
        <div class="search-input-box">
          <Search size={13} class="search-icon" />
          <input
            bind:this={searchInputNode}
            type="text"
            class="search-field"
            placeholder={searchPlaceholder}
            bind:value={searchTerm}
            oninput={handleSearchInput}
            onclick={stopPropagation}
            onpointerdown={stopPropagation}
            onkeydown={stopPropagation}
          />
        </div>
      {/if}

      <div class="options-scroll-container">
        {#each filteredOptions as opt (opt.value)}
          <button
            class="option-item"
            class:selected={opt.value === value}
            onclick={(e) => select(opt, e)}
            onpointerdown={stopPropagation}
            type="button"
          >
            {opt.label}
          </button>
        {:else}
          <div class="no-options">No results found</div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .custom-select-wrapper {
    position: relative;
    display: inline-block;
    width: auto;
    pointer-events: auto;
  }

  .custom-select-wrapper.full-width {
    width: 100%;
  }

  .select-trigger {
    background: var(--c-contrast-bg, rgba(0, 0, 0, 0.3));
    color: var(--c-text-primary);
    border: 1px solid var(--c-border);
    padding: 6px 10px;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 500;
    outline: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 8px;
    justify-content: space-between;
    min-width: 140px;
    width: 100%;
    box-sizing: border-box;
    transition:
      border-color 0.2s,
      background-color 0.2s;
    pointer-events: auto;
  }

  :global([data-theme='light']) .select-trigger {
    background: rgba(255, 255, 255, 0.85);
  }

  .select-trigger:focus,
  .custom-select-wrapper:focus-within .select-trigger,
  .select-trigger:hover {
    border-color: var(--c-primary);
  }

  .label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: left;
    flex: 1;
  }

  :global(.chevron) {
    transition: transform 0.2s ease;
    flex-shrink: 0;
  }

  :global(.chevron.open) {
    transform: rotate(180deg);
  }

  :global(.select-dropdown.portaled) {
    pointer-events: auto !important;
    user-select: auto;
    max-height: 220px;
    background: var(--c-bg);
    border: 1px solid var(--c-border);
    border-radius: 8px;
    z-index: 999999;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
    display: flex;
    flex-direction: column;
    padding: 6px;
    box-sizing: border-box;
  }

  :global(.search-input-box) {
    display: flex;
    align-items: center;
    gap: 6px;
    background: var(--c-item-bg, rgba(255, 255, 255, 0.05));
    border: 1px solid var(--c-border);
    border-radius: 6px;
    padding: 4px 8px;
    margin-bottom: 6px;
    flex-shrink: 0;
  }

  :global(.search-icon) {
    color: var(--c-text-muted);
    flex-shrink: 0;
  }

  :global(.search-field) {
    flex: 1;
    background: transparent;
    border: none;
    outline: none;
    color: var(--c-text-primary);
    font-size: 0.78rem;
    width: 100%;
  }

  :global(.options-scroll-container) {
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 2px;
    max-height: 170px;
  }

  :global(.options-scroll-container::-webkit-scrollbar) {
    width: 4px;
  }
  :global(.options-scroll-container::-webkit-scrollbar-track) {
    background: transparent;
  }
  :global(.options-scroll-container::-webkit-scrollbar-thumb) {
    background: var(--c-border);
    border-radius: 4px;
  }

  :global(.no-options) {
    padding: 8px;
    font-size: 0.78rem;
    color: var(--c-text-muted);
    text-align: center;
  }

  .option-item {
    background: transparent;
    border: none;
    color: var(--c-text-secondary);
    padding: 6px 10px;
    text-align: left;
    font-size: 0.8rem;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.15s ease;
    flex-shrink: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    pointer-events: auto;
  }

  .option-item:hover {
    background: var(--c-hover);
    color: var(--c-text-primary);
  }

  .option-item.selected {
    background: var(--c-item-bg);
    color: var(--c-primary);
    font-weight: 600;
  }
</style>
