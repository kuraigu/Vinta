<script lang="ts">
  import { appIcons, gameIcons, findAppIconConflicts } from '../appIcons'
  import {
    SendTestNotification,
    SendTestLongMessageToast,
    SendTestNoTitleToast,
    SendTestBadge
  } from '../../wailsjs/go/devtest/DevTest.js'
  import CustomSelect from './CustomSelect.svelte'

  let selectedAction = $state('native')
  let customAppName = $state('')
  let customIconUrl = $state('')
  let conflicts = $state<string[]>([])

  const actionLabels: Record<string, string> = {
    native: 'Native Toast',
    notitle: 'No Title',
    longmsg: 'Long Message',
    badge: 'Badge Notif'
  }

  const allIcons = [...gameIcons, ...appIcons]
  const allIconsOptions = [
    { value: '', label: '-- No Mock App --' },
    ...allIcons.map((brand) => ({
      value: brand.matchers[0],
      label: brand.name
    }))
  ]

  function checkConflicts() {
    const found = findAppIconConflicts()
    if (found.length === 0) {
      conflicts = ['✅ No shadowing conflicts found!']
    } else {
      conflicts = found
    }
  }

  function triggerAction() {
    const mockApp = customAppName.trim()

    if (mockApp !== '') {
      // Direct-emit path: simulates what the WinRT backend would emit.
      // Useful for testing appIcons.ts matchers without firing a real toast.
      // The app name here should mimic the real OS display name (e.g. "Discord").
      let finalName = mockApp
      const matchString = mockApp.toLowerCase()
      for (const brand of allIcons) {
        if (brand.matchers.some((m) => matchString.includes(m))) {
          finalName = brand.name
          break
        }
      }

      // Emit the event so App.svelte catches it natively (triggers sound + showWidget)
      ;(
        window as Window &
          typeof globalThis & {
            runtime: {
              EventsEmit: (
                eventName: string,
                payload: Record<string, string>
              ) => void
            }
          }
      ).runtime.EventsEmit('notification', {
        appName: mockApp,
        title: 'Mock App Notification',
        message: `Resolved to: ${finalName}`,
        icon: customIconUrl.trim(),
        heroImage: ''
      })
      return
    }

    // No custom name — fire a real Windows toast through the backend.
    // The WinRT listener will intercept it and emit it through the normal pipeline.
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

  <CustomSelect options={allIconsOptions} bind:value={customAppName} />

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
  <button class="action-btn" onclick={checkConflicts}>Find Conflicts</button>
</div>

{#if conflicts.length > 0}
  <div class="conflicts-panel">
    <button class="close-btn" onclick={() => (conflicts = [])}>✖</button>
    {#each conflicts as conflict (conflict)}
      <div class="conflict-item">{conflict}</div>
    {/each}
  </div>
{/if}

<style>
  .dev-controls {
    display: flex;
    flex-wrap: wrap; /* Allows controls to flow onto next line instead of stretching horizontal */
    gap: 8px;
    align-items: center;
    max-width: 250px; /* Constrain width so it forces vertical wrapping */
    justify-content: center;
  }

  .conflicts-panel {
    margin-top: 10px;
    padding: 8px 24px 8px 8px;
    background: rgba(255, 60, 60, 0.15);
    border: 1px solid rgba(255, 60, 60, 0.3);
    border-radius: 6px;
    font-size: 0.75rem;
    color: var(--c-text-primary);
    max-height: 150px;
    overflow-y: auto;
    width: 100%;
    box-sizing: border-box;
    position: relative;
  }

  .conflict-item {
    margin-bottom: 6px;
    word-break: break-all;
    line-height: 1.2;
  }

  .conflict-item:last-child {
    margin-bottom: 0;
  }

  .close-btn {
    position: absolute;
    top: 4px;
    right: 4px;
    background: transparent;
    border: none;
    color: var(--c-text-secondary);
    cursor: pointer;
    font-size: 0.8rem;
  }

  .close-btn:hover {
    color: var(--c-text-primary);
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

  /* Force the native select to have a higher stacking context than sibling buttons */
  select.action-select {
    position: relative;
    z-index: 50;
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
