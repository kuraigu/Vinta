<script lang="ts">
  /**
   * WeatherArt
   *
   * Reusable vector weather illustration supporting rich hero artwork
   * and compact glyph rendering across all weather conditions.
   */
  export type WeatherConditionType =
    | 'sunny'
    | 'night'
    | 'cloudy'
    | 'cloudy-night'
    | 'rain'
    | 'thunder'
    | 'snow'
    | 'fog'

  let {
    type = 'sunny',
    variant = 'hero',
    class: className = ''
  }: {
    type?: WeatherConditionType
    variant?: 'hero' | 'compact'
    class?: string
  } = $props()
</script>

{#if variant === 'hero'}
  <!-- =========================================================================
       HERO VOLUMETRIC ARTWORK (Banner Cards and Full Weather Page)
       ========================================================================= -->
  <div class="weather-art-hero {className}">
    {#if type === 'sunny'}
      <!-- Radiant Sun with Soft, Warm Golden Palette -->
      <svg class="hero-svg" viewBox="0 0 120 100" aria-label="Sunny weather">
        <defs>
          <radialGradient id="sunAura" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stop-color="#ffffff" stop-opacity="0.3" />
            <stop offset="50%" stop-color="#fef08a" stop-opacity="0.12" />
            <stop offset="100%" stop-color="#fef08a" stop-opacity="0" />
          </radialGradient>
          <linearGradient id="sunBodyGrad" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stop-color="#ffffff" />
            <stop offset="40%" stop-color="#fef9c3" />
            <stop offset="100%" stop-color="#fde047" />
          </linearGradient>
        </defs>

        <!-- Ambient Soft Atmosphere Halo -->
        <circle cx="60" cy="50" r="44" fill="url(#sunAura)" />

        <!-- Soft Corona Rays with Rounded Ends -->
        <g
          stroke="#fef08a"
          stroke-width="2.2"
          stroke-linecap="round"
          opacity="0.75"
        >
          <line x1="60" y1="17" x2="60" y2="24" />
          <line x1="60" y1="76" x2="60" y2="83" />
          <line x1="27" y1="50" x2="34" y2="50" />
          <line x1="86" y1="50" x2="93" y2="50" />
          <line x1="37" y1="27" x2="42" y2="32" />
          <line x1="78" y1="68" x2="83" y2="73" />
          <line x1="37" y1="73" x2="42" y2="68" />
          <line x1="78" y1="32" x2="83" y2="27" />
        </g>

        <!-- Luminous Warm Sun Orb -->
        <circle
          cx="60"
          cy="50"
          r="20"
          fill="url(#sunBodyGrad)"
          filter="drop-shadow(0 4px 14px rgba(254, 240, 138, 0.35))"
        />
      </svg>
    {:else if type === 'night'}
      <!-- Luminous Full Celestial Moon with Crater Maria & Sparkling Constellations -->
      <svg class="hero-svg" viewBox="0 0 120 100" aria-label="Clear night sky">
        <defs>
          <radialGradient id="moonAura" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stop-color="#ffffff" stop-opacity="0.38" />
            <stop offset="45%" stop-color="#c7d2fe" stop-opacity="0.22" />
            <stop offset="75%" stop-color="#818cf8" stop-opacity="0.08" />
            <stop offset="100%" stop-color="#4f46e5" stop-opacity="0" />
          </radialGradient>
          <linearGradient id="moonBodyGrad" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stop-color="#ffffff" />
            <stop offset="35%" stop-color="#f8fafc" />
            <stop offset="70%" stop-color="#e2e8f0" />
            <stop offset="100%" stop-color="#cbd5e1" />
          </linearGradient>
          <radialGradient id="moonLimbShadow" cx="35%" cy="35%" r="65%">
            <stop offset="60%" stop-color="#ffffff" stop-opacity="0" />
            <stop offset="88%" stop-color="#818cf8" stop-opacity="0.2" />
            <stop offset="100%" stop-color="#312e81" stop-opacity="0.45" />
          </radialGradient>
          <filter
            id="moonGlowFilter"
            x="-30%"
            y="-30%"
            width="160%"
            height="160%"
          >
            <feDropShadow
              dx="0"
              dy="0"
              stdDeviation="6"
              flood-color="#c7d2fe"
              flood-opacity="0.5"
            />
            <feDropShadow
              dx="0"
              dy="4"
              stdDeviation="12"
              flood-color="#818cf8"
              flood-opacity="0.35"
            />
          </filter>
        </defs>

        <!-- Ambient Celestial Lunar Halo -->
        <circle cx="60" cy="50" r="44" fill="url(#moonAura)" />

        <!-- Starfield & Constellation Sparkles -->
        <g fill="#ffffff">
          <!-- 4-Point Primary Cross Sparkle -->
          <path
            d="M 28 22 L 29 25 L 32 26 L 29 27 L 28 30 L 27 27 L 24 26 L 27 25 Z"
            opacity="0.95"
            filter="drop-shadow(0 0 3px #ffffff)"
          />
          <path
            d="M 94 24 L 95 26 L 97 27 L 95 28 L 94 30 L 93 28 L 91 27 L 93 26 Z"
            opacity="0.9"
            filter="drop-shadow(0 0 3px #ffffff)"
          />
          <!-- Star Dots -->
          <circle
            cx="20"
            cy="52"
            r="1.4"
            opacity="0.8"
            filter="drop-shadow(0 0 2px #ffffff)"
          />
          <circle cx="36" cy="74" r="1.6" opacity="0.85" />
          <circle
            cx="98"
            cy="62"
            r="1.8"
            opacity="0.9"
            filter="drop-shadow(0 0 2px #ffffff)"
          />
          <circle cx="84" cy="78" r="1.3" opacity="0.75" />
          <circle cx="102" cy="42" r="1.2" opacity="0.7" />
        </g>

        <!-- Luminous Celestial Moon Sphere -->
        <g filter="url(#moonGlowFilter)">
          <circle cx="60" cy="50" r="22" fill="url(#moonBodyGrad)" />
          <!-- Limb Depth Overlay -->
          <circle cx="60" cy="50" r="22" fill="url(#moonLimbShadow)" />

          <!-- Soft Lunar Maria & Crater Formations -->
          <!-- Mare Tranquillitatis -->
          <circle cx="53" cy="43" r="4.8" fill="#94a3b8" opacity="0.28" />
          <circle cx="51.5" cy="42" r="3.6" fill="#64748b" opacity="0.18" />
          <!-- Mare Imbrium -->
          <circle cx="68" cy="42" r="3.8" fill="#94a3b8" opacity="0.26" />
          <!-- Tycho Crater -->
          <circle cx="62" cy="60" r="3.2" fill="#94a3b8" opacity="0.32" />
          <!-- Oceanus Procellarum / Minor Craters -->
          <circle cx="72" cy="52" r="2.4" fill="#94a3b8" opacity="0.22" />
          <circle cx="48" cy="56" r="2.0" fill="#94a3b8" opacity="0.2" />
          <circle cx="61" cy="49" r="1.6" fill="#94a3b8" opacity="0.16" />
        </g>
      </svg>
    {:else if type === 'rain'}
      <!-- Wide Volumetric Rain Cloud with Compact Fluid Raindrops -->
      <svg class="hero-svg" viewBox="0 0 120 100" aria-label="Rainy weather">
        <defs>
          <linearGradient id="rainBackCloud" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#64748b" />
            <stop offset="100%" stop-color="#334155" />
          </linearGradient>
          <linearGradient id="rainFrontCloud" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#ffffff" />
            <stop offset="45%" stop-color="#e2e8f0" />
            <stop offset="100%" stop-color="#94a3b8" />
          </linearGradient>
          <linearGradient id="rainDropGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#7dd3fc" />
            <stop offset="100%" stop-color="#0284c7" />
          </linearGradient>
        </defs>

        <!-- Secondary Background Storm Cloud -->
        <path
          d="M 42 50 A 14 14 0 0 1 30 36 A 16 16 0 0 1 48 20 A 18 18 0 0 1 80 18 A 18 18 0 0 1 98 34 A 14 14 0 0 1 98 50 Z"
          fill="url(#rainBackCloud)"
          opacity="0.6"
        />

        <!-- Wide Primary Foreground Cloud -->
        <path
          d="M 32 56 A 16 16 0 0 1 18 40 A 18 18 0 0 1 34 22 A 22 22 0 0 1 70 18 A 22 22 0 0 1 98 32 A 16 16 0 0 1 98 56 Z"
          fill="url(#rainFrontCloud)"
          filter="drop-shadow(0 4px 10px rgba(15, 23, 42, 0.32))"
        />

        <!-- Compact Angled Fluid Raindrops Spread Horizontally -->
        <g
          stroke="url(#rainDropGrad)"
          stroke-width="2.6"
          stroke-linecap="round"
          opacity="0.95"
        >
          <line x1="28" y1="63" x2="24" y2="74" />
          <line x1="43" y1="64" x2="39" y2="75" />
          <line x1="58" y1="63" x2="54" y2="74" />
          <line x1="73" y1="64" x2="69" y2="75" />
          <line x1="88" y1="63" x2="84" y2="74" />
        </g>
      </svg>
    {:else if type === 'thunder'}
      <!-- Volumetric Cumulonimbus Thunderhead with Electric Discharge -->
      <svg
        class="hero-svg"
        viewBox="0 0 120 100"
        aria-label="Thunderstorm weather"
      >
        <defs>
          <radialGradient id="thunderFlashGlow" cx="54%" cy="48%" r="48%">
            <stop offset="0%" stop-color="#ffffff" stop-opacity="0.5" />
            <stop offset="35%" stop-color="#c7d2fe" stop-opacity="0.25" />
            <stop offset="70%" stop-color="#818cf8" stop-opacity="0.08" />
            <stop offset="100%" stop-color="#4338ca" stop-opacity="0" />
          </radialGradient>
          <linearGradient id="thunderBackCloud" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#475569" />
            <stop offset="50%" stop-color="#312e81" />
            <stop offset="100%" stop-color="#0f172a" />
          </linearGradient>
          <linearGradient id="thunderFrontCloud" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#f8fafc" />
            <stop offset="16%" stop-color="#cbd5e1" />
            <stop offset="55%" stop-color="#475569" />
            <stop offset="100%" stop-color="#1e1b4b" />
          </linearGradient>
          <linearGradient id="thunderBoltGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#ffffff" />
            <stop offset="20%" stop-color="#e0f2fe" />
            <stop offset="50%" stop-color="#fef08a" />
            <stop offset="85%" stop-color="#facc15" />
            <stop offset="100%" stop-color="#f59e0b" />
          </linearGradient>
          <filter
            id="boltElectricGlow"
            x="-40%"
            y="-40%"
            width="180%"
            height="180%"
          >
            <feGaussianBlur
              in="SourceGraphic"
              stdDeviation="3.2"
              result="blurWide"
            />
            <feGaussianBlur
              in="SourceGraphic"
              stdDeviation="1.2"
              result="blurTight"
            />
            <feMerge>
              <feMergeNode in="blurWide" />
              <feMergeNode in="blurTight" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        <!-- Internal Lightning Discharge Glow -->
        <ellipse
          cx="60"
          cy="46"
          rx="48"
          ry="28"
          fill="url(#thunderFlashGlow)"
        />

        <!-- Rear Anvil Tempest Mass -->
        <path
          d="M 42 50 A 14 14 0 0 1 30 36 A 16 16 0 0 1 48 20 A 18 18 0 0 1 80 18 A 18 18 0 0 1 98 34 A 14 14 0 0 1 98 50 Z"
          fill="url(#thunderBackCloud)"
          opacity="0.6"
        />

        <!-- Realistic Branching Fractal Lightning Discharge (Positioned Behind Foreground Cloud) -->
        <g filter="url(#boltElectricGlow)">
          <!-- Outer Atmospheric Ionization Aura (Diffused Electric Violet-Cyan Halo) -->
          <g
            stroke="#818cf8"
            stroke-linecap="round"
            stroke-linejoin="miter"
            opacity="0.45"
            fill="none"
          >
            <!-- Primary Downward Strike Trunk -->
            <path
              d="M 58 26 L 54 36 L 62 45 L 53 54 L 59 63 L 49 74 L 55 82 L 47 91 L 51 96"
              stroke-width="3.8"
            />
            <!-- Major Right Lateral Fork -->
            <path d="M 59 63 L 69 69 L 76 77 L 84 82" stroke-width="2.6" />
            <path d="M 69 69 L 66 78 L 70 85" stroke-width="1.8" />
            <!-- Major Left Stepped Fork -->
            <path d="M 53 54 L 43 62 L 37 72 L 31 79" stroke-width="2.6" />
            <path d="M 43 62 L 47 70 L 44 77" stroke-width="1.8" />
            <!-- Lower Tertiary Step -->
            <path d="M 55 82 L 62 87 L 65 93" stroke-width="1.8" />
            <!-- Upper Intracloud Arc Behind Anvil -->
            <path d="M 54 36 L 43 32 L 36 36 L 29 33" stroke-width="2.2" />
          </g>

          <!-- Mid High-Energy Ionization Plasma Sheath (Electric Cyan-Gold) -->
          <g
            stroke="url(#thunderBoltGrad)"
            stroke-linecap="round"
            stroke-linejoin="miter"
            opacity="0.92"
            fill="none"
          >
            <!-- Primary Downward Strike Trunk -->
            <path
              d="M 58 26 L 54 36 L 62 45 L 53 54 L 59 63 L 49 74 L 55 82 L 47 91 L 51 96"
              stroke-width="2.0"
            />
            <!-- Major Right Lateral Fork -->
            <path d="M 59 63 L 69 69 L 76 77 L 84 82" stroke-width="1.3" />
            <path d="M 69 69 L 66 78 L 70 85" stroke-width="0.9" />
            <!-- Major Left Stepped Fork -->
            <path d="M 53 54 L 43 62 L 37 72 L 31 79" stroke-width="1.3" />
            <path d="M 43 62 L 47 70 L 44 77" stroke-width="0.9" />
            <!-- Lower Tertiary Step -->
            <path d="M 55 82 L 62 87 L 65 93" stroke-width="0.9" />
            <!-- Upper Intracloud Arc Behind Anvil -->
            <path d="M 54 36 L 43 32 L 36 36 L 29 33" stroke-width="1.1" />
          </g>

          <!-- Incandescent Superheated Plasma Core (Pure White Filament) -->
          <g
            stroke="#ffffff"
            stroke-linecap="round"
            stroke-linejoin="miter"
            opacity="1"
            fill="none"
          >
            <!-- Primary Downward Strike Trunk -->
            <path
              d="M 58 26 L 54 36 L 62 45 L 53 54 L 59 63 L 49 74 L 55 82 L 47 91 L 51 96"
              stroke-width="1.0"
            />
            <!-- Major Right Lateral Fork -->
            <path d="M 59 63 L 69 69 L 76 77 L 84 82" stroke-width="0.65" />
            <path d="M 69 69 L 66 78 L 70 85" stroke-width="0.45" />
            <!-- Major Left Stepped Fork -->
            <path d="M 53 54 L 43 62 L 37 72 L 31 79" stroke-width="0.65" />
            <path d="M 43 62 L 47 70 L 44 77" stroke-width="0.45" />
            <!-- Lower Tertiary Step -->
            <path d="M 55 82 L 62 87 L 65 93" stroke-width="0.45" />
            <!-- Upper Intracloud Arc Behind Anvil -->
            <path d="M 54 36 L 43 32 L 36 36 L 29 33" stroke-width="0.55" />
          </g>
        </g>

        <!-- Primary Billowing Storm Cloud with Silver Rim Highlight (In Front of Lightning) -->
        <path
          d="M 32 56 A 16 16 0 0 1 18 40 A 18 18 0 0 1 34 22 A 22 22 0 0 1 70 18 A 22 22 0 0 1 98 32 A 16 16 0 0 1 98 56 Z"
          fill="url(#thunderFrontCloud)"
          filter="drop-shadow(0 4px 14px rgba(15, 23, 42, 0.45))"
        />

        <!-- Ambient Ionized Precipitation Streaks -->
        <g
          stroke="#93c5fd"
          stroke-width="1.8"
          stroke-linecap="round"
          opacity="0.45"
        >
          <line x1="28" y1="62" x2="25" y2="72" />
          <line x1="38" y1="65" x2="35" y2="76" />
          <line x1="82" y1="64" x2="79" y2="75" />
          <line x1="92" y1="62" x2="89" y2="72" />
        </g>
      </svg>
    {:else if type === 'snow'}
      <!-- Wide Winter Cloud with Floating Snow Crystals -->
      <svg class="hero-svg" viewBox="0 0 120 100" aria-label="Snowy weather">
        <defs>
          <linearGradient id="snowCloud" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#ffffff" />
            <stop offset="60%" stop-color="#e2e8f0" />
            <stop offset="100%" stop-color="#94a3b8" />
          </linearGradient>
        </defs>

        <!-- Wide Frosty Cloud Body -->
        <path
          d="M 32 56 A 16 16 0 0 1 18 40 A 18 18 0 0 1 34 22 A 22 22 0 0 1 70 18 A 22 22 0 0 1 98 32 A 16 16 0 0 1 98 56 Z"
          fill="url(#snowCloud)"
          filter="drop-shadow(0 4px 10px rgba(15, 23, 42, 0.28))"
        />

        <!-- Falling Snow Crystals Spread Horizontally -->
        <g
          fill="#ffffff"
          opacity="0.95"
          filter="drop-shadow(0 0 3px rgba(255, 255, 255, 0.8))"
        >
          <circle cx="28" cy="67" r="2.2" />
          <circle cx="43" cy="73" r="1.8" />
          <circle cx="58" cy="66" r="2.4" />
          <circle cx="73" cy="73" r="1.8" />
          <circle cx="88" cy="67" r="2.2" />
        </g>
      </svg>
    {:else if type === 'fog'}
      <!-- Atmospheric Misty Cloud with Layered Drifting Fog Ribbons -->
      <svg class="hero-svg" viewBox="0 0 120 100" aria-label="Foggy weather">
        <defs>
          <linearGradient id="fogCloudBody" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#ffffff" />
            <stop offset="55%" stop-color="#e2e8f0" />
            <stop offset="100%" stop-color="#94a3b8" />
          </linearGradient>
          <linearGradient id="fogRibbonGrad1" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0%" stop-color="#cbd5e1" stop-opacity="0.15" />
            <stop offset="25%" stop-color="#ffffff" stop-opacity="0.9" />
            <stop offset="75%" stop-color="#e2e8f0" stop-opacity="0.9" />
            <stop offset="100%" stop-color="#cbd5e1" stop-opacity="0.15" />
          </linearGradient>
          <linearGradient id="fogRibbonGrad2" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0%" stop-color="#94a3b8" stop-opacity="0.15" />
            <stop offset="25%" stop-color="#cbd5e1" stop-opacity="0.8" />
            <stop offset="75%" stop-color="#e2e8f0" stop-opacity="0.8" />
            <stop offset="100%" stop-color="#94a3b8" stop-opacity="0.15" />
          </linearGradient>
        </defs>

        <!-- Wide Misty Floating Cloud Base -->
        <path
          d="M 32 54 A 16 16 0 0 1 18 38 A 18 18 0 0 1 34 20 A 22 22 0 0 1 70 16 A 22 22 0 0 1 98 30 A 16 16 0 0 1 98 54 Z"
          fill="url(#fogCloudBody)"
          filter="drop-shadow(0 4px 10px rgba(15, 23, 42, 0.28))"
          opacity="0.92"
        />

        <!-- Drifting Atmospheric Fog Ribbons Floating Across & Underneath Cloud -->
        <g>
          <rect
            x="20"
            y="53"
            width="80"
            height="5.5"
            rx="2.75"
            fill="url(#fogRibbonGrad1)"
          />
          <rect
            x="14"
            y="63"
            width="92"
            height="6"
            rx="3"
            fill="url(#fogRibbonGrad1)"
          />
          <rect
            x="26"
            y="73"
            width="68"
            height="5"
            rx="2.5"
            fill="url(#fogRibbonGrad2)"
          />
          <rect
            x="38"
            y="82"
            width="44"
            height="4"
            rx="2"
            fill="url(#fogRibbonGrad2)"
            opacity="0.6"
          />
        </g>
      </svg>
    {:else if type === 'cloudy-night'}
      <!-- Cloudy Night: Layered Volumetric Nocturnal Clouds with Ambient Starfield (No Moon) -->
      <svg class="hero-svg" viewBox="0 0 120 100" aria-label="Cloudy night sky">
        <defs>
          <radialGradient id="nightSkyAura" cx="50%" cy="40%" r="55%">
            <stop offset="0%" stop-color="#475569" stop-opacity="0.35" />
            <stop offset="60%" stop-color="#312e81" stop-opacity="0.18" />
            <stop offset="100%" stop-color="#1e1b4b" stop-opacity="0" />
          </radialGradient>
          <linearGradient id="nightCloudBack" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#475569" />
            <stop offset="100%" stop-color="#1e293b" />
          </linearGradient>
          <linearGradient id="nightCloudFront" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#cbd5e1" />
            <stop offset="40%" stop-color="#94a3b8" />
            <stop offset="100%" stop-color="#334155" />
          </linearGradient>
        </defs>

        <!-- Subtle Ambient Night Aura -->
        <circle cx="60" cy="44" r="36" fill="url(#nightSkyAura)" />

        <!-- Starfield Glimpses in Upper Night Sky -->
        <g fill="#ffffff">
          <circle
            cx="26"
            cy="22"
            r="1.6"
            opacity="0.85"
            filter="drop-shadow(0 0 3px #ffffff)"
          />
          <circle
            cx="94"
            cy="20"
            r="1.8"
            opacity="0.9"
            filter="drop-shadow(0 0 3px #ffffff)"
          />
          <circle cx="60" cy="14" r="1.3" opacity="0.75" />
          <circle cx="84" cy="34" r="1.2" opacity="0.65" />
        </g>

        <!-- Secondary Rear Ambient Night Cloud Depth -->
        <path
          d="M 44 56 A 14 14 0 0 1 32 42 A 16 16 0 0 1 50 26 A 18 18 0 0 1 82 24 A 18 18 0 0 1 100 40 A 14 14 0 0 1 100 56 Z"
          fill="url(#nightCloudBack)"
          opacity="0.65"
        />

        <!-- Wide Primary Foreground Nocturnal Fluffy Cloud -->
        <path
          d="M 32 64 A 16 16 0 0 1 18 48 A 18 18 0 0 1 34 30 A 22 22 0 0 1 70 26 A 22 22 0 0 1 98 40 A 16 16 0 0 1 98 64 Z"
          fill="url(#nightCloudFront)"
          filter="drop-shadow(0 4px 12px rgba(15, 23, 42, 0.45))"
        />
      </svg>
    {:else}
      <!-- Cloudy (Partly Cloudy): Sun Directly at the Back of Wide Fluffy Cloud -->
      <svg
        class="hero-svg"
        viewBox="0 0 120 100"
        aria-label="Partly cloudy weather"
      >
        <defs>
          <radialGradient id="sunBackAura" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stop-color="#ffffff" stop-opacity="0.45" />
            <stop offset="50%" stop-color="#fef08a" stop-opacity="0.2" />
            <stop offset="100%" stop-color="#fef08a" stop-opacity="0" />
          </radialGradient>
          <linearGradient id="sunBackBody" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stop-color="#ffffff" />
            <stop offset="35%" stop-color="#fef9c3" />
            <stop offset="100%" stop-color="#fde047" />
          </linearGradient>
          <linearGradient id="cloudyBackCloud" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#cbd5e1" />
            <stop offset="100%" stop-color="#94a3b8" />
          </linearGradient>
          <linearGradient id="cloudyFrontCloud" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#ffffff" />
            <stop offset="50%" stop-color="#f1f5f9" />
            <stop offset="100%" stop-color="#cbd5e1" />
          </linearGradient>
        </defs>

        <!-- Ambient Sun Glow at Back -->
        <circle cx="50" cy="34" r="28" fill="url(#sunBackAura)" />

        <!-- Corona Rays Radiating from the Sun Behind Cloud -->
        <g
          stroke="#fef08a"
          stroke-width="2.2"
          stroke-linecap="round"
          opacity="0.8"
        >
          <line x1="50" y1="8" x2="50" y2="14" />
          <line x1="30" y1="16" x2="35" y2="21" />
          <line x1="70" y1="16" x2="65" y2="21" />
          <line x1="22" y1="34" x2="28" y2="34" />
          <line x1="78" y1="34" x2="72" y2="34" />
        </g>

        <!-- Radiant Golden Sun Positioned Directly at the Back of Cloud -->
        <circle
          cx="50"
          cy="34"
          r="19"
          fill="url(#sunBackBody)"
          filter="drop-shadow(0 0 10px rgba(254, 240, 138, 0.45))"
        />

        <!-- Secondary Rear Ambient Cloud Depth -->
        <path
          d="M 44 56 A 14 14 0 0 1 32 42 A 16 16 0 0 1 50 26 A 18 18 0 0 1 82 24 A 18 18 0 0 1 100 40 A 14 14 0 0 1 100 56 Z"
          fill="url(#cloudyBackCloud)"
          opacity="0.55"
        />

        <!-- Wide Primary Foreground Fluffy Cloud in Front of Sun -->
        <path
          d="M 32 64 A 16 16 0 0 1 18 48 A 18 18 0 0 1 34 30 A 22 22 0 0 1 70 26 A 22 22 0 0 1 98 40 A 16 16 0 0 1 98 64 Z"
          fill="url(#cloudyFrontCloud)"
          filter="drop-shadow(0 4px 12px rgba(15, 23, 42, 0.28))"
        />
      </svg>
    {/if}
  </div>
{:else}
  <!-- =========================================================================
       COMPACT GLYPH VARIANT (Minimized Dynamic Island Capsule Pill)
       ========================================================================= -->
  <div class="weather-art-compact {className}">
    {#if type === 'sunny'}
      <svg
        class="compact-svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="#f59e0b"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <circle cx="12" cy="12" r="4" fill="none" />
        <line x1="12" y1="2" x2="12" y2="4" />
        <line x1="12" y1="20" x2="12" y2="22" />
        <line x1="4.93" y1="4.93" x2="6.34" y2="6.34" />
        <line x1="17.66" y1="17.66" x2="19.07" y2="19.07" />
        <line x1="2" y1="12" x2="4" y2="12" />
        <line x1="20" y1="12" x2="22" y2="12" />
        <line x1="4.93" y1="19.07" x2="6.34" y2="17.66" />
        <line x1="17.66" y1="6.34" x2="19.07" y2="4.93" />
      </svg>
    {:else if type === 'night'}
      <svg
        class="compact-svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="#818cf8"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M 12 3 A 9 9 0 1 0 21 16 A 7 7 0 0 1 12 3 Z" fill="none" />
      </svg>
    {:else if type === 'rain'}
      <svg
        class="compact-svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="#38bdf8"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path
          d="M4 14.899A7 7 0 1 1 15.71 8h1.79a4.5 4.5 0 0 1 2.5 8.242"
          fill="none"
        />
        <path d="M16 14v6" />
        <path d="M8 14v6" />
        <path d="M12 16v6" />
      </svg>
    {:else if type === 'thunder'}
      <svg
        class="compact-svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="#eab308"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path
          d="M6 16.326A7 7 0 1 1 15.71 8h1.79a4.5 4.5 0 0 1 .5 8.973"
          fill="none"
        />
        <path d="m13 12-3 5h4l-3 5" />
      </svg>
    {:else if type === 'snow'}
      <svg
        class="compact-svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="#38bdf8"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path
          d="M4 14.899A7 7 0 1 1 15.71 8h1.79a4.5 4.5 0 0 1 2.5 8.242"
          fill="none"
        />
        <path d="M8 15h.01" />
        <path d="M8 19h.01" />
        <path d="M12 17h.01" />
        <path d="M12 21h.01" />
        <path d="M16 15h.01" />
        <path d="M16 19h.01" />
      </svg>
    {:else if type === 'fog'}
      <!-- Standard well-proportioned CloudFog outline -->
      <svg
        class="compact-svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="#94a3b8"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path
          d="M4 14.899A7 7 0 1 1 15.71 8h1.79a4.5 4.5 0 0 1 2.5 8.242"
          fill="none"
        />
        <line x1="7" y1="17" x2="16" y2="17" />
        <line x1="9" y1="21" x2="17" y2="21" />
      </svg>
    {:else if type === 'cloudy-night'}
      <!-- Cloudy Night: Wide Fluffy Nocturnal Cloud (No Moon) -->
      <svg
        class="compact-svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="#818cf8"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path
          d="M17.5 19H6.5a4.5 4.5 0 0 1-1.2-8.8A5.5 5.5 0 0 1 16 9.5a4.5 4.5 0 0 1 1.5 9.5Z"
        />
      </svg>
    {:else}
      <!-- Cloudy (Partly Cloudy) with Wide Aspect Ratio and Sun at the Back -->
      <svg
        class="compact-svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="#60a5fa"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <line x1="12" y1="2" x2="12" y2="4" />
        <line x1="6.8" y1="4.2" x2="8.2" y2="5.6" />
        <line x1="17.2" y1="4.2" x2="15.8" y2="5.6" />
        <path d="M9 7a4.5 4.5 0 0 1 7.5 0" />
        <path
          d="M17.5 20H6.5a4.5 4.5 0 0 1-1.2-8.8A5.5 5.5 0 0 1 16 9.5a4.5 4.5 0 0 1 1.5 10.5Z"
          fill="none"
        />
      </svg>
    {/if}
  </div>
{/if}

<style>
  .weather-art-hero {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    pointer-events: none;
    user-select: none;
  }

  .hero-svg {
    width: 100%;
    height: 100%;
    overflow: visible;
    object-fit: contain;
  }

  .weather-art-compact {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    pointer-events: none;
    user-select: none;
  }

  .compact-svg {
    width: 100%;
    height: 100%;
  }
</style>
