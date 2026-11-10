<!--
  @component BehaviorTab
  @description Island interaction settings. Controls auto-hide timeout delays, cursor hover
  expand sensitivity, click-to-hide triggers, and drag runway options.
-->
<script lang="ts">
  import {
    Layout,
    EyeOff,
    Timer,
    MousePointer,
    RotateCcw,
    Monitor,
    MonitorSmartphone,
    Maximize2
  } from 'lucide-svelte'
  import BentoCard from '../../BentoCard.svelte'
  import ToggleSwitch from '../../ToggleSwitch.svelte'
  import SegmentedControl from '../../SegmentedControl.svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingSlider from '../SettingSlider.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import type { MonitorInfo } from '../../../../bindings/Vinta/backend/models.js'
  import {
    homeAutoHideStore,
    mediaAutoHideStore,
    autoHideDelayStore,
    autoHideDisplayStore,
    wakeMethodStore,
    hiddenPillOffsetXStore,
    allowDragPositionStore,
    displayLayoutStore,
    autoHideBorderlessStore,
    followActiveDisplayStore
  } from '../../../store'

  const widgetContentOptions = [
    { label: 'Compact Dashboard', value: 'dash' },
    { label: 'Compact Clock', value: 'clock' }
  ]

  let isResettingOffset = $state(false)

  /**
   * Switches the active display and centers the island on the selected monitor.
   */
  function handleSwitchMonitor(mon: MonitorInfo) {
    const baseCenter =
      $displayLayoutStore.virtualLeft + $displayLayoutStore.primaryCenter
    const monCenter = mon.left + mon.width / 2
    const targetOffset = Math.round(monCenter - baseCenter)
    $hiddenPillOffsetXStore = targetOffset
    $displayLayoutStore = {
      ...$displayLayoutStore,
      currentMonitor: mon
    }
  }

  /**
   * Resets the persistent horizontal position offset of the island back to monitor center.
   */
  function handleResetOffset() {
    $hiddenPillOffsetXStore = 0
    isResettingOffset = true
    setTimeout(() => {
      isResettingOffset = false
    }, 1200)
  }
</script>

<div class="setting-group">
  <BentoCard variant="group">
    <SettingItem label="Auto-Hide Home Page" icon={EyeOff}>
      <ToggleSwitch
        bind:checked={$homeAutoHideStore}
        ariaLabel="Auto-Hide Home Page"
      />
    </SettingItem>

    <SettingItem label="Auto-Hide Media Widget" icon={EyeOff}>
      <ToggleSwitch
        bind:checked={$mediaAutoHideStore}
        ariaLabel="Auto-Hide Media Widget"
      />
    </SettingItem>

    <SettingItem
      label="Auto-Hide Delay"
      description="Inactivity duration before collapsing"
      icon={Timer}
      flexCol={true}
      isDefault={$autoHideDelayStore === autoHideDelayStore.defaultValue}
      onReset={() => autoHideDelayStore.reset()}
    >
      <SettingSlider
        min={5}
        max={30}
        step={1}
        unit="s"
        defaultValue={autoHideDelayStore.defaultValue}
        bind:value={$autoHideDelayStore}
      />
    </SettingItem>

    <SettingItem
      label="Collapsed Pill Content"
      description="Widget display when collapsed"
      icon={Layout}
      flexCol={true}
      isDefault={$autoHideDisplayStore === autoHideDisplayStore.defaultValue}
      onReset={() => autoHideDisplayStore.reset()}
    >
      <CustomSelect
        options={widgetContentOptions}
        bind:value={$autoHideDisplayStore}
        defaultValue={autoHideDisplayStore.defaultValue}
        fullWidth={true}
        direction="auto"
      />
    </SettingItem>
  </BentoCard>

  <BentoCard variant="group">
    <SettingItem
      label="Wake Trigger Method"
      description="Interaction used to expand the island"
      icon={MousePointer}
      flexCol={true}
      isDefault={$wakeMethodStore === wakeMethodStore.defaultValue}
      onReset={() => wakeMethodStore.reset()}
    >
      <SegmentedControl
        fullWidth={true}
        options={[
          { value: 'hover', label: 'Hover' },
          { value: 'click', label: 'Click' }
        ]}
        bind:value={$wakeMethodStore}
      />
    </SettingItem>

    <SettingItem label="Allow Dragging Position" icon={MousePointer}>
      <ToggleSwitch
        bind:checked={$allowDragPositionStore}
        ariaLabel="Allow Dragging Position"
      />
    </SettingItem>

    <SettingItem
      label="Center Island Position"
      description="Resets the horizontal offset of the Dynamic Island to monitor center"
      icon={RotateCcw}
      flexCol={true}
    >
      <SettingActionButton
        variant="default"
        fullWidth={true}
        onclick={handleResetOffset}
      >
        <RotateCcw size={13} />
        <span>{isResettingOffset ? 'Centered' : 'Reset Offset to Center'}</span>
      </SettingActionButton>
    </SettingItem>
  </BentoCard>

  <BentoCard variant="group">
    {#if $displayLayoutStore.monitors && $displayLayoutStore.monitors.length > 1}
      <SettingItem
        label="Follow Active Display"
        description="Automatically move Vinta to the display containing the active window"
        icon={MonitorSmartphone}
      >
        <ToggleSwitch
          bind:checked={$followActiveDisplayStore}
          ariaLabel="Follow Active Display"
        />
      </SettingItem>

      <SettingItem
        label="Active Display"
        description="Select which connected monitor Vinta is positioned on"
        icon={Monitor}
        flexCol={true}
      >
        <SegmentedControl
          fullWidth={true}
          options={$displayLayoutStore.monitors.map((mon) => ({
            value: String(mon.id),
            label: mon.name || `Display ${mon.id}`
          }))}
          value={String($displayLayoutStore.currentMonitor?.id ?? '')}
          onchange={(val) => {
            const found = $displayLayoutStore.monitors.find(
              (m) => String(m.id) === val
            )
            if (found) handleSwitchMonitor(found)
          }}
        />
      </SettingItem>
    {/if}

    <SettingItem
      label="Auto-Hide for Borderless Fullscreen"
      description="Automatically hide the island when borderless full-screen games or apps are active"
      icon={Maximize2}
    >
      <ToggleSwitch
        bind:checked={$autoHideBorderlessStore}
        ariaLabel="Auto-Hide for Borderless Fullscreen"
      />
    </SettingItem>
  </BentoCard>
</div>

<style src="../../styles/settings/BehaviorTab.css"></style>
