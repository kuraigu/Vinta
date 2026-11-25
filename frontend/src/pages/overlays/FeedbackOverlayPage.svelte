<!--
  @fileoverview Overlay page for submitting pre-filled Bug Reports and Feature Requests
  with automatic sanitized desktop diagnostics.
-->
<script lang="ts">
  import { onMount } from 'svelte'
  import {
    AlertTriangle,
    ArrowUpRight,
    Bug,
    ImagePlus,
    Lightbulb,
    Monitor,
    RotateCcw
  } from 'lucide-svelte'
  import OverlayPage from '../../components/OverlayPage.svelte'
  import HorizontalTabSelection, {
    type TabItem
  } from '../../components/HorizontalTabSelection.svelte'
  import ScrollView from '../../components/ScrollView.svelte'
  import SettingActionButton from '../../components/settings/SettingActionButton.svelte'
  import {
    getSystemDiagnostics,
    reportFeedback,
    requestFeature,
    type SystemDiagnostics
  } from '../../services/feedbackService'
  import {
    clearFeedbackDraft,
    feedbackDraftsStore,
    feedbackOverlayKindStore,
    tStore,
    type FeedbackDraft
  } from '../../store'

  interface Props {
    onClose: () => void
  }

  let { onClose }: Props = $props()

  let diagnostics = $state<SystemDiagnostics | null>(null)

  let activeDraft = $derived(
    $feedbackDraftsStore[$feedbackOverlayKindStore] ?? {
      title: '',
      describe: '',
      extra: ''
    }
  )

  function updateDraftField(field: keyof FeedbackDraft, value: string): void {
    const kind = $feedbackOverlayKindStore
    feedbackDraftsStore.update((drafts) => ({
      ...drafts,
      [kind]: {
        ...drafts[kind],
        [field]: value
      }
    }))
  }

  function hasRepeatedBlock(raw: string): boolean {
    const trimmed = raw.trim()
    if (trimmed.length < 8) return false

    const lines = trimmed
      .split(/\r?\n/)
      .map((l) => l.trim().toLowerCase().replace(/\s+/g, ' '))
      .filter((l) => l.length >= 4)

    if (lines.length >= 2) {
      for (let i = 1; i < lines.length; i++) {
        if (lines[i] === lines[i - 1]) return true
      }
      if (lines.length % 2 === 0) {
        const half = lines.length / 2
        if (lines.slice(0, half).join('\n') === lines.slice(half).join('\n')) {
          return true
        }
      }
    }

    const normalized = trimmed.toLowerCase().replace(/\s+/g, ' ')
    const mid = Math.floor(normalized.length / 2)
    if (
      normalized.length % 2 === 1 &&
      normalized[mid] === ' ' &&
      normalized.slice(0, mid) === normalized.slice(mid + 1)
    ) {
      return true
    }
    if (
      normalized.length % 2 === 0 &&
      normalized.slice(0, mid) === normalized.slice(mid)
    ) {
      return true
    }
    return false
  }

  let hasDuplicateContent = $derived.by(() => {
    const norm = (s: string) => s.trim().toLowerCase().replace(/\s+/g, ' ')
    const t = norm(activeDraft.title)
    const d = norm(activeDraft.describe)
    const e = norm(activeDraft.extra)

    if (
      (t.length >= 3 && d.length >= 3 && t === d) ||
      (d.length >= 3 && e.length >= 3 && d === e) ||
      (t.length >= 3 && e.length >= 3 && t === e)
    ) {
      return true
    }

    return (
      hasRepeatedBlock(activeDraft.title) ||
      hasRepeatedBlock(activeDraft.describe) ||
      hasRepeatedBlock(activeDraft.extra)
    )
  })

  let hasAnyContent = $derived(
    activeDraft.title.trim().length > 0 ||
      activeDraft.describe.trim().length > 0 ||
      activeDraft.extra.trim().length > 0
  )

  let headerTabs = $derived<TabItem[]>([
    { id: 'bug', label: $tStore('about.feedback_type_bug'), icon: Bug },
    {
      id: 'feature',
      label: $tStore('about.feedback_type_feature'),
      icon: Lightbulb
    }
  ])

  let canSubmit = $derived(
    activeDraft.title.trim().length > 0 &&
      activeDraft.describe.trim().length > 0 &&
      !hasDuplicateContent
  )

  function handleClear() {
    clearFeedbackDraft($feedbackOverlayKindStore)
  }

  function handleSubmit() {
    if (!canSubmit) return
    const { title, describe, extra } = activeDraft
    if ($feedbackOverlayKindStore === 'bug') {
      reportFeedback(title, describe, extra).catch(() => {})
    } else {
      requestFeature(title, describe, extra).catch(() => {})
    }
  }

  onMount(() => {
    getSystemDiagnostics()
      .then((d) => {
        diagnostics = d
      })
      .catch(() => {})
  })
</script>

<OverlayPage
  title={$feedbackOverlayKindStore === 'bug'
    ? $tStore('about.feedback_bug_title')
    : $tStore('about.feedback_feature_title')}
  dataVinta="feedback-overlay"
  {onClose}
>
  {#snippet headerLeft()}
    <HorizontalTabSelection
      tabs={headerTabs}
      bind:value={$feedbackOverlayKindStore}
      colorVariant="subtle"
      ariaLabel={$tStore('about.report_issue')}
    />
  {/snippet}

  <div class="feedback-overlay-shell">
    <ScrollView class="feedback-overlay-scroll">
      <div class="feedback-overlay-body">
        <div class="feedback-fields">
          <label class="feedback-field">
            <span class="feedback-label">
              {$tStore('about.feedback_title_label')}
              <span class="feedback-required" aria-hidden="true">*</span>
            </span>
            <input
              type="text"
              class="feedback-input"
              value={activeDraft.title}
              oninput={(e) => updateDraftField('title', e.currentTarget.value)}
              placeholder={$feedbackOverlayKindStore === 'bug'
                ? $tStore('about.feedback_title_bug_placeholder')
                : $tStore('about.feedback_title_feature_placeholder')}
            />
          </label>

          <label class="feedback-field">
            <span class="feedback-label">
              {$feedbackOverlayKindStore === 'bug'
                ? $tStore('about.feedback_describe_bug')
                : $tStore('about.feedback_describe_feature')}
              <span class="feedback-required" aria-hidden="true">*</span>
            </span>
            <textarea
              rows="3"
              class="feedback-textarea"
              value={activeDraft.describe}
              oninput={(e) =>
                updateDraftField('describe', e.currentTarget.value)}
              placeholder={$feedbackOverlayKindStore === 'bug'
                ? $tStore('about.feedback_describe_bug_placeholder')
                : $tStore('about.feedback_describe_feature_placeholder')}
            ></textarea>
          </label>

          <label class="feedback-field">
            <span class="feedback-label">
              {$feedbackOverlayKindStore === 'bug'
                ? $tStore('about.feedback_reproduce')
                : $tStore('about.feedback_use_case')}
            </span>
            <textarea
              rows="3"
              class="feedback-textarea"
              value={activeDraft.extra}
              oninput={(e) => updateDraftField('extra', e.currentTarget.value)}
              placeholder={$feedbackOverlayKindStore === 'bug'
                ? $tStore('about.feedback_reproduce_placeholder')
                : $tStore('about.feedback_use_case_placeholder')}></textarea>
          </label>

          {#if hasDuplicateContent}
            <div class="feedback-warning-pill" role="alert">
              <AlertTriangle size={14} strokeWidth={2} />
              <span>{$tStore('about.feedback_duplicate_warning')}</span>
            </div>
          {/if}

          <div class="feedback-meta-stack">
            <div class="feedback-info-pill">
              <ImagePlus size={14} strokeWidth={2} />
              <span>{$tStore('about.feedback_attachments_hint')}</span>
            </div>

            {#if $feedbackOverlayKindStore === 'bug'}
              <div class="feedback-info-pill">
                <Monitor size={14} strokeWidth={2} />
                <div class="feedback-info-copy">
                  <span>{$tStore('about.feedback_desktop_auto')}</span>
                  {#if diagnostics}
                    <span class="feedback-diagnostics-meta">
                      v{diagnostics.appVersion} · {diagnostics.osName} ({diagnostics.windowsBuild})
                    </span>
                  {/if}
                </div>
              </div>
            {/if}
          </div>
        </div>
      </div>
    </ScrollView>

    <div class="feedback-actions-bar">
      <div class="feedback-actions-left">
        <SettingActionButton
          variant="danger"
          disabled={!hasAnyContent}
          onclick={handleClear}
        >
          <RotateCcw size={12} strokeWidth={2} />
          <span>{$tStore('about.feedback_clear')}</span>
        </SettingActionButton>
      </div>
      <div class="feedback-actions-right">
        <SettingActionButton onclick={onClose}>
          <span>{$tStore('about.feedback_cancel')}</span>
        </SettingActionButton>
        <span class="feedback-submit-wrap">
          <SettingActionButton disabled={!canSubmit} onclick={handleSubmit}>
            <span>{$tStore('about.feedback_open_github')}</span>
            <ArrowUpRight size={14} strokeWidth={2} />
          </SettingActionButton>
        </span>
      </div>
    </div>
  </div>
</OverlayPage>

<style src="../styles/FeedbackOverlayPage.css"></style>
