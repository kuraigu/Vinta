<script module lang="ts">
  export interface InspectorHoverPayload {
    tag: string
    classes: string[]
    dataVinta: string | null
    selector: string
    width: number
    height: number
  }
</script>

<script lang="ts">
  import { onMount } from 'svelte'
  import { Events } from '@wailsio/runtime'
  import { debugClassInspector } from './debugConfig'

  let hoveredRect = $state<{
    left: number
    top: number
    width: number
    height: number
  } | null>(null)
  let hoveredInfo = $state<InspectorHoverPayload | null>(null)
  let islandRect = $state<DOMRect | null>(null)

  /**
   * Filters out internal Svelte generated hash classes to present clean semantic class names.
   */
  function extractCleanClasses(el: HTMLElement): string[] {
    if (!el.classList) return []
    return Array.from(el.classList).filter(
      (c) => !/^s-[A-Za-z0-9_-]{8,}$/.test(c) && !/^svelte-[a-z0-9]+$/i.test(c)
    )
  }

  /**
   * Finds the nearest ancestor or self carrying a data-vinta selector attribute.
   */
  function findNearestDataVinta(el: HTMLElement): string | null {
    let curr: HTMLElement | null = el
    while (curr && curr !== document.body) {
      const attr = curr.getAttribute('data-vinta')
      if (attr) return attr
      curr = curr.parentElement
    }
    return null
  }

  function handlePointerMove(e: PointerEvent): void {
    if (!$debugClassInspector) return

    const target = e.target as HTMLElement | null
    if (
      !target ||
      target === document.body ||
      target === document.documentElement ||
      target.closest('.class-inspector-layer') ||
      target.closest('.island-custom-tooltip')
    ) {
      hoveredRect = null
      hoveredInfo = null
      return
    }

    const rect = target.getBoundingClientRect()
    if (rect.width === 0 && rect.height === 0) {
      hoveredRect = null
      hoveredInfo = null
      return
    }

    const mainIslandEl = document.querySelector(
      '[data-vinta="main-island"]'
    ) as HTMLElement | null
    islandRect = mainIslandEl ? mainIslandEl.getBoundingClientRect() : null

    const tag = target.tagName.toLowerCase()
    const classes = extractCleanClasses(target)
    const dataVinta = findNearestDataVinta(target)
    const classPart = classes.length > 0 ? `.${classes.join('.')}` : ''
    const vintaPrefix = dataVinta ? `[data-vinta="${dataVinta}"] ` : ''
    const selector = `${vintaPrefix}${tag}${classPart}`

    hoveredRect = {
      left: Math.round(rect.left),
      top: Math.round(rect.top),
      width: Math.round(rect.width),
      height: Math.round(rect.height)
    }

    const payload: InspectorHoverPayload = {
      tag,
      classes,
      dataVinta,
      selector,
      width: Math.round(rect.width),
      height: Math.round(rect.height)
    }
    hoveredInfo = payload

    try {
      Events.Emit('devtest-inspector-hover', payload)
    } catch {
      // Ignore when running outside Wails runtime
    }
  }

  function handlePointerLeave(): void {
    hoveredRect = null
    hoveredInfo = null
  }

  function syncInspectorState(): void {
    if (!$debugClassInspector) {
      hoveredRect = null
      hoveredInfo = null
    }
  }
  $effect(syncInspectorState)

  onMount(() => {
    window.addEventListener('pointermove', handlePointerMove, { passive: true })
    document.documentElement.addEventListener(
      'pointerleave',
      handlePointerLeave
    )
    return () => {
      window.removeEventListener('pointermove', handlePointerMove)
      document.documentElement.removeEventListener(
        'pointerleave',
        handlePointerLeave
      )
    }
  })

  function computePillPosition(): { left: number; top: number } {
    if (!hoveredRect) return { left: 0, top: 0 }
    if (islandRect) {
      return {
        left: Math.round(islandRect.left + islandRect.width / 2),
        top: Math.max(4, Math.round(islandRect.bottom - 26))
      }
    }
    return {
      left: Math.round(hoveredRect.left + hoveredRect.width / 2),
      top: Math.max(4, hoveredRect.top - 24)
    }
  }
  let pillPos = $derived.by(computePillPosition)
</script>

{#if $debugClassInspector && hoveredRect && hoveredInfo}
  <div class="class-inspector-layer" aria-hidden="true">
    <div
      class="inspector-highlight-box"
      style="
        left: {hoveredRect.left}px;
        top: {hoveredRect.top}px;
        width: {hoveredRect.width}px;
        height: {hoveredRect.height}px;
      "
    ></div>

    <div
      class="inspector-hud-pill"
      style="
        left: {pillPos.left}px;
        top: {pillPos.top}px;
      "
    >
      {#if hoveredInfo.dataVinta}
        <span class="hud-vinta">[{hoveredInfo.dataVinta}]</span>
      {/if}
      <span class="hud-tag">{hoveredInfo.tag}</span>
      {#if hoveredInfo.classes.length > 0}
        <span class="hud-classes">.{hoveredInfo.classes.slice(0, 3).join('.')}</span>
      {/if}
      <span class="hud-dims">{hoveredInfo.width}×{hoveredInfo.height}</span>
    </div>
  </div>
{/if}

<style>
  .class-inspector-layer {
    position: fixed;
    inset: 0;
    pointer-events: none;
    z-index: 9998;
    overflow: hidden;
  }

  .inspector-highlight-box {
    position: fixed;
    box-sizing: border-box;
    border: 1.5px dashed #38bdf8;
    background: rgba(56, 189, 248, 0.1);
    border-radius: 4px;
    pointer-events: none;
    transition:
      left 0.05s ease-out,
      top 0.05s ease-out,
      width 0.05s ease-out,
      height 0.05s ease-out;
  }

  .inspector-hud-pill {
    position: fixed;
    transform: translateX(-50%);
    display: inline-flex;
    align-items: center;
    gap: 5px;
    max-width: 340px;
    padding: 3px 9px;
    border-radius: 999px;
    background: rgba(15, 23, 42, 0.94);
    border: 1px solid rgba(56, 189, 248, 0.45);
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.55);
    font-family: 'JetBrains Mono', 'Cascadia Code', Consolas, monospace;
    font-size: 10px;
    line-height: 1.3;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    pointer-events: none;
  }

  .hud-vinta {
    color: #c084fc;
    font-weight: 700;
  }

  .hud-tag {
    color: #38bdf8;
    font-weight: 700;
  }

  .hud-classes {
    color: #fbbf24;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .hud-dims {
    color: #94a3b8;
    font-size: 9.5px;
    margin-left: 2px;
  }
</style>
