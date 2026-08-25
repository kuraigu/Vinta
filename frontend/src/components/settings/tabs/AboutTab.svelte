<script lang="ts">
  import { onMount } from 'svelte'
  import {
    Cpu,
    Layers,
    Code,
    ExternalLink,
    ShieldCheck,
    Heart,
    Crown
  } from 'lucide-svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import appIcon from '@appicon'
  import {
    GetAppInfo,
    GetContributors
  } from '../../../../wailsjs/go/backend/App.js'

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

<div class="setting-group">
  <div class="about-hero-card">
    <div class="about-logo-wrapper">
      <img src={appIcon} alt="Vinta" class="about-logo-img" />
    </div>
    <div class="about-hero-text">
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
        {appInfo.version?.startsWith('v') ? appInfo.version : 'v' + (appInfo.version || '0.1.0-dev')}
      </span>
    </div>
  </div>

  <div class="about-info-grid">
    <div class="about-info-tile left-aligned">
      <div class="tile-header">
        <Cpu size={14} strokeWidth={2} class="tile-icon" />
        <span class="tile-label">Architecture</span>
      </div>
      <span class="tile-value">{appInfo.os}/{appInfo.arch}</span>
    </div>

    <div class="about-info-tile left-aligned">
      <div class="tile-header">
        <Layers size={14} strokeWidth={2} class="tile-icon" />
        <span class="tile-label">Build</span>
      </div>
      <span class="tile-value">{appInfo.buildTime || 'Dev Build'}</span>
    </div>
  </div>

  <div class="about-actions-row">
    <SettingActionButton
      href={appInfo.repoUrl}
      target="_blank"
      rel="noreferrer"
      leftAligned={true}
      fullWidth={true}
    >
      <Code size={14} strokeWidth={2} />
      <span>GitHub Repository</span>
      <ExternalLink size={12} strokeWidth={2} class="link-icon" />
    </SettingActionButton>
    <SettingActionButton
      href={`${appInfo.repoUrl}/releases`}
      target="_blank"
      rel="noreferrer"
      leftAligned={true}
      fullWidth={true}
    >
      <ShieldCheck size={14} strokeWidth={2} />
      <span>Release Notes</span>
      <ExternalLink size={12} strokeWidth={2} class="link-icon" />
    </SettingActionButton>
  </div>

  {#if maintainers.length > 0}
    <div class="about-section-box">
      <div class="about-box-title">
        <Crown
          size={14}
          strokeWidth={2}
          style="color: var(--c-warning, #eab308)"
        />
        <span>Project Maintainers</span>
      </div>
      <div class="contributors-grid">
        {#each maintainers as c (c.handle)}
          <a
            href={c.url}
            target="_blank"
            rel="noreferrer"
            class="contributor-card"
          >
            <div class="contributor-meta">
              <span class="contributor-name">{c.name}</span>
              <span class="contributor-handle">@{c.handle}</span>
            </div>
            <span class="role-pill">{c.role}</span>
          </a>
        {/each}
      </div>
      {#if generalContributors.length === 0}
        <a
          href={`${appInfo.repoUrl}/blob/main/docs/CONTRIBUTORS.md`}
          target="_blank"
          rel="noreferrer"
          class="contributor-link-btn"
        >
          <span>View Full Contributor List</span>
          <ExternalLink size={11} />
        </a>
      {/if}
    </div>
  {/if}

  {#if generalContributors.length > 0}
    <div class="about-section-box">
      <div class="about-box-title">
        <Heart size={14} strokeWidth={2} style="color: #ef4444" />
        <span>Contributors</span>
      </div>
      <div class="contributors-grid">
        {#each generalContributors as c (c.handle)}
          <a
            href={c.url}
            target="_blank"
            rel="noreferrer"
            class="contributor-card"
          >
            <div class="contributor-meta">
              <span class="contributor-name">{c.name}</span>
              <span class="contributor-handle">@{c.handle}</span>
            </div>
            <span class="role-pill">{c.role}</span>
          </a>
        {/each}
      </div>
      <a
        href={`${appInfo.repoUrl}/blob/main/docs/CONTRIBUTORS.md`}
        target="_blank"
        rel="noreferrer"
        class="contributor-link-btn"
      >
        <span>View Full Contributor List</span>
        <ExternalLink size={11} />
      </a>
    </div>
  {/if}
</div>

<style>
  .setting-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .about-hero-card {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 14px;
    border-radius: 14px;
    background: var(--c-item-bg, rgba(255, 255, 255, 0.06));
    border: 1px solid var(--c-border);
  }

  .about-logo-wrapper {
    width: 38px;
    height: 38px;
    border-radius: 10px;
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

  .about-hero-text {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    justify-content: center;
    gap: 2px;
    text-align: left;
  }

  .about-brand-row {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 8px;
  }

  .about-brand-title {
    font-size: 15px;
    font-weight: 700;
    letter-spacing: -0.01em;
    line-height: 1.2;
  }

  .about-version-text {
    font-size: 11px;
    font-weight: 500;
    color: var(--c-text-secondary);
    opacity: 0.8;
    text-align: left;
    line-height: 1.2;
    margin: 0;
    padding: 0;
  }

  .about-info-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
  }

  .about-info-tile {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 8px 10px;
    border-radius: 10px;
    background: var(--c-item-bg, rgba(255, 255, 255, 0.05));
    border: 1px solid var(--c-border);
  }

  .about-info-tile.left-aligned {
    align-items: flex-start;
    text-align: left;
  }

  .tile-header {
    display: flex;
    align-items: center;
    gap: 4px;
    width: 100%;
  }

  .tile-header :global(.tile-icon) {
    color: var(--c-text-secondary);
  }

  .tile-label {
    font-size: 9px;
    color: var(--c-text-secondary);
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  .tile-value {
    font-size: 11px;
    font-weight: 600;
    color: var(--c-text-primary);
  }

  .build-badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 9px;
    font-weight: 700;
    padding: 2px 7px;
    border-radius: 8px;
  }

  .build-badge.official {
    background: rgba(52, 199, 89, 0.15);
    color: #10b981;
    border: 1px solid rgba(52, 199, 89, 0.3);
  }

  .build-badge.dev {
    background: rgba(255, 149, 0, 0.15);
    color: #f59e0b;
    border: 1px solid rgba(255, 149, 0, 0.3);
  }

  .about-actions-row {
    display: flex;
    gap: 8px;
  }

  .about-actions-row :global(.link-icon) {
    margin-left: auto;
    opacity: 0.6;
  }

  .about-section-box {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 10px 12px;
    border-radius: 12px;
    background: var(--c-item-bg, rgba(255, 255, 255, 0.05));
    border: 1px solid var(--c-border);
  }

  .about-box-title {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    font-weight: 600;
    color: var(--c-text-primary);
  }

  .contributors-grid {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-top: 4px;
    width: 100%;
  }

  .contributor-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--c-border);
    text-decoration: none;
    color: var(--c-text-primary);
    transition: all 0.15s ease;
    width: 100%;
    box-sizing: border-box;
  }

  .contributor-card:hover {
    background: rgba(255, 255, 255, 0.08);
    transform: translateY(-1px);
  }

  .contributor-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex: 1;
    overflow: hidden;
    margin-right: 8px;
  }

  .contributor-name {
    font-size: 11px;
    font-weight: 700;
    color: var(--c-text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .contributor-handle {
    font-size: 10px;
    color: var(--c-text-secondary);
    opacity: 0.6;
    font-weight: 500;
    white-space: nowrap;
  }

  .role-pill {
    font-size: 9px;
    font-weight: 700;
    padding: 2px 7px;
    border-radius: 6px;
    background: var(--c-primary, #6366f1);
    color: #ffffff;
    flex-shrink: 0;
  }

  .contributor-link-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 10px;
    font-weight: 600;
    color: var(--c-primary, #6366f1);
    text-decoration: none;
    padding: 4px 8px;
    border-radius: 8px;
    transition: all 0.15s ease;
    margin-top: 4px;
  }

  .contributor-link-btn:hover {
    color: var(--c-text-primary);
  }
</style>
