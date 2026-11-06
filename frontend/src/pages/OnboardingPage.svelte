<script module lang="ts">
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 320, h: 36, r: 18 },
    expanded: { w: 440, h: 440, r: 28, minH: 380, maxH: 520 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { getContext } from 'svelte'
  import {
    ChevronRight,
    ChevronLeft,
    MousePointer,
    Move,
    LayoutGrid,
    Activity,
    Power,
    User,
    Keyboard,
    SlidersHorizontal
  } from 'lucide-svelte'
  import {
    hasCompletedOnboardingStore,
    userNameStore,
    themeStore,
    hiddenPillPositionStore,
    followActiveDisplayStore,
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
    titleKey: string
    subtitleKey: string
    pose: MascotPose
  }[] = [
    {
      id: 'welcome',
      titleKey: 'onboarding.step_welcome_title',
      subtitleKey: 'onboarding.step_welcome_sub',
      pose: 'welcome'
    },
    {
      id: 'positioning',
      titleKey: 'onboarding.step_positioning_title',
      subtitleKey: 'onboarding.step_positioning_sub',
      pose: 'position'
    },
    {
      id: 'theme',
      titleKey: 'onboarding.step_theme_title',
      subtitleKey: 'onboarding.step_theme_sub',
      pose: 'theme'
    },
    {
      id: 'system',
      titleKey: 'onboarding.step_system_title',
      subtitleKey: 'onboarding.step_system_sub',
      pose: 'settings'
    },
    {
      id: 'ready',
      titleKey: 'onboarding.step_ready_title',
      subtitleKey: 'onboarding.step_ready_sub',
      pose: 'celebrate'
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

<div data-vinta="onboarding-page" class="onboarding-page-layout">
  <header class="onboarding-header">
    <div class="onboarding-mascot-wrap">
      <VintaMascot pose={stepData.pose} />
    </div>
    <div class="onboarding-header-titles">
      <h2 class="onboarding-title">{$tStore(stepData.titleKey)}</h2>
      <p class="onboarding-subtitle">{$tStore(stepData.subtitleKey)}</p>
    </div>
  </header>

  <div class="onboarding-body">
    {#if currentStep === 0}
      <div class="onboarding-feature-list">
        <div class="onboarding-inputs-row">
          <div class="onboarding-input-card">
            <span class="onboarding-input-label">
              {$tStore('settings.language_label')}
            </span>
            <CustomSelect
              options={localeOptions}
              bind:value={$localeStore}
              fullWidth={true}
            />
          </div>

          <div class="onboarding-input-card">
            <label for="onboarding-name-input" class="onboarding-input-label">
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

        <div class="onboarding-feature-item">
          <div class="feature-item-icon">
            <MousePointer size={14} />
          </div>
          <div class="feature-item-text">
            <span class="feature-item-title"
              >{$tStore('onboarding.feature_expand_title')}</span
            >
            <span class="feature-item-desc">
              {$tStore('onboarding.feature_expand_desc')}
            </span>
          </div>
        </div>

        <div class="onboarding-feature-item">
          <div class="feature-item-icon">
            <Move size={14} />
          </div>
          <div class="feature-item-text">
            <span class="feature-item-title"
              >{$tStore('onboarding.feature_move_title')}</span
            >
            <span class="feature-item-desc">
              {$tStore('onboarding.feature_move_desc')}
            </span>
          </div>
        </div>

        <div class="onboarding-feature-item">
          <div class="feature-item-icon">
            <LayoutGrid size={14} />
          </div>
          <div class="feature-item-text">
            <span class="feature-item-title"
              >{$tStore('onboarding.feature_glance_title')}</span
            >
            <span class="feature-item-desc">
              {$tStore('onboarding.feature_glance_desc')}
            </span>
          </div>
        </div>
      </div>
    {:else if currentStep === 1}
      <div class="onboarding-feature-list">
        <div class="onboarding-row-item">
          <div class="onboarding-row-label">
            <span class="row-label-title"
              >{$tStore('onboarding.resting_alignment')}</span
            >
            <span class="row-label-desc"
              >{$tStore('onboarding.resting_alignment_desc')}</span
            >
          </div>
          <div class="onboarding-action-buttons">
            {#each ['top-left', 'center', 'top-right'] as pos (pos)}
              <button
                type="button"
                class="onboarding-btn"
                class:primary={$hiddenPillPositionStore === pos}
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

        <div class="onboarding-row-item">
          <div class="onboarding-row-label">
            <span class="row-label-title"
              >{$tStore('onboarding.follow_screen')}</span
            >
            <span class="row-label-desc"
              >{$tStore('onboarding.follow_screen_desc')}</span
            >
          </div>
          <button
            type="button"
            class="onboarding-btn"
            class:primary={$followActiveDisplayStore}
            onclick={() =>
              ($followActiveDisplayStore = !$followActiveDisplayStore)}
          >
            {$followActiveDisplayStore
              ? $tStore('common.enabled')
              : $tStore('common.disabled')}
          </button>
        </div>
      </div>
    {:else if currentStep === 2}
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
    {:else if currentStep === 3}
      <div class="onboarding-feature-list">
        <div class="onboarding-row-item">
          <div class="onboarding-row-label">
            <span class="row-label-title"
              >{$tStore('onboarding.start_with_windows')}</span
            >
            <span class="row-label-desc"
              >{$tStore('onboarding.start_with_windows_desc')}</span
            >
          </div>
          <button
            type="button"
            class="onboarding-btn"
            class:primary={$autostartStore}
            onclick={handleToggleAutostart}
          >
            <Power size={13} />
            <span
              >{$autostartStore
                ? $tStore('common.enabled')
                : $tStore('common.disabled')}</span
            >
          </button>
        </div>

        <div class="onboarding-row-item">
          <div class="onboarding-row-label">
            <span class="row-label-title"
              >{$tStore('onboarding.system_stats')}</span
            >
            <span class="row-label-desc"
              >{$tStore('onboarding.system_stats_desc')}</span
            >
          </div>
          <button
            type="button"
            class="onboarding-btn"
            class:primary={Boolean($homeWidgetsVisibilityStore.systemStats)}
            onclick={handleToggleSystemStats}
          >
            <Activity size={13} />
            <span
              >{$homeWidgetsVisibilityStore.systemStats
                ? $tStore('common.enabled')
                : $tStore('common.disabled')}</span
            >
          </button>
        </div>
      </div>
    {:else if currentStep === 4}
      <div class="onboarding-feature-list">
        <div class="onboarding-feature-item">
          <div class="feature-item-icon">
            <Keyboard size={14} />
          </div>
          <div class="feature-item-text">
            <span class="feature-item-title"
              >{$tStore('onboarding.tip_shortcut_title')}</span
            >
            <span class="feature-item-desc">
              {$tStore('onboarding.tip_shortcut_desc', { key: 'Alt+Shift+V' })}
            </span>
          </div>
        </div>

        <div class="onboarding-feature-item">
          <div class="feature-item-icon">
            <SlidersHorizontal size={14} />
          </div>
          <div class="feature-item-text">
            <span class="feature-item-title"
              >{$tStore('onboarding.tip_satellite_title')}</span
            >
            <span class="feature-item-desc">
              {$tStore('onboarding.tip_satellite_desc')}
            </span>
          </div>
        </div>
      </div>
    {/if}
  </div>

  <footer class="onboarding-footer">
    <div class="onboarding-steps-indicator">
      {#each steps as s, idx (s.id)}
        <span class="step-indicator-dot" class:active={idx === currentStep}
        ></span>
      {/each}
    </div>

    <div class="onboarding-action-buttons">
      {#if currentStep > 0}
        <button type="button" class="onboarding-btn" onclick={handlePrev}>
          <ChevronLeft size={13} strokeWidth={2.2} />
          <span>{$tStore('common.back')}</span>
        </button>
      {:else}
        <button type="button" class="onboarding-btn" onclick={handleComplete}>
          <span>{$tStore('common.skip')}</span>
        </button>
      {/if}

      <button type="button" class="onboarding-btn primary" onclick={handleNext}>
        <span
          >{currentStep === steps.length - 1
            ? $tStore('onboarding.btn_get_started')
            : $tStore('common.continue')}</span
        >
        {#if currentStep < steps.length - 1}
          <ChevronRight size={13} strokeWidth={2.2} />
        {/if}
      </button>
    </div>
  </footer>
</div>

<style src="./styles/OnboardingPage.css"></style>
