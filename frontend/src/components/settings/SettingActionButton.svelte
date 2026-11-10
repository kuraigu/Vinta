<!--
  @component SettingActionButton
  @description Clickable settings action button used for secondary actions like
  opening folders in Windows Explorer, testing audio chimes, or resetting data.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  interface Props {
    variant?: 'default' | 'danger' | 'success'
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
    href,
    target,
    rel,
    disabled = false,
    fullWidth = false,
    leftAligned = false,
    onclick,
    children
  }: Props = $props()
</script>

{#if href}
  <a
    {href}
    {target}
    {rel}
    class="setting-btn"
    class:danger={variant === 'danger'}
    class:success={variant === 'success'}
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
    {disabled}
    {onclick}
    class="setting-btn"
    class:danger={variant === 'danger'}
    class:success={variant === 'success'}
    class:full-width={fullWidth}
    class:left-aligned={leftAligned}
  >
    {#if children}
      {@render children()}
    {/if}
  </button>
{/if}

<style src="../styles/settings/SettingActionButton.css"></style>
