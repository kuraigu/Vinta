<script lang="ts">
  import { Globe } from 'lucide-svelte'

  interface Props {
    code: string
    size?: number
  }

  let { code = 'en', size = 16 }: Props = $props()

  let normalized = $derived(code.toLowerCase().trim())
</script>

<span
  class="flag-icon-wrap"
  style="--flag-w: {Math.round((size * 4) / 3)}px; --flag-h: {size}px;"
>
  {#if normalized === 'en' || normalized === 'en-us'}
    <!-- United States Flag (Vexillologically accurate: 13 stripes, 50 stars in 9 alternating rows of 6 & 5) -->
    <svg viewBox="0 0 1235 650" class="flag-svg" aria-label="English (US)">
      <defs>
        <!-- Standard 5-pointed star -->
        <polygon
          id="us-star"
          points="0,-14 3.14,-4.33 13.31,-4.33 5.09,1.65 8.23,11.33 0,5.35 -8.23,11.33 -5.09,1.65 -13.31,-4.33 -3.14,-4.33"
          fill="#fff"
        />
        <!-- Row of 6 stars -->
        <g id="us-row6">
          <use href="#us-star" x="41.17" />
          <use href="#us-star" x="123.5" />
          <use href="#us-star" x="205.83" />
          <use href="#us-star" x="288.17" />
          <use href="#us-star" x="370.5" />
          <use href="#us-star" x="452.83" />
        </g>
        <!-- Row of 5 stars -->
        <g id="us-row5">
          <use href="#us-star" x="82.33" />
          <use href="#us-star" x="164.67" />
          <use href="#us-star" x="247" />
          <use href="#us-star" x="329.33" />
          <use href="#us-star" x="411.67" />
        </g>
      </defs>
      <!-- 13 Stripes: Red background with 6 white stripes -->
      <rect width="1235" height="650" fill="#b22234" />
      <path
        d="M0,50h1235v50H0z M0,150h1235v50H0z M0,250h1235v50H0z M0,350h1235v50H0z M0,450h1235v50H0z M0,550h1235v50H0z"
        fill="#fff"
      />
      <!-- Canton: Blue field covering 7 stripes -->
      <rect width="494" height="350" fill="#3c3b6e" />
      <!-- 50 Stars: 9 rows alternating 6 and 5 stars -->
      <g>
        <use href="#us-row6" y="35" />
        <use href="#us-row5" y="70" />
        <use href="#us-row6" y="105" />
        <use href="#us-row5" y="140" />
        <use href="#us-row6" y="175" />
        <use href="#us-row5" y="210" />
        <use href="#us-row6" y="245" />
        <use href="#us-row5" y="280" />
        <use href="#us-row6" y="315" />
      </g>
    </svg>
  {:else if normalized === 'en-gb'}
    <!-- United Kingdom Flag (Official Union Jack geometry with counterchanged St. Patrick saltire) -->
    <svg viewBox="0 0 60 30" class="flag-svg" aria-label="English (UK)">
      <clipPath id="uk-clip"><path d="M0,0 v30 h60 v-30 z" /></clipPath>
      <clipPath id="uk-counter"><path d="M30,15 h30 v15 z v15 h-30 z h-30 v-15 z v-15 h30 z" /></clipPath>
      <g clip-path="url(#uk-clip)">
        <path d="M0,0 v30 h60 v-30 z" fill="#012169" />
        <path d="M0,0 L60,30 M60,0 L0,30" stroke="#fff" stroke-width="6" />
        <path d="M0,0 L60,30 M60,0 L0,30" clip-path="url(#uk-counter)" stroke="#C8102E" stroke-width="4" />
        <path d="M30,0 v30 M0,15 h60" stroke="#fff" stroke-width="10" />
        <path d="M30,0 v30 M0,15 h60" stroke="#C8102E" stroke-width="6" />
      </g>
    </svg>
  {:else if normalized === 'de'}
    <!-- Germany Flag -->
    <svg viewBox="0 0 5 3" class="flag-svg" aria-label="Deutsch">
      <rect width="5" height="1" y="0" fill="#000" />
      <rect width="5" height="1" y="1" fill="#dd0000" />
      <rect width="5" height="1" y="2" fill="#ffce00" />
    </svg>
  {:else if normalized === 'fr'}
    <!-- France Flag -->
    <svg viewBox="0 0 3 2" class="flag-svg" aria-label="Français">
      <rect width="1" height="2" x="0" fill="#002395" />
      <rect width="1" height="2" x="1" fill="#fff" />
      <rect width="1" height="2" x="2" fill="#ed2939" />
    </svg>
  {:else if normalized === 'es'}
    <!-- Spain Flag -->
    <svg viewBox="0 0 3 2" class="flag-svg" aria-label="Español">
      <rect width="3" height="0.5" y="0" fill="#aa151b" />
      <rect width="3" height="1" y="0.5" fill="#f1bf00" />
      <rect width="3" height="0.5" y="1.5" fill="#aa151b" />
    </svg>
  {:else if normalized === 'it'}
    <!-- Italy Flag -->
    <svg viewBox="0 0 3 2" class="flag-svg" aria-label="Italiano">
      <rect width="1" height="2" x="0" fill="#009246" />
      <rect width="1" height="2" x="1" fill="#fff" />
      <rect width="1" height="2" x="2" fill="#ce2b37" />
    </svg>
  {:else if normalized === 'ja'}
    <!-- Japan Flag -->
    <svg viewBox="0 0 3 2" class="flag-svg" aria-label="日本語">
      <rect width="3" height="2" fill="#fff" />
      <circle cx="1.5" cy="1" r="0.6" fill="#bc002d" />
    </svg>
  {:else if normalized === 'zh'}
    <!-- China Flag (5 stars: 1 large and 4 small arced stars pointing to center) -->
    <svg viewBox="0 0 30 20" class="flag-svg" aria-label="中文">
      <defs>
        <polygon
          id="zh-star"
          points="0,-1 0.225,-0.309 0.951,-0.309 0.363,0.118 0.588,0.809 0,0.382 -0.588,0.809 -0.363,0.118 -0.951,-0.309 -0.225,-0.309"
          fill="#ffde00"
        />
      </defs>
      <rect width="30" height="20" fill="#de2910" />
      <!-- Main big star at (5, 5), radius 3 -->
      <use href="#zh-star" transform="translate(5, 5) scale(3)" />
      <!-- 4 smaller stars pointing toward (5, 5) -->
      <use href="#zh-star" transform="translate(10, 2) rotate(108.43)" />
      <use href="#zh-star" transform="translate(12, 4) rotate(126.87)" />
      <use href="#zh-star" transform="translate(12, 7) rotate(161.57)" />
      <use href="#zh-star" transform="translate(10, 9) rotate(198.43)" />
    </svg>
  {:else if normalized === 'ru'}
    <!-- Russia Flag -->
    <svg viewBox="0 0 3 2" class="flag-svg" aria-label="Русский">
      <rect width="3" height="0.66" y="0" fill="#fff" />
      <rect width="3" height="0.66" y="0.66" fill="#0039a6" />
      <rect width="3" height="0.68" y="1.32" fill="#d52b1e" />
    </svg>
  {:else if normalized === 'uk'}
    <!-- Ukraine Flag -->
    <svg viewBox="0 0 3 2" class="flag-svg" aria-label="Українська">
      <rect width="3" height="1" y="0" fill="#0057b7" />
      <rect width="3" height="1" y="1" fill="#ffd700" />
    </svg>
  {:else if normalized === 'pl'}
    <!-- Poland Flag -->
    <svg viewBox="0 0 3 2" class="flag-svg" aria-label="Polski">
      <rect width="3" height="1" y="0" fill="#fff" />
      <rect width="3" height="1" y="1" fill="#dc143c" />
    </svg>
  {:else if normalized === 'nl'}
    <!-- Netherlands Flag -->
    <svg viewBox="0 0 3 2" class="flag-svg" aria-label="Nederlands">
      <rect width="3" height="0.66" y="0" fill="#ae1c28" />
      <rect width="3" height="0.66" y="0.66" fill="#fff" />
      <rect width="3" height="0.68" y="1.32" fill="#21468b" />
    </svg>
  {:else if normalized === 'fil' || normalized === 'tl' || normalized === 'ph'}
    <!-- Philippines Flag -->
    <svg viewBox="0 0 60 30" class="flag-svg" aria-label="Filipino">
      <defs>
        <!-- 5-pointed star -->
        <polygon
          id="ph-star"
          points="0,-1.1 0.32,-0.34 1.05,-0.34 0.46,0.1 0.68,0.85 0,0.39 -0.68,0.85 -0.46,0.1 -1.05,-0.34 -0.32,-0.34"
          fill="#fcd116"
        />
        <!-- 8-ray sun assembly: 4 orthogonal rays + 4 diagonal rays with center disk -->
        <g id="ph-sun">
          <circle cx="0" cy="0" r="2.8" fill="#fcd116" />
          <path
            d="M0,-5.2 L0.6,-3 L-0.6,-3 Z M0,5.2 L0.6,3 L-0.6,3 Z M-5.2,0 L-3,0.6 L-3,-0.6 Z M5.2,0 L3,0.6 L3,-0.6 Z"
            fill="#fcd116"
          />
          <path
            d="M-3.68,-3.68 L-1.7,-2.55 L-2.55,-1.7 Z M3.68,3.68 L1.7,2.55 L2.55,1.7 Z M-3.68,3.68 L-2.55,1.7 L-1.7,2.55 Z M3.68,-3.68 L2.55,-1.7 L1.7,-2.55 Z"
            fill="#fcd116"
          />
        </g>
      </defs>
      <!-- Royal Blue top stripe -->
      <rect width="60" height="15" y="0" fill="#0038a8" />
      <!-- Crimson Red bottom stripe -->
      <rect width="60" height="15" y="15" fill="#ce1126" />
      <!-- White equilateral hoist triangle -->
      <polygon points="0,0 25.98,15 0,30" fill="#fff" />
      <!-- Golden Sun centered at (9.6, 15) -->
      <use href="#ph-sun" transform="translate(9.6, 15)" />
      <!-- 3 Golden Stars at triangle corners -->
      <use href="#ph-star" transform="translate(3.2, 5.2) rotate(15)" />
      <use href="#ph-star" transform="translate(3.2, 24.8) rotate(-15)" />
      <use href="#ph-star" transform="translate(22.2, 15) rotate(90)" />
    </svg>
  {:else}
    <!-- Custom community translation fallback -->
    <div class="custom-locale-badge">
      <Globe size={Math.max(12, size - 2)} />
    </div>
  {/if}
</span>

<style>
  .flag-icon-wrap {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: var(--flag-w);
    height: var(--flag-h);
    border-radius: 2px;
    overflow: hidden;
    box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.15);
    flex-shrink: 0;
    vertical-align: middle;
  }

  .flag-svg {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .custom-locale-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    background: rgba(255, 255, 255, 0.08);
    color: var(--vinta-text-secondary);
  }
</style>
