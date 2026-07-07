<script lang="ts">
  import { onMount, onDestroy } from 'svelte'

  export let expanded: boolean = false

  let time = new Date()
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
    background: rgba(255, 255, 255, 0.05);
    border-radius: 24px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-sizing: border-box;
    backdrop-filter: blur(10px);
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.1),
      0 10px 30px rgba(0, 0, 0, 0.2);
    width: 100%;
    height: 100%;
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
    margin-left: 8px; /* Offset from the green dot */
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
