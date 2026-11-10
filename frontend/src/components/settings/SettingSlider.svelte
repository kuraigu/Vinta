<!--
  @component SettingSlider
  @description Custom range slider for numeric preferences (volume, delay timeouts,
  opacity). Renders an active fill track, draggable thumb, and live value readout with unit labels.
-->
<script lang="ts">
  interface Props {
    value: number
    min: number
    max: number
    step?: number
    unit?: string
    defaultValue?: number
    onchange?: (val: number) => void
  }

  let {
    value = $bindable(0),
    min,
    max,
    step = 1,
    unit = '',
    defaultValue,
    onchange
  }: Props = $props()

  let defaultTickPercent = $derived(
    defaultValue !== undefined && max > min
      ? Math.max(0, Math.min(100, ((defaultValue - min) / (max - min)) * 100))
      : null
  )

  function clamp(val: number): number {
    if (isNaN(val)) return min
    return Math.min(max, Math.max(min, val))
  }

  function handleInput(e: Event) {
    const target = e.target as HTMLInputElement
    const parsed = parseFloat(target.value)
    if (!isNaN(parsed)) {
      if (parsed > max) {
        value = max
        target.value = String(max)
        if (onchange) onchange(max)
      } else {
        value = parsed
        if (onchange) onchange(parsed)
      }
    }
  }

  function handleBlur(e: Event) {
    const target = e.target as HTMLInputElement
    const parsed = parseFloat(target.value)
    const clamped = clamp(parsed)
    value = clamped
    target.value = String(clamped)
    if (onchange) onchange(clamped)
  }

  function handleSliderChange(e: Event) {
    const target = e.target as HTMLInputElement
    const parsed = parseFloat(target.value)
    if (!isNaN(parsed)) {
      value = parsed
      if (onchange) onchange(parsed)
    }
  }
</script>

<div class="setting-slider-container">
  <div class="slider-track-wrapper">
    {#if defaultTickPercent !== null}
      <span
        class="slider-default-tick"
        style="left: calc(7px + (100% - 14px) * {defaultTickPercent / 100});"
        aria-hidden="true"
      ></span>
    {/if}
    <input
      type="range"
      {min}
      {max}
      {step}
      {value}
      oninput={handleSliderChange}
      class="range-slider no-swipe"
    />
  </div>

  <div class="number-input-wrapper">
    <input
      type="number"
      {min}
      {max}
      {step}
      {value}
      oninput={handleInput}
      onblur={handleBlur}
      class="number-input no-swipe"
    />
    {#if unit}
      <span class="unit-label">{unit}</span>
    {/if}
  </div>
</div>

<style>
  .setting-slider-container {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
  }

  .slider-track-wrapper {
    position: relative;
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
  }

  .slider-default-tick {
    position: absolute;
    top: 50%;
    width: 2px;
    height: 10px;
    border-radius: 1px;
    background: var(--vinta-text-muted, rgba(255, 255, 255, 0.35));
    transform: translate(-50%, -50%);
    pointer-events: none;
    z-index: 0;
    opacity: 0.65;
  }

  .range-slider {
    position: relative;
    z-index: 1;
    width: 100%;
    min-width: 0;
    accent-color: var(--vinta-accent-primary, var(--vinta-primary, #38bdf8));
    height: 4px;
    border-radius: 2px;
    cursor: pointer;
  }

  .number-input-wrapper {
    display: flex;
    align-items: center;
    gap: 3px;
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid var(--vinta-border);
    border-radius: 6px;
    padding: 2px 6px;
    flex-shrink: 0;
  }

  .number-input {
    width: 42px;
    background: transparent;
    border: none;
    color: var(--vinta-text-primary);
    font-size: 11px;
    font-weight: 600;
    text-align: right;
    outline: none;
    appearance: textfield;
    -moz-appearance: textfield;
  }

  .number-input::-webkit-outer-spin-button,
  .number-input::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  .unit-label {
    font-size: 10px;
    font-weight: 500;
    color: var(--vinta-text-secondary);
  }
</style>
