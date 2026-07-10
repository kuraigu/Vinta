<script lang="ts">
  import { notifData } from '../store'

  function handleTest1() {
    notifData.update((n) => [
      {
        id: Date.now(),
        title: 'System Alert',
        message: 'Battery is running low (10%).'
      },
      ...n
    ])
  }

  function handleTest2() {
    notifData.update((n) => [
      {
        id: Date.now() + 1,
        title: 'Achievement Unlocked',
        message: 'You completed the first task!'
      },
      ...n
    ])
  }
</script>

<button class="action-btn" on:click={handleTest1}>Send 1</button>
<button class="action-btn" on:click={handleTest2}>Send 2</button>

<style>
  .action-btn {
    background: var(--c-item-bg);
    border: 1px solid var(--c-border);
    color: var(--c-text-primary);
    padding: 6px 12px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.8rem;
    transition: background 0.2s;
  }

  .action-btn:hover {
    background: var(--c-hover);
  }
</style>
