<!--
  @fileoverview First-run interactive setup and onboarding wizard.

  Guides new users through island basics, gestures & shortcuts, pages & widgets,
  overlay limitations, and personalization using SidebarLayout, BentoCard, and IconBadge.
-->
<script module lang="ts">
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 320, h: 36, r: 18 },
    expanded: { w: 590, h: 480, r: 30, minH: 440, maxH: 520 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { getContext } from 'svelte'
  import {
    Hand,
    Maximize2,
    Minimize2,
    Move,
    Keyboard,
    StickyNote,
    EyeOff,
    LayoutGrid,
    Music,
    Timer,
    Calendar,
    ShieldAlert,
    Gamepad2,
    Lock,
    Monitor,
    Palette,
    User,
    Power,
    Activity,
    ChevronLeft,
    ChevronRight
  } from 'lucide-svelte'
  import {
    hasCompletedOnboardingStore,
    userNameStore,
    themeStore,
    hiddenPillPositionStore,
    autostartStore,
    homeWidgetsVisibilityStore,
    syncSettingsToDisk,
    localeStore,
    availableLocalesStore,
    tStore
  } from '../store'
  import { SetAutostart } from '../../bindings/Vinta/backend/app.js'
  import HomePage, { config as homeConfig } from './HomePage.svelte'
  import SidebarLayout, {
    type SidebarNavItem
  } from '../components/SidebarLayout.svelte'
  import BentoCard from '../components/BentoCard.svelte'
  import IconBadge, {
    type IconBadgeColor
  } from '../components/IconBadge.svelte'
  import CustomSelect from '../components/CustomSelect.svelte'
  import ToggleSwitch from '../components/ToggleSwitch.svelte'
  import SegmentedControl from '../components/SegmentedControl.svelte'
  import VintaMascot, {
    type MascotPose
  } from '../illustrations/mascot/VintaMascot.svelte'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let currentStep = $state(0)

  const steps: {
    id: string
    navKey: string
    titleKey: string
    subtitleKey: string
    pose: MascotPose
    badgeColor: IconBadgeColor
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon: any
  }[] = [
    {
      id: 'welcome',
      navKey: 'onboarding.nav_welcome',
      titleKey: 'onboarding.step_welcome_title',
      subtitleKey: 'onboarding.step_welcome_sub',
      pose: 'welcome',
      badgeColor: 'blue',
      icon: Hand
    },
    {
      id: 'basics',
      navKey: 'onboarding.nav_basics',
      titleKey: 'onboarding.step_basics_title',
      subtitleKey: 'onboarding.step_basics_sub',
      pose: 'position',
      badgeColor: 'cyan',
      icon: Maximize2
    },
    {
      id: 'shortcuts',
      navKey: 'onboarding.nav_shortcuts',
      titleKey: 'onboarding.step_shortcuts_title',
      subtitleKey: 'onboarding.step_shortcuts_sub',
      pose: 'settings',
      badgeColor: 'purple',
      icon: Keyboard
    },
    {
      id: 'widgets',
      navKey: 'onboarding.nav_widgets',
      titleKey: 'onboarding.step_widgets_title',
      subtitleKey: 'onboarding.step_widgets_sub',
      pose: 'theme',
      badgeColor: 'emerald',
      icon: LayoutGrid
    },
    {
      id: 'limitations',
      navKey: 'onboarding.nav_limitations',
      titleKey: 'onboarding.step_limitations_title',
      subtitleKey: 'onboarding.step_limitations_sub',
      pose: 'settings',
      badgeColor: 'amber',
      icon: ShieldAlert
    },
    {
      id: 'personalize',
      navKey: 'onboarding.nav_personalize',
      titleKey: 'onboarding.step_personalize_title',
      subtitleKey: 'onboarding.step_personalize_sub',
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

  function handleSelectStepId(id: string) {
    const idx = steps.findIndex((s) => s.id === id)
    if (idx !== -1) currentStep = idx
  }

  let localeOptions = $derived(
    $availableLocalesStore.map((l) => ({
      value: l.id,
      label: l.name,
      flag: l.flag || l.id
    }))
  )

  /**
   * Completes the onboarding wizard, persists status, and returns to HomePage.
   */
  function handleComplete() {
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
      currentStep++
    } else {
      handleComplete()
    }
  }

  /**
   * Navigates back to the preceding step.
   */
  function handlePrev() {
    if (currentStep > 0) {
      currentStep--
    }
  }

  /**
   * Handles keyboard navigation (ArrowLeft, ArrowRight, Enter) when not typing in an input.
   */
  function handleWindowKeyDown(e: KeyboardEvent) {
    const tag = (e.target as HTMLElement | null)?.tagName?.toLowerCase()
    if (tag === 'input' || tag === 'textarea') {
      if (e.key === 'Enter') {
        e.preventDefault()
        handleNext()
      }
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
  function handleToggleAutostart() {
    const nextVal = !$autostartStore
    $autostartStore = nextVal
    if (typeof SetAutostart === 'function') {
      SetAutostart(nextVal).catch(() => {})
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
          <h2 class="pane-title">{$tStore(stepData.titleKey)}</h2>
          <div
            class="pane-step-pills"
            role="tablist"
            aria-label="Onboarding progress"
          >
            {#each steps as s, idx (s.id)}
              <button
                type="button"
                class="step-progress-pill"
                class:active={currentStep === idx}
                class:completed={idx < currentStep}
                onclick={() => (currentStep = idx)}
                aria-label={$tStore(s.navKey)}
              ></button>
            {/each}
          </div>
        </div>
        <p class="pane-subtitle">{$tStore(stepData.subtitleKey)}</p>
      </header>

      <div class="pane-body">
        {#key currentStep}
          <div class="onboarding-step-transition">
            {#if currentStep === 0}
          <!-- Step 1: Welcome, Language & Name -->
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
              <label for="onboarding-name-input" class="field-card-label">
                {$tStore('onboarding.name_prompt')}
              </label>
              <div class="onboarding-input-wrapper">
                <User size={13} class="input-leading-icon" />
                <input
                  id="onboarding-name-input"
                  type="text"
                  class="onboarding-name-field"
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

          <BentoCard variant="group">
            <div class="example-row">
              <IconBadge icon={Maximize2} color="cyan" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.feature_expand_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.feature_expand_desc')}</span
                >
              </div>
            </div>

            <div class="example-row">
              <IconBadge icon={LayoutGrid} color="emerald" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.feature_glance_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.feature_glance_desc')}</span
                >
              </div>
            </div>
          </BentoCard>
        {:else if currentStep === 1}
          <!-- Step 2: Island Basics -->
          <BentoCard variant="group">
            <div class="example-row">
              <IconBadge icon={Maximize2} color="cyan" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.basics_expand_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.basics_expand_desc')}</span
                >
              </div>
            </div>

            <div class="example-row">
              <IconBadge icon={Minimize2} color="blue" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.basics_collapse_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.basics_collapse_desc')}</span
                >
              </div>
            </div>

            <div class="example-row">
              <IconBadge icon={Move} color="amber" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.feature_move_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.feature_move_desc')}</span
                >
              </div>
            </div>
          </BentoCard>

          <BentoCard variant="setting" class="onboarding-setting-row">
            <div class="setting-row-text">
              <span class="setting-row-title"
                >{$tStore('onboarding.resting_alignment')}</span
              >
              <span class="setting-row-desc"
                >{$tStore('onboarding.resting_alignment_desc')}</span
              >
            </div>
            <SegmentedControl
              size="sm"
              options={[
                { value: 'top-left', label: $tStore('onboarding.align_left') },
                { value: 'center', label: $tStore('onboarding.align_center') },
                { value: 'top-right', label: $tStore('onboarding.align_right') }
              ]}
              value={$hiddenPillPositionStore}
              onchange={(val) => ($hiddenPillPositionStore = val as any)}
            />
          </BentoCard>
        {:else if currentStep === 2}
          <!-- Step 3: Gestures & Shortcuts -->
          <BentoCard variant="group">
            <div class="example-row">
              <IconBadge icon={Keyboard} color="purple" size="md" />
              <div class="example-text-col">
                <div class="example-title-with-kbd">
                  <span class="example-title"
                    >{$tStore('onboarding.shortcut_toggle_title')}</span
                  >
                  <kbd class="kbd-pill">Alt + Shift + V</kbd>
                </div>
                <span class="example-desc"
                  >{$tStore('onboarding.shortcut_toggle_desc')}</span
                >
              </div>
            </div>

            <div class="example-row">
              <IconBadge icon={StickyNote} color="amber" size="md" />
              <div class="example-text-col">
                <div class="example-title-with-kbd">
                  <span class="example-title"
                    >{$tStore('onboarding.shortcut_notes_title')}</span
                  >
                  <kbd class="kbd-pill">Alt + Shift + N</kbd>
                </div>
                <span class="example-desc"
                  >{$tStore('onboarding.shortcut_notes_desc')}</span
                >
              </div>
            </div>

            <div class="example-row">
              <IconBadge icon={EyeOff} color="blue" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.gesture_autohide_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.gesture_autohide_desc')}</span
                >
              </div>
            </div>
          </BentoCard>
        {:else if currentStep === 3}
          <!-- Step 4: Pages & Widgets (Complete Feature Grouping) -->
          <BentoCard variant="group">
            <div class="example-row">
              <IconBadge icon={Calendar} color="blue" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.widget_calendar_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.widget_calendar_desc')}</span
                >
              </div>
            </div>

            <div class="example-row">
              <IconBadge icon={Timer} color="amber" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.widget_timer_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.widget_timer_desc')}</span
                >
              </div>
            </div>

            <div class="example-row">
              <IconBadge icon={Music} color="rose" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.widget_media_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.widget_media_desc')}</span
                >
              </div>
            </div>

            <div class="example-row">
              <IconBadge icon={LayoutGrid} color="emerald" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.widget_utilities_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.widget_utilities_desc')}</span
                >
              </div>
            </div>
          </BentoCard>
        {:else if currentStep === 4}
          <!-- Step 5: Limitations & Overlay Compatibility -->
          <BentoCard variant="group">
            <div class="example-row">
              <IconBadge icon={Gamepad2} color="amber" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.limit_fullscreen_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.limit_fullscreen_desc')}</span
                >
              </div>
            </div>

            <div class="example-row">
              <IconBadge icon={ShieldAlert} color="purple" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.limit_admin_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.limit_admin_desc')}</span
                >
              </div>
            </div>

            <div class="example-row">
              <IconBadge icon={Lock} color="rose" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.limit_drm_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.limit_drm_desc')}</span
                >
              </div>
            </div>

            <div class="example-row">
              <IconBadge icon={Monitor} color="blue" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.limit_docking_title')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.limit_docking_desc')}</span
                >
              </div>
            </div>
          </BentoCard>
        {:else if currentStep === 5}
          <!-- Step 6: Personalize & Finish -->
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
            <div class="example-row">
              <IconBadge icon={Power} color="emerald" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.start_with_windows')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.start_with_windows_desc')}</span
                >
              </div>
              <ToggleSwitch
                checked={$autostartStore}
                onchange={() => handleToggleAutostart()}
                ariaLabel={$tStore('onboarding.start_with_windows')}
              />
            </div>

            <div class="example-row">
              <IconBadge icon={Activity} color="cyan" size="md" />
              <div class="example-text-col">
                <span class="example-title"
                  >{$tStore('onboarding.system_stats')}</span
                >
                <span class="example-desc"
                  >{$tStore('onboarding.system_stats_desc')}</span
                >
              </div>
              <ToggleSwitch
                checked={Boolean($homeWidgetsVisibilityStore.systemStats)}
                onchange={() => handleToggleSystemStats()}
                ariaLabel={$tStore('onboarding.system_stats')}
              />
            </div>
          </BentoCard>
            {/if}
          </div>
        {/key}
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

<style src="./styles/OnboardingPage.css"></style>
