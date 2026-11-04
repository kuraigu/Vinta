<script lang="ts">
  import { onMount } from 'svelte'
  import { SvelteMap } from 'svelte/reactivity'
  import {
    Volume2,
    VolumeX,
    Volume1,
    Speaker,
    SlidersHorizontal,
    Waves,
    CloudRain,
    Radio,
    Play,
    Pause,
    Check
  } from 'lucide-svelte'
  import DrawerPanel from './DrawerPanel.svelte'
  import ScrollView from './ScrollView.svelte'
  import { globalMediaStore } from '../store'
  import { islandTooltip } from '../services/islandTooltip'
  import { resolveAppBrand } from '../services/iconService'
  import {
    audioSessionsStore,
    audioEndpointsStore,
    masterVolumeStore,
    masterMuteStore,
    setAppVolume,
    toggleAppMute,
    switchDefaultEndpoint,
    setMasterVolume,
    toggleMasterMute,
    startMixerPolling
  } from '../stores/audioMixer'
  import {
    soundscapeStore,
    playPreset,
    setSoundscapeVolume,
    toggleSoundscape,
    type SoundscapePreset
  } from '../services/soundscapeService'

  interface Props {
    onClose: () => void
  }

  let { onClose }: Props = $props()

  type MixerTab = 'apps' | 'ambience' | 'devices'
  let activeTab = $state<MixerTab>('apps')

  let media = $derived($globalMediaStore)
  let sessions = $derived($audioSessionsStore)
  let endpoints = $derived($audioEndpointsStore)
  let masterVol = $derived($masterVolumeStore)
  let masterMuted = $derived($masterMuteStore)
  let soundscape = $derived($soundscapeStore)

  let displaySessions = $derived.by(() => {
    const list = sessions || []
    const map = new SvelteMap<number, (typeof list)[0]>()
    for (const s of list) {
      if (!map.has(s.pid)) {
        map.set(s.pid, s)
      } else {
        const existing = map.get(s.pid)!
        if (s.peak > existing.peak) {
          map.set(s.pid, s)
        }
      }
    }
    return Array.from(map.values())
  })

  function syncMasterAudioEffect() {
    if (media.volume !== undefined && $masterVolumeStore === 100) {
      $masterVolumeStore = media.volume
    }
    if (media.isMuted !== undefined) {
      $masterMuteStore = media.isMuted
    }
  }
  $effect(syncMasterAudioEffect)

  onMount(() => {
    const stopPolling = startMixerPolling()
    return () => {
      stopPolling()
    }
  })

  function handleMasterVolInput(e: Event) {
    const val = Number((e.target as HTMLInputElement).value)
    void setMasterVolume(val)
  }

  function handleAppVolInput(pid: number, e: Event) {
    const val = Number((e.target as HTMLInputElement).value)
    void setAppVolume(pid, val)
  }

  function handleSoundscapeVolInput(e: Event) {
    const val = Number((e.target as HTMLInputElement).value)
    setSoundscapeVolume(val)
  }

  function handleSelectPreset(preset: SoundscapePreset) {
    if (soundscape.preset === preset && soundscape.isPlaying) {
      playPreset('none')
    } else {
      playPreset(preset)
    }
  }
</script>

<DrawerPanel
  title="Audio Mixer"
  tag={activeTab === 'apps' ? displaySessions.length : undefined}
  closeTooltip="Close Mixer"
  variant="embedded"
  {onClose}
>
  <div class="mixer-tab-row">
    <button
      class="mixer-tab-btn"
      class:is-active={activeTab === 'apps'}
      onclick={() => (activeTab = 'apps')}
      use:islandTooltip={{ text: 'Application Audio Volumes' }}
    >
      <SlidersHorizontal size={12} />
      <span>Apps</span>
    </button>

    <button
      class="mixer-tab-btn"
      class:is-active={activeTab === 'ambience'}
      onclick={() => (activeTab = 'ambience')}
      use:islandTooltip={{ text: 'Ambient Focus Soundscapes' }}
    >
      <Waves size={12} />
      <span>Ambience</span>
    </button>

    <button
      class="mixer-tab-btn"
      class:is-active={activeTab === 'devices'}
      onclick={() => (activeTab = 'devices')}
      use:islandTooltip={{ text: 'Audio Output Devices' }}
    >
      <Speaker size={12} />
      <span>Devices</span>
    </button>
  </div>

  <ScrollView class="mixer-drawer-scroll">
    <div class="mixer-content-area">
      {#if activeTab === 'apps'}
      <!-- Master Audio Card -->
      <div class="master-volume-card">
        <div class="master-label-group">
          <button
            class="mini-mute-btn"
            class:is-muted={masterMuted}
            onclick={() => void toggleMasterMute()}
            use:islandTooltip={{ text: masterMuted ? 'Unmute Master' : 'Mute Master' }}
            aria-label="Toggle Master Mute"
          >
            {#if masterMuted}
              <VolumeX size={13} />
            {:else if masterVol > 50}
              <Volume2 size={13} />
            {:else}
              <Volume1 size={13} />
            {/if}
          </button>
          <span class="master-title">Master</span>
        </div>

        <div class="session-slider-cluster">
          <div class="slider-track-wrap">
            <div
              class="slider-track-fill"
              style="width: {masterMuted ? 0 : masterVol}%;"
            ></div>
            <input
              type="range"
              min="0"
              max="100"
              value={masterMuted ? 0 : masterVol}
              class="slider-native-input"
              oninput={handleMasterVolInput}
            />
          </div>
          <span class="vol-readout-text">{masterMuted ? 0 : Math.round(masterVol)}%</span>
        </div>
      </div>

      <!-- App Sessions List -->
      {#if displaySessions.length === 0}
        <div class="empty-mixer-state">
          <span>No active application audio streams</span>
        </div>
      {:else}
        {#each displaySessions as s, idx (s.pid + '-' + (s.processName || s.name) + '-' + idx)}
          {@const brand = resolveAppBrand(s.processName || s.name)}
          <div class="app-session-item">
            <div class="app-icon-wrap">
              {#if brand.icon}
                <img
                  src={brand.icon}
                  alt={brand.name}
                  class="app-icon-img"
                  draggable="false"
                  decoding="async"
                />
              {:else}
                <Volume2 size={12} class="app-fallback-icon" />
              {/if}
            </div>

            <div class="app-name-wrap">
              <span class="app-name-text" title={brand.name || s.name}>
                {brand.name || s.name}
              </span>
              <span class="app-proc-text" title={s.processName || s.name}>
                {s.processName || s.name}
              </span>
            </div>

            <!-- Peak Meter Visualizer -->
            <div class="session-peak-bar">
              <div
                class="session-peak-fill"
                style="height: {s.isMuted ? 0 : Math.min(100, Math.round(s.peak * 100))}%;"
              ></div>
            </div>

            <div class="session-slider-cluster">
              <button
                class="mini-mute-btn"
                class:is-muted={s.isMuted}
                onclick={() => void toggleAppMute(s.pid, s.isMuted)}
                use:islandTooltip={{ text: s.isMuted ? 'Unmute' : 'Mute' }}
                aria-label="Toggle App Mute"
              >
                {#if s.isMuted}
                  <VolumeX size={12} />
                {:else if s.volume > 50}
                  <Volume2 size={12} />
                {:else}
                  <Volume1 size={12} />
                {/if}
              </button>

              <div class="slider-track-wrap">
                <div
                  class="slider-track-fill"
                  style="width: {s.isMuted ? 0 : s.volume}%;"
                ></div>
                <input
                  type="range"
                  min="0"
                  max="100"
                  value={s.isMuted ? 0 : s.volume}
                  class="slider-native-input"
                  oninput={(e) => handleAppVolInput(s.pid, e)}
                />
              </div>

              <span class="vol-readout-text">
                {s.isMuted ? 0 : Math.round(s.volume)}%
              </span>
            </div>
          </div>
        {/each}
      {/if}
    {:else if activeTab === 'ambience'}
      <!-- Ambient Focus Soundscapes -->
      <div class="soundscape-panel">
        <div class="preset-pill-group">
          <button
            class="preset-pill-btn"
            class:is-active={soundscape.preset === 'brown'}
            onclick={() => handleSelectPreset('brown')}
            use:islandTooltip={{ text: 'Deep Brownian rumble for deep concentration' }}
          >
            <Waves size={12} />
            <span>Brown Noise</span>
          </button>

          <button
            class="preset-pill-btn"
            class:is-active={soundscape.preset === 'rain'}
            onclick={() => handleSelectPreset('rain')}
            use:islandTooltip={{ text: 'Gentle rainfall for calming focus' }}
          >
            <CloudRain size={12} />
            <span>Gentle Rain</span>
          </button>

          <button
            class="preset-pill-btn"
            class:is-active={soundscape.preset === 'static'}
            onclick={() => handleSelectPreset('static')}
            use:islandTooltip={{ text: 'Warm analog tape hiss' }}
          >
            <Radio size={12} />
            <span>Soft Static</span>
          </button>
        </div>

        <div class="soundscape-controls-row">
          <button
            class="soundscape-play-btn"
            onclick={toggleSoundscape}
            use:islandTooltip={{ text: soundscape.isPlaying ? 'Pause Ambience' : 'Play Ambience' }}
            aria-label="Toggle Ambient Sound"
          >
            {#if soundscape.isPlaying}
              <Pause size={14} />
            {:else}
              <Play size={14} style="margin-left: 2px;" />
            {/if}
          </button>

          <div class="session-slider-cluster">
            <div class="slider-track-wrap">
              <div
                class="slider-track-fill"
                style="width: {soundscape.volume}%;"
              ></div>
              <input
                type="range"
                min="0"
                max="100"
                value={soundscape.volume}
                class="slider-native-input"
                oninput={handleSoundscapeVolInput}
              />
            </div>
            <span class="vol-readout-text">{soundscape.volume}%</span>
          </div>

          {#if soundscape.isDucked}
            <span class="ducking-indicator-badge">Ducked</span>
          {/if}
        </div>
      </div>
    {:else if activeTab === 'devices'}
      <!-- Audio Output Devices -->
      <div class="endpoint-list">
        {#if endpoints.length === 0}
          <div class="empty-mixer-state">
            <span>No playback devices found</span>
          </div>
        {:else}
          {#each endpoints as ep (ep.id)}
            <button
              class="endpoint-item"
              class:is-default={ep.isDefault}
              onclick={() => void switchDefaultEndpoint(ep.id)}
            >
              <div class="endpoint-left-meta">
                <div class="endpoint-icon-wrap">
                  <Speaker size={13} class="endpoint-icon" />
                </div>
                <span class="endpoint-name">{ep.name}</span>
              </div>
              {#if ep.isDefault}
                <span class="endpoint-badge">
                  <Check size={11} class="endpoint-badge-icon" />
                  <span>Default</span>
                </span>
              {/if}
            </button>
          {/each}
        {/if}
      </div>
    {/if}
    </div>
  </ScrollView>
</DrawerPanel>

<style src="./styles/AudioMixerDrawer.css"></style>

