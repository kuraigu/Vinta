<script module lang="ts">
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 320, h: 36, r: 18 },
    expanded: { w: 440, h: 390, r: 28 },
    borderColor: 'var(--c-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import Weather from '../components/Weather.svelte'

  let { expanded = false }: { expanded?: boolean } = $props()
</script>

<div class="weather-page-layout" class:expanded class:minimal={!expanded}>
  <Weather {expanded} variant="page" />
</div>

<style src="./styles/WeatherPage.css"></style>
