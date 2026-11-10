<!--
  @component UpdateGlanceCard
  @description Prompt card alerting the user when a new Vinta build is ready to install.
-->
<script lang="ts">
  import { RefreshCw, Info, ChevronRight } from 'lucide-svelte'
  import {
    updateInfoStore,
    updateLifecycleStatusStore,
    triggerApplyUpdate
  } from '../store'
  import CirclePattern from './patterns/CirclePattern.svelte'
  import ActionCapsule from './ActionCapsule.svelte'

  let {
    onClick
  }: {
    onClick?: () => void
  } = $props()

  let isReady = $derived($updateLifecycleStatusStore === 'ready')

  function handleCardClick() {
    if (onClick) onClick()
  }

  function handleRestart(e: MouseEvent) {
    e.stopPropagation()
    triggerApplyUpdate()
  }
</script>

<!-- Shared Reusable Update Card with Atmospheric Bento Backdrop -->
<section
  class="update-glance-card {isReady ? 'ready' : 'available'}"
  role="button"
  tabindex="0"
  onclick={handleCardClick}
  onkeydown={(e) => e.key === 'Enter' && handleCardClick()}
>
  <div class="update-glance-pattern-wrap" aria-hidden="true">
    <CirclePattern align="right" rows={8} cols={18} />
  </div>

  <div class="update-glance-body">
    <div class="update-glance-icon">
      {#if isReady}
        <RefreshCw size={14} class="update-spin-icon" />
      {:else}
        <Info size={14} />
      {/if}
    </div>

    <div class="update-glance-text">
      <span class="update-glance-title">
        {isReady
          ? 'Update Ready to Install'
          : `Vinta ${$updateInfoStore?.latestVersion || 'Update'}`}
      </span>
      <span class="update-glance-sub">
        {#if isReady}
          Update package downloaded and verified against SHA-256 manifest.
        {:else if $updateLifecycleStatusStore === 'verifying'}
          Verifying SHA-256 package checksum against release manifest...
        {:else if $updateLifecycleStatusStore === 'downloading'}
          Downloading update package...
        {:else if $updateLifecycleStatusStore === 'paused'}
          Download paused. Click to view status in settings.
        {:else}
          A new release is available for download and installation.
        {/if}
      </span>
    </div>

    <div class="update-glance-actions">
      {#if isReady}
        <ActionCapsule
          label="Restart"
          variant="frosted"
          onclick={handleRestart}
        />
      {:else}
        <ChevronRight size={16} class="update-glance-arrow" />
      {/if}
    </div>
  </div>
</section>

<style src="./styles/UpdateGlanceCard.css"></style>
