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

  let rawNotes = $state('')

  function formatVersion(ver: string | undefined): string {
    if (!ver) return 'v0.1.0'
    const clean = ver.trim().replace(/^v+/i, '')
    return `v${clean}`
  }

  /**
   * Scans release notes markdown for the first valid non-badge preview image.
   */
  function extractFirstImage(text: string | undefined): string | null {
    if (!text) return null
    const mdRegex = /!\[.*?\]\((https?:\/\/[^\s)"]+)(?:\s+["'][^"']*["'])?\)/g
    let match: RegExpExecArray | null
    while ((match = mdRegex.exec(text)) !== null) {
      const url = match[1]
      if (!url.includes('shields.io') && !url.includes('badge')) {
        return url
      }
    }
    const htmlRegex = /<img[^>]+src=["'](https?:\/\/[^"']+)["']/gi
    while ((match = htmlRegex.exec(text)) !== null) {
      const url = match[1]
      if (!url.includes('shields.io') && !url.includes('badge')) {
        return url
      }
    }
    return null
  }

  /**
   * Cleans and renders inline markdown formatting into safe HTML for the glance card preview.
   */
  function renderGlanceSnippet(raw: string | undefined, maxLen = 110): string {
    if (!raw) return ''

    let text = raw
      .replace(/!\[.*?\]\((https?:\/\/[^\s)"]+)(?:\s+["'][^"']*["'])?\)/g, '')
      .replace(/<img[^>]*>/gi, '')
      .replace(/<details[\s\S]*?<\/details>/gi, '')
      .replace(/^#+\s+/gm, '')
      .replace(/^---+$/gm, '')
      .replace(/\r?\n+/g, ' ')
      .trim()

    text = text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')

    const isTruncated = text.length > maxLen
    if (isTruncated) {
      text = text.slice(0, maxLen).trim()
    }

    text = text.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    text = text.replace(/(?<!\*)\*([^*\s][^*]*[^*\s]|\S)\*(?!\*)/g, '<em>$1</em>')
    text = text.replace(/`([^`]+)`/g, '<code class="glance-inline-code">$1</code>')
    text = text.replace(/\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g, '<span class="glance-inline-link">$1</span>')

    return isTruncated ? text + '...' : text
  }

  function getPreviewImage(): string | null {
    return extractFirstImage(rawNotes)
  }

  function getSnippetHtml(): string {
    if (!rawNotes) {
      return 'Discover the latest features, improvements, and fixes in this build.'
    }
    return renderGlanceSnippet(rawNotes, previewImage ? 85 : 120)
  }

  let previewImage = $derived(getPreviewImage())
  let snippetHtml = $derived(getSnippetHtml())

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

    <div class="glance-content-preview" class:has-image={!!previewImage}>
      <div class="glance-text-col">
        <h4 class="glance-title">
          What's New in Vinta {formatVersion($appInfoStore?.version)}
        </h4>
        <p class="glance-message">
          <!-- eslint-disable-next-line svelte/no-at-html-tags -->
          {@html snippetHtml}
        </p>
      </div>

      {#if previewImage}
        <div class="glance-thumb-wrapper">
          <img
            src={previewImage}
            alt=""
            class="glance-thumb-image"
            onerror={(e) => {
              const target = e.currentTarget as HTMLImageElement
              if (target.parentElement)
                target.parentElement.style.display = 'none'
            }}
          />
        </div>
      {/if}
    </div>
  </BentoCard>
</div>

<style src="./styles/ReleaseNotesGlance.css"></style>
