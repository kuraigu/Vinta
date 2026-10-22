<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 },
    expanded: { w: 580, h: 480, r: 32 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import SidebarLayout from '../components/SidebarLayout.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import SettingsSidebar from '../components/settings/SettingsSidebar.svelte'
  import AppearanceTab from '../components/settings/tabs/AppearanceTab.svelte'
  import AppsModulesTab from '../components/settings/tabs/AppsModulesTab.svelte'
  import BehaviorTab from '../components/settings/tabs/BehaviorTab.svelte'
  import ShortcutsTab from '../components/settings/tabs/ShortcutsTab.svelte'
  import NotificationsTab from '../components/settings/tabs/NotificationsTab.svelte'
  import ClockTab from '../components/settings/tabs/ClockTab.svelte'
  import WeatherTab from '../components/settings/tabs/WeatherTab.svelte'
  import SystemTab from '../components/settings/tabs/SystemTab.svelte'
  import UpdatesTab from '../components/settings/tabs/UpdatesTab.svelte'
  import AboutTab from '../components/settings/tabs/AboutTab.svelte'
  import { activeSettingsTabStore } from '../store'

  let activeTab = $derived($activeSettingsTabStore)
</script>

<SidebarLayout dataVinta="settings-page">
  {#snippet sidebar()}
    <SettingsSidebar />
  {/snippet}

  {#snippet content()}
    <ScrollView class="settings-content-scroll">
      <div class="tab-content">
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
        {:else if activeTab === 'system'}
          <SystemTab />
        {:else if activeTab === 'updates'}
          <UpdatesTab />
        {:else if activeTab === 'about'}
          <AboutTab />
        {/if}
      </div>
    </ScrollView>
  {/snippet}
</SidebarLayout>

<style src="./styles/SettingsPage.css"></style>
