<script lang="ts">
  import {
    Languages,
    FolderOpen,
    RefreshCw,
    Check
  } from 'lucide-svelte'
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
</div>

<style>
  .setting-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .actions-stack {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;
  }

  .actions-stack :global(button) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
  }

  :global(.spinning) {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>

