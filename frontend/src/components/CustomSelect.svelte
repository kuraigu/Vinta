<!--
  @component CustomSelect
  @description Custom dropdown picker that replaces native HTML <select>. Uses Vinta's
  token styling, keyboard navigation (Arrow Up/Down, Enter, Esc), click-outside dismissal,
  and notifies the native backend when opened so the dropdown menu isn't clipped by the window frame.
-->
<script lang="ts">
  import { slide } from 'svelte/transition'
  import { Events } from '$lib/events'
  import { ChevronDown, Search } from 'lucide-svelte'
  import FlagIcon from './FlagIcon.svelte'
  import { portal, getUiScale } from '../utils/portal'
  import { tStore } from '../store'

  export interface SelectOption {
    value: string
    label: string
    flag?: string
    payload?: unknown
    isDefault?: boolean
  }

  interface Props {
    options: SelectOption[]
    value: string
    defaultValue?: string
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
    defaultValue,
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
      onSearch?.('')
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
    Events.Emit('vinta:overlay', open ? { open, ...rect } : { open })
  }

  let selectedOption = $derived(options.find((o) => o.value === value))
  let selectedLabel = $derived(selectedOption?.label || value || placeholder)

  function mountSelectPortal(node: HTMLElement) {
    return portal(node, {
      onMount: () => {
        updateRect()
        const r = node.getBoundingClientRect()
        notifyOverlay(true, { x: r.left, y: r.top, w: r.width, h: r.height })
      },
      onDestroy: () => notifyOverlay(false)
    })
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
      <span class="label">{selectedLabel}</span>
    </div>
    <ChevronDown size={14} class="chevron {isOpen ? 'open' : ''}" />
  </button>

  {#if isOpen}
    {@const uiScale = getUiScale()}
    <div
      use:mountSelectPortal
      use:clickOutside
      class="select-dropdown portaled no-swipe"
      transition:slide={{ duration: 120 }}
      style="
        position: fixed;
        zoom: {uiScale};
        left: {dropdownRect.left / uiScale}px;
        width: {dropdownRect.width / uiScale}px;
        {dropdownRect.dropUp
        ? `bottom: ${(window.innerHeight - dropdownRect.top) / uiScale}px;`
        : `top: ${dropdownRect.top / uiScale}px;`}
        max-height: {dropdownRect.maxHeight / uiScale}px;
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
            {#if opt.isDefault || (defaultValue !== undefined && opt.value === defaultValue)}
              <span class="option-default-tag">{$tStore('common.default')}</span
              >
            {/if}
          </button>
        {:else}
          <div class="no-options">No results found</div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style src="./styles/CustomSelect.css"></style>
