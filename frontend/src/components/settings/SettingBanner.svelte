<script lang="ts">
  /**
   * Reusable notification and safety banner for settings tabs.
   * Displays an atmospheric pill with an icon, title, and description.
   */
  import type { Snippet, ComponentType } from 'svelte'

  interface Props {
    title?: string
    description?: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon?: ComponentType<any>
    variant?: 'blue' | 'amber' | 'primary' | 'danger'
    children?: Snippet
  }

  let {
    title,
    description,
    icon: IconComponent,
    variant = 'blue',
    children
  }: Props = $props()
</script>

<div class="setting-banner setting-banner--{variant}">
  {#if IconComponent}
    <div class="setting-banner-icon-box">
      <IconComponent size={16} />
    </div>
  {/if}
  <div class="setting-banner-content">
    {#if title}
      <div class="setting-banner-title">{title}</div>
    {/if}
    {#if description}
      <div class="setting-banner-desc">{description}</div>
    {/if}
    {#if children}
      {@render children()}
    {/if}
  </div>
</div>

<style src="../styles/settings/SettingBanner.css"></style>
