<script lang="ts">
  interface Props {
    color?: string
    size?: string
  }

  let {
    color = 'var(--vinta-success)', // default to a pleasant green
    size = '8px'
  }: Props = $props()
</script>

<div class="dot-container" style="--dot-color: {color}; --dot-size: {size};">
  <div class="pulse"></div>
  <div class="dot"></div>
</div>

<style>
  .dot-container {
    position: relative;
    width: var(--dot-size);
    height: var(--dot-size);
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }

  .dot {
    position: absolute;
    width: 100%;
    height: 100%;
    background-color: var(--dot-color);
    border-radius: 50%;
    z-index: 2;
  }

  .pulse {
    position: absolute;
    width: 100%;
    height: 100%;
    background-color: var(--dot-color);
    border-radius: 50%;
    z-index: 1;
    animation: pulse-animation 2s infinite cubic-bezier(0.2, 0, 0, 1);
  }

  @keyframes pulse-animation {
    0% {
      transform: scale(1);
      opacity: 0.8;
    }
    100% {
      transform: scale(2.5);
      opacity: 0;
    }
  }
</style>
