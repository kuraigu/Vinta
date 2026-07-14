<script lang="ts">
  import { flip } from 'svelte/animate'
  import { fly, fade } from 'svelte/transition'
  import { Bell, Trash2, X } from 'lucide-svelte'
  import { notifData } from '../store'
  import { onMount } from 'svelte'

  let { expanded = false, notifications = [] } = $props<{
    expanded?: boolean
    notifications?: Array<{
      id: number
      appName?: string
      icon?: string
      title: string
      message: string
      timestamp?: number
    }>
  }>()

  let latest = $derived(
    notifications[0] ||
      ({
        id: 0,
        title: 'No notifications',
        message: ''
      } as (typeof notifications)[0])
  )

  let now = $state(Date.now())

  onMount(() => {
    const interval = setInterval(() => {
      now = Date.now()
    }, 1000)
    return () => clearInterval(interval)
  })

  function getRelativeTime(timestamp?: number): string {
    if (!timestamp) return 'NOW'
    const diffSeconds = Math.max(0, Math.floor((now - timestamp) / 1000))
    if (diffSeconds < 5) return 'NOW'
    if (diffSeconds < 60) return `${diffSeconds}s ago`
    
    const diffMinutes = Math.floor(diffSeconds / 60)
    if (diffMinutes < 60) return `${diffMinutes}m ago`
    
    const diffHours = Math.floor(diffMinutes / 60)
    if (diffHours < 24) return `${diffHours}h ago`
    
    const diffDays = Math.floor(diffHours / 24)
    return `${diffDays}d ago`
  }

  function clearAll() {
    notifData.set([])
  }

  function dismiss(id: number) {
    notifData.update((n) => n.filter((x) => x.id !== id))
  }

  function isImageUrl(url: string | undefined): boolean {
    if (!url) return false
    return (
      url.startsWith('http') ||
      url.startsWith('data:') ||
      url.startsWith('/') ||
      url.startsWith('file:') ||
      url.includes(':\\')
    )
  }
</script>

<div class="notification-container">
  {#if !expanded}
    <div class="banner-view-container">
      {#key latest.id}
        <div
          class="banner-view"
          in:fly={{ y: -10, duration: 300 }}
          out:fade={{ duration: 150 }}
        >
          <div class="banner-icon">
            {#if isImageUrl(latest.icon)}
              <img src={latest.icon} alt="app icon" class="notif-img-icon" />
            {:else if latest.icon}
              <span class="icon-text">{latest.icon}</span>
            {:else}
              <Bell size={18} color="var(--c-text-secondary)" />
            {/if}
          </div>
          <div class="banner-content">
            <span class="banner-app">{latest.appName || 'Notification'}</span>
            <span class="banner-title">{latest.title}</span>
            {#if latest.message}
              <span class="banner-message">{latest.message}</span>
            {/if}
          </div>
        </div>
      {/key}
    </div>
  {:else}
    <div class="expanded-view">
      <div class="notif-header">
        <h3>Notifications</h3>
        {#if notifications.length > 0}
          <button class="clear-btn" onclick={clearAll}>
            <Trash2 size={14} />
            Clear
          </button>
        {/if}
      </div>
      <div class="notif-list">
        {#each notifications as notif (notif.id)}
          <div
            class="notif-item"
            animate:flip={{ duration: 300 }}
            in:fly={{ x: -20, duration: 300 }}
            out:fade={{ duration: 150 }}
          >
            <div class="notif-item-header">
              <div class="notif-item-app-group">
                <div class="notif-item-icon">
                  {#if isImageUrl(notif.icon)}
                    <img
                      src={notif.icon}
                      alt="app icon"
                      class="notif-img-icon"
                    />
                  {:else if notif.icon}
                    <span class="icon-text">{notif.icon}</span>
                  {:else}
                    <Bell size={16} color="var(--c-text-secondary)" />
                  {/if}
                </div>
                <span class="notif-item-app"
                  >{notif.appName || 'Notification'}</span
                >
              </div>
              <div class="time-and-dismiss">
                <span class="notif-item-time">{getRelativeTime(notif.timestamp)}</span>
                <button class="dismiss-item-btn" onclick={(e) => { e.stopPropagation(); dismiss(notif.id); }}>
                  <X size={14} />
                </button>
              </div>
            </div>
            <div class="notif-item-body">
              <span class="notif-item-title">{notif.title}</span>
              <p class="notif-item-message">{notif.message}</p>
            </div>
          </div>
        {/each}
        {#if notifications.length === 0}
          <div class="empty-state">No new notifications</div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .notification-container {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
    color: var(--c-text-primary);
  }

  .banner-view-container {
    width: 100%;
    height: 100%;
    display: grid;
    place-items: center;
  }

  .banner-view {
    grid-area: 1 / 1;
    display: flex;
    align-items: center;
    width: 100%;
    height: 100%;
    padding: 0 16px 0 20px;
    box-sizing: border-box;
    gap: 14px;
  }

  .banner-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    flex-shrink: 0;
    font-size: 22px;
    background: var(--c-item-bg);
    border-radius: 8px;
    overflow: hidden;
  }

  .notif-img-icon {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .banner-content {
    display: flex;
    flex-direction: column;
    justify-content: center;
    overflow: hidden;
    text-align: left;
    flex-grow: 1;
  }

  .banner-app {
    font-size: 0.7rem;
    color: var(--c-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    font-weight: 700;
    line-height: 1;
    margin-bottom: 3px;
  }

  .banner-title {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--c-text-primary);
    line-height: 1.2;
    margin-bottom: 2px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .banner-message {
    font-size: 0.85rem;
    color: var(--c-text-secondary);
    line-height: 1.2;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .expanded-view {
    padding: 20px 20px 60px 20px;
    height: 100%;
    width: 100%;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
  }

  .notif-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 10px;
    width: 100%;
    margin-bottom: 24px;
    flex-shrink: 0;
  }

  h3 {
    color: var(--c-text-primary);
    font-size: 16px;
    font-weight: 600;
    letter-spacing: -0.01em;
  }

  .dismiss-item-btn {
    background: rgba(255, 255, 255, 0.06);
    border: none;
    color: var(--c-text-muted);
    width: 24px;
    height: 24px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s ease;
    opacity: 0;
  }

  .notif-item:hover .dismiss-item-btn {
    opacity: 1;
  }

  .dismiss-item-btn:hover {
    background: rgba(255, 255, 255, 0.15);
    color: var(--c-text-primary);
  }

  .clear-btn {
    background: rgba(255, 255, 255, 0.05);
    border: none;
    color: var(--c-text-secondary);
    font-size: 12px;
    font-weight: 500;
    padding: 6px 10px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    gap: 6px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .clear-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: var(--c-text-primary);
  }

  .notif-list {
    width: 100%;
    flex: 1;
    min-height: 0; /* CRITICAL: prevents flex child from overflowing parent */
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding-right: 5px; /* for scrollbar */
  }

  .notif-list::-webkit-scrollbar {
    width: 4px;
  }
  .notif-list::-webkit-scrollbar-track {
    background: transparent;
  }
  .notif-list::-webkit-scrollbar-thumb {
    background: var(--c-scrollbar-thumb);
    border-radius: 4px;
  }

  .notif-item {
    padding: 12px 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    position: relative;
  }

  .notif-item:last-child {
    border-bottom: none;
  }

  .notif-item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    margin-bottom: 4px;
  }

  .time-and-dismiss {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .notif-item-app-group {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .notif-item-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    flex-shrink: 0;
    font-size: 14px;
    background: transparent;
    border-radius: 4px;
    overflow: hidden;
  }

  .notif-item-app {
    font-size: 0.8rem;
    font-weight: 500;
    color: var(--c-text-muted);
  }

  .notif-item-time {
    font-size: 0.75rem;
    color: var(--c-text-muted);
  }

  .notif-item-body {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    text-align: left;
    gap: 2px;
    min-width: 0;
  }

  .notif-item-title {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--c-text-primary);
  }

  .notif-item-message {
    margin: 0;
    font-size: 0.85rem;
    color: var(--c-text-muted);
    line-height: 1.3;
  }

  .empty-state {
    text-align: center;
    color: var(--c-text-muted);
    margin-top: 2rem;
    font-size: 0.9rem;
  }
</style>
