<script lang="ts">
  import { onMount } from 'svelte'
  import { ArrowUpRight, ShieldCheck, GitBranch, Bug } from 'lucide-svelte'
  import BentoCard from '../../BentoCard.svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import ReleaseNotesGlance from '../../ReleaseNotesGlance.svelte'
  import appIcon from '@appicon'
  import { reportFeedback } from '../../../services/feedbackService'
  import { tStore } from '../../../store'
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

  function handleReportFeedback() {
    reportFeedback().catch(() => {})
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
  <!-- Brand Hero with Consolidated Version and Build Specs -->
  <BentoCard variant="announcement" class="about-hero-card">
    <div class="about-logo-wrapper">
      <img src={appIcon} alt="Vinta" class="about-logo-img" />
    </div>
    <div class="about-hero-details">
      <div class="about-brand-header-row">
        <span class="about-brand-title">{appInfo.name}</span>
        <span class="about-brand-studio"
          >{$tStore('about.by_freematrix')}</span
        >
      </div>

      <div class="about-release-capsule">
        <span
          class="capsule-channel"
          class:official={appInfo.isOfficialBuild}
          class:dev={!appInfo.isOfficialBuild}
        >
          {#if appInfo.isOfficialBuild}
            <ShieldCheck size={11} strokeWidth={2.2} />
            <span>{$tStore('about.official')}</span>
          {:else}
            <GitBranch size={11} strokeWidth={2.2} />
            <span>{$tStore('about.dev')}</span>
          {/if}
        </span>
        <span class="capsule-divider"></span>
        <span class="capsule-version">v{displayVersion}</span>
        <span class="capsule-divider"></span>
        <span class="capsule-build"
          >{$tStore('about.build', { build: displayBuild })}</span
        >
      </div>
    </div>
  </BentoCard>

  <!-- Project Links Bento -->
  <SettingItem label={$tStore('about.project_resources')} flexCol={true}>
    <div class="about-links-row">
      <SettingActionButton
        href={appInfo.repoUrl}
        target="_blank"
        rel="noreferrer"
        fullWidth={true}
      >
        <ArrowUpRight size={13} strokeWidth={2} />
        <span>{$tStore('about.repository')}</span>
      </SettingActionButton>
      <SettingActionButton onclick={handleReportFeedback} fullWidth={true}>
        <Bug size={13} strokeWidth={2} />
        <span>{$tStore('about.report_issue')}</span>
      </SettingActionButton>
    </div>
  </SettingItem>

  <!-- Release Notes Glance -->
  <ReleaseNotesGlance />

  <!-- Maintainers and Contributors -->
  {#if maintainers.length > 0 || generalContributors.length > 0}
    <BentoCard variant="bento" class="contributors-card">
      <div class="specs-header">{$tStore('about.project_credits')}</div>
      <div class="contributors-list">
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
      </div>
      <a
        href={`${appInfo.repoUrl}/blob/main/docs/CONTRIBUTORS.md`}
        target="_blank"
        rel="noreferrer"
        class="contributor-full-link"
      >
        <span>{$tStore('about.view_full_roster')}</span>
        <ArrowUpRight size={12} />
      </a>
    </BentoCard>
  {/if}
</div>

<style>
  .setting-group.about-page {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  :global(.about-hero-card) {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .about-logo-wrapper {
    width: 44px;
    height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    flex-shrink: 0;
  }

  .about-logo-img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }

  .about-hero-details {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
    flex: 1;
    min-width: 0;
    text-align: left;
  }

  .about-brand-header-row {
    display: flex;
    align-items: baseline;
    gap: 6px;
    flex-wrap: wrap;
  }

  .about-brand-title {
    font-size: 16px;
    font-weight: 700;
    letter-spacing: -0.01em;
    color: var(--vinta-text-primary);
    line-height: 1.2;
  }

  .about-brand-studio {
    font-size: 11px;
    font-weight: 500;
    color: var(--vinta-text-muted);
    letter-spacing: -0.01em;
  }

  .about-release-capsule {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    padding: 3px 9px;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--vinta-border);
    font-size: 11px;
    line-height: 1.2;
    user-select: text;
  }

  .capsule-channel {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.02em;
    color: var(--vinta-text-secondary);
  }

  .capsule-channel.official {
    color: #4ade80;
  }

  .capsule-channel.dev {
    color: #a5b4fc;
  }

  .capsule-divider {
    width: 1px;
    height: 9px;
    background: rgba(255, 255, 255, 0.14);
  }

  .capsule-version {
    font-weight: 600;
    color: var(--vinta-text-primary);
    font-variant-numeric: tabular-nums;
    font-size: 11px;
  }

  .capsule-build {
    color: var(--vinta-text-muted);
    font-variant-numeric: tabular-nums;
    font-size: 10.5px;
  }

  .about-links-row {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
  }

  .about-links-row > :global(*) {
    flex: 1;
  }

  .specs-header {
    font-size: 11.5px;
    font-weight: 600;
    color: var(--vinta-text-primary);
    text-align: left;
    margin-bottom: 8px;
    letter-spacing: -0.01em;
  }

  .contributors-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    width: 100%;
  }

  .contributor-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 10px;
    border-radius: 9px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--vinta-border);
    text-decoration: none;
    color: var(--vinta-text-primary);
    transition:
      background 0.15s ease,
      border-color 0.15s ease;
    width: 100%;
    box-sizing: border-box;
  }

  .contributor-row:hover {
    background: rgba(255, 255, 255, 0.07);
    border-color: var(--vinta-border-hover, rgba(255, 255, 255, 0.18));
  }

  .contributor-info {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex: 1;
    overflow: hidden;
  }

  .contributor-name {
    font-size: 11.5px;
    font-weight: 600;
    color: var(--vinta-text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .contributor-handle {
    font-size: 10px;
    color: var(--vinta-text-muted);
    font-weight: 500;
    white-space: nowrap;
  }

  .role-badge {
    font-size: 9px;
    font-weight: 600;
    padding: 2px 7px;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.08);
    color: var(--vinta-text-secondary);
    border: 1px solid var(--vinta-border);
    flex-shrink: 0;
  }

  .role-badge.maintainer {
    background: rgba(99, 102, 241, 0.15);
    color: #a5b4fc;
    border-color: rgba(99, 102, 241, 0.35);
  }

  .contributor-full-link {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-size: 10.5px;
    font-weight: 600;
    color: var(--vinta-accent-primary, var(--vinta-primary, #38bdf8));
    text-decoration: none;
    margin-top: 8px;
    align-self: flex-start;
    transition: opacity 0.15s ease;
  }

  .contributor-full-link:hover {
    opacity: 0.8;
  }
</style>
