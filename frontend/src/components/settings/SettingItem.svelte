<!--
  @component SettingItem
  @description Standard settings row layout. Displays a shared IconBadge squircle,
  title label, and description subtitle alongside a right-aligned control slot.
-->
<script lang="ts">
  import type { Snippet, ComponentType } from 'svelte'
  import { RotateCcw } from 'lucide-svelte'
  import BentoCard from '../BentoCard.svelte'
  import IconBadge, { type IconBadgeColor } from '../IconBadge.svelte'
  import { islandTooltip } from '../../services/islandTooltip'
  import { tStore } from '../../store'

  interface Props {
    label: string
    description?: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon?: ComponentType<any>
    badgeColor?: IconBadgeColor
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
    badgeColor = 'surface',
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
      <div class="setting-icon-wrap">
        <IconBadge icon={IconComponent} color={badgeColor} size="md" />
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

<style src="../styles/settings/SettingItem.css"></style>
