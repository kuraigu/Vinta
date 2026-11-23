<!--
  @component NotificationGlance
  @description Compact notification ticker in the collapsed island showing the latest toast snippet with sender avatar support.
-->
<script lang="ts">
  import { getContext } from 'svelte'
  import { Bell, X } from 'lucide-svelte'
  import {
    notifData,
    privacyModeStore,
    autoPrivacyModeStore,
    screenShareDetectedStore,
    pageVisibilityStore,
    tStore,
    type IslandConfig
  } from '../store'
  import Notification, { config as notifConfig } from './Notification.svelte'
  import {
    nativeIconsStore,
    parseAumid,
    requestNativeAppIcon
  } from '../services/iconService'
  import { islandTooltip } from '../services/islandTooltip'
  import BentoCard from './BentoCard.svelte'
  import appIcon from '@appicon'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  const SENDER_AVATAR_STORAGE_KEY = 'vinta_sender_avatars'

  /**
   * Retrieves persistent contact/sender profile pictures from local storage.
   */
  function loadSenderAvatars(): Record<string, string> {
    try {
      const raw = localStorage.getItem(SENDER_AVATAR_STORAGE_KEY)
      if (raw) return JSON.parse(raw)
    } catch {
      // Storage unavailable or unparseable
    }
    return {}
  }

  let notifications = $derived(
    $notifData.filter(
      (n) =>
        !n.isUpdate &&
        !n.title?.toLowerCase().includes('update available') &&
        !(
          (n.appName?.toLowerCase() === 'vinta' ||
            n.aumid?.toLowerCase().includes('vinta')) &&
          n.title?.toLowerCase().includes('update')
        )
    )
  )
  let latestNotif = $derived(notifications.length > 0 ? notifications[0] : null)
  let unreadCount = $derived(notifications.length)
  let nativeIcons = $derived($nativeIconsStore)

  function openNotifications(e?: Event) {
    if (e) e.stopPropagation()
    if (!$pageVisibilityStore.notifications) return
    if (navigate) {
      navigate(Notification, notifConfig, true)
    }
  }

  function dismissLatest(e: Event) {
    e.stopPropagation()
    if (!latestNotif) return
    const idToDismiss = latestNotif.id
    notifData.update((items) => items.filter((n) => n.id !== idToDismiss))
  }

  function getSenderKey(notif: typeof latestNotif): string {
    if (!notif) return ''
    const app = (notif.appName || notif.aumid || '').trim().toLowerCase()
    if (app.includes('vinta')) return ''
    const sender = (notif.title || '').trim().toLowerCase()
    if (!sender) return ''
    return `${app}::${sender}`
  }

  /**
   * Resolves the parent application icon for the top brand header.
   */
  function resolveAppIcon(notif: typeof latestNotif): string | null {
    if (!notif) return null

    const rawInput = notif.appName || notif.aumid || ''
    const isVinta =
      rawInput.toLowerCase().includes('vinta') ||
      (notif.title || '').toLowerCase().includes('vinta')

    if (isVinta) {
      return appIcon
    }

    if (notif.appLogo && notif.appLogo.length > 0) {
      return notif.appLogo
    }

    const parsed = parseAumid(rawInput)
    const searchApp = (parsed.processName || notif.appName || '').trim()
    const searchAumid = (notif.aumid || '').trim()
    const iconKey = `${searchApp.toLowerCase()}|${searchAumid.toLowerCase()}`

    if (nativeIcons[iconKey]) {
      return nativeIcons[iconKey]
    }
    if (notif.aumid && nativeIcons[notif.aumid]) {
      return nativeIcons[notif.aumid]
    }
    if (notif.appName && nativeIcons[notif.appName]) {
      return nativeIcons[notif.appName]
    }

    if (searchApp || searchAumid) {
      const extracted = requestNativeAppIcon(searchApp, searchAumid)
      if (extracted) return extracted
    }

    return null
  }

  /**
   * Resolves custom sender/contact avatar image when distinct from the app logo.
   */
  function resolveAvatarIcon(
    notif: typeof latestNotif,
    appLogo: string | null
  ): string | null {
    if (!notif) return null

    const rawInput = notif.appName || notif.aumid || ''
    if (
      rawInput.toLowerCase().includes('vinta') ||
      (notif.title || '').toLowerCase().includes('vinta')
    ) {
      return null
    }

    if (notif.icon && notif.icon.length > 0) {
      if (!appLogo || notif.icon !== appLogo) {
        return notif.icon
      }
    }

    if (notif.heroImage && notif.heroImage.length > 0) {
      const isAvatar =
        notif.heroImage.includes('/avatars/') ||
        notif.heroImage.includes('/avatar/') ||
        notif.heroImage.includes('avatar') ||
        notif.heroImage.includes('profile') ||
        notif.heroImage.includes('user')
      if (isAvatar) {
        return notif.heroImage
      }
    }

    const sKey = getSenderKey(notif)
    if (sKey) {
      const savedAvatars = loadSenderAvatars()
      if (savedAvatars[sKey]) {
        return savedAvatars[sKey]
      }
    }

    return null
  }

  let resolvedAppLogo = $derived(resolveAppIcon(latestNotif))
  let resolvedAvatar = $derived(resolveAvatarIcon(latestNotif, resolvedAppLogo))
  let isPrivacyActive = $derived(
    $privacyModeStore || ($autoPrivacyModeStore && $screenShareDetectedStore)
  )

  let topRowIcon = $derived(resolvedAppLogo || resolvedAvatar)
  let bodyAvatar = $derived(
    resolvedAvatar && resolvedAvatar !== resolvedAppLogo ? resolvedAvatar : null
  )

  let appDisplayName = $derived(
    latestNotif?.appName ||
      (latestNotif?.aumid
        ? parseAumid(latestNotif.aumid).displayName
        : $tStore('notification.title'))
  )
</script>

{#if latestNotif}
  <BentoCard
    variant="bento"
    class="notification-glance-card"
    onclick={openNotifications}
  >
    <div class="glance-top-row">
      <div class="glance-brand-group">
        <div class="glance-avatar-container">
          {#key latestNotif?.id}
            {#if topRowIcon}
              <img
                src={topRowIcon}
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
                <Bell size={14} class="notif-brand-icon" />
              </div>
            {:else}
              <Bell size={14} class="notif-brand-icon" />
            {/if}
          {/key}
        </div>
        <span class="glance-app-name">{appDisplayName}</span>
        {#if unreadCount > 1}
          <span class="glance-count-badge"
            >{unreadCount > 9 ? '9+' : unreadCount}</span
          >
        {/if}
      </div>

      <div class="glance-actions-group">
        <button
          type="button"
          class="glance-action-btn dismiss"
          onclick={dismissLatest}
          use:islandTooltip={$tStore('notification.dismiss_tooltip')}
          aria-label={$tStore('notification.dismiss_tooltip')}
        >
          <X size={12} />
        </button>
      </div>
    </div>

    <div class="glance-content-preview" class:has-avatar={!!bodyAvatar}>
      {#if bodyAvatar}
        <div class="glance-avatar-bubble">
          <img
            src={bodyAvatar}
            alt=""
            class="glance-avatar-img"
            onerror={(e) => {
              const target = e.currentTarget as HTMLImageElement
              if (target.parentElement)
                target.parentElement.style.display = 'none'
            }}
          />
        </div>
      {/if}
      <div class="glance-text-col">
        {#if isPrivacyActive}
          <h4 class="glance-title">{appDisplayName}</h4>
          <p class="glance-message">
            {$tStore('notification.privacy_hidden_preview')}
          </p>
        {:else}
          {#if latestNotif.title}
            <h4 class="glance-title">{latestNotif.title}</h4>
          {/if}
          {#if latestNotif.message}
            <p class="glance-message">{latestNotif.message}</p>
          {/if}
        {/if}
      </div>
    </div>
  </BentoCard>
{/if}

<style src="./styles/NotificationGlance.css"></style>
