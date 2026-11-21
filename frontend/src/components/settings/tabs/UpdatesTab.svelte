<!--
  @component UpdatesTab
  @description App update manager. Checks for new releases, displays download progress,
  shows release changelogs, and triggers restart to apply pending updates.
-->
<script lang="ts">
  import { onMount } from 'svelte'
  import {
    RefreshCw,
    Download,
    ArrowUpRight,
    Pause,
    Play,
    X,
    CheckCircle2,
    ShieldCheck,
    AlertCircle,
    GitBranch,
    SlidersHorizontal,
    Info
  } from 'lucide-svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingBanner from '../SettingBanner.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import BentoCard from '../../BentoCard.svelte'
  import ToggleSwitch from '../../ToggleSwitch.svelte'
  import CirclePattern from '../../patterns/CirclePattern.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import ReleaseNotesGlance from '../../ReleaseNotesGlance.svelte'
  import {
    updateInfoStore,
    updateCheckingStore,
    autoCheckUpdatesStore,
    updateChannelStore,
    downloadProgressStore,
    updateLifecycleStatusStore,
    triggerUpdateCheck,
    triggerDownload,
    triggerPauseDownload,
    triggerResumeDownload,
    triggerCancelDownload,
    triggerApplyUpdate,
    triggerDevSimulation,
    triggerDevDownloadCycle,
    tStore
  } from '../../../store'
  import {
    OpenURL,
    GetAppInfo
  } from '../../../../bindings/Vinta/backend/app.js'
  import { Events } from '$lib/events'

  let isDevBuild = $state(false)
  let isOfficialBuild = $state(true)
  let showDevSimulation = $state(false)

  function formatBytes(bytes: number): string {
    if (bytes <= 0) return '0 B'
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  function computeChannelOptions() {
    if (isDevBuild) {
      return [
        { value: 'stable', label: $tStore('settings.updates_channel_stable') },
        {
          value: 'prerelease',
          label: $tStore('settings.updates_channel_prerelease')
        },
        {
          value: 'testing',
          label: $tStore('settings.updates_channel_testing')
        }
      ]
    }
    return [
      { value: 'stable', label: $tStore('settings.updates_channel_stable') },
      {
        value: 'prerelease',
        label: $tStore('settings.updates_channel_prerelease')
      }
    ]
  }

  let channelOptions = $derived.by(computeChannelOptions)

  function enforceValidChannel() {
    if (!isDevBuild && $updateChannelStore === 'testing') {
      $updateChannelStore = 'stable'
    }
  }

  $effect(enforceValidChannel)

  function handleCheckUpdates() {
    if (!isOfficialBuild) return
    triggerUpdateCheck(true)
  }

  function handleOpenURL(url: string) {
    if (url && typeof OpenURL === 'function') {
      OpenURL(url).catch(() => {})
    }
  }

  function handleChannelChange(newChannel: string) {
    $updateChannelStore = newChannel as 'stable' | 'prerelease' | 'testing'
    if (isOfficialBuild) {
      triggerUpdateCheck(true)
    }
  }

  function handleStartDownload() {
    const installerUrl = $updateInfoStore?.release?.installerUrl
    const checksumsUrl = $updateInfoStore?.release?.checksumsUrl || ''
    if (!installerUrl) {
      handleOpenURL(
        $updateInfoStore?.release?.htmlUrl ||
          'https://github.com/kuraigu/Vinta/releases'
      )
      return
    }
    triggerDownload(installerUrl, checksumsUrl)
  }

  function handleProgressUpdate(ev: { data: typeof $downloadProgressStore }) {
    if (ev?.data) {
      const data = ev.data
      downloadProgressStore.set(data)
      if (data.status === 'downloading') {
        updateLifecycleStatusStore.set('downloading')
      } else if (data.status === 'paused') {
        updateLifecycleStatusStore.set('paused')
      } else if (data.status === 'verifying') {
        updateLifecycleStatusStore.set('verifying')
      } else if (data.status === 'ready') {
        updateLifecycleStatusStore.set('ready')
      } else if (data.status === 'error') {
        updateLifecycleStatusStore.set('error')
      }
    }
  }

  async function initUpdatesTab() {
    if (typeof GetAppInfo === 'function') {
      try {
        const info = await GetAppInfo()
        if (info) {
          isDevBuild = info.buildType === 'Development'
          isOfficialBuild = Boolean(info.isOfficialBuild)
        }
      } catch {
        isDevBuild = false
        isOfficialBuild = false
      }
    }

    if (isOfficialBuild && !$updateInfoStore) {
      triggerUpdateCheck(false)
    }
  }

  Events.On('update:download-progress', handleProgressUpdate)

  onMount(() => {
    initUpdatesTab()
  })
</script>

<div class="updates-container">
  <!-- Non-Official Release Information Banner: placed at the very top of the Updates page -->
  {#if !isOfficialBuild}
    <SettingBanner
      variant="amber"
      icon={AlertCircle}
      title={$tStore('settings.updates_nonofficial_title')}
      description={$tStore('settings.updates_nonofficial_desc')}
    />
  {/if}

  <!-- Settings Group -->
  <div class="setting-group">
    <SettingItem
      label={$tStore('settings.updates_check_label')}
      description={!isOfficialBuild
        ? $tStore('settings.updates_check_disabled_desc')
        : $updateInfoStore?.status === 'error'
          ? $updateInfoStore.message || $tStore('settings.updates_check_failed')
          : $updateInfoStore?.checkedAt
            ? $tStore('settings.updates_last_checked', {
                time: $updateInfoStore.checkedAt
              })
            : $tStore('settings.updates_check_new_desc')}
      icon={RefreshCw}
      flexCol={true}
      disabled={!isOfficialBuild}
    >
      <SettingActionButton
        variant="default"
        fullWidth={true}
        disabled={!isOfficialBuild || $updateCheckingStore}
        onclick={handleCheckUpdates}
      >
        <RefreshCw
          size={13}
          strokeWidth={2}
          class={$updateCheckingStore ? 'spinning' : ''}
        />
        <span
          >{$updateCheckingStore
            ? $tStore('settings.updates_checking_btn')
            : $tStore('settings.updates_check_btn')}</span
        >
      </SettingActionButton>
    </SettingItem>

    <!-- Update Available & Download Card: directly below Check for Updates -->
    {#if $updateInfoStore?.updateAvailable}
      <BentoCard
        variant="announcement"
        color={$updateLifecycleStatusStore === 'ready' ? 'emerald' : 'blue'}
        flexCol={true}
        class="update-card"
      >
        <div class="update-card-pattern-wrap" aria-hidden="true">
          <CirclePattern align="right" rows={10} cols={20} />
        </div>

        <div class="update-header">
          <div class="update-badge-icon">
            {#if $updateLifecycleStatusStore === 'ready'}
              <CheckCircle2 size={16} />
            {:else if $updateLifecycleStatusStore === 'verifying'}
              <ShieldCheck size={16} class="spinning" />
            {:else}
              <Info size={16} />
            {/if}
          </div>
          <div class="update-header-info">
            <div class="update-title-row">
              <span class="update-title"
                >Vinta {$updateInfoStore.latestVersion}</span
              >
            </div>
            <span class="update-sub">
              {#if $updateLifecycleStatusStore === 'ready'}
                {$tStore('settings.updates_ready_desc')}
              {:else if $updateLifecycleStatusStore === 'verifying'}
                {$tStore('settings.updates_verifying_desc')}
              {:else if $updateLifecycleStatusStore === 'downloading'}
                {$tStore('settings.updates_downloading_desc')}
              {:else if $updateLifecycleStatusStore === 'paused'}
                {$tStore('settings.updates_paused_desc')}
              {:else}
                {$tStore('settings.updates_available_desc')}
              {/if}
            </span>
          </div>
        </div>

        <!-- Progress Bar & Download Stats -->
        {#if $updateLifecycleStatusStore === 'downloading' || $updateLifecycleStatusStore === 'paused' || $updateLifecycleStatusStore === 'verifying'}
          <div class="progress-section">
            <div class="progress-bar-track">
              <div
                class="progress-bar-fill {$updateLifecycleStatusStore ===
                'paused'
                  ? 'paused'
                  : ''}"
                style="transform: scaleX({Math.min(
                  1,
                  Math.max(0, $downloadProgressStore.percentage / 100)
                )});"
              ></div>
            </div>
            <div class="progress-meta-row">
              <span class="progress-text">
                {#if $updateLifecycleStatusStore === 'verifying'}
                  {$tStore('settings.updates_validating_checksum')}
                {:else}
                  {formatBytes($downloadProgressStore.downloadedBytes)} /
                  {$downloadProgressStore.totalBytes > 0
                    ? formatBytes($downloadProgressStore.totalBytes)
                    : '...'}
                  {#if $updateLifecycleStatusStore === 'downloading' && $downloadProgressStore.speedBytesPerSec > 0}
                    ({formatBytes($downloadProgressStore.speedBytesPerSec)}/s)
                  {/if}
                {/if}
              </span>
              <span class="progress-percent">
                {Math.round($downloadProgressStore.percentage)}%
              </span>
            </div>
          </div>
        {/if}

        <!-- Action Buttons Row -->
        <div class="update-actions-row">
          {#if $updateLifecycleStatusStore === 'ready'}
            <SettingActionButton variant="default" onclick={triggerApplyUpdate}>
              <CheckCircle2 size={12} strokeWidth={2} />
              <span>{$tStore('settings.updates_restart_install')}</span>
            </SettingActionButton>
            <SettingActionButton
              variant="default"
              onclick={() => updateLifecycleStatusStore.set('available')}
            >
              <span>{$tStore('settings.updates_install_on_exit')}</span>
            </SettingActionButton>
          {:else if $updateLifecycleStatusStore === 'downloading'}
            <SettingActionButton
              variant="default"
              onclick={triggerPauseDownload}
            >
              <Pause size={12} strokeWidth={2} />
              <span>{$tStore('settings.updates_pause')}</span>
            </SettingActionButton>
            <SettingActionButton
              variant="default"
              onclick={triggerCancelDownload}
            >
              <X size={12} strokeWidth={2} />
              <span>{$tStore('settings.updates_cancel')}</span>
            </SettingActionButton>
          {:else if $updateLifecycleStatusStore === 'paused'}
            <SettingActionButton
              variant="default"
              onclick={triggerResumeDownload}
            >
              <Play size={12} strokeWidth={2} />
              <span>{$tStore('settings.updates_resume')}</span>
            </SettingActionButton>
            <SettingActionButton
              variant="default"
              onclick={triggerCancelDownload}
            >
              <X size={12} strokeWidth={2} />
              <span>{$tStore('settings.updates_cancel')}</span>
            </SettingActionButton>
          {:else if $updateLifecycleStatusStore === 'verifying'}
            <SettingActionButton variant="default" disabled={true}>
              <RefreshCw size={12} strokeWidth={2} class="spinning" />
              <span>{$tStore('settings.updates_verifying_btn')}</span>
            </SettingActionButton>
          {:else}
            <SettingActionButton
              variant="default"
              onclick={handleStartDownload}
            >
              <Download size={12} strokeWidth={2} />
              <span>{$tStore('settings.updates_download_install')}</span>
            </SettingActionButton>
          {/if}

          {#if $updateInfoStore.release?.htmlUrl}
            <SettingActionButton
              variant="default"
              onclick={() => handleOpenURL($updateInfoStore.release.htmlUrl)}
            >
              <ArrowUpRight size={13} strokeWidth={2} />
              <span>{$tStore('settings.updates_github')}</span>
            </SettingActionButton>
          {/if}
        </div>

        {#if $downloadProgressStore.error || ($updateInfoStore.status === 'error' && $updateInfoStore.message)}
          <div class="update-error-banner">
            <AlertCircle size={14} />
            <span
              >{$downloadProgressStore.error || $updateInfoStore.message}</span
            >
          </div>
        {/if}
      </BentoCard>
    {/if}

    <!-- Current Installed Version Release Notes -->
    <ReleaseNotesGlance />

    <BentoCard variant="group">
      <SettingItem
        label={$tStore('settings.updates_channel_label')}
        description={$tStore('settings.updates_channel_desc')}
        icon={GitBranch}
        flexCol={true}
        isDefault={$updateChannelStore === updateChannelStore.defaultValue}
        onReset={() => handleChannelChange(updateChannelStore.defaultValue)}
      >
        <CustomSelect
          options={channelOptions}
          value={$updateChannelStore}
          fullWidth={true}
          onChange={handleChannelChange}
        />
      </SettingItem>

      <SettingItem
        label={$tStore('settings.updates_auto_check_label')}
        description={!isOfficialBuild
          ? $tStore('settings.updates_check_disabled_desc')
          : $tStore('settings.updates_auto_check_desc')}
        icon={Download}
        disabled={!isOfficialBuild}
        isDefault={$autoCheckUpdatesStore ===
          autoCheckUpdatesStore.defaultValue}
        onReset={() => autoCheckUpdatesStore.reset()}
      >
        <ToggleSwitch
          disabled={!isOfficialBuild}
          bind:checked={$autoCheckUpdatesStore}
          ariaLabel={$tStore('settings.updates_auto_check_label')}
        />
      </SettingItem>
    </BentoCard>
  </div>

  <!-- Local Dev Simulation Controls (Strictly visible only during dev builds) -->
  {#if isDevBuild}
    <div class="setting-group">
      <SettingItem
        label={$tStore('settings.updates_dev_title')}
        description={$tStore('settings.updates_dev_desc')}
        icon={SlidersHorizontal}
        flexCol={true}
      >
        <SettingActionButton
          variant="default"
          fullWidth={true}
          onclick={() => (showDevSimulation = !showDevSimulation)}
        >
          <span
            >{showDevSimulation
              ? $tStore('settings.updates_hide_controls')
              : $tStore('settings.updates_show_controls')}</span
          >
        </SettingActionButton>
      </SettingItem>

      {#if showDevSimulation}
        <div class="devtest-sim-card">
          <div class="devtest-sim-buttons">
            <SettingActionButton
              variant="default"
              fullWidth={true}
              onclick={() => triggerDevSimulation('v0.4.0-alpha')}
            >
              <span>{$tStore('settings.updates_sim_available')}</span>
            </SettingActionButton>
            <SettingActionButton
              variant="default"
              fullWidth={true}
              onclick={() => triggerDevDownloadCycle(false)}
            >
              <span>{$tStore('settings.updates_sim_success')}</span>
            </SettingActionButton>
            <SettingActionButton
              variant="default"
              fullWidth={true}
              onclick={() => triggerDevDownloadCycle(true)}
            >
              <span>{$tStore('settings.updates_sim_fail')}</span>
            </SettingActionButton>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style src="../../styles/settings/UpdatesTab.css"></style>
