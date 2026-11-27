<!--
  @component ExamplePage
  @description Starter workspace page using IslandPage (pre-warmed minimal/expanded slots),
  PageHeader, and ScrollView with 56px bottom dock clearance.
-->
<script module lang="ts">
  import { ISLAND_WIDTH_TIERS, type IslandConfig } from '../../frontend/src/store'

  export const config: IslandConfig = {
    minimal: { w: 200, h: 36, r: 18 },
    expanded: { w: ISLAND_WIDTH_TIERS.standard, h: 420, r: 28 },
    borderColor: 'var(--vinta-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import IslandPage from '../../frontend/src/components/IslandPage.svelte'
  import PageHeader from '../../frontend/src/components/PageHeader.svelte'
  import ScrollView from '../../frontend/src/components/ScrollView.svelte'
  import BentoCard from '../../frontend/src/components/BentoCard.svelte'

  interface Props {
    expanded?: boolean
  }

  let { expanded = false }: Props = $props()
</script>

<IslandPage {expanded}>
  {#snippet minimal()}
    <div class="example-minimal">
      <span>Example</span>
    </div>
  {/snippet}

  {#snippet expandedContent()}
    <div class="example-page" data-vinta="example-page">
      <PageHeader title="Example Page" />

      <ScrollView class="example-scroll-area">
        <BentoCard variant="bento">
          <p class="example-body">
            Page content here. ScrollView leaves 56px bottom clearance for the dock.
          </p>
        </BentoCard>
      </ScrollView>
    </div>
  {/snippet}
</IslandPage>

<style>
  .example-minimal {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    font-weight: 600;
    color: var(--vinta-text-primary);
  }

  .example-page {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 12px 14px 0 14px;
    box-sizing: border-box;
    color: var(--vinta-text-primary);
  }

  :global(.example-scroll-area) {
    flex: 1;
    padding-bottom: 56px;
  }

  .example-body {
    margin: 0;
    font-size: 12px;
    color: var(--vinta-text-secondary);
  }
</style>
