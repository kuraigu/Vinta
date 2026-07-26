<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 340, h: 48, r: 24 },
    expanded: { w: 420, h: 220, r: 28 },
    borderColor: 'rgba(255, 60, 100, 0.4)',
    glowColor: 'rgba(255, 60, 100, 0.35)'
  }
</script>

<script lang="ts">
  import { SendMediaControl } from '../../wailsjs/go/backend/App.js'
  import Clock from './Clock.svelte'
  import { globalMediaStore, notifData } from '../store'
  import { Volume2, VolumeX, Laptop, Music, Bell } from 'lucide-svelte'
  import { getContext } from 'svelte'
  import NotificationPage, { config as notifConfig } from '../pages/NotificationPage.svelte'

  const navigate = getContext<(page: unknown, pageConfig: IslandConfig, expand?: boolean) => void>('navigate')

  interface Props {
    expanded?: boolean
  }

  let { expanded = false }: Props = $props()

  let media = $derived($globalMediaStore)

  let isDraggingSeek = $state(false)
  let seekVal = $state(0)
  let isDraggingVol = $state(false)
  let volVal = $state(50)

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

  $effect(() => {
    if (!isDraggingSeek) {
      seekVal = media.positionMs || 0
    }
    if (!isDraggingVol && media.volume !== undefined) {
      volVal = media.volume
    }
  })

  function handleControl(action: string, e: MouseEvent) {
    e.stopPropagation()
    SendMediaControl(action)
  }

  function handleSeekInput(e: Event) {
    seekVal = Number((e.target as HTMLInputElement).value)
  }

  function handleSeekChange(e: Event) {
    const val = Number((e.target as HTMLInputElement).value)
    isDraggingSeek = false
    SendMediaControl(`seek:${val}`)
  }

  function handleVolInput(e: Event) {
    volVal = Number((e.target as HTMLInputElement).value)
    SendMediaControl(`set_volume:${volVal}`)
  }

  function handleVolChange(e: Event) {
    const val = Number((e.target as HTMLInputElement).value)
    isDraggingVol = false
    SendMediaControl(`set_volume:${val}`)
  }

  function formatTime(ms: number): string {
    if (!ms || isNaN(ms) || ms <= 0) return '0:00'
    const totalSecs = Math.floor(ms / 1000)
    const mins = Math.floor(totalSecs / 60)
    const secs = totalSecs % 60
    return `${mins}:${secs < 10 ? '0' : ''}${secs}`
  }

  let progressPercent = $derived(
    media.durationMs > 0
      ? Math.min(
          100,
          Math.max(
            0,
            ((isDraggingSeek ? seekVal : media.positionMs) / media.durationMs) * 100
          )
        )
      : 0
  )

  // 24 Bold Audio Spectrum Waveform Heights for Soundcloud/iOS Visualizer
  const waveformHeights = [
    30, 50, 85, 45, 65, 95, 75, 40, 70, 90, 100, 75, 55, 85, 95, 65, 45, 80, 90,
    60, 85, 50, 70, 40
  ]
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
            <img src={media.albumArt} alt={media.title} class="vinyl-art" />
          {:else}
            <div class="vinyl-placeholder">
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
        <span class="track-title-mini">{media.title || 'No Media Playing'}</span>
        <span class="artist-name-mini">{media.artist || 'Connect Media Player'}</span>
      </div>

      <!-- Minimal Re-used <Clock>, Bouncing Equalizer & Subtle Notification Indicator -->
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
            <span class="mini-notif-count">{$notifData.length > 9 ? '9+' : $notifData.length}</span>
          </button>
        {/if}
        <div class="minimal-clock-box">
          <Clock expanded={false} />
        </div>
        {#if media.isPlaying}
          <div class="equalizer">
            <span class="bar bar-1"></span>
            <span class="bar bar-2"></span>
            <span class="bar bar-3"></span>
          </div>
        {/if}
      </div>

      <!-- Compact Visible Audio Waveform Bar Track -->
      <div class="mini-wave-track">
        {#each waveformHeights as height, i (i)}
          {@const barPercent = (i / (waveformHeights.length - 1)) * 100}
          {@const isPlayed = barPercent <= progressPercent}
          <div
            class="mini-wave-bar {isPlayed
              ? 'played'
              : 'unplayed'} {media.isPlaying && isPlayed ? 'dancing' : ''}"
            style="height: {height}%; animation-delay: {i * 0.05}s;"
          ></div>
        {/each}
      </div>
    </div>
  {:else}
    <!-- Full-Card Background Image -->
    <div class="card-bg-wrapper">
      {#if media.albumArt}
        <img src={media.albumArt} alt={media.title} class="card-bg-image" />
      {/if}
      <div class="card-bg-overlay"></div>
    </div>

    <!-- Expanded Player Content -->
    <div class="expanded-card">
      <!-- Top Bar: Dynamic App Icon & App Name -->
      <div class="top-meta-bar">
        <div class="app-brand">
          {#if media.appIcon}
            <img src={media.appIcon} alt={media.appName} class="app-icon-img" />
          {:else}
            <Music size={16} class="brand-icon" />
          {/if}
          <span class="app-name-text">{media.appName || 'Media Player'}</span>
        </div>

        <div class="device-badge">
          <Laptop size={11} />
          <span>System Audio</span>
        </div>
      </div>

      <!-- Middle Row: Centered Track Title/Artist + Play/Pause Button -->
      <div class="middle-section">
        <div class="middle-left-spacer"></div>
        <div class="track-details">
          <div class="track-title">{media.title || 'No Media Playing'}</div>
          <div class="artist-name">
            {media.artist || 'Connect Media Player'}
          </div>
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
        <span class="time-label"
          >{formatTime(isDraggingSeek ? seekVal : media.positionMs)}</span
        >
        <div class="waveform-scrub-wrapper">
          <div class="spectrum-bars-container">
            {#each waveformHeights as height, i (i)}
              {@const barPercent = (i / (waveformHeights.length - 1)) * 100}
              {@const isPlayed = barPercent <= progressPercent}
              <div
                class="spectrum-bar {isPlayed
                  ? 'played'
                  : 'unplayed'} {media.isPlaying && isPlayed ? 'dancing' : ''}"
                style="height: {height}%; animation-delay: {i * 0.04}s;"
              ></div>
            {/each}
          </div>
          <input
            type="range"
            min="0"
            max={media.durationMs || 100}
            value={isDraggingSeek ? seekVal : media.positionMs}
            class="scrub-input"
            onpointerdown={() => (isDraggingSeek = true)}
            oninput={handleSeekInput}
            onchange={handleSeekChange}
          />
        </div>
        <span class="time-label">{formatTime(media.durationMs)}</span>
      </div>

      <!-- Bottom Controls Row: Prev, Next, Volume Slider -->
      <div class="bottom-controls-row">
        <div class="nav-controls">
          <button
            class="action-btn"
            title="Previous Track"
            onclick={(e) => handleControl('prev', e)}
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor">
              <path d="M6 6h2v12H6zm3.5 6l8.5 6V6z" />
            </svg>
          </button>

          <button
            class="action-btn"
            title="Next Track"
            onclick={(e) => handleControl('next', e)}
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor">
              <path d="M6 18l8.5-6L6 6v12zM16 6v12h2V6h-2z" />
            </svg>
          </button>
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

          <div class="vol-slider-wrapper">
            <div
              class="vol-fill"
              style="width: {isDraggingVol ? volVal : media.volume || 50}%"
            ></div>
            <input
              type="range"
              min="0"
              max="100"
              value={isDraggingVol ? volVal : media.volume || 50}
              class="vol-input"
              onpointerdown={() => (isDraggingVol = true)}
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

  /* Minimal Layout - Perfectly centered 3-column layout with safe pill-curve padding */
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

  /* Spinning Vinyl Disc */
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
    opacity: 0.6;
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

  /* Compact Visible Audio Waveform Bar Track */
  .mini-wave-track {
    position: absolute;
    bottom: 2px;
    left: 20px;
    right: 20px;
    height: 7px;
    display: flex;
    align-items: flex-end;
    gap: 2.5px;
    pointer-events: none;
  }

  .mini-wave-bar {
    flex: 1;
    border-radius: 1px;
    transition: background 0.2s ease;
  }

  .mini-wave-bar.played {
    background: linear-gradient(180deg, #ff3c64 0%, #1db954 100%);
    box-shadow: 0 0 4px rgba(255, 60, 100, 0.8);
  }

  .mini-wave-bar.unplayed {
    background: rgba(255, 255, 255, 0.25);
  }

  .mini-wave-bar.dancing {
    animation: waveBounce 0.8s infinite ease-in-out alternate;
  }

  /* Equalizer Animation */
  .equalizer {
    display: flex;
    align-items: flex-end;
    gap: 2px;
    height: 12px;
  }

  .bar {
    width: 2.5px;
    background-color: #1db954;
    border-radius: 1px;
    animation: bounce 1s infinite ease-in-out alternate;
  }

  .bar-1 {
    height: 60%;
    animation-delay: 0.1s;
  }
  .bar-2 {
    height: 100%;
    animation-delay: 0.3s;
  }
  .bar-3 {
    height: 40%;
    animation-delay: 0.2s;
  }

  @keyframes bounce {
    0% {
      transform: scaleY(0.3);
    }
    100% {
      transform: scaleY(1);
    }
  }

  /* Full-Card Background Styling - Theme Aware Cover Card */
  .card-bg-wrapper {
    position: absolute;
    inset: 0;
    z-index: 0;
    pointer-events: none;
    overflow: hidden;
    border-radius: 16px;
    border: 1px solid var(--c-border, rgba(255, 255, 255, 0.12));
  }

  .card-bg-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
    opacity: 0.85;
    filter: brightness(0.65);
  }

  .card-bg-overlay {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      180deg,
      rgba(0, 0, 0, 0.35) 0%,
      rgba(0, 0, 0, 0.65) 60%,
      rgba(0, 0, 0, 0.88) 100%
    );
  }

  :global([data-theme='light']) .card-bg-overlay {
    background: linear-gradient(
      180deg,
      rgba(0, 0, 0, 0.25) 0%,
      rgba(0, 0, 0, 0.55) 60%,
      rgba(0, 0, 0, 0.78) 100%
    );
  }

  /* Expanded Card Layout */
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

  /* Top Bar */
  .top-meta-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .app-brand {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .app-icon-img {
    width: 16px;
    height: 16px;
    border-radius: 3px;
    object-fit: contain;
  }

  .app-name-text {
    font-size: 0.72rem;
    font-weight: 600;
    opacity: 0.9;
  }

  .device-badge {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 2px 7px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.18);
    backdrop-filter: blur(8px);
    font-size: 0.62rem;
    font-weight: 600;
  }

  /* Middle Section - Centered Song Info */
  .middle-section {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-top: 2px;
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
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.38), 0 0 14px rgba(255, 255, 255, 0.25);
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    flex-shrink: 0;
  }

  .hero-play-btn:hover {
    transform: scale(1.08);
    background: #ffffff;
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.48), 0 0 20px rgba(255, 255, 255, 0.4);
  }

  /* Soundcloud / iOS Spectrum Audio Waveform Scrub Bar */
  .scrub-container {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 2px;
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
    inset: 0;
    display: flex;
    align-items: flex-end;
    gap: 3px;
    pointer-events: none;
  }

  .spectrum-bar {
    flex: 1;
    border-radius: 2px;
    transition: background 0.15s ease;
  }

  .spectrum-bar.played {
    background: linear-gradient(180deg, #ff3c64 0%, #1db954 100%);
    box-shadow: 0 0 6px rgba(255, 60, 100, 0.8);
  }

  .spectrum-bar.unplayed {
    background: rgba(255, 255, 255, 0.2);
  }

  .spectrum-bar.dancing {
    animation: waveBounce 0.8s infinite ease-in-out alternate;
  }

  @keyframes waveBounce {
    0% {
      transform: scaleY(0.45);
    }
    100% {
      transform: scaleY(1.2);
    }
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

  /* Bottom Controls Row */
  .bottom-controls-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
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

  /* Volume Slider Control */
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
