<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 368, h: 44, r: 16 },
    expanded: { w: 400, h: 225, r: 28 },
    borderColor: 'rgba(255, 60, 100, 0.4)',
    glowColor: 'rgba(255, 60, 100, 0.35)'
  }
</script>

<script lang="ts">
  import {
    SendMediaControl,
    SelectMediaSession,
    FocusNotificationApp,
    RequeryMediaArt
  } from '../../bindings/Vinta/backend/app.js'
  import Clock from './Clock.svelte'
  import BatteryWidget from './BatteryWidget.svelte'
  import DynamicEqualizer from './DynamicEqualizer.svelte'
  import {
    globalMediaStore,
    audioPeakStore,
    notifData,
    batteryDataStore,
    homeWidgetsVisibilityStore
  } from '../store'
  import {
    Volume1,
    Volume2,
    VolumeX,
    Bell,
    Shuffle,
    Repeat,
    Repeat1,
    ListMusic,
    Music2,
    Check,
    X
  } from 'lucide-svelte'
  import { getContext, untrack } from 'svelte'
  import NotificationPage, {
    config as notifConfig
  } from '../pages/NotificationPage.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import { resolveAppBrand, nativeIconsStore } from '../services/iconService'
  import { enhanceArtwork } from '../utils/imageUpscaler'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  interface Props {
    expanded?: boolean
  }

  let { expanded = false }: Props = $props()

  let media = $derived($globalMediaStore)
  let hasBattery = $derived($batteryDataStore.hasBattery)
  let nativeIcons = $derived($nativeIconsStore)

  function computeAppBrand() {
    void nativeIcons
    return resolveAppBrand(media.sourceApp)
  }
  let appBrand = $derived(computeAppBrand())

  function getTrackTextureStyle(
    title?: string,
    artist?: string,
    sourceApp?: string
  ) {
    const str = `${title || ''}-${artist || ''}-${sourceApp || ''}`.trim()
    let hash = 0
    for (let i = 0; i < str.length; i++) {
      hash = str.charCodeAt(i) + ((hash << 5) - hash)
    }
    const h1 = Math.abs(hash % 360)
    const h2 = (h1 + 45 + Math.abs((hash >> 3) % 90)) % 360
    const h3 = (h1 + 180 + Math.abs((hash >> 5) % 60)) % 360

    const c1 = `hsl(${h1}, 75%, 28%)`
    const c2 = `hsl(${h2}, 80%, 20%)`
    const c3 = `hsl(${h3}, 70%, 15%)`
    const base = `hsl(${h1}, 40%, 12%)`

    const p1x = 20 + Math.abs((hash >> 2) % 60)
    const p1y = 20 + Math.abs((hash >> 4) % 60)
    const p2x = 80 - Math.abs((hash >> 6) % 50)
    const p2y = 75 - Math.abs((hash >> 8) % 45)

    return {
      background: `radial-gradient(circle at ${p1x}% ${p1y}%, ${c1} 0%, transparent 65%), radial-gradient(circle at ${p2x}% ${p2y}%, ${c2} 0%, transparent 60%), radial-gradient(circle at 50% 100%, ${c3} 0%, transparent 70%), ${base}`,
      accentColor: c1
    }
  }

  function computeTrackTexture() {
    return getTrackTextureStyle(media.title, media.artist, media.sourceApp)
  }
  let trackTexture = $derived(computeTrackTexture())

  let isDraggingSeek = $state(false)
  let seekVal = $state(0)
  let localPosMs = $state(0)
  let isDraggingVol = $state(false)
  let volVal = $state(50)
  let isVolLocked = $state(false)
  let volLockTimer: ReturnType<typeof setTimeout> | null = null
  let isSessionMenuOpen = $state(false)
  let lastNonZeroVol = $state(50)

  function handleFocusApp(e: MouseEvent) {
    e.stopPropagation()
    const process = appBrand.processName || media.sourceApp || ''
    const aumid =
      appBrand.aumid ||
      (media.sourceApp?.includes('!') ? media.sourceApp : '') ||
      ''
    if (process || aumid) {
      FocusNotificationApp(process, aumid)
    }
  }

  function toggleSessionMenu(e: MouseEvent) {
    e.stopPropagation()
    isSessionMenuOpen = !isSessionMenuOpen
  }

  function handleSelectSession(sessionId: string, e: MouseEvent) {
    e.stopPropagation()
    SelectMediaSession(sessionId)
    isSessionMenuOpen = false
  }

  function handleSessionDismissEffect() {
    function onPointerDown(e: PointerEvent) {
      if (isSessionMenuOpen) {
        const target = e.target as HTMLElement
        if (
          !target.closest('.session-drawer-panel') &&
          !target.closest('.session-btn')
        ) {
          isSessionMenuOpen = false
        }
      }
    }
    window.addEventListener('pointerdown', onPointerDown)
    return () => {
      window.removeEventListener('pointerdown', onPointerDown)
    }
  }
  $effect(handleSessionDismissEffect)

  function lockVolumeWithGracePeriod() {
    isVolLocked = true
    if (volLockTimer) clearTimeout(volLockTimer)
    volLockTimer = setTimeout(() => {
      isVolLocked = false
      volLockTimer = null
    }, 850)
  }

  let prevNotifCount = $state(0)
  let isNewNotif = $state(false)
  let pulseTimeout: number

  function handleNotifPulseEffect() {
    const currentCount = $notifData.length
    if (currentCount > prevNotifCount) {
      isNewNotif = true
      clearTimeout(pulseTimeout)
      pulseTimeout = window.setTimeout(() => {
        isNewNotif = false
      }, 3500)
    }
    prevNotifCount = currentCount
  }
  $effect(handleNotifPulseEffect)

  let isSeekingLocked = $state(false)
  let targetSeekPos = $state(0)
  let driftError = 0
  let seekLockTimer: ReturnType<typeof setTimeout> | null = null

  let playLockTimer: ReturnType<typeof setTimeout> | null = null

  // Track synchronization with incoming media store updates
  let prevTitle = ''
  let prevArtist = ''
  let prevSessionId = ''
  let prevSourceApp = ''

  let isPlayingState = $derived(Boolean(media.isPlaying))
  let durationMsState = $derived(Number(media.durationMs || 0))

  function handleMediaSyncEffect() {
    const curTitle = media.title || ''
    const curArtist = media.artist || ''
    const curSessionId = media.activeSessionId || ''
    const curSourceApp = media.sourceApp || ''
    const backendPos = media.positionMs || 0
    const playing = Boolean(media.isPlaying)

    const isTrackChanged =
      curTitle !== prevTitle ||
      curArtist !== prevArtist ||
      (curSessionId !== '' && curSessionId !== prevSessionId) ||
      (curSourceApp !== '' && curSourceApp !== prevSourceApp)

    if (!isDraggingSeek) {
      if (isTrackChanged || (!playing && !isSeekingLocked)) {
        // Track change, session switch, or paused (not actively seeking):
        // sync directly to backend snapshot
        localPosMs = backendPos
        seekVal = backendPos
        driftError = 0
        isSeekingLocked = false
        if (seekLockTimer) {
          clearTimeout(seekLockTimer)
          seekLockTimer = null
        }
      } else if (isSeekingLocked) {
        // While locked after a seek, wait for external player to arrive near target
        if (Math.abs(backendPos - targetSeekPos) < 2000) {
          isSeekingLocked = false
          if (seekLockTimer) {
            clearTimeout(seekLockTimer)
            seekLockTimer = null
          }
          const currentLocal = untrack(() => localPosMs)
          driftError = backendPos - currentLocal
        }
      } else {
        // While actively playing, calculate drift between backend and local playhead
        const currentLocal = untrack(() => localPosMs)
        const diff = backendPos - currentLocal

        if (Math.abs(diff) > 3000 || (currentLocal === 0 && backendPos > 0)) {
          // Major seek or skip: snap immediately
          localPosMs = backendPos
          seekVal = backendPos
          driftError = 0
        } else if (Math.abs(diff) > 40) {
          // Routine drift: assign to driftError for smooth damping
          driftError = diff
        } else {
          // In sync within 40ms
          driftError = 0
        }
      }
    }

    prevTitle = curTitle
    prevArtist = curArtist
    prevSessionId = curSessionId
    prevSourceApp = curSourceApp

    if (!isDraggingVol && !isVolLocked && media.volume !== undefined) {
      volVal = media.volume
    }
  }
  $effect(handleMediaSyncEffect)

  // Artwork fallback: if an active media session is reported without album artwork,
  // request a thumbnail re-query after a brief settling delay.
  let requeryTimer: ReturnType<typeof setTimeout> | null = null
  function handleArtRequeryEffect() {
    const isAvailable = Boolean(media.isAvailable)
    const hasArt = Boolean(media.albumArt)
    const track = media.title || ''

    if (requeryTimer) {
      clearTimeout(requeryTimer)
      requeryTimer = null
    }

    if (isAvailable && !hasArt && track) {
      requeryTimer = setTimeout(() => {
        if (typeof RequeryMediaArt === 'function') {
          RequeryMediaArt()
        }
        requeryTimer = null
      }, 1500)
    }

    return () => {
      if (requeryTimer) {
        clearTimeout(requeryTimer)
        requeryTimer = null
      }
    }
  }
  $effect(handleArtRequeryEffect)

  let imgLoadFailed = $state(false)
  let lastArtSrc = $state('')
  let displayAlbumArt = $state('')
  let isArtUpscaled = $state(false)

  /**
   * Synchronizes album artwork and initiates non-AI detail enhancement for low-resolution thumbnails.
   */
  function handleArtSyncEffect() {
    const rawArt = media.albumArt
    if (rawArt !== lastArtSrc) {
      lastArtSrc = rawArt
      imgLoadFailed = false

      if (!rawArt) {
        displayAlbumArt = ''
        isArtUpscaled = false
        return
      }

      displayAlbumArt = rawArt
      isArtUpscaled = false

      enhanceArtwork(rawArt).then((res) => {
        if (media.albumArt === rawArt && res.isUpscaled) {
          displayAlbumArt = res.url
          isArtUpscaled = true
        }
      })
    }
  }
  $effect(handleArtSyncEffect)

  let cardTintColor = $state('')

  /**
   * Samples dominant artwork hue across downsampled pixels to supply ambient backdrop tinting.
   *
   * @param src - Base64 data URI or image URL of the album artwork.
   */
  function extractArtworkColor(src: string) {
    if (!src) {
      cardTintColor = trackTexture.accentColor
      return
    }
    const img = new Image()
    img.crossOrigin = 'anonymous'
    img.onload = () => {
      try {
        const canvas = document.createElement('canvas')
        canvas.width = 16
        canvas.height = 16
        const ctx = canvas.getContext('2d')
        if (!ctx) {
          cardTintColor = trackTexture.accentColor
          return
        }
        ctx.drawImage(img, 0, 0, 16, 16)
        const data = ctx.getImageData(0, 0, 16, 16).data
        let r = 0
        let g = 0
        let b = 0
        let count = 0
        for (let i = 0; i < data.length; i += 4) {
          r += data[i]
          g += data[i + 1]
          b += data[i + 2]
          count++
        }
        if (count > 0) {
          cardTintColor = `rgba(${Math.round(r / count)}, ${Math.round(g / count)}, ${Math.round(b / count)}, 0.45)`
        }
      } catch {
        cardTintColor = trackTexture.accentColor
      }
    }
    img.onerror = () => {
      cardTintColor = trackTexture.accentColor
    }
    img.src = src
  }

  /**
   * Harmonizes dynamic background scrim tint with active album artwork or procedural texture palette.
   */
  function handleCardTintEffect() {
    if (displayAlbumArt && !imgLoadFailed) {
      extractArtworkColor(displayAlbumArt)
    } else {
      cardTintColor = trackTexture.accentColor
    }
  }
  $effect(handleCardTintEffect)

  function handleImgError() {
    imgLoadFailed = true
  }

  // Smooth local timeline progression between backend SMTC updates.
  // Reactive ONLY to isPlayingState and isDraggingSeek to prevent routine
  // media-update events (Discord, volume, audio peaks) from cancelling the rAF loop.
  function handleTimelineProgressEffect() {
    if (!isPlayingState || isDraggingSeek) return

    let lastTick = performance.now()
    let rafId: number

    const tick = () => {
      const now = performance.now()
      const elapsed = now - lastTick
      lastTick = now

      const totalDur = untrack(() => durationMsState)

      if (!isDraggingSeek && totalDur > 0) {
        // Smooth proportional drift damping: bleed off ~10% of remaining error per frame
        const err = driftError
        let correction = 0
        if (Math.abs(err) > 5) {
          correction = err * 0.1
          driftError -= correction
        } else if (err !== 0) {
          driftError = 0
        }

        const nextPos = Math.min(
          totalDur,
          Math.max(0, localPosMs + elapsed + correction)
        )
        localPosMs = nextPos
        seekVal = nextPos
      }

      if (untrack(() => isPlayingState) && !isDraggingSeek) {
        rafId = requestAnimationFrame(tick)
      }
    }

    rafId = requestAnimationFrame(tick)
    return () => {
      if (rafId) cancelAnimationFrame(rafId)
    }
  }
  $effect(handleTimelineProgressEffect)

  function handleControl(action: string, e?: MouseEvent) {
    if (e) e.stopPropagation()
    if (action === 'play_pause' || action === 'play' || action === 'pause') {
      const nextPlaying =
        action === 'play_pause'
          ? !media.isPlaying
          : action === 'play'
            ? true
            : false

      if (playLockTimer) clearTimeout(playLockTimer)
      playLockTimer = setTimeout(() => {
        globalMediaStore.update((m) => ({
          ...m,
          _isPlayStateLocked: false
        }))
        playLockTimer = null
      }, 1200)

      globalMediaStore.update((m) => ({
        ...m,
        isPlaying: nextPlaying,
        _isPlayStateLocked: true
      }))
    } else if (action === 'prev' || action === 'previous') {
      localPosMs = 0
      seekVal = 0
      targetSeekPos = 0
      driftError = 0
      isSeekingLocked = true
      if (seekLockTimer) clearTimeout(seekLockTimer)
      seekLockTimer = setTimeout(() => {
        isSeekingLocked = false
        seekLockTimer = null
      }, 1000)
    } else if (action === 'shuffle') {
      globalMediaStore.update((m) => ({
        ...m,
        isShuffleActive: !m.isShuffleActive
      }))
    } else if (action === 'repeat') {
      globalMediaStore.update((m) => ({
        ...m,
        autoRepeatMode: ((m.autoRepeatMode || 0) + 1) % 3
      }))
    } else if (action === 'toggle_mute' || action === 'mute') {
      const willMute = !media.isMuted
      if (willMute) {
        if (volVal > 0) lastNonZeroVol = volVal
        volVal = 0
        globalMediaStore.update((m) => ({
          ...m,
          isMuted: true,
          volume: 0
        }))
        SendMediaControl(`set_volume:0:${media.sourceApp || ''}`)
        SendMediaControl(`toggle_mute:${media.sourceApp || ''}`)
      } else {
        const restoreVol = lastNonZeroVol > 0 ? lastNonZeroVol : 50
        volVal = restoreVol
        globalMediaStore.update((m) => ({
          ...m,
          isMuted: false,
          volume: restoreVol
        }))
        SendMediaControl(
          `set_volume:${Math.round(restoreVol)}:${media.sourceApp || ''}`
        )
        SendMediaControl(`toggle_mute:${media.sourceApp || ''}`)
      }
      return
    }
    SendMediaControl(action)
  }

  function handleSeekStart() {
    if (seekLockTimer) {
      clearTimeout(seekLockTimer)
      seekLockTimer = null
    }
    isDraggingSeek = true
    isSeekingLocked = true
    driftError = 0
  }

  function handleSeekInput(e: Event) {
    isDraggingSeek = true
    isSeekingLocked = true
    driftError = 0
    if (seekLockTimer) {
      clearTimeout(seekLockTimer)
      seekLockTimer = null
    }
    const val = Number((e.target as HTMLInputElement).value)
    seekVal = val
    localPosMs = val
    targetSeekPos = val
  }

  let lastCommittedSeek = -1
  let lastCommittedTime = 0

  function handleSeekCommit(e: Event) {
    const now = Date.now()
    const val = Number((e.target as HTMLInputElement).value)
    if (
      now - lastCommittedTime < 100 &&
      Math.abs(val - lastCommittedSeek) < 50
    ) {
      return
    }
    lastCommittedTime = now
    lastCommittedSeek = val

    seekVal = val
    localPosMs = val
    targetSeekPos = val
    driftError = 0
    isDraggingSeek = false
    isSeekingLocked = true
    const sessionTarget = media.activeSessionId || media.sourceApp || ''
    SendMediaControl(`seek:${Math.round(val)}:${sessionTarget}`)

    if (seekLockTimer) clearTimeout(seekLockTimer)
    seekLockTimer = setTimeout(() => {
      isSeekingLocked = false
      seekLockTimer = null
    }, 2500)
  }

  let volThrottleTimer: ReturnType<typeof setTimeout> | null = null
  function handleVolInput(e: Event) {
    const val = Number((e.target as HTMLInputElement).value)
    volVal = val
    isDraggingVol = true
    lockVolumeWithGracePeriod()
    if (!volThrottleTimer) {
      SendMediaControl(`set_volume:${val}:${media.sourceApp || ''}`)
      volThrottleTimer = setTimeout(() => {
        volThrottleTimer = null
      }, 75)
    }
  }

  function handleVolChange(e: Event) {
    const val = Number((e.target as HTMLInputElement).value)
    if (volThrottleTimer) clearTimeout(volThrottleTimer)
    volThrottleTimer = null
    volVal = val
    lockVolumeWithGracePeriod()
    SendMediaControl(`set_volume:${val}:${media.sourceApp || ''}`)
    globalMediaStore.update((m) => ({ ...m, volume: val }))
    isDraggingVol = false
  }

  function handleVolPointerDown() {
    isDraggingVol = true
    lockVolumeWithGracePeriod()
  }

  function handleVolPointerUp() {
    isDraggingVol = false
    lockVolumeWithGracePeriod()
  }

  let isScrubHovering = $state(false)
  let scrubHoverPercent = $state(0)
  let scrubHoverTimeMs = $state(0)

  /**
   * Tracks cursor/pointer position along the timeline scrubber to render a live timestamp preview bubble.
   *
   * @param e - Pointer movement event over the scrubber container.
   */
  function handleScrubPointerMove(e: PointerEvent) {
    if (media.durationMs <= 0) return
    const target = e.currentTarget as HTMLElement
    const rect = target.getBoundingClientRect()
    if (rect.width <= 0) return
    const ratio = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width))
    scrubHoverPercent = ratio * 100
    scrubHoverTimeMs = Math.round(ratio * media.durationMs)
    isScrubHovering = true
  }

  /**
   * Hides the timestamp preview bubble when pointer departs the waveform scrubber track.
   */
  function handleScrubPointerLeave() {
    isScrubHovering = false
  }

  function formatTime(ms: number): string {
    if (!ms || isNaN(ms) || ms <= 0) return '0:00'
    const totalSecs = Math.floor(ms / 1000)
    const mins = Math.floor(totalSecs / 60)
    const secs = totalSecs % 60
    return `${mins}:${secs < 10 ? '0' : ''}${secs}`
  }

  let currentPos = $derived(isDraggingSeek ? seekVal : localPosMs)

  let progressPercent = $derived(
    media.durationMs > 0
      ? Math.min(100, Math.max(0, (currentPos / media.durationMs) * 100))
      : 0
  )

  let currentVol = $derived(
    isDraggingVol || isVolLocked ? volVal : (media.volume ?? 50)
  )

  function getSessionBrand(sourceApp: string, sessionId: string) {
    void nativeIcons
    return resolveAppBrand(sourceApp, sessionId)
  }

  let miniTitleNode = $state<HTMLElement | null>(null)
  let expandedTitleNode = $state<HTMLElement | null>(null)

  let isMiniTitleOverflowing = $state(false)
  let miniTitleDist = $state(0)
  let miniTitleDuration = $state(8)

  let isExpandedTitleOverflowing = $state(false)
  let expandedTitleDist = $state(0)
  let expandedTitleDuration = $state(8)

  /**
   * Calculates text overflow metrics and configures the marquee animation speed.
   *
   * @param node Target element to evaluate
   * @returns Object containing overflow status, slide distance, and duration
   */
  function measureMarquee(node: HTMLElement | null): {
    overflowing: boolean
    dist: number
    duration: number
  } {
    if (!node || !node.parentElement) {
      return { overflowing: false, dist: 0, duration: 8 }
    }
    const containerWidth = node.parentElement.clientWidth
    const contentWidth = node.scrollWidth
    const diff = contentWidth - containerWidth

    if (diff > 14 && containerWidth > 0) {
      const dist = Math.ceil(diff + 8)
      const duration = Math.max(10, Math.min(38, Math.round(dist / 14 + 7)))
      return { overflowing: true, dist, duration }
    }
    return { overflowing: false, dist: 0, duration: 8 }
  }

  /**
   * Observes and updates marquee metrics when track information or element bounds change.
   */
  function handleTitleMarqueeEffect() {
    const _title = media.title
    const _expanded = expanded
    const _mini = miniTitleNode
    const _exp = expandedTitleNode

    let rafId: number | null = null

    function evaluate() {
      if (!_expanded && _mini) {
        const res = measureMarquee(_mini)
        isMiniTitleOverflowing = res.overflowing
        miniTitleDist = res.dist
        miniTitleDuration = res.duration
      } else if (_expanded && _exp) {
        const res = measureMarquee(_exp)
        isExpandedTitleOverflowing = res.overflowing
        expandedTitleDist = res.dist
        expandedTitleDuration = res.duration
      }
    }

    rafId = requestAnimationFrame(evaluate)

    const ro = new ResizeObserver(() => {
      evaluate()
    })

    if (_mini?.parentElement) {
      ro.observe(_mini.parentElement)
    }
    if (_exp?.parentElement) {
      ro.observe(_exp.parentElement)
    }

    return () => {
      if (rafId !== null) cancelAnimationFrame(rafId)
      ro.disconnect()
    }
  }
  $effect(handleTitleMarqueeEffect)
</script>

<div class="media-container {expanded ? 'expanded' : 'minimal'}">
  {#if !expanded}
    <!-- Compact Minimal View with Spinning Vinyl Disc, Track Info, Clock, and Audio Wave Bars -->
    <div class="minimal-layout">
      <!-- Spinning Vinyl Disc Album Art -->
      <div
        class="vinyl-disc {media.isPlaying ? 'spinning' : 'paused'}"
      >
        <div class="vinyl-groove">
          {#if displayAlbumArt && !imgLoadFailed}
            <img
              src={displayAlbumArt}
              alt={media.title}
              class="vinyl-art"
              draggable="false"
              decoding="async"
              onerror={handleImgError}
            />
          {:else}
            <div
              class="vinyl-placeholder"
              style="background: {trackTexture.background};"
            >
              <div class="vinyl-noise-ring"></div>
              <svg
                viewBox="0 0 24 24"
                width="12"
                height="12"
                fill="currentColor"
              >
                <path
                  d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z"
                />
              </svg>
            </div>
          {/if}
          <!-- Vinyl Spindle Hole -->
          <div class="vinyl-hole"></div>
        </div>
      </div>

      <!-- Track Info -->
      <div class="track-info-minimal">
        <div
          class="track-title-mini-wrapper"
          class:is-overflowing={isMiniTitleOverflowing}
          style={isMiniTitleOverflowing
            ? `--marquee-dist: ${miniTitleDist}px; --marquee-duration: ${miniTitleDuration}s;`
            : undefined}
        >
          <span
            bind:this={miniTitleNode}
            class="track-title-mini"
            class:is-marquee={isMiniTitleOverflowing}
            >{media.title || 'No Media Playing'}</span
          >
        </div>
        {#if media.artist}
          <span class="artist-name-mini">{media.artist}</span>
        {:else if !media.title}
          <span class="artist-name-mini">Connect Media Player</span>
        {/if}
      </div>

      <!-- Minimal Re-used <Clock>, Bouncing Equalizer, Notification Indicator & Battery Indicator -->
      <div class="minimal-right-meta">
        {#if $notifData.length > 0}
          <button
            class="mini-notif-indicator {isNewNotif ? 'pulse-new' : ''}"
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
        {#if $homeWidgetsVisibilityStore.battery && hasBattery}
          <BatteryWidget />
        {/if}
        <div class="minimal-clock-box">
          <Clock expanded={false} />
        </div>
        {#if media.title || media.isPlaying}
          <DynamicEqualizer
            isPlaying={media.isPlaying}
            audioPeak={$audioPeakStore}
            variant="minimal"
            barCount={4}
          />
        {/if}
      </div>

      <!-- Compact Minimal Progress Track with Ghost Background & Capsule Laser Fill -->
      <div class="mini-progress-track {isDraggingSeek ? 'is-dragging' : ''}">
        <div class="mini-progress-ghost"></div>
        <div
          class="mini-progress-fill"
          style="width: {progressPercent}%;"
        ></div>
      </div>
    </div>
  {:else}
    <!-- Full-Card Background Image / Procedural Texture -->
    <div class="card-bg-wrapper" style="--card-tint: {cardTintColor};">
      {#if displayAlbumArt && !imgLoadFailed}
        <img
          src={displayAlbumArt}
          alt={media.title}
          class="card-bg-image"
          class:is-upscaled={isArtUpscaled}
          draggable="false"
          decoding="async"
          onerror={handleImgError}
        />
      {:else}
        <div
          class="card-bg-texture"
          style="background: {trackTexture.background};"
        >
          <div class="texture-grain-overlay"></div>
          <div class="texture-wave-overlay"></div>
        </div>
      {/if}
      <div class="card-bg-overlay"></div>
      <!-- Bottom legibility scrim: fades from --c-scrim to transparent -->
      <div class="card-bg-scrim"></div>
    </div>

    <!-- Expanded Player Content -->
    <div class="expanded-card">
      <!-- Top Bar: Glass Header Backdrop with Media Player title, Center Clock, App Badge -->
      <div class="top-meta-bar">
        <div class="app-brand">
          <span class="app-name-text">Media Player</span>
        </div>

        <div class="top-clock-wrapper">
          <Clock expanded={false} />
        </div>

        <button
          class="device-badge clickable"
          use:islandTooltip={{ text: `Switch to ${appBrand.name}`, position: 'bottom' }}
          aria-label="Switch to {appBrand.name}"
          onclick={handleFocusApp}
        >
          <span>{appBrand.name}</span>
        </button>
      </div>

      {#if isSessionMenuOpen && media.sessions && media.sessions.length > 1}
        <!-- Dedicated Media Sources View Mode -->
        <div class="session-drawer-panel">
          <div class="session-drawer-header">
            <div class="session-header-title-group">
              <span class="session-header-title">Media Sources</span>
              <span class="session-count-tag">{media.sessions.length}</span>
            </div>
            <button
              class="session-close-btn"
              use:islandTooltip={{ text: 'Close Sources', align: 'right' }}
              aria-label="Close Sources"
              onclick={(e) => {
                e.stopPropagation()
                isSessionMenuOpen = false
              }}
            >
              <X size={13} />
            </button>
          </div>

          <div class="session-drawer-list">
            {#each media.sessions as s (s.id)}
              {@const isSelected =
                s.id === media.activeSessionId ||
                (Boolean(media.title) && s.title === media.title) ||
                (media.sessions.length === 1 &&
                  s.sourceApp === media.sourceApp)}
              {@const sBrand = getSessionBrand(s.sourceApp, s.id)}
              <button
                class="session-menu-item"
                class:is-selected={isSelected}
                onclick={(e) => handleSelectSession(s.id, e)}
                role="menuitem"
              >
                <div class="session-item-icon">
                  {#if sBrand.icon}
                    <img
                      src={sBrand.icon}
                      alt={sBrand.name}
                      class="session-app-img"
                      draggable="false"
                      decoding="async"
                      onerror={(e) => {
                        const el = e.currentTarget as HTMLImageElement
                        el.style.display = 'none'
                      }}
                    />
                  {:else}
                    <Music2 size={13} class="session-item-fallback-icon" />
                  {/if}
                </div>
                <div class="session-item-info">
                  <div
                    class="session-item-title"
                    title={s.title || sBrand.name}
                  >
                    {s.title || sBrand.name}
                  </div>
                  <div
                    class="session-item-subtitle"
                    title={s.artist || sBrand.name}
                  >
                    {s.artist || sBrand.name}
                  </div>
                </div>
                <div class="session-item-status">
                  {#if isSelected}
                    <Check size={13} class="selected-check" />
                  {:else if s.isPlaying}
                    <div class="live-playing-dot" title="Playing"></div>
                  {/if}
                </div>
              </button>
            {/each}
          </div>
        </div>
      {:else}
        <!-- Middle Row: Left-aligned Artist (Line 1) + Track Title (Line 2) + Hero Play/Pause Button -->
        <div class="middle-section">
          <div class="track-details">
            <!-- Line 1: Artist / Album Name (Subtle Subtitle Style) -->
            <div class="artist-name">
              {media.artist || media.album || (media.title ? 'Unknown Artist' : 'Media Player')}
            </div>

            <!-- Line 2: Track Title (Prominent Headline Style with marquee support) -->
            <div
              class="track-title-wrapper"
              class:is-overflowing={isExpandedTitleOverflowing}
              style={isExpandedTitleOverflowing
                ? `--marquee-dist: ${expandedTitleDist}px; --marquee-duration: ${expandedTitleDuration}s;`
                : undefined}
            >
              <div
                bind:this={expandedTitleNode}
                class="track-title"
                class:is-marquee={isExpandedTitleOverflowing}
              >
                {media.title || 'Connect Media Player'}
              </div>
            </div>
          </div>

          <button
            class="hero-play-btn"
            use:islandTooltip={media.isPlaying ? 'Pause' : 'Play'}
            aria-label={media.isPlaying ? 'Pause' : 'Play'}
            onclick={(e) => handleControl('play_pause', e)}
          >
            {#if media.isPlaying}
              <svg
                viewBox="0 0 24 24"
                width="22"
                height="22"
                fill="currentColor"
              >
                <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" />
              </svg>
            {:else}
              <svg
                viewBox="0 0 24 24"
                width="22"
                height="22"
                fill="currentColor"
              >
                <path d="M8 5v14l11-7z" />
              </svg>
            {/if}
          </button>
        </div>

        <!-- Controls Row: Shuffle, Prev, Next, Repeat, Source on Left, Volume Slider on Right -->
        <div class="bottom-controls-row" class:is-playing={media.isPlaying}>
          <div class="nav-controls">
            {#if media.isShuffleEnabled !== false && media.isShuffleActive !== undefined}
              <button
                class="action-btn"
                class:is-active={media.isShuffleActive}
                disabled={media.isShuffleActive === undefined}
                use:islandTooltip={{
                  text: media.isShuffleActive === undefined
                    ? 'Shuffle: Not supported'
                    : media.isShuffleActive
                      ? 'Shuffle: On'
                      : 'Shuffle: Off',
                  align: 'left'
                }}
                aria-label="Shuffle"
                onclick={(e) => handleControl('shuffle', e)}
              >
                <Shuffle size={13} />
              </button>
            {/if}

            <button
              class="action-btn"
              use:islandTooltip={{ text: 'Previous Track', align: 'left' }}
              aria-label="Previous Track"
              onclick={(e) => handleControl('prev', e)}
            >
              <svg
                viewBox="0 0 24 24"
                width="14"
                height="14"
                fill="currentColor"
              >
                <path d="M6 6h2v12H6zm3.5 6l8.5 6V6z" />
              </svg>
            </button>

            <button
              class="action-btn"
              use:islandTooltip={'Next Track'}
              aria-label="Next Track"
              onclick={(e) => handleControl('next', e)}
            >
              <svg
                viewBox="0 0 24 24"
                width="14"
                height="14"
                fill="currentColor"
              >
                <path
                  d="M6 6h2v12H6zm3.5 6l8.5 6V6z"
                  transform="rotate(180 12 12)"
                />
              </svg>
            </button>

            {#if media.isRepeatEnabled !== false && media.autoRepeatMode !== undefined}
              <button
                class="action-btn"
                class:is-active={media.autoRepeatMode &&
                  media.autoRepeatMode > 0}
                disabled={media.autoRepeatMode === undefined}
                use:islandTooltip={{
                  text: media.autoRepeatMode === undefined
                    ? 'Repeat: Not supported'
                    : media.autoRepeatMode === 1
                      ? 'Repeat: One Track'
                      : media.autoRepeatMode === 2
                        ? 'Repeat: All Tracks'
                        : 'Repeat: Off'
                }}
                aria-label="Repeat"
                onclick={(e) => handleControl('repeat', e)}
              >
                {#if media.autoRepeatMode === 1}
                  <Repeat1 size={13} />
                {:else}
                  <Repeat size={13} />
                {/if}
              </button>
            {/if}

            {#if media.sessions && media.sessions.length > 1}
              <button
                class="action-btn session-btn"
                class:is-active={isSessionMenuOpen}
                use:islandTooltip={{ text: `Media Sources (${media.sessions.length})` }}
                aria-label="Media Sources"
                onclick={toggleSessionMenu}
              >
                <ListMusic size={14} />
                <span class="session-indicator-dot"></span>
              </button>
            {/if}
          </div>

          <!-- Volume Slider Control -->
          <div class="volume-control">
            <button
              class="vol-icon-btn"
              use:islandTooltip={{ text: media.isMuted ? 'Unmute' : 'Mute', align: 'right' }}
              aria-label={media.isMuted ? 'Unmute' : 'Mute'}
              onclick={(e) => handleControl('toggle_mute', e)}
            >
              {#if media.isMuted || currentVol === 0}
                <VolumeX size={12} />
              {:else if currentVol > 50}
                <Volume2 size={12} />
              {:else}
                <Volume1 size={12} />
              {/if}
            </button>
            <span class="vol-percentage-text">{Math.round(currentVol)}%</span>
            <div class="vol-slider-wrapper">
              <div class="vol-fill" style="width: {currentVol}%;"></div>
              <input
                type="range"
                min="0"
                max="100"
                value={currentVol}
                class="vol-input"
                onpointerdown={handleVolPointerDown}
                onpointerup={handleVolPointerUp}
                oninput={handleVolInput}
                onchange={handleVolChange}
              />
            </div>
          </div>
        </div>

        <!-- Bottom Audio Spectrum Waveform Scrub Bar with aligned Time Labels -->
        <div class="scrub-container">
          <span class="time-label">{formatTime(currentPos)}</span>
          <div
            class="waveform-scrub-wrapper"
            role="presentation"
            onpointermove={handleScrubPointerMove}
            onpointerleave={handleScrubPointerLeave}
          >
            {#if (isScrubHovering || isDraggingSeek) && media.durationMs > 0}
              <div
                class="scrub-preview-bubble"
                style="left: {isDraggingSeek
                  ? (seekVal / media.durationMs) * 100
                  : scrubHoverPercent}%;"
              >
                {formatTime(isDraggingSeek ? seekVal : scrubHoverTimeMs)}
              </div>
            {/if}
            <!-- Inactive State: Sleek Single Block Progress Bar -->
            <div
              class="single-block-progress {isDraggingSeek
                ? 'is-dragging'
                : ''}"
              class:is-active={media.isPlaying}
            >
              <div class="single-block-ghost"></div>
              <div
                class="single-block-fill"
                style="width: {progressPercent}%;"
              ></div>
            </div>

            <!-- Active State: Dynamic Equalizer Waveform -->
            <div
              class="spectrum-bars-container"
              class:is-active={media.isPlaying}
            >
              <DynamicEqualizer
                isPlaying={media.isPlaying}
                audioPeak={$audioPeakStore}
                variant="full"
                barCount={32}
                {progressPercent}
              />
            </div>
            <input
              type="range"
              min="0"
              max={media.durationMs > 0 ? media.durationMs : 100}
              value={currentPos}
              class="scrub-input"
              onpointerdown={handleSeekStart}
              oninput={handleSeekInput}
              onpointerup={handleSeekCommit}
              onchange={handleSeekCommit}
            />
          </div>
          <span class="time-label">{formatTime(media.durationMs)}</span>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style src="./styles/MediaWidget.css"></style>
