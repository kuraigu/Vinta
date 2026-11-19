<!--
  @fileoverview First-run interactive setup and onboarding wizard.

  Guides new users through Vinta's top-edge overlay model, live window positioning,
  expand/collapse/hide states & control satellite, behaviors, shortcuts & compatibility,
  and appearance/startup preferences using SidebarLayout, BentoCard, SettingItem, and
  SegmentedControl primitives.
-->
<script module lang="ts">
  import { ISLAND_WIDTH_TIERS, type IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 320, h: 36, r: 18 },
    expanded: {
      w: ISLAND_WIDTH_TIERS.wide,
      h: 495,
      r: 30,
      minH: 440,
      maxH: 520
    },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }

  let savedStepIndex = 0
</script>

<script lang="ts">
  import { flushSync, getContext } from 'svelte'
  import {
    Hand,
    Move,
    Minimize2,
    Keyboard,
    Palette,
    Layers,
    LayoutGrid,
    SlidersHorizontal,
    MousePointerClick,
    EyeOff,
    Clock as ClockIcon,
    Timer,
    FolderInput,
    Bell,
    Gamepad2,
    ShieldAlert,
    Gauge,
    Scale,
    Power,
    Activity,
    User,
    Monitor,
    ChevronLeft,
    ChevronRight
  } from 'lucide-svelte'
  import { Events } from '$lib/events'
  import {
    hasCompletedOnboardingStore,
    userNameStore,
    themeStore,
    hiddenPillPositionStore,
    hiddenPillOffsetXStore,
    displayLayoutStore,
    uiScaleStore,
    monitorUiScalesStore,
    getMonitorScaleKey,
    getMonitorBasePhysicalScale,
    getMaxScaleForMonitor,
    autoHideDisplayStore,
    wakeMethodStore,
    homeAutoHideStore,
    clickOutsideMinimizeStore,
    autoHideDelayStore,
    globalHotkeyStore,
    modeSwitchHotkeyStore,
    quickNotesHotkeyStore,
    animScaleStore,
    autostartStore,
    homeWidgetsVisibilityStore,
    syncSettingsToDisk,
    localeStore,
    availableLocalesStore,
    tStore,
    type HiddenPillPosition
  } from '../store'
  import { SetAutostart } from '../../bindings/Vinta/backend/app.js'
  import { computeAllowedOffsets } from '../services/islandGestures'
  import HomePage, { config as homeConfig } from './HomePage.svelte'
  import Clock from '../components/Clock.svelte'
  import SidebarLayout, {
    type SidebarNavItem
  } from '../components/SidebarLayout.svelte'
  import BentoCard from '../components/BentoCard.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import SettingItem from '../components/settings/SettingItem.svelte'
  import CustomSelect from '../components/CustomSelect.svelte'
  import ToggleSwitch from '../components/ToggleSwitch.svelte'
  import SegmentedControl from '../components/SegmentedControl.svelte'
  import OnboardingDiagrams from '../illustrations/onboarding/OnboardingDiagrams.svelte'
  import VintaMascot, {
    type MascotPose
  } from '../illustrations/mascot/VintaMascot.svelte'

  let { expanded = true }: { expanded?: boolean } = $props()

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let currentStep = $state(savedStepIndex)

  function syncSavedStepIndex() {
    savedStepIndex = currentStep
  }
  $effect(syncSavedStepIndex)

  const steps: {
    id: string
    navKey: string
    titleKey: string
    pose: MascotPose
    badgeColor: 'blue' | 'cyan' | 'purple' | 'emerald' | 'amber' | 'rose'
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon: any
  }[] = [
    {
      id: 'welcome',
      navKey: 'onboarding.nav_welcome',
      titleKey: 'onboarding.step_welcome_title',
      pose: 'welcome',
      badgeColor: 'blue',
      icon: Hand
    },
    {
      id: 'position',
      navKey: 'onboarding.nav_position',
      titleKey: 'onboarding.step_position_title',
      pose: 'position',
      badgeColor: 'cyan',
      icon: Move
    },
    {
      id: 'states',
      navKey: 'onboarding.nav_states',
      titleKey: 'onboarding.step_states_title',
      pose: 'settings',
      badgeColor: 'purple',
      icon: Minimize2
    },
    {
      id: 'behaviors',
      navKey: 'onboarding.nav_behaviors',
      titleKey: 'onboarding.step_behaviors_title',
      pose: 'theme',
      badgeColor: 'emerald',
      icon: SlidersHorizontal
    },
    {
      id: 'shortcuts',
      navKey: 'onboarding.nav_shortcuts',
      titleKey: 'onboarding.step_shortcuts_title',
      pose: 'settings',
      badgeColor: 'amber',
      icon: Keyboard
    },
    {
      id: 'personalize',
      navKey: 'onboarding.nav_personalize',
      titleKey: 'onboarding.step_personalize_title',
      pose: 'celebrate',
      badgeColor: 'rose',
      icon: Palette
    }
  ]

  let stepData = $derived(steps[currentStep] || steps[0])

  let sidebarItems: SidebarNavItem[] = $derived(
    steps.map((s) => ({
      id: s.id,
      label: $tStore(s.navKey),
      icon: s.icon,
      badge: s.badgeColor
    }))
  )

  let navDirection = $state<'forward' | 'backward'>('forward')

  /**
   * Transitions to the target onboarding step with directional View Transitions when supported.
   */
  function navigateToStep(nextIdx: number) {
    if (nextIdx === currentStep || nextIdx < 0 || nextIdx >= steps.length) {
      return
    }
    const direction: 'forward' | 'backward' =
      nextIdx > currentStep ? 'forward' : 'backward'
    navDirection = direction

    const prefersReducedMotion =
      typeof window !== 'undefined' &&
      window.matchMedia('(prefers-reduced-motion: reduce)').matches

    const startViewTransition =
      typeof document !== 'undefined'
        ? (
            document as Document & {
              startViewTransition?: (options: {
                update: () => void
                types?: string[]
              }) => unknown
            }
          ).startViewTransition
        : undefined

    if (
      typeof startViewTransition === 'function' &&
      !prefersReducedMotion &&
      $animScaleStore > 0
    ) {
      startViewTransition.call(document, {
        update: () => {
          flushSync(() => {
            currentStep = nextIdx
          })
        },
        types: [direction]
      })
      return
    }

    currentStep = nextIdx
  }

  function handleSelectStepId(id: string) {
    const idx = steps.findIndex((s) => s.id === id)
    if (idx !== -1) navigateToStep(idx)
  }

  let localeOptions = $derived(
    $availableLocalesStore.map((l) => ({
      value: l.id,
      label: l.name,
      flag: l.flag || l.id
    }))
  )

  let positionGeometry = $derived.by(() => {
    const layout = $displayLayoutStore
    const mon = layout?.currentMonitor || {
      id: 1,
      name: 'Primary Display',
      left: 0,
      top: 0,
      width: 1920,
      height: 1080,
      isPrimary: true,
      isPortrait: false,
      dpiScale: 1.0
    }
    const dpr =
      typeof window !== 'undefined' && window.devicePixelRatio > 0
        ? window.devicePixelRatio
        : 1.0
    const monKey = getMonitorScaleKey(mon)
    const perMonitor = $monitorUiScalesStore?.[monKey]
    const customScale =
      typeof perMonitor === 'number' && perMonitor >= 1.0
        ? perMonitor
        : mon.isPrimary
          ? $uiScaleStore || 1.0
          : 1.0
    const maxScale = getMaxScaleForMonitor(mon.width, mon.height, mon.dpiScale)
    const clampedCustom = Math.max(1.0, Math.min(maxScale, customScale))
    const basePhysical = getMonitorBasePhysicalScale(
      mon.width,
      mon.height,
      mon.dpiScale
    )
    const scale = clampedCustom * (basePhysical / dpr)
    const baseCenter =
      (layout?.virtualLeft ?? 0) + (layout?.primaryCenter ?? 960)
    const rightFootprint = 44 * scale
    const range = computeAllowedOffsets(
      config.expanded.w * scale,
      mon,
      baseCenter,
      0,
      rightFootprint,
      dpr
    )
    const centerOffset = Math.max(
      range.min,
      Math.min(
        range.max,
        Math.round((mon.left + mon.width / 2 - baseCenter) / dpr)
      )
    )
    const live = $hiddenPillOffsetXStore
    let ratio = 0
    if (live <= centerOffset) {
      const span = centerOffset - range.min
      ratio = span > 0 ? -((centerOffset - live) / span) : 0
    } else {
      const span = range.max - centerOffset
      ratio = span > 0 ? (live - centerOffset) / span : 0
    }
    const expected =
      $hiddenPillPositionStore === 'top-left'
        ? range.min
        : $hiddenPillPositionStore === 'top-right'
          ? range.max
          : centerOffset
    const isCustom = Math.abs(live - expected) > 24

    return {
      ratio: Math.max(-1, Math.min(1, ratio)),
      isCustom
    }
  })

  function handleSelectAlignment(align: HiddenPillPosition) {
    $hiddenPillPositionStore = align
    Events.Emit('vinta:apply-alignment', { alignment: align, animate: true })
  }

  function handleSelectMonitor(monIdStr: string) {
    Events.Emit('vinta:switch-monitor', {
      monitorId: monIdStr,
      alignment: $hiddenPillPositionStore
    })
  }

  let delayOptions = $derived([
    { value: '5', label: '5s' },
    { value: '10', label: '10s' },
    { value: '15', label: '15s' },
    { value: '30', label: '30s' }
  ])

  let activeDelayPreset = $derived(
    ['5', '10', '15', '30'].includes(String($autoHideDelayStore))
      ? String($autoHideDelayStore)
      : ''
  )

  let animPresetValue = $derived.by(() => {
    const val = $animScaleStore
    if (val <= 0) return '0'
    if (val <= 0.85) return '0.7'
    if (val <= 1.15) return '1'
    return '1.3'
  })

  const currentMonitor = $derived($displayLayoutStore.currentMonitor)
  const monitorWidth = $derived(currentMonitor?.width || 1920)
  const monitorHeight = $derived(currentMonitor?.height || 1080)
  const monitorDpi = $derived(
    currentMonitor?.dpiScale && currentMonitor.dpiScale > 0
      ? currentMonitor.dpiScale
      : 1.0
  )
  const monitorKey = $derived(getMonitorScaleKey(currentMonitor))
  const maxAllowedScale = $derived(
    getMaxScaleForMonitor(monitorWidth, monitorHeight, monitorDpi)
  )

  const uiScaleOptions = $derived([
    { label: '100% (1.0x)', value: '1', isDefault: true },
    {
      label:
        maxAllowedScale <= 1.1
          ? $tStore('appearance.scale_max', { pct: 110, scale: '1.1' })
          : '110% (1.1x)',
      value: '1.1'
    },
    ...(maxAllowedScale >= 1.2
      ? [
          {
            label:
              maxAllowedScale <= 1.2
                ? $tStore('appearance.scale_max', { pct: 120, scale: '1.2' })
                : '120% (1.2x)',
            value: '1.2'
          }
        ]
      : []),
    ...(maxAllowedScale >= 1.25
      ? [{ label: $tStore('appearance.scale_large'), value: '1.25' }]
      : []),
    ...(maxAllowedScale >= 1.5
      ? [
          {
            label:
              maxAllowedScale <= 1.5
                ? $tStore('appearance.scale_max', { pct: 150, scale: '1.5' })
                : '150% (1.5x)',
            value: '1.5'
          }
        ]
      : []),
    ...(maxAllowedScale >= 1.75
      ? [{ label: $tStore('appearance.scale_xl'), value: '1.75' }]
      : []),
    ...(maxAllowedScale >= 2.0
      ? [{ label: '200% (2.0x - 4K)', value: '2' }]
      : [])
  ])

  const activeMonitorScale = $derived.by(() => {
    const perMonitor = $monitorUiScalesStore?.[monitorKey]
    if (typeof perMonitor === 'number' && perMonitor >= 1.0) {
      return perMonitor
    }
    return currentMonitor?.isPrimary ? $uiScaleStore || 1.0 : 1.0
  })

  function handleScaleChange(val: string) {
    const num = parseFloat(val)
    if (isNaN(num) || num < 1.0) return
    const clamped = Math.max(1.0, Math.min(maxAllowedScale, num))
    monitorUiScalesStore.update((prev) => ({
      ...prev,
      [monitorKey]: clamped
    }))
    if (currentMonitor?.isPrimary && $uiScaleStore !== clamped) {
      $uiScaleStore = clamped
    }
  }

  function handleResetScale() {
    monitorUiScalesStore.update((prev) => {
      const next = { ...prev }
      delete next[monitorKey]
      return next
    })
    if (currentMonitor?.isPrimary) {
      uiScaleStore.reset()
    }
  }

  /**
   * Completes the onboarding wizard, persists status, and returns to HomePage.
   */
  function handleComplete() {
    savedStepIndex = 0
    hasCompletedOnboardingStore.set(true)
    try {
      localStorage.setItem(
        'vinta_has_completed_onboarding',
        JSON.stringify(true)
      )
    } catch {}
    syncSettingsToDisk(true)
    if (typeof navigate === 'function') {
      navigate(HomePage, homeConfig, false)
    }
  }

  /**
   * Advances to next step or completes on the final slide.
   */
  function handleNext() {
    if (currentStep < steps.length - 1) {
      navigateToStep(currentStep + 1)
    } else {
      handleComplete()
    }
  }

  /**
   * Navigates back to the preceding step.
   */
  function handlePrev() {
    if (currentStep > 0) {
      navigateToStep(currentStep - 1)
    }
  }

  /**
   * Handles keyboard navigation when expanded and focus is not inside a control.
   */
  function handleWindowKeyDown(e: KeyboardEvent) {
    if (!expanded) return
    const target = e.target as HTMLElement | null
    const tag = target?.tagName?.toLowerCase()
    if (tag === 'input' || tag === 'textarea') {
      if (e.key === 'Enter') {
        e.preventDefault()
        handleNext()
      }
      return
    }
    if (
      e.key === 'Enter' &&
      target?.closest('button, [role="button"], [role="tab"], [role="radio"]')
    ) {
      return
    }
    if (e.key === 'ArrowRight' || e.key === 'Enter') {
      e.preventDefault()
      handleNext()
    } else if (e.key === 'ArrowLeft' && currentStep > 0) {
      e.preventDefault()
      handlePrev()
    }
  }

  /**
   * Toggles Windows autostart on system boot.
   */
  async function handleToggleAutostart() {
    const nextVal = !$autostartStore
    $autostartStore = nextVal
    if (typeof SetAutostart === 'function') {
      try {
        await SetAutostart(nextVal)
      } catch {
        $autostartStore = !nextVal
      }
    }
  }

  /**
   * Toggles system hardware telemetry glance on home feed.
   */
  function handleToggleSystemStats() {
    $homeWidgetsVisibilityStore = {
      ...$homeWidgetsVisibilityStore,
      systemStats: !$homeWidgetsVisibilityStore.systemStats
    }
  }

  const themes = [
    { id: 'dark-gold', name: 'Dark Gold' },
    { id: 'nord-frost', name: 'Nord Frost' },
    { id: 'colorblind-safe', name: 'Colorblind Safe' }
  ]
</script>

<svelte:window onkeydown={handleWindowKeyDown} />

{#if !expanded}
  <div class="onboarding-compact-pill">
    <Clock expanded={false} />
    <span class="compact-pill-divider" aria-hidden="true"></span>
    <span class="compact-pill-text">
      {$tStore('onboarding.compact_pill_resume', {
        step: currentStep + 1,
        total: steps.length
      })}
    </span>
  </div>
{:else}
  <SidebarLayout
    dataVinta="onboarding-page"
    noDockClearance={true}
    sidebarWidth="168px"
    items={sidebarItems}
    activeId={stepData.id}
    onSelect={handleSelectStepId}
  >
    {#snippet sidebarFooter()}
      <div class="sidebar-mascot-dock">
        <div class="sidebar-mascot-frame">
          <VintaMascot pose={stepData.pose} />
        </div>
      </div>
    {/snippet}

    {#snippet content()}
      <section class="onboarding-main-pane">
        <header class="pane-header">
          <div class="pane-header-top">
            <h1 class="pane-title">{$tStore(stepData.titleKey)}</h1>
            <div
              class="pane-step-pills"
              role="group"
              aria-label={$tStore('onboarding.progress_aria')}
            >
              {#each steps as s, idx (s.id)}
                <button
                  type="button"
                  class="step-progress-pill"
                  class:active={currentStep === idx}
                  class:completed={idx < currentStep}
                  aria-current={currentStep === idx ? 'step' : undefined}
                  onclick={() => navigateToStep(idx)}
                  aria-label={$tStore(s.navKey)}
                ></button>
              {/each}
            </div>
          </div>
        </header>

        <div class="pane-body">
          <ScrollView class="onboarding-scroll">
            {#key currentStep}
              <div
                class="onboarding-step-transition"
                data-direction={navDirection}
              >
                {#if currentStep === 0}
                  <!-- Step 1: Welcome, Language & Profile -->
                  <div class="onboarding-inputs-grid">
                    <BentoCard variant="bento" class="onboarding-field-card">
                      <span class="field-card-label">
                        {$tStore('settings.language_label')}
                      </span>
                      <CustomSelect
                        options={localeOptions}
                        bind:value={$localeStore}
                        fullWidth={true}
                      />
                    </BentoCard>

                    <BentoCard variant="bento" class="onboarding-field-card">
                      <label
                        for="onboarding-name-input"
                        class="field-card-label"
                      >
                        {$tStore('onboarding.name_prompt')}
                      </label>
                      <div class="onboarding-input-wrapper">
                        <User size={13} class="input-leading-icon" />
                        <input
                          id="onboarding-name-input"
                          type="text"
                          class="onboarding-name-field"
                          autocomplete="name"
                          placeholder={$tStore('onboarding.name_placeholder')}
                          maxlength="20"
                          bind:value={$userNameStore}
                          onblur={() => {
                            if (!$userNameStore || !$userNameStore.trim()) {
                              $userNameStore = 'User'
                            } else {
                              $userNameStore = $userNameStore.trim()
                            }
                          }}
                        />
                      </div>
                    </BentoCard>
                  </div>

                  <OnboardingDiagrams mode="welcome" />

                  <BentoCard variant="group">
                    <SettingItem
                      icon={Layers}
                      badgeColor="cyan"
                      label={$tStore('onboarding.welcome_overlay_title')}
                      description={$tStore('onboarding.welcome_overlay_desc')}
                    />
                    <SettingItem
                      icon={LayoutGrid}
                      badgeColor="emerald"
                      label={$tStore('onboarding.welcome_dock_title')}
                      description={$tStore('onboarding.welcome_dock_desc')}
                    />
                  </BentoCard>
                {:else if currentStep === 1}
                  <!-- Step 2: Position & Display -->
                  <OnboardingDiagrams mode="position" />

                  <BentoCard variant="group">
                    <SettingItem
                      icon={Move}
                      badgeColor="cyan"
                      label={$tStore('onboarding.resting_alignment')}
                      description={$tStore('onboarding.resting_alignment_desc')}
                      flexCol={true}
                      isDefault={!positionGeometry.isCustom}
                      onReset={() =>
                        handleSelectAlignment($hiddenPillPositionStore)}
                    >
                      <SegmentedControl
                        size="sm"
                        fullWidth={true}
                        options={[
                          {
                            value: 'top-left',
                            label: $tStore('onboarding.align_left')
                          },
                          {
                            value: 'center',
                            label: $tStore('onboarding.align_center')
                          },
                          {
                            value: 'top-right',
                            label: $tStore('onboarding.align_right')
                          }
                        ]}
                        value={$hiddenPillPositionStore}
                        onchange={(val) =>
                          handleSelectAlignment(val as HiddenPillPosition)}
                      />
                    </SettingItem>

                    {#if $displayLayoutStore.monitors && $displayLayoutStore.monitors.length > 1}
                      <SettingItem
                        icon={Monitor}
                        badgeColor="blue"
                        label={$tStore('settings.active_display_label')}
                        flexCol={true}
                      >
                        <SegmentedControl
                          size="sm"
                          fullWidth={true}
                          options={$displayLayoutStore.monitors.map((mon) => ({
                            value: String(mon.id),
                            label:
                              mon.name ||
                              $tStore('settings.display_fallback_name', {
                                id: mon.id
                              })
                          }))}
                          value={String(
                            $displayLayoutStore.currentMonitor?.id ?? ''
                          )}
                          onchange={handleSelectMonitor}
                        />
                      </SettingItem>
                    {/if}

                    <SettingItem
                      icon={Hand}
                      badgeColor="amber"
                      label={$tStore('onboarding.position_drag_title')}
                      description={$tStore('onboarding.position_drag_desc')}
                    />
                  </BentoCard>
                {:else if currentStep === 2}
                  <!-- Step 3: Expand, Collapse & Hide -->
                  <OnboardingDiagrams
                    mode="states"
                    hiddenDisplay={$autoHideDisplayStore}
                    wakeMethod={$wakeMethodStore}
                  />

                  <BentoCard variant="group">
                    <SettingItem
                      icon={EyeOff}
                      badgeColor="cyan"
                      label={$tStore('onboarding.states_hidden_style')}
                      description={$tStore(
                        'onboarding.states_hidden_style_desc'
                      )}
                      flexCol={true}
                    >
                      <SegmentedControl
                        size="sm"
                        fullWidth={true}
                        options={[
                          {
                            value: 'dash',
                            label: $tStore('onboarding.hidden_style_dash')
                          },
                          {
                            value: 'clock',
                            label: $tStore('onboarding.hidden_style_clock')
                          }
                        ]}
                        bind:value={$autoHideDisplayStore}
                      />
                    </SettingItem>

                    <SettingItem
                      icon={MousePointerClick}
                      badgeColor="emerald"
                      label={$tStore('onboarding.states_wake_method')}
                      description={$tStore(
                        'onboarding.states_wake_method_desc'
                      )}
                      flexCol={true}
                    >
                      <SegmentedControl
                        size="sm"
                        fullWidth={true}
                        options={[
                          {
                            value: 'click',
                            label: $tStore('settings.wake_method_click')
                          },
                          {
                            value: 'hover',
                            label: $tStore('settings.wake_method_hover')
                          }
                        ]}
                        bind:value={$wakeMethodStore}
                      />
                    </SettingItem>

                    <SettingItem
                      icon={MousePointerClick}
                      badgeColor="indigo"
                      label={$tStore('settings.click_outside_minimize_label')}
                      description={$tStore(
                        'settings.click_outside_minimize_desc'
                      )}
                    >
                      <ToggleSwitch
                        bind:checked={$clickOutsideMinimizeStore}
                        ariaLabel={$tStore(
                          'settings.click_outside_minimize_label'
                        )}
                      />
                    </SettingItem>
                  </BentoCard>

                  <BentoCard variant="group">
                    <SettingItem
                      icon={SlidersHorizontal}
                      badgeColor="purple"
                      label={$tStore('onboarding.states_satellite_title')}
                      description={$tStore('onboarding.states_satellite_desc')}
                    />
                    <SettingItem
                      icon={MousePointerClick}
                      badgeColor="blue"
                      label={$tStore('onboarding.states_gestures_title')}
                      description={$tStore('onboarding.states_gestures_desc')}
                    />
                  </BentoCard>
                {:else if currentStep === 3}
                  <!-- Step 4: How It Behaves -->
                  <BentoCard variant="group">
                    <SettingItem
                      icon={EyeOff}
                      badgeColor="blue"
                      label={$tStore('onboarding.behavior_autohide_title')}
                      description={$tStore('onboarding.behavior_autohide_desc')}
                    >
                      <ToggleSwitch
                        bind:checked={$homeAutoHideStore}
                        ariaLabel={$tStore(
                          'onboarding.behavior_autohide_title'
                        )}
                      />
                    </SettingItem>

                    <SettingItem
                      icon={ClockIcon}
                      badgeColor="cyan"
                      label={$tStore('onboarding.behavior_delay_title')}
                      description={$tStore('onboarding.behavior_delay_desc')}
                      flexCol={true}
                      disabled={!$homeAutoHideStore}
                    >
                      <SegmentedControl
                        size="sm"
                        fullWidth={true}
                        options={delayOptions}
                        value={activeDelayPreset}
                        onchange={(val) => {
                          const sec = parseInt(val, 10)
                          if (!isNaN(sec)) $autoHideDelayStore = sec
                        }}
                      />
                    </SettingItem>
                  </BentoCard>

                  <BentoCard variant="group">
                    <SettingItem
                      icon={Timer}
                      badgeColor="amber"
                      label={$tStore('onboarding.behavior_lock_title')}
                      description={$tStore('onboarding.behavior_lock_desc')}
                    />
                    <SettingItem
                      icon={FolderInput}
                      badgeColor="emerald"
                      label={$tStore('onboarding.behavior_shelf_title')}
                      description={$tStore('onboarding.behavior_shelf_desc')}
                    />
                    <SettingItem
                      icon={Bell}
                      badgeColor="rose"
                      label={$tStore('onboarding.behavior_notif_title')}
                      description={$tStore('onboarding.behavior_notif_desc')}
                    />
                  </BentoCard>
                {:else if currentStep === 4}
                  <!-- Step 5: Shortcuts & Windows Limits -->
                  <div class="onboarding-hotkey-grid">
                    <BentoCard variant="bento" class="hotkey-chip-card">
                      <span class="hotkey-chip-label"
                        >{$tStore('onboarding.hotkey_toggle')}</span
                      >
                      <kbd class="kbd-pill"
                        >{$globalHotkeyStore || 'Alt+Shift+V'}</kbd
                      >
                    </BentoCard>

                    <BentoCard variant="bento" class="hotkey-chip-card">
                      <span class="hotkey-chip-label"
                        >{$tStore('onboarding.hotkey_mode')}</span
                      >
                      <kbd class="kbd-pill"
                        >{$modeSwitchHotkeyStore || 'Alt+Shift+D'}</kbd
                      >
                    </BentoCard>

                    <BentoCard variant="bento" class="hotkey-chip-card">
                      <span class="hotkey-chip-label"
                        >{$tStore('onboarding.hotkey_notes')}</span
                      >
                      <kbd class="kbd-pill"
                        >{$quickNotesHotkeyStore || 'Alt+Shift+N'}</kbd
                      >
                    </BentoCard>

                    <BentoCard variant="bento" class="hotkey-chip-card">
                      <span class="hotkey-chip-label"
                        >{$tStore('onboarding.hotkey_collapse')}</span
                      >
                      <kbd class="kbd-pill">Esc</kbd>
                    </BentoCard>
                  </div>

                  <BentoCard variant="group">
                    <SettingItem
                      icon={Gamepad2}
                      badgeColor="purple"
                      label={$tStore('onboarding.limit_fullscreen_title')}
                      description={$tStore('onboarding.limit_fullscreen_short')}
                    />
                    <SettingItem
                      icon={ShieldAlert}
                      badgeColor="rose"
                      label={$tStore('onboarding.limit_admin_drm_title')}
                      description={$tStore('onboarding.limit_admin_drm_desc')}
                    />
                    <SettingItem
                      icon={Monitor}
                      badgeColor="blue"
                      label={$tStore('onboarding.limit_gpu_title')}
                      description={$tStore('onboarding.limit_gpu_desc')}
                    />
                  </BentoCard>
                {:else if currentStep === 5}
                  <!-- Step 6: Look & Startup -->
                  <div class="onboarding-theme-grid">
                    {#each themes as t (t.id)}
                      <button
                        type="button"
                        class="theme-preview-card"
                        class:active={$themeStore === t.id}
                        onclick={() => ($themeStore = t.id)}
                      >
                        <span class="theme-dot {t.id}"></span>
                        <span class="theme-name">{t.name}</span>
                      </button>
                    {/each}
                  </div>

                  <BentoCard variant="group">
                    <SettingItem
                      icon={Gauge}
                      badgeColor="purple"
                      label={$tStore('appearance.anim_scale_label')}
                      description={$tStore('onboarding.anim_speed_desc')}
                      flexCol={true}
                    >
                      <SegmentedControl
                        size="sm"
                        fullWidth={true}
                        options={[
                          {
                            value: '0',
                            label: $tStore('appearance.anim_off')
                          },
                          {
                            value: '0.7',
                            label: $tStore('appearance.anim_snappy')
                          },
                          {
                            value: '1',
                            label: $tStore('appearance.anim_default')
                          },
                          {
                            value: '1.3',
                            label: $tStore('appearance.anim_smooth')
                          }
                        ]}
                        value={animPresetValue}
                        onchange={(val) => {
                          $animScaleStore = parseFloat(val)
                        }}
                      />
                    </SettingItem>

                    <SettingItem
                      icon={Scale}
                      badgeColor="blue"
                      label={$tStore('appearance.scale_label')}
                      description={`${$tStore('appearance.scale_desc')} (${monitorWidth}×${monitorHeight})`}
                      flexCol={true}
                      isDefault={activeMonitorScale === 1.0}
                      onReset={handleResetScale}
                    >
                      <CustomSelect
                        options={uiScaleOptions}
                        value={String(activeMonitorScale)}
                        onChange={handleScaleChange}
                        defaultValue="1"
                        fullWidth={true}
                        direction="auto"
                      />
                    </SettingItem>

                    <SettingItem
                      icon={Power}
                      badgeColor="emerald"
                      label={$tStore('onboarding.start_with_windows')}
                      description={$tStore(
                        'onboarding.start_with_windows_desc'
                      )}
                    >
                      <ToggleSwitch
                        checked={$autostartStore}
                        onchange={() => handleToggleAutostart()}
                        ariaLabel={$tStore('onboarding.start_with_windows')}
                      />
                    </SettingItem>

                    <SettingItem
                      icon={Activity}
                      badgeColor="cyan"
                      label={$tStore('onboarding.system_stats')}
                      description={$tStore('onboarding.system_stats_desc')}
                    >
                      <ToggleSwitch
                        checked={Boolean(
                          $homeWidgetsVisibilityStore.systemStats
                        )}
                        onchange={() => handleToggleSystemStats()}
                        ariaLabel={$tStore('onboarding.system_stats')}
                      />
                    </SettingItem>
                  </BentoCard>
                {/if}
              </div>
            {/key}
          </ScrollView>
        </div>

        <footer class="pane-footer">
          <button
            type="button"
            class="onboarding-skip-link"
            onclick={handleComplete}
          >
            {$tStore('common.skip')}
          </button>

          <div class="pane-footer-actions">
            {#if currentStep > 0}
              <button
                type="button"
                class="onboarding-nav-btn secondary"
                onclick={handlePrev}
              >
                <ChevronLeft size={14} strokeWidth={2.2} />
                <span>{$tStore('common.back')}</span>
              </button>
            {/if}

            <button
              type="button"
              class="onboarding-nav-btn primary"
              onclick={handleNext}
            >
              <span
                >{currentStep === steps.length - 1
                  ? $tStore('onboarding.btn_get_started')
                  : $tStore('common.continue')}</span
              >
              {#if currentStep < steps.length - 1}
                <ChevronRight size={14} strokeWidth={2.2} />
              {/if}
            </button>
          </div>
        </footer>
      </section>
    {/snippet}
  </SidebarLayout>
{/if}

<style src="./styles/OnboardingPage.css"></style>
