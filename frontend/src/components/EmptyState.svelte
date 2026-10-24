<script lang="ts">
  import type { ComponentType } from 'svelte'

  interface Props {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon?: ComponentType<any>
    title: string
    description?: string
    actionLabel?: string
    onAction?: () => void
  }

  let {
    icon: IconComponent,
    title,
    description = '',
    actionLabel = '',
    onAction
  }: Props = $props()
</script>

<div class="empty-state-container">
  {#if IconComponent}
    <div class="empty-state-icon">
      <IconComponent size={32} />
    </div>
  {/if}
  <h3 class="empty-state-title">{title}</h3>
  {#if description}
    <p class="empty-state-description">{description}</p>
  {/if}
  {#if actionLabel && onAction}
    <button type="button" class="empty-state-action" onclick={onAction}>
      {actionLabel}
    </button>
  {/if}
</div>

<style>
  .empty-state-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 24px 16px;
    border-radius: 18px;
    background: linear-gradient(
      145deg,
      rgba(255, 255, 255, 0.04) 0%,
      rgba(255, 255, 255, 0.01) 100%
    );
    border: 1px solid var(--vinta-border);
    margin: 8px 0;
  }

  .empty-state-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 52px;
    height: 52px;
    border-radius: 50%;
    background: var(--vinta-item-bg);
    color: var(--vinta-text-secondary);
    margin-bottom: 12px;
    box-shadow: inset 0 1px 0 0 rgba(255, 255, 255, 0.08);
  }

  .empty-state-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--vinta-text-primary);
    margin: 0 0 6px 0;
  }

  .empty-state-description {
    font-size: 12px;
    color: var(--vinta-text-muted);
    line-height: 1.4;
    margin: 0;
    max-width: 260px;
  }

  .empty-state-action {
    margin-top: 14px;
    height: 32px;
    padding: 0 14px;
    border-radius: 16px;
    background: var(--vinta-primary);
    color: #ffffff;
    font-size: 12px;
    font-weight: 600;
    border: none;
    cursor: pointer;
    transition:
      transform 0.15s ease,
      opacity 0.15s ease;
  }

  .empty-state-action:hover {
    transform: translateY(-1px);
    opacity: 0.92;
  }

  .empty-state-action:active {
    transform: translateY(0);
  }
</style>
