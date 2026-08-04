<script lang="ts">
  import { onDestroy } from 'svelte'

  interface Props {
    isPlaying?: boolean
    audioPeak?: number
    barCount?: number
    variant?: 'minimal' | 'full'
    progressPercent?: number
  }

  let {
    isPlaying = false,
    audioPeak = 0,
    barCount = 24,
    variant = 'minimal',
    progressPercent
  }: Props = $props()

  const MIN_HEIGHT = 25 // 25% baseline (~5-6px) minimum height when paused/quiet/muted

  let dynamicHeights = $state<number[]>([])
  let animFrame: number = 0

  $effect(() => {
    const count = barCount || 24
    if (dynamicHeights.length !== count) {
      dynamicHeights = Array(count).fill(MIN_HEIGHT)
    }
  })

  // Cleanly manage animation loop based on isPlaying state
  $effect(() => {
    if (isPlaying) {
      if (!animFrame) animFrame = requestAnimationFrame(updateHeights)
    } else {
      if (animFrame) {
        cancelAnimationFrame(animFrame)
        animFrame = 0
      }
      // Set to baseline height when paused
      const count = barCount || 24
      for (let i = 0; i < count; i++) {
        dynamicHeights[i] = MIN_HEIGHT
      }
    }
  })

  function updateHeights(time: number) {
    if (!isPlaying) return

    const t = time * 0.0045
    const peak = audioPeak || 0
    // Real audio peak signal gating: 0 when video/media has NO audio or is silent/muted; active wave scaling when audio is present
    const amplitude =
      peak > 0.008 ? Math.min(1.5, Math.max(0.35, peak * 2.8)) : 0

    const count = barCount || 24
    for (let i = 0; i < count; i++) {
      if (amplitude === 0) {
        dynamicHeights[i] = MIN_HEIGHT
      } else {
        const wave1 = Math.sin(t * 2.8 + i * 0.45)
        const wave2 = Math.cos(t * 4.2 + i * 0.75) * 0.5
        const wave3 = Math.sin(t * 6.5 + i * 1.3) * 0.25
        const combined = Math.abs(wave1 + wave2 + wave3) / 1.75
        const val = Math.round(MIN_HEIGHT + combined * 75 * amplitude)
        dynamicHeights[i] = Math.max(MIN_HEIGHT, Math.min(100, val))
      }
    }

    animFrame = requestAnimationFrame(updateHeights)
  }

  onDestroy(() => {
    if (animFrame) cancelAnimationFrame(animFrame)
  })
</script>

<div
  class="dynamic-equalizer {variant || 'minimal'}"
  class:is-playing={isPlaying}
>
  {#each dynamicHeights as h, i (i)}
    {@const barPercent = (i / Math.max(1, (barCount || 24) - 1)) * 100}
    {@const isPlayed =
      progressPercent === undefined || barPercent <= progressPercent}
    <div
      class="eq-bar {isPlayed ? 'is-played' : 'is-unplayed'}"
      style="height: {h}%;"
    ></div>
  {/each}
</div>

<style>
  .dynamic-equalizer {
    display: flex;
    align-items: flex-end;
    gap: 3px;
    height: 100%;
    width: 100%;
    pointer-events: none;
  }

  .dynamic-equalizer.minimal {
    gap: 2px;
    height: 12px;
    width: 18px;
  }

  .eq-bar {
    flex: 1;
    min-height: 5px;
    border-radius: 2px;
    transform-origin: bottom;
    transition:
      opacity 0.3s ease,
      background 0.3s ease;
    background: rgba(255, 255, 255, 0.45);
    opacity: 0.7;
    will-change: height;
  }

  .eq-bar.is-played {
    background: linear-gradient(180deg, #00f0ff 0%, #a855f7 50%, #ff007f 100%);
    opacity: 0.95;
  }

  .dynamic-equalizer.is-playing .eq-bar.is-played {
    opacity: 1;
    box-shadow:
      0 0 8px rgba(0, 240, 255, 0.6),
      0 0 14px rgba(255, 0, 127, 0.4);
  }

  .dynamic-equalizer.is-playing .eq-bar.is-unplayed {
    opacity: 0.4;
  }
</style>
