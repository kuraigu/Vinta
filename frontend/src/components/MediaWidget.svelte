<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 368, h: 44, r: 22 },
    expanded: { w: 440, h: 235, r: 28 },
    borderColor: 'rgba(255, 60, 100, 0.4)',
    glowColor: 'rgba(255, 60, 100, 0.35)'
  }
</script>

<script lang="ts">
  import { SendMediaControl } from '../../wailsjs/go/backend/App.js'
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
    Repeat1
  } from 'lucide-svelte'
  import { getContext } from 'svelte'
  import NotificationPage, {
    config as notifConfig
  } from '../pages/NotificationPage.svelte'
  import { resolveAppBrand } from '../appIcons'

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

  function computeAppBrand() {
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
  let seekLockTimer: ReturnType<typeof setTimeout> | null = null

  let playLockTimer: ReturnType<typeof setTimeout> | null = null

  // Sync with incoming media store updates when track changes, play state changes, or external seek occurs
  let prevTitle = $state('')
  let prevArtist = $state('')
  let prevBackendPos = $state(0)

  function handleMediaSyncEffect() {
    const curTitle = media.title || ''
    const curArtist = media.artist || ''
    const backendPos = media.positionMs || 0

    const isTrackChanged = curTitle !== prevTitle || curArtist !== prevArtist
    const isExternalSeek = Math.abs(backendPos - prevBackendPos) > 3000

    if (!isDraggingSeek && !isSeekingLocked) {
      if (isTrackChanged || !media.isPlaying || isExternalSeek) {
        localPosMs = backendPos
        seekVal = backendPos
      } else if (Math.abs(localPosMs - backendPos) > 1500) {
        localPosMs = backendPos
        seekVal = backendPos
      }
    }

    prevTitle = curTitle
    prevArtist = curArtist
    prevBackendPos = backendPos

    if (!isDraggingVol && media.volume !== undefined) {
      volVal = media.volume
    }
  }
  $effect(handleMediaSyncEffect)

  let lastSeekTimestamp = $state(0)

  function handleImgError() {
    globalMediaStore.update((m) => ({ ...m, albumArt: '' }))
  }

  // Smooth local timeline progression between backend SMTC updates
  function handleTimelineProgressEffect() {
    if (!media.isPlaying || isDraggingSeek) return

    let lastTick = performance.now()
    let rafId: number

    const tick = () => {
      const now = performance.now()
      const elapsed = now - lastTick
      lastTick = now

      const isSettlingAfterSeek = now - lastSeekTimestamp < 500

      if (
        !isDraggingSeek &&
        !isSettlingAfterSeek &&
        !isSeekingLocked &&
        media.durationMs > 0 &&
        media.isPlaying
      ) {
        localPosMs = Math.min(media.durationMs, localPosMs + elapsed)
        seekVal = localPosMs
      }

      if (media.isPlaying && !isDraggingSeek) {
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
      isSeekingLocked = true
      lastSeekTimestamp = performance.now()
      if (seekLockTimer) clearTimeout(seekLockTimer)
      seekLockTimer = setTimeout(() => {
        isSeekingLocked = false
        seekLockTimer = null
      }, 1800)
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
      globalMediaStore.update((m) => ({
        ...m,
        isMuted: !m.isMuted
      }))
      SendMediaControl(`toggle_mute:${media.sourceApp || ''}`)
      return
    }
    SendMediaControl(action)
  }

  function handleSeekStart(e: PointerEvent) {
    if (seekLockTimer) {
      clearTimeout(seekLockTimer)
      seekLockTimer = null
    }
    isDraggingSeek = true
    isSeekingLocked = true
    const val = Number((e.target as HTMLInputElement).value)
    seekVal = val
    localPosMs = val
  }

  function handleSeekInput(e: Event) {
    isDraggingSeek = true
    isSeekingLocked = true
    if (seekLockTimer) {
      clearTimeout(seekLockTimer)
      seekLockTimer = null
    }
    const val = Number((e.target as HTMLInputElement).value)
    seekVal = val
    localPosMs = val
  }

  function handleSeekCommit(e: Event) {
    const val = Number((e.target as HTMLInputElement).value)
    seekVal = val
    localPosMs = val
    isDraggingSeek = false
    isSeekingLocked = true
    lastSeekTimestamp = performance.now()
    SendMediaControl(`seek:${val}`)

    if (seekLockTimer) clearTimeout(seekLockTimer)
    seekLockTimer = setTimeout(() => {
      isSeekingLocked = false
      seekLockTimer = null
    }, 1800)
  }

  let volThrottleTimer: ReturnType<typeof setTimeout> | null = null
  function handleVolInput(e: Event) {
    const val = Number((e.target as HTMLInputElement).value)
    volVal = val
    isDraggingVol = true
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
    SendMediaControl(`set_volume:${val}:${media.sourceApp || ''}`)
    globalMediaStore.update((m) => ({ ...m, volume: val }))
    isDraggingVol = false
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

  let currentVol = $derived(isDraggingVol ? volVal : (media.volume ?? 50))
</script>

<div class="media-container {expanded ? 'expanded' : 'minimal'}">
  {#if !expanded}
    <!-- Compact Minimal View with Spinning Vinyl Disc, Track Info, Clock, and Audio Wave Bars -->
    <div class="minimal-layout">
      <!-- Spinning Vinyl Disc Album Art -->
      <div
        class="vinyl-disc {media.isPlaying ? 'spinning' : 'paused'}"
        title={media.title}
      >
        <div class="vinyl-groove">
          {#if media.albumArt}
            <img
              src={media.albumArt}
              alt={media.title}
              class="vinyl-art"
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
        <span class="track-title-mini">{media.title || 'No Media Playing'}</span
        >
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
            color="var(--c-primary, #38bdf8)"
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
        <div class="mini-progress-head" style="left: {progressPercent}%;"></div>
      </div>
    </div>
  {:else}
    <!-- Full-Card Background Image / Procedural Texture -->
    <div class="card-bg-wrapper">
      {#if media.albumArt}
        <img
          src={media.albumArt}
          alt={media.title}
          class="card-bg-image"
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

        <div class="device-badge" title="Source: {appBrand.name}">
          <span>{appBrand.name}</span>
        </div>
      </div>

      <!-- Middle Row: Left-aligned Track Title/Artist + Hero Play/Pause Button -->
      <div class="middle-section">
        <div class="track-details">
          <div class="track-title" title={media.title || 'No Media Playing'}>
            {media.title || 'No Media Playing'}
          </div>
          {#if media.artist}
            <div class="artist-name" title={media.artist}>{media.artist}</div>
          {:else if !media.title}
            <div class="artist-name">Connect Media Player</div>
          {/if}
        </div>

        <button
          class="hero-play-btn"
          title={media.isPlaying ? 'Pause' : 'Play'}
          onclick={(e) => handleControl('play_pause', e)}
        >
          {#if media.isPlaying}
            <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
              <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" />
            </svg>
          {:else}
            <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
              <path d="M8 5v14l11-7z" />
            </svg>
          {/if}
        </button>
      </div>

      <!-- Controls Row: Shuffle, Prev, Next, Repeat on Left, Volume Slider on Right -->
      <div class="bottom-controls-row">
        <div class="nav-controls">
          {#if media.isShuffleEnabled !== false && media.isShuffleActive !== undefined}
            <button
              class="action-btn"
              class:is-active={media.isShuffleActive}
              disabled={media.isShuffleActive === undefined}
              title={media.isShuffleActive === undefined
                ? 'Shuffle: Not supported by current playback'
                : media.isShuffleActive
                  ? 'Shuffle: On'
                  : 'Shuffle: Off'}
              onclick={(e) => handleControl('shuffle', e)}
            >
              <Shuffle size={13} />
            </button>
          {/if}

          <button
            class="action-btn"
            title="Previous Track"
            onclick={(e) => handleControl('prev', e)}
          >
            <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
              <path d="M6 6h2v12H6zm3.5 6l8.5 6V6z" />
            </svg>
          </button>

          <button
            class="action-btn"
            title="Next Track"
            onclick={(e) => handleControl('next', e)}
          >
            <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
              <path
                d="M6 6h2v12H6zm3.5 6l8.5 6V6z"
                transform="rotate(180 12 12)"
              />
            </svg>
          </button>

          {#if media.isRepeatEnabled !== false && media.autoRepeatMode !== undefined}
            <button
              class="action-btn"
              class:is-active={media.autoRepeatMode && media.autoRepeatMode > 0}
              disabled={media.autoRepeatMode === undefined}
              title={media.autoRepeatMode === undefined
                ? 'Repeat: Not supported by current playback'
                : media.autoRepeatMode === 1
                  ? 'Repeat: One Track'
                  : media.autoRepeatMode === 2
                    ? 'Repeat: All Tracks'
                    : 'Repeat: Off'}
              onclick={(e) => handleControl('repeat', e)}
            >
              {#if media.autoRepeatMode === 1}
                <Repeat1 size={13} />
              {:else}
                <Repeat size={13} />
              {/if}
            </button>
          {/if}
        </div>

        <!-- Volume Slider Control -->
        <div class="volume-control">
          <button
            class="vol-icon-btn"
            title={media.isMuted ? 'Unmute' : 'Mute'}
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
              oninput={handleVolInput}
              onchange={handleVolChange}
            />
          </div>
        </div>
      </div>

      <!-- Bottom Audio Spectrum Waveform Scrub Bar with aligned Time Labels -->
      <div class="scrub-container">
        <span class="time-label">{formatTime(currentPos)}</span>
        <div class="waveform-scrub-wrapper">
          <div class="spectrum-bars-container">
            <DynamicEqualizer
              isPlaying={media.isPlaying}
              audioPeak={$audioPeakStore}
              variant="full"
              barCount={32}
              {progressPercent}
            />
          </div>
          <div
            class="whole-capsule-track {isDraggingSeek ? 'is-dragging' : ''}"
          >
            <div class="capsule-track-ghost"></div>
            <div
              class="capsule-track-fill"
              style="width: {progressPercent}%;"
            ></div>
            <div
              class="capsule-track-head"
              style="left: {progressPercent}%;"
            ></div>
          </div>
          <input
            type="range"
            min="0"
            max={media.durationMs > 0 ? media.durationMs : 100}
            value={currentPos}
            class="scrub-input"
            onpointerdown={handleSeekStart}
            oninput={handleSeekInput}
            onchange={handleSeekCommit}
          />
        </div>
        <span class="time-label">{formatTime(media.durationMs)}</span>
      </div>
    </div>
  {/if}
</div>

<style src="./styles/MediaWidget.css"></style>
