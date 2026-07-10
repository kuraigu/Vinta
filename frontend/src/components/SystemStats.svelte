<script lang="ts">
  // Note: This component is currently a mock component. It is simply graphical for now and does not have
  // the necessary monitoring features yet. - kuraigu (Innoh Reloza)

  import { onMount } from 'svelte'

  export let expanded: boolean = false

  // Mock data for CPU and RAM
  let cpuUsage = 0
  let ramUsage = 0

  onMount(() => {
    // Animate to some initial value shortly after mount
    setTimeout(() => {
      cpuUsage = 42
      ramUsage = 68
    }, 300)

    // Randomly fluctuate every few seconds to simulate real activity
    const interval = setInterval(() => {
      cpuUsage = Math.max(
        10,
        Math.min(95, cpuUsage + (Math.random() * 20 - 10))
      )
      ramUsage = Math.max(20, Math.min(90, ramUsage + (Math.random() * 10 - 5)))
    }, 2500)

    return () => clearInterval(interval)
  })

  // Calculate SVG stroke dashes
  const radius = 24
  const circumference = 2 * Math.PI * radius
  $: cpuOffset = circumference - (cpuUsage / 100) * circumference
  $: ramOffset = circumference - (ramUsage / 100) * circumference
</script>

{#if expanded}
  <div class="stats-card">
    <div class="stat-item">
      <div class="ring-container">
        <svg class="ring" width="60" height="60">
          <circle class="ring-bg" cx="30" cy="30" r={radius} />
          <circle
            class="ring-progress cpu"
            cx="30"
            cy="30"
            r={radius}
            style="stroke-dasharray: {circumference}; stroke-dashoffset: {cpuOffset};"
          />
        </svg>
        <div class="ring-label">
          <span class="value">{Math.round(cpuUsage)}%</span>
        </div>
      </div>
      <span class="label">CPU</span>
    </div>

    <div class="stat-item">
      <div class="ring-container">
        <svg class="ring" width="60" height="60">
          <circle class="ring-bg" cx="30" cy="30" r={radius} />
          <circle
            class="ring-progress ram"
            cx="30"
            cy="30"
            r={radius}
            style="stroke-dasharray: {circumference}; stroke-dashoffset: {ramOffset};"
          />
        </svg>
        <div class="ring-label">
          <span class="value">{Math.round(ramUsage)}%</span>
        </div>
      </div>
      <span class="label">RAM</span>
    </div>
  </div>
{/if}

<style>
  .stats-card {
    padding: 0.8rem 1rem;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-around;
    height: 100%;
    box-sizing: border-box;
    /* Fade in when expanded */
    animation: fadeScale 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  }

  .stat-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  .ring-container {
    position: relative;
    width: 60px;
    height: 60px;
  }

  .ring {
    transform: rotate(-90deg);
  }

  .ring-bg {
    fill: none;
    stroke: var(--c-border);
    stroke-width: 6;
  }

  .ring-progress {
    fill: none;
    stroke-width: 6;
    stroke-linecap: round;
    transition: stroke-dashoffset 1.5s cubic-bezier(0.22, 1, 0.36, 1);
  }

  .cpu {
    stroke: #ff3b30;
    filter: drop-shadow(0 0 6px rgba(255, 59, 48, 0.5));
  }

  .ram {
    stroke: #0a84ff;
    filter: drop-shadow(0 0 6px rgba(10, 132, 255, 0.5));
  }

  .ring-label {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .value {
    font-size: 0.95rem;
    font-weight: 700;
    color: var(--c-text-primary);
    letter-spacing: -0.5px;
  }

  .label {
    font-size: 0.75rem;
    font-weight: 700;
    color: var(--c-text-secondary);
    letter-spacing: 1px;
    text-transform: uppercase;
  }

  @keyframes fadeScale {
    0% {
      opacity: 0;
      transform: scale(0.9);
    }
    100% {
      opacity: 1;
      transform: scale(1);
    }
  }
</style>
