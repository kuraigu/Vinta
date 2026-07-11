<script lang="ts">
  import { onMount, onDestroy } from 'svelte'

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
</script>

<div class="clock-container {expanded ? 'expanded' : 'minimal'}">
  <div class="time">
    {time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
  </div>
  {#if expanded}
    <div class="seconds">
      {time.getSeconds().toString().padStart(2, '0')} seconds
    </div>
  {/if}
</div>

<style>
  .clock-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }

  .clock-container.minimal {
    flex-direction: row;
  }

  .clock-container.expanded {
    width: 100%;
    height: 100%;
    box-sizing: border-box;
    padding: 1rem;
  }

  .time {
    line-height: 1;
  }

  .expanded .time {
    font-size: 3.2rem;
    font-weight: 800;
    letter-spacing: -2px;
    margin-bottom: 5px;
    background: linear-gradient(180deg, #ffffff 0%, #a0a0a0 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    filter: drop-shadow(0 4px 12px rgba(255, 255, 255, 0.15));
  }

  .minimal .time {
    font-size: 1rem;
    font-weight: 400;
  }

  .seconds {
    font-size: 1.1rem;
    color: #808080;
    font-weight: 600;
    margin-top: -5px;
    letter-spacing: 2px;
    text-transform: uppercase;
  }
</style>
