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
    Check,
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
    syncSettingsToDisk
  } from '../store'
  import { SetAutostart } from '../../bindings/Vinta/backend/app.js'
  import HomePage, { config as homeConfig } from './HomePage.svelte'
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
    title: string
    subtitle: string
    pose: MascotPose
  }[] = [
    {
      id: 'welcome',
      title: 'Welcome to Vinta',
      subtitle: 'Your dynamic island assistant for Windows',
      pose: 'welcome'
    },
    {
      id: 'positioning',
      title: 'Screen & Alignment',
      subtitle: 'Choose where the island rests on your display',
      pose: 'position'
    },
    {
      id: 'theme',
      title: 'Pick a Theme',
      subtitle: 'Choose a style that matches your desktop',
      pose: 'theme'
    },
    {
      id: 'system',
      title: 'Quick Settings',
      subtitle: 'A few helpful preferences you can change anytime',
      pose: 'settings'
    },
    {
      id: 'ready',
      title: "You're All Set",
      subtitle: 'Vinta is ready to use on your PC',
      pose: 'celebrate'
    }
  ]

  let stepData = $derived(steps[currentStep] || steps[0])

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
      <h2 class="onboarding-title">{stepData.title}</h2>
      <p class="onboarding-subtitle">{stepData.subtitle}</p>
    </div>
  </header>

  <div class="onboarding-body">
    {#if currentStep === 0}
      <div class="onboarding-feature-list">
        <div class="onboarding-input-card">
          <label for="onboarding-name-input" class="onboarding-input-label">
            What should we call you?
          </label>
          <div class="onboarding-input-wrapper">
            <User size={13} class="input-leading-icon" />
            <input
              id="onboarding-name-input"
              type="text"
              class="onboarding-name-field"
              placeholder="Your name (e.g. Alex)"
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

        <div class="onboarding-feature-item">
          <div class="feature-item-icon">
            <MousePointer size={14} />
          </div>
          <div class="feature-item-text">
            <span class="feature-item-title">Click to expand</span>
            <span class="feature-item-desc">
              Click the top pill anytime to open widgets, or hover to take a quick peek.
            </span>
          </div>
        </div>

        <div class="onboarding-feature-item">
          <div class="feature-item-icon">
            <Move size={14} />
          </div>
          <div class="feature-item-text">
            <span class="feature-item-title">Move it anywhere</span>
            <span class="feature-item-desc">
              Drag the top handle to place the island wherever you prefer.
            </span>
          </div>
        </div>

        <div class="onboarding-feature-item">
          <div class="feature-item-icon">
            <LayoutGrid size={14} />
          </div>
          <div class="feature-item-text">
            <span class="feature-item-title">Glance at what matters</span>
            <span class="feature-item-desc">
              Check your calendar, weather, music player, and quick controls at a glance.
            </span>
          </div>
        </div>
      </div>
    {:else if currentStep === 1}
      <div class="onboarding-feature-list">
        <div class="onboarding-row-item">
          <div class="onboarding-row-label">
            <span class="row-label-title">Resting Alignment</span>
            <span class="row-label-desc">Where the island docks when idle</span>
          </div>
          <div class="onboarding-action-buttons">
            {#each ['top-left', 'center', 'top-right'] as pos (pos)}
              <button
                type="button"
                class="onboarding-btn"
                class:primary={$hiddenPillPositionStore === pos}
                onclick={() => ($hiddenPillPositionStore = pos as any)}
              >
                {pos === 'top-left' ? 'Left' : pos === 'top-right' ? 'Right' : 'Center'}
              </button>
            {/each}
          </div>
        </div>

        <div class="onboarding-row-item">
          <div class="onboarding-row-label">
            <span class="row-label-title">Follow Active Screen</span>
            <span class="row-label-desc">Move the island to whichever monitor you are working on</span>
          </div>
          <button
            type="button"
            class="onboarding-btn"
            class:primary={$followActiveDisplayStore}
            onclick={() => ($followActiveDisplayStore = !$followActiveDisplayStore)}
          >
            {$followActiveDisplayStore ? 'Enabled' : 'Disabled'}
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
            <span class="row-label-title">Start with Windows</span>
            <span class="row-label-desc">Launch Vinta automatically on startup</span>
          </div>
          <button
            type="button"
            class="onboarding-btn"
            class:primary={$autostartStore}
            onclick={handleToggleAutostart}
          >
            <Power size={13} />
            <span>{$autostartStore ? 'Enabled' : 'Disabled'}</span>
          </button>
        </div>

        <div class="onboarding-row-item">
          <div class="onboarding-row-label">
            <span class="row-label-title">System Stats</span>
            <span class="row-label-desc">Show CPU, RAM, and GPU stats in your home feed</span>
          </div>
          <button
            type="button"
            class="onboarding-btn"
            class:primary={Boolean($homeWidgetsVisibilityStore.systemStats)}
            onclick={handleToggleSystemStats}
          >
            <Activity size={13} />
            <span>{$homeWidgetsVisibilityStore.systemStats ? 'Enabled' : 'Disabled'}</span>
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
            <span class="feature-item-title">Keyboard shortcut</span>
            <span class="feature-item-desc">
              Press <strong>Alt+Shift+V</strong> anytime to quickly expand or hide the island.
            </span>
          </div>
        </div>

        <div class="onboarding-feature-item">
          <div class="feature-item-icon">
            <SlidersHorizontal size={14} />
          </div>
          <div class="feature-item-text">
            <span class="feature-item-title">Control dock satellite</span>
            <span class="feature-item-desc">
              You can also open quick controls and settings anytime using the floating control dock.
            </span>
          </div>
        </div>
      </div>
    {/if}
  </div>

  <footer class="onboarding-footer">
    <div class="onboarding-steps-indicator">
      {#each steps as s, idx (s.id)}
        <span
          class="step-indicator-dot"
          class:active={idx === currentStep}
        ></span>
      {/each}
    </div>

    <div class="onboarding-action-buttons">
      {#if currentStep > 0}
        <button
          type="button"
          class="onboarding-btn"
          onclick={handlePrev}
        >
          <ChevronLeft size={13} strokeWidth={2.2} />
          <span>Back</span>
        </button>
      {:else}
        <button
          type="button"
          class="onboarding-btn"
          onclick={handleComplete}
        >
          <span>Skip</span>
        </button>
      {/if}

      <button
        type="button"
        class="onboarding-btn primary"
        onclick={handleNext}
      >
        <span>{currentStep === steps.length - 1 ? 'Get Started' : 'Continue'}</span>
        {#if currentStep < steps.length - 1}
          <ChevronRight size={13} strokeWidth={2.2} />
        {/if}
      </button>
    </div>
  </footer>
</div>

<style src="./styles/OnboardingPage.css"></style>

