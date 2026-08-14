<script lang="ts">
  import type { Snippet, ComponentType } from 'svelte'

  interface Props {
    label: string
    description?: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon?: ComponentType<any>
    flexCol?: boolean
    children?: Snippet
  }

  let {
    label,
    description,
    icon: IconComponent,
    flexCol = false,
    children
  }: Props = $props()
</script>

<div class="setting-item" class:flex-col={flexCol}>
  <div class="setting-label-group">
    <div class="setting-label">
      {#if IconComponent}
        <IconComponent size={14} class="item-icon" />
      {/if}
      <span>{label}</span>
    </div>
    {#if description}
      <span class="setting-description">{description}</span>
    {/if}
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
  }

  .setting-item.flex-col {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }

  .setting-label-group {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .setting-label {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    font-weight: 600;
    color: var(--c-text-primary);
  }

  .setting-label :global(.item-icon) {
    color: var(--c-text-secondary);
  }

  .setting-description {
    font-size: 10px;
    color: var(--c-text-muted);
  }

  .setting-control {
    display: flex;
    align-items: center;
    gap: 8px;
  }
</style>
