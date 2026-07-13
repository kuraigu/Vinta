<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 }, // Settings is never meant to be 'minimal'
    expanded: { w: 340, h: 260, r: 24 }, // Slightly taller for the Dock
    borderColor: 'var(--c-border)',
    glowColor: 'transparent',
  }
</script>

<script lang="ts">
  import {
    themeStore,
    shapeStore,
    autoHideStore,
    wakeMethodStore,
    type AppTheme,
    type IslandShape
  } from '../store'
  import { ShieldCheck, Moon, Type, EyeOff, Pointer } from 'lucide-svelte'

  export let expanded = true
</script>

<div class="settings-page">
  <div class="header">
    <div class="title">Settings</div>
  </div>

  <div class="content-scroll">
    <div class="section">
      <div class="section-title">Appearance</div>
      
      <div class="setting-item">
        <span>Theme</span>
        <div class="segment-control">
          <button class:active={$themeStore === 'dark'} on:click={() => $themeStore = 'dark'}>Dark</button>
          <button class:active={$themeStore === 'light'} on:click={() => $themeStore = 'light'}>Light</button>
        </div>
      </div>

      <div class="setting-item">
        <span>Shape</span>
        <div class="segment-control">
          <button class:active={$shapeStore === 'pill'} on:click={() => $shapeStore = 'pill'}>Pill</button>
          <button class:active={$shapeStore === 'box'} on:click={() => $shapeStore = 'box'}>Box</button>
        </div>
      </div>
    </div>

    <div class="section">
      <div class="section-title">Behavior</div>
      
      <div class="setting-item">
        <span>Auto-hide Pill</span>
        <label class="toggle">
          <input type="checkbox" bind:checked={$autoHideStore} />
          <span class="slider"></span>
        </label>
      </div>

      <div class="setting-item">
        <span>Wake Method</span>
        <div class="segment-control">
          <button class:active={$wakeMethodStore === 'hover'} on:click={() => $wakeMethodStore = 'hover'}>Hover</button>
          <button class:active={$wakeMethodStore === 'click'} on:click={() => $wakeMethodStore = 'click'}>Click</button>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .settings-page {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 16px;
    padding-bottom: 56px; /* Space for Universal Dock */
    box-sizing: border-box;
    color: var(--c-text-primary);
  }

  .header {
    display: flex;
    align-items: center;
    margin-bottom: 16px;
    justify-content: center;
  }

  .title {
    font-weight: 600;
    font-size: 14px;
    text-align: center;
  }

  .content-scroll {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .content-scroll::-webkit-scrollbar {
    width: 4px;
  }
  .content-scroll::-webkit-scrollbar-thumb {
    background: var(--c-border);
    border-radius: 4px;
  }

  .section {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .section-title {
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--c-text-muted);
    font-weight: 600;
  }

  .setting-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: var(--c-item-bg);
    padding: 10px 12px;
    border-radius: 8px;
    font-size: 13px;
    border: 1px solid var(--c-border);
  }

  .segment-control {
    display: flex;
    background: rgba(0, 0, 0, 0.2);
    border-radius: 6px;
    padding: 2px;
    gap: 2px;
  }

  :global([data-theme='light']) .segment-control {
    background: rgba(0, 0, 0, 0.05);
  }

  .segment-control button {
    background: transparent;
    border: none;
    color: var(--c-text-secondary);
    padding: 4px 12px;
    font-size: 11px;
    font-weight: 600;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .segment-control button:hover {
    color: var(--c-text-primary);
  }

  .segment-control button.active {
    background: var(--c-border);
    color: var(--c-text-primary);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .theme-select {
    background: var(--c-bg);
    color: var(--c-text-primary);
    border: 1px solid var(--c-border);
    border-radius: 8px;
    padding: 6px 12px;
    font-size: 14px;
    font-family: inherit;
    outline: none;
    cursor: pointer;
    min-width: 120px;
  }
  .theme-select:focus {
    border-color: var(--c-accent, #667EEA);
  }

  /* Toggle Switch CSS */
  .toggle {
    position: relative;
    display: inline-block;
    width: 44px;
    height: 24px;
  }
  .toggle input {
    opacity: 0;
    width: 0;
    height: 0;
  }
  .slider {
    position: absolute;
    cursor: pointer;
    top: 0; left: 0; right: 0; bottom: 0;
    background-color: var(--c-border);
    transition: .4s;
    border-radius: 24px;
  }
  .slider:before {
    position: absolute;
    content: "";
    height: 18px;
    width: 18px;
    left: 3px;
    bottom: 3px;
    background-color: white;
    transition: .4s;
    border-radius: 50%;
  }
  input:checked + .slider {
    background-color: var(--c-accent, #667EEA);
  }
  input:checked + .slider:before {
    transform: translateX(20px);
  }

  .about-info {
    background: var(--c-item-bg);
    padding: 10px 12px;
    border-radius: 8px;
    font-size: 13px;
    border: 1px solid var(--c-border);
    text-align: center;
    line-height: 1.4;
  }
  .about-info small {
    color: var(--c-text-muted);
    font-size: 11px;
  }
</style>
