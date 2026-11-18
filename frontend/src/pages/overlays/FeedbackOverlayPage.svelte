<!--
  @fileoverview Overlay page for submitting pre-filled Bug Reports and Feature Requests
  with automatic sanitized desktop diagnostics.
-->
<script lang="ts">
  import { onMount } from 'svelte'
  import {
    ArrowUpRight,
    Bug,
    ImagePlus,
    Lightbulb,
    Monitor
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
  import { feedbackOverlayKindStore, tStore } from '../../store'

  interface Props {
    onClose: () => void
  }

  let { onClose }: Props = $props()

  let title = $state('')
  let describe = $state('')
  let extra = $state('')
  let diagnostics = $state<SystemDiagnostics | null>(null)

  let headerTabs = $derived<TabItem[]>([
    { id: 'bug', label: $tStore('about.feedback_type_bug'), icon: Bug },
    {
      id: 'feature',
      label: $tStore('about.feedback_type_feature'),
      icon: Lightbulb
    }
  ])

  let canSubmit = $derived(
    title.trim().length > 0 &&
      describe.trim().length > 0 &&
      ($feedbackOverlayKindStore === 'feature' || extra.trim().length > 0)
  )

  function handleSubmit() {
    if (!canSubmit) return
    if ($feedbackOverlayKindStore === 'bug') {
      reportFeedback(title, describe, extra).catch(() => {})
    } else {
      requestFeature(title, describe, extra).catch(() => {})
    }
    onClose()
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
              bind:value={title}
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
              bind:value={describe}
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
              {#if $feedbackOverlayKindStore === 'bug'}
                <span class="feedback-required" aria-hidden="true">*</span>
              {/if}
            </span>
            <textarea
              rows="3"
              class="feedback-textarea"
              bind:value={extra}
              placeholder={$feedbackOverlayKindStore === 'bug'
                ? $tStore('about.feedback_reproduce_placeholder')
                : $tStore('about.feedback_use_case_placeholder')}></textarea>
          </label>

          <div class="feedback-meta-stack">
            <div class="feedback-info-pill">
              <ImagePlus size={13} strokeWidth={2} />
              <span>{$tStore('about.feedback_attachments_hint')}</span>
            </div>

            {#if $feedbackOverlayKindStore === 'bug'}
              <div class="feedback-info-pill">
                <Monitor size={13} strokeWidth={2} />
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
      <SettingActionButton onclick={onClose}>
        <span>{$tStore('about.feedback_cancel')}</span>
      </SettingActionButton>
      <span class="feedback-submit-wrap">
        <SettingActionButton disabled={!canSubmit} onclick={handleSubmit}>
          <span>{$tStore('about.feedback_open_github')}</span>
          <ArrowUpRight size={13} strokeWidth={2} />
        </SettingActionButton>
      </span>
    </div>
  </div>
</OverlayPage>

<style src="../styles/FeedbackOverlayPage.css"></style>
