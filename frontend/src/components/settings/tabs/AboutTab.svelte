<script lang="ts">
  import { onMount } from 'svelte'
  import { ExternalLink } from 'lucide-svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import appIcon from '@appicon'
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
  <!-- Brand Hero -->
  <div class="about-hero-card">
    <div class="about-logo-wrapper">
      <img src={appIcon} alt="Vinta" class="about-logo-img" />
    </div>
    <div class="about-hero-details">
      <div class="about-brand-row">
        <span class="about-brand-title">{appInfo.name}</span>
        <span
          class="build-badge"
          class:official={appInfo.isOfficialBuild}
          class:dev={!appInfo.isOfficialBuild}
        >
          {appInfo.buildType}
        </span>
      </div>
      <span class="about-version-text">
        {appInfo.version?.startsWith('v')
          ? appInfo.version
          : 'v' + (appInfo.version || '0.1.0-dev')}
      </span>
    </div>
  </div>

  <!-- System Specifications Bento -->
  <SettingItem
    label="System Specifications"
    flexCol={true}
  >
    <div class="specs-compact-list">
      <div class="spec-compact-row">
        <span class="spec-compact-label">Version</span>
        <span class="spec-compact-val">
          {appInfo.version?.startsWith('v')
            ? appInfo.version
            : 'v' + (appInfo.version || '0.1.0-dev')}
        </span>
      </div>
      <div class="spec-compact-row">
        <span class="spec-compact-label">Build</span>
        <span class="spec-compact-val">{appInfo.buildTime || 'Development Build'}</span>
      </div>
    </div>
  </SettingItem>

  <!-- Project Links Bento -->
  <SettingItem
    label="Project Resources"
    flexCol={true}
  >
    <div class="about-links-row">
      <SettingActionButton
        href={appInfo.repoUrl}
        target="_blank"
        rel="noreferrer"
        fullWidth={true}
      >
        <ExternalLink size={12} strokeWidth={2} />
        <span>Repository</span>
      </SettingActionButton>
      <SettingActionButton
        href={`${appInfo.repoUrl}/releases`}
        target="_blank"
        rel="noreferrer"
        fullWidth={true}
      >
        <ExternalLink size={12} strokeWidth={2} />
        <span>Release Notes</span>
      </SettingActionButton>
    </div>
  </SettingItem>

  <!-- Maintainers and Contributors -->
  {#if maintainers.length > 0 || generalContributors.length > 0}
    <div class="about-card contributors-card">
      <div class="specs-header">Project Credits</div>
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
        <span>View Full Contributor Roster</span>
        <ExternalLink size={11} />
      </a>
    </div>
  {/if}
</div>

<style>
  .setting-group.about-page {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .about-hero-card {
    display: flex;
    align-items: flex-start;
    gap: 14px;
    padding: 15px 17px;
    border-radius: 18px;
    background: linear-gradient(
        145deg,
        rgba(255, 255, 255, 0.07) 0%,
        rgba(255, 255, 255, 0.02) 100%
      ),
      #15161c;
    border: 1px solid rgba(255, 255, 255, 0.09);
    box-shadow:
      inset 0 1px 0 0 rgba(255, 255, 255, 0.12),
      0 4px 18px rgba(0, 0, 0, 0.15);
  }

  .about-logo-wrapper {
    width: 44px;
    height: 44px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    flex-shrink: 0;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.18);
    border: 1px solid var(--c-border);
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
    gap: 3px;
    flex: 1;
    min-width: 0;
    text-align: left;
  }

  .about-brand-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .about-brand-title {
    font-size: 16px;
    font-weight: 700;
    letter-spacing: -0.01em;
    color: var(--c-text-primary);
    line-height: 1.2;
  }

  .build-badge {
    display: inline-flex;
    align-items: center;
    font-size: 9px;
    font-weight: 700;
    padding: 2px 7px;
    border-radius: 6px;
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  .build-badge.official {
    background: rgba(16, 185, 129, 0.15);
    color: #10b981;
    border: 1px solid rgba(16, 185, 129, 0.3);
  }

  .build-badge.dev {
    background: rgba(245, 158, 11, 0.15);
    color: #f59e0b;
    border: 1px solid rgba(245, 158, 11, 0.3);
  }

  .about-version-text {
    font-size: 11px;
    font-weight: 600;
    color: var(--c-text-muted);
    font-variant-numeric: tabular-nums;
    line-height: 1.2;
  }



  .about-card {
    display: flex;
    flex-direction: column;
    padding: 13px 15px;
    border-radius: 16px;
    background: linear-gradient(
        145deg,
        rgba(255, 255, 255, 0.055) 0%,
        rgba(255, 255, 255, 0.018) 100%
      ),
      #15161c;
    border: 1px solid rgba(255, 255, 255, 0.08);
    box-shadow:
      inset 0 1px 0 0 rgba(255, 255, 255, 0.1),
      0 3px 12px rgba(0, 0, 0, 0.12);
  }

  .specs-compact-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    width: 100%;
    padding: 8px 10px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.06);
    box-sizing: border-box;
  }

  .spec-compact-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 11px;
  }

  .spec-compact-label {
    color: var(--c-text-muted);
    font-weight: 500;
  }

  .spec-compact-val {
    color: var(--c-text-primary);
    font-weight: 600;
    font-variant-numeric: tabular-nums;
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
    color: var(--c-text-primary);
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
    border: 1px solid var(--c-border);
    text-decoration: none;
    color: var(--c-text-primary);
    transition: background 0.15s ease, border-color 0.15s ease;
    width: 100%;
    box-sizing: border-box;
  }

  .contributor-row:hover {
    background: rgba(255, 255, 255, 0.07);
    border-color: var(--c-border-hover, rgba(255, 255, 255, 0.18));
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
    color: var(--c-text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .contributor-handle {
    font-size: 10px;
    color: var(--c-text-muted);
    font-weight: 500;
    white-space: nowrap;
  }

  .role-badge {
    font-size: 9px;
    font-weight: 600;
    padding: 2px 7px;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.08);
    color: var(--c-text-secondary);
    border: 1px solid var(--c-border);
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
    color: var(--c-primary, #6366f1);
    text-decoration: none;
    margin-top: 8px;
    align-self: flex-start;
    transition: opacity 0.15s ease;
  }

  .contributor-full-link:hover {
    opacity: 0.8;
  }
</style>
