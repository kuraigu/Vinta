<script lang="ts">
  import { Events } from '@wailsio/runtime'
  import { parseAumid } from '../services/iconService'
  import {
    SendTestNotification,
    SendTestLongMessageToast,
    SendTestNoTitleToast,
    SendTestBadge
  } from '../../bindings/Vinta/backend/devtest/devtest.js'

  let selectedAction = $state('native')
  let customAppName = $state('')
  let customIconUrl = $state('')

  const actionLabels: Record<string, string> = {
    native: 'Native Toast',
    notitle: 'No Title',
    longmsg: 'Long Message',
    badge: 'Badge Notif'
  }

  function triggerAction() {
    const mockApp = customAppName.trim()

    if (mockApp !== '') {
      const parsed = parseAumid(mockApp)
      Events.Emit('notification', {
        appName: mockApp,
        title: 'Mock App Notification',
        message: `Parsed to: ${parsed.displayName} (${parsed.processName})`,
        icon: customIconUrl.trim(),
        heroImage: ''
      })
      return
    }

    // No custom name: fire a real Windows toast through the backend.
    if (selectedAction === 'native') SendTestNotification()
    if (selectedAction === 'notitle') SendTestNoTitleToast()
    if (selectedAction === 'longmsg') SendTestLongMessageToast()
    if (selectedAction === 'badge') {
      const count = Math.floor(Math.random() * 9) + 1
      SendTestBadge(count.toString())
    }
  }
</script>

<div class="dev-controls">
  <input
    type="text"
    class="action-input icon-input"
    placeholder="Test custom icon URL..."
    bind:value={customIconUrl}
  />

  <input
    type="text"
    class="action-input"
    placeholder="Mock App or AUMID..."
    bind:value={customAppName}
  />

  <button
    class="action-select"
    onclick={() => {
      const actions = ['native', 'notitle', 'longmsg', 'badge']
      selectedAction =
        actions[(actions.indexOf(selectedAction) + 1) % actions.length]
    }}
  >
    {actionLabels[selectedAction]} 🔁
  </button>
  <button class="action-btn" onclick={triggerAction}>Execute</button>
</div>

<style>
  .dev-controls {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
    max-width: 250px;
    justify-content: center;
  }

  .action-input,
  .action-select {
    background: var(--c-bg);
    color: var(--c-text-primary);
    border: 1px solid var(--c-border);
    padding: 6px 8px;
    border-radius: 6px;
    font-size: 0.8rem;
    outline: none;
  }

  .action-input {
    width: 160px;
  }

  .action-input:focus,
  .action-select:focus {
    border-color: var(--c-primary);
  }

  .action-btn {
    background: var(--c-item-bg);
    border: 1px solid var(--c-border);
    color: var(--c-text-primary);
    padding: 6px 12px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.8rem;
    position: relative;
    z-index: 1;
    /* Removed transition because it forces WebView2 to create an overlapping compositing layer */
  }

  .action-btn:hover {
    background: var(--c-hover);
  }
</style>
