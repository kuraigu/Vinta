<svelte:options />

<script lang="ts">
  import { onMount, tick } from 'svelte'
  import { Clock, ChevronDown } from 'lucide-svelte'
  import ScrollView from './ScrollView.svelte'
  import { timeFormatStore } from '../store'

  let {
    value = $bindable('09:00'),
    ariaLabel = 'Select time',
    align = 'left'
  }: {
    value?: string
    ariaLabel?: string
    align?: 'left' | 'right'
  } = $props()

  let isOpen = $state(false)
  let containerRef = $state<HTMLDivElement | null>(null)
  let activeHourRef = $state<HTMLButtonElement | null>(null)
  let activeMinRef = $state<HTMLButtonElement | null>(null)

  function computeIs24H(): boolean {
    return $timeFormatStore === '24h'
  }
  let is24H = $derived.by(computeIs24H)

  // Internal state
  let selectedHour = $state(9)
  let selectedMinute = $state(0)
  let selectedAmPm = $state<'AM' | 'PM'>('AM')

  /**
   * Syncs internal components from the 24h HH:MM value prop.
   */
  function syncFromValue(): void {
    if (!value || !/^\d{1,2}:\d{2}$/.test(value)) {
      value = '09:00'
    }
    const [hStr, mStr] = value.split(':')
    const h24 = parseInt(hStr, 10) || 0
    const m = parseInt(mStr, 10) || 0

    selectedMinute = Math.max(0, Math.min(59, m))

    if (is24H) {
      selectedHour = Math.max(0, Math.min(23, h24))
    } else {
      selectedHour = h24 % 12 || 12
      selectedAmPm = h24 >= 12 ? 'PM' : 'AM'
    }
  }

  $effect(() => {
    syncFromValue()
  })

  /**
   * Updates the bound 24h value string from internal state components.
   */
  function emitValue(): void {
    let h24 = selectedHour
    if (!is24H) {
      if (selectedAmPm === 'AM') {
        h24 = selectedHour === 12 ? 0 : selectedHour
      } else {
        h24 = selectedHour === 12 ? 12 : selectedHour + 12
      }
    }
    const hh = String(Math.max(0, Math.min(23, h24))).padStart(2, '0')
    const mm = String(Math.max(0, Math.min(59, selectedMinute))).padStart(2, '0')
    value = `${hh}:${mm}`
  }

  /**
   * Computes the formatted readable label for the trigger button.
   */
  function computeDisplayLabel(): string {
    if (!value) return 'Pick time'
    const [hStr, mStr] = value.split(':')
    const h24 = parseInt(hStr, 10) || 0
    const mm = (parseInt(mStr, 10) || 0).toString().padStart(2, '0')

    if (is24H) {
      return `${String(h24).padStart(2, '0')}:${mm}`
    }
    const h12 = h24 % 12 || 12
    const ampm = h24 >= 12 ? 'PM' : 'AM'
    return `${h12}:${mm} ${ampm}`
  }
  let displayLabel = $derived.by(computeDisplayLabel)

  // Full 60-minute list for standard complete selection
  const MINUTES_LIST = Array.from({ length: 60 }, (_, i) => i)

  function computeHoursList(): number[] {
    if (is24H) {
      return Array.from({ length: 24 }, (_, i) => i)
    }
    return [12, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
  }
  let hoursList = $derived.by(computeHoursList)

  function handleSelectHour(h: number): void {
    selectedHour = h
    emitValue()
  }

  function handleSelectMinute(m: number): void {
    selectedMinute = m
    emitValue()
  }

  function handleSelectAmPm(ampm: 'AM' | 'PM'): void {
    selectedAmPm = ampm
    emitValue()
  }

  function handleSetNow(): void {
    const now = new Date()
    const hh = String(now.getHours()).padStart(2, '0')
    const mm = String(now.getMinutes()).padStart(2, '0')
    value = `${hh}:${mm}`
    syncFromValue()
    isOpen = false
  }

  async function scrollToActiveTime(): Promise<void> {
    await tick()
    if (activeHourRef) {
      activeHourRef.scrollIntoView({ block: 'center', behavior: 'smooth' })
    }
    if (activeMinRef) {
      activeMinRef.scrollIntoView({ block: 'center', behavior: 'smooth' })
    }
  }

  function bindActiveHour(node: HTMLButtonElement, isActive: boolean) {
    if (isActive) activeHourRef = node
    return {
      update(nextActive: boolean) {
        if (nextActive) activeHourRef = node
      }
    }
  }

  function bindActiveMin(node: HTMLButtonElement, isActive: boolean) {
    if (isActive) activeMinRef = node
    return {
      update(nextActive: boolean) {
        if (nextActive) activeMinRef = node
      }
    }
  }

  function toggleOpen(): void {
    isOpen = !isOpen
    if (isOpen) {
      syncFromValue()
      scrollToActiveTime()
    }
  }

  onMount(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        isOpen &&
        containerRef &&
        !containerRef.contains(event.target as Node)
      ) {
        isOpen = false
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  })
</script>

<div class="custom-time-picker" bind:this={containerRef}>
  <button
    type="button"
    class="time-trigger-btn"
    class:open={isOpen}
    onclick={toggleOpen}
    aria-label={ariaLabel}
    aria-haspopup="dialog"
    aria-expanded={isOpen}
  >
    <div class="trigger-label-group">
      <Clock size={13} class="trigger-icon" />
      <span class="trigger-text">{displayLabel}</span>
    </div>
    <ChevronDown size={13} class={`chevron-icon ${isOpen ? 'rotate' : ''}`} />
  </button>

  {#if isOpen}
    <div
      class="time-popover"
      class:align-right={align === 'right'}
      role="dialog"
      aria-modal="true"
    >
      <!-- Direct Digital Editable Input Header -->
      <div class="digital-input-header">
        <div class="digital-inputs-pair">
          <input
            type="number"
            class="digit-box"
            min={is24H ? 0 : 1}
            max={is24H ? 23 : 12}
            bind:value={selectedHour}
            oninput={() => emitValue()}
            aria-label="Hours"
          />
          <span class="digit-colon">:</span>
          <input
            type="number"
            class="digit-box"
            min="0"
            max="59"
            bind:value={selectedMinute}
            oninput={() => emitValue()}
            aria-label="Minutes"
          />
        </div>

        {#if !is24H}
          <div class="ampm-header-toggle">
            <button
              type="button"
              class="ampm-btn"
              class:active={selectedAmPm === 'AM'}
              onclick={() => handleSelectAmPm('AM')}
            >
              AM
            </button>
            <button
              type="button"
              class="ampm-btn"
              class:active={selectedAmPm === 'PM'}
              onclick={() => handleSelectAmPm('PM')}
            >
              PM
            </button>
          </div>
        {/if}
      </div>

      <!-- Industry Standard Columns (Hours, Minutes, Period) -->
      <div class="time-columns-row" class:has-ampm={!is24H}>
        <!-- Hours column -->
        <div class="time-col">
          <span class="col-title">Hour</span>
          <ScrollView class="col-scroll" enableSwipe={false}>
            <div class="col-items-list">
              {#each hoursList as h (h)}
                {@const isActive = selectedHour === h}
                <button
                  type="button"
                  class="col-item-btn"
                  class:active={isActive}
                  use:bindActiveHour={isActive}
                  onclick={() => handleSelectHour(h)}
                >
                  {String(h).padStart(2, '0')}
                </button>
              {/each}
            </div>
          </ScrollView>
        </div>

        <!-- Minutes column with full 60-minute scroll list -->
        <div class="time-col">
          <span class="col-title">Min</span>
          <ScrollView class="col-scroll" enableSwipe={false}>
            <div class="col-items-list">
              {#each MINUTES_LIST as m (m)}
                {@const isActive = selectedMinute === m}
                <button
                  type="button"
                  class="col-item-btn"
                  class:active={isActive}
                  use:bindActiveMin={isActive}
                  onclick={() => handleSelectMinute(m)}
                >
                  {String(m).padStart(2, '0')}
                </button>
              {/each}
            </div>
          </ScrollView>
        </div>

        <!-- AM/PM column in 12h mode -->
        {#if !is24H}
          <div class="time-col ampm-col">
            <span class="col-title">Period</span>
            <div class="ampm-col-list">
              <button
                type="button"
                class="col-item-btn ampm-btn-item"
                class:active={selectedAmPm === 'AM'}
                onclick={() => handleSelectAmPm('AM')}
              >
                AM
              </button>
              <button
                type="button"
                class="col-item-btn ampm-btn-item"
                class:active={selectedAmPm === 'PM'}
                onclick={() => handleSelectAmPm('PM')}
              >
                PM
              </button>
            </div>
          </div>
        {/if}
      </div>

      <!-- Popover Footer -->
      <div class="popover-footer">
        <button
          type="button"
          class="popover-act-btn now"
          onclick={handleSetNow}
        >
          Now
        </button>
        <button
          type="button"
          class="popover-act-btn done"
          onclick={() => (isOpen = false)}
        >
          Done
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
  .custom-time-picker {
    position: relative;
    width: 100%;
    user-select: none;
  }

  .time-trigger-btn {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    height: 32px;
    padding: 0 10px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--c-border, rgba(255, 255, 255, 0.1));
    border-radius: 9px;
    color: var(--c-text-primary, #ffffff);
    font-size: 11.5px;
    font-weight: 500;
    cursor: pointer;
    outline: none;
    transition: all 0.16s ease;
    box-sizing: border-box;
  }

  .time-trigger-btn:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(255, 255, 255, 0.18);
  }

  .time-trigger-btn.open {
    border-color: var(--c-primary, #38bdf8);
    background: rgba(255, 255, 255, 0.09);
    box-shadow: 0 0 0 2px rgba(56, 189, 248, 0.2);
  }

  .trigger-label-group {
    display: flex;
    align-items: center;
    gap: 7px;
    min-width: 0;
  }

  :global(.trigger-icon) {
    color: var(--c-primary, #38bdf8);
    flex-shrink: 0;
  }

  .trigger-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  :global(.chevron-icon) {
    color: var(--c-text-secondary, #a1a1aa);
    transition: transform 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
    flex-shrink: 0;
  }

  :global(.chevron-icon.rotate) {
    transform: rotate(180deg);
  }

  /* Popover Floating Card */
  .time-popover {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    width: 200px;
    z-index: 250;
    background: #181920;
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 14px;
    box-shadow:
      inset 0 1px 0 0 rgba(255, 255, 255, 0.14),
      0 12px 32px rgba(0, 0, 0, 0.65);
    padding: 8px 10px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 6px;
    animation: popover-fade 0.15s ease-out;
  }

  .time-popover.align-right {
    left: auto;
    right: 0;
  }

  @keyframes popover-fade {
    from {
      opacity: 0;
      transform: translateY(-4px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  /* Direct Digital Input Header */
  .digital-input-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 6px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--c-border, rgba(255, 255, 255, 0.08));
    border-radius: 8px;
    padding: 3px 6px;
  }

  .digital-inputs-pair {
    display: flex;
    align-items: center;
    gap: 2px;
  }

  .digit-box {
    width: 24px;
    height: 20px;
    background: transparent;
    border: none;
    outline: none;
    color: var(--c-text-primary, #ffffff);
    font-size: 12px;
    font-weight: 700;
    text-align: center;
    font-variant-numeric: tabular-nums;
    padding: 0;
    margin: 0;
    -moz-appearance: textfield;
    appearance: textfield;
  }

  .digit-box::-webkit-outer-spin-button,
  .digit-box::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  .digit-colon {
    font-size: 11px;
    font-weight: 800;
    color: var(--c-text-muted, #71717a);
  }

  .ampm-header-toggle {
    display: flex;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 6px;
    padding: 1.5px;
    gap: 1.5px;
  }

  .ampm-btn {
    border: none;
    background: transparent;
    border-radius: 5px;
    font-size: 9.5px;
    font-weight: 700;
    color: var(--c-text-secondary, #a1a1aa);
    padding: 2px 5px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .ampm-btn:hover {
    color: var(--c-text-primary, #ffffff);
  }

  .ampm-btn.active {
    background: var(--c-primary, #38bdf8);
    color: #0b0f19;
  }

  .time-columns-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 6px;
  }

  .time-columns-row.has-ampm {
    grid-template-columns: 1fr 1fr 0.85fr;
  }

  .time-col {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .col-title {
    font-size: 9px;
    font-weight: 700;
    color: var(--c-text-muted, #71717a);
    text-transform: uppercase;
    text-align: center;
  }

  :global(.col-scroll) {
    max-height: 110px;
    height: 110px;
    width: 100%;
  }

  .col-items-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding-right: 2px;
  }

  .ampm-col-list {
    display: flex;
    flex-direction: column;
    gap: 3px;
    padding-top: 2px;
  }

  .col-item-btn {
    border: none;
    background: transparent;
    border-radius: 6px;
    padding: 3px 0;
    text-align: center;
    font-size: 11px;
    font-weight: 500;
    color: var(--c-text-primary, #e4e4e7);
    cursor: pointer;
    transition: all 0.12s ease;
  }

  .col-item-btn:hover {
    background: rgba(255, 255, 255, 0.09);
    color: #ffffff;
  }

  .col-item-btn.active {
    background: rgba(56, 189, 248, 0.2);
    color: var(--c-primary, #38bdf8);
    font-weight: 700;
  }

  .ampm-btn-item {
    padding: 6px 0;
  }

  .popover-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-top: 5px;
    border-top: 1px solid rgba(255, 255, 255, 0.08);
  }

  .popover-act-btn {
    font-size: 10px;
    font-weight: 600;
    padding: 2px 8px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
    border: none;
  }

  .popover-act-btn.now {
    background: transparent;
    color: var(--c-text-secondary, #a1a1aa);
  }

  .popover-act-btn.now:hover {
    color: var(--c-text-primary, #ffffff);
  }

  .popover-act-btn.done {
    background: var(--c-primary, #38bdf8);
    color: #0b0f19;
  }

  .popover-act-btn.done:hover {
    background: #7dd3fc;
  }
</style>
