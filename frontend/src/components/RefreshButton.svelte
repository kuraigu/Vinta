<script lang="ts">
  import { RefreshCw } from 'lucide-svelte'
  import { islandTooltip } from '../services/islandTooltip'

  interface Props {
    onclick?: (e: MouseEvent) => void
    loading?: boolean
    title?: string
    size?: number
  }

  let {
    onclick,
    loading = false,
    title = 'Refresh',
    size = 14
  }: Props = $props()

  function handleClick(e: MouseEvent) {
    if (onclick) {
      onclick(e)
    }
  }
</script>

<button
  type="button"
  class="refresh-btn"
  class:spinning={loading}
  onclick={handleClick}
  use:islandTooltip={title}
  aria-label={title}
>
  <RefreshCw {size} />
</button>

<style>
  .refresh-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 30px;
    height: 30px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.05);
    border: none;
    color: var(--vinta-text-primary, #ffffff);
    cursor: pointer;
    transition: all 0.18s cubic-bezier(0.2, 0.8, 0.2, 1);
    box-sizing: border-box;
    flex-shrink: 0;
  }

  .refresh-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #ffffff;
    transform: scale(1.04);
  }

  .refresh-btn:active {
    transform: scale(0.96);
  }

  .refresh-btn.spinning :global(svg) {
    animation: vinta-spin 0.9s linear infinite;
  }

  @keyframes vinta-spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
