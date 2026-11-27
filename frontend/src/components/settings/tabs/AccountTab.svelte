<!--
  @component AccountTab
  @description Settings tab hosting Google Sign-In + Cloud Sync,
  GitHub Sign-In, and Windows DPAPI-encrypted AI provider keys.
-->
<script lang="ts">
  import { onMount } from 'svelte'
  import {
    Cloud,
    CloudDownload,
    CloudUpload,
    KeyRound,
    LogOut,
    Terminal
  } from 'lucide-svelte'
  import SettingSectionHeader from '../SettingSectionHeader.svelte'
  import SettingItem from '../SettingItem.svelte'
  import BentoCard from '../../BentoCard.svelte'
  import BrandLogo from '../../BrandLogo.svelte'
  import ToggleSwitch from '../../ToggleSwitch.svelte'
  import { tStore } from '../../../store'
  import {
    googleAccountStore,
    githubStatusStore,
    aiKeysStatusStore,
    initDevIntegrations,
    signInGoogleAccount,
    signOutGoogleAccount,
    pushCloudBackup,
    pullCloudBackup,
    startGitHubDeviceAuth,
    completeGitHubDeviceAuth,
    disconnectGitHubAccount,
    saveAIKey,
    deleteAIKey
  } from '../../../stores/devIntegrationsStore'

  let isGoogleBusy = $state(false)
  let isSyncBusy = $state(false)
  let googleFeedback = $state('')
  let googleError = $state('')

  let isGitHubBusy = $state(false)
  let githubUserCode = $state('')
  let githubError = $state('')
  let includeGitHubNotifications = $state(false)
  let includeGitHubPrivateRepos = $state(false)

  let claudeKeyInput = $state('')
  let openaiKeyInput = $state('')
  let aiKeyFeedback = $state('')

  onMount(() => {
    void initDevIntegrations()
  })

  function formatSyncTime(unixSec: number): string {
    if (!unixSec) return $tStore('account.never_synced')
    return new Date(unixSec * 1000).toLocaleString()
  }

  async function handleGoogleSignIn() {
    isGoogleBusy = true
    googleError = ''
    googleFeedback = ''
    const res = await signInGoogleAccount()
    isGoogleBusy = false
    if (!res.ok && res.error) {
      googleError = res.error
    }
  }

  async function handlePushSync() {
    isSyncBusy = true
    googleError = ''
    googleFeedback = ''
    const res = await pushCloudBackup()
    isSyncBusy = false
    if (res.ok) {
      googleFeedback = $tStore('account.sync_pushed')
    } else if (res.error) {
      googleError = res.error
    }
  }

  async function handlePullSync() {
    isSyncBusy = true
    googleError = ''
    googleFeedback = ''
    const res = await pullCloudBackup()
    isSyncBusy = false
    if (res.ok) {
      googleFeedback = $tStore('account.sync_pulled')
    } else if (res.error) {
      googleError = res.error
    }
  }

  async function handleGitHubDeviceSignIn() {
    isGitHubBusy = true
    githubError = ''
    githubUserCode = ''
    const start = await startGitHubDeviceAuth({
      includeNotifications: includeGitHubNotifications,
      includePrivateRepos: includeGitHubPrivateRepos
    })
    if (!start.ok || !start.info) {
      isGitHubBusy = false
      githubError = start.error || $tStore('account.github_failed')
      return
    }
    githubUserCode = start.info.userCode
    const done = await completeGitHubDeviceAuth(
      start.info.deviceCode,
      start.info.interval
    )
    isGitHubBusy = false
    githubUserCode = ''
    if (!done.ok && done.error) {
      githubError = done.error
    }
  }

  async function handleSaveAIKey(provider: 'claude' | 'openai', value: string) {
    if (!value.trim()) return
    aiKeyFeedback = ''
    const res = await saveAIKey(provider, value.trim())
    if (res.ok) {
      if (provider === 'claude') claudeKeyInput = ''
      if (provider === 'openai') openaiKeyInput = ''
      aiKeyFeedback = $tStore('account.key_saved')
    } else if (res.error) {
      aiKeyFeedback = res.error
    }
  }
</script>

<div class="setting-group" data-vinta="account-tab">
  <SettingSectionHeader title={$tStore('account.google_section')} />

  <BentoCard variant="group">
    {#if $googleAccountStore.connected}
      <SettingItem
        label={$googleAccountStore.name || $googleAccountStore.email}
        description="{$googleAccountStore.email} · {formatSyncTime(
          $googleAccountStore.lastCloudSyncAt
        )}"
        icon={Cloud}
      >
        <button
          type="button"
          class="account-ghost-btn danger"
          onclick={() => void signOutGoogleAccount()}
        >
          <LogOut size={14} />
          <span>{$tStore('account.sign_out')}</span>
        </button>
      </SettingItem>

      {#if $googleAccountStore.cloudSyncReady}
        <SettingItem
          label={$tStore('account.cloud_backup_label')}
          description={$tStore('account.cloud_backup_desc')}
          flexCol={true}
        >
          <div class="account-actions-row">
            <button
              type="button"
              class="account-action-btn primary"
              disabled={isSyncBusy}
              onclick={handlePushSync}
            >
              <CloudUpload size={14} />
              <span>{$tStore('account.sync_now')}</span>
            </button>
            <button
              type="button"
              class="account-action-btn"
              disabled={isSyncBusy}
              onclick={handlePullSync}
            >
              <CloudDownload size={14} />
              <span>{$tStore('account.restore_cloud')}</span>
            </button>
          </div>
        </SettingItem>
      {/if}
    {:else}
      <SettingItem
        label={$tStore('account.google_title')}
        description={$tStore('account.google_desc')}
        icon={Cloud}
        flexCol={true}
      >
        <button
          type="button"
          class="oauth-brand-btn google"
          disabled={isGoogleBusy}
          onclick={handleGoogleSignIn}
        >
          <BrandLogo name="google" size={16} />
          <span>
            {isGoogleBusy
              ? $tStore('account.signing_in')
              : $tStore('account.sign_in_google')}
          </span>
        </button>
      </SettingItem>
    {/if}

    {#if googleFeedback}
      <p class="account-feedback">{googleFeedback}</p>
    {/if}
    {#if googleError}
      <p class="account-error" role="alert">{googleError}</p>
    {/if}
  </BentoCard>

  <SettingSectionHeader title={$tStore('account.github_section')} />

  <BentoCard variant="group">
    {#if $githubStatusStore.connected}
      <SettingItem
        label="@{$githubStatusStore.username}"
        description={$tStore('account.github_connected_sub')}
        icon={Terminal}
      >
        <button
          type="button"
          class="account-ghost-btn danger"
          onclick={() => void disconnectGitHubAccount()}
        >
          <LogOut size={14} />
          <span>{$tStore('account.disconnect')}</span>
        </button>
      </SettingItem>
    {:else}
      <SettingItem
        label={$tStore('account.github_title')}
        description={$tStore('account.github_desc')}
        icon={Terminal}
        flexCol={true}
      >
        <button
          type="button"
          class="oauth-brand-btn github"
          disabled={isGitHubBusy}
          onclick={handleGitHubDeviceSignIn}
        >
          <BrandLogo name="github" size={16} />
          <span>
            {isGitHubBusy
              ? $tStore('account.waiting_github')
              : $tStore('account.sign_in_github')}
          </span>
        </button>

        {#if githubUserCode}
          <div class="device-code-box">
            <span class="device-code-label"
              >{$tStore('account.device_code_label')}</span
            >
            <code class="device-code-value">{githubUserCode}</code>
          </div>
        {/if}
      </SettingItem>

      <SettingItem
        label={$tStore('account.github_opt_notifications_label')}
        description={$tStore('account.github_opt_notifications_desc')}
      >
        <ToggleSwitch
          bind:checked={includeGitHubNotifications}
          ariaLabel={$tStore('account.github_opt_notifications_label')}
        />
      </SettingItem>

      <SettingItem
        label={$tStore('account.github_opt_private_label')}
        description={$tStore('account.github_opt_private_desc')}
      >
        <ToggleSwitch
          bind:checked={includeGitHubPrivateRepos}
          ariaLabel={$tStore('account.github_opt_private_label')}
        />
      </SettingItem>
    {/if}

    {#if githubError}
      <p class="account-error" role="alert">{githubError}</p>
    {/if}
  </BentoCard>

  <SettingSectionHeader title={$tStore('account.ai_keys_section')} />
  <p class="account-section-note">{$tStore('account.ai_keys_desc')}</p>

  <BentoCard variant="group">
    <SettingItem
      label="Claude (Anthropic)"
      description={$aiKeysStatusStore.hasClaude
        ? $tStore('account.stored_dpapi')
        : $tStore('account.not_set')}
      icon={KeyRound}
      flexCol={true}
    >
      <div class="account-inline-row">
        <input
          id="ai-key-claude"
          type="password"
          autocomplete="off"
          spellcheck={false}
          class="account-input"
          placeholder="sk-ant-admin-... or sk-ant-..."
          aria-label="Claude API Key"
          bind:value={claudeKeyInput}
        />
        <button
          type="button"
          class="account-action-btn"
          disabled={!claudeKeyInput.trim()}
          onclick={() => void handleSaveAIKey('claude', claudeKeyInput)}
        >
          {$tStore('common.save')}
        </button>
        {#if $aiKeysStatusStore.hasClaude}
          <button
            type="button"
            class="account-ghost-btn danger"
            onclick={() => void deleteAIKey('claude')}
          >
            {$tStore('common.delete')}
          </button>
        {/if}
      </div>
    </SettingItem>

    <SettingItem
      label="OpenAI"
      description={$aiKeysStatusStore.hasOpenAI
        ? $tStore('account.stored_dpapi')
        : $tStore('account.not_set')}
      icon={KeyRound}
      flexCol={true}
    >
      <div class="account-inline-row">
        <input
          id="ai-key-openai"
          type="password"
          autocomplete="off"
          spellcheck={false}
          class="account-input"
          placeholder="sk-admin-... or sk-..."
          aria-label="OpenAI API Key"
          bind:value={openaiKeyInput}
        />
        <button
          type="button"
          class="account-action-btn"
          disabled={!openaiKeyInput.trim()}
          onclick={() => void handleSaveAIKey('openai', openaiKeyInput)}
        >
          {$tStore('common.save')}
        </button>
        {#if $aiKeysStatusStore.hasOpenAI}
          <button
            type="button"
            class="account-ghost-btn danger"
            onclick={() => void deleteAIKey('openai')}
          >
            {$tStore('common.delete')}
          </button>
        {/if}
      </div>
    </SettingItem>

    {#if aiKeyFeedback}
      <p class="account-feedback">{aiKeyFeedback}</p>
    {/if}
  </BentoCard>
</div>

<style src="../../styles/settings/AccountTab.css"></style>
