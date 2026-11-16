<!--
  @component MediaGlanceWidget
  @description Compact Home widget reusing MediaWidget in minimal glance mode.
  Clicking the card opens MediaPage.
-->
<script lang="ts">
  import { getContext } from 'svelte'
  import WidgetFrame from './WidgetFrame.svelte'
  import BentoCard from '../BentoCard.svelte'
  import MediaWidget from '../MediaWidget.svelte'
  import MediaPage, {
    config as mediaConfig
  } from '../../pages/MediaPage.svelte'
  import {
    globalMediaStore,
    type IslandConfig,
    type WidgetSizePreset
  } from '../../store'

  interface Props {
    size?: WidgetSizePreset
    showMenuButton?: boolean
    oncontextmenu?: (e: MouseEvent) => void
  }

  let {
    size = '2x1',
    showMenuButton = false,
    oncontextmenu = undefined
  }: Props = $props()

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let media = $derived($globalMediaStore)
  let hasMedia = $derived(
    Boolean(
      media &&
      media.isAvailable &&
      (media.title ||
        media.artist ||
        media.isPlaying ||
        (media.sessions && media.sessions.length > 0))
    )
  )

  function openMediaPage(e: MouseEvent) {
    e.stopPropagation()
    if (typeof navigate === 'function') {
      navigate(MediaPage, mediaConfig, true)
    }
  }
</script>

{#if hasMedia}
  <WidgetFrame id="media" {size} {showMenuButton} {oncontextmenu}>
    <BentoCard
      variant="bento"
      class="media-glance-card"
      onclick={openMediaPage}
    >
      <MediaWidget expanded={false} glanceMode={true} widgetSize={size} />
    </BentoCard>
  </WidgetFrame>
{/if}

<style src="../styles/MediaGlanceWidget.css"></style>
