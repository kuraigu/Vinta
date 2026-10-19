<script lang="ts">
  import { onMount } from 'svelte'
  import { RefreshCw, Info, X } from 'lucide-svelte'
  import {
    updateInfoStore,
    updateLifecycleStatusStore,
    triggerApplyUpdate
  } from '../store'

  let {
    onDismiss,
    onClick
  }: {
    onDismiss?: () => void
    onClick?: () => void
  } = $props()

  let isReady = $derived($updateLifecycleStatusStore === 'ready')

  function handleCardClick() {
    if (onClick) onClick()
  }

  function handleDismiss(e: MouseEvent) {
    e.stopPropagation()
    if (onDismiss) onDismiss()
  }

  function handleRestart(e: MouseEvent) {
    e.stopPropagation()
    triggerApplyUpdate()
  }
</script>

<!-- Shared Reusable Update Card -->
<section
  class="update-glance-card {isReady ? 'ready' : 'available'}"
  role="button"
  tabindex="0"
  onclick={handleCardClick}
  onkeydown={(e) => e.key === 'Enter' && handleCardClick()}
>
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
        Download paused. Click resume to continue from current offset.
      {:else}
        A new release is available for download and installation.
      {/if}
    </span>
  </div>

  <div class="update-glance-actions">
    {#if isReady}
      <button
        type="button"
        class="update-action-btn restart"
        onclick={handleRestart}
        title="Restart Vinta"
      >
        Restart
      </button>
    {/if}

    <button
      type="button"
      class="update-dismiss-btn"
      onclick={handleDismiss}
      title="Dismiss for this session"
      aria-label="Dismiss update banner"
    >
      <X size={12} />
    </button>
  </div>
</section>

<style>
  .update-glance-card {
    display: flex;
    align-items: center;
    text-align: left;
    gap: 12px;
    width: 100%;
    padding: 12px 14px;
    border-radius: 16px;
    box-sizing: border-box;
    cursor: pointer;
    user-select: none;
    -webkit-user-select: none;
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

  .update-glance-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.22);
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
    font-size: 13.5px;
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
    font-size: 11.5px;
    color: rgba(255, 255, 255, 0.86);
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

  .update-action-btn {
    font-size: 0.72rem;
    font-weight: 600;
    padding: 4px 10px;
    border-radius: 8px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    background: rgba(255, 255, 255, 0.14);
    color: #ffffff;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .update-action-btn:hover {
    background: rgba(255, 255, 255, 0.22);
    border-color: rgba(255, 255, 255, 0.32);
    transform: translateY(-0.5px);
  }

  .update-action-btn.restart {
    background: rgba(255, 255, 255, 0.25);
    border-color: rgba(255, 255, 255, 0.35);
    color: #ffffff;
  }

  .update-dismiss-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    border-radius: 6px;
    border: none;
    background: transparent;
    color: rgba(255, 255, 255, 0.75);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .update-dismiss-btn:hover {
    background: rgba(255, 255, 255, 0.18);
    color: #ffffff;
  }
</style>
