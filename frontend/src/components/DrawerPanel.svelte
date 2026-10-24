<script lang="ts">
  import type { Snippet } from 'svelte'
  import { X } from 'lucide-svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import './styles/DrawerPanel.css'

  interface Props {
    title: string
    tag?: string | number
    closeTooltip?: string
    variant?: 'overlay' | 'embedded'
    class?: string
    onClose: () => void
    children?: Snippet
  }

  let {
    title,
    tag,
    closeTooltip = 'Close',
    variant = 'embedded',
    class: className = '',
    onClose,
    children
  }: Props = $props()
</script>

<section
  class="vinta-drawer-panel variant-{variant} {className}"
  role="dialog"
  aria-label={title}
>
  <header class="vinta-drawer-header">
    <div class="vinta-drawer-title-group">
      <span class="vinta-drawer-title">{title}</span>
      {#if tag !== undefined}
        <span class="vinta-drawer-tag">{tag}</span>
      {/if}
    </div>
    <button
      type="button"
      class="vinta-drawer-close-btn"
      use:islandTooltip={{ text: closeTooltip, align: 'right' }}
      aria-label={closeTooltip}
      onclick={(e) => {
        e.stopPropagation()
        onClose()
      }}
    >
      <X size={13} />
    </button>
  </header>

  <div class="vinta-drawer-body">
    {#if children}
      {@render children()}
    {/if}
  </div>
</section>
