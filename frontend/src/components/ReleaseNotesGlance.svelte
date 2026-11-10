<!--
  @component ReleaseNotesGlance
  @description Preview card highlighting new features after a Vinta version update.
-->
<script lang="ts">
  import { onMount, getContext } from 'svelte'
  import { ScrollText, ChevronRight } from 'lucide-svelte'
  import BentoCard from './BentoCard.svelte'
  import { getReleaseNotes } from '../services/announcementsService'
  import ReleaseNotesPage, {
    config as releaseNotesConfig
  } from '../pages/ReleaseNotesPage.svelte'
  import { appInfoStore, tStore, type IslandConfig } from '../store'

  let {
    onClick
  }: {
    onClick?: () => void
  } = $props()

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let rawNotes = $state('')

  function formatVersion(ver: string | undefined): string {
    if (!ver) return 'v0.1.0'
    const clean = ver.trim().replace(/^v+/i, '')
    return `v${clean}`
  }

  /**
   * Extracts up to 3 clean bullet highlights from raw release notes markdown.
   */
  function extractHighlights(raw: string): string[] {
    if (!raw) {
      return [
        'Performance improvements and smoother island transitions',
        'Refined bento card surfaces and theme styling',
        'General stability and bug fixes'
      ]
    }

    const lines = raw
      .replace(/<details[\s\S]*?<\/details>/gi, '')
      .replace(/!\[.*?\]\([^)]+\)/g, '')
      .replace(/<img[^>]*>/gi, '')
      .split(/\r?\n/)

    const bullets: string[] = []
    for (const rawLine of lines) {
      const line = rawLine.trim()
      if (!line) continue
      const match = line.match(/^[-*•]\s+(.+)/)
      if (match) {
        const cleaned = match[1]
          .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
          .replace(/\*\*([^*]+)\*\*/g, '$1')
          .replace(/`([^`]+)`/g, '$1')
          .replace(/\s*\(#\d+\)/g, '')
          .replace(/\s+by\s+@\S+.*/i, '')
          .trim()
        if (cleaned.length > 6 && !cleaned.toLowerCase().startsWith('full changelog')) {
          bullets.push(cleaned)
        }
      }
      if (bullets.length >= 3) break
    }

    if (bullets.length > 0) return bullets

    const plainParagraphs = lines
      .filter(
        (l) =>
          l.trim().length > 12 &&
          !l.trim().startsWith('#') &&
          !l.trim().startsWith('---')
      )
      .slice(0, 2)
      .map((l) =>
        l
          .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
          .replace(/\*\*([^*]+)\*\*/g, '$1')
          .replace(/`([^`]+)`/g, '$1')
          .trim()
      )

    return plainParagraphs.length > 0
      ? plainParagraphs
      : ['Discover the latest features, improvements, and fixes in this build.']
  }

  let highlights = $derived(extractHighlights(rawNotes))

  onMount(async () => {
    try {
      const ver = $appInfoStore?.version || '0.1.0'
      const raw = await getReleaseNotes(ver)
      if (raw) {
        rawNotes = raw
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
          <ScrollText size={13} />
        </div>
        <div class="glance-heading-col">
          <div class="glance-title-line">
            <span class="glance-brand-name">What's New in Vinta</span>
            <span class="glance-version-pill"
              >{formatVersion($appInfoStore?.version)}</span
            >
          </div>
          <span class="glance-subtitle">{$tStore('release_notes.title')}</span>
        </div>
      </div>

      <span class="glance-open-cta">
        <span>{$tStore('common.view')}</span>
        <ChevronRight size={13} />
      </span>
    </div>

    <ul class="glance-highlights-list">
      {#each highlights as bullet, idx (idx)}
        <li class="glance-highlight-item">
          <span class="highlight-dot"></span>
          <span class="highlight-text">{bullet}</span>
        </li>
      {/each}
    </ul>
  </BentoCard>
</div>

<style src="./styles/ReleaseNotesGlance.css"></style>
