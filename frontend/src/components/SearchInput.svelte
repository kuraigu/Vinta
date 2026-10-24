<script lang="ts">
  import { Search, X } from 'lucide-svelte'

  interface Props {
    value: string
    placeholder?: string
    variant?: 'pill' | 'rounded'
    size?: 'sm' | 'md'
    maxWidth?: string
    class?: string
  }

  let {
    value = $bindable(''),
    placeholder = 'Search...',
    variant = 'pill',
    size = 'md',
    maxWidth,
    class: className = ''
  }: Props = $props()

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
</script>

<div
  class="vinta-search-input variant-{variant} size-{size} {className}"
  style:max-width={maxWidth}
>
  <Search size={13} class="vinta-search-icon" />
  <input
    type="text"
    {placeholder}
    bind:value
    onclick={(e) => e.stopPropagation()}
    onkeydown={handleKeyDown}
  />
  {#if value.trim().length > 0}
    <button
      type="button"
      class="vinta-clear-search-btn"
      aria-label="Clear search query"
      onclick={handleClear}
    >
      <X size={11} />
    </button>
  {/if}
</div>

<style>
  .vinta-search-input {
    display: flex;
    align-items: center;
    gap: 6px;
    background: var(--vinta-item-bg, rgba(255, 255, 255, 0.06));
    border: 1px solid var(--vinta-border, rgba(255, 255, 255, 0.08));
    padding: 4px 10px;
    flex: 1;
    width: 100%;
    box-sizing: border-box;
    transition: all 0.18s ease;
  }

  .vinta-search-input.variant-pill {
    border-radius: var(--vinta-radius-pill, 9999px);
  }

  .vinta-search-input.variant-rounded {
    border-radius: 8px;
  }

  .vinta-search-input.size-sm {
    height: 26px;
    padding: 0 9px;
  }

  .vinta-search-input.size-md {
    height: 28px;
    padding: 0 10px;
  }

  .vinta-search-input:focus-within {
    border-color: var(--vinta-border-accent, rgba(255, 255, 255, 0.24));
    background: var(--vinta-contrast-hover, rgba(255, 255, 255, 0.09));
    box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.08);
  }

  :global(.vinta-search-icon) {
    color: var(--vinta-text-muted);
    flex-shrink: 0;
  }

  .vinta-search-input input {
    background: transparent;
    border: none;
    outline: none;
    color: var(--vinta-text-primary);
    font-size: 11.5px;
    width: 100%;
    padding: 0;
  }

  .vinta-search-input input::placeholder {
    color: var(--vinta-text-muted);
  }

  .vinta-clear-search-btn {
    background: transparent;
    border: none;
    padding: 2px;
    margin: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--vinta-text-muted);
    cursor: pointer;
    border-radius: 50%;
    transition:
      color 0.15s ease,
      background 0.15s ease;
    flex-shrink: 0;
  }

  .vinta-clear-search-btn:hover {
    color: var(--vinta-text-primary);
    background: var(--vinta-hover, rgba(255, 255, 255, 0.1));
  }
</style>
