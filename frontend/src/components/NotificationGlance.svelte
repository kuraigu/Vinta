<script lang="ts">
  import { getContext } from 'svelte'
  import { Bell, X, ChevronRight } from 'lucide-svelte'
  import { notifData, type IslandConfig } from '../store'
  import NotificationPage, {
    config as notifConfig
  } from '../pages/NotificationPage.svelte'
  import { nativeIconsStore } from '../services/iconService'
  import { islandTooltip } from '../services/islandTooltip'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let notifications = $derived($notifData)
  let latestNotif = $derived(notifications.length > 0 ? notifications[0] : null)
  let unreadCount = $derived(notifications.length)
  let nativeIcons = $derived($nativeIconsStore)

  function openNotifications(e?: Event) {
    if (e) e.stopPropagation()
    if (navigate) {
      navigate(NotificationPage, notifConfig, true)
    }
  }

  function dismissLatest(e: Event) {
    e.stopPropagation()
    if (!latestNotif) return
    const idToDismiss = latestNotif.id
    notifData.update((items) => items.filter((n) => n.id !== idToDismiss))
  }

  function resolveIconUrl(notif: typeof latestNotif): string | null {
    if (!notif) return null
    if (notif.icon && notif.icon.length > 0) return notif.icon
    if (notif.aumid && nativeIcons[notif.aumid]) return nativeIcons[notif.aumid]
    if (notif.appName && nativeIcons[notif.appName])
      return nativeIcons[notif.appName]
    if (notif.appLogo) return notif.appLogo
    return null
  }

  let resolvedIcon = $derived(resolveIconUrl(latestNotif))
  let appDisplayName = $derived(
    latestNotif?.appName || latestNotif?.aumid || 'Notification'
  )
</script>

{#if latestNotif}
  <div
    class="notification-glance-card"
    role="button"
    tabindex="0"
    onclick={openNotifications}
    onkeydown={(e) => e.key === 'Enter' && openNotifications()}
  >
    <div class="glance-top-row">
      <div class="glance-brand-group">
        <div class="glance-avatar-container">
          {#if resolvedIcon}
            <img
              src={resolvedIcon}
              alt=""
              class="glance-app-icon"
              onerror={(e) => {
                const target = e.currentTarget as HTMLImageElement
                target.style.display = 'none'
                const fallback = target.nextElementSibling as HTMLElement
                if (fallback) fallback.style.display = 'flex'
              }}
            />
            <div class="glance-fallback-icon" style="display: none;">
              <Bell size={13} />
            </div>
          {:else}
            <div class="glance-fallback-icon">
              <Bell size={13} />
            </div>
          {/if}
        </div>
        <span class="glance-app-name">{appDisplayName}</span>
        {#if unreadCount > 1}
          <span class="glance-count-badge">+{unreadCount - 1}</span>
        {/if}
      </div>

      <div class="glance-actions-group">
        <button
          type="button"
          class="glance-action-btn view"
          onclick={openNotifications}
          use:islandTooltip={'Open Notification Center'}
          aria-label="View notifications"
        >
          <span>View</span>
          <ChevronRight size={13} />
        </button>
        <button
          type="button"
          class="glance-action-btn dismiss"
          onclick={dismissLatest}
          use:islandTooltip={'Dismiss notification'}
          aria-label="Dismiss this notification"
        >
          <X size={12} />
        </button>
      </div>
    </div>

    <div class="glance-content-preview">
      {#if latestNotif.title}
        <h4 class="glance-title">{latestNotif.title}</h4>
      {/if}
      {#if latestNotif.message}
        <p class="glance-message">{latestNotif.message}</p>
      {/if}
    </div>
  </div>
{/if}

<style>
  .notification-glance-card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;
    padding: 10px 14px;
    border-radius: var(--vinta-radius-card, 16px);
    background: var(--vinta-surface-card, rgba(255, 255, 255, 0.04));
    border: 1px solid var(--vinta-border, rgba(255, 255, 255, 0.08));
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    box-sizing: border-box;
    cursor: pointer;
    text-align: left;
    transition:
      background 0.2s cubic-bezier(0.16, 1, 0.3, 1),
      border-color 0.2s cubic-bezier(0.16, 1, 0.3, 1),
      transform 0.15s ease;
    user-select: none;
    -webkit-user-select: none;
  }

  .notification-glance-card:hover {
    background: var(--vinta-hover, rgba(255, 255, 255, 0.07));
    border-color: var(--vinta-border-light, rgba(255, 255, 255, 0.14));
  }

  .notification-glance-card:active {
    transform: scale(0.992);
  }

  .glance-top-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    gap: 8px;
  }

  .glance-brand-group {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex: 1;
  }

  .glance-avatar-container {
    width: 20px;
    height: 20px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    flex-shrink: 0;
    background: rgba(255, 255, 255, 0.08);
  }

  .glance-app-icon {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .glance-fallback-icon {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--vinta-accent-primary, #60a5fa);
  }

  .glance-app-name {
    font-size: 0.78rem;
    font-weight: 600;
    letter-spacing: -0.01em;
    color: var(--vinta-text-secondary, rgba(255, 255, 255, 0.65));
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-transform: capitalize;
  }

  .glance-count-badge {
    font-size: 0.68rem;
    font-weight: 700;
    color: var(--vinta-accent-primary, #60a5fa);
    background: var(--vinta-accent-muted, rgba(96, 165, 250, 0.16));
    border: 1px solid var(--vinta-accent-muted, rgba(96, 165, 250, 0.25));
    border-radius: 999px;
    padding: 1px 6px;
    line-height: 1.2;
    flex-shrink: 0;
  }

  .glance-actions-group {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
  }

  .glance-action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    cursor: pointer;
    border-radius: 6px;
    color: var(--vinta-text-muted, rgba(255, 255, 255, 0.45));
    transition: all 0.15s ease;
    padding: 3px 6px;
    font-size: 0.72rem;
    font-weight: 500;
    gap: 2px;
  }

  .glance-action-btn.view {
    background: rgba(255, 255, 255, 0.06);
    color: var(--vinta-text-secondary, rgba(255, 255, 255, 0.8));
    padding: 3px 8px;
    border-radius: 12px;
  }

  .glance-action-btn.view:hover {
    background: var(--vinta-accent-muted, rgba(96, 165, 250, 0.2));
    color: var(--vinta-accent-primary, #60a5fa);
  }

  .glance-action-btn.dismiss {
    width: 20px;
    height: 20px;
    padding: 0;
    border-radius: 50%;
  }

  .glance-action-btn.dismiss:hover {
    background: rgba(255, 255, 255, 0.12);
    color: var(--vinta-text-primary, #fff);
  }

  .glance-content-preview {
    display: flex;
    flex-direction: column;
    gap: 2px;
    width: 100%;
  }

  .glance-title {
    margin: 0;
    font-size: 0.84rem;
    font-weight: 600;
    letter-spacing: -0.015em;
    color: var(--vinta-text-primary, #ffffff);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .glance-message {
    margin: 0;
    font-size: 0.76rem;
    line-height: 1.35;
    color: var(--vinta-text-secondary, rgba(255, 255, 255, 0.65));
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    word-break: break-word;
  }
</style>
