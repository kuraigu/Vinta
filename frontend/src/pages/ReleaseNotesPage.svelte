<!--
  @fileoverview Changelog and GitHub release notes viewer.

  Fetches version release logs from GitHub, parses markdown formatting, and displays
  feature highlights, bug fixes, and contributor attributions.
-->
<script module lang="ts">
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 320, h: 36, r: 18 },
    expanded: { w: 440, h: 480, r: 28, minH: 320, maxH: 640 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import AnnouncementsPage from './AnnouncementsPage.svelte'
</script>

<AnnouncementsPage initialTab="releases" />
