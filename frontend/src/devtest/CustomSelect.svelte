<script lang="ts">
  import { slide } from 'svelte/transition'
  import { ChevronDown } from 'lucide-svelte'

  let {
    options,
    value = $bindable(),
    placeholder = 'Select...'
  } = $props<{
    options: { value: string; label: string }[]
    value: string
    placeholder?: string
  }>()

  let isOpen = $state(false)

  function toggle() {
    isOpen = !isOpen
    triggerBoundsUpdate(isOpen)
  }

  function select(val: string) {
    value = val
    isOpen = false
    triggerBoundsUpdate(false)
  }

  function triggerBoundsUpdate(open: boolean) {
    // Notify DevPill to expand its OS-level click region to encompass the absolute dropdown
    const event = new CustomEvent('dev-dropdown-toggle', { detail: open })
    window.dispatchEvent(event)
  }

  let selectedLabel = $derived(
    options.find((o) => o.value === value)?.label || placeholder
  )

  function clickOutside(node: HTMLElement) {
    const handleClick = (e: MouseEvent) => {
      if (!node.contains(e.target as Node)) {
        if (isOpen) {
          isOpen = false
          triggerBoundsUpdate(false)
        }
      }
    }
    document.addEventListener('click', handleClick, true)
    return {
      destroy() {
        document.removeEventListener('click', handleClick, true)
      }
    }
  }
</script>

<div class="custom-select-wrapper" use:clickOutside>
  <button class="select-trigger" onclick={toggle} type="button">
    <span class="label">{selectedLabel}</span>
    <ChevronDown size={14} />
  </button>

  {#if isOpen}
    <div class="select-dropdown" transition:slide={{ duration: 150 }}>
      {#each options as opt (opt.value)}
        <button
          class="option-item"
          class:selected={opt.value === value}
          onclick={() => select(opt.value)}
          type="button"
        >
          {opt.label}
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .custom-select-wrapper {
    position: relative;
    display: inline-block;
  }

  .select-trigger {
    background: var(--c-bg);
    color: var(--c-text-primary);
    border: 1px solid var(--c-border);
    padding: 6px 8px;
    border-radius: 6px;
    font-size: 0.8rem;
    outline: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 8px;
    justify-content: space-between;
    min-width: 140px;
    transition: border-color 0.2s;
  }

  .select-trigger:focus,
  .custom-select-wrapper:focus-within .select-trigger {
    border-color: var(--c-primary);
  }

  .label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 150px;
    text-align: left;
  }

  .select-dropdown {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    min-width: 100%;
    max-height: 200px;
    overflow-y: auto;
    background: var(--c-bg);
    border: 1px solid var(--c-border);
    border-radius: 6px;
    z-index: 99999; /* Pure HTML z-index superiority */
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    display: flex;
    flex-direction: column;
    padding: 4px;
    box-sizing: border-box;
  }

  .select-dropdown::-webkit-scrollbar {
    width: 4px;
  }
  .select-dropdown::-webkit-scrollbar-track {
    background: transparent;
  }
  .select-dropdown::-webkit-scrollbar-thumb {
    background: var(--c-scrollbar-thumb);
    border-radius: 4px;
  }

  .option-item {
    background: transparent;
    border: none;
    color: var(--c-text-secondary);
    padding: 6px 8px;
    text-align: left;
    font-size: 0.8rem;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.1s;
    flex-shrink: 0;
  }

  .option-item:hover {
    background: var(--c-hover);
    color: var(--c-text-primary);
  }

  .option-item.selected {
    background: var(--c-item-bg);
    color: var(--c-text-primary);
    font-weight: 500;
  }
</style>
