<!--
  @component IslandPage
  @description Shared page host that separates minimal pill and expanded card snippets,
  pre-warming the expanded view during idle time and toggling visibility via CSS display
  so pages never tear down or remount their DOM on expand/collapse.
-->
<script lang="ts">
  import { onMount, type Snippet } from 'svelte'

  interface Props {
    expanded: boolean
    minimal: Snippet
    expandedContent: Snippet
    prewarmMs?: number
  }

  let {
    expanded = false,
    minimal,
    expandedContent,
    prewarmMs = 1200
  }: Props = $props()

  let expandedMounted = $state(expanded)

  $effect(function ensureExpandedMountedEffect() {
    if (expanded && !expandedMounted) {
      expandedMounted = true
    }
  })

  onMount(() => {
    if (expandedMounted) return
    let idleHandle: number | undefined
    let fallbackTimer: ReturnType<typeof setTimeout> | undefined
    const prewarm = () => {
      if (!expandedMounted) expandedMounted = true
    }
    if (typeof requestIdleCallback === 'function') {
      idleHandle = requestIdleCallback(prewarm, { timeout: prewarmMs })
    } else {
      fallbackTimer = setTimeout(prewarm, Math.min(prewarmMs, 900))
    }
    return () => {
      if (
        idleHandle !== undefined &&
        typeof cancelIdleCallback === 'function'
      ) {
        cancelIdleCallback(idleHandle)
      }
      if (fallbackTimer !== undefined) {
        clearTimeout(fallbackTimer)
      }
    }
  })
</script>

<div style:display={expanded ? 'none' : 'contents'}>
  {@render minimal()}
</div>

{#if expandedMounted}
  <div style:display={expanded ? 'contents' : 'none'}>
    {@render expandedContent()}
  </div>
{/if}
