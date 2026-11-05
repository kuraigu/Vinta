<script lang="ts">
  /**
   * VintaMascot
   *
   * Reusable vector mascot component representing the official Vinta robot companion.
   * Supports dedicated expressive poses for onboarding, guides, and status states.
   */
  export type MascotPose =
    | 'welcome'
    | 'position'
    | 'theme'
    | 'settings'
    | 'celebrate'

  let {
    pose = 'welcome',
    class: className = ''
  }: {
    pose?: MascotPose
    class?: string
  } = $props()
</script>

<div class="vinta-mascot-root {className}">
  <svg class="mascot-svg" viewBox="24 30 72 62" aria-label="Vinta mascot">
    <defs>
      <clipPath id="vintaMascotHeadClip">
        <circle cx="60" cy="62" r="27" />
      </clipPath>

      <clipPath id="vintaMascotEarsClip">
        <rect x="28" y="51" width="6" height="20" rx="3" />
        <rect x="86" y="51" width="6" height="20" rx="3" />
      </clipPath>

      <g id="vintaMascotChassis">
        <g clip-path="url(#vintaMascotEarsClip)">
          <rect x="27" y="50" width="66" height="11" fill="#68788e" />
          <rect x="27" y="61" width="66" height="12" fill="#475569" />
        </g>

        <g clip-path="url(#vintaMascotHeadClip)">
          <rect x="30" y="32" width="60" height="30" fill="#94a3b8" />
          <rect x="30" y="62" width="60" height="30" fill="#475569" />
          <polygon points="36,41 40,37 51,48 48,48" fill="#182232" />
          <polygon points="84,41 80,37 69,48 72,48" fill="#182232" />
          <polygon points="40,37 46,33 55,48 51,48" fill="#3e4c5e" />
          <polygon points="80,37 74,33 65,48 69,48" fill="#3e4c5e" />
          <polygon points="60,33 69,41 64,48 56,48 51,41" fill="#2e3b4e" />
        </g>

        <rect
          x="34.5"
          y="46.5"
          width="51"
          height="34"
          rx="12.5"
          fill="#2b3748"
        />
        <rect x="37" y="49" width="46" height="29" rx="10.5" fill="#09101d" />
      </g>

      <filter id="vintaMascotEyeGlow" x="-40%" y="-40%" width="180%" height="180%">
        <feDropShadow
          dx="0"
          dy="0"
          stdDeviation="2.2"
          flood-color="#facc15"
          flood-opacity="0.85"
        />
      </filter>

      <filter id="vintaMascotShadow" x="-20%" y="-20%" width="140%" height="140%">
        <feDropShadow
          dx="0"
          dy="4"
          stdDeviation="4"
          flood-color="#0f172a"
          flood-opacity="0.35"
        />
      </filter>
    </defs>

    <g class="anim-float-calm" filter="url(#vintaMascotShadow)">
      <use href="#vintaMascotChassis" />

      {#if pose === 'welcome'}
        <g class="anim-look-around" filter="url(#vintaMascotEyeGlow)" fill="#facc15">
          <rect x="43" y="53" width="13" height="13" rx="5.5" />
          <rect x="64" y="53" width="13" height="13" rx="5.5" />
          <path d="M 52 68.5 Q 60 70.5 68 68.5 C 67 75.5 53 75.5 52 68.5 Z" />
        </g>
      {:else if pose === 'position'}
        <g filter="url(#vintaMascotEyeGlow)" fill="#facc15">
          <rect x="46" y="53" width="13" height="13" rx="5.5" />
          <rect x="67" y="53" width="13" height="13" rx="5.5" />
          <path d="M 53 69 Q 60 70.5 67 69 C 66 74 54 74 53 69 Z" />
        </g>
      {:else if pose === 'theme'}
        <g filter="url(#vintaMascotEyeGlow)">
          <rect x="43" y="53" width="13" height="13" rx="5.5" fill="#facc15" />
          <path
            d="M 64 60 Q 70 54 76 60"
            fill="none"
            stroke="#facc15"
            stroke-width="2.6"
            stroke-linecap="round"
          />
          <path
            d="M 52 68.5 Q 60 70.5 68 68.5 C 67 75.5 53 75.5 52 68.5 Z"
            fill="#facc15"
          />
        </g>
      {:else if pose === 'settings'}
        <g filter="url(#vintaMascotEyeGlow)" fill="#facc15">
          <rect x="44" y="54" width="12" height="12" rx="5" />
          <rect x="64" y="54" width="12" height="12" rx="5" />
          <path d="M 53 69.5 Q 60 71 67 69.5 C 66 73 54 73 53 69.5 Z" />
        </g>
      {:else if pose === 'celebrate'}
        <g filter="url(#vintaMascotEyeGlow)">
          <path
            d="M 43 60 Q 49.5 52 56 60"
            fill="none"
            stroke="#facc15"
            stroke-width="2.8"
            stroke-linecap="round"
          />
          <path
            d="M 64 60 Q 70.5 52 77 60"
            fill="none"
            stroke="#facc15"
            stroke-width="2.8"
            stroke-linecap="round"
          />
          <path
            d="M 51 68 Q 60 71.5 69 68 C 68 76 52 76 51 68 Z"
            fill="#facc15"
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
