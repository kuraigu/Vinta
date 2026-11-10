<!--
  @component BentoCard
  @description Interactive card surface with subtle hover lift, rounded corners, optional
  header icon/title, and slot content for widgets and controls.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  export type BentoVariant =
    | 'bento'
    | 'setting'
    | 'announcement'
    | 'group'
    | 'sidebar'
  export type BentoColor =
    'blue' | 'amber' | 'primary' | 'danger' | 'emerald' | 'purple'

  interface Props {
    variant?: BentoVariant
    color?: BentoColor | string
    flexCol?: boolean
    disabled?: boolean
    class?: string
    style?: string
    onclick?: (e: MouseEvent) => void
    children?: Snippet
  }

  let {
    variant = 'bento',
    color,
    flexCol = false,
    disabled = false,
    class: className = '',
    style = '',
    onclick,
    children
  }: Props = $props()
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="vinta-bento-card variant-{variant} {color
    ? `color-${color}`
    : ''} {className}"
  class:flex-col={flexCol}
  class:card-disabled={disabled}
  {style}
  {onclick}
>
  {#if children}
    {@render children()}
  {/if}
</div>

<style src="./styles/BentoCard.css"></style>
