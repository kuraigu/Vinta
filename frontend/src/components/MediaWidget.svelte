<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 340, h: 44, r: 22 },
    expanded: { w: 440, h: 270, r: 28 },
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
    volVal = Number((e.target as HTMLInputElement).value)
    isDraggingVol = true
    if (!volThrottleTimer) {
      SendMediaControl(`set_volume:${volVal}`)
      volThrottleTimer = setTimeout(() => {
        volThrottleTimer = null
      }, 75)
    }
  }

  function handleVolChange(e: Event) {
    const val = Number((e.target as HTMLInputElement).value)
    if (volThrottleTimer) clearTimeout(volThrottleTimer)
    volThrottleTimer = null
    SendMediaControl(`set_volume:${val}`)
    if (media) media.volume = val
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

      <!-- Minimal Re-used <Clock>, Bouncing Equalizer & Subtle Notification Indicator / Battery Indicator -->
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
        {:else if $homeWidgetsVisibilityStore.battery && hasBattery}
          <BatteryWidget />
        {/if}
        <div class="minimal-clock-box">
          <Clock expanded={false} />
        </div>
        {#if media.isPlaying}
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

      <!-- Middle Row: Centered Track Title/Artist + Hero Play/Pause Button -->
      <div class="middle-section">
        <div class="middle-left-spacer"></div>
        <div class="track-details">
          <div class="track-title">{media.title || 'No Media Playing'}</div>
          {#if media.artist}
            <div class="artist-name">{media.artist}</div>
          {:else if !media.title}
            <div class="artist-name">Connect Media Player</div>
          {/if}
          {#if media.album}
            <div class="album-name">{media.album}</div>
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

      <!-- Soundcloud / iOS Dynamic Audio Spectrum Waveform Scrub Bar -->
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

      <!-- Bottom Controls Row: Shuffle, Prev, Next, Repeat, Volume Slider -->
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
            onclick={(e) => handleControl('mute', e)}
          >
            {#if media.isMuted}
              <VolumeX size={14} />
            {:else}
              <Volume2 size={14} />
            {/if}
          </button>

          <span class="vol-percentage-text"
            >{Math.round(isDraggingVol ? volVal : (media.volume ?? 50))}%</span
          >

          <div class="vol-slider-wrapper">
            <div
              class="vol-fill"
              style="width: {isDraggingVol ? volVal : (media.volume ?? 50)}%"
            ></div>
            <input
              type="range"
              min="0"
              max="100"
              value={isDraggingVol ? volVal : (media.volume ?? 50)}
              class="vol-input"
              oninput={handleVolInput}
              onchange={handleVolChange}
            />
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .media-container {
    width: 100%;
    height: 100%;
    box-sizing: border-box;
    color: #ffffff;
    user-select: none;
    position: relative;
    overflow: hidden;
    pointer-events: none;
  }

  .minimal-layout {
    pointer-events: auto;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 0 18px;
    height: 100%;
    width: 100%;
    max-width: 100%;
    position: relative;
    box-sizing: border-box;
  }

  .vinyl-disc {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: #111;
    box-shadow:
      0 2px 8px rgba(0, 0, 0, 0.5),
      inset 0 0 0 2px rgba(255, 255, 255, 0.15);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    position: relative;
    overflow: hidden;
  }

  .vinyl-disc.spinning {
    animation: spinVinyl 6s linear infinite;
  }

  .vinyl-disc.paused {
    animation: spinVinyl 6s linear infinite;
    animation-play-state: paused;
  }

  @keyframes spinVinyl {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  .vinyl-groove {
    width: 100%;
    height: 100%;
    border-radius: 50%;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
  }

  .vinyl-art {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 50%;
  }

  .vinyl-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
    overflow: hidden;
  }

  .vinyl-noise-ring {
    position: absolute;
    inset: 0;
    opacity: 0.25;
    background: radial-gradient(
      circle,
      transparent 30%,
      rgba(255, 255, 255, 0.2) 50%,
      transparent 70%
    );
    pointer-events: none;
  }

  .vinyl-hole {
    position: absolute;
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: #0d0d12;
    border: 1px solid rgba(255, 255, 255, 0.4);
    box-shadow: 0 0 2px rgba(0, 0, 0, 0.8);
    pointer-events: none;
  }

  .track-info-minimal {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    text-align: left;
    overflow: hidden;
    flex: 1 1 0;
    min-width: 0;
  }

  .track-title-mini {
    font-size: 0.73rem;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    width: 100%;
  }

  .artist-name-mini {
    font-size: 0.62rem;
    opacity: 0.7;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    width: 100%;
  }

  .minimal-right-meta {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
  }

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

  .minimal-clock-box {
    font-size: 0.68rem;
    font-weight: 600;
    opacity: 0.75;
    white-space: nowrap;
  }

  .mini-progress-track {
    position: absolute;
    bottom: 3px;
    left: 45px;
    right: 45px;
    height: 4px;
    border-radius: 9999px;
    overflow: visible;
    pointer-events: none;
  }

  .mini-progress-track.is-dragging .mini-progress-fill,
  .mini-progress-track.is-dragging .mini-progress-head {
    transition: none !important;
  }

  .mini-progress-ghost {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    background: rgba(255, 255, 255, 0.16);
    border-radius: 9999px;
  }

  .mini-progress-fill {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    height: 100%;
    background: linear-gradient(90deg, #00f0ff 0%, #a855f7 50%, #ff007f 100%);
    box-shadow:
      0 0 8px #00f0ff,
      0 0 16px #ff007f,
      inset 0 0 4px #ffffff;
    border-radius: 9999px;
    transition: width 0.1s linear;
  }

  .mini-progress-head {
    position: absolute;
    top: 50%;
    transform: translate(-50%, -50%);
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #ffffff;
    box-shadow:
      0 0 10px #00f0ff,
      0 0 18px #ff007f,
      0 0 26px #a855f7;
    z-index: 10;
    pointer-events: none;
    transition: left 0.1s linear;
  }

  .mini-progress-head::after {
    content: '';
    position: absolute;
    inset: -3px;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      rgba(0, 240, 255, 0.6) 0%,
      rgba(255, 0, 127, 0.3) 60%,
      transparent 100%
    );
    box-shadow:
      0 0 10px rgba(0, 240, 255, 0.7),
      0 0 16px rgba(255, 0, 127, 0.5);
    animation: subtleHeadPulse 1.8s infinite ease-in-out alternate;
    pointer-events: none;
  }

  @keyframes subtleHeadPulse {
    0% {
      transform: scale(0.85);
      opacity: 0.5;
    }
    100% {
      transform: scale(1.4);
      opacity: 0.95;
    }
  }

  .card-bg-wrapper {
    position: absolute;
    inset: 0;
    z-index: 0;
    pointer-events: none;
    overflow: hidden;
  }

  .card-bg-wrapper::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: inherit;
    box-shadow: inset 0 0 0 1px var(--c-border, rgba(255, 255, 255, 0.12));
    pointer-events: none;
    z-index: 1;
  }

  .card-bg-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
    opacity: 0.65;
    filter: brightness(0.42);
    position: relative;
    z-index: 0;
  }

  .card-bg-overlay {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      180deg,
      var(--c-scrim, rgba(0, 0, 0, 0.35)) 0%,
      transparent 50%
    );
    opacity: 0.55;
  }

  /* Bottom legibility scrim — themed, anchored behind controls & metadata */
  .card-bg-scrim {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    height: 110px;
    background: linear-gradient(
      to top,
      var(--c-scrim, rgba(0, 0, 0, 0.72)) 0%,
      transparent 100%
    );
    pointer-events: none;
  }

  .expanded-card {
    position: relative;
    z-index: 2;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    padding: 10px 14px 8px 14px;
    height: 100%;
    width: 100%;
    box-sizing: border-box;
    pointer-events: auto;
  }

  .top-meta-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    margin: -10px -14px 6px -14px;
    border-radius: 20px 20px 0 0;
    background: rgba(50, 50, 50, 0.06);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.18);
  }

  .top-clock-wrapper {
    font-size: 0.85rem;
    font-weight: 700;
    letter-spacing: -0.02em;
    color: var(--c-text-primary, #ffffff);
    opacity: 0.95;
    font-feature-settings: 'tnum';
    text-shadow: 0 1px 4px rgba(0, 0, 0, 0.5);
  }

  .app-brand {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .app-name-text {
    font-size: 0.72rem;
    font-weight: 600;
    opacity: 0.9;
  }

  .card-bg-texture {
    width: 100%;
    height: 100%;
    position: absolute;
    inset: 0;
    opacity: 0.85;
    filter: brightness(0.65);
    z-index: 0;
    transition: background 0.4s ease;
    overflow: hidden;
  }

  .texture-grain-overlay {
    position: absolute;
    inset: 0;
    opacity: 0.18;
    mix-blend-mode: overlay;
    background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)'/%3E%3C/svg%3E");
    pointer-events: none;
  }

  .texture-wave-overlay {
    position: absolute;
    inset: 0;
    opacity: 0.12;
    mix-blend-mode: soft-light;
    background: repeating-radial-gradient(
      circle at 30% 40%,
      rgba(255, 255, 255, 0.15) 0,
      rgba(255, 255, 255, 0.15) 2px,
      transparent 3px,
      transparent 8px
    );
    pointer-events: none;
  }

  .device-badge {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 2px 7px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.14);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    font-size: 0.62rem;
    font-weight: 600;
    line-height: 1;
    max-width: 125px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  }

  .device-badge span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .middle-section {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-top: 4px;
    width: 100%;
  }

  .middle-left-spacer {
    width: 44px;
    flex-shrink: 0;
  }

  .track-details {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    overflow: hidden;
    flex: 1;
    min-width: 0;
  }

  .track-title {
    font-size: 1.05rem;
    font-weight: 800;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    text-shadow: 0 2px 6px rgba(0, 0, 0, 0.6);
    text-align: center;
    width: 100%;
  }

  .artist-name {
    font-size: 0.82rem;
    font-weight: 500;
    opacity: 0.85;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    text-align: center;
    width: 100%;
  }

  .album-name {
    font-size: 0.7rem;
    opacity: 0.55;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    text-align: center;
    width: 100%;
  }

  .hero-play-btn {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.92);
    color: #0d0d14;
    border: 1px solid rgba(255, 255, 255, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    box-shadow:
      0 4px 16px rgba(0, 0, 0, 0.38),
      0 0 14px rgba(255, 255, 255, 0.25);
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    flex-shrink: 0;
  }

  .hero-play-btn:hover {
    transform: scale(1.08);
    background: #ffffff;
    box-shadow:
      0 6px 20px rgba(0, 0, 0, 0.48),
      0 0 20px rgba(255, 255, 255, 0.4);
  }

  .scrub-container {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 8px;
    margin-bottom: 4px;
  }

  .time-label {
    font-size: 0.63rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    opacity: 0.8;
    min-width: 28px;
    text-align: center;
  }

  .waveform-scrub-wrapper {
    position: relative;
    flex: 1;
    height: 22px;
    display: flex;
    align-items: flex-end;
  }

  .spectrum-bars-container {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 8px;
    pointer-events: none;
  }

  .whole-capsule-track {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 4px;
    border-radius: 9999px;
    overflow: visible;
    pointer-events: none;
  }

  .whole-capsule-track.is-dragging .capsule-track-fill,
  .whole-capsule-track.is-dragging .capsule-track-head {
    transition: none !important;
  }

  .capsule-track-ghost {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    background: rgba(255, 255, 255, 0.16);
    border-radius: 9999px;
  }

  .capsule-track-fill {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    height: 100%;
    background: linear-gradient(90deg, #00f0ff 0%, #a855f7 50%, #ff007f 100%);
    box-shadow:
      0 0 10px #00f0ff,
      0 0 20px #ff007f,
      inset 0 0 4px #ffffff;
    border-radius: 9999px;
  }

  .capsule-track-head {
    position: absolute;
    top: 50%;
    transform: translate(-50%, -50%);
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #ffffff;
    box-shadow:
      0 0 10px #00f0ff,
      0 0 18px #ff007f,
      0 0 26px #a855f7;
    z-index: 10;
    pointer-events: none;
  }

  .scrub-input {
    position: absolute;
    width: 100%;
    height: 100%;
    opacity: 0;
    cursor: pointer;
    margin: 0;
    z-index: 5;
  }

  .bottom-controls-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-top: 6px;
  }

  .nav-controls {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .action-btn {
    background: rgba(255, 255, 255, 0.15);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: #ffffff;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    backdrop-filter: blur(8px);
    transition: all 0.15s ease;
  }

  .action-btn:hover {
    background: rgba(255, 255, 255, 0.3);
    transform: scale(1.08);
  }

  .action-btn:disabled {
    opacity: 0.35;
    cursor: not-allowed;
    pointer-events: none;
    transform: none !important;
  }

  .action-btn.is-active {
    background: var(--c-media-btn-active-bg, #ffffff);
    border-color: var(--c-media-btn-active-border, rgba(255, 255, 255, 0.6));
    color: var(--c-media-btn-active-text, #000000);
    box-shadow: var(
      --c-media-btn-active-glow,
      0 0 10px rgba(255, 255, 255, 0.6)
    );
  }

  .volume-control {
    display: flex;
    align-items: center;
    gap: 6px;
    background: rgba(0, 0, 0, 0.25);
    padding: 3px 8px;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .vol-icon-btn {
    background: transparent;
    border: none;
    color: #ffffff;
    display: flex;
    align-items: center;
    cursor: pointer;
    padding: 0;
    opacity: 0.85;
  }

  .vol-icon-btn:hover {
    opacity: 1;
  }

  .vol-percentage-text {
    font-size: 0.62rem;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    color: var(--c-text-primary, #ffffff);
    opacity: 0.85;
    min-width: 24px;
    text-align: right;
  }

  .vol-slider-wrapper {
    position: relative;
    width: 70px;
    height: 4px;
    border-radius: 2px;
    background: rgba(255, 255, 255, 0.2);
    overflow: hidden;
    display: flex;
    align-items: center;
  }

  .vol-fill {
    position: absolute;
    left: 0;
    top: 0;
    height: 100%;
    background: linear-gradient(90deg, var(--c-primary, #6366f1), #a855f7);
    box-shadow: 0 0 6px var(--c-primary, rgba(99, 102, 241, 0.6));
    border-radius: 2px;
    pointer-events: none;
  }

  .vol-input {
    position: absolute;
    width: 100%;
    height: 100%;
    opacity: 0;
    cursor: pointer;
    margin: 0;
  }
</style>
