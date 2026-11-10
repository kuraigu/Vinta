<!--
  @component SearchInput
  @description Filter text input with search icon, clear button, and debounced input
  events for real-time list filtering.
-->
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

<style src="./styles/SearchInput.css"></style>
