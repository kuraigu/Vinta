<!--
  @component SystemStats
  @description Mini telemetry display showing CPU load, RAM utilization, and GPU usage percentages.
-->
<script lang="ts">
  import { onMount, getContext } from 'svelte'
  import { Activity } from 'lucide-svelte'
  import {
    systemTelemetryStore,
    subscribeToTelemetry
  } from '../stores/systemTelemetry'
  import { islandTooltip } from '../services/islandTooltip'
  import BentoCard from './BentoCard.svelte'
  import ActionCapsule from './ActionCapsule.svelte'
  import IconBadge from './IconBadge.svelte'
  import PerformancePage, {
    config as performanceConfig
  } from '../pages/PerformancePage.svelte'
  import type { IslandConfig } from '../store'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  interface Props {
    expanded?: boolean
  }

  let { expanded = false }: Props = $props()

  let telemetry = $derived($systemTelemetryStore)

  onMount(() => {
    const unsubscribe = subscribeToTelemetry()
    return () => {
      unsubscribe()
    }
  })

  function handleCardClick() {
    if (typeof navigate === 'function') {
      navigate(PerformancePage, performanceConfig, true)
    }
  }

  const radius = 22
  const circumference = 2 * Math.PI * radius

  let cpuOffset = $derived(
    circumference - (telemetry.cpuPercent / 100) * circumference
  )
  let ramOffset = $derived(
    circumference - (telemetry.ramPercent / 100) * circumference
  )
  let gpuOffset = $derived(
    circumference - (telemetry.gpuPercent / 100) * circumference
  )
</script>

{#if expanded}
  <BentoCard
    variant="bento"
    class="system-stats-glance-card"
    onclick={handleCardClick}
  >
    <div class="glance-header">
      <div class="glance-title-group">
        <IconBadge icon={Activity} color="surface" size="sm" />
        <span>System Stats</span>
      </div>
      <div class="glance-actions-group">
        <ActionCapsule
          label="View"
          onclick={(e) => {
            e.stopPropagation()
            handleCardClick()
          }}
        />
      </div>
    </div>

    <div class="system-stats-rings-row">
      <!-- CPU Ring -->
      <div
        class="stat-item"
        use:islandTooltip={{ text: `CPU: ${telemetry.cpuPercent}%` }}
      >
        <div class="ring-container">
          <svg class="ring" width="56" height="56" viewBox="0 0 56 56">
            <circle class="ring-bg" cx="28" cy="28" r={radius} />
            <circle
              class="ring-progress cpu"
              cx="28"
              cy="28"
              r={radius}
              style="stroke-dasharray: {circumference}; stroke-dashoffset: {cpuOffset};"
            />
          </svg>
          <div class="ring-label">
            <span class="value">{telemetry.cpuPercent}%</span>
          </div>
        </div>
        <span class="label">CPU</span>
      </div>

      <!-- RAM Ring -->
      <div
        class="stat-item"
        use:islandTooltip={{
          text: `RAM: ${telemetry.ramUsedMB} MB / ${telemetry.ramTotalMB} MB (${telemetry.ramPercent}%)`
        }}
      >
        <div class="ring-container">
          <svg class="ring" width="56" height="56" viewBox="0 0 56 56">
            <circle class="ring-bg" cx="28" cy="28" r={radius} />
            <circle
              class="ring-progress ram"
              cx="28"
              cy="28"
              r={radius}
              style="stroke-dasharray: {circumference}; stroke-dashoffset: {ramOffset};"
            />
          </svg>
          <div class="ring-label">
            <span class="value">{telemetry.ramPercent}%</span>
          </div>
        </div>
        <span class="label">RAM</span>
      </div>

      <!-- GPU Ring -->
      <div
        class="stat-item"
        use:islandTooltip={{ text: `GPU: ${telemetry.gpuPercent}%` }}
      >
        <div class="ring-container">
          <svg class="ring" width="56" height="56" viewBox="0 0 56 56">
            <circle class="ring-bg" cx="28" cy="28" r={radius} />
            <circle
              class="ring-progress gpu"
              cx="28"
              cy="28"
              r={radius}
              style="stroke-dasharray: {circumference}; stroke-dashoffset: {gpuOffset};"
            />
          </svg>
          <div class="ring-label">
            <span class="value">{telemetry.gpuPercent}%</span>
          </div>
        </div>
        <span class="label">GPU</span>
      </div>
    </div>
  </BentoCard>
{/if}

<style src="./styles/SystemStats.css"></style>
