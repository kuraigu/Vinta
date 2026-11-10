<!--
  @component LanguageTab
  @description Internationalization settings. Lets users switch languages with flag
  icons, infer regional defaults, and open the custom translations folder in Explorer.
-->
<script lang="ts">
  import { Languages, FolderOpen, RefreshCw, Check } from 'lucide-svelte'
  import BentoCard from '../../BentoCard.svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingActionButton from '../SettingActionButton.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import {
    localeStore,
    availableLocalesStore,
    openTranslationsFolder,
    loadCustomLocales,
    tStore
  } from '../../../store'

  let isReloadingLocales = $state(false)
  let localesReloaded = $state(false)
  let reloadTimer: ReturnType<typeof setTimeout> | null = null

  let localeOptions = $derived(
    $availableLocalesStore.map((l) => ({
      value: l.id,
      label: l.name,
      flag: l.flag || l.id
    }))
  )

  async function handleOpenLocalesFolder() {
    await openTranslationsFolder()
  }

  async function handleReloadLocales() {
    isReloadingLocales = true
    try {
      await loadCustomLocales()
      localesReloaded = true
      if (reloadTimer) clearTimeout(reloadTimer)
      reloadTimer = setTimeout(() => {
        localesReloaded = false
      }, 2500)
    } finally {
      isReloadingLocales = false
    }
  }
</script>

<div class="setting-group">
  <BentoCard variant="group">
    <SettingItem
      label={$tStore('settings.language_label')}
      description={$tStore('settings.language_desc')}
      icon={Languages}
      flexCol={true}
    >
      <CustomSelect
        options={localeOptions}
        bind:value={$localeStore}
        fullWidth={true}
      />
    </SettingItem>

    <SettingItem
      label={$tStore('settings.language_section')}
      description={$tStore('settings.locales_help', { locale: 'fr' })}
      icon={FolderOpen}
      flexCol={true}
    >
      <div class="actions-stack">
        <SettingActionButton
          variant="default"
          fullWidth={true}
          onclick={handleOpenLocalesFolder}
        >
          <FolderOpen size={13} />
          {$tStore('settings.open_locales_folder')}
        </SettingActionButton>

        <SettingActionButton
          variant={localesReloaded ? 'success' : 'default'}
          fullWidth={true}
          disabled={isReloadingLocales}
          onclick={handleReloadLocales}
        >
          {#if localesReloaded}
            <Check size={13} />
            {$tStore('settings.reloaded')}
          {:else}
            <RefreshCw size={13} class={isReloadingLocales ? 'spinning' : ''} />
            {$tStore('settings.reload_locales')}
          {/if}
        </SettingActionButton>
      </div>
    </SettingItem>
  </BentoCard>
</div>

<style src="../../styles/settings/LanguageTab.css"></style>
