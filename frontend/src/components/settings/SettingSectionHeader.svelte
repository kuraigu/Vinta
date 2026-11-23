<!--
  @component SettingSectionHeader
  @description Unified section header for Settings tabs with optional leading icon and trailing actions.
-->
<script lang="ts">
  import type { ComponentType, Snippet } from 'svelte'

  interface Props {
    title: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon?: ComponentType<any>
    actions?: Snippet
  }

  let { title, icon: IconComponent, actions }: Props = $props()
</script>

<div class="setting-section-header">
  <div class="setting-section-title">
    {#if IconComponent}
      <IconComponent size={13} strokeWidth={2.2} />
    {/if}
    <span>{title}</span>
  </div>
  {#if actions}
    <div class="setting-section-actions">
      {@render actions()}
    </div>
  {/if}
</div>

<style src="../styles/settings/SettingSectionHeader.css"></style>
