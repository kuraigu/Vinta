<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 }, // Apps page is never minimal
    expanded: { w: 340, h: 260, r: 24 }, // Slightly taller for the Dock
    borderColor: 'var(--c-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { getContext } from 'svelte'
  import { Clock, Bell, Info, RefreshCw } from 'lucide-svelte'
  import ClockPage, { config as clockConfig } from './ClockPage.svelte'
  import NotificationPage, {
    config as notifConfig
  } from './NotificationPage.svelte'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  function openApp(page: unknown, config: IslandConfig) {
    navigate(page, config, true)
  }
</script>

<div class="apps-page">
  <div class="header">
    <div class="title">Apps</div>
  </div>

  <div class="grid">
    <button
      class="app-btn"
      onclick={(e) => {
        e.stopPropagation()
        openApp(ClockPage, clockConfig)
      }}
    >
      <div class="icon clock">
        <Clock size={24} />
      </div>
      <span class="label">Clock</span>
    </button>

    <button
      class="app-btn"
      onclick={(e) => {
        e.stopPropagation()
        openApp(NotificationPage, notifConfig)
      }}
    >
      <div class="icon notif">
        <Bell size={24} />
      </div>
      <span class="label">Alerts</span>
    </button>

    <button class="app-btn placeholder" onclick={(e) => e.stopPropagation()}>
      <div class="icon about">
        <Info size={24} />
      </div>
      <span class="label">About</span>
    </button>

    <button class="app-btn placeholder" onclick={(e) => e.stopPropagation()}>
      <div class="icon updates">
        <RefreshCw size={24} />
      </div>
      <span class="label">Updates</span>
    </button>
  </div>
</div>

<style>
  .apps-page {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 20px 20px 60px 20px; /* Bottom padding for dock */
    box-sizing: border-box;
    color: var(--c-text-primary);
  }

  .header {
    display: flex;
    justify-content: center;
    align-items: center;
    margin-bottom: 24px;
  }

  .title {
    font-size: 16px;
    font-weight: 600;
    letter-spacing: -0.01em;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
    padding: 0 10px;
  }

  .app-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-start;
    gap: 8px;
    background: transparent;
    border: none;
    cursor: pointer;
    padding: 0;
    transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .app-btn:hover {
    transform: scale(1.05);
  }

  .app-btn:active {
    transform: scale(0.95);
  }

  .icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .icon.clock {
    background: linear-gradient(135deg, #ff6b6b, #ff8e53);
  }
  .icon.notif {
    background: linear-gradient(135deg, #4e65ff, #92effd);
    color: #1a1a1a;
  }
  .icon.about {
    background: linear-gradient(135deg, #667eea, #764ba2);
  }
  .icon.updates {
    background: linear-gradient(135deg, #00c9ff, #92fe9d);
    color: #1a1a1a;
  }

  .label {
    font-size: 11px;
    font-weight: 500;
    color: var(--c-text-secondary);
    text-align: center;
  }

  .placeholder {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .placeholder:hover {
    transform: none;
  }
</style>
