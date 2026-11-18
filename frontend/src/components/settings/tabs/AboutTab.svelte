<!--
  @component AboutTab
  @description About Vinta information page. Displays current app version, links to the
  GitHub repository and issue tracker, and exports system diagnostics for bug reports.
-->
<script lang="ts">
  import { onMount, getContext } from 'svelte'
  import {
    ArrowUpRight,
    ShieldCheck,
    GitBranch,
    Bug,
    Globe,
    Lightbulb
  } from 'lucide-svelte'
  import BentoCard from '../../BentoCard.svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import appIcon from '@appicon-svg'
  import {
    tStore,
    type OverlayPageId,
    type FeedbackOverlayKind
  } from '../../../store'
  import {
    GetAppInfo,
    GetContributors
  } from '../../../../bindings/Vinta/backend/app.js'

  interface AppInfoData {
    name: string
    version: string
    gitCommit: string
    buildType: string
    buildTime: string
    isOfficialBuild: boolean
    goVersion: string
    arch: string
    os: string
    wailsVersion: string
    repoUrl: string
  }

  interface ContributorItem {
    name: string
    handle: string
    role: string
    avatar: string
    url: string
  }

  let appInfo = $state<AppInfoData>({
    name: 'Vinta',
    version: '0.1.0',
    gitCommit: 'dev',
    buildType: 'Development',
    buildTime: '',
    isOfficialBuild: false,
    goVersion: 'go1.23',
    arch: 'amd64',
    os: 'windows',
    wailsVersion: 'v2.13.0',
    repoUrl: 'https://github.com/kuraigu/Vinta'
  })

  let contributors = $state<ContributorItem[]>([])

  /**
   * Formats the application version for display.
   *
   * @param ver - Raw version string from app info
   * @returns Cleaned version number
   */
  function getDisplayVersion(ver: string | undefined): string {
    if (!ver) return '0.1.0'
    const trimmed = ver.replace(/^v/, '').replace(/-dev$/i, '')
    return trimmed || '0.1.0'
  }
  let displayVersion = $derived(getDisplayVersion(appInfo.version))

  /**
   * Formats the build identifier for display.
   *
   * @param buildTime - Raw build time stamp
   * @param commit - Git commit hash or branch
   * @returns Formatted build string
   */
  function getDisplayBuild(
    buildTime: string | undefined,
    commit: string | undefined
  ): string {
    if (buildTime && buildTime !== 'Development' && buildTime !== 'Unknown') {
      return buildTime
    }
    if (commit && commit !== 'dev' && commit !== 'Unknown') {
      return commit.slice(0, 7)
    }
    return buildTime || 'local'
  }
  let displayBuild = $derived(
    getDisplayBuild(appInfo.buildTime, appInfo.gitCommit)
  )

  function getMaintainers() {
    return contributors.filter(
      (c) => c.role === 'Founder' || c.role === 'Maintainer'
    )
  }
  let maintainers = $derived(getMaintainers())

  function getGeneralContributors() {
    return contributors.filter(
      (c) => c.role !== 'Founder' && c.role !== 'Maintainer'
    )
  }
  let generalContributors = $derived(getGeneralContributors())

  const openOverlay =
    getContext<
      (
        overlay: OverlayPageId,
        options?: { feedbackKind?: FeedbackOverlayKind }
      ) => void
    >('openOverlay')

  function openFeedbackModal(kind: FeedbackOverlayKind) {
    openOverlay?.('feedback', { feedbackKind: kind })
  }

  onMount(() => {
    GetAppInfo()
      .then((info) => {
        if (info) appInfo = info as AppInfoData
      })
      .catch(() => {})
    GetContributors()
      .then((list) => {
        if (list && Array.isArray(list) && list.length > 0) {
          contributors = list as ContributorItem[]
        }
      })
      .catch(() => {})
  })
</script>

<div class="setting-group about-page">
  <!-- Brand Hero Bento Card -->
  <BentoCard variant="bento" class="about-hero-card">
    <div class="about-logo-wrapper">
      <img src={appIcon} alt="Vinta" class="about-logo-img" draggable="false" />
    </div>
    <div class="about-hero-details">
      <div class="about-brand-header-row">
        <span class="about-brand-title">{appInfo.name}</span>
        <span class="about-brand-studio">{$tStore('about.by_freematrix')}</span>
      </div>

      <div class="about-release-badges">
        <span class="about-pill-badge version">v{displayVersion}</span>
        <span
          class="about-pill-badge channel"
          class:official={appInfo.isOfficialBuild}
          class:dev={!appInfo.isOfficialBuild}
        >
          {#if appInfo.isOfficialBuild}
            <ShieldCheck size={10} strokeWidth={2.3} />
            <span>{$tStore('about.official')}</span>
          {:else}
            <GitBranch size={10} strokeWidth={2.3} />
            <span>{$tStore('about.dev')}</span>
          {/if}
        </span>
        <span class="about-pill-badge build"
          >{$tStore('about.build', { build: displayBuild })}</span
        >
      </div>
    </div>
  </BentoCard>

  <!-- Project Links Bento -->
  <BentoCard variant="group">
    <SettingItem
      label={$tStore('about.project_resources')}
      icon={Globe}
      flexCol={true}
    >
      <div class="about-links-stack">
        <SettingActionButton
          href={appInfo.repoUrl}
          target="_blank"
          rel="noreferrer"
          fullWidth={true}
        >
          <ArrowUpRight size={13} strokeWidth={2} />
          <span>{$tStore('about.repository')}</span>
        </SettingActionButton>
        <div class="about-links-row">
          <SettingActionButton
            onclick={() => openFeedbackModal('bug')}
            fullWidth={true}
          >
            <Bug size={13} strokeWidth={2} />
            <span>{$tStore('about.report_issue')}</span>
          </SettingActionButton>
          <SettingActionButton
            onclick={() => openFeedbackModal('feature')}
            fullWidth={true}
          >
            <Lightbulb size={13} strokeWidth={2} />
            <span>{$tStore('about.request_feature')}</span>
          </SettingActionButton>
        </div>
      </div>
    </SettingItem>
  </BentoCard>

  <!-- Maintainers and Contributors -->
  {#if maintainers.length > 0 || generalContributors.length > 0}
    <div class="about-section-header">
      <span>{$tStore('about.project_credits')}</span>
      <a
        href={`${appInfo.repoUrl}/blob/main/docs/CONTRIBUTORS.md`}
        target="_blank"
        rel="noreferrer"
        class="contributor-full-link"
      >
        <span>{$tStore('about.view_full_roster')}</span>
        <ArrowUpRight size={11} />
      </a>
    </div>

    <BentoCard variant="group" class="contributors-card">
      {#each [...maintainers, ...generalContributors] as c (c.handle)}
        <a
          href={c.url}
          target="_blank"
          rel="noreferrer"
          class="contributor-row"
        >
          <div class="contributor-info">
            <span class="contributor-name">{c.name}</span>
            <span class="contributor-handle">@{c.handle}</span>
          </div>
          <span
            class="role-badge"
            class:maintainer={c.role === 'Founder' || c.role === 'Maintainer'}
          >
            {c.role}
          </span>
        </a>
      {/each}
    </BentoCard>
  {/if}
</div>

<style src="../../styles/settings/AboutTab.css"></style>
