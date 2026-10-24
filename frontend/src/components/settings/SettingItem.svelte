<script lang="ts">
  import type { Snippet, ComponentType } from 'svelte'
  import BentoCard from '../BentoCard.svelte'

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

<BentoCard variant="setting" {flexCol} {disabled}>
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
</BentoCard>

<style>
  :global(.vinta-bento-card.variant-setting.flex-col) .setting-control {
    width: 100%;
  }

  :global(.vinta-bento-card.variant-setting.flex-col) .setting-control > :global(*) {
    width: 100%;
  }

  :global(.vinta-bento-card.variant-setting.card-disabled) .setting-icon-col {
    opacity: 0.6;
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
    border: 1px solid var(--vinta-border);
    box-shadow: var(--vinta-contrast-border);
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
    background: var(--vinta-accent-muted);
    color: var(--vinta-accent-primary);
    border: 1px solid var(--vinta-border-accent);
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
