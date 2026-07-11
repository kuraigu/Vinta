<script context="module" lang="ts">
  // You can import these directly in your JS/TS logic if you ever need colors in code:
  // import { colors } from './Colors.svelte'
  export const colors = {
    primary: '#007aff', // Blue (e.g. notifications)
    success: '#34c759', // Green (e.g. active states)
    danger: '#ff3b30', // Red (e.g. errors, low battery)
    warning: '#ffcc00', // Yellow/Orange

    bg: 'rgba(20, 20, 20, 0.95)',
    border: 'rgba(255, 255, 255, 0.1)',
    hover: 'rgba(255, 255, 255, 0.08)',
    itemBg: 'rgba(255, 255, 255, 0.05)',

    textPrimary: '#ffffff',
    textSecondary: '#b0b0b0',
    textMuted: '#888888'
  }
</script>

<!-- 
  By adding the 'global' attribute, these CSS variables become available to EVERY 
  component in your app without needing to import this file!
  
  Usage in any style tag:
  color: var(--c-text-primary);
  background: var(--c-bg);
-->
<style>
  :global(:root) {
    /* DEFAULT DARK THEME */
    --c-primary: #007aff;
    --c-success: #34c759;
    --c-danger: #ff3b30;
    --c-warning: #ffcc00;

    --c-bg: rgba(20, 20, 20, 0.95);
    --c-border: rgba(255, 255, 255, 0.1);
    --c-hover: rgba(255, 255, 255, 0.08);
    --c-item-bg: rgba(255, 255, 255, 0.05);

    --c-text-primary: #ffffff;
    --c-text-secondary: #b0b0b0;
    --c-text-muted: #888888;
  }

  :global([data-theme='light']) {
    /* LIGHT THEME OVERRIDES */
    --c-bg: rgba(250, 250, 250, 0.9);
    --c-border: rgba(0, 0, 0, 0.1);
    --c-hover: rgba(0, 0, 0, 0.05);
    --c-item-bg: rgba(0, 0, 0, 0.03);

    --c-text-primary: #111111;
    --c-text-secondary: #666666;
    --c-text-muted: #999999;
  }
</style>
