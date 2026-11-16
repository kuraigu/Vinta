<!--
  @component CustomTimerDrawer
  @description Slide-over drawer for setting a custom countdown duration with hours,
  minutes, and seconds pickers.
-->
<script lang="ts">
  import { Check } from 'lucide-svelte'
  import DrawerPanel from '../DrawerPanel.svelte'

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
  let hoursInputText = $state('00')
  let minutesInputText = $state('10')
  let secondsInputText = $state('00')
  let isEditingHours = $state(false)
  let isEditingMinutes = $state(false)
  let isEditingSeconds = $state(false)

  const QUICK_SECOND_PRESETS = [
    { label: '3s', secs: 3 },
    { label: '5s', secs: 5 },
    { label: '10s', secs: 10 },
    { label: '30s', secs: 30 },
    { label: '1m', secs: 60 },
    { label: '5m', secs: 300 }
  ]

  function syncTextFromValues(): void {
    if (!isEditingHours) {
      hoursInputText = String(customHours).padStart(2, '0')
    }
    if (!isEditingMinutes) {
      minutesInputText = String(customMinutes).padStart(2, '0')
    }
    if (!isEditingSeconds) {
      secondsInputText = String(customSeconds).padStart(2, '0')
    }
  }

  function syncFromProps() {
    if (isOpen) {
      customHours = Math.floor(initialTotalSeconds / 3600)
      customMinutes = Math.floor((initialTotalSeconds % 3600) / 60)
      customSeconds = initialTotalSeconds % 60
      syncTextFromValues()
    }
  }
  $effect(syncFromProps)

  function setQuickSecondPreset(secs: number): void {
    customHours = Math.floor(secs / 3600)
    customMinutes = Math.floor((secs % 3600) / 60)
    customSeconds = secs % 60
    syncTextFromValues()
  }

  function stepHours(delta: number): void {
    customHours = Math.max(0, Math.min(23, customHours + delta))
    hoursInputText = String(customHours).padStart(2, '0')
  }

  function stepMinutes(delta: number): void {
    customMinutes = Math.max(0, Math.min(59, customMinutes + delta))
    minutesInputText = String(customMinutes).padStart(2, '0')
  }

  function stepSeconds(delta: number): void {
    customSeconds = Math.max(0, Math.min(59, customSeconds + delta))
    secondsInputText = String(customSeconds).padStart(2, '0')
  }

  function handleHoursInput(e: Event): void {
    const raw = (e.currentTarget as HTMLInputElement).value.replace(/\D/g, '')
    hoursInputText = raw.slice(0, 2)
    const parsed = parseInt(hoursInputText, 10)
    if (!isNaN(parsed)) {
      customHours = Math.max(0, Math.min(23, parsed))
    }
  }

  function handleMinutesInput(e: Event): void {
    const raw = (e.currentTarget as HTMLInputElement).value.replace(/\D/g, '')
    minutesInputText = raw.slice(0, 2)
    const parsed = parseInt(minutesInputText, 10)
    if (!isNaN(parsed)) {
      customMinutes = Math.max(0, Math.min(59, parsed))
    }
  }

  function handleSecondsInput(e: Event): void {
    const raw = (e.currentTarget as HTMLInputElement).value.replace(/\D/g, '')
    secondsInputText = raw.slice(0, 2)
    const parsed = parseInt(secondsInputText, 10)
    if (!isNaN(parsed)) {
      customSeconds = Math.max(0, Math.min(59, parsed))
    }
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
  <DrawerPanel
    title="Custom Duration"
    variant="overlay"
    closeTooltip="Close"
    {onClose}
  >
    <div class="drawer-content">
      <div class="drawer-row">
        <span class="drawer-row-label">Hours</span>
        <div class="drawer-stepper">
          <button class="stepper-btn" onclick={() => stepHours(-1)}>-</button>
          <input
            type="text"
            inputmode="numeric"
            maxlength="2"
            class="stepper-input"
            value={hoursInputText}
            onfocus={(e) => {
              isEditingHours = true
              e.currentTarget.select()
            }}
            oninput={handleHoursInput}
            onblur={() => {
              isEditingHours = false
              hoursInputText = String(customHours).padStart(2, '0')
            }}
            aria-label="Hours"
          />
          <button class="stepper-btn" onclick={() => stepHours(1)}>+</button>
        </div>
      </div>

      <div class="drawer-row">
        <span class="drawer-row-label">Minutes</span>
        <div class="drawer-stepper">
          <button class="stepper-btn" onclick={() => stepMinutes(-1)}>-</button>
          <input
            type="text"
            inputmode="numeric"
            maxlength="2"
            class="stepper-input"
            value={minutesInputText}
            onfocus={(e) => {
              isEditingMinutes = true
              e.currentTarget.select()
            }}
            oninput={handleMinutesInput}
            onblur={() => {
              isEditingMinutes = false
              minutesInputText = String(customMinutes).padStart(2, '0')
            }}
            aria-label="Minutes"
          />
          <button class="stepper-btn" onclick={() => stepMinutes(1)}>+</button>
        </div>
      </div>

      <div class="drawer-row">
        <span class="drawer-row-label">Seconds</span>
        <div class="drawer-stepper">
          <button class="stepper-btn" onclick={() => stepSeconds(-1)}>-</button>
          <input
            type="text"
            inputmode="numeric"
            maxlength="2"
            class="stepper-input"
            value={secondsInputText}
            onfocus={(e) => {
              isEditingSeconds = true
              e.currentTarget.select()
            }}
            oninput={handleSecondsInput}
            onblur={() => {
              isEditingSeconds = false
              secondsInputText = String(customSeconds).padStart(2, '0')
            }}
            aria-label="Seconds"
          />
          <button class="stepper-btn" onclick={() => stepSeconds(1)}>+</button>
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
  </DrawerPanel>
{/if}

<style src="../../pages/styles/TimerPage.css"></style>
