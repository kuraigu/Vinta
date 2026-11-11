<script lang="ts">
  /**
   * Reusable 36x20px pill toggle switch primitive.
   */
  interface Props {
    checked?: boolean
    disabled?: boolean
    ariaLabel?: string
    onchange?: (checked: boolean, event: Event) => void
    class?: string
  }

  let {
    checked = $bindable(false),
    disabled = false,
    ariaLabel,
    onchange,
    class: className = ''
  }: Props = $props()

  function handleChange(e: Event) {
    const target = e.currentTarget as HTMLInputElement
    checked = target.checked
    onchange?.(checked, e)
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<label
  class="vinta-toggle-switch {className}"
  class:disabled
  onclick={(e) => e.stopPropagation()}
>
  <input
    type="checkbox"
    bind:checked
    {disabled}
    aria-label={ariaLabel}
    onchange={handleChange}
  />
  <span class="vinta-toggle-slider"></span>
</label>

<style src="./styles/ToggleSwitch.css"></style>
