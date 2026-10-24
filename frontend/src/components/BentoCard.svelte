<script lang="ts">
  import type { Snippet } from 'svelte'
  import './styles/BentoCard.css'

  export type BentoVariant = 'bento' | 'setting' | 'announcement'

  interface Props {
    variant?: BentoVariant
    flexCol?: boolean
    disabled?: boolean
    class?: string
    onclick?: (e: MouseEvent) => void
    children?: Snippet
  }

  let {
    variant = 'bento',
    flexCol = false,
    disabled = false,
    class: className = '',
    onclick,
    children
  }: Props = $props()
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="vinta-bento-card variant-{variant} {className}"
  class:flex-col={flexCol}
  class:card-disabled={disabled}
  {onclick}
>
  {#if children}
    {@render children()}
  {/if}
</div>
