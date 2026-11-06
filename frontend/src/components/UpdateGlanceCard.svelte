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
          tooltip="Restart Vinta to complete update"
        />
      {:else}
        <ChevronRight size={16} class="update-glance-arrow" />
      {/if}
    </div>
  </div>
</section>

<style>
  .update-glance-card {
    position: relative;
    width: 100%;
    border-radius: var(--vinta-radius-card, 16px);
    box-sizing: border-box;
    cursor: pointer;
    user-select: none;
    -webkit-user-select: none;
    overflow: hidden;
    background: linear-gradient(135deg, #0284c7 0%, #0369a1 50%, #1d4ed8 100%);
    border: 1px solid rgba(255, 255, 255, 0.25);
    box-shadow:
      inset 0 1px 0 0 rgba(255, 255, 255, 0.28),
      0 4px 16px rgba(2, 132, 199, 0.28);
    color: #ffffff;
    transition:
      transform 0.2s cubic-bezier(0.2, 0.8, 0.2, 1),
      border-color 0.2s ease,
      box-shadow 0.2s ease;
  }

  .update-glance-card:hover {
    transform: translateY(-1px);
    border-color: rgba(255, 255, 255, 0.38);
    box-shadow:
      inset 0 1px 0 0 rgba(255, 255, 255, 0.35),
      0 6px 20px rgba(2, 132, 199, 0.35);
  }

  .update-glance-card.ready {
    background: linear-gradient(135deg, #059669 0%, #047857 50%, #065f46 100%);
    border-color: rgba(255, 255, 255, 0.3);
    box-shadow:
      inset 0 1px 0 0 rgba(255, 255, 255, 0.28),
      0 4px 16px rgba(5, 150, 105, 0.28);
  }

  .update-glance-pattern-wrap {
    position: absolute;
    inset: 0;
    pointer-events: none;
    opacity: 0.32;
    z-index: 0;
  }

  .update-glance-body {
    position: relative;
    z-index: 1;
    display: flex;
    align-items: center;
    text-align: left;
    gap: 12px;
    width: 100%;
    padding: 11px 14px;
    box-sizing: border-box;
  }

  .update-glance-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 30px;
    height: 30px;
    border-radius: 9px;
    background: rgba(255, 255, 255, 0.2);
    color: #ffffff;
    flex-shrink: 0;
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.25);
  }

  :global(.update-spin-icon) {
    animation: spinSlow 3s linear infinite;
  }

  @keyframes spinSlow {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }

  .update-glance-text {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    text-align: left;
    gap: 2px;
    min-width: 0;
    flex: 1;
  }

  .update-glance-title {
    display: block;
    font-size: 13px;
    font-weight: 700;
    letter-spacing: -0.01em;
    color: #ffffff;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    text-align: left;
  }

  .update-glance-sub {
    display: block;
    font-size: 11px;
    color: rgba(255, 255, 255, 0.88);
    line-height: 1.35;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    text-align: left;
  }

  .update-glance-actions {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
  }

  :global(.update-glance-arrow) {
    color: rgba(255, 255, 255, 0.7);
    transition:
      transform 0.15s ease,
      color 0.15s ease;
    flex-shrink: 0;
  }

  .update-glance-card:hover :global(.update-glance-arrow) {
    color: #ffffff;
    transform: translateX(2px);
  }
</style>
