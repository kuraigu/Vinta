<script lang="ts">
  interface Props {
    isPlaying?: boolean
    barCount?: number
    variant?: 'minimal' | 'full'
  }

  let {
    isPlaying = false,
    barCount = 24,
    variant = 'minimal'
  }: Props = $props()

  // Pre-generated seed heights for organic visual diversity
  const seedHeights = [
    35, 60, 90, 50, 75, 100, 80, 45, 70, 95, 85, 60,
    40, 80, 95, 65, 50, 85, 90, 55, 75, 45, 65, 35
  ]

  let heights = $derived(
    Array.from({ length: barCount }, (_, i) => seedHeights[i % seedHeights.length])
  )
</script>

<div class="dynamic-equalizer {variant}" class:is-playing={isPlaying}>
  {#each heights as h, i (i)}
    <div
      class="eq-bar"
      style="
        --bar-height: {h}%;
        --delay: {((i % 12) * 0.07).toFixed(2)}s;
        --duration: {(0.6 + (i % 5) * 0.12).toFixed(2)}s;
      "
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
    gap: 2.5px;
    height: 12px;
  }

  .eq-bar {
    flex: 1;
    height: var(--bar-height);
    background: var(--c-primary, #6366f1);
    border-radius: 2px;
    opacity: 0.4;
    transform-origin: bottom;
    transition: transform 0.3s cubic-bezier(0.2, 0.8, 0.2, 1), opacity 0.3s ease, background-color 0.3s ease;
  }

  .dynamic-equalizer.is-playing .eq-bar {
    opacity: 0.95;
    background: linear-gradient(180deg, var(--c-purple, #a855f7) 0%, var(--c-primary, #6366f1) 100%);
    animation: eqPulse var(--duration) infinite ease-in-out alternate;
    animation-delay: var(--delay);
  }

  @keyframes eqPulse {
    0% {
      transform: scaleY(0.25);
      filter: brightness(0.9);
    }
    50% {
      transform: scaleY(1.15);
      filter: brightness(1.2);
    }
    100% {
      transform: scaleY(0.4);
      filter: brightness(1);
    }
  }
</style>
