<!--
  @component EmptyState
  @description Clean placeholder view with icon, title, description, and optional action button
  shown when a list or feed has no items.
-->
<script lang="ts">
  import type { ComponentType } from 'svelte'

  interface Props {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon?: ComponentType<any>
    title: string
    description?: string
    actionLabel?: string
    onAction?: () => void
    seamless?: boolean
  }

  let {
    icon: IconComponent,
    title,
    description = '',
    actionLabel = '',
    onAction,
    seamless = false
  }: Props = $props()
</script>

<div class="empty-state-container" class:seamless>
  {#if IconComponent}
    <div class="empty-state-icon">
      <IconComponent size={32} />
    </div>
  {/if}
  <h3 class="empty-state-title">{title}</h3>
  {#if description}
    <p class="empty-state-description">{description}</p>
  {/if}
  {#if actionLabel && onAction}
    <button type="button" class="empty-state-action" onclick={onAction}>
      {actionLabel}
    </button>
  {/if}
</div>

<style src="./styles/EmptyState.css"></style>
