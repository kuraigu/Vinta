<script module lang="ts">
  import type { IslandConfig } from '../../frontend/src/store'

  /**
   * Defines island dimensions for minimal and expanded views.
   */
  export const config: IslandConfig = {
    minimal: { w: 320, h: 42, r: 21 },
    expanded: { w: 440, h: 420, r: 28 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { ChevronLeft } from 'lucide-svelte'
  import { AppControls, type NavFn } from '../../frontend/src/utils/AppControls.svelte'

  /**
   * Component properties for ExamplePage.
   */
  interface Props {
    expanded?: boolean
    navigate?: NavFn
  }

  let { expanded = false, navigate }: Props = $props()

  function handleMinimize(e: MouseEvent) {
    if (navigate) {
      AppControls.doMinimize(navigate, e)
    }
  }
</script>

<div
  data-vinta="example-page"
  class="example-page-container"
  class:is-expanded={expanded}
>
  <header class="page-header">
    <button
      type="button"
      class="back-btn"
      onclick={handleMinimize}
      aria-label="Back"
    >
      <ChevronLeft size={16} />
    </button>
    <h1 class="page-title">Starter Page</h1>
  </header>

  <main class="page-content">
    <p>Page content here. Bottom clearance leaves 56px for peripheral dock.</p>
  </main>
</div>

<style>
  .example-page-container {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 12px 14px 56px 14px;
    box-sizing: border-box;
    color: var(--vinta-text-primary, #ffffff);
  }

  .page-header {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 32px;
    margin-bottom: 8px;
  }

  .back-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: var(--vinta-radius-button, 8px);
    background: var(--vinta-item-bg, rgba(255, 255, 255, 0.06));
    border: 1px solid var(--vinta-border, rgba(255, 255, 255, 0.1));
    color: var(--vinta-text-secondary, rgba(255, 255, 255, 0.7));
    cursor: pointer;
  }

  .page-title {
    margin: 0;
    font-size: 13px;
    font-weight: 700;
  }

  .page-content {
    flex: 1;
    overflow-y: auto;
    font-size: 12px;
    color: var(--vinta-text-secondary, rgba(255, 255, 255, 0.7));
  }
</style>
