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
    padding: 13px 15px;
    border-radius: 16px;
    background:
      linear-gradient(
        145deg,
        rgba(255, 255, 255, 0.055) 0%,
        rgba(255, 255, 255, 0.018) 100%
      ),
      #15161c;
    border: 1px solid rgba(255, 255, 255, 0.08);
    box-shadow:
      inset 0 1px 0 0 rgba(255, 255, 255, 0.1),
      0 3px 12px rgba(0, 0, 0, 0.12);
    transition:
      opacity 0.2s ease,
      border-color 0.2s ease,
      box-shadow 0.2s ease,
      background 0.2s ease;
  }

  .setting-item:hover {
    border-color: rgba(255, 255, 255, 0.15);
    background:
      linear-gradient(
        145deg,
        rgba(255, 255, 255, 0.075) 0%,
        rgba(255, 255, 255, 0.028) 100%
      ),
      #15161c;
    box-shadow:
      inset 0 1px 0 0 rgba(255, 255, 255, 0.14),
      0 5px 16px rgba(0, 0, 0, 0.16);
  }

  .setting-item.item-disabled {
    opacity: 0.42;
    cursor: not-allowed;
  }

  .setting-item.item-disabled:hover {
    border-color: rgba(255, 255, 255, 0.08);
    background:
      linear-gradient(
        145deg,
        rgba(255, 255, 255, 0.055) 0%,
        rgba(255, 255, 255, 0.018) 100%
      ),
      #15161c;
    box-shadow:
      inset 0 1px 0 0 rgba(255, 255, 255, 0.1),
      0 3px 12px rgba(0, 0, 0, 0.12);
  }

  .setting-item.item-disabled .setting-icon-col {
    opacity: 0.6;
  }

  .setting-item.flex-col {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }

  .setting-item.flex-col .setting-control {
    width: 100%;
  }

  .setting-item.flex-col .setting-control > :global(*) {
    width: 100%;
  }

  .setting-label-group {
    display: flex;
    align-items: center;
    gap: 10px;
    flex: 1;
    min-width: 0;
    text-align: left;
  }

  .setting-icon-col {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 9px;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.09);
    box-shadow: inset 0 1px 0 0 rgba(255, 255, 255, 0.08);
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
    font-size: 12.5px;
    font-weight: 600;
    color: var(--vinta-text-primary);
    text-align: left;
    line-height: 1.25;
  }

  .setting-badge {
    font-size: 9px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 2px 7px;
    border-radius: 9999px;
    background: var(--vinta-accent-muted, rgba(56, 189, 248, 0.12));
    color: var(--vinta-accent-primary, var(--vinta-primary, #38bdf8));
    border: 1px solid var(--vinta-border-accent, rgba(56, 189, 248, 0.25));
  }

  .setting-icon-col :global(.item-icon) {
    color: var(--vinta-text-secondary);
  }

  .setting-description {
    font-size: 10.5px;
    line-height: 1.4;
    color: var(--vinta-text-muted);
    text-align: left;
  }

  .setting-control {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }
</style>
