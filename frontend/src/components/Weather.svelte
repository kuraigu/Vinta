<script lang="ts">
  // This component is currently running on mock data -kuraigu (Innoh Reloza)
  import { Sun, CloudSun, Wind, Droplets, Compass } from 'lucide-svelte'

  let { expanded = false } = $props<{ expanded?: boolean }>()

  // Mock Weather data
  let temp = $state(72)
  let condition = $state('Partly Cloudy')
  let location = $state('Manila')
  let high = $state(75)
  let low = $state(62)
  let humidity = $state(58)
  let windSpeed = $state(8)
  let uvIndex = $state('3 (Low)')

  const forecast = [
    { day: 'Today', temp: '72°', icon: CloudSun, high: 75, low: 62 },
    { day: 'Sat', temp: '76°', icon: Sun, high: 78, low: 64 },
    { day: 'Sun', temp: '69°', icon: CloudSun, high: 71, low: 58 },
    { day: 'Mon', temp: '71°', icon: Sun, high: 74, low: 60 }
  ]
</script>

{#if !expanded}
  <div class="weather-minimal">
    <CloudSun size={16} color="var(--c-amber, #f59e0b)" />
    <span class="temp-minimal">{temp}°</span>
  </div>
{:else}
  <div class="weather-card">
    <div class="glow-backdrop"></div>

    <!-- Top Main Overview Row -->
    <div class="top-row">
      <div class="main-left">
        <div class="location" title={location}>{location}</div>
        <div class="temp-group">
          <span class="temp-val">{temp}°</span>
          <div class="temp-details">
            <span class="hilo-badge">H: {high}° &bull; L: {low}°</span>
            <span class="condition-badge">{condition}</span>
          </div>
        </div>
      </div>

      <div class="main-right">
        <div class="icon-wrapper">
          <CloudSun size={52} class="weather-icon-animated" />
        </div>
      </div>
    </div>

    <!-- Metrics Stats Bar -->
    <div class="metrics-bar">
      <div class="metric-item">
        <Droplets size={14} class="metric-icon" />
        <span class="metric-val">{humidity}%</span>
        <span class="metric-lbl">Humidity</span>
      </div>
      <div class="metric-sep"></div>
      <div class="metric-item">
        <Wind size={14} class="metric-icon" />
        <span class="metric-val">{windSpeed} mph</span>
        <span class="metric-lbl">Wind</span>
      </div>
      <div class="metric-sep"></div>
      <div class="metric-item">
        <Compass size={14} class="metric-icon" />
        <span class="metric-val">{uvIndex}</span>
        <span class="metric-lbl">UV</span>
      </div>
    </div>

    <!-- 4-Day Forecast Bar -->
    <div class="forecast-bar">
      {#each forecast as f (f.day)}
        <div class="forecast-item">
          <span class="f-day">{f.day}</span>
          <f.icon size={18} class="f-icon" />
          <span class="f-temp">{f.temp}</span>
        </div>
      {/each}
    </div>
  </div>
{/if}

<style>
  /* MINIMAL */
  .weather-minimal {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    color: var(--c-text-primary);
  }

  .temp-minimal {
    font-size: 1rem;
    font-weight: 500;
  }

  /* EXPANDED DESIGN */
  .weather-card {
    position: relative;
    width: 100%;
    height: 100%;
    padding: 0.9rem 1.1rem;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    overflow: hidden;
    color: var(--c-text-primary);
  }

  .glow-backdrop {
    position: absolute;
    top: -20px;
    right: -20px;
    width: 180px;
    height: 180px;
    background: radial-gradient(
      circle,
      rgba(245, 158, 11, 0.15) 0%,
      rgba(234, 88, 12, 0.05) 50%,
      transparent 75%
    );
    filter: blur(24px);
    pointer-events: none;
  }

  .top-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    z-index: 2;
  }

  .main-left {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
  }

  .location {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--c-text-muted);
    letter-spacing: 0.04em;
    text-transform: uppercase;
    max-width: 170px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .temp-group {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 2px;
  }

  .temp-val {
    font-size: 3.2rem;
    font-weight: 800;
    line-height: 1;
    letter-spacing: -1.8px;
    color: var(--c-text-primary);
  }

  .temp-details {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 3px;
  }

  .hilo-badge {
    font-size: 0.72rem;
    font-weight: 600;
    color: var(--c-text-muted);
    letter-spacing: 0.02em;
  }

  .condition-badge {
    font-size: 0.72rem;
    font-weight: 600;
    color: var(--c-warning);
    background: var(--c-item-bg);
    padding: 3px 10px;
    border-radius: 12px;
    border: 1px solid var(--c-border);
  }

  .main-right {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .icon-wrapper {
    color: var(--c-warning);
    filter: drop-shadow(0 4px 14px rgba(245, 158, 11, 0.4));
  }

  :global(.weather-icon-animated) {
    animation: float 4s ease-in-out infinite alternate;
  }

  @keyframes float {
    0% {
      transform: translateY(0px) rotate(0deg);
    }
    100% {
      transform: translateY(-4px) rotate(3deg);
    }
  }

  .hilo-badge {
    font-size: 0.72rem;
    font-weight: 600;
    color: var(--c-text-muted);
    margin-bottom: 2px;
  }

  /* Metrics Bar */
  .metrics-bar {
    display: flex;
    align-items: center;
    justify-content: space-around;
    background: var(--c-item-bg);
    border: 1px solid var(--c-border);
    border-radius: 12px;
    padding: 8px 10px;
    z-index: 2;
  }

  .metric-item {
    display: flex;
    align-items: center;
    gap: 5px;
  }

  :global(.metric-icon) {
    color: var(--c-primary);
  }

  .metric-val {
    font-size: 0.78rem;
    font-weight: 700;
    color: var(--c-text-primary);
  }

  .metric-lbl {
    font-size: 0.68rem;
    color: var(--c-text-muted);
  }

  .metric-sep {
    width: 1px;
    height: 14px;
    background: var(--c-border);
  }

  /* Forecast Bar */
  .forecast-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    z-index: 2;
  }

  .forecast-item {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 5px;
    padding: 8px 4px;
    background: var(--c-item-bg);
    border-radius: 12px;
    border: 1px solid var(--c-border);
  }

  .f-day {
    font-size: 0.72rem;
    font-weight: 600;
    color: var(--c-text-muted);
  }

  :global(.f-icon) {
    color: var(--c-warning);
  }

  .f-temp {
    font-size: 0.78rem;
    font-weight: 700;
    color: var(--c-text-primary);
  }
</style>
