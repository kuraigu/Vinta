/**
 * @fileoverview Generates procedural application audio and signature notification chimes.
 *
 * Synthesizes Vinta's signature multi-harmonic crystal notification chime, focus completion
 * sequences, and energetic alarm motifs (melodic chime, tactical radar, digital chronograph,
 * and acoustic marimba) using the native Web Audio API without external audio assets.
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
 * Plays Vinta's signature crystal droplet chime for incoming notifications.
 *
 * Combines a subtle acoustic liquid-glass transient tap with an uplifting
 * three-note harmonic crystal triad (F#5 -> B5 -> E6) and shimmering resonant
 * overtones for an unmistakable, modern sonic identity.
 *
 * @param volume Master volume gain multiplier between 0 and 1.
 */
export function playNotifChime(volume = 0.8): void {
  try {
    const ctx = getAudioContext()
    const now = ctx.currentTime

    const masterGain = ctx.createGain()
    masterGain.gain.setValueAtTime(Math.max(0, Math.min(1, volume)), now)
    masterGain.connect(ctx.destination)

    // 1. Tactile liquid glass transient tap (very soft, fast 24ms pitch drop)
    const tapOsc = ctx.createOscillator()
    const tapGain = ctx.createGain()
    tapOsc.type = 'sine'
    tapOsc.frequency.setValueAtTime(2600, now)
    tapOsc.frequency.exponentialRampToValueAtTime(1600, now + 0.024)

    tapGain.gain.setValueAtTime(0.001, now)
    tapGain.gain.linearRampToValueAtTime(0.12, now + 0.003)
    tapGain.gain.exponentialRampToValueAtTime(0.0001, now + 0.028)

    tapOsc.connect(tapGain)
    tapGain.connect(masterGain)
    tapOsc.start(now)
    tapOsc.stop(now + 0.03)

    // 2. Signature Ascending Crystal Triad: F#5 (739.99Hz) -> B5 (987.77Hz) -> E6 (1318.51Hz)
    const triad = [
      {
        freq: 739.99,
        overtone: 2042.37,
        start: 0.015,
        duration: 0.32,
        gain: 0.28
      },
      {
        freq: 987.77,
        overtone: 2726.24,
        start: 0.075,
        duration: 0.36,
        gain: 0.25
      },
      {
        freq: 1318.51,
        overtone: 2637.02,
        start: 0.145,
        duration: 0.44,
        gain: 0.32
      }
    ]

    for (const note of triad) {
      const startTime = now + note.start

      // Fundamental tone (warm sine)
      const osc = ctx.createOscillator()
      const gain = ctx.createGain()
      osc.type = 'sine'
      osc.frequency.setValueAtTime(note.freq, startTime)

      gain.gain.setValueAtTime(0.0001, startTime)
      gain.gain.linearRampToValueAtTime(note.gain, startTime + 0.008)
      gain.gain.exponentialRampToValueAtTime(0.0001, startTime + note.duration)

      osc.connect(gain)
      gain.connect(masterGain)
      osc.start(startTime)
      osc.stop(startTime + note.duration)

      // Crystalline overtone partial (adds metallic bell/glass shimmer)
      const overtoneOsc = ctx.createOscillator()
      const overtoneGain = ctx.createGain()
      overtoneOsc.type = 'sine'
      overtoneOsc.frequency.setValueAtTime(note.overtone, startTime)

      overtoneGain.gain.setValueAtTime(0.0001, startTime)
      overtoneGain.gain.linearRampToValueAtTime(
        note.gain * 0.22,
        startTime + 0.006
      )
      overtoneGain.gain.exponentialRampToValueAtTime(
        0.0001,
        startTime + note.duration * 0.65
      )

      overtoneOsc.connect(overtoneGain)
      overtoneGain.connect(masterGain)
      overtoneOsc.start(startTime)
      overtoneOsc.stop(startTime + note.duration * 0.65)
    }
  } catch {
    // Ignore audio playback errors if audio context is blocked
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
): void {
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
 * Plays a complete musical motif of a chosen alarm preset.
 *
 * Each preset is synthesized procedurally with rich harmonics, tight rhythmic
 * cadence, and sufficient acoustic urgency to function effectively as an alarm.
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
      // Tactical Sonar Radar: Dual swept pings with acoustic reverberation and alert climax
      const pings = [
        {
          freqStart: 1650,
          freqEnd: 1220,
          start: 0.0,
          duration: 0.26,
          gain: 0.32
        },
        {
          freqStart: 1220,
          freqEnd: 1220,
          start: 0.13,
          duration: 0.16,
          gain: 0.14
        }, // Echo 1
        {
          freqStart: 1960,
          freqEnd: 1440,
          start: 0.34,
          duration: 0.3,
          gain: 0.38
        }, // Accented ping
        {
          freqStart: 1440,
          freqEnd: 1440,
          start: 0.48,
          duration: 0.18,
          gain: 0.16
        }, // Echo 2
        {
          freqStart: 2180,
          freqEnd: 1650,
          start: 0.72,
          duration: 0.36,
          gain: 0.42
        } // Alert climax
      ]

      for (const p of pings) {
        const osc = ctx.createOscillator()
        const gain = ctx.createGain()
        const filter = ctx.createBiquadFilter()

        osc.type = 'sine'
        filter.type = 'bandpass'
        filter.frequency.setValueAtTime(
          (p.freqStart + p.freqEnd) / 2,
          now + p.start
        )
        filter.Q.setValueAtTime(4.0, now + p.start)

        const st = now + p.start
        osc.frequency.setValueAtTime(p.freqStart, st)
        osc.frequency.exponentialRampToValueAtTime(
          Math.max(20, p.freqEnd),
          st + p.duration
        )

        gain.gain.setValueAtTime(0.0001, st)
        gain.gain.linearRampToValueAtTime(p.gain, st + 0.008)
        gain.gain.exponentialRampToValueAtTime(0.0001, st + p.duration)

        osc.connect(filter)
        filter.connect(gain)
        gain.connect(masterGain)

        osc.start(st)
        osc.stop(st + p.duration)
      }
    } else if (preset === 'digital') {
      // Classic Digital Chronograph Alarm: High-energy two-phrase 4-burst cadence
      const bursts = [
        // Phrase 1 (1150Hz)
        { freq: 1150, start: 0.0, duration: 0.065 },
        { freq: 1150, start: 0.1, duration: 0.065 },
        { freq: 1150, start: 0.2, duration: 0.065 },
        { freq: 1150, start: 0.3, duration: 0.065 },
        // Phrase 2 (1380Hz)
        { freq: 1380, start: 0.48, duration: 0.065 },
        { freq: 1380, start: 0.58, duration: 0.065 },
        { freq: 1380, start: 0.68, duration: 0.065 },
        { freq: 1380, start: 0.78, duration: 0.065 }
      ]

      const filter = ctx.createBiquadFilter()
      filter.type = 'lowpass'
      filter.frequency.setValueAtTime(3200, now)
      filter.connect(masterGain)

      for (const b of bursts) {
        const osc = ctx.createOscillator()
        const gain = ctx.createGain()

        osc.type = 'square'
        const st = now + b.start
        osc.frequency.setValueAtTime(b.freq, st)

        gain.gain.setValueAtTime(0.0001, st)
        gain.gain.linearRampToValueAtTime(0.22, st + 0.006)
        gain.gain.linearRampToValueAtTime(0.2, st + b.duration - 0.006)
        gain.gain.exponentialRampToValueAtTime(0.0001, st + b.duration)

        osc.connect(gain)
        gain.connect(filter)

        osc.start(st)
        osc.stop(st + b.duration)
      }
    } else if (preset === 'marimba') {
      // Resonant Wooden Marimba: Upbeat syncopated 8-note melody
      const notes = [
        { freq: 523.25, start: 0.0, duration: 0.26 }, // C5
        { freq: 783.99, start: 0.12, duration: 0.24 }, // G5
        { freq: 1046.5, start: 0.23, duration: 0.24 }, // C6
        { freq: 1318.51, start: 0.35, duration: 0.28 }, // E6
        { freq: 1174.66, start: 0.48, duration: 0.24 }, // D6
        { freq: 783.99, start: 0.6, duration: 0.24 }, // G5
        { freq: 987.77, start: 0.72, duration: 0.26 }, // B5
        { freq: 1046.5, start: 0.85, duration: 0.38 } // C6 (held)
      ]

      for (const n of notes) {
        const st = now + n.start

        // Warm wooden bar fundamental (triangle)
        const osc = ctx.createOscillator()
        const gain = ctx.createGain()
        osc.type = 'triangle'
        osc.frequency.setValueAtTime(n.freq, st)

        gain.gain.setValueAtTime(0.0001, st)
        gain.gain.linearRampToValueAtTime(0.38, st + 0.004)
        gain.gain.exponentialRampToValueAtTime(0.0001, st + n.duration)

        osc.connect(gain)
        gain.connect(masterGain)
        osc.start(st)
        osc.stop(st + n.duration)

        // Percussive mallet tap overtone (short 4x partial)
        const tap = ctx.createOscillator()
        const tapGain = ctx.createGain()
        tap.type = 'sine'
        tap.frequency.setValueAtTime(n.freq * 3.98, st)

        tapGain.gain.setValueAtTime(0.0001, st)
        tapGain.gain.linearRampToValueAtTime(0.12, st + 0.002)
        tapGain.gain.exponentialRampToValueAtTime(0.0001, st + 0.035)

        tap.connect(tapGain)
        tapGain.connect(masterGain)
        tap.start(st)
        tap.stop(st + 0.04)
      }
    } else {
      // Radiant Campanile Chimes: Cascading 6-note bell arpeggio sequence
      const chimeMelody = [
        { freq: 659.25, start: 0.0, duration: 0.45 }, // E5
        { freq: 830.61, start: 0.14, duration: 0.42 }, // G#5
        { freq: 987.77, start: 0.28, duration: 0.42 }, // B5
        { freq: 1318.51, start: 0.42, duration: 0.5 }, // E6
        { freq: 987.77, start: 0.6, duration: 0.42 }, // B5
        { freq: 1318.51, start: 0.76, duration: 0.6 } // E6 (resonant tail)
      ]

      for (const bell of chimeMelody) {
        const st = now + bell.start

        const osc = ctx.createOscillator()
        const gain = ctx.createGain()
        osc.type = 'sine'
        osc.frequency.setValueAtTime(bell.freq, st)

        gain.gain.setValueAtTime(0.0001, st)
        gain.gain.linearRampToValueAtTime(0.32, st + 0.008)
        gain.gain.exponentialRampToValueAtTime(0.0001, st + bell.duration)

        osc.connect(gain)
        gain.connect(masterGain)
        osc.start(st)
        osc.stop(st + bell.duration)

        // Bell shimmer harmonic overtone
        const overtone = ctx.createOscillator()
        const otGain = ctx.createGain()
        overtone.type = 'sine'
        overtone.frequency.setValueAtTime(bell.freq * 2.76, st)

        otGain.gain.setValueAtTime(0.0001, st)
        otGain.gain.linearRampToValueAtTime(0.08, st + 0.006)
        otGain.gain.exponentialRampToValueAtTime(
          0.0001,
          st + bell.duration * 0.55
        )

        overtone.connect(otGain)
        otGain.connect(masterGain)
        overtone.start(st)
        overtone.stop(st + bell.duration * 0.55)
      }
    }
  } catch {
    // Ignore audio errors if audio context is blocked by browser
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
 * Initiates an energetic, repeating procedural ringing cadence for active alarms.
 *
 * Repeats every 1750ms with a brief 650ms rhythmic pause between musical phrases
 * to sustain urgency without acoustic fatigue.
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
  }, 1750)
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
