<script lang="ts">
  import DotIndicator from './DotIndicator.svelte'

  export let expanded: boolean = false
  export let notifications: Array<{
    id: number
    title: string
    message: string
  }> = []

  $: latest = notifications[0] || { title: 'No notifications', message: '' }
</script>

<div class="notification-container">
  {#if !expanded}
    <div class="banner-view">
      <div class="banner-icon">
        <DotIndicator color="#007aff" size="12px" />
      </div>
      <div class="banner-content">
        <span class="banner-title">{latest.title}</span>
        <span class="banner-message">{latest.message}</span>
      </div>
    </div>
  {:else}
    <div class="expanded-view">
      <div class="notif-header">
        <DotIndicator color="#007aff" />
        <h3>Notifications</h3>
      </div>
      <div class="notif-list">
        {#each notifications as notif (notif.id)}
          <div class="notif-item">
            <div class="notif-item-header">
              <span class="notif-item-title">{notif.title}</span>
            </div>
            <p class="notif-item-message">{notif.message}</p>
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

  .banner-view {
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
    width: 24px;
    height: 24px;
    flex-shrink: 0;
  }

  .banner-content {
    display: flex;
    flex-direction: column;
    justify-content: center;
    overflow: hidden;
    white-space: nowrap;
    text-align: left;
    flex: 1;
    gap: 2px;
  }

  .banner-title {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--c-text-primary);
    line-height: 1.1;
  }

  .banner-message {
    font-size: 0.85rem;
    color: #b0b0b0;
    text-overflow: ellipsis;
    overflow: hidden;
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
    background: rgba(255, 255, 255, 0.2);
    border-radius: 4px;
  }

  .stack-item {
    background: var(--c-item-bg);
    border-radius: 12px;
    padding: 12px 16px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    transition: background 0.2s;
  }

  .stack-item:hover {
    background: var(--c-hover);
  }

  .notif-item-title {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--c-text-primary);
  }

  .notif-item-message {
    margin: 0;
    font-size: 0.85rem;
    color: #b0b0b0;
    line-height: 1.3;
  }

  .empty-state {
    text-align: center;
    color: #888;
    margin-top: 2rem;
    font-size: 0.9rem;
  }
</style>
