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
    pageVisibilityStore,
    type IslandConfig,
    type WidgetSizePreset
  } from '../../store'

  interface Props {
    size?: WidgetSizePreset
    active?: boolean
    showMenuButton?: boolean
    oncontextmenu?: (e: MouseEvent) => void
  }

  let {
    size = '2x1',
    active = true,
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

  let isMediaEnabled = $derived($pageVisibilityStore.media !== false)

  function openMediaPage(e: MouseEvent) {
    e.stopPropagation()
    if (!isMediaEnabled) return
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
      onclick={isMediaEnabled ? openMediaPage : undefined}
    >
      <MediaWidget
        expanded={false}
        glanceMode={true}
        {active}
        widgetSize={size}
      />
    </BentoCard>
  </WidgetFrame>
{/if}

<style src="../styles/MediaGlanceWidget.css"></style>
