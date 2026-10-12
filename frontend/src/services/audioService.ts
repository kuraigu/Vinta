// Synthetic procedural audio chime synthesizer using Web Audio API
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

let alarmChimeIntervalId: number | undefined

/**
 * Initiates a repeating procedural dual-tone ringing chime for active alarms.
 *
 * @param volume Master volume gain multiplier between 0 and 1.
 */
export function startAlarmRingingChime(volume = 0.85): void {
  stopAlarmRingingChime()
  playNotifChime(volume)
  alarmChimeIntervalId = window.setInterval(() => {
    playNotifChime(volume)
  }, 2500)
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
