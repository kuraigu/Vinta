<script lang="ts">
  import type { Snippet } from 'svelte'
  import BackButton from './BackButton.svelte'

  interface Props {
    title?: string
    showBackButton?: boolean
    onback?: () => void
    sidebarWidth?: string
    dataVinta?: string
    sidebar?: Snippet
    content?: Snippet
    children?: Snippet
  }

  let {
    title = '',
    showBackButton = false,
    onback,
    sidebarWidth = '160px',
    dataVinta,
    sidebar,
    content,
    children
  }: Props = $props()
</script>

<div data-vinta={dataVinta} class="sidebar-page-layout">
  {#if title}
    <div class="page-header" class:has-back={showBackButton}>
      {#if showBackButton}
        <BackButton {onback} />
      {/if}
      <div class="page-title">{title}</div>
      {#if showBackButton}
        <div class="header-spacer"></div>
      {/if}
    </div>
  {/if}

  <div class="sidebar-body">
    <div class="sidebar-pane" style="width: {sidebarWidth}">
      {@render sidebar?.()}
    </div>
    <div class="content-pane">
      {@render (content || children)?.()}
    </div>
  </div>
</div>

<style>
  .sidebar-page-layout {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 14px 16px 58px 16px;
    box-sizing: border-box;
    color: var(--c-text-primary);
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 12px;
    flex-shrink: 0;
    width: 100%;
    min-height: 28px;
  }

  .page-header.has-back {
    justify-content: space-between;
  }

  .header-spacer {
    width: 28px;
    height: 28px;
    flex-shrink: 0;
  }

  .page-title {
    font-size: 15px;
    font-weight: 700;
    letter-spacing: -0.01em;
    color: var(--c-text-primary);
    text-align: center;
  }

  .sidebar-body {
    flex: 1;
    display: flex;
    gap: 12px;
    min-height: 0;
    width: 100%;
  }

  .sidebar-pane {
    display: flex;
    flex-direction: column;
    gap: 6px;
    border-right: 1px solid var(--c-border);
    padding-right: 10px;
    flex-shrink: 0;
    box-sizing: border-box;
  }

  .content-pane {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    box-sizing: border-box;
  }
</style>
