<!--
  @fileoverview Developer Hub page surfacing GitHub contributions, PRs, issues,
  notifications, and AI provider telemetry (Claude, OpenAI, Gemini, Antigravity).
-->
<script lang="ts" module>
  import { ISLAND_WIDTH_TIERS, type IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 230, h: 42, r: 16 },
    expanded: { w: ISLAND_WIDTH_TIERS.wide, h: 495, r: 28 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { onMount, getContext } from 'svelte'
  import {
    Bell,
    ChevronRight,
    Cpu,
    ExternalLink,
    Flame,
    GitPullRequest,
    CircleDot,
    RefreshCw,
    Settings,
    Terminal
  } from 'lucide-svelte'
  import IslandPage from '../components/IslandPage.svelte'
  import PageHeader from '../components/PageHeader.svelte'
  import HeaderActionButton from '../components/HeaderActionButton.svelte'
  import BentoCard from '../components/BentoCard.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import ToggleSwitch from '../components/ToggleSwitch.svelte'
  import HorizontalTabSelection, {
    type TabItem
  } from '../components/HorizontalTabSelection.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import SettingsPage, { config as settingsConfig } from './SettingsPage.svelte'
  import BrandLogo from '../components/BrandLogo.svelte'
  import { activeSettingsTabStore, tStore } from '../store'
  import {
    githubStatusStore,
    githubDashboardStore,
    aiUsageStore,
    initDevIntegrations,
    startGitHubDeviceAuth,
    completeGitHubDeviceAuth,
    refreshGitHubDashboard,
    refreshAIUsage
  } from '../stores/devIntegrationsStore'
  import { OpenURL } from '../../bindings/Vinta/backend/app.js'

  let { expanded = false }: { expanded?: boolean } = $props()

  const navigate = getContext<
    | ((page: unknown, pageConfig: IslandConfig, expand?: boolean) => void)
    | undefined
  >('navigate')

  let activeTab = $state<'github' | 'ai'>('github')
  let ghSubTab = $state<'prs' | 'issues' | 'notifications'>('prs')
  let isRefreshing = $state(false)
  let isSigningIn = $state(false)
  let deviceUserCode = $state('')
  let signInError = $state('')
  let includeNotifications = $state(false)
  let includePrivateRepos = $state(false)

  const mainTabs: TabItem<'github' | 'ai'>[] = $derived([
    {
      id: 'github',
      label: $tStore('devhub.tab_github'),
      icon: Terminal,
      tooltip: ''
    },
    { id: 'ai', label: $tStore('devhub.tab_ai'), icon: Cpu, tooltip: '' }
  ])

  const ghSubTabs: TabItem<'prs' | 'issues' | 'notifications'>[] = $derived([
    {
      id: 'prs',
      label: $tStore('devhub.sub_prs'),
      badge: $githubDashboardStore?.pullRequests.length || undefined,
      tooltip: ''
    },
    {
      id: 'issues',
      label: $tStore('devhub.sub_issues'),
      badge: $githubDashboardStore?.assignedIssues.length || undefined,
      tooltip: ''
    },
    {
      id: 'notifications',
      label: $tStore('devhub.sub_notifications'),
      badge: $githubDashboardStore?.unreadCount || undefined,
      tooltip: ''
    }
  ])

  const activeItems = $derived.by(() => {
    const dash = $githubDashboardStore
    if (!dash) return []
    if (ghSubTab === 'prs') return dash.pullRequests
    if (ghSubTab === 'issues') return dash.assignedIssues
    return dash.notifications
  })

  const todayContributions = $derived.by(() => {
    const days = $githubDashboardStore?.contributions
    if (!days || days.length === 0) return 0
    return days[days.length - 1]?.count ?? 0
  })

  function formatContributionTooltip(dateStr: string, count: number): string {
    const parsed = new Date(`${dateStr}T00:00:00`)
    const label = isNaN(parsed.getTime())
      ? dateStr
      : parsed.toLocaleDateString(undefined, {
          month: 'short',
          day: 'numeric',
          year: 'numeric'
        })
    if (count === 0) return `${label}: ${$tStore('devhub.no_contributions')}`
    return `${label}: ${count} ${count === 1 ? $tStore('devhub.contribution_one') : $tStore('devhub.contribution_many')}`
  }

  onMount(() => {
    void initDevIntegrations()
  })

  function openAccountSettings() {
    $activeSettingsTabStore = 'account'
    if (navigate) {
      navigate(SettingsPage, settingsConfig, true)
    }
  }

  function openExternal(url: string) {
    if (!url) return
    if (typeof OpenURL === 'function') {
      void OpenURL(url)
    } else {
      window.open(url, '_blank', 'noopener,noreferrer')
    }
  }

  async function handleRefresh() {
    isRefreshing = true
    await Promise.all([refreshGitHubDashboard(true), refreshAIUsage(true)])
    isRefreshing = false
  }

  async function handleQuickGitHubSignIn() {
    isSigningIn = true
    signInError = ''
    deviceUserCode = ''
    const start = await startGitHubDeviceAuth({
      includeNotifications,
      includePrivateRepos
    })
    if (!start.ok || !start.info) {
      isSigningIn = false
      signInError = start.error || $tStore('account.github_failed')
      return
    }
    deviceUserCode = start.info.userCode
    const done = await completeGitHubDeviceAuth(
      start.info.deviceCode,
      start.info.interval
    )
    isSigningIn = false
    deviceUserCode = ''
    if (!done.ok && done.error) {
      signInError = done.error
    }
  }
</script>

<IslandPage {expanded}>
  {#snippet minimal()}
    <div class="devhub-minimal-pill">
      <BrandLogo name="github" size={16} />
      {#if $githubStatusStore.connected && $githubDashboardStore}
        <span class="devhub-min-metric">
          <Flame size={12} />
          <span>{$githubDashboardStore.currentStreak}d</span>
        </span>
        <span class="devhub-min-sep">·</span>
        <span class="devhub-min-metric">
          <GitPullRequest size={12} />
          <span>{$githubDashboardStore.pullRequests.length}</span>
        </span>
        {#if $githubDashboardStore.unreadCount > 0}
          <span class="devhub-min-badge"
            >{$githubDashboardStore.unreadCount}</span
          >
        {/if}
      {:else}
        <span class="devhub-min-label">{$tStore('devhub.title')}</span>
      {/if}
    </div>
  {/snippet}

  {#snippet expandedContent()}
    <div class="devhub-page" data-vinta="devhub-page">
      <PageHeader title={$tStore('devhub.title')}>
        {#snippet center()}
          <HorizontalTabSelection
            tabs={mainTabs}
            bind:value={activeTab}
            ariaLabel={$tStore('devhub.title')}
          />
        {/snippet}
        {#snippet actions()}
          <HeaderActionButton
            variant="ghost"
            disabled={isRefreshing}
            ariaLabel={$tStore('devhub.refresh')}
            tooltip={{ text: $tStore('devhub.refresh'), position: 'bottom' }}
            onclick={handleRefresh}
          >
            <RefreshCw size={16} />
          </HeaderActionButton>
          <HeaderActionButton
            variant="ghost"
            ariaLabel={$tStore('settings.tabs.account')}
            tooltip={{
              text: $tStore('settings.tabs.account'),
              position: 'bottom'
            }}
            onclick={openAccountSettings}
          >
            <Settings size={16} />
          </HeaderActionButton>
        {/snippet}
      </PageHeader>

      <ScrollView class="devhub-scroll">
        {#if activeTab === 'github'}
          {#if !$githubStatusStore.connected}
            <BentoCard variant="bento" class="devhub-empty-card">
              <article class="devhub-empty-article">
                <header class="devhub-empty-header">
                  <BrandLogo name="github" size={24} />
                  <h3 class="devhub-empty-title">
                    {$tStore('devhub.connect_title')}
                  </h3>
                </header>
                <p class="devhub-empty-desc">
                  {$tStore('devhub.connect_desc')}
                </p>
                <div class="devhub-scope-opts">
                  <div class="devhub-scope-row">
                    <div class="devhub-scope-text">
                      <span class="devhub-scope-label"
                        >{$tStore('account.github_opt_notifications_label')}</span
                      >
                      <span class="devhub-scope-desc"
                        >{$tStore('account.github_opt_notifications_desc')}</span
                      >
                    </div>
                    <ToggleSwitch
                      bind:checked={includeNotifications}
                      ariaLabel={$tStore('account.github_opt_notifications_label')}
                    />
                  </div>
                  <div class="devhub-scope-row">
                    <div class="devhub-scope-text">
                      <span class="devhub-scope-label"
                        >{$tStore('account.github_opt_private_label')}</span
                      >
                      <span class="devhub-scope-desc"
                        >{$tStore('account.github_opt_private_desc')}</span
                      >
                    </div>
                    <ToggleSwitch
                      bind:checked={includePrivateRepos}
                      ariaLabel={$tStore('account.github_opt_private_label')}
                    />
                  </div>
                </div>
                <footer class="devhub-empty-actions">
                  <button
                    type="button"
                    class="devhub-gh-btn"
                    disabled={isSigningIn}
                    onclick={handleQuickGitHubSignIn}
                  >
                    <BrandLogo name="github" size={16} />
                    <span>
                      {isSigningIn
                        ? $tStore('account.waiting_github')
                        : $tStore('account.sign_in_github')}
                    </span>
                  </button>
                </footer>
                {#if deviceUserCode}
                  <p class="devhub-code-hint">
                    {$tStore('account.device_code_label')}:
                    <code>{deviceUserCode}</code>
                  </p>
                {/if}
                {#if signInError}
                  <p class="devhub-error" role="alert">{signInError}</p>
                {/if}
              </article>
            </BentoCard>
          {:else if $githubDashboardStore}
            <div class="devhub-stack">
              <BentoCard variant="bento" class="devhub-stats-card">
                <div class="devhub-stats-header">
                  <div class="devhub-user-group">
                    <span class="devhub-user"
                      >@{$githubDashboardStore.username}</span
                    >
                    <span class="devhub-user-sub">
                      {$githubDashboardStore.totalRecent.toLocaleString()}
                      {$tStore('devhub.contributions_year')} ·
                      {todayContributions}
                      {$tStore('devhub.today_short')}
                    </span>
                  </div>
                  <div class="devhub-kpis">
                    <span class="devhub-kpi">
                      <Flame size={12} />
                      {$githubDashboardStore.currentStreak}d
                    </span>
                    <span class="devhub-kpi">
                      <Bell size={12} />
                      {$githubDashboardStore.unreadCount}
                    </span>
                  </div>
                </div>
                {#if $githubDashboardStore.contributions.length > 0}
                  <div
                    class="devhub-heatmap"
                    aria-label={$tStore('devhub.heatmap_label')}
                  >
                    {#each $githubDashboardStore.contributions as day (day.date)}
                      <span
                        class="heatmap-cell level-{day.level}"
                        use:islandTooltip={formatContributionTooltip(
                          day.date,
                          day.count
                        )}
                        aria-label={formatContributionTooltip(
                          day.date,
                          day.count
                        )}
                      ></span>
                    {/each}
                  </div>
                  <div class="devhub-heatmap-footer">
                    <span>{$tStore('devhub.last_12_weeks')}</span>
                    <div class="devhub-legend" aria-hidden="true">
                      <span>{$tStore('devhub.legend_less')}</span>
                      <span class="heatmap-cell level-0 legend-dot"></span>
                      <span class="heatmap-cell level-1 legend-dot"></span>
                      <span class="heatmap-cell level-2 legend-dot"></span>
                      <span class="heatmap-cell level-3 legend-dot"></span>
                      <span class="heatmap-cell level-4 legend-dot"></span>
                      <span>{$tStore('devhub.legend_more')}</span>
                    </div>
                  </div>
                {/if}
              </BentoCard>

              <HorizontalTabSelection
                tabs={ghSubTabs}
                bind:value={ghSubTab}
                colorVariant="primary"
                ariaLabel={$tStore('devhub.tab_github')}
              />

              {#if activeItems.length === 0}
                <BentoCard variant="bento" class="devhub-list-empty">
                  <span>{$tStore('devhub.empty_list')}</span>
                </BentoCard>
              {:else}
                <div class="devhub-item-list">
                  {#each activeItems as item (item.id)}
                    <button
                      type="button"
                      class="devhub-item-row"
                      onclick={() => openExternal(item.url)}
                    >
                      <span class="devhub-item-icon">
                        {#if ghSubTab === 'prs'}
                          <GitPullRequest size={14} />
                        {:else if ghSubTab === 'issues'}
                          <CircleDot size={14} />
                        {:else}
                          <Bell size={14} />
                        {/if}
                      </span>
                      <span class="devhub-item-body">
                        <span class="devhub-item-title">{item.title}</span>
                        <span class="devhub-item-repo">{item.repo}</span>
                      </span>
                      <ChevronRight size={14} class="devhub-ext-icon" />
                    </button>
                  {/each}
                </div>
              {/if}
            </div>
          {:else}
            <div
              class="devhub-stack"
              aria-busy="true"
              aria-label={$tStore('devhub.loading')}
            >
              <BentoCard variant="bento" class="devhub-stats-card">
                <div class="devhub-stats-header">
                  <span class="skeleton-bar sk-user"></span>
                  <span class="skeleton-bar sk-kpi"></span>
                </div>
                <div class="skeleton-heatmap"></div>
              </BentoCard>
              <div class="devhub-item-list">
                {#each [0, 1, 2] as idx (idx)}
                  <div class="devhub-skeleton-row">
                    <span class="skeleton-circle"></span>
                    <div class="skeleton-lines">
                      <span class="skeleton-bar sk-title"></span>
                      <span class="skeleton-bar sk-sub"></span>
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        {:else}
          <BentoCard variant="group" class="devhub-ai-list">
            {#each $aiUsageStore as provider (provider.id)}
              <div class="devhub-ai-row">
                <div class="devhub-ai-left">
                  <span
                    class="devhub-ai-dot"
                    class:connected={provider.configured}
                    aria-hidden="true"
                  ></span>
                  <div class="devhub-ai-meta">
                    <span class="devhub-ai-name">{provider.name}</span>
                    <span class="devhub-ai-host">{provider.detail}</span>
                  </div>
                </div>
                <div class="devhub-ai-right">
                  <span
                    class="devhub-ai-state"
                    class:connected={provider.configured}
                  >
                    {provider.status}
                  </span>
                  {#if !provider.configured}
                    <button
                      type="button"
                      class="devhub-ai-btn"
                      onclick={openAccountSettings}
                    >
                      {$tStore('devhub.add_key')}
                    </button>
                  {/if}
                  <button
                    type="button"
                    class="devhub-ai-icon-btn"
                    aria-label="{$tStore(
                      'devhub.open_usage'
                    )} ({provider.name})"
                    onclick={() => openExternal(provider.dashboardUrl)}
                  >
                    <ExternalLink size={14} />
                  </button>
                </div>
              </div>
            {/each}
          </BentoCard>
        {/if}
      </ScrollView>
    </div>
  {/snippet}
</IslandPage>

<style src="./styles/DevHubPage.css"></style>
