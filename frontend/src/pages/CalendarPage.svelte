<script module lang="ts">
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 320, h: 42, r: 16 },
    expanded: { w: 440, h: 470, r: 28 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  /**
   * @fileoverview Page container hosting the full-page Calendar and Agenda interface.
   *
   * Configures the expanded island viewport dimensions for calendar operations and mounts
   * the interactive Calendar component.
   */

  import Calendar from '../components/Calendar.svelte'

  let { expanded = false }: { expanded?: boolean } = $props()
</script>

<div
  data-vinta="calendar-page"
  class="calendar-page-layout"
  class:expanded
  class:minimal={!expanded}
>
  <Calendar {expanded} variant="page" />
</div>

<style src="./styles/CalendarPage.css"></style>
