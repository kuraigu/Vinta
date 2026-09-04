<script lang="ts">
  import { getContext } from 'svelte'
  import { Bell } from '@lucide/svelte'
  import { notifData } from '../store'
  import type { IslandConfig } from '../store'
  import NotificationPage, {
    config as notifConfig
  } from '../pages/NotificationPage.svelte'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let prevNotifCount = $state(0)
  let isNewNotif = $state(false)
  let pulseTimeout: number

  $effect(() => {
    const currentCount = $notifData.length
    if (currentCount > prevNotifCount) {
      isNewNotif = true
      clearTimeout(pulseTimeout)
      pulseTimeout = window.setTimeout(() => {
        isNewNotif = false
      }, 3500)
    }
    prevNotifCount = currentCount
  })
</script>

{#if $notifData.length > 0}
  <button
    class="mini-notif-indicator {isNewNotif ? 'pulse-new' : ''}"
    title="{$notifData.length} unread notification(s)"
    onclick={(e) => {
      e.stopPropagation()
      if (navigate) navigate(NotificationPage, notifConfig, true)
    }}
  >
    <Bell size={10} />
    <span class="mini-notif-count"
      >{$notifData.length > 9 ? '9+' : $notifData.length}</span
    >
  </button>
{/if}

<style>
  .mini-notif-indicator {
    position: relative;
    background: rgba(99, 102, 241, 0.25);
    border: 1px solid rgba(99, 102, 241, 0.45);
    color: #a5b4fc;
    padding: 1px 5px;
    height: 18px;
    border-radius: 9px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 3px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .mini-notif-indicator:hover {
    background: rgba(99, 102, 241, 0.45);
    transform: scale(1.08);
  }

  .mini-notif-indicator.pulse-new {
    animation: notifPing 0.7s infinite alternate ease-in-out;
    background: rgba(239, 68, 68, 0.4);
    border-color: #ef4444;
    color: #ffffff;
  }

  @keyframes notifPing {
    0% {
      transform: scale(1);
      box-shadow: 0 0 4px #ef4444;
    }
    100% {
      transform: scale(1.12);
      box-shadow: 0 0 10px #ef4444;
    }
  }

  .mini-notif-count {
    font-size: 8px;
    font-weight: 700;
    line-height: 1;
  }
</style>
