<!--
  @component SettingBanner
  @description Informational banner card placed at the top of settings tabs to highlight
  tips, restart requirements, or active system states.
-->
<script lang="ts">
  import type { Snippet, ComponentType } from 'svelte'
  import BentoCard from '../BentoCard.svelte'
  import CirclePattern from '../patterns/CirclePattern.svelte'

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

<BentoCard
  variant="announcement"
  color={variant}
  class="setting-banner setting-banner--{variant}"
>
  <div class="setting-banner-pattern-wrap" aria-hidden="true">
    <CirclePattern align="right" rows={8} cols={18} />
  </div>

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
</BentoCard>

<style src="../styles/settings/SettingBanner.css"></style>
