<!--
  @component BehaviorTab
  @description Island interaction settings. Controls auto-hide timeout delays, cursor hover
  expand sensitivity, click-to-hide triggers, and drag runway options.
-->
<script lang="ts">
  import {
    Layout,
    EyeOff,
    Home,
    Music,
    MousePointerClick,
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
  import SettingSectionHeader from '../SettingSectionHeader.svelte'
  import SettingSlider from '../SettingSlider.svelte'
  import SettingActionButton, {
    type SettingButtonStatus
  } from '../SettingActionButton.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import type { MonitorInfo } from '../../../../bindings/Vinta/backend/models.js'
  import {
    homeAutoHideStore,
    mediaAutoHideStore,
    clickOutsideMinimizeStore,
    autoHideDelayStore,
    feedbackTimeoutSecStore,
    autoHideDisplayStore,
    wakeMethodStore,
    hiddenPillOffsetXStore,
    allowDragPositionStore,
    minimalPillNavStore,
    displayLayoutStore,
    autoHideBorderlessStore,
    fullscreenBehaviorStore,
    returnOnFullscreenExitStore,
    tStore
  } from '../../../store'

  let widgetContentOptions = $derived([
    { label: $tStore('settings.collapsed_pill_dash'), value: 'dash' },
    { label: $tStore('settings.collapsed_pill_clock'), value: 'clock' }
  ])

  let wakeMethodOptions = $derived([
    { value: 'hover', label: $tStore('settings.wake_method_hover') },
    { value: 'click', label: $tStore('settings.wake_method_click') }
  ])

  let fullscreenBehaviorOptions = $derived([
    { label: $tStore('settings.fullscreen_move_display'), value: 'relocate' },
    { label: $tStore('settings.fullscreen_hide_island'), value: 'hide' }
  ])

  let offsetResetStatus = $state<SettingButtonStatus>('idle')

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
    offsetResetStatus = 'success'
    setTimeout(() => {
      offsetResetStatus = 'idle'
    }, 1200)
  }
</script>

<div class="setting-group">
  <BentoCard variant="group">
    <SettingItem
      label={$tStore('settings.autohide_home_label')}
      icon={Home}
      isDefault={$homeAutoHideStore === homeAutoHideStore.defaultValue}
      onReset={() => homeAutoHideStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$homeAutoHideStore}
        ariaLabel={$tStore('settings.autohide_home_label')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.autohide_media_label')}
      icon={Music}
      isDefault={$mediaAutoHideStore === mediaAutoHideStore.defaultValue}
      onReset={() => mediaAutoHideStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$mediaAutoHideStore}
        ariaLabel={$tStore('settings.autohide_media_label')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.click_outside_minimize_label')}
      description={$tStore('settings.click_outside_minimize_desc')}
      icon={MousePointerClick}
      isDefault={$clickOutsideMinimizeStore ===
        clickOutsideMinimizeStore.defaultValue}
      onReset={() => clickOutsideMinimizeStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$clickOutsideMinimizeStore}
        ariaLabel={$tStore('settings.click_outside_minimize_label')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.autohide_delay_label')}
      description={$tStore('settings.autohide_delay_desc')}
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
      label={$tStore('settings.feedback_timeout_label')}
      description={$tStore('settings.feedback_timeout_desc')}
      icon={Timer}
      flexCol={true}
      isDefault={$feedbackTimeoutSecStore ===
        feedbackTimeoutSecStore.defaultValue}
      onReset={() => feedbackTimeoutSecStore.reset()}
    >
      <SettingSlider
        min={1}
        max={10}
        step={0.5}
        unit="s"
        defaultValue={feedbackTimeoutSecStore.defaultValue}
        bind:value={$feedbackTimeoutSecStore}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.collapsed_pill_label')}
      description={$tStore('settings.collapsed_pill_desc')}
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
    <SettingItem
      label={$tStore('settings.minimal_pill_nav_label')}
      description={$tStore('settings.minimal_pill_nav_desc')}
      icon={Layout}
      isDefault={$minimalPillNavStore === minimalPillNavStore.defaultValue}
      onReset={() => minimalPillNavStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$minimalPillNavStore}
        ariaLabel={$tStore('settings.minimal_pill_nav_label')}
      />
    </SettingItem>
  </BentoCard>

  <BentoCard variant="group">
    <SettingItem
      label={$tStore('settings.wake_method_label')}
      description={$tStore('settings.wake_method_desc')}
      icon={MousePointer}
      flexCol={true}
      isDefault={$wakeMethodStore === wakeMethodStore.defaultValue}
      onReset={() => wakeMethodStore.reset()}
    >
      <SegmentedControl
        fullWidth={true}
        options={wakeMethodOptions}
        bind:value={$wakeMethodStore}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.allow_drag_label')}
      icon={MousePointer}
      isDefault={$allowDragPositionStore ===
        allowDragPositionStore.defaultValue}
      onReset={() => allowDragPositionStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$allowDragPositionStore}
        ariaLabel={$tStore('settings.allow_drag_label')}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.center_island_label')}
      description={$tStore('settings.center_island_desc')}
      icon={RotateCcw}
      flexCol={true}
    >
      <SettingActionButton
        variant="default"
        status={offsetResetStatus}
        successLabel={$tStore('settings.center_island_done')}
        fullWidth={true}
        onclick={handleResetOffset}
      >
        <RotateCcw size={14} />
        <span>{$tStore('settings.center_island_btn')}</span>
      </SettingActionButton>
    </SettingItem>
  </BentoCard>

  {#if $displayLayoutStore.monitors && $displayLayoutStore.monitors.length > 1}
    <SettingSectionHeader title={$tStore('settings.section_display')} />

    <BentoCard variant="group">
      <SettingItem
        label={$tStore('settings.active_display_label')}
        description={$tStore('settings.active_display_desc')}
        icon={Monitor}
        flexCol={true}
      >
        <SegmentedControl
          fullWidth={true}
          options={$displayLayoutStore.monitors.map((mon) => ({
            value: String(mon.id),
            label:
              mon.name ||
              $tStore('settings.display_fallback_name', { id: mon.id })
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
    </BentoCard>
  {/if}

  <SettingSectionHeader title={$tStore('settings.section_fullscreen')} />

  <BentoCard variant="group">
    <SettingItem
      label={$tStore('settings.autohide_borderless_label')}
      description={$tStore('settings.autohide_borderless_desc')}
      icon={Maximize2}
      isDefault={$autoHideBorderlessStore ===
        autoHideBorderlessStore.defaultValue}
      onReset={() => autoHideBorderlessStore.reset()}
    >
      <ToggleSwitch
        bind:checked={$autoHideBorderlessStore}
        ariaLabel={$tStore('settings.autohide_borderless_label')}
      />
    </SettingItem>

    {#if $displayLayoutStore.monitors && $displayLayoutStore.monitors.length > 1}
      <SettingItem
        label={$tStore('settings.fullscreen_behavior_label')}
        description={$tStore('settings.fullscreen_behavior_desc')}
        icon={MonitorSmartphone}
        flexCol={true}
        isDefault={$fullscreenBehaviorStore ===
          fullscreenBehaviorStore.defaultValue}
        onReset={() => fullscreenBehaviorStore.reset()}
      >
        <SegmentedControl
          fullWidth={true}
          options={fullscreenBehaviorOptions}
          value={$fullscreenBehaviorStore}
          onchange={(val) => {
            $fullscreenBehaviorStore = val as 'relocate' | 'hide'
          }}
        />
      </SettingItem>

      <SettingItem
        label={$tStore('settings.return_fullscreen_label')}
        description={$tStore('settings.return_fullscreen_desc')}
        icon={RotateCcw}
        disabled={$fullscreenBehaviorStore !== 'relocate'}
        isDefault={$returnOnFullscreenExitStore ===
          returnOnFullscreenExitStore.defaultValue}
        onReset={() => returnOnFullscreenExitStore.reset()}
      >
        <ToggleSwitch
          bind:checked={$returnOnFullscreenExitStore}
          disabled={$fullscreenBehaviorStore !== 'relocate'}
          ariaLabel={$tStore('settings.return_fullscreen_label')}
        />
      </SettingItem>
    {/if}
  </BentoCard>
</div>

<style src="../../styles/settings/BehaviorTab.css"></style>
