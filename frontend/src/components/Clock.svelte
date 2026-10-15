<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import {
    timeFormatStore,
    timezoneStore,
    dateFormatStore,
    formatUserDate
  } from '../store'

  let { expanded = false } = $props<{ expanded?: boolean }>()

  let time = $state(new Date())
  let timer: ReturnType<typeof setInterval>

  onMount(() => {
    timer = setInterval(() => {
      time = new Date()
    }, 1000)
  })

  onDestroy(() => {
    clearInterval(timer)
  })

  // Selected timezone from store (defaulting to Asia/Manila)
  let currentTimezone = $derived($timezoneStore || 'Asia/Manila')

  // Format parts according to selected timezone and date format
  let timeParts = $derived.by(() => {
    try {
      const formatter = new Intl.DateTimeFormat('en-US', {
        timeZone: currentTimezone,
        hour: 'numeric',
        minute: 'numeric',
        second: 'numeric',
        hour12: false,
        weekday: 'long',
        month: 'short',
        monthNumeric: '2-digit',
        day: '2-digit',
        year: 'numeric'
      } as Intl.DateTimeFormatOptions)
      const parts = formatter.formatToParts(time)
      const getVal = (type: string) =>
        parts.find((p) => p.type === type)?.value || '0'
      const h24 = parseInt(getVal('hour'), 10)
      const m = getVal('minute').padStart(2, '0')
      const s = parseInt(getVal('second'), 10)
      const weekday = getVal('weekday')
      const month = getVal('month')
      const monthNum = (time.getMonth() + 1).toString().padStart(2, '0')
      const day = getVal('day').padStart(2, '0')
      const year = getVal('year')

      const dateBadgeText = formatUserDate(time, $dateFormatStore, {
        includeWeekday: true
      })

      return {
        h24,
        m,
        s,
        fullDateStr: dateBadgeText
      }
    } catch {
      const dateBadgeText = formatUserDate(time, $dateFormatStore, {
        includeWeekday: true
      })

      return {
        h24: time.getHours(),
        m: time.getMinutes().toString().padStart(2, '0'),
        s: time.getSeconds(),
        fullDateStr: dateBadgeText
      }
    }
  })

  // Formatting helpers
  let is24H = $derived($timeFormatStore === '24h')
  let hours = $derived.by(() => {
    const h24 = timeParts.h24
    if (is24H) return h24.toString().padStart(2, '0')
    const h = h24 % 12 || 12
    return h.toString().padStart(2, '0')
  })
  let minutes = $derived(timeParts.m)
  let ampm = $derived(timeParts.h24 >= 12 ? 'PM' : 'AM')
  let seconds = $derived(timeParts.s)
  let fullDateStr = $derived(timeParts.fullDateStr)

  // GMT Offset calculation based on selected timezone
  let gmtOffset = $derived.by(() => {
    try {
      const parts = new Intl.DateTimeFormat('en-US', {
        timeZone: currentTimezone,
        timeZoneName: 'shortOffset'
      }).formatToParts(time)
      const tzPart = parts.find((p) => p.type === 'timeZoneName')
      if (tzPart && tzPart.value.startsWith('GMT')) {
        return tzPart.value
      }
    } catch {
      // Fallback to local offset if timezone string is invalid
    }
    const offsetMin = -time.getTimezoneOffset()
    const sign = offsetMin >= 0 ? '+' : '-'
    const absMin = Math.abs(offsetMin)
    const h = Math.floor(absMin / 60)
    const m = absMin % 60
    return m === 0
      ? `GMT${sign}${h}`
      : `GMT${sign}${h}:${m.toString().padStart(2, '0')}`
  })

  let timezoneName = $derived(
    currentTimezone.split('/').pop()?.replace('_', ' ') || 'Manila'
  )
</script>

<div class="clock-container {expanded ? 'expanded' : 'minimal'}">
  {#if !expanded}
    <div class="time">
      {hours}:{minutes}{#if !is24H}<span class="ampm-minimal">
          {ampm}</span
        >{/if}
    </div>
  {:else}
    <div class="clock-card">
      <div class="ambient-glow"></div>

      <!-- Date Badge Header -->
      <div class="date-header">
        <span class="date-badge">{fullDateStr}</span>
      </div>

      <!-- Circular Progress Dial Hub -->
      <div class="circular-dial-wrapper">
        <!-- SVG Thick Glass Capsule Ring with Linear Gradient & Glow -->
        <svg class="dial-svg" viewBox="0 0 150 150">
          <defs>
            <linearGradient
              id="clockRingGrad"
              x1="0%"
              y1="0%"
              x2="100%"
              y2="100%"
            >
              <stop offset="0%" stop-color="var(--c-primary)" />
              <stop offset="100%" stop-color="var(--c-purple)" />
            </linearGradient>
            <filter id="ringGlow" x="-20%" y="-20%" width="140%" height="140%">
              <feDropShadow
                dx="0"
                dy="2"
                stdDeviation="4"
                flood-color="var(--c-primary)"
                flood-opacity="0.4"
              />
            </filter>
          </defs>
          <circle class="dial-bg-ring" cx="75" cy="75" r="58" />
          <circle
            class="dial-fill-ring"
            class:no-transition={seconds === 0}
            cx="75"
            cy="75"
            r="58"
            style="stroke-dasharray: 364.42; stroke-dashoffset: {364.42 -
              (seconds / 60) * 364.42};"
          />
        </svg>

        <!-- Dial Inner Center Content -->
        <div class="dial-center-stack">
          {#if !is24H}
            <span class="dial-ampm-badge">{ampm}</span>
          {/if}
          <div class="dial-time-digits">
            <span class="hours">{hours}</span>
            <span class="colon">:</span>
            <span class="minutes">{minutes}</span>
          </div>
          <span class="dial-seconds-pill"
            >{seconds.toString().padStart(2, '0')}s</span
          >
        </div>
      </div>

      <!-- Normalized Bottom Footer Bar -->
      <div class="bottom-footer-bar">
        <span class="footer-meta-item timezone-text" title={timezoneName}
          >{timezoneName}</span
        >
        <span class="footer-dot">•</span>
        <span class="footer-meta-item">{gmtOffset}</span>
      </div>
    </div>
  {/if}
</div>

<style>
  .clock-container {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
  }

  .clock-container.minimal {
    width: auto;
    justify-content: flex-start;
  }

  .clock-container.minimal .time {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--c-text-primary);
    font-variant-numeric: tabular-nums;
    margin: 0;
    padding: 0;
    display: inline-flex;
    align-items: center;
  }

  /* EXPANDED DESIGN */
  .clock-card {
    position: relative;
    width: 100%;
    height: 100%;
    padding: 0.9rem 1.1rem;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    align-items: center;
    overflow: hidden;
  }

  .ambient-glow {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 220px;
    height: 140px;
    background: radial-gradient(
      circle,
      rgba(99, 102, 241, 0.15) 0%,
      rgba(168, 85, 247, 0.08) 50%,
      transparent 75%
    );
    filter: blur(28px);
    pointer-events: none;
  }

  .date-header {
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2;
  }

  .date-badge {
    font-size: 0.78rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--c-text-muted);
    background: var(--c-item-bg);
    padding: 4px 14px;
    border-radius: 20px;
    border: 1px solid var(--c-border);
  }

  /* CIRCULAR DIAL HUB DESIGN (FLOATING RING) */
  .circular-dial-wrapper {
    position: relative;
    width: 136px;
    height: 136px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2;
    margin: 4px 0;
  }

  .dial-svg {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%) rotate(-90deg);
    width: 154px;
    height: 154px;
    pointer-events: none;
  }

  .dial-bg-ring {
    fill: none;
    stroke: var(--c-contrast-bg);
    stroke-width: 14;
    opacity: 0.8;
  }

  .dial-fill-ring {
    fill: none;
    stroke: url(#clockRingGrad);
    stroke-width: 14;
    stroke-linecap: round;
    filter: url(#ringGlow);
    transition: stroke-dashoffset 0.3s linear;
  }

  .dial-fill-ring.no-transition {
    transition: none !important;
  }

  .dial-center-stack {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1px;
    z-index: 3;
  }

  .dial-ampm-badge {
    font-size: 0.65rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    color: var(--c-primary);
    text-transform: uppercase;
    line-height: 1;
  }

  .dial-time-digits {
    display: flex;
    align-items: center;
    font-size: 1.85rem;
    font-weight: 800;
    letter-spacing: -1px;
    line-height: 1.1;
    color: var(--c-text-primary);
  }

  .colon {
    margin: 0 -1px;
    opacity: 0.9;
  }

  .dial-seconds-pill {
    font-size: 0.68rem;
    font-weight: 700;
    color: var(--c-text-muted);
    font-variant-numeric: tabular-nums;
    line-height: 1;
    letter-spacing: 0.04em;
  }

  /* NORMALIZED BOTTOM FOOTER BAR */
  .bottom-footer-bar {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    z-index: 2;
    padding-top: 2px;
    max-width: 100%;
  }

  .footer-meta-item {
    font-size: 0.82rem;
    font-weight: 700;
    color: var(--c-text-primary);
    letter-spacing: 0.02em;
  }

  .footer-meta-item.timezone-text {
    max-width: 140px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .footer-dot {
    font-size: 0.75rem;
    color: var(--c-text-muted);
    opacity: 0.5;
  }
</style>
