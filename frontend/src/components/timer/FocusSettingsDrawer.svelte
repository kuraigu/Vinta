<script lang="ts">
  import DrawerPanel from '../DrawerPanel.svelte'
  import { islandTooltip } from '../../services/islandTooltip'
  import { timerAppStore, updateFocusSettings } from '../../store'

  interface Props {
    isOpen: boolean
    onClose: () => void
  }

  let { isOpen, onClose }: Props = $props()
</script>

{#if isOpen}
  <DrawerPanel
    title="Focus Settings"
    variant="overlay"
    closeTooltip="Close settings"
    {onClose}
  >
    <div class="drawer-content">
      <div class="focus-presets-row">
        <button
          class="focus-preset-chip"
          class:active={$timerAppStore.focus.settings.focusDuration === 25 &&
            $timerAppStore.focus.settings.shortBreakDuration === 5}
          onclick={() =>
            updateFocusSettings({
              focusDuration: 25,
              shortBreakDuration: 5
            })}
          use:islandTooltip={'Standard Pomodoro (25m focus / 5m break)'}
        >
          25 / 5m
        </button>
        <button
          class="focus-preset-chip"
          class:active={$timerAppStore.focus.settings.focusDuration === 45 &&
            $timerAppStore.focus.settings.shortBreakDuration === 15}
          onclick={() =>
            updateFocusSettings({
              focusDuration: 45,
              shortBreakDuration: 15
            })}
          use:islandTooltip={'Deep Work (45m focus / 15m break)'}
        >
          45 / 15m
        </button>
        <button
          class="focus-preset-chip"
          class:active={$timerAppStore.focus.settings.focusDuration === 50 &&
            $timerAppStore.focus.settings.shortBreakDuration === 10}
          onclick={() =>
            updateFocusSettings({
              focusDuration: 50,
              shortBreakDuration: 10
            })}
          use:islandTooltip={'Extended Focus (50m focus / 10m break)'}
        >
          50 / 10m
        </button>
      </div>

      <div class="drawer-row">
        <span class="drawer-row-label">Focus Length</span>
        <div class="drawer-stepper">
          <button
            class="stepper-btn"
            onclick={() =>
              updateFocusSettings({
                focusDuration: Math.max(
                  1,
                  $timerAppStore.focus.settings.focusDuration - 5
                )
              })}>-</button
          >
          <span class="stepper-value"
            >{$timerAppStore.focus.settings.focusDuration}m</span
          >
          <button
            class="stepper-btn"
            onclick={() =>
              updateFocusSettings({
                focusDuration: Math.min(
                  90,
                  $timerAppStore.focus.settings.focusDuration + 5
                )
              })}>+</button
          >
        </div>
      </div>

      <div class="drawer-row">
        <span class="drawer-row-label">Short Break</span>
        <div class="drawer-stepper">
          <button
            class="stepper-btn"
            onclick={() =>
              updateFocusSettings({
                shortBreakDuration: Math.max(
                  1,
                  $timerAppStore.focus.settings.shortBreakDuration - 1
                )
              })}>-</button
          >
          <span class="stepper-value"
            >{$timerAppStore.focus.settings.shortBreakDuration}m</span
          >
          <button
            class="stepper-btn"
            onclick={() =>
              updateFocusSettings({
                shortBreakDuration: Math.min(
                  30,
                  $timerAppStore.focus.settings.shortBreakDuration + 1
                )
              })}>+</button
          >
        </div>
      </div>

      <div class="drawer-row">
        <span class="drawer-row-label">Long Break</span>
        <div class="drawer-stepper">
          <button
            class="stepper-btn"
            onclick={() =>
              updateFocusSettings({
                longBreakDuration: Math.max(
                  1,
                  $timerAppStore.focus.settings.longBreakDuration - 5
                )
              })}>-</button
          >
          <span class="stepper-value"
            >{$timerAppStore.focus.settings.longBreakDuration}m</span
          >
          <button
            class="stepper-btn"
            onclick={() =>
              updateFocusSettings({
                longBreakDuration: Math.min(
                  60,
                  $timerAppStore.focus.settings.longBreakDuration + 5
                )
              })}>+</button
          >
        </div>
      </div>

      <div class="drawer-row">
        <span class="drawer-row-label">Long Break Cycle</span>
        <div class="drawer-stepper">
          <button
            class="stepper-btn"
            onclick={() =>
              updateFocusSettings({
                longBreakInterval: Math.max(
                  2,
                  $timerAppStore.focus.settings.longBreakInterval - 1
                )
              })}>-</button
          >
          <span class="stepper-value"
            >{$timerAppStore.focus.settings.longBreakInterval}</span
          >
          <button
            class="stepper-btn"
            onclick={() =>
              updateFocusSettings({
                longBreakInterval: Math.min(
                  10,
                  $timerAppStore.focus.settings.longBreakInterval + 1
                )
              })}>+</button
          >
        </div>
      </div>

      <div class="drawer-row">
        <span class="drawer-row-label">Sound Chimes</span>
        <label class="drawer-switch">
          <input
            type="checkbox"
            checked={$timerAppStore.focus.settings.soundEnabled}
            onchange={(e) =>
              updateFocusSettings({
                soundEnabled: (e.target as HTMLInputElement).checked
              })}
          />
          <span class="switch-slider"></span>
        </label>
      </div>

      <div class="drawer-row">
        <span class="drawer-row-label">Auto-start Breaks</span>
        <label class="drawer-switch">
          <input
            type="checkbox"
            checked={$timerAppStore.focus.settings.autoStartBreaks}
            onchange={(e) =>
              updateFocusSettings({
                autoStartBreaks: (e.target as HTMLInputElement).checked
              })}
          />
          <span class="switch-slider"></span>
        </label>
      </div>

      <div class="drawer-row">
        <span class="drawer-row-label">Auto-start Focus</span>
        <label class="drawer-switch">
          <input
            type="checkbox"
            checked={$timerAppStore.focus.settings.autoStartFocus}
            onchange={(e) =>
              updateFocusSettings({
                autoStartFocus: (e.target as HTMLInputElement).checked
              })}
          />
          <span class="switch-slider"></span>
        </label>
      </div>
    </div>
  </DrawerPanel>
{/if}
