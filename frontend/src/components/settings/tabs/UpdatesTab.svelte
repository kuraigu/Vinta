<script lang="ts">
  import { onMount } from 'svelte'
  import {
    RefreshCw,
    Sparkles,
    Download,
    ExternalLink,
    FileText
  } from '@lucide/svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import {
    updateInfoStore,
    updateCheckingStore,
    autoCheckUpdatesStore,
    triggerUpdateCheck
  } from '../../../store'
  import { OpenURL, GetAppInfo } from '../../../../wailsjs/go/backend/App.js'

  let isDevBuild = $state(false)

  function handleCheckUpdates() {
    triggerUpdateCheck(true)
  }

  function handleOpenURL(url: string) {
    if (url && typeof OpenURL === 'function') {
      OpenURL(url).catch(() => {})
    }
  }

  function initUpdatesTab() {
    if (typeof GetAppInfo === 'function') {
      GetAppInfo()
        .then((info) => {
          if (info) {
            isDevBuild = info.buildType === 'Development'
          }
        })
        .catch(() => {})
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
      icon={Sparkles}
    >
      <label class="toggle-switch">
        <input type="checkbox" bind:checked={$autoCheckUpdatesStore} />
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
  {#if $updateInfoStore?.updateAvailable}
    <div class="update-alert-card">
      <div class="update-alert-header">
        <div class="update-alert-icon">
          <Sparkles size={16} />
        </div>
        <div class="update-alert-meta">
          <span class="update-alert-title"
            >Update Available: {$updateInfoStore.latestVersion}</span
          >
          <span class="update-alert-sub"
            >{$updateInfoStore.release?.name ||
              'A new release of Vinta is ready for installation.'}</span
          >
        </div>
      </div>
      <div class="update-alert-actions">
        {#if $updateInfoStore.release?.installerUrl}
          <SettingActionButton
            variant="default"
            onclick={() => handleOpenURL($updateInfoStore.release.installerUrl)}
          >
            <Download size={12} strokeWidth={2} />
            <span>Installer (.exe)</span>
          </SettingActionButton>
        {/if}
        {#if $updateInfoStore.release?.portableUrl}
          <SettingActionButton
            variant="default"
            onclick={() => handleOpenURL($updateInfoStore.release.portableUrl)}
          >
            <Download size={12} strokeWidth={2} />
            <span>Portable (.zip)</span>
          </SettingActionButton>
        {/if}
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
    </div>
  {/if}
</div>

<style src="../../styles/settings/UpdatesTab.css"></style>
