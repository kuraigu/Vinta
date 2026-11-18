<script lang="ts">
  /**
   * VintaMascot
   *
   * Reusable vector mascot component representing the official Vinta robot companion.
   * Supports dedicated expressive poses for onboarding, guides, and status states.
   */
  export type MascotPose =
    'welcome' | 'position' | 'theme' | 'settings' | 'celebrate'

  let {
    pose = 'welcome',
    class: className = ''
  }: {
    pose?: MascotPose
    class?: string
  } = $props()
  const uid = `vm_${Math.random().toString(36).slice(2, 8)}`
</script>

<div class="vinta-mascot-root {className}">
  <svg class="mascot-svg" viewBox="24 30 72 62" aria-label="Vinta mascot">
    <defs>
      <g id="{uid}-chassis">
        <!-- Side Comms Pods (Exact 1:1 Match with appicon.svg) -->
        <rect x="29" y="54.5" width="5" height="15.2" rx="2.5" fill="#F59E0B" />
        <rect x="86" y="54.5" width="5" height="15.2" rx="2.5" fill="#F59E0B" />

        <!-- Dark-Gold Rounded Bot Helmet -->
        <rect
          x="35.6"
          y="42"
          width="48.8"
          height="40.9"
          rx="18.5"
          fill="#F59E0B"
        />

        <!-- Dynamic-Island Top Notch Cutout -->
        <rect x="53.4" y="41" width="13.2" height="5" rx="2.5" fill="#111925" />

        <!-- Obsidian Visor -->
        <rect
          x="40.5"
          y="50"
          width="39"
          height="27.7"
          rx="11.2"
          fill="#111925"
        />
      </g>

      <filter id="{uid}-shadow" x="-20%" y="-20%" width="140%" height="140%">
        <feDropShadow
          dx="0"
          dy="4"
          stdDeviation="4"
          flood-color="#0f172a"
          flood-opacity="0.35"
        />
      </filter>
    </defs>

    <g class="anim-float-calm" filter="url(#{uid}-shadow)">
      <use href="#{uid}-chassis" />

      {#if pose === 'welcome'}
        <g class="anim-look-around" fill="#F59E0B">
          <path
            d="M 48.8 57.5 L 53.4 61 L 48.8 64.5"
            fill="none"
            stroke="#F59E0B"
            stroke-width="3"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          <rect x="65.6" y="56" width="6.3" height="9.9" rx="3.15" />
          <path
            d="M 54.4 70.6 Q 60 73.9 65.6 70.6"
            fill="none"
            stroke="#F59E0B"
            stroke-width="2.6"
            stroke-linecap="round"
          />
        </g>
      {:else if pose === 'position'}
        <g fill="#F59E0B">
          <path
            d="M 51.5 57.5 L 56.1 61 L 51.5 64.5"
            fill="none"
            stroke="#F59E0B"
            stroke-width="3"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          <rect x="68" y="56" width="6.3" height="9.9" rx="3.15" />
          <path
            d="M 56.5 70.6 Q 62 73.6 67.5 70.6"
            fill="none"
            stroke="#F59E0B"
            stroke-width="2.6"
            stroke-linecap="round"
          />
        </g>
      {:else if pose === 'theme'}
        <g fill="#F59E0B">
          <path
            d="M 48.8 57.5 L 53.4 61 L 48.8 64.5"
            fill="none"
            stroke="#F59E0B"
            stroke-width="3"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          <path
            d="M 64.5 62 Q 69 57 73.5 62"
            fill="none"
            stroke="#F59E0B"
            stroke-width="2.8"
            stroke-linecap="round"
          />
          <path
            d="M 54.4 70.6 Q 60 73.9 65.6 70.6"
            fill="none"
            stroke="#F59E0B"
            stroke-width="2.6"
            stroke-linecap="round"
          />
        </g>
      {:else if pose === 'settings'}
        <g fill="#F59E0B">
          <path
            d="M 48.8 57.5 L 53.4 61 L 48.8 64.5"
            fill="none"
            stroke="#F59E0B"
            stroke-width="3"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          <rect x="65.6" y="56.5" width="6.3" height="9" rx="3.15" />
          <path
            d="M 55 70.6 Q 60 73 65 70.6"
            fill="none"
            stroke="#F59E0B"
            stroke-width="2.6"
            stroke-linecap="round"
          />
        </g>
      {:else if pose === 'celebrate'}
        <g>
          <path
            d="M 47 62.5 Q 51.5 56.5 56 62.5"
            fill="none"
            stroke="#F59E0B"
            stroke-width="2.8"
            stroke-linecap="round"
          />
          <path
            d="M 64 62.5 Q 68.5 56.5 73 62.5"
            fill="none"
            stroke="#F59E0B"
            stroke-width="2.8"
            stroke-linecap="round"
          />
          <path
            d="M 53.5 70 Q 60 74.5 66.5 70"
            fill="none"
            stroke="#F59E0B"
            stroke-width="2.8"
            stroke-linecap="round"
          />
        </g>
      {/if}
    </g>
  </svg>
</div>

<style>
  .vinta-mascot-root {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    pointer-events: none;
    user-select: none;
  }

  .mascot-svg {
    width: 100%;
    height: 100%;
    overflow: visible;
    object-fit: contain;
  }

  .anim-float-calm {
    animation: floatCalm 4s ease-in-out infinite;
  }

  .anim-look-around {
    animation: lookAround 6s ease-in-out infinite;
  }

  @keyframes floatCalm {
    0%,
    100% {
      transform: translateY(0px);
    }
    50% {
      transform: translateY(-3.5px);
    }
  }

  @keyframes lookAround {
    0%,
    100% {
      transform: translate(0px, 0px);
    }
    30% {
      transform: translate(2px, -1.2px);
    }
    70% {
      transform: translate(-1.8px, -0.8px);
    }
  }
</style>
