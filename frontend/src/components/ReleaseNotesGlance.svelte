<!--
  @component ReleaseNotesGlance
  @description Preview card highlighting new features after a Vinta version update.
-->
<script lang="ts">
  import { onMount, getContext } from 'svelte'
  import { ScrollText } from 'lucide-svelte'
  import BentoCard from './BentoCard.svelte'
  import ActionCapsule from './ActionCapsule.svelte'
  import { getReleaseNotes } from '../services/announcementsService'
  import ReleaseNotesPage, {
    config as releaseNotesConfig
  } from '../pages/ReleaseNotesPage.svelte'
  import { appInfoStore, type IslandConfig } from '../store'

  let {
    onClick
  }: {
    onClick?: () => void
  } = $props()

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let notesSnippet = $state('')

  function formatVersion(ver: string | undefined): string {
    if (!ver) return 'v0.1.0'
    const clean = ver.trim().replace(/^v+/i, '')
    return `v${clean}`
  }

  onMount(async () => {
    try {
      const ver = $appInfoStore?.version || '0.1.0'
      const raw = await getReleaseNotes(ver)
      if (raw) {
        notesSnippet =
          raw
            .replace(/[#*`>-]+/g, ' ')
            .replace(/\s+/g, ' ')
            .trim()
            .slice(0, 110) + '...'
      }
    } catch {
      // Ignore background fetch error
    }
  })

  function handleOpen() {
    if (onClick) {
      onClick()
    } else if (typeof navigate === 'function') {
      navigate(ReleaseNotesPage, releaseNotesConfig, true)
    }
  }
</script>

<div
  class="release-notes-glance-wrap"
  role="button"
  tabindex="0"
  onclick={handleOpen}
  onkeydown={(e) => e.key === 'Enter' && handleOpen()}
>
  <BentoCard variant="bento" class="release-notes-glance-card">
    <div class="glance-top-row">
      <div class="glance-brand-group">
        <div class="glance-icon-badge">
          <ScrollText size={12} />
        </div>
        <span class="glance-brand-name">Release Notes</span>
        <span class="glance-version-pill"
          >{formatVersion($appInfoStore?.version)}</span
        >
      </div>

      <ActionCapsule
        label="View"
        onclick={handleOpen}
        tooltip="View full release notes"
      />
    </div>

    <div class="glance-content-preview">
      <h4 class="glance-title">
        What's New in Vinta {formatVersion($appInfoStore?.version)}
      </h4>
      <p class="glance-message">
        {notesSnippet ||
          'Discover the latest features, improvements, and fixes in this build.'}
      </p>
    </div>
  </BentoCard>
</div>

<style src="./styles/ReleaseNotesGlance.css"></style>
