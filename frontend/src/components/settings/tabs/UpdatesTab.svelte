<script lang="ts">
  import { onMount } from 'svelte'
  import { RefreshCw, Download, ExternalLink, FileText } from 'lucide-svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import {
    updateInfoStore,
    updateCheckingStore,
    autoCheckUpdatesStore,
    triggerUpdateCheck
  } from '../../../store'
  import {
    OpenURL,
    GetAppInfo,
    DownloadAndInstallUpdate
  } from '../../../../bindings/Vinta/backend/app.js'

  let isDevBuild = $state(false)
  let isDownloadingUpdate = $state(false)
  let updateDownloadError = $state('')

  function handleCheckUpdates() {
    if (isDevBuild) return
    triggerUpdateCheck(true)
  }

  function handleOpenURL(url: string) {
    if (url && typeof OpenURL === 'function') {
      OpenURL(url).catch(() => {})
    }
  }

  async function handleDownloadAndInstall() {
    const installerUrl = $updateInfoStore?.release?.installerUrl
    if (!installerUrl) {
      handleOpenURL(
        $updateInfoStore?.release?.htmlUrl ||
          'https://github.com/kuraigu/Vinta/releases'
      )
      return
    }

    isDownloadingUpdate = true
    updateDownloadError = ''
    try {
      if (typeof DownloadAndInstallUpdate === 'function') {
        await DownloadAndInstallUpdate(installerUrl)
      } else {
        handleOpenURL(installerUrl)
      }
    } catch (err) {
      updateDownloadError = `Installation failed: ${err}`
      isDownloadingUpdate = false
    }
  }

  async function initUpdatesTab() {
    if (typeof GetAppInfo === 'function') {
      try {
        const info = await GetAppInfo()
        if (info) {
          isDevBuild = info.buildType === 'Development'
        }
      } catch {
        isDevBuild = false
      }
    }

    if (!$updateInfoStore && !isDevBuild) {
      triggerUpdateCheck(false)
    }
  }

  onMount(() => {
    initUpdatesTab()
  })
</script>

<div class="updates-container">
  <!-- Top Settings Rows -->
  <div class="setting-group">
    <SettingItem
      label="Check for Updates"
      description={isDevBuild
        ? 'Disabled in development mode'
        : $updateInfoStore?.status === 'error'
          ? 'Check failed'
          : $updateInfoStore?.checkedAt
            ? `Last checked: ${$updateInfoStore.checkedAt}`
            : 'Check for new versions'}
      icon={RefreshCw}
      disabled={isDevBuild}
    >
      <SettingActionButton
        variant="default"
        disabled={$updateCheckingStore || isDevBuild}
        onclick={handleCheckUpdates}
      >
        <RefreshCw
          size={12}
          strokeWidth={2}
          class={$updateCheckingStore ? 'spinning' : ''}
        />
        <span>{$updateCheckingStore ? 'Checking...' : 'Check Now'}</span>
      </SettingActionButton>
    </SettingItem>

    <SettingItem
      label="Automatically Check for Updates"
      description="Check for new releases in the background"
      icon={Download}
      disabled={isDevBuild}
    >
      <label class="toggle-switch">
        <input
          type="checkbox"
          bind:checked={$autoCheckUpdatesStore}
          disabled={isDevBuild}
        />
        <span class="slider"></span>
      </label>
    </SettingItem>

    <SettingItem
      label="Release Notes & Highlights"
      description="View release notes and downloads on GitHub"
      icon={FileText}
    >
      <SettingActionButton
        variant="default"
        onclick={() =>
          handleOpenURL('https://github.com/kuraigu/Vinta/releases')}
      >
        <span>All Releases</span>
        <ExternalLink size={11} strokeWidth={2} />
      </SettingActionButton>
    </SettingItem>
  </div>

  <!-- Update Available Banner -->
  {#if $updateInfoStore?.updateAvailable && !isDevBuild}
    <div class="update-alert-card">
      <div class="update-alert-header">
        <div class="update-alert-icon">
          <Download size={16} />
        </div>
        <div class="update-alert-meta">
          <span class="update-alert-title">Update Available</span>
          <span class="update-alert-sub"
            >Version {$updateInfoStore.latestVersion} is ready to download and install.</span
          >
        </div>
      </div>
      <div class="update-alert-actions">
        <SettingActionButton
          variant="default"
          disabled={isDownloadingUpdate}
          onclick={handleDownloadAndInstall}
        >
          {#if isDownloadingUpdate}
            <RefreshCw size={12} strokeWidth={2} class="spinning" />
            <span>Downloading...</span>
          {:else}
            <Download size={12} strokeWidth={2} />
            <span>Download</span>
          {/if}
        </SettingActionButton>

        {#if $updateInfoStore.release?.htmlUrl}
          <SettingActionButton
            variant="default"
            onclick={() => handleOpenURL($updateInfoStore.release.htmlUrl)}
          >
            <ExternalLink size={12} strokeWidth={2} />
            <span>GitHub</span>
          </SettingActionButton>
        {/if}
      </div>
      {#if updateDownloadError}
        <span class="update-error-text">{updateDownloadError}</span>
      {/if}
    </div>
  {/if}
</div>

<style src="../../styles/settings/UpdatesTab.css"></style>
