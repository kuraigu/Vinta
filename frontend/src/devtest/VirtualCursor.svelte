<!--
  @component VirtualCursor
  @description Simulated desktop cursor for automated showcases and recordings.
  Renders an on-screen SVG pointer with smooth cubic Bezier easing, visual click ripple rings,
  tactile button clicking sequences, and Promise-chained event synchronization.
-->
<script lang="ts">
  import { onMount } from 'svelte'
  import { fade, scale } from 'svelte/transition'
  import { Events } from '@wailsio/runtime'
  import {
    Search,
    Minus,
    Square,
    X,
    Plus,
    ChevronDown,
    ArrowLeft,
    ArrowRight,
    ArrowUp,
    RotateCw,
    FileText,
    Image,
    Folder,
    FolderDown,
    Download,
    Scissors,
    Copy,
    Trash2,
    LayoutGrid,
    List
  } from 'lucide-svelte'

  interface CursorPayload {
    action:
      | 'show'
      | 'hide'
      | 'move'
      | 'click'
      | 'keypad-sequence'
      | 'scroll'
      | 'type'
      | 'drag-file'
      | 'reset'
    selector?: string
    containsText?: string
    keys?: string[]
    x?: number
    y?: number
    xRatio?: number
    yRatio?: number
    fromXRatio?: number
    fromYRatio?: number
    toXRatio?: number
    toYRatio?: number
    fromSelector?: string
    toSelector?: string
    holdProgress?: boolean
    fileBadge?: {
      name: string
      source: string
      icon?: string
      iconType?: 'pdf' | 'img' | 'folder' | 'doc'
    }
    duration?: number
    deltaY?: number
    text?: string
    delay?: number
    keyDelay?: number
    holdDurationMs?: number
    pickRandom?: boolean
  }

  interface Ripple {
    id: number
    x: number
    y: number
  }

  interface Props {
    primaryCenter?: number
    clampedOffsetX?: number
    monitorTop?: number
    monitorWidth?: number
  }

  let {
    primaryCenter = typeof window !== 'undefined' ? window.innerWidth / 2 : 960,
    clampedOffsetX = 0,
    monitorTop = 0,
    monitorWidth: _monitorWidth = typeof window !== 'undefined'
      ? window.innerWidth
      : 1920
  }: Props = $props()

  const CURSOR_VISIBLE_STORAGE_KEY = 'vinta_showcase_cursor_visible'
  const CURSOR_TRAIL_STORAGE_KEY = 'vinta_showcase_cursor_trail'

  function loadInitialCursorVisible(): boolean {
    try {
      const saved = localStorage.getItem(CURSOR_VISIBLE_STORAGE_KEY)
      if (saved !== null) return JSON.parse(saved)
    } catch {}
    return true
  }

  function loadInitialCursorTrail(): boolean {
    try {
      const saved = localStorage.getItem(CURSOR_TRAIL_STORAGE_KEY)
      if (saved !== null) return JSON.parse(saved)
    } catch {}
    return true
  }

  let visible = $state(loadInitialCursorVisible())
  let motionTrailEnabled = $state(loadInitialCursorTrail())
  let carriedBadge = $state<{
    name: string
    source: string
    icon?: string
    iconType?: 'pdf' | 'img' | 'folder' | 'doc'
  } | null>(null)
  let showcaseFoldersVisible = $state(false)
  let destinationFileCount = $state(0)
  let posX = $state(typeof window !== 'undefined' ? window.innerWidth / 2 : 500)
  let posY = $state(
    typeof window !== 'undefined' ? window.innerHeight * 0.35 : 200
  )
  let cursorTilt = $state(0)
  let isClicking = $state(false)
  let animDuration = $state(0.7)
  let ripples = $state<Ripple[]>([])
  let nextRippleId = 0
  let currentHoveredEl: HTMLElement | null = null

  let cursorContainer = $state<HTMLDivElement | null>(null)
  let cursorBody = $state<HTMLDivElement | null>(null)

  /**
   * Human Kinematic Flight Engine.
   * Interpolates the cursor along an organic Quadratic Bezier curve with natural arc curvature,
   * Fitts's Law ease-out deceleration, and real-time directional wrist tilt.
   */
  function glideCursor(
    targetX: number,
    targetY: number,
    durationMs: number
  ): Promise<void> {
    return new Promise((resolve) => {
      const startX = posX
      const startY = posY
      const dx = targetX - startX
      const dy = targetY - startY
      const dist = Math.hypot(dx, dy)

      if (dist < 4 || durationMs <= 20) {
        posX = targetX
        posY = targetY
        cursorTilt = 0
        if (cursorContainer) {
          cursorContainer.style.transform = `translate3d(${targetX}px, ${targetY}px, 0)`
        }
        if (cursorBody) {
          cursorBody.style.transform = `rotate(0deg) scale(${isClicking ? 0.88 : 1})`
        }
        resolve()
        return
      }

      // Calculate an organic human arc control point perpendicular to flight vector
      const perpX = -dy / dist
      const perpY = dx / dist
      // Subtle natural arc height (4% to 8% of distance, capped at 28px)
      const arcMagnitude =
        Math.min(28, Math.max(5, dist * 0.07)) * (dx >= 0 ? 1 : -0.75)
      const ctrlX = (startX + targetX) / 2 + perpX * arcMagnitude
      const ctrlY = (startY + targetY) / 2 + perpY * arcMagnitude

      const startTime = performance.now()

      function step(now: number) {
        const elapsed = now - startTime
        const rawProgress = Math.min(1, elapsed / durationMs)

        // Human kinematic ease: soft initial acceleration, swift glide, organic Fitts deceleration
        const p =
          rawProgress < 0.2
            ? (rawProgress / 0.2) * (rawProgress / 0.2) * 0.08
            : 0.08 + 0.92 * (1 - Math.pow(1 - (rawProgress - 0.2) / 0.8, 3.2))

        const progress = Math.min(1, Math.max(0, p))
        const inv = 1 - progress

        // Quadratic Bezier curve: B(p) = (1-p)^2 P0 + 2(1-p)p Pctrl + p^2 P1
        const curX =
          inv * inv * startX +
          2 * inv * progress * ctrlX +
          progress * progress * targetX
        const curY =
          inv * inv * startY +
          2 * inv * progress * ctrlY +
          progress * progress * targetY

        // Tangent derivative for natural wrist tilt
        const derivX =
          2 * inv * (ctrlX - startX) + 2 * progress * (targetX - ctrlX)
        const dampener = 1 - Math.pow(progress, 4)
        const tilt = Math.max(-6, Math.min(6, (derivX / 90) * 3.2 * dampener))

        posX = curX
        posY = curY
        cursorTilt = tilt

        if (cursorContainer) {
          cursorContainer.style.transform = `translate3d(${curX}px, ${curY}px, 0)`
        }
        if (cursorBody) {
          cursorBody.style.transform = `rotate(${tilt}deg) scale(${isClicking ? 0.88 : 1})`
        }

        if (rawProgress < 1) {
          requestAnimationFrame(step)
        } else {
          posX = targetX
          posY = targetY
          cursorTilt = 0
          if (cursorContainer) {
            cursorContainer.style.transform = `translate3d(${targetX}px, ${targetY}px, 0)`
          }
          if (cursorBody) {
            cursorBody.style.transform = `rotate(0deg) scale(${isClicking ? 0.88 : 1})`
          }
          resolve()
        }
      }

      requestAnimationFrame(step)
    })
  }

  function updateHover(el: HTMLElement | null) {
    if (currentHoveredEl === el) return
    if (currentHoveredEl) {
      currentHoveredEl.classList.remove('vinta-virtual-hover')
      currentHoveredEl.dispatchEvent(
        new PointerEvent('pointerleave', {
          bubbles: true,
          clientX: posX,
          clientY: posY
        })
      )
      currentHoveredEl.dispatchEvent(
        new MouseEvent('mouseleave', {
          bubbles: true,
          clientX: posX,
          clientY: posY
        })
      )
    }
    currentHoveredEl = el
    if (currentHoveredEl) {
      currentHoveredEl.classList.add('vinta-virtual-hover')
      currentHoveredEl.dispatchEvent(
        new PointerEvent('pointerenter', {
          bubbles: true,
          clientX: posX,
          clientY: posY
        })
      )
      currentHoveredEl.dispatchEvent(
        new MouseEvent('mouseenter', {
          bubbles: true,
          clientX: posX,
          clientY: posY
        })
      )
      currentHoveredEl.dispatchEvent(
        new MouseEvent('mouseover', {
          bubbles: true,
          clientX: posX,
          clientY: posY
        })
      )
    }
  }

  function sleep(ms: number): Promise<void> {
    return new Promise((r) => setTimeout(r, ms))
  }

  /**
   * Triggers a visual expanding ripple ring at cursor position.
   */
  function spawnRipple(x: number, y: number) {
    const id = nextRippleId++
    ripples = [...ripples, { id, x, y }]
    setTimeout(() => {
      ripples = ripples.filter((r) => r.id !== id)
    }, 600)
  }

  /**
   * Resolves target coordinates based on a DOM selector or relative ratio.
   */
  function resolveCoordinates(payload: CursorPayload): {
    x: number
    y: number
    el: HTMLElement | null
  } {
    let targetX = posX
    let targetY = posY
    let targetEl: HTMLElement | null = null

    if (payload.selector) {
      let candidates: HTMLElement[] = []
      try {
        candidates = Array.from(
          document.querySelectorAll<HTMLElement>(payload.selector)
        )
      } catch {
        candidates = []
      }

      // Filter for elements that are actually laid out with nonzero dimensions
      const visibleCandidates = candidates.filter((c) => {
        const r = c.getBoundingClientRect()
        return r.width > 0 && r.height > 0
      })
      const list = visibleCandidates.length > 0 ? visibleCandidates : candidates

      if (payload.containsText) {
        const query = payload.containsText.toLowerCase().trim()
        targetEl =
          list.find((c) => c.textContent?.toLowerCase().includes(query)) || null
      }
      if (!targetEl && list.length > 0) {
        if (payload.pickRandom) {
          const nonToday = list.filter(
            (c) =>
              !c.classList.contains('today') && !c.classList.contains('empty')
          )
          const pool = nonToday.length > 0 ? nonToday : list
          const randIdx = Math.floor(Math.random() * pool.length)
          targetEl = pool[randIdx]
        } else {
          targetEl = list[0]
        }
      }
      if (
        !targetEl &&
        (payload.selector.includes('floating-controls') ||
          payload.selector.includes('controls'))
      ) {
        targetEl = document.querySelector<HTMLElement>(
          '[data-satellite-host="controls"] button, .satellite-host button, .floating-controls button'
        )
      }
      if (targetEl) {
        const rect = targetEl.getBoundingClientRect()
        // Organic human landing offset: real humans don't strike dead geometric centers
        const hash =
          Math.sin(rect.left * 12.9898 + rect.top * 78.233) * 43758.5453
        const varianceX =
          (hash - Math.floor(hash) - 0.5) * Math.min(5, rect.width * 0.22)
        const hashY =
          Math.cos(rect.left * 39.346 + rect.top * 11.135) * 23421.631
        const varianceY =
          (hashY - Math.floor(hashY) - 0.5) * Math.min(4, rect.height * 0.2)
        targetX = rect.left + rect.width / 2 + varianceX
        targetY = rect.top + rect.height / 2 + varianceY
      } else if (
        payload.selector.includes('floating-controls') ||
        payload.selector.includes('minimize-2')
      ) {
        const island = document.querySelector<HTMLElement>(
          '.main-island, .island-shell, .island-wrapper'
        )
        if (island) {
          const rect = island.getBoundingClientRect()
          targetX = rect.right + 22
          targetY = rect.top + 24
        }
      }
    } else if (
      typeof payload.xRatio === 'number' &&
      typeof payload.yRatio === 'number'
    ) {
      targetX = window.innerWidth * payload.xRatio
      targetY = window.innerHeight * payload.yRatio
    } else if (typeof payload.x === 'number' && typeof payload.y === 'number') {
      targetX = payload.x
      targetY = payload.y
    }

    return { x: targetX, y: targetY, el: targetEl }
  }

  /**
   * Finds a calculator button matching a character token.
   */
  function findCalcButton(token: string): HTMLElement | null {
    const buttons = Array.from(
      document.querySelectorAll<HTMLElement>('.calc-btn')
    )
    for (const b of buttons) {
      const text = b.textContent?.trim()
      if (text === token) return b
      if (token === '*' && text === '×') return b
      if (token === '/' && text === '÷') return b
      if (token === '-' && text === '−') return b
      if (token === '+' && text === '+') return b
      if (token === '=' && (text === '=' || b.classList.contains('equals-btn')))
        return b
    }
    return null
  }

  onMount(() => {
    const unsubTrail = Events.On('devtest-cursor-trail', (ev: any) => {
      const payload = ev?.data?.[0] ?? ev?.data ?? ev
      if (payload?.enabled !== undefined) {
        motionTrailEnabled = payload.enabled
      } else {
        motionTrailEnabled = !motionTrailEnabled
      }
      try {
        localStorage.setItem(
          CURSOR_TRAIL_STORAGE_KEY,
          JSON.stringify(motionTrailEnabled)
        )
      } catch {}
    })

    const unsubAction = Events.On('devtest-cursor-action', async (ev: any) => {
      const payload: CursorPayload = ev?.data?.[0] ?? ev?.data ?? ev
      if (!payload) return

      if (payload.action === 'show') {
        visible = true
        try {
          localStorage.setItem(CURSOR_VISIBLE_STORAGE_KEY, JSON.stringify(true))
        } catch {}
        Events.Emit('devtest-cursor-done', { action: 'show' })
        return
      }

      if (payload.action === 'hide') {
        updateHover(null)
        visible = false
        try {
          localStorage.setItem(
            CURSOR_VISIBLE_STORAGE_KEY,
            JSON.stringify(false)
          )
        } catch {}
        Events.Emit('devtest-cursor-done', { action: 'hide' })
        return
      }

      if (payload.action === 'reset') {
        updateHover(null)
        visible = true
        try {
          localStorage.setItem(CURSOR_VISIBLE_STORAGE_KEY, JSON.stringify(true))
        } catch {}
        posX = window.innerWidth / 2
        posY = window.innerHeight * 0.35
        Events.Emit('devtest-cursor-done', { action: 'reset' })
        return
      }

      if (payload.action === 'scroll') {
        const container = payload.selector
          ? document.querySelector<HTMLElement>(payload.selector)
          : document.querySelector<HTMLElement>('.scroll-view')
        if (container) {
          container.scrollBy({
            top: payload.deltaY || 200,
            behavior: 'smooth'
          })
        }
        await sleep(700)
        Events.Emit('devtest-cursor-done', { action: 'scroll' })
        return
      }

      if (payload.action === 'drag-file' && payload.fileBadge) {
        updateHover(null)
        visible = true

        let startX = window.innerWidth * (payload.fromXRatio ?? 0.15)
        let startY = window.innerHeight * (payload.fromYRatio ?? 0.5)
        if (payload.fromSelector) {
          const startEl = document.querySelector<HTMLElement>(
            payload.fromSelector
          )
          if (startEl) {
            const rect = startEl.getBoundingClientRect()
            startX = rect.left + rect.width / 2
            startY = rect.top + rect.height / 2
          }
        }

        let endX = window.innerWidth * (payload.toXRatio ?? 0.5)
        let endY = window.innerHeight * (payload.toYRatio ?? 0.04)
        if (payload.toSelector) {
          const endEl = document.querySelector<HTMLElement>(payload.toSelector)
          if (endEl) {
            const rect = endEl.getBoundingClientRect()
            endX = rect.left + rect.width / 2
            endY = rect.top + rect.height / 2
          }
        }

        // 1. Move cursor smoothly along natural arc to start item
        await glideCursor(startX, startY, 350)

        // 2. Press mouse down to grasp file badge
        isClicking = true
        spawnRipple(posX, posY)
        carriedBadge = payload.fileBadge
        await sleep(100)

        // 3. Drag file smoothly across to destination with natural carrying curve
        const dragDurationMs =
          typeof payload.duration === 'number'
            ? Math.round(payload.duration * 1000)
            : 750
        await glideCursor(endX, endY, dragDurationMs)

        // 4. If holdProgress requested (e.g. dropping onto Island), run authentic 950ms hold sequence
        if (payload.holdProgress) {
          const holdTotalMs = payload.holdDurationMs || 950
          const holdSteps = 24
          const stepDuration = Math.max(16, Math.round(holdTotalMs / holdSteps))
          for (let i = 1; i <= holdSteps; i++) {
            const p = Math.min(100, Math.round((i / holdSteps) * 100))
            Events.Emit('devtest-file-drop-hold', { active: true, progress: p })
            await sleep(stepDuration)
          }
          await sleep(80)
          Events.Emit('devtest-file-drop-hold', { active: false, progress: 0 })
        }

        // 5. Release mouse at destination
        spawnRipple(posX, posY)
        isClicking = false
        carriedBadge = null
        await sleep(60)
        Events.Emit('devtest-cursor-done', { action: 'drag-file' })
        return
      }

      if (payload.action === 'keypad-sequence' && payload.keys) {
        visible = true
        const keyDelay = payload.keyDelay || 420
        const glideTime = Math.round(keyDelay * 0.65)
        const pressTime = Math.round(keyDelay * 0.22)
        const pauseTime = Math.round(keyDelay * 0.13)

        for (const key of payload.keys) {
          const btn = findCalcButton(key)
          if (btn) {
            const rect = btn.getBoundingClientRect()
            const targetX =
              rect.left + rect.width / 2 + (Math.random() - 0.5) * 4
            const targetY =
              rect.top + rect.height / 2 + (Math.random() - 0.5) * 3

            // Fluid organic human typing glide
            await glideCursor(targetX, targetY, glideTime)
            updateHover(btn)

            isClicking = true
            spawnRipple(posX, posY)
            btn.dispatchEvent(
              new MouseEvent('mousedown', {
                bubbles: true,
                clientX: posX,
                clientY: posY
              })
            )
            btn.click()
            btn.dispatchEvent(
              new MouseEvent('mouseup', {
                bubbles: true,
                clientX: posX,
                clientY: posY
              })
            )
            await sleep(pressTime)
            isClicking = false
            await sleep(pauseTime)
          }
        }

        updateHover(null)
        Events.Emit('devtest-cursor-done', { action: 'keypad-sequence' })
        return
      }

      // Standard move or click
      visible = true
      const durationMs =
        typeof payload.duration === 'number'
          ? Math.round(payload.duration * 1000)
          : 450

      const { x, y, el } = resolveCoordinates(payload)
      await glideCursor(x, y, durationMs)
      updateHover(el)

      if (payload.action === 'click') {
        let targetEl = el
        if (payload.selector) {
          if (payload.pickRandom && targetEl) {
            const freshRect = targetEl.getBoundingClientRect()
            const freshX = freshRect.left + freshRect.width / 2
            const freshY = freshRect.top + freshRect.height / 2
            if (Math.hypot(freshX - posX, freshY - posY) > 4) {
              await glideCursor(freshX, freshY, 80)
            }
          } else {
            const fresh = resolveCoordinates(payload)
            targetEl = fresh.el
            if (Math.hypot(fresh.x - posX, fresh.y - posY) > 4) {
              await glideCursor(fresh.x, fresh.y, 80)
            }
          }
        }

        updateHover(targetEl)
        // Subtle human pause before press
        await sleep(30)
        isClicking = true
        spawnRipple(posX, posY)

        if (targetEl) {
          const evtInit = {
            bubbles: true,
            cancelable: true,
            view: window,
            clientX: posX,
            clientY: posY
          }
          targetEl.dispatchEvent(new PointerEvent('pointerdown', evtInit))
          targetEl.dispatchEvent(new MouseEvent('mousedown', evtInit))
          targetEl.dispatchEvent(new PointerEvent('pointerup', evtInit))
          targetEl.dispatchEvent(new MouseEvent('mouseup', evtInit))
          const btn = targetEl.closest('button')
          if (btn && !btn.closest('.select-trigger, .option-item')) {
            btn.click()
          } else {
            targetEl.dispatchEvent(new MouseEvent('click', evtInit))
          }
        }

        // Real human click contact dwell
        await sleep(65)
        isClicking = false
        await sleep(35)
      }

      if (payload.action === 'type' && payload.text) {
        if (el) {
          el.focus()
          if (
            el instanceof HTMLInputElement ||
            el instanceof HTMLTextAreaElement
          ) {
            el.value = ''
            el.dispatchEvent(new Event('input', { bubbles: true }))
          } else if ((el as HTMLElement).isContentEditable) {
            ;(el as HTMLElement).textContent = ''
            el.dispatchEvent(new Event('input', { bubbles: true }))
          }
        }
        const text = payload.text
        const delay = payload.delay || 70

        for (let i = 0; i < text.length; i++) {
          const char = text[i]
          if (el && (el as HTMLElement).isContentEditable) {
            document.execCommand('insertText', false, char)
            el.dispatchEvent(new Event('input', { bubbles: true }))
          } else if (
            el &&
            (el instanceof HTMLInputElement ||
              el instanceof HTMLTextAreaElement)
          ) {
            el.value += char
            el.dispatchEvent(new Event('input', { bubbles: true }))
          }
          window.dispatchEvent(
            new KeyboardEvent('keydown', { key: char, bubbles: true })
          )
          window.dispatchEvent(
            new KeyboardEvent('keypress', { key: char, bubbles: true })
          )
          window.dispatchEvent(
            new KeyboardEvent('keyup', { key: char, bubbles: true })
          )
          await sleep(delay)
        }
      }

      Events.Emit('devtest-cursor-done', { action: payload.action })
    })

    const unsubFolders = Events.On('devtest-showcase-folders', (ev: any) => {
      const payload = ev?.data?.[0] ?? ev?.data ?? ev
      if (typeof payload?.visible === 'boolean') {
        showcaseFoldersVisible = payload.visible
      }
      if (typeof payload?.destinationCount === 'number') {
        destinationFileCount = payload.destinationCount
      }
      if (payload?.visible === false) {
        destinationFileCount = 0
      }
      // Resync hit bounds so folder windows are immediately clickable by real mouse
      setTimeout(() => {
        if (typeof window !== 'undefined') {
          window.dispatchEvent(new CustomEvent('vinta-resync-bounds'))
        }
      }, 30)
    })

    return () => {
      unsubTrail()
      unsubAction()
      unsubFolders()
    }
  })
</script>

{#if visible}
  {#if motionTrailEnabled}
    <!-- Motion blur ghost trailing cursor 2 -->
    <div
      class="virtual-cursor-container ghost ghost-2"
      style="
        transform: translate3d({posX}px, {posY}px, 0);
        transition: transform {animDuration +
        0.12}s cubic-bezier(0.25, 0.8, 0.45, 1);
      "
    >
      <div class="virtual-cursor-body">
        <svg
          width="26"
          height="26"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          class="cursor-arrow-svg"
        >
          <path
            d="M5.5 3.21V20.8c0 .45.54.67.85.35l4.86-4.86a.5.5 0 0 1 .35-.15h6.87c.45 0 .67-.54.35-.85L6.35 2.86a.5.5 0 0 0-.85.35Z"
            fill="#e2a03f"
            stroke="#ffffff"
            stroke-width="1.6"
            stroke-linejoin="round"
          />
        </svg>
      </div>
    </div>

    <!-- Motion blur ghost trailing cursor 1 -->
    <div
      class="virtual-cursor-container ghost ghost-1"
      style="
        transform: translate3d({posX}px, {posY}px, 0);
        transition: transform {animDuration +
        0.05}s cubic-bezier(0.2, 0.9, 0.4, 1);
      "
    >
      <div class="virtual-cursor-body">
        <svg
          width="26"
          height="26"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          class="cursor-arrow-svg"
        >
          <path
            d="M5.5 3.21V20.8c0 .45.54.67.85.35l4.86-4.86a.5.5 0 0 1 .35-.15h6.87c.45 0 .67-.54.35-.85L6.35 2.86a.5.5 0 0 0-.85.35Z"
            fill="#1a1a24"
            stroke="#ffffff"
            stroke-width="1.6"
            stroke-linejoin="round"
          />
        </svg>
      </div>
    </div>
  {/if}

  <div
    bind:this={cursorContainer}
    class="virtual-cursor-container"
    style="transform: translate3d({posX}px, {posY}px, 0);"
  >
    <div
      bind:this={cursorBody}
      class="virtual-cursor-body"
      class:is-clicking={isClicking}
      style="transform: rotate({cursorTilt}deg) scale({isClicking ? 0.88 : 1});"
    >
      <svg
        width="26"
        height="26"
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        class="cursor-arrow-svg"
      >
        <path
          d="M5.5 3.21V20.8c0 .45.54.67.85.35l4.86-4.86a.5.5 0 0 1 .35-.15h6.87c.45 0 .67-.54.35-.85L6.35 2.86a.5.5 0 0 0-.85.35Z"
          fill="#111116"
          stroke="#ffffff"
          stroke-width="1.6"
          stroke-linejoin="round"
        />
      </svg>
      {#if carriedBadge}
        <div class="carried-badge-pill">
          <div class="badge-icon-box">
            {#if carriedBadge.iconType === 'pdf' || carriedBadge.name.endsWith('.pdf')}
              <FileText size={15} color="#ef4444" />
            {:else if carriedBadge.iconType === 'img' || carriedBadge.name.endsWith('.png')}
              <Image size={15} color="#10b981" />
            {:else}
              <Folder size={15} color="#f59e0b" />
            {/if}
          </div>
          <div class="badge-text-col">
            <span class="badge-filename">{carriedBadge.name}</span>
            <span class="badge-source-folder">{carriedBadge.source}</span>
          </div>
        </div>
      {/if}
    </div>
  </div>

  {#each ripples as r (r.id)}
    <div class="click-ripple" style="left: {r.x}px; top: {r.y}px;"></div>
  {/each}
{/if}

{#if showcaseFoldersVisible}
  <div
    class="showcase-folders-layer"
    aria-hidden="true"
    style="left: {primaryCenter + clampedOffsetX}px; top: {monitorTop}px;"
    transition:fade={{ duration: 280 }}
  >
    <!-- Source Folder Window A: Downloads (Top-Left of Island) -->
    <div
      class="desktop-folder-window source-a"
      data-showcase-folder="downloads"
      class:is-active={carriedBadge?.source?.includes('Downloads')}
      transition:scale={{ duration: 300, start: 0.92, opacity: 0 }}
    >
      <!-- Win11 Tab / Title Bar -->
      <div class="folder-tab-bar">
        <div class="folder-tabs-left">
          <div class="win11-tab active">
            <svg
              class="folder-win-icon"
              width="15"
              height="15"
              viewBox="0 0 24 24"
              fill="none"
            >
              <path
                d="M2.5 7C2.5 5.61929 3.61929 4.5 5 4.5H9.37868C10.0417 4.5 10.6777 4.76339 11.1464 5.23223L12.7678 6.85355C13.2366 7.32239 13.8726 7.58579 14.5355 7.58579H19C20.3807 7.58579 21.5 8.70507 21.5 10.0858V17C21.5 18.3807 20.3807 19.5 19 19.5H5C3.61929 19.5 2.5 18.3807 2.5 17V7Z"
                fill="#38bdf8"
              />
              <path
                d="M2.5 10H21.5V17C21.5 18.3807 20.3807 19.5 19 19.5H5C3.61929 19.5 2.5 18.3807 2.5 17V10Z"
                fill="#0284c7"
              />
            </svg>
            <span class="win11-tab-title">Downloads</span>
            <span class="win11-tab-close"><X size={10} /></span>
          </div>
          <span class="win11-tab-new"><Plus size={11} /></span>
        </div>
        <div class="win-controls">
          <span class="win-control-btn"><Minus size={11} /></span>
          <span class="win-control-btn"><Square size={10} /></span>
          <span class="win-control-btn close"><X size={12} /></span>
        </div>
      </div>

      <!-- Address / Search Navigation Bar -->
      <div class="folder-nav-bar">
        <div class="folder-nav-actions">
          <span class="nav-arrow-btn"><ArrowLeft size={13} /></span>
          <span class="nav-arrow-btn disabled"><ArrowRight size={13} /></span>
          <span class="nav-arrow-btn"><ArrowUp size={13} /></span>
          <span class="nav-arrow-btn"><RotateCw size={11} /></span>
        </div>
        <div class="folder-address-box">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="#38bdf8"
            ><path
              d="M4 4h6l2 2h8a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z"
            /></svg
          >
          <span class="crumb">This PC</span>
          <span class="sep">›</span>
          <span class="crumb active">Downloads</span>
        </div>
        <div class="folder-search-input">
          <span class="search-placeholder">Search Downloads</span>
          <Search size={11} class="search-icon" />
        </div>
      </div>

      <!-- Command Ribbon Toolbar -->
      <div class="folder-command-bar">
        <div class="cmd-group">
          <span class="cmd-btn"
            ><Plus size={12} /> New <ChevronDown size={10} /></span
          >
        </div>
        <span class="cmd-sep"></span>
        <div class="cmd-group">
          <span class="cmd-icon-btn"><Scissors size={12} /></span>
          <span class="cmd-icon-btn"><Copy size={12} /></span>
          <span class="cmd-icon-btn"><Trash2 size={12} /></span>
        </div>
        <span class="cmd-sep"></span>
        <div class="cmd-group">
          <span class="cmd-btn">Sort <ChevronDown size={10} /></span>
          <span class="cmd-btn">View <ChevronDown size={10} /></span>
        </div>
      </div>

      <!-- Body: Sidebar + Main Files Grid -->
      <div class="folder-body-split">
        <div class="folder-sidebar">
          <div class="sidebar-link"><Folder size={12} /> Gallery</div>
          <div class="sidebar-link"><Folder size={12} /> Desktop</div>
          <div class="sidebar-link active">
            <Download size={12} /> Downloads
          </div>
          <div class="sidebar-link"><Folder size={12} /> Documents</div>
          <div class="sidebar-link"><Folder size={12} /> Pictures</div>
        </div>
        <div class="folder-content-area">
          <div class="folder-section-header">
            <ChevronDown size={11} />
            <span>Today</span>
          </div>
          <div class="folder-items-grid">
            <div
              class="folder-file-item is-primary"
              data-file-name="proposal_final.pdf"
            >
              <div class="file-icon-wrapper pdf">
                <svg width="24" height="28" viewBox="0 0 24 28" fill="none">
                  <path
                    d="M2 3a2 2 0 0 1 2-2h11l7 7v17a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V3z"
                    fill="#dc2626"
                  />
                  <path d="M15 1v7h7" fill="#f87171" />
                  <text
                    x="12"
                    y="21"
                    font-size="7"
                    font-weight="bold"
                    fill="#ffffff"
                    text-anchor="middle">PDF</text
                  >
                </svg>
              </div>
              <div class="file-info-stack">
                <span class="file-label">proposal_final.pdf</span>
                <span class="file-size-meta"
                  >Adobe Acrobat Document · 2.4 MB</span
                >
              </div>
            </div>
            <div class="folder-file-item" data-file-name="summary_q4.docx">
              <div class="file-icon-wrapper doc">
                <svg width="24" height="28" viewBox="0 0 24 28" fill="none">
                  <path
                    d="M2 3a2 2 0 0 1 2-2h11l7 7v17a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V3z"
                    fill="#2563eb"
                  />
                  <path d="M15 1v7h7" fill="#60a5fa" />
                  <text
                    x="12"
                    y="21"
                    font-size="6"
                    font-weight="bold"
                    fill="#ffffff"
                    text-anchor="middle">DOCX</text
                  >
                </svg>
              </div>
              <div class="file-info-stack">
                <span class="file-label">summary_q4.docx</span>
                <span class="file-size-meta">Word Document · 1.1 MB</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Status Bar -->
      <div class="folder-status-bar">
        <div class="status-left">
          <span>2 items</span>
          <span class="status-sep">|</span>
          <span>1 item selected (2.4 MB)</span>
        </div>
        <div class="status-right">
          <LayoutGrid size={11} />
          <List size={11} />
        </div>
      </div>
    </div>

    <!-- Source Folder Window B: Design Assets (Bottom-Left of Island) -->
    <div
      class="desktop-folder-window source-b"
      data-showcase-folder="design"
      class:is-active={carriedBadge?.source?.includes('Design')}
      transition:scale={{ duration: 320, start: 0.92, opacity: 0 }}
    >
      <!-- Win11 Tab / Title Bar -->
      <div class="folder-tab-bar">
        <div class="folder-tabs-left">
          <div class="win11-tab active">
            <svg
              class="folder-win-icon"
              width="15"
              height="15"
              viewBox="0 0 24 24"
              fill="none"
            >
              <path
                d="M2.5 7C2.5 5.61929 3.61929 4.5 5 4.5H9.37868C10.0417 4.5 10.6777 4.76339 11.1464 5.23223L12.7678 6.85355C13.2366 7.32239 13.8726 7.58579 14.5355 7.58579H19C20.3807 7.58579 21.5 8.70507 21.5 10.0858V17C21.5 18.3807 20.3807 19.5 19 19.5H5C3.61929 19.5 2.5 18.3807 2.5 17V7Z"
                fill="#eab308"
              />
              <path
                d="M2.5 10H21.5V17C21.5 18.3807 20.3807 19.5 19 19.5H5C3.61929 19.5 2.5 18.3807 2.5 17V10Z"
                fill="#facc15"
              />
            </svg>
            <span class="win11-tab-title">Design Assets</span>
            <span class="win11-tab-close"><X size={10} /></span>
          </div>
          <span class="win11-tab-new"><Plus size={11} /></span>
        </div>
        <div class="win-controls">
          <span class="win-control-btn"><Minus size={11} /></span>
          <span class="win-control-btn"><Square size={10} /></span>
          <span class="win-control-btn close"><X size={12} /></span>
        </div>
      </div>

      <!-- Address / Search Navigation Bar -->
      <div class="folder-nav-bar">
        <div class="folder-nav-actions">
          <span class="nav-arrow-btn"><ArrowLeft size={13} /></span>
          <span class="nav-arrow-btn disabled"><ArrowRight size={13} /></span>
          <span class="nav-arrow-btn"><ArrowUp size={13} /></span>
          <span class="nav-arrow-btn"><RotateCw size={11} /></span>
        </div>
        <div class="folder-address-box">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="#facc15"
            ><path
              d="M4 4h6l2 2h8a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z"
            /></svg
          >
          <span class="crumb">This PC</span>
          <span class="sep">›</span>
          <span class="crumb active">Design Assets</span>
        </div>
        <div class="folder-search-input">
          <span class="search-placeholder">Search Assets</span>
          <Search size={11} class="search-icon" />
        </div>
      </div>

      <!-- Command Ribbon Toolbar -->
      <div class="folder-command-bar">
        <div class="cmd-group">
          <span class="cmd-btn"
            ><Plus size={12} /> New <ChevronDown size={10} /></span
          >
        </div>
        <span class="cmd-sep"></span>
        <div class="cmd-group">
          <span class="cmd-icon-btn"><Scissors size={12} /></span>
          <span class="cmd-icon-btn"><Copy size={12} /></span>
          <span class="cmd-icon-btn"><Trash2 size={12} /></span>
        </div>
        <span class="cmd-sep"></span>
        <div class="cmd-group">
          <span class="cmd-btn">Sort <ChevronDown size={10} /></span>
          <span class="cmd-btn">View <ChevronDown size={10} /></span>
        </div>
      </div>

      <!-- Body: Sidebar + Main Files Grid -->
      <div class="folder-body-split">
        <div class="folder-sidebar">
          <div class="sidebar-link"><Folder size={12} /> Desktop</div>
          <div class="sidebar-link"><Folder size={12} /> Downloads</div>
          <div class="sidebar-link active">
            <Folder size={12} /> Design Assets
          </div>
          <div class="sidebar-link"><Folder size={12} /> Documents</div>
          <div class="sidebar-link"><Folder size={12} /> Pictures</div>
        </div>
        <div class="folder-content-area">
          <div class="folder-section-header">
            <ChevronDown size={11} />
            <span>Today</span>
          </div>
          <div class="folder-items-grid">
            <div
              class="folder-file-item is-primary"
              data-file-name="hero_render_4k.png"
            >
              <div class="file-icon-wrapper img">
                <svg width="24" height="28" viewBox="0 0 24 28" fill="none">
                  <path
                    d="M2 3a2 2 0 0 1 2-2h11l7 7v17a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V3z"
                    fill="#059669"
                  />
                  <path d="M15 1v7h7" fill="#34d399" />
                  <circle cx="8" cy="12" r="2" fill="#ffffff" />
                  <path d="M5 20l4-5 3 3 4-6 4 8H5z" fill="#ffffff" />
                </svg>
              </div>
              <div class="file-info-stack">
                <span class="file-label">hero_render_4k.png</span>
                <span class="file-size-meta">PNG Image · 8.9 MB</span>
              </div>
            </div>
            <div class="folder-file-item" data-file-name="brand_logo.svg">
              <div class="file-icon-wrapper vector">
                <svg width="24" height="28" viewBox="0 0 24 28" fill="none">
                  <path
                    d="M2 3a2 2 0 0 1 2-2h11l7 7v17a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V3z"
                    fill="#7c3aed"
                  />
                  <path d="M15 1v7h7" fill="#a78bfa" />
                  <text
                    x="12"
                    y="21"
                    font-size="6.5"
                    font-weight="bold"
                    fill="#ffffff"
                    text-anchor="middle">SVG</text
                  >
                </svg>
              </div>
              <div class="file-info-stack">
                <span class="file-label">brand_logo.svg</span>
                <span class="file-size-meta">Vector Graphic · 450 KB</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Status Bar -->
      <div class="folder-status-bar">
        <div class="status-left">
          <span>2 items</span>
          <span class="status-sep">|</span>
          <span>1 item selected (8.9 MB)</span>
        </div>
        <div class="status-right">
          <LayoutGrid size={11} />
          <List size={11} />
        </div>
      </div>
    </div>

    <!-- Destination Folder Window: Client Deliverables (Right of Island) -->
    <div
      class="desktop-folder-window destination"
      data-showcase-folder="destination"
      class:is-active={carriedBadge?.source?.includes('Client')}
      transition:scale={{ duration: 320, start: 0.92, opacity: 0 }}
    >
      <!-- Win11 Tab / Title Bar -->
      <div class="folder-tab-bar">
        <div class="folder-tabs-left">
          <div class="win11-tab active">
            <svg
              class="folder-win-icon"
              width="15"
              height="15"
              viewBox="0 0 24 24"
              fill="none"
            >
              <path
                d="M2.5 7C2.5 5.61929 3.61929 4.5 5 4.5H9.37868C10.0417 4.5 10.6777 4.76339 11.1464 5.23223L12.7678 6.85355C13.2366 7.32239 13.8726 7.58579 14.5355 7.58579H19C20.3807 7.58579 21.5 8.70507 21.5 10.0858V17C21.5 18.3807 20.3807 19.5 19 19.5H5C3.61929 19.5 2.5 18.3807 2.5 17V7Z"
                fill="#eab308"
              />
              <path
                d="M2.5 10H21.5V17C21.5 18.3807 20.3807 19.5 19 19.5H5C3.61929 19.5 2.5 18.3807 2.5 17V10Z"
                fill="#facc15"
              />
            </svg>
            <span class="win11-tab-title">Client Deliverables</span>
            <span class="win11-tab-close"><X size={10} /></span>
          </div>
          <span class="win11-tab-new"><Plus size={11} /></span>
        </div>
        <div class="win-controls">
          <span class="win-control-btn"><Minus size={11} /></span>
          <span class="win-control-btn"><Square size={10} /></span>
          <span class="win-control-btn close"><X size={12} /></span>
        </div>
      </div>

      <!-- Address / Search Navigation Bar -->
      <div class="folder-nav-bar">
        <div class="folder-nav-actions">
          <span class="nav-arrow-btn"><ArrowLeft size={13} /></span>
          <span class="nav-arrow-btn disabled"><ArrowRight size={13} /></span>
          <span class="nav-arrow-btn"><ArrowUp size={13} /></span>
          <span class="nav-arrow-btn"><RotateCw size={11} /></span>
        </div>
        <div class="folder-address-box">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="#facc15"
            ><path
              d="M4 4h6l2 2h8a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z"
            /></svg
          >
          <span class="crumb">This PC</span>
          <span class="sep">›</span>
          <span class="crumb active">Client Deliverables</span>
        </div>
        <div class="folder-search-input">
          <span class="search-placeholder">Search Folder</span>
          <Search size={11} class="search-icon" />
        </div>
      </div>

      <!-- Command Ribbon Toolbar -->
      <div class="folder-command-bar">
        <div class="cmd-group">
          <span class="cmd-btn"
            ><Plus size={12} /> New <ChevronDown size={10} /></span
          >
        </div>
        <span class="cmd-sep"></span>
        <div class="cmd-group">
          <span class="cmd-icon-btn"><Scissors size={12} /></span>
          <span class="cmd-icon-btn"><Copy size={12} /></span>
          <span class="cmd-icon-btn"><Trash2 size={12} /></span>
        </div>
        <span class="cmd-sep"></span>
        <div class="cmd-group">
          <span class="cmd-btn">Sort <ChevronDown size={10} /></span>
          <span class="cmd-btn">View <ChevronDown size={10} /></span>
        </div>
      </div>

      <!-- Body: Sidebar + Main Files Grid / Dropzone -->
      <div class="folder-body-split">
        <div class="folder-sidebar">
          <div class="sidebar-link"><Folder size={12} /> Desktop</div>
          <div class="sidebar-link"><Folder size={12} /> Downloads</div>
          <div class="sidebar-link active">
            <Folder size={12} /> Deliverables
          </div>
          <div class="sidebar-link"><Folder size={12} /> Documents</div>
          <div class="sidebar-link"><Folder size={12} /> Pictures</div>
        </div>
        <div class="folder-content-area">
          {#if destinationFileCount > 0}
            <div class="folder-section-header">
              <ChevronDown size={11} />
              <span>Received Deliverables ({destinationFileCount})</span>
            </div>
            <div class="folder-items-grid">
              <div
                class="folder-file-item received"
                data-file-name="proposal_final.pdf"
              >
                <div class="file-icon-wrapper pdf">
                  <svg width="24" height="28" viewBox="0 0 24 28" fill="none">
                    <path
                      d="M2 3a2 2 0 0 1 2-2h11l7 7v17a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V3z"
                      fill="#dc2626"
                    />
                    <path d="M15 1v7h7" fill="#f87171" />
                    <text
                      x="12"
                      y="21"
                      font-size="7"
                      font-weight="bold"
                      fill="#ffffff"
                      text-anchor="middle">PDF</text
                    >
                  </svg>
                </div>
                <div class="file-info-stack">
                  <span class="file-label">proposal_final.pdf</span>
                  <span class="file-size-meta"
                    >Adobe Acrobat Document · 2.45 MB</span
                  >
                </div>
              </div>
              {#if destinationFileCount >= 2}
                <div
                  class="folder-file-item received"
                  data-file-name="hero_render_4k.png"
                >
                  <div class="file-icon-wrapper img">
                    <svg width="24" height="28" viewBox="0 0 24 28" fill="none">
                      <path
                        d="M2 3a2 2 0 0 1 2-2h11l7 7v17a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V3z"
                        fill="#059669"
                      />
                      <path d="M15 1v7h7" fill="#34d399" />
                      <circle cx="8" cy="12" r="2" fill="#ffffff" />
                      <path d="M5 20l4-5 3 3 4-6 4 8H5z" fill="#ffffff" />
                    </svg>
                  </div>
                  <div class="file-info-stack">
                    <span class="file-label">hero_render_4k.png</span>
                    <span class="file-size-meta">PNG Image · 8.94 MB</span>
                  </div>
                </div>
              {/if}
            </div>
          {:else}
            <div class="destination-dropzone-prompt">
              <div class="dropzone-box">
                <FolderDown size={30} class="dropzone-svg" />
                <span class="dropzone-dash-box">Drop Staged Files Here</span>
                <span class="dropzone-subhint"
                  >Drag out of Vinta File Shelf to deliver</span
                >
              </div>
            </div>
          {/if}
        </div>
      </div>

      <!-- Status Bar -->
      <div class="folder-status-bar">
        <div class="status-left">
          {#if destinationFileCount > 0}
            <span>2 items</span>
            <span class="status-sep">|</span>
            <span>11.39 MB</span>
          {:else}
            <span>0 items</span>
          {/if}
        </div>
        <div class="status-right">
          <LayoutGrid size={11} />
          <List size={11} />
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .virtual-cursor-container {
    position: fixed;
    top: 0;
    left: 0;
    pointer-events: none;
    z-index: 2147483647;
    will-change: transform;
  }

  .virtual-cursor-container.ghost {
    pointer-events: none;
    z-index: 2147483645;
  }

  .virtual-cursor-container.ghost.ghost-1 {
    opacity: 0.38;
    filter: blur(1px);
  }

  .virtual-cursor-container.ghost.ghost-2 {
    opacity: 0.18;
    filter: blur(2.5px);
  }

  .virtual-cursor-body {
    transform-origin: top left;
    filter: drop-shadow(0 4px 10px rgba(0, 0, 0, 0.65));
    transition: transform 0.12s ease;
  }

  .virtual-cursor-body.is-clicking {
    transform: scale(0.85);
  }

  .cursor-arrow-svg {
    display: block;
  }

  .click-ripple {
    position: fixed;
    pointer-events: none;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    border: 2px solid rgba(255, 255, 255, 0.85);
    background: radial-gradient(
      circle,
      rgba(226, 160, 63, 0.45) 0%,
      transparent 70%
    );
    box-shadow: 0 0 12px rgba(226, 160, 63, 0.65);
    z-index: 2147483646;
    animation: rippleWave 0.55s cubic-bezier(0.1, 0.8, 0.2, 1) forwards;
  }

  .carried-badge-pill {
    position: absolute;
    left: 22px;
    top: 16px;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    background: rgba(22, 22, 28, 0.94);
    border: 1px solid rgba(226, 160, 63, 0.4);
    border-radius: 10px;
    box-shadow:
      0 8px 24px rgba(0, 0, 0, 0.55),
      0 0 12px rgba(226, 160, 63, 0.25);
    white-space: nowrap;
    animation: carriedFloat 1.8s ease-in-out infinite alternate;
  }

  .badge-icon {
    font-size: 16px;
    line-height: 1;
  }

  .badge-text-col {
    display: flex;
    flex-direction: column;
    gap: 1px;
    text-align: left;
  }

  .badge-filename {
    font-size: 11px;
    font-weight: 600;
    color: #ffffff;
    font-family:
      system-ui,
      -apple-system,
      BlinkMacSystemFont,
      sans-serif;
  }

  .badge-source-folder {
    font-size: 9px;
    font-weight: 500;
    color: #e2a03f;
    letter-spacing: 0.3px;
    text-transform: uppercase;
  }

  @keyframes carriedFloat {
    0% {
      transform: translateY(0);
    }
    100% {
      transform: translateY(-3px);
    }
  }

  @keyframes rippleWave {
    0% {
      transform: translate(-50%, -50%) scale(0.3);
      opacity: 1;
    }
    100% {
      transform: translate(-50%, -50%) scale(2.6);
      opacity: 0;
    }
  }

  .showcase-folders-layer {
    position: absolute;
    top: 0;
    left: 0;
    width: 0;
    height: 0;
    pointer-events: none;
    z-index: 2147483640;
  }

  .desktop-folder-window {
    position: absolute;
    display: flex;
    flex-direction: column;
    border-radius: 8px;
    background: #1f1f1f;
    border: 1px solid #333333;
    box-shadow:
      0 20px 50px rgba(0, 0, 0, 0.9),
      0 2px 8px rgba(0, 0, 0, 0.5);
    overflow: hidden;
    pointer-events: auto;
    font-family:
      'Segoe UI Variable',
      'Segoe UI',
      -apple-system,
      system-ui,
      sans-serif;
    transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .desktop-folder-window.source-a {
    left: -740px;
    top: 90px;
    width: 470px;
    height: 285px;
  }

  .desktop-folder-window.source-b {
    left: -740px;
    top: 395px;
    width: 470px;
    height: 275px;
  }

  .desktop-folder-window.destination {
    left: 275px;
    top: 110px;
    width: 480px;
    height: 335px;
  }

  .desktop-folder-window.is-active {
    border-color: rgba(226, 160, 63, 0.7);
    box-shadow:
      0 20px 50px rgba(0, 0, 0, 0.95),
      0 0 20px rgba(226, 160, 63, 0.3);
  }

  /* Win11 Tab / Title Bar */
  .folder-tab-bar {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    height: 38px;
    background: #181818;
    padding: 0 0 0 8px;
    border-bottom: 1px solid #282828;
    user-select: none;
  }

  .folder-tabs-left {
    display: flex;
    align-items: flex-end;
    gap: 4px;
    height: 100%;
  }

  .win11-tab {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 31px;
    padding: 0 10px;
    background: #1f1f1f;
    border: 1px solid #333333;
    border-bottom: none;
    border-radius: 6px 6px 0 0;
    color: #f1f5f9;
    font-size: 11.5px;
    font-weight: 500;
  }

  .win11-tab-title {
    white-space: nowrap;
  }

  .win11-tab-close {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    border-radius: 3px;
    color: rgba(255, 255, 255, 0.55);
    margin-left: 2px;
  }

  .win11-tab-new {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    margin-bottom: 3px;
    border-radius: 4px;
    color: rgba(255, 255, 255, 0.6);
  }

  .win-controls {
    display: flex;
    align-items: center;
    height: 100%;
  }

  .win-control-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 44px;
    height: 100%;
    color: rgba(255, 255, 255, 0.7);
  }

  .win-control-btn.close {
    font-size: 10px;
  }

  /* Address / Navigation Bar */
  .folder-nav-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 38px;
    padding: 0 10px;
    background: #1f1f1f;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .folder-nav-actions {
    display: flex;
    align-items: center;
    gap: 2px;
  }

  .nav-arrow-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 4px;
    color: rgba(255, 255, 255, 0.75);
  }

  .nav-arrow-btn.disabled {
    color: rgba(255, 255, 255, 0.25);
  }

  .folder-address-box {
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 1;
    min-width: 0;
    height: 26px;
    padding: 0 10px;
    background: #282828;
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 4px;
    font-size: 11.5px;
    color: rgba(255, 255, 255, 0.85);
  }

  .folder-address-box .crumb {
    color: rgba(255, 255, 255, 0.7);
  }

  .folder-address-box .crumb.active {
    color: #ffffff;
    font-weight: 500;
  }

  .folder-address-box .sep {
    color: rgba(255, 255, 255, 0.35);
  }

  .folder-search-input {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 140px;
    height: 26px;
    padding: 0 8px;
    background: #282828;
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 4px;
    font-size: 10.5px;
    color: rgba(255, 255, 255, 0.45);
  }

  :global(.search-icon) {
    flex-shrink: 0;
    color: rgba(255, 255, 255, 0.45);
  }

  /* Command Toolbar */
  .folder-command-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 30px;
    padding: 0 10px;
    background: #1f1f1f;
    border-bottom: 1px solid rgba(255, 255, 255, 0.06);
    font-size: 11px;
    color: rgba(255, 255, 255, 0.7);
  }

  .cmd-group {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .cmd-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 2px 6px;
    border-radius: 4px;
    color: rgba(255, 255, 255, 0.75);
    cursor: default;
  }

  .cmd-icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 22px;
    border-radius: 4px;
    color: rgba(255, 255, 255, 0.75);
  }

  .cmd-sep {
    width: 1px;
    height: 14px;
    background: rgba(255, 255, 255, 0.12);
    margin: 0 2px;
  }

  /* Body: Sidebar + Main Content */
  .folder-body-split {
    display: flex;
    flex-direction: row;
    flex: 1;
    min-height: 0;
    background: #1f1f1f;
  }

  .folder-sidebar {
    display: flex;
    flex-direction: column;
    gap: 2px;
    width: 115px;
    padding: 6px 4px;
    background: #191919;
    border-right: 1px solid rgba(255, 255, 255, 0.06);
    user-select: none;
  }

  .sidebar-link {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 5px 8px;
    border-radius: 4px;
    font-size: 11px;
    color: rgba(255, 255, 255, 0.7);
  }

  .sidebar-link.active {
    background: rgba(255, 255, 255, 0.08);
    color: #ffffff;
    font-weight: 500;
  }

  .folder-content-area {
    display: flex;
    flex-direction: column;
    flex: 1;
    padding: 8px 12px;
    overflow: hidden;
  }

  .folder-section-header {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 11px;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.65);
    margin-bottom: 6px;
    user-select: none;
  }

  .folder-items-grid {
    display: flex;
    flex-direction: column;
    gap: 4px;
    flex: 1;
  }

  .folder-file-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 10px;
    border-radius: 5px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.05);
    transition:
      background 0.15s ease,
      border-color 0.15s ease;
  }

  .folder-file-item.is-primary {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(255, 255, 255, 0.18);
  }

  .folder-file-item.received {
    border-color: rgba(255, 255, 255, 0.08);
    background: rgba(255, 255, 255, 0.04);
  }

  .file-icon-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 30px;
    flex-shrink: 0;
  }

  .file-info-stack {
    display: flex;
    flex-direction: column;
    gap: 1px;
    min-width: 0;
    flex: 1;
  }

  .file-label {
    font-size: 12px;
    font-weight: 500;
    color: #f8fafc;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .file-size-meta {
    font-size: 10px;
    color: rgba(255, 255, 255, 0.45);
    white-space: nowrap;
  }

  .destination-dropzone-prompt {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    flex: 1;
    min-height: 120px;
  }

  .dropzone-box {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 5px;
    padding: 18px 22px;
    border: 1.5px dashed rgba(255, 255, 255, 0.18);
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.02);
    color: rgba(255, 255, 255, 0.5);
    text-align: center;
  }

  .dropzone-svg {
    stroke: rgba(255, 255, 255, 0.4);
    margin-bottom: 2px;
  }

  .dropzone-dash-box {
    font-size: 12px;
    font-weight: 500;
    color: #e2e8f0;
  }

  .dropzone-subhint {
    font-size: 10px;
    color: rgba(255, 255, 255, 0.4);
  }

  .folder-status-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 22px;
    padding: 0 10px;
    background: #191919;
    border-top: 1px solid rgba(255, 255, 255, 0.06);
    font-size: 10.5px;
    color: rgba(255, 255, 255, 0.5);
    user-select: none;
  }

  .status-left {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .status-right {
    display: flex;
    align-items: center;
    gap: 6px;
    color: rgba(255, 255, 255, 0.4);
  }

  .status-sep {
    color: rgba(255, 255, 255, 0.2);
  }

  :global(.vinta-virtual-hover) {
    filter: brightness(1.22) !important;
  }

  :global(button.vinta-virtual-hover:not(.hero-play-btn)),
  :global(.dock-btn.vinta-virtual-hover) {
    background: rgba(255, 255, 255, 0.22) !important;
    transform: scale(1.12) !important;
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.45) !important;
  }

  :global(.hero-play-btn),
  :global(.hero-play-btn.vinta-virtual-hover) {
    background: #ffffff !important;
    opacity: 1 !important;
  }

  :global(.hero-play-btn svg) {
    fill: currentColor !important;
    color: #0d0d14 !important;
    opacity: 1 !important;
  }

  :global(.weather-art-link.vinta-virtual-hover) {
    transform: scale(1.1) !important;
    filter: brightness(1.25) drop-shadow(0 0 16px rgba(255, 255, 255, 0.4)) !important;
  }

  :global(.app-tile.vinta-virtual-hover) {
    background: rgba(255, 255, 255, 0.16) !important;
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.3) !important;
  }

  :global(.nav-item.vinta-virtual-hover) {
    background: rgba(255, 255, 255, 0.12) !important;
    color: #ffffff !important;
  }

  :global(.option-item.vinta-virtual-hover) {
    background: rgba(255, 255, 255, 0.18) !important;
  }

  :global(.collapse-btn.vinta-virtual-hover) {
    background: rgba(255, 255, 255, 0.22) !important;
    transform: scale(1.15) !important;
  }

  :global(.day-cell.vinta-virtual-hover) {
    background: rgba(255, 255, 255, 0.16) !important;
    border-radius: 8px !important;
  }

  :global(.calc-btn.vinta-virtual-hover) {
    filter: brightness(1.3) !important;
    transform: scale(1.05) !important;
  }
</style>
