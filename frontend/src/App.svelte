<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  let expanded = false;
  let time = new Date();
  let timer: ReturnType<typeof setInterval>;

  onMount(() => {
    timer = setInterval(() => {
      time = new Date();
    }, 1000);
  });

  onDestroy(() => {
    clearInterval(timer);
  });

  function toggle() {
    expanded = !expanded;
  }
</script>

<main>
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div 
    class="island {expanded ? 'expanded' : ''}" 
    on:click={toggle}
    on:mouseenter={() => expanded = true}
    on:mouseleave={() => expanded = false}
  >
    {#if expanded}
      <div class="content expanded-content">
        <div class="time">{time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</div>
        <div class="date">{time.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })}</div>
      </div>
    {:else}
      <div class="content minimal-content">
        <!-- Minimal state indicator -->
        <div class="dot"></div>
      </div>
    {/if}
  </div>
</main>

<style>
  main {
    width: 100vw;
    height: 100vh;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    padding-top: 15px; /* Distance from top of screen */
  }

  .island {
    background-color: #000000;
    border-radius: 40px;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.7);
    cursor: pointer;
    overflow: hidden;
    
    /* Default collapsed state */
    width: 150px;
    height: 35px;
    
    /* Smooth spring-like transition */
    transition: all 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
    
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
  }

  .island.expanded {
    width: 320px;
    height: 130px;
    border-radius: 35px;
  }

  .content {
    /* Smooth fade in for content */
    animation: fadeIn 0.4s ease-in-out;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }

  .expanded-content {
    width: 100%;
    height: 100%;
  }

  .time {
    font-size: 3rem;
    font-weight: 700;
    letter-spacing: -1.5px;
    line-height: 1;
    margin-bottom: 5px;
  }

  .date {
    font-size: 1rem;
    color: #a0a0a0;
    font-weight: 500;
  }

  .minimal-content {
    width: 100%;
    height: 100%;
  }

  .dot {
    width: 6px;
    height: 6px;
    background-color: #34c759; /* Green active dot */
    border-radius: 50%;
  }

  @keyframes fadeIn {
    0% { opacity: 0; transform: scale(0.9); }
    50% { opacity: 0; }
    100% { opacity: 1; transform: scale(1); }
  }
</style>
