<!--
  @component IconBadge
  @description Unified gradient squircle icon badge used across sidebars, settings rows,
  onboarding cards, and spotlight carousel items.
-->
<script lang="ts" module>
  export type IconBadgeColor =
    | 'blue'
    | 'cyan'
    | 'purple'
    | 'emerald'
    | 'amber'
    | 'rose'
    | 'indigo'
    | 'slate'
    | 'teal'
    | 'surface'

  export type IconBadgeSize = 'sm' | 'md' | 'lg'
</script>

<script lang="ts">
  import type { ComponentType, Snippet } from 'svelte'

  interface Props {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon?: ComponentType<any>
    color?: IconBadgeColor
    size?: IconBadgeSize
    iconSize?: number
    strokeWidth?: number
    customBg?: string
    customColor?: string
    class?: string
    children?: Snippet
  }

  let {
    icon: IconComponent,
    color = 'blue',
    size = 'md',
    iconSize,
    strokeWidth = 2.2,
    customBg,
    customColor,
    class: className = '',
    children
  }: Props = $props()

  let resolvedIconSize = $derived(
    iconSize ?? (size === 'sm' ? 13 : size === 'lg' ? 18 : 15)
  )
</script>

<span
  class="vinta-icon-badge size-{size} color-{color} {className}"
  style:background={customBg}
  style:color={customColor}
>
  {#if IconComponent}
    <IconComponent size={resolvedIconSize} {strokeWidth} />
  {:else if children}
    {@render children()}
  {/if}
</span>

<style src="./styles/IconBadge.css"></style>

