<script lang="ts">
  // This component is currently running on mock data -kuraigu (Innoh Reloza)

  let { expanded = false } = $props<{ expanded?: boolean }>()

  // Mock data
  let temp = $state(72)
  let condition = $state('Partly Cloudy')
  let location = $state('San Francisco')
  let high = $state(75)
  let low = $state(62)
</script>

{#if !expanded}
  <div class="weather-minimal">
    <!-- Tiny icon -->
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="#ffd60a"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <path d="M17.5 19a4.5 4.5 0 1 0-2.3-8.3A7 7 0 1 0 5 15" />
      <path d="M12 2v2" />
      <path d="M12 20v2" />
      <path d="m4.93 4.93 1.41 1.41" />
      <path d="m17.66 17.66 1.41 1.41" />
      <path d="M2 12h2" />
      <path d="M20 12h2" />
      <path d="m6.34 17.66-1.41 1.41" />
      <path d="m19.07 4.93-1.41 1.41" />
    </svg>
    <span class="temp-minimal">{temp}°</span>
  </div>
{:else}
  <div class="weather-card">
    <div class="left-col">
      <div class="location">{location}</div>
      <div class="temp">{temp}°</div>
    </div>
    <div class="right-col">
      <svg
        class="icon"
        width="38"
        height="38"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M17.5 19a4.5 4.5 0 1 0-2.3-8.3A7 7 0 1 0 5 15" />
        <path d="M12 2v2" />
        <path d="M12 20v2" />
        <path d="m4.93 4.93 1.41 1.41" />
        <path d="m17.66 17.66 1.41 1.41" />
        <path d="M2 12h2" />
        <path d="M20 12h2" />
        <path d="m6.34 17.66-1.41 1.41" />
        <path d="m19.07 4.93-1.41 1.41" />
      </svg>
      <div class="condition">{condition}</div>
      <div class="hi-lo">H:{high}° L:{low}°</div>
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

  /* EXPANDED */
  .weather-card {
    padding: 1.2rem 1.5rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    height: 100%;
    box-sizing: border-box;
    color: var(--c-text-primary);
  }

  .left-col {
    display: flex;
    flex-direction: column;
    justify-content: center;
    height: 100%;
  }

  .location {
    font-size: 1.1rem;
    font-weight: 600;
    margin-bottom: 0.2rem;
  }

  .temp {
    font-size: 3.5rem;
    font-weight: 300;
    line-height: 1;
    letter-spacing: -2px;
  }

  .right-col {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    justify-content: center;
    height: 100%;
  }

  .icon {
    color: #ffd60a;
    filter: drop-shadow(0 0 10px rgba(255, 214, 10, 0.4));
    margin-bottom: 0.4rem;
  }

  .condition {
    font-size: 0.9rem;
    font-weight: 600;
    margin-bottom: 0.1rem;
  }

  .hi-lo {
    font-size: 0.8rem;
    color: var(--c-text-secondary);
    font-weight: 500;
  }
</style>
