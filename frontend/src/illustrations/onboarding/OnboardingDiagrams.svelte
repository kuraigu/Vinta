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
    SkipForward,
    Plus,
    Archive,
    Calendar,
    Bell,
    Check,
    SlidersHorizontal
  } from 'lucide-svelte'
  import IconBadge from '../../components/IconBadge.svelte'
  import BentoCard from '../../components/BentoCard.svelte'
  import SegmentedControl from '../../components/SegmentedControl.svelte'
  import { tStore, animScaleStore } from '../../store'

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

  function handleSelectWelcomeSlide(idx: number) {
    welcomeStartIndex = idx
    welcomeCycleKey++
  }

  function welcomeDelay(slideIdx: number): string {
    return `${((slideIdx - welcomeStartIndex + 7) % 7) * 3}s`
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
        role="region"
        aria-label={$tStore('onboarding.step_welcome_title')}
      >
        <!-- Slide 1: Media Controls -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={$animScaleStore <= 0 &&
            welcomeStartIndex === 0}
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
              <Play size={10} strokeWidth={2.5} />
            </span>
            <span class="mini-action-btn">
              <SkipForward size={10} strokeWidth={2.4} />
            </span>
          </div>
        </div>

        <!-- Slide 2: Timer & Focus -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={$animScaleStore <= 0 &&
            welcomeStartIndex === 1}
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
              <Play size={10} strokeWidth={2.5} />
            </span>
          </div>
        </div>

        <!-- Slide 3: Quick Notes -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={$animScaleStore <= 0 &&
            welcomeStartIndex === 2}
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
              <Plus size={10} strokeWidth={2.6} />
            </span>
          </div>
        </div>

        <!-- Slide 4: File Shelf -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={$animScaleStore <= 0 &&
            welcomeStartIndex === 3}
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
              <Archive size={10} strokeWidth={2.4} />
            </span>
          </div>
        </div>

        <!-- Slide 5: Calendar & Agenda -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={$animScaleStore <= 0 &&
            welcomeStartIndex === 4}
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
              <Plus size={10} strokeWidth={2.6} />
            </span>
          </div>
        </div>

        <!-- Slide 6: Notifications -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={$animScaleStore <= 0 &&
            welcomeStartIndex === 5}
          style:animation-delay={welcomeDelay(5)}
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
              <Check size={10} strokeWidth={2.6} />
            </span>
          </div>
        </div>

        <!-- Slide 7: Customize & Settings -->
        <div
          class="welcome-feature-slide"
          class:is-static-active={$animScaleStore <= 0 &&
            welcomeStartIndex === 6}
          style:animation-delay={welcomeDelay(6)}
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
              <SlidersHorizontal size={10} strokeWidth={2.4} />
            </span>
          </div>
        </div>
      </div>

      <div
        class="welcome-progress-dots"
        class:animations-paused={$animScaleStore <= 0}
        role="tablist"
      >
        {#each ['onboarding.welcome_feat_media_title', 'onboarding.welcome_feat_timer_title', 'onboarding.welcome_feat_notes_title', 'onboarding.welcome_feat_shelf_title', 'onboarding.welcome_feat_calendar_title', 'onboarding.welcome_feat_notif_title', 'onboarding.welcome_feat_customize_title'] as labelKey, idx (labelKey)}
          <button
            type="button"
            role="tab"
            class="welcome-dot-btn"
            aria-label={$tStore(labelKey)}
            onclick={() => handleSelectWelcomeSlide(idx)}
          >
            <span
              class="welcome-dot-bar"
              class:is-static-active={$animScaleStore <= 0 &&
                welcomeStartIndex === idx}
              style:animation-delay={welcomeDelay(idx)}
            ></span>
          </button>
        {/each}
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
                <Sun size={10} strokeWidth={2.4} />
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
    <SegmentedControl
      size="sm"
      fullWidth={true}
      options={stateOptions}
      value={effectiveStateMode}
      onchange={(val) => (selectedStatePreview = val as StatePreviewMode)}
    />

    <div
      class="mini-desktop-screen states-stage mode-{effectiveStateMode} hidden-{hiddenDisplay}"
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
                      <Music size={10} strokeWidth={2.2} />
                    </span>
                    <span class="morph-tile-sub">Playing</span>
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
                  <Sun size={10} strokeWidth={2.3} />
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
            <span class="morph-sat-btn highlight">
              <Minimize2 size={9} strokeWidth={2.4} />
            </span>
            <span class="morph-sat-btn highlight">
              <EyeOff size={9} strokeWidth={2.4} />
            </span>
            <span class="morph-sat-divider"></span>
            <span class="morph-sat-btn danger">
              <Power size={9} strokeWidth={2.4} />
            </span>
          </div>
        </div>

        <!-- Animated White Pointer Cursor across States -->
        <div class="morph-state-cursor wake-{wakeMethod}" aria-hidden="true">
          <span class="cursor-click-ring"></span>
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
            {$tStore('onboarding.diag_expanded_hint')}
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
  </BentoCard>
{/if}

<style src="../../components/styles/OnboardingDiagrams.css"></style>
