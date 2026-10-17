<script lang="ts">
  import { X, Check } from 'lucide-svelte'
  import { islandTooltip } from '../../services/islandTooltip'

  interface Props {
    isOpen: boolean
    initialTotalSeconds?: number
    onClose: () => void
    onApply: (seconds: number) => void
  }

  let { isOpen, initialTotalSeconds = 600, onClose, onApply }: Props = $props()

  let customHours = $state(0)
  let customMinutes = $state(10)
  let customSeconds = $state(0)

  const QUICK_SECOND_PRESETS = [
    { label: '3s', secs: 3 },
    { label: '5s', secs: 5 },
    { label: '10s', secs: 10 },
    { label: '30s', secs: 30 },
    { label: '1m', secs: 60 },
    { label: '5m', secs: 300 }
  ]

  function syncFromProps() {
    if (isOpen) {
      customHours = Math.floor(initialTotalSeconds / 3600)
      customMinutes = Math.floor((initialTotalSeconds % 3600) / 60)
      customSeconds = initialTotalSeconds % 60
    }
  }
  $effect(syncFromProps)

  function setQuickSecondPreset(secs: number): void {
    customHours = Math.floor(secs / 3600)
    customMinutes = Math.floor((secs % 3600) / 60)
    customSeconds = secs % 60
  }

  function handleApply(): void {
    const totalSecs = customHours * 3600 + customMinutes * 60 + customSeconds
    if (totalSecs > 0) {
      onApply(totalSecs)
    }
    onClose()
  }
</script>

{#if isOpen}
  <div class="drawer-overlay">
    <div class="drawer-header">
      <span class="drawer-title">Custom Duration</span>
      <button
        class="header-tool-btn"
        onclick={onClose}
        use:islandTooltip={'Close'}
      >
        <X size={15} />
      </button>
    </div>

    <div class="drawer-content">
      <div class="drawer-row">
        <span class="drawer-row-label">Hours</span>
        <div class="drawer-stepper">
          <button
            class="stepper-btn"
            onclick={() => {
              customHours = Math.max(0, customHours - 1)
            }}>-</button
          >
          <input
            type="number"
            min="0"
            max="23"
            class="stepper-input"
            bind:value={customHours}
            aria-label="Hours"
          />
          <button
            class="stepper-btn"
            onclick={() => {
              customHours = Math.min(23, customHours + 1)
            }}>+</button
          >
        </div>
      </div>

      <div class="drawer-row">
        <span class="drawer-row-label">Minutes</span>
        <div class="drawer-stepper">
          <button
            class="stepper-btn"
            onclick={() => {
              customMinutes = Math.max(0, customMinutes - 1)
            }}>-</button
          >
          <input
            type="number"
            min="0"
            max="59"
            class="stepper-input"
            bind:value={customMinutes}
            aria-label="Minutes"
          />
          <button
            class="stepper-btn"
            onclick={() => {
              customMinutes = Math.min(59, customMinutes + 1)
            }}>+</button
          >
        </div>
      </div>

      <div class="drawer-row">
        <span class="drawer-row-label">Seconds</span>
        <div class="drawer-stepper">
          <button
            class="stepper-btn"
            onclick={() => {
              customSeconds = Math.max(0, customSeconds - 1)
            }}>-</button
          >
          <input
            type="number"
            min="0"
            max="59"
            class="stepper-input"
            bind:value={customSeconds}
            aria-label="Seconds"
          />
          <button
            class="stepper-btn"
            onclick={() => {
              customSeconds = Math.min(59, customSeconds + 1)
            }}>+</button
          >
        </div>
      </div>

      <div class="quick-seconds-row">
        {#each QUICK_SECOND_PRESETS as qp (qp.secs)}
          <button
            class="quick-sec-chip"
            onclick={() => setQuickSecondPreset(qp.secs)}
          >
            {qp.label}
          </button>
        {/each}
      </div>

      <button class="drawer-confirm-btn" onclick={handleApply}>
        <Check size={15} />
        <span>Set Duration</span>
      </button>
    </div>
  </div>
{/if}
