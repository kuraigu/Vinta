<script module lang="ts">
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 368, h: 44, r: 16 },
    expanded: { w: 400, h: 420, r: 28 },
    borderColor: 'var(--vinta-island-border-idle)',
    glowColor: 'var(--vinta-island-glow-idle)'
  }
</script>

<script lang="ts">
  import { onMount } from 'svelte'
  import { Cpu, Activity, Zap } from 'lucide-svelte'
  import BackButton from '../components/BackButton.svelte'
  import BentoCard from '../components/BentoCard.svelte'
  import ScrollView from '../components/ScrollView.svelte'
  import { tStore } from '../store'
  import {
    systemTelemetryStore,
    cpuHistoryStore,
    ramHistoryStore,
    gpuHistoryStore,
    subscribeToTelemetry
  } from '../stores/systemTelemetry'

  interface Props {
    expanded?: boolean
  }

  let { expanded = true }: Props = $props()

  let telemetry = $derived($systemTelemetryStore)
  let cpuHistory = $derived($cpuHistoryStore)
  let ramHistory = $derived($ramHistoryStore)
  let gpuHistory = $derived($gpuHistoryStore)

  onMount(() => {
    const unsubscribe = subscribeToTelemetry()
    return () => {
      unsubscribe()
    }
  })

  /**
   * Generates SVG polyline coordinate points for trajectory sparklines.
   *
   * @param history - Array of numeric load values from 0 to 100.
   * @param width - Viewport pixel width of the target sparkline.
   * @param height - Viewport pixel height of the target sparkline.
   * @returns Formatted SVG coordinates string.
   */
  function renderSparkline(
    history: number[],
    width = 340,
    height = 32
  ): string {
    if (!history || history.length < 2) return `0,${height} ${width},${height}`
    const step = width / (history.length - 1)
    return history
      .map((val, idx) => {
        const x = Math.round(idx * step * 10) / 10
        const clamped = Math.max(0, Math.min(100, val))
        const y =
          Math.round((height - (clamped / 100) * (height - 6) - 3) * 10) / 10
        return `${x},${y}`
      })
      .join(' ')
  }

  let cpuLine = $derived(renderSparkline(cpuHistory, 340, 32))
  let cpuArea = $derived(`0,32 ${cpuLine} 340,32`)

  let ramLine = $derived(renderSparkline(ramHistory, 340, 32))
  let ramArea = $derived(`0,32 ${ramLine} 340,32`)

  let gpuLine = $derived(renderSparkline(gpuHistory, 340, 32))
  let gpuArea = $derived(`0,32 ${gpuLine} 340,32`)

  /**
   * Formats a megabyte count into a legible gigabyte string with one decimal place.
   *
   * @param mb - Memory allocation in megabytes.
   * @returns Formatted string in gigabytes.
   */
  function formatGB(mb: number): string {
    if (!mb || isNaN(mb)) return '0.0'
    return (mb / 1024).toFixed(1)
  }
</script>

<div
  data-vinta="performance-page"
  class="performance-page-layout"
  class:expanded
>
  <!-- Main Metrics Feed -->
  <ScrollView class="perf-scroll-view">
    <div class="perf-cards-stack">
      <!-- CPU Metric Card -->
      <BentoCard variant="bento" color="danger" class="perf-metric-card">
        <div class="perf-card-header">
          <div class="perf-card-label">
            <Cpu size={14} class="perf-card-icon cpu" />
            <span class="perf-card-name">{$tStore('performance.processor')}</span>
          </div>
          <span class="perf-card-percent cpu">{telemetry.cpuPercent}%</span>
        </div>

        <div class="perf-sparkline-wrap">
          <svg
            class="perf-sparkline-svg"
            width="100%"
            height="32"
            viewBox="0 0 340 32"
            preserveAspectRatio="none"
          >
            <defs>
              <linearGradient id="grad-cpu" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="0%"
                  stop-color="var(--vinta-accent-danger, #d55e00)"
                  stop-opacity="0.32"
                />
                <stop
                  offset="100%"
                  stop-color="var(--vinta-accent-danger, #d55e00)"
                  stop-opacity="0.0"
                />
              </linearGradient>
            </defs>
            <polygon points={cpuArea} fill="url(#grad-cpu)" />
            <polyline points={cpuLine} class="sparkline-trace cpu" />
          </svg>
        </div>

        <div class="perf-card-meta-row">
          <span class="perf-meta-sub">{$tStore('performance.usage')}</span>
          <span class="perf-meta-val">
            {telemetry.cpuPercent > 75
              ? $tStore('performance.high')
              : telemetry.cpuPercent > 35
                ? $tStore('performance.moderate')
                : $tStore('performance.normal')}
          </span>
        </div>
      </BentoCard>

      <!-- RAM Metric Card -->
      <BentoCard variant="bento" color="primary" class="perf-metric-card">
        <div class="perf-card-header">
          <div class="perf-card-label">
            <Activity size={14} class="perf-card-icon ram" />
            <span class="perf-card-name">{$tStore('performance.memory')}</span>
          </div>
          <span class="perf-card-percent ram">{telemetry.ramPercent}%</span>
        </div>

        <div class="perf-sparkline-wrap">
          <svg
            class="perf-sparkline-svg"
            width="100%"
            height="32"
            viewBox="0 0 340 32"
            preserveAspectRatio="none"
          >
            <defs>
              <linearGradient id="grad-ram" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="0%"
                  stop-color="var(--vinta-primary, #56b4e9)"
                  stop-opacity="0.32"
                />
                <stop
                  offset="100%"
                  stop-color="var(--vinta-primary, #56b4e9)"
                  stop-opacity="0.0"
                />
              </linearGradient>
            </defs>
            <polygon points={ramArea} fill="url(#grad-ram)" />
            <polyline points={ramLine} class="sparkline-trace ram" />
          </svg>
        </div>

        <div class="perf-card-meta-row">
          <span class="perf-meta-sub">
            {formatGB(telemetry.ramUsedMB)} GB / {formatGB(
              telemetry.ramTotalMB
            )} GB
          </span>
          <span class="perf-meta-val">
            {$tStore('performance.gb_free', {
              free: formatGB(telemetry.ramTotalMB - telemetry.ramUsedMB)
            })}
          </span>
        </div>
      </BentoCard>

      <!-- GPU Metric Card -->
      <BentoCard variant="bento" color="blue" class="perf-metric-card">
        <div class="perf-card-header">
          <div class="perf-card-label">
            <Zap size={14} class="perf-card-icon gpu" />
            <span class="perf-card-name">{$tStore('performance.graphics')}</span>
          </div>
          <span class="perf-card-percent gpu">{telemetry.gpuPercent}%</span>
        </div>

        <div class="perf-sparkline-wrap">
          <svg
            class="perf-sparkline-svg"
            width="100%"
            height="32"
            viewBox="0 0 340 32"
            preserveAspectRatio="none"
          >
            <defs>
              <linearGradient id="grad-gpu" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="0%"
                  stop-color="var(--vinta-accent-purple, #cc79a7)"
                  stop-opacity="0.32"
                />
                <stop
                  offset="100%"
                  stop-color="var(--vinta-accent-purple, #cc79a7)"
                  stop-opacity="0.0"
                />
              </linearGradient>
            </defs>
            <polygon points={gpuArea} fill="url(#grad-gpu)" />
            <polyline points={gpuLine} class="sparkline-trace gpu" />
          </svg>
        </div>

        <div class="perf-card-meta-row">
          <span class="perf-meta-sub">{$tStore('performance.engine_3d')}</span>
          <span class="perf-meta-val">
            {telemetry.gpuPercent > 5
              ? $tStore('performance.active')
              : $tStore('performance.idle')}
          </span>
        </div>
      </BentoCard>
    </div>
  </ScrollView>

  <!-- Bottom Navigation Footer Bar (Swapped with Top Dock) -->
  <footer class="perf-meta-footer">
    <div class="perf-footer-left">
      <BackButton title={$tStore('performance.back_to_apps')} />
      <span class="perf-footer-title">{$tStore('performance.title')}</span>
    </div>
  </footer>
</div>

<style src="./styles/PerformancePage.css"></style>
