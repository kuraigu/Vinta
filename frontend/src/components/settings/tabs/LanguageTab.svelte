<!--
  @component LanguageTab
  @description Internationalization settings. Lets users switch languages with flag
  icons, infer regional defaults, and open the custom translations folder in Explorer.
-->
<script lang="ts">
  import { Languages, FolderOpen, RefreshCw } from 'lucide-svelte'
  import BentoCard from '../../BentoCard.svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingActionButton, {
    type SettingButtonStatus
  } from '../SettingActionButton.svelte'
  import CustomSelect from '../../CustomSelect.svelte'
  import {
    localeStore,
    availableLocalesStore,
    customCatalogsStore,
    getLocaleMetadata,
    openTranslationsFolder,
    loadCustomLocales,
    tStore
  } from '../../../store'

  let reloadStatus = $state<SettingButtonStatus>('idle')
  let reloadTimer: ReturnType<typeof setTimeout> | null = null

  let localeOptions = $derived(
    $availableLocalesStore.map((l) => ({
      value: l.id,
      label: l.name,
      flag: l.flag || l.id
    }))
  )

  let activeLocaleMeta = $derived(
    getLocaleMetadata($localeStore, $customCatalogsStore)
  )

  async function handleOpenLocalesFolder() {
    await openTranslationsFolder()
  }

  async function handleReloadLocales() {
    reloadStatus = 'loading'
    try {
      await loadCustomLocales()
      reloadStatus = 'success'
      if (reloadTimer) clearTimeout(reloadTimer)
      reloadTimer = setTimeout(() => {
        reloadStatus = 'idle'
      }, 2500)
    } catch {
      reloadStatus = 'error'
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
      isDefault={$localeStore === localeStore.defaultValue}
      onReset={() => localeStore.reset()}
    >
      <CustomSelect
        options={localeOptions}
        bind:value={$localeStore}
        fullWidth={true}
      />
      {#if activeLocaleMeta && activeLocaleMeta.authors.length > 0}
        <p class="locale-authors-caption">
          {$tStore('settings.translated_by', {
            authors: activeLocaleMeta.authors.join(', ')
          })}
        </p>
      {/if}
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
          variant="default"
          status={reloadStatus}
          successLabel={$tStore('settings.reloaded')}
          fullWidth={true}
          onclick={handleReloadLocales}
        >
          <RefreshCw size={13} />
          {$tStore('settings.reload_locales')}
        </SettingActionButton>
      </div>
    </SettingItem>
  </BentoCard>
</div>

<style>
  .locale-authors-caption {
    margin: 6px 0 0;
    font-size: 11px;
    color: var(--vinta-text-secondary);
  }
</style>
