<!--
  @component CustomSelect
  @description Custom dropdown picker that replaces native HTML <select>. Uses Vinta's
  token styling, keyboard navigation (Arrow Up/Down, Enter, Esc), click-outside dismissal,
  and notifies the native backend when opened so the dropdown menu isn't clipped by the window frame.
-->
<script lang="ts">
  import { slide } from 'svelte/transition'
  import { ChevronDown, Search } from 'lucide-svelte'
  import FlagIcon from './FlagIcon.svelte'

  export interface SelectOption {
    value: string
    label: string
    flag?: string
    payload?: unknown
  }

  interface Props {
    options: SelectOption[]
    value: string
    placeholder?: string
    fullWidth?: boolean
    direction?: 'auto' | 'down' | 'up'
    searchable?: boolean
    searchPlaceholder?: string
    onSearch?: (query: string) => Promise<void> | void
    onSelectPayload?: (payload: unknown) => void
    onChange?: (value: string) => void
  }

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
  }: Props = $props()

  let isOpen = $state(false)
  let searchTerm = $state('')
  let wrapperNode = $state<HTMLElement>()
  let dropdownRect = $state({
    top: 0,
    left: 0,
    width: 0,
    dropUp: false,
    maxHeight: 220
  })
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
    const island =
      wrapperNode.closest('.island') ||
      wrapperNode.closest('.island-wrapper') ||
      wrapperNode.closest('.island-container')
    const containerRect = island
      ? island.getBoundingClientRect()
      : {
          top: 0,
          bottom: window.innerHeight,
          left: 0,
          right: window.innerWidth
        }

    const spaceBelow = containerRect.bottom - rect.bottom
    const spaceAbove = rect.top - containerRect.top
    const isUp =
      direction === 'up' ||
      (direction === 'auto' && spaceBelow < 220 && spaceAbove > spaceBelow)

    const availableSpace = isUp ? spaceAbove : spaceBelow
    const calculatedMaxHeight = Math.max(
      120,
      Math.min(220, availableSpace - 12)
    )

    dropdownRect = {
      top: isUp ? rect.top - 4 : rect.bottom + 4,
      left: rect.left,
      width: rect.width,
      dropUp: isUp,
      maxHeight: calculatedMaxHeight
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

  let selectedOption = $derived(options.find((o) => o.value === value))
  let selectedLabel = $derived(selectedOption?.label || value || placeholder)

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
    <div class="trigger-content">
      {#if selectedOption?.flag}
        <FlagIcon code={selectedOption.flag} size={14} />
      {/if}
      <span class="label" title={selectedLabel}>{selectedLabel}</span>
    </div>
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
        max-height: {dropdownRect.maxHeight}px;
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
            {#if opt.flag}
              <FlagIcon code={opt.flag} size={14} />
            {/if}
            <span class="option-label">{opt.label}</span>
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
    background: var(--vinta-contrast-bg, rgba(0, 0, 0, 0.3));
    color: var(--vinta-text-primary);
    border: 1px solid var(--vinta-border);
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
    border-color: var(--vinta-primary);
  }

  .trigger-content {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    overflow: hidden;
    flex: 1;
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
    background: var(--vinta-bg);
    border: 1px solid var(--vinta-border);
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
    background: var(--vinta-item-bg, rgba(255, 255, 255, 0.05));
    border: 1px solid var(--vinta-border);
    border-radius: 6px;
    padding: 4px 8px;
    margin-bottom: 6px;
    flex-shrink: 0;
  }

  :global(.search-icon) {
    color: var(--vinta-text-muted);
    flex-shrink: 0;
  }

  :global(.search-field) {
    flex: 1;
    background: transparent;
    border: none;
    outline: none;
    color: var(--vinta-text-primary);
    font-size: 0.78rem;
    width: 100%;
  }

  :global(.options-scroll-container) {
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 2px;
    flex: 1;
    min-height: 0;
  }

  :global(.options-scroll-container::-webkit-scrollbar) {
    width: 4px;
  }
  :global(.options-scroll-container::-webkit-scrollbar-track) {
    background: transparent;
  }
  :global(.options-scroll-container::-webkit-scrollbar-thumb) {
    background: var(--vinta-border);
    border-radius: 4px;
  }

  :global(.no-options) {
    padding: 8px;
    font-size: 0.78rem;
    color: var(--vinta-text-muted);
    text-align: center;
  }

  .option-item {
    display: flex;
    align-items: center;
    gap: 8px;
    background: transparent;
    border: none;
    color: var(--vinta-text-secondary);
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

  .option-label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }

  .option-item:hover {
    background: var(--vinta-hover);
    color: var(--vinta-text-primary);
  }

  .option-item.selected {
    background: var(--vinta-item-bg);
    color: var(--vinta-primary);
    font-weight: 600;
  }
</style>
