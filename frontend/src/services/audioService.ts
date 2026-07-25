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
