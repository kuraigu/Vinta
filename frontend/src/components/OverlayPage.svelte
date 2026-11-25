<!--
  @component OverlayPage
  @description Solid elevated overlay shell rendered above the active Workspace Page
  without unmounting underlying workspace state or stacking a second backdrop-filter blur pass.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import { fade, fly } from 'svelte/transition'
  import { cubicOut } from 'svelte/easing'
  import { X } from 'lucide-svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import { animScaleStore, tStore } from '../store'

  interface Props {
    title?: string
    dataVinta?: string
    onClose: () => void
    headerLeft?: Snippet
    headerActions?: Snippet
    children: Snippet
  }

  let {
    title = '',
    dataVinta = 'overlay-page',
    onClose,
    headerLeft = undefined,
    headerActions = undefined,
    children
  }: Props = $props()
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div
  data-vinta={dataVinta}
  class="overlay-page-shell"
  role="dialog"
  aria-modal="true"
  aria-label={title}
  transition:fade={{
    duration: Math.round(340 * $animScaleStore),
    easing: cubicOut
  }}
  onclick={(e) => {
    if (e.target === e.currentTarget) {
      e.stopPropagation()
      onClose()
    }
  }}
  onkeydown={(e) => {
    if (e.key === 'Escape') {
      e.stopPropagation()
      onClose()
    }
  }}
>
  <div
    class="overlay-page-card"
    in:fly={{
      y: 56,
      duration: Math.round(520 * $animScaleStore),
      opacity: 0.2,
      easing: cubicOut
    }}
    out:fly={{
      y: 40,
      duration: Math.round(360 * $animScaleStore),
      opacity: 0,
      easing: cubicOut
    }}
  >
    <header class="overlay-page-header">
      <div class="overlay-header-left">
        {#if headerLeft}
          {@render headerLeft()}
        {:else if title}
          <h2 class="overlay-title">{title}</h2>
        {/if}
      </div>

      <div class="overlay-header-actions">
        {#if headerActions}
          {@render headerActions()}
        {/if}
        <button
          type="button"
          class="overlay-close-btn"
          use:islandTooltip={{
            text: $tStore('common.close'),
            subtitle: 'Esc',
            position: 'bottom'
          }}
          aria-label={$tStore('common.close')}
          onclick={(e) => {
            e.stopPropagation()
            onClose()
          }}
        >
          <X size={16} strokeWidth={2.2} />
        </button>
      </div>
    </header>

    <div class="overlay-page-body">
      {@render children()}
    </div>
  </div>
</div>

<style src="./styles/OverlayPage.css"></style>
