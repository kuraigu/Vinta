<script lang="ts">
  import type { Snippet, ComponentType } from 'svelte'

  interface Props {
    label: string
    description?: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon?: ComponentType<any>
    flexCol?: boolean
    badge?: string
    disabled?: boolean
    children?: Snippet
  }

  let {
    label,
    description,
    icon: IconComponent,
    flexCol = false,
    badge,
    disabled = false,
    children
  }: Props = $props()
</script>

<div
  class="setting-item"
  class:flex-col={flexCol}
  class:item-disabled={disabled}
>
  <div class="setting-label-group">
    {#if IconComponent}
      <div class="setting-icon-col">
        <IconComponent size={14} class="item-icon" />
      </div>
    {/if}
    <div class="setting-text-col">
      <div class="setting-label">
        <span>{label}</span>
        {#if badge}
          <span class="setting-badge">{badge}</span>
        {/if}
      </div>
      {#if description}
        <span class="setting-description">{description}</span>
      {/if}
    </div>
  </div>

  {#if children}
    <div class="setting-control">
      {@render children()}
    </div>
  {/if}
</div>

<style>
  .setting-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 12px;
    border-radius: 12px;
    background: var(--c-item-bg, rgba(255, 255, 255, 0.05));
    border: 1px solid var(--c-border);
    transition: opacity 0.2s ease;
  }

  .setting-item.item-disabled {
    opacity: 0.55;
  }

  .setting-item.flex-col {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }

  .setting-label-group {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    flex: 1;
    min-width: 0;
    text-align: left;
  }

  .setting-icon-col {
    display: flex;
    align-items: center;
    justify-content: center;
    padding-top: 1px;
    flex-shrink: 0;
  }

  .setting-text-col {
    display: flex;
    flex-direction: column;
    gap: 2px;
    flex: 1;
    min-width: 0;
    text-align: left;
  }

  .setting-label {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    font-weight: 600;
    color: var(--c-text-primary);
    text-align: left;
  }

  .setting-badge {
    font-size: 9px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 2px 6px;
    border-radius: 6px;
    background: var(--c-contrast-bg, rgba(255, 255, 255, 0.08));
    color: var(--c-text-muted);
    border: 1px solid var(--c-border);
  }

  .setting-icon-col :global(.item-icon) {
    color: var(--c-text-secondary);
  }

  .setting-description {
    font-size: 10px;
    line-height: 1.4;
    color: var(--c-text-muted);
    text-align: left;
  }

  .setting-control {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }
</style>
