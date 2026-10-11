<script lang="ts">
  import type { Snippet } from 'svelte'

  interface Props {
    variant?: 'default' | 'danger' | 'success'
    href?: string
    target?: string
    rel?: string
    disabled?: boolean
    fullWidth?: boolean
    leftAligned?: boolean
    onclick?: (e: MouseEvent) => void
    children?: Snippet
  }

  let {
    variant = 'default',
    href,
    target,
    rel,
    disabled = false,
    fullWidth = false,
    leftAligned = false,
    onclick,
    children
  }: Props = $props()
</script>

{#if href}
  <a
    {href}
    {target}
    {rel}
    class="setting-btn"
    class:danger={variant === 'danger'}
    class:success={variant === 'success'}
    class:full-width={fullWidth}
    class:left-aligned={leftAligned}
  >
    {#if children}
      {@render children()}
    {/if}
  </a>
{:else}
  <button
    type="button"
    {disabled}
    {onclick}
    class="setting-btn"
    class:danger={variant === 'danger'}
    class:success={variant === 'success'}
    class:full-width={fullWidth}
    class:left-aligned={leftAligned}
  >
    {#if children}
      {@render children()}
    {/if}
  </button>
{/if}

<style>
  .setting-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 7px;
    height: 32px;
    padding: 0 14px;
    border-radius: 9px;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid var(--c-border);
    box-shadow: inset 0 1px 0 0 rgba(255, 255, 255, 0.06);
    color: var(--c-text-primary);
    font-size: 11.5px;
    font-weight: 600;
    cursor: pointer;
    white-space: nowrap;
    text-decoration: none;
    box-sizing: border-box;
    transition: all 0.15s ease;
  }

  .setting-btn:hover {
    background: rgba(255, 255, 255, 0.12);
    border-color: var(--c-border-hover, rgba(255, 255, 255, 0.18));
  }

  .setting-btn:active {
    transform: scale(0.99);
  }

  .setting-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }

  .setting-btn.full-width {
    width: 100%;
  }

  .setting-btn.left-aligned {
    justify-content: flex-start;
    text-align: left;
  }

  .setting-btn.danger {
    background: rgba(239, 68, 68, 0.15);
    border-color: rgba(239, 68, 68, 0.35);
    color: #ef4444;
  }

  .setting-btn.danger:hover {
    background: rgba(239, 68, 68, 0.25);
    border-color: rgba(239, 68, 68, 0.45);
  }

  .setting-btn.success {
    background: rgba(16, 185, 129, 0.15);
    border-color: rgba(16, 185, 129, 0.35);
    color: #10b981;
  }

  .setting-btn.success:hover {
    background: rgba(16, 185, 129, 0.25);
    border-color: rgba(16, 185, 129, 0.45);
  }
</style>
