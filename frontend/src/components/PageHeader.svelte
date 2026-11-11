<!--
  @component PageHeader
  @description Standardized header bar for expanded overlay pages. Houses navigation back button,
  page title, centered controls, and trailing action capsules.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import BackButton from './BackButton.svelte'

  interface Props {
    title?: string
    showBackButton?: boolean
    onback?: () => void
    position?: 'top' | 'bottom'
    center?: Snippet
    actions?: Snippet
    children?: Snippet
    class?: string
  }

  let {
    title,
    showBackButton = true,
    onback,
    position = 'top',
    center,
    actions,
    children,
    class: className = ''
  }: Props = $props()
</script>

<header class="vinta-page-header position-{position} {className}">
  <div class="header-leading">
    {#if showBackButton}
      <BackButton {onback} />
    {/if}
    {#if title}
      <span class="header-title">{title}</span>
    {/if}
  </div>

  {#if center}
    <div class="header-center">
      {@render center()}
    </div>
  {/if}

  {#if actions}
    <div class="header-actions">
      {@render actions()}
    </div>
  {:else if showBackButton}
    <div class="header-spacer" aria-hidden="true"></div>
  {/if}

  {#if children}
    {@render children()}
  {/if}
</header>

<style src="./styles/PageHeader.css"></style>
