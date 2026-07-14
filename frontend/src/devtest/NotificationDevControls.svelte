<script lang="ts">
  import { notifData } from '../store'

  import { SendTestNotification } from '../../wailsjs/go/backend/App.js'

  function triggerNativeNotification() {
    SendTestNotification()
  }
</script>

<button class="action-btn" onclick={triggerNativeNotification}>Send Native Windows Toast</button>

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
