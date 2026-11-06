<!--
  @component EditorModal
  @description Centered modal dialog wrapper with backdrop blur, title bar, action buttons,
  and Escape key dismissal.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import { X } from 'lucide-svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import './styles/EditorModal.css'

  interface Props {
    isOpen: boolean
    title: string
    maxWidth?: string
    maxHeight?: string
    onClose: () => void
    footer?: Snippet
    children?: Snippet
  }

  let {
    isOpen,
    title,
    maxWidth = '380px',
    maxHeight = '90%',
    onClose,
    footer,
    children
  }: Props = $props()

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      onClose()
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div
    class="vinta-modal-backdrop"
    onclick={handleBackdropClick}
    role="presentation"
  >
    <div
      class="vinta-modal-dialog"
      style:max-width={maxWidth}
      style:max-height={maxHeight}
      role="dialog"
      tabindex="-1"
      aria-modal="true"
    >
      <header class="vinta-modal-header">
        <h3 class="vinta-modal-title">{title}</h3>
        <button
          type="button"
          class="vinta-modal-close-btn"
          onclick={onClose}
          use:islandTooltip={'Close'}
          aria-label="Close dialog"
        >
          <X size={15} />
        </button>
      </header>

      <div class="vinta-modal-body">
        {#if children}
          {@render children()}
        {/if}
      </div>

      {#if footer}
        <footer class="vinta-modal-footer">
          {@render footer()}
        </footer>
      {/if}
    </div>
  </div>
{/if}
