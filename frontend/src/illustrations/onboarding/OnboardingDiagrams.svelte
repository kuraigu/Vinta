<!--
  @component OnboardingDiagrams
  @description Animated token-driven desktop previews for onboarding.
  Renders a cursor-drag animation for island repositioning and a clean 3-state
  morphing preview (Expanded + Control Satellite, Compact Pill, Hidden edge).
-->
<script lang="ts">
  import {
    Minimize2,
    EyeOff,
    Power,
    Move,
    Sun,
    Music,
    Timer,
    StickyNote,
    FolderInput,
    Play,
    Pause,
    SkipForward,
    Plus,
    Archive,
    Calendar,
    Bell,
    Check,
    SlidersHorizontal,
    Pin,
    BookOpen,
    Terminal,
    LayoutGrid,
    Settings
  } from 'lucide-svelte'
  import IconBadge from '../../components/IconBadge.svelte'
  import BentoCard from '../../components/BentoCard.svelte'
  import SegmentedControl from '../../components/SegmentedControl.svelte'
  import {
    tStore,
    animScaleStore,
    clickOutsideMinimizeStore
  } from '../../store'

  export type StatePreviewMode = 'loop' | 'expanded' | 'compact' | 'hidden'

  interface Props {
    mode: 'welcome' | 'position' | 'states'
    hiddenDisplay?: 'dash' | 'clock'
    wakeMethod?: 'click' | 'hover'
  }

  let { mode, hiddenDisplay = 'dash', wakeMethod = 'click' }: Props = $props()

  let selectedStatePreview = $state<StatePreviewMode>('loop')
  let welcomeStartIndex = $state(0)
  let welcomeCycleKey = $state(0)
  let isWelcomePaused = $state(false)
  let isStatesPaused = $state(false)

  function handleSelectWelcomeSlide(idx: number) {
    welcomeStartIndex = idx
    welcomeCycleKey++
  }

  function welcomeDelay(slideIdx: number): string {
    return `${((slideIdx - welcomeStartIndex + 9) % 9) * 3}s`
  }

  let effectiveStateMode = $derived<StatePreviewMode>(
    $animScaleStore <= 0 && selectedStatePreview === 'loop'
      ? 'expanded'
      : selectedStatePreview
  )

  let stateOptions = $derived([
    ...($animScaleStore > 0
      ? [{ value: 'loop', label: $tStore('onboarding.state_auto') }]
      : []),
    { value: 'expanded', label: $tStore('onboarding.state_expanded') },
    { value: 'compact', label: $tStore('onboarding.state_compact') },
    { value: 'hidden', label: $tStore('onboarding.state_hidden') }
  ])
</script>

{#if mode === 'welcome'}
  <BentoCard
    variant="bento"
    class="onboarding-diagram-card welcome-showcase-card"
  >
    {#key welcomeCycleKey}
      <div
        class="welcome-feature-stage"
        class:animations-paused={$animScaleStore <= 0}
        class:user-paused={isWelcomePaused}
        role="region"
        aria-label={$tStore('onboarding.step_welcome_title')}
      >
        <!-- Slide 1: Media Controls -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={welcomeStartIndex === 0}
          style:animation-delay={welcomeDelay(0)}
        >
          <div class="welcome-slide-left">
            <IconBadge icon={Music} color="rose" size="md" />
            <div class="welcome-slide-text">
              <span class="welcome-slide-title">
                {$tStore('onboarding.welcome_feat_media_title')}
              </span>
              <span class="welcome-slide-action">
                {$tStore('onboarding.welcome_feat_media_action')}
              </span>
            </div>
          </div>
          <div class="welcome-action-pill action-media" aria-hidden="true">
            <span class="mini-eq-bars">
              <span></span><span></span><span></span>
            </span>
            <span class="mini-action-btn press-pulse">
              <Play size={12} strokeWidth={2.4} />
            </span>
            <span class="mini-action-btn">
              <SkipForward size={12} strokeWidth={2.4} />
            </span>
          </div>
        </div>

        <!-- Slide 2: Timer & Focus -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={welcomeStartIndex === 1}
          style:animation-delay={welcomeDelay(1)}
        >
          <div class="welcome-slide-left">
            <IconBadge icon={Timer} color="amber" size="md" />
            <div class="welcome-slide-text">
              <span class="welcome-slide-title">
                {$tStore('onboarding.welcome_feat_timer_title')}
              </span>
              <span class="welcome-slide-action">
                {$tStore('onboarding.welcome_feat_timer_action')}
              </span>
            </div>
          </div>
          <div class="welcome-action-pill action-timer" aria-hidden="true">
            <span class="mini-timer-readout">25:00</span>
            <span class="mini-action-btn press-pulse">
              <Play size={12} strokeWidth={2.4} />
            </span>
          </div>
        </div>

        <!-- Slide 3: Quick Notes -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={welcomeStartIndex === 2}
          style:animation-delay={welcomeDelay(2)}
        >
          <div class="welcome-slide-left">
            <IconBadge icon={StickyNote} color="purple" size="md" />
            <div class="welcome-slide-text">
              <span class="welcome-slide-title">
                {$tStore('onboarding.welcome_feat_notes_title')}
              </span>
              <span class="welcome-slide-action">
                {$tStore('onboarding.welcome_feat_notes_action')}
              </span>
            </div>
          </div>
          <div class="welcome-action-pill action-notes" aria-hidden="true">
            <span class="mini-note-lines">
              <span></span><span></span>
            </span>
            <span class="mini-action-btn press-pulse">
              <Plus size={12} strokeWidth={2.5} />
            </span>
          </div>
        </div>

        <!-- Slide 4: File Shelf -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={welcomeStartIndex === 3}
          style:animation-delay={welcomeDelay(3)}
        >
          <div class="welcome-slide-left">
            <IconBadge icon={FolderInput} color="emerald" size="md" />
            <div class="welcome-slide-text">
              <span class="welcome-slide-title">
                {$tStore('onboarding.welcome_feat_shelf_title')}
              </span>
              <span class="welcome-slide-action">
                {$tStore('onboarding.welcome_feat_shelf_action')}
              </span>
            </div>
          </div>
          <div class="welcome-action-pill action-shelf" aria-hidden="true">
            <span class="mini-file-chip">.pdf</span>
            <span class="mini-action-btn press-pulse">
              <Archive size={12} strokeWidth={2.3} />
            </span>
          </div>
        </div>

        <!-- Slide 5: Calendar & Agenda -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={welcomeStartIndex === 4}
          style:animation-delay={welcomeDelay(4)}
        >
          <div class="welcome-slide-left">
            <IconBadge icon={Calendar} color="blue" size="md" />
            <div class="welcome-slide-text">
              <span class="welcome-slide-title">
                {$tStore('onboarding.welcome_feat_calendar_title')}
              </span>
              <span class="welcome-slide-action">
                {$tStore('onboarding.welcome_feat_calendar_action')}
              </span>
            </div>
          </div>
          <div class="welcome-action-pill action-calendar" aria-hidden="true">
            <span class="mini-timer-readout">10:00</span>
            <span class="mini-action-btn press-pulse">
              <Plus size={12} strokeWidth={2.5} />
            </span>
          </div>
        </div>

        <!-- Slide 6: Flashcards & Daily Habits -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={welcomeStartIndex === 5}
          style:animation-delay={welcomeDelay(5)}
        >
          <div class="welcome-slide-left">
            <IconBadge icon={BookOpen} color="emerald" size="md" />
            <div class="welcome-slide-text">
              <span class="welcome-slide-title">
                {$tStore('onboarding.welcome_feat_companion_title')}
              </span>
              <span class="welcome-slide-action">
                {$tStore('onboarding.welcome_feat_companion_action')}
              </span>
            </div>
          </div>
          <div class="welcome-action-pill action-companion" aria-hidden="true">
            <span class="mini-file-chip">4/5</span>
            <span class="mini-action-btn press-pulse">
              <Check size={12} strokeWidth={2.5} />
            </span>
          </div>
        </div>

        <!-- Slide 7: Developer Hub & Cloud Sync -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={welcomeStartIndex === 6}
          style:animation-delay={welcomeDelay(6)}
        >
          <div class="welcome-slide-left">
            <IconBadge icon={Terminal} color="purple" size="md" />
            <div class="welcome-slide-text">
              <span class="welcome-slide-title">
                {$tStore('onboarding.welcome_feat_devhub_title')}
              </span>
              <span class="welcome-slide-action">
                {$tStore('onboarding.welcome_feat_devhub_action')}
              </span>
            </div>
          </div>
          <div class="welcome-action-pill action-devhub" aria-hidden="true">
            <span class="mini-file-chip">PR</span>
            <span class="mini-action-btn press-pulse">
              <Terminal size={12} strokeWidth={2.3} />
            </span>
          </div>
        </div>

        <!-- Slide 8: Notifications -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={welcomeStartIndex === 7}
          style:animation-delay={welcomeDelay(7)}
        >
          <div class="welcome-slide-left">
            <IconBadge icon={Bell} color="indigo" size="md" />
            <div class="welcome-slide-text">
              <span class="welcome-slide-title">
                {$tStore('onboarding.welcome_feat_notif_title')}
              </span>
              <span class="welcome-slide-action">
                {$tStore('onboarding.welcome_feat_notif_action')}
              </span>
            </div>
          </div>
          <div class="welcome-action-pill action-notif" aria-hidden="true">
            <span class="mini-notif-dot"></span>
            <span class="mini-file-chip">2</span>
            <span class="mini-action-btn press-pulse">
              <Check size={12} strokeWidth={2.5} />
            </span>
          </div>
        </div>

        <!-- Slide 9: Customize & Settings -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={welcomeStartIndex === 8}
          style:animation-delay={welcomeDelay(8)}
        >
          <div class="welcome-slide-left">
            <IconBadge icon={SlidersHorizontal} color="cyan" size="md" />
            <div class="welcome-slide-text">
              <span class="welcome-slide-title">
                {$tStore('onboarding.welcome_feat_customize_title')}
              </span>
              <span class="welcome-slide-action">
                {$tStore('onboarding.welcome_feat_customize_action')}
              </span>
            </div>
          </div>
          <div class="welcome-action-pill action-customize" aria-hidden="true">
            <span class="mini-swatch-dots">
              <span></span><span></span>
            </span>
            <span class="mini-action-btn press-pulse">
              <SlidersHorizontal size={12} strokeWidth={2.3} />
            </span>
          </div>
        </div>
      </div>

      <div class="welcome-controls-row">
        <div
          class="welcome-progress-dots"
          class:animations-paused={$animScaleStore <= 0}
          class:user-paused={isWelcomePaused}
          role="group"
          aria-label={$tStore('onboarding.welcome_slides_aria')}
        >
          {#each ['onboarding.welcome_feat_media_title', 'onboarding.welcome_feat_timer_title', 'onboarding.welcome_feat_notes_title', 'onboarding.welcome_feat_shelf_title', 'onboarding.welcome_feat_calendar_title', 'onboarding.welcome_feat_companion_title', 'onboarding.welcome_feat_devhub_title', 'onboarding.welcome_feat_notif_title', 'onboarding.welcome_feat_customize_title'] as labelKey, idx (labelKey)}
            <button
              type="button"
              class="welcome-dot-btn"
              aria-label={$tStore(labelKey)}
              aria-current={welcomeStartIndex === idx ? 'true' : undefined}
              onclick={() => handleSelectWelcomeSlide(idx)}
            >
              <span
                class="welcome-dot-bar"
                class:is-static-active={welcomeStartIndex === idx}
                style:animation-delay={welcomeDelay(idx)}
              ></span>
            </button>
          {/each}
        </div>
        {#if $animScaleStore > 0}
          <button
            type="button"
            class="diagram-motion-toggle"
            aria-label={isWelcomePaused
              ? $tStore('onboarding.diagram_play')
              : $tStore('onboarding.diagram_pause')}
            onclick={() => (isWelcomePaused = !isWelcomePaused)}
          >
            {#if isWelcomePaused}
              <Play size={12} strokeWidth={2.3} />
            {:else}
              <Pause size={12} strokeWidth={2.3} />
            {/if}
          </button>
        {/if}
      </div>
    {/key}
  </BentoCard>
{:else if mode === 'position'}
  <BentoCard variant="bento" class="onboarding-diagram-card">
    <div
      class="mini-desktop-screen position-stage"
      role="img"
      aria-label={$tStore('onboarding.position_drag_title')}
    >
      <div class="mini-bg-window" aria-hidden="true">
        <div class="mini-bg-dots">
          <span></span><span></span><span></span>
        </div>
        <div class="mini-bg-lines">
          <span></span><span></span>
        </div>
      </div>

      <!-- Animated Dragging Island + Cursor -->
      <div class="drag-demo-track" aria-hidden="true">
        <div class="drag-demo-island-group">
          <div class="mini-island-pill-preview">
            <div class="mini-drag-bar"></div>
            <div class="mini-pill-row">
              <span>9:41</span>
              <span class="mini-pill-meta">
                <Sun size={12} strokeWidth={2.3} />
                <span>24°</span>
              </span>
            </div>
          </div>

          <div class="mini-satellite-preview">
            <span class="mini-sat-dot"></span>
            <span class="mini-sat-dot accent"></span>
            <span class="mini-sat-dot danger"></span>
          </div>

          <div class="drag-demo-cursor">
            <svg viewBox="0 0 18 22" width="15" height="18" aria-hidden="true">
              <path
                d="M2 1.5L15.5 11.2L9.4 12.1L12.6 18.6L10.2 19.7L7 13.2L2.5 16.8Z"
                fill="#ffffff"
                stroke="rgba(15, 23, 42, 0.78)"
                stroke-width="1.5"
                stroke-linejoin="round"
              />
            </svg>
          </div>
        </div>
      </div>
    </div>

    <div class="diagram-drag-footer">
      <Move size={12} strokeWidth={2.2} />
      <span>{$tStore('onboarding.diag_drag_hint')}</span>
    </div>
  </BentoCard>
{:else}
  <BentoCard variant="bento" class="onboarding-diagram-card">
    <div class="states-diagram-toolbar">
      <div class="states-diagram-segmented">
        <SegmentedControl
          size="sm"
          fullWidth={true}
          options={stateOptions}
          value={effectiveStateMode}
          onchange={(val) => (selectedStatePreview = val as StatePreviewMode)}
        />
      </div>
      {#if effectiveStateMode === 'loop' && $animScaleStore > 0}
        <button
          type="button"
          class="diagram-motion-toggle"
          aria-label={isStatesPaused
            ? $tStore('onboarding.diagram_play')
            : $tStore('onboarding.diagram_pause')}
          onclick={() => (isStatesPaused = !isStatesPaused)}
        >
          {#if isStatesPaused}
            <Play size={12} strokeWidth={2.3} />
          {:else}
            <Pause size={12} strokeWidth={2.3} />
          {/if}
        </button>
      {/if}
    </div>

    {#key `${effectiveStateMode}-${hiddenDisplay}-${wakeMethod}-${$clickOutsideMinimizeStore}`}
      <div
        class="mini-desktop-screen states-stage mode-{effectiveStateMode} hidden-{hiddenDisplay}"
        class:click-outside-enabled={$clickOutsideMinimizeStore}
        class:user-paused={isStatesPaused}
        role="region"
        aria-label={$tStore('onboarding.step_states_title')}
      >
        <div class="mini-bg-window" aria-hidden="true">
          <div class="mini-bg-dots">
            <span></span><span></span><span></span>
          </div>
          <div class="mini-bg-lines">
            <span></span><span></span>
          </div>
        </div>

        <div class="morph-stage-center">
          <div class="morph-island-cluster">
            <!-- Single Morphing Island Container -->
            <div class="morph-island-shell">
              <!-- Expanded View Layer -->
              <div class="morph-layer-slot slot-expanded">
                <div class="morph-expanded-inner">
                  <div class="morph-drag-handle"></div>
                  <div class="morph-bento-pair">
                    <div class="morph-bento-tile">
                      <span class="morph-tile-title">9:41</span>
                      <span class="morph-tile-sub">24°</span>
                    </div>
                    <div class="morph-bento-tile">
                      <span class="morph-tile-title">
                        <Music size={12} strokeWidth={2.2} />
                      </span>
                      <span class="morph-tile-sub"
                        >{$tStore('onboarding.diag_playing')}</span
                      >
                    </div>
                  </div>
                  <div class="morph-dock-strip">
                    <span></span><span></span><span></span><span></span>
                  </div>
                </div>
              </div>

              <!-- Compact Pill View Layer -->
              <div class="morph-layer-slot slot-compact">
                <div class="morph-compact-row">
                  <span>9:41</span>
                  <span class="morph-compact-right">
                    <Sun size={12} strokeWidth={2.3} />
                    <span>24°</span>
                  </span>
                </div>
                <div class="morph-timeout-track"></div>
              </div>

              <!-- Hidden Edge View Layer -->
              <div class="morph-layer-slot slot-hidden">
                {#if hiddenDisplay === 'clock'}
                  <div class="morph-hidden-clock">
                    <span>9:41</span>
                    <span class="morph-hidden-clock-dot"></span>
                  </div>
                {:else}
                  <div class="morph-hidden-dash"></div>
                {/if}
              </div>
            </div>

            <!-- Floating Right Control Satellite -->
            <div class="morph-satellite-dock" aria-hidden="true">
              {#if $clickOutsideMinimizeStore}
                <span class="morph-sat-btn highlight morph-sat-pin">
                  <Pin size={12} strokeWidth={2.2} />
                </span>
              {/if}
              <span class="morph-sat-btn highlight">
                <Minimize2 size={12} strokeWidth={2.2} />
              </span>
              <span class="morph-sat-btn highlight">
                <EyeOff size={12} strokeWidth={2.2} />
              </span>
              <span class="morph-sat-divider"></span>
              <span class="morph-sat-btn danger">
                <Power size={12} strokeWidth={2.2} />
              </span>
            </div>
          </div>

          <!-- Miniature Right-Click Context Menu Popover -->
          <div class="morph-context-menu" aria-hidden="true">
            <div class="morph-ctx-item">
              <Pin size={12} strokeWidth={2.2} />
              <span>{$tStore('onboarding.diag_ctx_pin')}</span>
            </div>
            <div class="morph-ctx-item">
              <LayoutGrid size={12} strokeWidth={2.2} />
              <span>{$tStore('onboarding.diag_ctx_apps')}</span>
            </div>
            <div class="morph-ctx-item">
              <Settings size={12} strokeWidth={2.2} />
              <span>{$tStore('onboarding.diag_ctx_settings')}</span>
            </div>
            <div class="morph-ctx-divider"></div>
            <div class="morph-ctx-item is-hide">
              <EyeOff size={12} strokeWidth={2.2} />
              <span>{$tStore('onboarding.diag_ctx_hide')}</span>
            </div>
          </div>

          <!-- Animated White Pointer Cursor across States -->
          <div class="morph-state-cursor wake-{wakeMethod}" aria-hidden="true">
            <span class="cursor-click-ring"></span>
            <span class="cursor-click-badge cursor-click-left">
              {$tStore('onboarding.diag_click_left')}
            </span>
            <span class="cursor-click-badge cursor-click-right">
              {$tStore('onboarding.diag_click_right')}
            </span>
            <svg viewBox="0 0 18 22" width="15" height="18">
              <path
                d="M2 1.5L15.5 11.2L9.4 12.1L12.6 18.6L10.2 19.7L7 13.2L2.5 16.8Z"
                fill="#ffffff"
                stroke="rgba(15, 23, 42, 0.78)"
                stroke-width="1.5"
                stroke-linejoin="round"
              />
            </svg>
          </div>

          <!-- Single-Line Bottom Hint Zone (Consistent Styling) -->
          <div class="morph-stage-hints hints-expanded">
            <span class="stage-hint-chip">
              {$clickOutsideMinimizeStore
                ? $tStore('onboarding.diag_expanded_hint_outside')
                : $tStore('onboarding.diag_expanded_hint')}
            </span>
          </div>

          <div class="morph-stage-hints hints-compact">
            <span class="stage-hint-chip">
              {$tStore('onboarding.diag_compact_hint')}
            </span>
          </div>

          <div class="morph-stage-hints hints-hidden">
            <span class="stage-hint-chip">
              {wakeMethod === 'hover'
                ? $tStore('onboarding.diag_wake_hover')
                : $tStore('onboarding.diag_wake_click')}
            </span>
          </div>
        </div>
      </div>
    {/key}
  </BentoCard>
{/if}

<style src="../../components/styles/OnboardingDiagrams.css"></style>
