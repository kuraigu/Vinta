<script lang="ts">
  import { flip } from 'svelte/animate'
  import { fly, fade } from 'svelte/transition'

  let { expanded = false, notifications = [] } = $props<{
    expanded?: boolean
    notifications?: Array<{
      id: number
      appName?: string
      icon?: string
      title: string
      message: string
      time?: string
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

  function isImageUrl(url: string | undefined): boolean {
    if (!url) return false
    return (
      url.startsWith('http') || url.startsWith('data:') || url.startsWith('/')
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
            {:else}
              <span class="icon-text">{latest.icon || '💬'}</span>
            {/if}
          </div>
          <div class="banner-content">
            <span class="banner-title">{latest.title}</span>
            <span class="banner-message">{latest.message}</span>
          </div>
        </div>
      {/key}
    </div>
  {:else}
    <div class="expanded-view">
      <div class="notif-header">
        <h3>Notifications</h3>
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
                  {:else}
                    <span class="icon-text">{notif.icon || '💬'}</span>
                  {/if}
                </div>
                <span class="notif-item-app"
                  >{notif.appName || 'Notification'}</span
                >
              </div>
              <span class="notif-item-time">{notif.time || 'Now'}</span>
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
    padding: 0 15px;
    box-sizing: border-box;
    gap: 12px;
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
    flex: 1;
    min-width: 0;
    gap: 2px;
  }

  .banner-title,
  .banner-message {
    display: block;
    white-space: nowrap;
    text-overflow: ellipsis;
    overflow: hidden;
    width: 100%;
  }

  .banner-title {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--c-text-primary);
    line-height: 1.1;
  }

  .banner-message {
    font-size: 0.85rem;
    color: var(--c-text-muted);
    line-height: 1.1;
  }

  .expanded-view {
    padding: 1.5rem;
    width: 100%;
    height: 100%;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
  }

  .notif-header {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    margin-bottom: 1rem;
    flex-shrink: 0;
  }

  h3 {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
  }

  .notif-list {
    width: 100%;
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding-right: 5px; /* for scrollbar */
  }

  /* Custom scrollbar */
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
    background: var(--c-item-bg);
    border-radius: 12px;
    padding: 12px 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    transition: background 0.2s;
  }

  .notif-item:hover {
    background: var(--c-hover);
  }

  .notif-item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
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
