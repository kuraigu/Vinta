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
  import { onMount, getContext } from 'svelte'
  import { ScrollText, ArrowUpRight } from 'lucide-svelte'
  import BackButton from '../components/BackButton.svelte'
  import BentoCard from '../components/BentoCard.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import MarkdownViewer from '../components/MarkdownViewer.svelte'
  import { getReleaseNotes } from '../services/announcementsService'
  import { appInfoStore, tStore } from '../store'
  import HomePage, { config as homeConfig } from './HomePage.svelte'

  let releaseNotes = $state('')
  let isLoading = $state(true)

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  /**
   * Navigates island viewport back to previous view or Home feed.
   */
  function handleBack() {
    if (typeof navigate === 'function') {
      navigate(HomePage, homeConfig, true)
    }
  }

  /**
   * Normalizes version string to ensure a single consistent leading 'v' prefix.
   */
  function formatVersion(ver: string | undefined): string {
    if (!ver) return 'v0.1.0'
    const clean = ver.trim().replace(/^v+/i, '')
    return `v${clean}`
  }

  /**
   * Fetches latest release notes from the GitHub API and extracts markdown changelog text.
   */
  async function loadNotes() {
    isLoading = true
    const version = $appInfoStore?.version || '0.1.0'
    releaseNotes = await getReleaseNotes(version)
    isLoading = false
  }

  onMount(() => {
    void loadNotes()
  })
</script>

<div data-vinta="release-notes-page" class="release-notes-page-layout">
  <header class="release-notes-header">
    <div class="header-left">
      <BackButton onback={handleBack} title={$tStore('common.back_to_home')} />
      <div class="header-titles">
        <h2 class="header-main-title">{$tStore('release_notes.title')}</h2>
        <span class="header-subtitle">{$tStore('release_notes.subtitle')}</span>
      </div>
    </div>

    <div class="version-tag-pill">
      <ScrollText size={12} />
      <span>{formatVersion($appInfoStore?.version)}</span>
    </div>
  </header>

  <div class="release-notes-body-wrap">
    <ScrollView class="release-notes-scroll-view">
      {#if isLoading}
        <div class="notes-loading-state">
          <span>{$tStore('release_notes.loading')}</span>
        </div>
      {:else}
        <div class="notes-container">
          <BentoCard variant="bento" class="release-notes-card">
            <MarkdownViewer content={releaseNotes} />

            {#if $appInfoStore?.repoUrl}
              <a
                href="{$appInfoStore.repoUrl}/releases"
                target="_blank"
                rel="noreferrer"
                class="view-all-releases-link"
              >
                <span>{$tStore('release_notes.view_all_github')}</span>
                <ArrowUpRight size={12} />
              </a>
            {/if}
          </BentoCard>
        </div>
      {/if}
    </ScrollView>
  </div>
</div>

<style src="./styles/ReleaseNotesPage.css"></style>
