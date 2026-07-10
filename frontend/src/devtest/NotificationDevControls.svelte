<script lang="ts">
  import { notifData } from '../store'

  function handleTest1() {
    notifData.update(n => [{ id: Date.now(), title: "System Alert", message: "Battery is running low (10%)." }, ...n])
  }

  function handleTest2() {
    notifData.update(n => [{ id: Date.now() + 1, title: "Achievement Unlocked", message: "You completed the first task!" }, ...n])
  }
</script>

<button class="test-btn" on:click={handleTest1}>Send 1</button>
<button class="test-btn" on:click={handleTest2}>Send 2</button>

<style>
  button {
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: white;
    padding: 5px 12px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.8rem;
    transition: background 0.2s;
  }

  button:hover {
    background: rgba(255, 255, 255, 0.2);
  }
</style>
