<!--
  @component SettingItem
  @description Standard settings row layout. Displays an icon badge, title label,
  and description subtitle alongside a right-aligned control slot (toggle switch,
  select dropdown, or action button).
-->
<script lang="ts">
  import type { Snippet, ComponentType } from 'svelte'
  import { RotateCcw } from 'lucide-svelte'
  import BentoCard from '../BentoCard.svelte'
  import { islandTooltip } from '../../services/islandTooltip'
  import { tStore } from '../../store'

  interface Props {
    label: string
    description?: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon?: ComponentType<any>
    flexCol?: boolean
    badge?: string
    disabled?: boolean
    isDefault?: boolean
    onReset?: () => void
    children?: Snippet
  }

  let {
    label,
    description,
    icon: IconComponent,
    flexCol = false,
    badge,
    disabled = false,
    isDefault = true,
    onReset,
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
        {#if onReset && !isDefault && !disabled}
          <button
            type="button"
            class="setting-reset-btn"
            onclick={(e) => {
              e.stopPropagation()
              onReset()
            }}
            use:islandTooltip={$tStore('common.reset')}
            aria-label={$tStore('common.reset')}
          >
            <RotateCcw size={11} strokeWidth={2.2} />
          </button>
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

  :global(.vinta-bento-card.variant-setting.flex-col)
    .setting-control
    > :global(*) {
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

  .setting-reset-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    padding: 0;
    border-radius: 5px;
    border: none;
    background: transparent;
    color: var(--vinta-text-muted);
    cursor: pointer;
    opacity: 0.72;
    transition: all 0.15s ease;
  }

  .setting-reset-btn:hover {
    opacity: 1;
    background: rgba(255, 255, 255, 0.08);
    color: var(--vinta-accent-primary);
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
