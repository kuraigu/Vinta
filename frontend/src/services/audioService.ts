/**
 * @fileoverview Generates app sounds and notification chimes.
 *
 * Synthesizes two-tone notification bells, focus completion sounds, and
 * alarm sounds (chime, radar, digital, marimba) using the Web Audio API without
 * needing audio asset files.
 */

let audioCtx: AudioContext | null = null

function getAudioContext(): AudioContext {
  if (!audioCtx) {
    const AudioContextClass =
      window.AudioContext ||
      (window as unknown as { webkitAudioContext: typeof AudioContext })
        .webkitAudioContext
    audioCtx = new AudioContextClass()
  }
  if (audioCtx.state === 'suspended') {
    void audioCtx.resume()
  }
  return audioCtx
}

/**
 * Plays a two-tone bell chime for incoming notifications.
 */
export function playNotifChime(volume = 0.8) {
  try {
    const ctx = getAudioContext()
    const now = ctx.currentTime

    const masterGain = ctx.createGain()
    masterGain.gain.setValueAtTime(Math.max(0, Math.min(1, volume)), now)
    masterGain.connect(ctx.destination)

    // Two-tone soft bell chime (E5: 659.25Hz -> B5: 987.77Hz)
    const tones = [
      { freq: 659.25, start: 0, duration: 0.25 },
      { freq: 987.77, start: 0.08, duration: 0.35 }
    ]

    for (const t of tones) {
      const osc = ctx.createOscillator()
      const gain = ctx.createGain()

      osc.type = 'sine'
      osc.frequency.setValueAtTime(t.freq, now + t.start)

      const startTime = now + t.start
      gain.gain.setValueAtTime(0.001, startTime)
      gain.gain.exponentialRampToValueAtTime(0.3, startTime + 0.02)
      gain.gain.exponentialRampToValueAtTime(0.0001, startTime + t.duration)

      osc.connect(gain)
      gain.connect(masterGain)

      osc.start(startTime)
      osc.stop(startTime + t.duration)
    }
  } catch {
    // Ignore audio playback errors if audio context blocked
  }
}

/**
 * Plays a procedural focus session or break completion chime using the Web Audio API.
 * Focus completion uses an uplifting ascending harmonic sequence, while break completion
 * uses a calm settling sequence.
 *
 * @param type The completion event type.
 * @param volume Master volume gain multiplier between 0 and 1.
 */
export function playFocusChime(
  type: 'focusComplete' | 'breakComplete' = 'focusComplete',
  volume = 0.75
) {
  try {
    const ctx = getAudioContext()
    const now = ctx.currentTime

    const masterGain = ctx.createGain()
    masterGain.gain.setValueAtTime(Math.max(0, Math.min(1, volume)), now)
    masterGain.connect(ctx.destination)

    const tones =
      type === 'focusComplete'
        ? [
            { freq: 523.25, start: 0, duration: 0.35 }, // C5
            { freq: 659.25, start: 0.12, duration: 0.38 }, // E5
            { freq: 783.99, start: 0.24, duration: 0.42 }, // G5
            { freq: 1046.5, start: 0.36, duration: 0.65 } // C6
          ]
        : [
            { freq: 880.0, start: 0, duration: 0.35 }, // A5
            { freq: 659.25, start: 0.14, duration: 0.45 }, // E5
            { freq: 523.25, start: 0.28, duration: 0.55 } // C5
          ]

    for (const t of tones) {
      const osc = ctx.createOscillator()
      const gain = ctx.createGain()

      osc.type = 'sine'
      osc.frequency.setValueAtTime(t.freq, now + t.start)

      const startTime = now + t.start
      gain.gain.setValueAtTime(0.001, startTime)
      gain.gain.exponentialRampToValueAtTime(0.28, startTime + 0.02)
      gain.gain.exponentialRampToValueAtTime(0.0001, startTime + t.duration)

      osc.connect(gain)
      gain.connect(masterGain)

      osc.start(startTime)
      osc.stop(startTime + t.duration)
    }
  } catch {
    // Ignore audio context errors if playback is blocked
  }
}

export type AlarmSoundPreset = 'chime' | 'radar' | 'digital' | 'marimba'

/**
 * Plays a single iteration of a procedural alarm sound preset.
 *
 * @param preset The chosen sound preset identifier.
 * @param volume Master volume gain multiplier between 0 and 1.
 */
export function playAlarmPreset(
  preset: AlarmSoundPreset = 'chime',
  volume = 0.85
): void {
  try {
    const ctx = getAudioContext()
    const now = ctx.currentTime

    const masterGain = ctx.createGain()
    masterGain.gain.setValueAtTime(Math.max(0, Math.min(1, volume)), now)
    masterGain.connect(ctx.destination)

    if (preset === 'radar') {
      // High-frequency sonar ping with subtle frequency drop
      const osc = ctx.createOscillator()
      const gain = ctx.createGain()

      osc.type = 'sine'
      osc.frequency.setValueAtTime(1320, now)
      osc.frequency.exponentialRampToValueAtTime(880, now + 0.35)

      gain.gain.setValueAtTime(0.001, now)
      gain.gain.exponentialRampToValueAtTime(0.35, now + 0.02)
      gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.5)

      osc.connect(gain)
      gain.connect(masterGain)
      osc.start(now)
      osc.stop(now + 0.5)
    } else if (preset === 'digital') {
      // Classic digital alarm double-beep burst
      const beeps = [
        { start: 0, duration: 0.12 },
        { start: 0.16, duration: 0.12 }
      ]
      for (const b of beeps) {
        const osc = ctx.createOscillator()
        const gain = ctx.createGain()

        osc.type = 'square'
        osc.frequency.setValueAtTime(1046.5, now + b.start)

        const st = now + b.start
        gain.gain.setValueAtTime(0.001, st)
        gain.gain.linearRampToValueAtTime(0.18, st + 0.01)
        gain.gain.linearRampToValueAtTime(0.001, st + b.duration)

        osc.connect(gain)
        gain.connect(masterGain)
        osc.start(st)
        osc.stop(st + b.duration)
      }
    } else if (preset === 'marimba') {
      // Woody acoustic triad chords with short percussive decay
      const notes = [
        { freq: 523.25, start: 0, duration: 0.3 }, // C5
        { freq: 659.25, start: 0.08, duration: 0.3 }, // E5
        { freq: 783.99, start: 0.16, duration: 0.45 } // G5
      ]
      for (const n of notes) {
        const osc = ctx.createOscillator()
        const gain = ctx.createGain()

        osc.type = 'triangle'
        osc.frequency.setValueAtTime(n.freq, now + n.start)

        const st = now + n.start
        gain.gain.setValueAtTime(0.001, st)
        gain.gain.exponentialRampToValueAtTime(0.35, st + 0.01)
        gain.gain.exponentialRampToValueAtTime(0.0001, st + n.duration)

        osc.connect(gain)
        gain.connect(masterGain)
        osc.start(st)
        osc.stop(st + n.duration)
      }
    } else {
      // Default soft chime
      playNotifChime(volume)
    }
  } catch {
    // Ignore audio errors if audio context blocked by browser
  }
}

/**
 * Previews an alarm sound preset once.
 *
 * @param preset The sound preset to audition.
 * @param volume Audition volume.
 */
export function previewAlarmSound(
  preset: AlarmSoundPreset,
  volume = 0.85
): void {
  playAlarmPreset(preset, volume)
}

let alarmChimeIntervalId: number | undefined

/**
 * Initiates a repeating procedural ringing chime for active alarms.
 *
 * @param volume Master volume gain multiplier between 0 and 1.
 * @param preset The alarm sound preset to ring.
 */
export function startAlarmRingingChime(
  volume = 0.85,
  preset: AlarmSoundPreset = 'chime'
): void {
  stopAlarmRingingChime()
  playAlarmPreset(preset, volume)
  alarmChimeIntervalId = window.setInterval(() => {
    playAlarmPreset(preset, volume)
  }, 2400)
}

/**
 * Halts any actively ringing alarm chime interval.
 */
export function stopAlarmRingingChime(): void {
  if (alarmChimeIntervalId !== undefined) {
    clearInterval(alarmChimeIntervalId)
    alarmChimeIntervalId = undefined
  }
}
