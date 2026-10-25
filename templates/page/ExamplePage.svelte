<script module lang="ts">
  import type { IslandConfig } from '../../frontend/src/store'

  export const config: IslandConfig = {
    minimal: { w: 320, h: 42, r: 21 },
    expanded: { w: 440, h: 420, r: 28 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import BackButton from '../../frontend/src/components/BackButton.svelte'
  import { AppControls, type NavFn } from '../../frontend/src/utils/AppControls.svelte'

  interface Props {
    expanded?: boolean
    navigate?: NavFn
  }

  let { expanded = false, navigate }: Props = $props()

  function handleMinimize() {
    if (navigate) {
      AppControls.doMinimize(navigate)
    }
  }
</script>

<div
  data-vinta="example-page"
  class="example-page-container"
  class:is-expanded={expanded}
>
  <header class="page-header">
    <BackButton onback={handleMinimize} title="Back" />
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
    color: var(--vinta-text-primary);
  }

  .page-header {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 32px;
    margin-bottom: 8px;
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
    color: var(--vinta-text-secondary);
  }
</style>
