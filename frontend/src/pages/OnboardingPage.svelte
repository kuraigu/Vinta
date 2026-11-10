<!--
  @fileoverview First-run interactive setup and onboarding wizard.

  Guides new users through island basics, gestures & shortcuts, pages & widgets,
  overlay limitations, and personalization using a two-column sidebar layout.
-->
<script module lang="ts">
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 320, h: 36, r: 18 },
    expanded: { w: 580, h: 460, r: 30, minH: 420, maxH: 520 },
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
    FolderOpen,
    Calendar,
    ShieldAlert,
    Gamepad2,
    Lock,
    Sparkles,
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
  import CustomSelect from '../components/CustomSelect.svelte'
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
    badgeColor: string
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
      icon: Sparkles
    }
  ]

  let stepData = $derived(steps[currentStep] || steps[0])

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

<div data-vinta="onboarding-page" class="onboarding-shell">
  <!-- Left Navigation Sidebar -->
  <aside class="onboarding-sidebar" aria-label="Onboarding steps">
    <div class="sidebar-steps-list">
      {#each steps as step, idx (step.id)}
        {@const StepIcon = step.icon}
        <button
          type="button"
          class="sidebar-step-tab"
          class:active={currentStep === idx}
          onclick={() => (currentStep = idx)}
        >
          <span class="step-tab-badge badge-{step.badgeColor}">
            <StepIcon size={13} strokeWidth={2.2} />
          </span>
          <span class="step-tab-label">{$tStore(step.navKey)}</span>
        </button>
      {/each}
    </div>

    <div class="sidebar-mascot-dock">
      <div class="sidebar-mascot-frame">
        <VintaMascot pose={stepData.pose} />
      </div>
    </div>
  </aside>

  <!-- Right Top-Aligned Content Pane -->
  <section class="onboarding-main-pane">
    <header class="pane-header">
      <div class="pane-header-top">
        <h2 class="pane-title">{$tStore(stepData.titleKey)}</h2>
        <span class="pane-step-counter"
          >{currentStep + 1} / {steps.length}</span
        >
      </div>
      <p class="pane-subtitle">{$tStore(stepData.subtitleKey)}</p>
    </header>

    <div class="pane-body">
      {#if currentStep === 0}
        <!-- Step 1: Welcome, Language & Name -->
        <div class="onboarding-inputs-grid">
          <div class="onboarding-field-card">
            <span class="field-card-label">
              {$tStore('settings.language_label')}
            </span>
            <CustomSelect
              options={localeOptions}
              bind:value={$localeStore}
              fullWidth={true}
            />
          </div>

          <div class="onboarding-field-card">
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
          </div>
        </div>

        <div class="example-group-card">
          <div class="example-row">
            <div class="example-icon-box badge-cyan">
              <Maximize2 size={15} />
            </div>
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
            <div class="example-icon-box badge-emerald">
              <LayoutGrid size={15} />
            </div>
            <div class="example-text-col">
              <span class="example-title"
                >{$tStore('onboarding.feature_glance_title')}</span
              >
              <span class="example-desc"
                >{$tStore('onboarding.feature_glance_desc')}</span
              >
            </div>
          </div>
        </div>
      {:else if currentStep === 1}
        <!-- Step 2: Island Basics (Screenshot 2 Reference) -->
        <div class="example-group-card">
          <div class="example-row">
            <div class="example-icon-box badge-cyan">
              <Maximize2 size={15} />
            </div>
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
            <div class="example-icon-box badge-blue">
              <Minimize2 size={15} />
            </div>
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
            <div class="example-icon-box badge-amber">
              <Move size={15} />
            </div>
            <div class="example-text-col">
              <span class="example-title"
                >{$tStore('onboarding.feature_move_title')}</span
              >
              <span class="example-desc"
                >{$tStore('onboarding.feature_move_desc')}</span
              >
            </div>
          </div>
        </div>

        <div class="onboarding-setting-row">
          <div class="setting-row-text">
            <span class="setting-row-title"
              >{$tStore('onboarding.resting_alignment')}</span
            >
            <span class="setting-row-desc"
              >{$tStore('onboarding.resting_alignment_desc')}</span
            >
          </div>
          <div class="onboarding-segment-group">
            {#each ['top-left', 'center', 'top-right'] as pos (pos)}
              <button
                type="button"
                class="segment-pill-btn"
                class:active={$hiddenPillPositionStore === pos}
                onclick={() => ($hiddenPillPositionStore = pos as any)}
              >
                {pos === 'top-left'
                  ? $tStore('onboarding.align_left')
                  : pos === 'top-right'
                    ? $tStore('onboarding.align_right')
                    : $tStore('onboarding.align_center')}
              </button>
            {/each}
          </div>
        </div>
      {:else if currentStep === 2}
        <!-- Step 3: Gestures & Shortcuts -->
        <div class="example-group-card">
          <div class="example-row">
            <div class="example-icon-box badge-purple">
              <Keyboard size={15} />
            </div>
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
            <div class="example-icon-box badge-amber">
              <StickyNote size={15} />
            </div>
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
            <div class="example-icon-box badge-blue">
              <EyeOff size={15} />
            </div>
            <div class="example-text-col">
              <span class="example-title"
                >{$tStore('onboarding.gesture_autohide_title')}</span
              >
              <span class="example-desc"
                >{$tStore('onboarding.gesture_autohide_desc')}</span
              >
            </div>
          </div>
        </div>
      {:else if currentStep === 3}
        <!-- Step 4: Pages & Widgets -->
        <div class="example-group-card">
          <div class="example-row">
            <div class="example-icon-box badge-rose">
              <Music size={15} />
            </div>
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
            <div class="example-icon-box badge-cyan">
              <FolderOpen size={15} />
            </div>
            <div class="example-text-col">
              <span class="example-title"
                >{$tStore('onboarding.widget_shelf_title')}</span
              >
              <span class="example-desc"
                >{$tStore('onboarding.widget_shelf_desc')}</span
              >
            </div>
          </div>

          <div class="example-row">
            <div class="example-icon-box badge-emerald">
              <Calendar size={15} />
            </div>
            <div class="example-text-col">
              <span class="example-title"
                >{$tStore('onboarding.widget_productivity_title')}</span
              >
              <span class="example-desc"
                >{$tStore('onboarding.widget_productivity_desc')}</span
              >
            </div>
          </div>
        </div>
      {:else if currentStep === 4}
        <!-- Step 5: Limitations & Overlay Compatibility -->
        <div class="example-group-card">
          <div class="example-row">
            <div class="example-icon-box badge-amber">
              <Gamepad2 size={15} />
            </div>
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
            <div class="example-icon-box badge-purple">
              <ShieldAlert size={15} />
            </div>
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
            <div class="example-icon-box badge-rose">
              <Lock size={15} />
            </div>
            <div class="example-text-col">
              <span class="example-title"
                >{$tStore('onboarding.limit_drm_title')}</span
              >
              <span class="example-desc"
                >{$tStore('onboarding.limit_drm_desc')}</span
              >
            </div>
          </div>
        </div>
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

        <div class="onboarding-setting-row">
          <div class="setting-row-text">
            <span class="setting-row-title"
              >{$tStore('onboarding.start_with_windows')}</span
            >
            <span class="setting-row-desc"
              >{$tStore('onboarding.start_with_windows_desc')}</span
            >
          </div>
          <button
            type="button"
            class="onboarding-toggle-pill"
            class:active={$autostartStore}
            onclick={handleToggleAutostart}
          >
            <Power size={12} />
            <span
              >{$autostartStore
                ? $tStore('common.enabled')
                : $tStore('common.disabled')}</span
            >
          </button>
        </div>

        <div class="onboarding-setting-row">
          <div class="setting-row-text">
            <span class="setting-row-title"
              >{$tStore('onboarding.system_stats')}</span
            >
            <span class="setting-row-desc"
              >{$tStore('onboarding.system_stats_desc')}</span
            >
          </div>
          <button
            type="button"
            class="onboarding-toggle-pill"
            class:active={Boolean($homeWidgetsVisibilityStore.systemStats)}
            onclick={handleToggleSystemStats}
          >
            <Activity size={12} />
            <span
              >{$homeWidgetsVisibilityStore.systemStats
                ? $tStore('common.enabled')
                : $tStore('common.disabled')}</span
            >
          </button>
        </div>
      {/if}
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
</div>

<style src="./styles/OnboardingPage.css"></style>
