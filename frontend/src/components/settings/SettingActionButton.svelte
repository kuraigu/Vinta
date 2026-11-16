<!--
  @component SettingActionButton
  @description Clickable settings action button used for secondary actions like
  opening folders in Windows Explorer, testing audio chimes, or resetting data,
  with unified loading, success, and error feedback states.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import { Loader2, Check, AlertCircle } from 'lucide-svelte'

  export type SettingButtonStatus = 'idle' | 'loading' | 'success' | 'error'

  interface Props {
    variant?: 'default' | 'danger' | 'success' | 'error'
    status?: SettingButtonStatus
    loadingLabel?: string
    successLabel?: string
    errorLabel?: string
    href?: string
    target?: string
    rel?: string
    disabled?: boolean
    fullWidth?: boolean
    leftAligned?: boolean
    onclick?: (e: MouseEvent) => void
    children?: Snippet
  }

  let {
    variant = 'default',
    status = 'idle',
    loadingLabel,
    successLabel,
    errorLabel,
    href,
    target,
    rel,
    disabled = false,
    fullWidth = false,
    leftAligned = false,
    onclick,
    children
  }: Props = $props()

  let effectiveDanger = $derived(
    status === 'error' || variant === 'danger' || variant === 'error'
  )
  let effectiveSuccess = $derived(status === 'success' || variant === 'success')
  let effectiveDisabled = $derived(disabled || status === 'loading')
</script>

{#if href}
  <a
    {href}
    {target}
    {rel}
    class="setting-btn"
    class:danger={effectiveDanger}
    class:success={effectiveSuccess}
    class:full-width={fullWidth}
    class:left-aligned={leftAligned}
  >
    {#if children}
      {@render children()}
    {/if}
  </a>
{:else}
  <button
    type="button"
    disabled={effectiveDisabled}
    {onclick}
    class="setting-btn"
    class:danger={effectiveDanger}
    class:success={effectiveSuccess}
    class:is-loading={status === 'loading'}
    class:full-width={fullWidth}
    class:left-aligned={leftAligned}
  >
    {#if status === 'loading'}
      <Loader2 size={13} class="setting-btn-spin" />
      {#if loadingLabel}
        <span>{loadingLabel}</span>
      {:else if children}
        {@render children()}
      {/if}
    {:else if status === 'success'}
      <Check size={13} />
      {#if successLabel}
        <span>{successLabel}</span>
      {:else if children}
        {@render children()}
      {/if}
    {:else if status === 'error'}
      <AlertCircle size={13} />
      {#if errorLabel}
        <span>{errorLabel}</span>
      {:else if children}
        {@render children()}
      {/if}
    {:else if children}
      {@render children()}
    {/if}
  </button>
{/if}

<style src="../styles/settings/SettingActionButton.css"></style>
