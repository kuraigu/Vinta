<script lang="ts">
  /**
   * Reusable segmented pill selector primitive.
   */
  export interface SegmentOption {
    value: string
    label: string
    disabled?: boolean
  }

  interface Props {
    options: SegmentOption[]
    value?: string
    size?: 'sm' | 'md'
    fullWidth?: boolean
    onchange?: (value: string) => void
    class?: string
  }

  let {
    options = [],
    value = $bindable(''),
    size = 'md',
    fullWidth = false,
    onchange,
    class: className = ''
  }: Props = $props()

  function selectOption(optValue: string) {
    if (value === optValue) return
    value = optValue
    onchange?.(optValue)
  }
</script>

<div
  class="vinta-segmented-control size-{size} {className}"
  class:full-width={fullWidth}
  role="radiogroup"
>
  {#each options as option (option.value)}
    <button
      type="button"
      role="radio"
      aria-checked={value === option.value}
      disabled={option.disabled}
      class="vinta-segment-btn"
      class:active={value === option.value}
      onclick={() => selectOption(option.value)}
    >
      <span class="vinta-segment-label">{option.label}</span>
    </button>
  {/each}
</div>

<style src="./styles/SegmentedControl.css"></style>
