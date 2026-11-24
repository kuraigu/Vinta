<!--
  @fileoverview Primary settings hub hosting specialized preference tabs.

  Renders the responsive settings sidebar navigation, top search bar (Ctrl+F / /),
  and mounts active tab views for appearance, apps & widgets, clock, weather,
  language, and system.
-->
<script lang="ts" module>
  import { ISLAND_WIDTH_TIERS, type IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 },
    expanded: { w: ISLAND_WIDTH_TIERS.wide, h: 495, r: 30 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { setContext, tick } from 'svelte'
  import SidebarLayout from '../components/SidebarLayout.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import SearchInput from '../components/SearchInput.svelte'
  import SettingsSidebar from '../components/settings/SettingsSidebar.svelte'
  import AppearanceTab from '../components/settings/tabs/AppearanceTab.svelte'
  import AppsModulesTab from '../components/settings/tabs/AppsModulesTab.svelte'
  import BehaviorTab from '../components/settings/tabs/BehaviorTab.svelte'
  import ShortcutsTab from '../components/settings/tabs/ShortcutsTab.svelte'
  import NotificationsTab from '../components/settings/tabs/NotificationsTab.svelte'
  import ClockTab from '../components/settings/tabs/ClockTab.svelte'
  import WeatherTab from '../components/settings/tabs/WeatherTab.svelte'
  import LanguageTab from '../components/settings/tabs/LanguageTab.svelte'
  import AccountTab from '../components/settings/tabs/AccountTab.svelte'
  import SystemTab from '../components/settings/tabs/SystemTab.svelte'
  import UpdatesTab from '../components/settings/tabs/UpdatesTab.svelte'
  import AboutTab from '../components/settings/tabs/AboutTab.svelte'
  import { activeSettingsTabStore, tStore } from '../store'

  let { expanded: _expanded = true }: { expanded?: boolean } = $props()
  let activeTab = $derived($activeSettingsTabStore)

  let searchQuery = $state('')
  let isSearching = $derived(searchQuery.trim().length > 0)
  let tabContentEl = $state<HTMLElement | null>(null)
  let hasSearchMatches = $state(true)

  setContext('settingsSearchQuery', () => searchQuery)

  $effect(() => {
    const q = searchQuery.trim()
    if (!q || !tabContentEl) {
      hasSearchMatches = true
      return
    }
    void tick().then(() => {
      if (!tabContentEl) return
      hasSearchMatches =
        tabContentEl.querySelectorAll('.vinta-bento-card.variant-setting')
          .length > 0
    })
  })
</script>

<SidebarLayout dataVinta="settings-page">
  {#snippet sidebar()}
    <SettingsSidebar />
  {/snippet}

  {#snippet content()}
    <div class="settings-search-bar">
      <SearchInput
        bind:value={searchQuery}
        placeholder={$tStore('settings.search_placeholder')}
      />
    </div>

    <ScrollView class="settings-content-scroll">
      <div
        bind:this={tabContentEl}
        class="tab-content"
        class:is-searching={isSearching}
      >
        {#if isSearching}
          <div class="settings-search-results">
            <AppearanceTab />
            <AppsModulesTab />
            <BehaviorTab />
            <ShortcutsTab />
            <NotificationsTab />
            <ClockTab />
            <WeatherTab />
            <LanguageTab />
            <AccountTab />
            <SystemTab />
            <UpdatesTab />
          </div>
          {#if !hasSearchMatches}
            <div class="settings-search-empty">
              {$tStore('settings.no_matching_settings', { query: searchQuery })}
            </div>
          {/if}
        {:else}
          {#key activeTab}
            <div class="tab-pane-transition">
              {#if activeTab === 'appearance'}
                <AppearanceTab />
              {:else if activeTab === 'modules'}
                <AppsModulesTab />
              {:else if activeTab === 'behavior'}
                <BehaviorTab />
              {:else if activeTab === 'shortcuts'}
                <ShortcutsTab />
              {:else if activeTab === 'notifications'}
                <NotificationsTab />
              {:else if activeTab === 'clock'}
                <ClockTab />
              {:else if activeTab === 'weather'}
                <WeatherTab />
              {:else if activeTab === 'language'}
                <LanguageTab />
              {:else if activeTab === 'account'}
                <AccountTab />
              {:else if activeTab === 'system'}
                <SystemTab />
              {:else if activeTab === 'updates'}
                <UpdatesTab />
              {:else if activeTab === 'about'}
                <AboutTab />
              {/if}
            </div>
          {/key}
        {/if}
      </div>
    </ScrollView>
  {/snippet}
</SidebarLayout>

<style src="./styles/SettingsPage.css"></style>
