import { writable, get } from 'svelte/store'
import { globalMediaStore } from '../store'

export type SoundscapePreset = 'none' | 'brown' | 'rain' | 'static'

export interface SoundscapeState {
  preset: SoundscapePreset
  isPlaying: boolean
  volume: number
  isDucked: boolean
}

const STORAGE_KEY_PRESET = 'vinta_soundscape_preset'
const STORAGE_KEY_VOL = 'vinta_soundscape_volume'

const initialPreset =
  (localStorage.getItem(STORAGE_KEY_PRESET) as SoundscapePreset) || 'none'
const initialVol = Number(localStorage.getItem(STORAGE_KEY_VOL) ?? '65')

export const soundscapeStore = writable<SoundscapeState>({
  preset: initialPreset,
  isPlaying: false,
  volume: isNaN(initialVol) ? 65 : Math.max(0, Math.min(100, initialVol)),
  isDucked: false
})

let audioCtx: AudioContext | null = null
let currentSourceNode: AudioBufferSourceNode | null = null
let masterGainNode: GainNode | null = null
let filterNode: BiquadFilterNode | null = null
let isMediaDucking = false

/**
 * Initializes the procedural Web Audio API context.
 */
function ensureAudioContext(): AudioContext {
  if (!audioCtx) {
    const AudioContextClass =
      window.AudioContext ||
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (window as any).webkitAudioContext
    audioCtx = new AudioContextClass()
  }
  if (audioCtx.state === 'suspended') {
    void audioCtx.resume()
  }
  return audioCtx
}

/**
 * Creates a procedural seamless audio buffer for the chosen preset.
 */
function createProceduralBuffer(
  ctx: AudioContext,
  preset: SoundscapePreset
): AudioBuffer {
  const sampleRate = ctx.sampleRate
  const durationSeconds = 5
  const frameCount = sampleRate * durationSeconds
  const buffer = ctx.createBuffer(2, frameCount, sampleRate)

  for (let channel = 0; channel < 2; channel++) {
    const data = buffer.getChannelData(channel)

    if (preset === 'brown') {
      let lastOut = 0.0
      for (let i = 0; i < frameCount; i++) {
        const white = Math.random() * 2 - 1
        lastOut = (lastOut + 0.02 * white) / 1.02
        data[i] = lastOut * 3.2
      }
    } else if (preset === 'rain') {
      let b0 = 0
      let b1 = 0
      let b2 = 0
      for (let i = 0; i < frameCount; i++) {
        const white = Math.random() * 2 - 1
        b0 = 0.99886 * b0 + white * 0.0555179
        b1 = 0.99332 * b1 + white * 0.0750759
        b2 = 0.969 * b2 + white * 0.153852
        let pink = b0 + b1 + b2 + white * 0.5362
        // Random droplet impulses
        if (Math.random() < 0.0007) {
          pink += (Math.random() - 0.5) * 1.6
        }
        data[i] = pink * 0.16
      }
    } else if (preset === 'static') {
      let b0 = 0
      let b1 = 0
      let b2 = 0
      for (let i = 0; i < frameCount; i++) {
        const white = Math.random() * 2 - 1
        b0 = 0.99886 * b0 + white * 0.0555179
        b1 = 0.99332 * b1 + white * 0.0750759
        b2 = 0.969 * b2 + white * 0.153852
        data[i] = (b0 + b1 + b2) * 0.12
      }
    }

    // Apply seamless crossfade envelope at seam to eliminate loop boundary clicks
    const fadeSamples = Math.min(2200, Math.floor(frameCount / 4))
    for (let i = 0; i < fadeSamples; i++) {
      const envelope = 0.5 * (1 - Math.cos((Math.PI * i) / fadeSamples))
      data[i] *= envelope
      data[frameCount - 1 - i] *= envelope
    }
  }

  return buffer
}

/**
 * Updates gain target respecting user volume setting and active ducking state.
 */
function updateGainTarget(instant = false): void {
  if (!masterGainNode || !audioCtx) return

  const state = get(soundscapeStore)
  const baseGain = (state.volume / 100) * 0.5
  const effectiveGain = isMediaDucking ? baseGain * 0.25 : baseGain

  const now = audioCtx.currentTime
  if (instant) {
    masterGainNode.gain.cancelScheduledValues(now)
    masterGainNode.gain.setValueAtTime(effectiveGain, now)
  } else {
    masterGainNode.gain.setTargetAtTime(
      effectiveGain,
      now,
      isMediaDucking ? 0.35 : 0.55
    )
  }

  soundscapeStore.update((s) => ({
    ...s,
    isDucked: isMediaDucking
  }))
}

/**
 * Starts or switches ambient soundscape playback.
 */
export function playPreset(preset: SoundscapePreset): void {
  if (preset === 'none') {
    stopSoundscape()
    return
  }

  const ctx = ensureAudioContext()
  stopSoundscape(false)

  const buffer = createProceduralBuffer(ctx, preset)
  const source = ctx.createBufferSource()
  source.buffer = buffer
  source.loop = true

  const filter = ctx.createBiquadFilter()
  if (preset === 'brown') {
    filter.type = 'lowpass'
    filter.frequency.setValueAtTime(320, ctx.currentTime)
  } else if (preset === 'rain') {
    filter.type = 'bandpass'
    filter.frequency.setValueAtTime(850, ctx.currentTime)
    filter.Q.setValueAtTime(0.65, ctx.currentTime)
  } else if (preset === 'static') {
    filter.type = 'lowpass'
    filter.frequency.setValueAtTime(2200, ctx.currentTime)
  }

  const gain = ctx.createGain()
  source.connect(filter)
  filter.connect(gain)
  gain.connect(ctx.destination)

  currentSourceNode = source
  filterNode = filter
  masterGainNode = gain

  source.start()

  soundscapeStore.update((s) => ({
    ...s,
    preset,
    isPlaying: true
  }))
  localStorage.setItem(STORAGE_KEY_PRESET, preset)

  updateGainTarget(true)
}

/**
 * Stops the active soundscape.
 */
export function stopSoundscape(clearPreset = true): void {
  if (currentSourceNode) {
    try {
      currentSourceNode.stop()
      currentSourceNode.disconnect()
    } catch {
      // Ignore cleanup error
    }
    currentSourceNode = null
  }

  if (filterNode) {
    try {
      filterNode.disconnect()
    } catch {
      // Ignore cleanup error
    }
    filterNode = null
  }

  if (masterGainNode) {
    try {
      masterGainNode.disconnect()
    } catch {
      // Ignore cleanup error
    }
    masterGainNode = null
  }

  soundscapeStore.update((s) => ({
    ...s,
    preset: clearPreset ? 'none' : s.preset,
    isPlaying: false,
    isDucked: false
  }))

  if (clearPreset) {
    localStorage.setItem(STORAGE_KEY_PRESET, 'none')
  }
}

/**
 * Adjusts the ambient soundscape volume (0-100).
 */
export function setSoundscapeVolume(volume: number): void {
  const clamped = Math.max(0, Math.min(100, volume))
  soundscapeStore.update((s) => ({
    ...s,
    volume: clamped
  }))
  localStorage.setItem(STORAGE_KEY_VOL, String(clamped))
  updateGainTarget(false)
}

/**
 * Toggles play/pause for the active soundscape preset.
 */
export function toggleSoundscape(): void {
  const state = get(soundscapeStore)
  if (state.isPlaying) {
    stopSoundscape(false)
  } else {
    const targetPreset = state.preset === 'none' ? 'brown' : state.preset
    playPreset(targetPreset)
  }
}

// Wire automatic smart ducking on media playback changes
globalMediaStore.subscribe((media) => {
  const isPlaying = Boolean(media?.isPlaying)
  if (isPlaying !== isMediaDucking) {
    isMediaDucking = isPlaying
    updateGainTarget(false)
  }
})
