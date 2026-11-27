<!--
  @component ExampleOverlayPage
  @description Starter overlay view rendered inside OverlayPage above the active workspace.
-->
<script lang="ts">
  import OverlayPage from '../../components/OverlayPage.svelte'
  import ScrollView from '../../components/ScrollView.svelte'
  import BentoCard from '../../components/BentoCard.svelte'

  interface Props {
    onClose: () => void
  }

  let { onClose }: Props = $props()
</script>

<OverlayPage
  title="Example Overlay"
  dataVinta="example-overlay"
  {onClose}
>
  <ScrollView class="example-overlay-scroll">
    <BentoCard variant="bento">
      <h3 class="overlay-card-title">Overlay Content</h3>
      <p class="overlay-card-desc">
        Uses solid --vinta-bg-opaque surfaces inside OverlayPage without stacking backdrop blur.
      </p>
    </BentoCard>
  </ScrollView>
</OverlayPage>

<style>
  .overlay-card-title {
    margin: 0 0 4px 0;
    font-size: 13px;
    font-weight: 700;
    color: var(--vinta-text-primary);
  }

  .overlay-card-desc {
    margin: 0;
    font-size: 11px;
    color: var(--vinta-text-secondary);
  }
</style>
