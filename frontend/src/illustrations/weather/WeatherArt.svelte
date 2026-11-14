<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import { isIslandHiddenStore } from '../../store'

  /**
   * WeatherArt
   *
   * Reusable vector weather illustration supporting animated Vinta mascot hero artwork
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
    monochrome = false,
    class: className = ''
  }: {
    type?: WeatherConditionType
    variant?: 'hero' | 'compact' | 'mini'
    monochrome?: boolean
    class?: string
  } = $props()

  let heroEl = $state<HTMLDivElement | null>(null)
  let isInView = $state(true)
  let observer: IntersectionObserver | null = null

  onMount(() => {
    if (heroEl && typeof IntersectionObserver !== 'undefined') {
      observer = new IntersectionObserver(
        (entries) => {
          isInView = entries[0]?.isIntersecting ?? true
        },
        { threshold: 0.01 }
      )
      observer.observe(heroEl)
    }
  })

  onDestroy(() => {
    observer?.disconnect()
    observer = null
  })
</script>

{#if variant === 'hero'}
  <!-- =========================================================================
       HERO ANIMATED VINTA MASCOT ARTWORK (Banner Cards and Full Weather Page)
       ========================================================================= -->
  <div
    bind:this={heroEl}
    class="weather-art-hero {className}"
    class:is-paused={!isInView || $isIslandHiddenStore}
  >
    <svg class="hero-svg" viewBox="0 0 120 100" aria-label="{type} weather">
      <defs>
        <!-- Vinta Mascot Clip Paths -->
        <clipPath id="vintaHeadClip">
          <circle cx="60" cy="62" r="27" />
        </clipPath>

        <clipPath id="vintaEarsClip">
          <rect x="28" y="51" width="6" height="20" rx="3" />
          <rect x="86" y="51" width="6" height="20" rx="3" />
        </clipPath>

        <!-- Reusable Vinta Robot Base Chassis (Centered at 60, 62) -->
        <g id="vintaBotChassis">
          <!-- Soft Vector Drop Shadow Beneath Chassis -->
          <ellipse cx="60" cy="89" rx="22" ry="4.5" fill="rgba(15, 23, 42, 0.35)" />

          <!-- Side Ear Pods (Split Two-Tone Slate) -->
          <g clip-path="url(#vintaEarsClip)">
            <rect x="27" y="50" width="66" height="11" fill="#68788e" />
            <rect x="27" y="61" width="66" height="12" fill="#475569" />
          </g>

          <!-- Main Spherical Head Dome -->
          <g clip-path="url(#vintaHeadClip)">
            <!-- Upper Light Slate Metallic Half -->
            <rect x="30" y="32" width="60" height="30" fill="#94a3b8" />
            <!-- Lower Dark Slate Metallic Half -->
            <rect x="30" y="62" width="60" height="30" fill="#475569" />

            <!-- Forehead Geometric Crest: Outer Dark Navy Wedges -->
            <polygon points="36,41 40,37 51,48 48,48" fill="#182232" />
            <polygon points="84,41 80,37 69,48 72,48" fill="#182232" />

            <!-- Forehead Geometric Crest: Mid Slate Stripes -->
            <polygon points="40,37 46,33 55,48 51,48" fill="#3e4c5e" />
            <polygon points="80,37 74,33 65,48 69,48" fill="#3e4c5e" />

            <!-- Forehead Geometric Crest: Center Shield Diamond -->
            <polygon points="60,33 69,41 64,48 56,48 51,41" fill="#2e3b4e" />
          </g>

          <!-- Outer Visor Frame Bezel -->
          <rect
            x="34.5"
            y="46.5"
            width="51"
            height="34"
            rx="12.5"
            fill="#2b3748"
          />

          <!-- Inner Deep Space Visor Screen -->
          <rect x="37" y="49" width="46" height="29" rx="10.5" fill="#09101d" />
        </g>

        <!-- Weather Element Gradients & Filters -->
        <linearGradient id="sunCoreGrad" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#fef08a" />
          <stop offset="50%" stop-color="#facc15" />
          <stop offset="100%" stop-color="#f97316" />
        </linearGradient>

        <radialGradient id="sunAuraGrad" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stop-color="#fde047" stop-opacity="0.45" />
          <stop offset="60%" stop-color="#f59e0b" stop-opacity="0.15" />
          <stop offset="100%" stop-color="#f59e0b" stop-opacity="0" />
        </radialGradient>

        <linearGradient id="moonGrad" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#ffffff" />
          <stop offset="45%" stop-color="#fef9c3" />
          <stop offset="100%" stop-color="#facc15" />
        </linearGradient>

        <linearGradient id="cloudLightGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#ffffff" />
          <stop offset="60%" stop-color="#f1f5f9" />
          <stop offset="100%" stop-color="#cbd5e1" />
        </linearGradient>

        <linearGradient id="cloudNightLitGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#cbd5e1" />
          <stop offset="50%" stop-color="#64748b" />
          <stop offset="100%" stop-color="#334155" />
        </linearGradient>

        <linearGradient id="sweatGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#bae6fd" />
          <stop offset="100%" stop-color="#0284c7" />
        </linearGradient>

        <linearGradient id="umbrellaGrad" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#38bdf8" />
          <stop offset="100%" stop-color="#0284c7" />
        </linearGradient>

        <linearGradient id="beanieGrad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stop-color="#ef4444" />
          <stop offset="50%" stop-color="#dc2626" />
          <stop offset="100%" stop-color="#991b1b" />
        </linearGradient>

        <linearGradient id="fogBandGrad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stop-color="#cbd5e1" stop-opacity="0" />
          <stop offset="25%" stop-color="#f8fafc" stop-opacity="0.85" />
          <stop offset="75%" stop-color="#e2e8f0" stop-opacity="0.85" />
          <stop offset="100%" stop-color="#cbd5e1" stop-opacity="0" />
        </linearGradient>
      </defs>

      {#if type === 'sunny'}
        <!-- ===============================================================
             SUNNY: Blazing Sun Behind + Sweating Overheated Vinta Bot
             =============================================================== -->
        <circle
          class="anim-pulse-aura"
          cx="60"
          cy="40"
          r="46"
          fill="url(#sunAuraGrad)"
        />

        <!-- Rotating Sun at the Back -->
        <g class="anim-spin-sun" style="transform-origin: 60px 32px;">
          <g
            stroke="#fde047"
            stroke-width="2.8"
            stroke-linecap="round"
            opacity="0.85"
          >
            <line x1="60" y1="2" x2="60" y2="8" />
            <line x1="60" y1="56" x2="60" y2="62" />
            <line x1="30" y1="32" x2="36" y2="32" />
            <line x1="84" y1="32" x2="90" y2="32" />
            <line x1="39" y1="11" x2="43" y2="15" />
            <line x1="77" y1="49" x2="81" y2="53" />
            <line x1="39" y1="53" x2="43" y2="49" />
            <line x1="77" y1="15" x2="81" y2="11" />
          </g>
          <circle cx="60" cy="32" r="18" fill="url(#sunCoreGrad)" />
        </g>

        <!-- Vinta Mascot Character -->
        <g class="anim-float-hot">
          <use href="#vintaBotChassis" />

          <!-- Heat Flush Cheeks -->
          <ellipse
            cx="43"
            cy="69"
            rx="3.5"
            ry="2"
            fill="#fb7185"
            opacity="0.5"
          />
          <ellipse
            cx="77"
            cy="69"
            rx="3.5"
            ry="2"
            fill="#fb7185"
            opacity="0.5"
          />

          <!-- Squinting Golden Eyes & Overheated Wavy Mouth -->
          <g
            stroke="#facc15"
            stroke-width="2.6"
            stroke-linecap="round"
            stroke-linejoin="round"
            fill="none"
          >
            <path d="M 45 57 L 52 61 L 45 65" />
            <path d="M 75 57 L 68 61 L 75 65" />
            <path
              d="M 50 71 Q 53.5 68 57 71 T 64 71 T 70 71"
              stroke-width="2.2"
            />
          </g>

          <!-- Animated Sweat Drops -->
          <g class="anim-sweat-1">
            <path
              d="M 83 45 C 83 45 87 50 87 52.5 A 2.5 2.5 0 1 1 82 52.5 C 82 50 83 45 83 45 Z"
              fill="url(#sweatGrad)"
            />
          </g>
          <g class="anim-sweat-2">
            <path
              d="M 36 49 C 36 49 39 53 39 55 A 2 2 0 1 1 35 55 C 35 53 36 49 36 49 Z"
              fill="url(#sweatGrad)"
            />
          </g>
        </g>
      {:else if type === 'night'}
        <!-- ===============================================================
             NIGHT: Crescent Moon, Twinkling Stars + Sleeping Vinta Bot
             =============================================================== -->
        <g fill="#ffffff">
          <circle class="anim-twinkle-1" cx="22" cy="24" r="1.5" />
          <circle class="anim-twinkle-2" cx="46" cy="14" r="1.2" />
          <circle class="anim-twinkle-1" cx="100" cy="48" r="1.6" />
          <circle class="anim-twinkle-2" cx="18" cy="54" r="1.2" />
        </g>

        <!-- Glowing Crescent Moon in Background -->
        <path
          class="anim-moon-rock"
          style="transform-origin: 84px 26px;"
          d="M 86 10 C 75 10 66 19 66 30 C 66 37 70 44 76 47 C 72 44 70 39 70 34 C 70 24 78 16 88 16 C 91 16 94 17 97 19 C 94 13 90 10 86 10 Z"
          fill="url(#moonGrad)"
        />

        <!-- Sleeping Vinta Mascot -->
        <g class="anim-float-sleep">
          <use href="#vintaBotChassis" />

          <!-- Peaceful Sleeping Golden Arcs & Breathing Mouth -->
          <g
            stroke="#facc15"
            stroke-width="2.6"
            stroke-linecap="round"
            fill="none"
          >
            <path d="M 44 60 Q 49 66 54 60" />
            <path d="M 66 60 Q 71 66 76 60" />
            <circle
              class="anim-sleep-mouth"
              style="transform-origin: 60px 71px;"
              cx="60"
              cy="71"
              r="2.4"
              fill="#facc15"
              stroke="none"
            />
          </g>
        </g>

        <!-- Floating Zzz Particles -->
        <g
          fill="#fef08a"
          font-family="system-ui, sans-serif"
          font-weight="800"
          text-anchor="middle"
        >
          <text class="anim-zzz-1" x="33" y="38" font-size="9">Z</text>
          <text class="anim-zzz-2" x="25" y="28" font-size="7">z</text>
          <text class="anim-zzz-3" x="19" y="20" font-size="5.5">z</text>
        </g>
      {:else if type === 'cloudy'}
        <!-- ===============================================================
             CLOUDY: Peeking Sun + Drifting Fluffy Clouds + Signature Vinta Bot
             =============================================================== -->
        <circle
          cx="34"
          cy="24"
          r="14"
          fill="url(#sunCoreGrad)"
        />

        <!-- Background Cloud Puff -->
        <path
          class="anim-cloud-drift-slow"
          d="M 22 42 A 9 9 0 0 1 15 33 A 10 10 0 0 1 26 23 A 12 12 0 0 1 46 21 A 11 11 0 0 1 56 32 A 9 9 0 0 1 56 42 Z"
          fill="url(#cloudLightGrad)"
          opacity="0.88"
        />

        <!-- Vinta Mascot with Exact Signature Squircle Eyes & Filled Smile -->
        <g class="anim-float-calm">
          <use href="#vintaBotChassis" />

          <g
            class="anim-look-around"
            fill="#facc15"
          >
            <!-- Left & Right Signature Rounded Squircle Eyes -->
            <rect x="43" y="53" width="13" height="13" rx="5.5" />
            <rect x="64" y="53" width="13" height="13" rx="5.5" />
            <!-- Signature Filled Crescent Smile -->
            <path d="M 52 68.5 Q 60 70.5 68 68.5 C 67 75.5 53 75.5 52 68.5 Z" />
          </g>
        </g>

        <!-- Foreground Fluffy Cloud Overlapping Top-Right of Bot -->
        <path
          class="anim-cloud-drift"
          d="M 62 48 A 8 8 0 0 1 55 40 A 9 9 0 0 1 65 31 A 11 11 0 0 1 84 29 A 10 10 0 0 1 95 39 A 8 8 0 0 1 95 48 Z"
          fill="url(#cloudLightGrad)"
        />
      {:else if type === 'cloudy-night'}
        <!-- ===============================================================
             CLOUDY NIGHT: Moon + Nocturnal Clouds + Cozy Vinta Bot
             =============================================================== -->
        <g fill="#ffffff">
          <circle class="anim-twinkle-1" cx="22" cy="20" r="1.4" />
          <circle class="anim-twinkle-2" cx="98" cy="18" r="1.5" />
        </g>

        <!-- Crescent Moon Behind Cloud -->
        <path
          d="M 42 10 C 33 10 26 17 26 26 C 26 32 29 37 34 40 C 31 37 29 33 29 29 C 29 21 35 15 43 15 C 46 15 48 16 50 17 C 48 12 45 10 42 10 Z"
          fill="url(#moonGrad)"
        />

        <!-- Vinta Mascot with Happy Nocturnal Expression -->
        <g class="anim-float-calm">
          <use href="#vintaBotChassis" />

          <g
            class="anim-look-around"
            fill="#facc15"
          >
            <rect x="43" y="53.5" width="13" height="12.5" rx="5.5" />
            <rect x="64" y="53.5" width="13" height="12.5" rx="5.5" />
            <path d="M 52.5 69 Q 60 70.5 67.5 69 C 66.5 75 53.5 75 52.5 69 Z" />
          </g>
        </g>

        <!-- Foreground Nocturnal Cloud -->
        <path
          class="anim-cloud-drift"
          d="M 58 47 A 8 8 0 0 1 51 39 A 9 9 0 0 1 61 30 A 11 11 0 0 1 80 28 A 10 10 0 0 1 91 38 A 8 8 0 0 1 91 47 Z"
          fill="url(#cloudNightLitGrad)"
          stroke="rgba(226, 232, 240, 0.38)"
          stroke-width="0.8"
        />
      {:else if type === 'rain'}
        <!-- ===============================================================
             RAIN: Storm Cloud + Falling Raindrops + Vinta Bot Under Umbrella
             =============================================================== -->
        <!-- Upper Rain Cloud -->
        <path
          class="anim-cloud-drift-slow"
          d="M 32 29 A 10 10 0 0 1 23 19 A 12 12 0 0 1 36 7 A 15 15 0 0 1 62 5 A 14 14 0 0 1 80 17 A 10 10 0 0 1 80 29 Z"
          fill="url(#cloudNightLitGrad)"
          stroke="rgba(226, 232, 240, 0.3)"
          stroke-width="0.8"
          opacity="0.92"
        />

        <!-- Animated Falling Rain Streaks -->
        <g
          stroke="#38bdf8"
          stroke-width="2"
          stroke-linecap="round"
          opacity="0.85"
        >
          <line class="anim-rain-1" x1="25" y1="33" x2="22" y2="42" />
          <line class="anim-rain-2" x1="42" y1="31" x2="39" y2="40" />
          <line class="anim-rain-3" x1="60" y1="33" x2="57" y2="42" />
          <line class="anim-rain-1" x1="78" y1="32" x2="75" y2="41" />
          <line class="anim-rain-2" x1="93" y1="39" x2="90" y2="48" />
        </g>

        <!-- Vinta Mascot Sheltering Under Umbrella -->
        <g class="anim-float-calm">
          <use href="#vintaBotChassis" />

          <!-- Pensive / Gloomy Golden Eyes & Pout -->
          <g fill="#facc15">
            <!-- Flat-topped Droopy Squircle Eyes -->
            <path
              d="M 43 57 L 56 59 L 56 63 A 4 4 0 0 1 52 67 L 47 67 A 4 4 0 0 1 43 63 Z"
            />
            <path
              d="M 77 57 L 64 59 L 64 63 A 4 4 0 0 0 68 67 L 73 67 A 4 4 0 0 0 77 63 Z"
            />
            <!-- Small Pout Mouth -->
            <path
              d="M 55 73 Q 60 69.5 65 73"
              fill="none"
              stroke="#facc15"
              stroke-width="2.4"
              stroke-linecap="round"
            />
          </g>

          <!-- Sleek Cyber Umbrella Canopy Over Vinta Bot -->
          <g class="anim-umbrella-tilt" style="transform-origin: 60px 38px;">
            <path
              d="M 60 24 L 60 39"
              stroke="#94a3b8"
              stroke-width="2.5"
              stroke-linecap="round"
            />
            <path
              d="M 22 39 C 22 19 98 19 98 39 Q 85 35 72 39 Q 60 35 48 39 Q 35 35 22 39 Z"
              fill="url(#umbrellaGrad)"
            />
            <circle cx="60" cy="21" r="2.5" fill="#38bdf8" />
          </g>
        </g>
      {:else if type === 'thunder'}
        <!-- ===============================================================
             THUNDER: Flashing Lightning + Startled Shivering Vinta Bot
             =============================================================== -->
        <!-- Dark Thunderhead at Top -->
        <path
          d="M 26 31 A 11 11 0 0 1 16 20 A 13 13 0 0 1 30 7 A 16 16 0 0 1 62 4 A 15 15 0 0 1 88 17 A 11 11 0 0 1 88 31 Z"
          fill="url(#cloudNightLitGrad)"
          stroke="rgba(226, 232, 240, 0.32)"
          stroke-width="0.8"
        />

        <!-- Flashing Lightning Bolts Left & Right -->
        <polygon
          class="anim-lightning-1"
          points="27,28 17,46 26,46 20,64 36,42 26,42"
          fill="#facc15"
        />
        <polygon
          class="anim-lightning-2"
          points="89,26 80,43 88,43 83,59 97,39 88,39"
          fill="#fef08a"
        />

        <!-- Startled Jittering Vinta Mascot -->
        <g class="anim-shock-shake">
          <use href="#vintaBotChassis" />

          <!-- Shocked Hollow Golden Squircle Eyes & Surprised Mouth -->
          <g>
            <rect
              x="44"
              y="54"
              width="11"
              height="11"
              rx="4.5"
              fill="none"
              stroke="#facc15"
              stroke-width="2.5"
            />
            <circle cx="49.5" cy="59.5" r="1.8" fill="#fef08a" />

            <rect
              x="65"
              y="54"
              width="11"
              height="11"
              rx="4.5"
              fill="none"
              stroke="#facc15"
              stroke-width="2.5"
            />
            <circle cx="70.5" cy="59.5" r="1.8" fill="#fef08a" />

            <ellipse cx="60" cy="71.5" rx="4.2" ry="3.2" fill="#facc15" />
          </g>
        </g>
      {:else if type === 'snow'}
        <!-- ===============================================================
             SNOW: Falling Snowflakes + Chilly Vinta Bot in Winter Beanie
             =============================================================== -->
        <!-- Falling Snowflakes -->
        <g fill="#ffffff" opacity="0.9">
          <circle class="anim-snow-1" cx="22" cy="24" r="2.2" />
          <circle class="anim-snow-2" cx="40" cy="16" r="1.8" />
          <circle class="anim-snow-3" cx="82" cy="18" r="2.4" />
          <circle class="anim-snow-1" cx="98" cy="28" r="1.7" />
          <circle class="anim-snow-2" cx="16" cy="55" r="1.9" />
          <circle class="anim-snow-3" cx="102" cy="60" r="2.1" />
        </g>

        <!-- Chilly Shivering Vinta Mascot with Knit Beanie -->
        <g class="anim-shiver-cold">
          <use href="#vintaBotChassis" />

          <!-- Rosy Cold Cheeks -->
          <ellipse
            cx="43"
            cy="69"
            rx="3.5"
            ry="2.2"
            fill="#f43f5e"
            opacity="0.55"
          />
          <ellipse
            cx="77"
            cy="69"
            rx="3.5"
            ry="2.2"
            fill="#f43f5e"
            opacity="0.55"
          />

          <!-- Chilly Golden Squircle Eyes & Shivering Wavy Mouth -->
          <g>
            <rect
              x="44"
              y="54"
              width="11.5"
              height="11.5"
              rx="5"
              fill="#facc15"
            />
            <rect
              x="64.5"
              y="54"
              width="11.5"
              height="11.5"
              rx="5"
              fill="#facc15"
            />
            <path
              d="M 51 71.5 Q 53.2 69 55.5 71.5 T 60 71.5 T 64.5 71.5 T 69 71.5"
              fill="none"
              stroke="#facc15"
              stroke-width="2.2"
              stroke-linecap="round"
            />
          </g>

          <!-- Cozy Winter Beanie Hat Sitting on Vinta Dome -->
          <g>
            <!-- Fluffy Pom-Pom -->
            <circle cx="60" cy="21" r="5.5" fill="#ffffff" />
            <!-- Beanie Dome -->
            <path d="M 35 43 C 35 23 85 23 85 43 Z" fill="url(#beanieGrad)" />
            <!-- Ribbed Beanie Cuff -->
            <rect
              x="32"
              y="38.5"
              width="56"
              height="8.5"
              rx="4.25"
              fill="#f8fafc"
            />
            <line
              x1="44"
              y1="39.5"
              x2="44"
              y2="46"
              stroke="#cbd5e1"
              stroke-width="1.5"
              stroke-linecap="round"
            />
            <line
              x1="54"
              y1="39.5"
              x2="54"
              y2="46"
              stroke="#cbd5e1"
              stroke-width="1.5"
              stroke-linecap="round"
            />
            <line
              x1="66"
              y1="39.5"
              x2="66"
              y2="46"
              stroke="#cbd5e1"
              stroke-width="1.5"
              stroke-linecap="round"
            />
            <line
              x1="76"
              y1="39.5"
              x2="76"
              y2="46"
              stroke="#cbd5e1"
              stroke-width="1.5"
              stroke-linecap="round"
            />
          </g>
        </g>
      {:else if type === 'fog'}
        <!-- ===============================================================
             FOG: Drifting Mist Layers + Vinta Bot Peering Side-to-Side
             =============================================================== -->
        <!-- Back Fog Ribbon -->
        <rect
          class="anim-fog-back"
          x="10"
          y="33"
          width="90"
          height="7"
          rx="3.5"
          fill="url(#fogBandGrad)"
          opacity="0.65"
        />

        <!-- Vinta Mascot Peering Through the Mist -->
        <g class="anim-float-calm">
          <use href="#vintaBotChassis" />

          <!-- Narrowed Golden Squircle Eyes Scanning Left & Right -->
          <g class="anim-scan-eyes" fill="#facc15">
            <rect x="43" y="56.5" width="12.5" height="6.5" rx="3.2" />
            <rect x="64.5" y="56.5" width="12.5" height="6.5" rx="3.2" />
            <path
              d="M 54 70.5 L 65 69"
              stroke="#facc15"
              stroke-width="2.4"
              stroke-linecap="round"
            />
          </g>
        </g>

        <!-- Foreground Drifting Fog Layers Partially Obscuring Vinta Bot -->
        <rect
          class="anim-fog-front-1"
          x="14"
          y="65"
          width="92"
          height="6.5"
          rx="3.25"
          fill="url(#fogBandGrad)"
        />
        <rect
          class="anim-fog-front-2"
          x="22"
          y="77"
          width="76"
          height="6"
          rx="3"
          fill="url(#fogBandGrad)"
          opacity="0.9"
        />
      {/if}
    </svg>
  </div>
{:else}
  <!-- =========================================================================
       COMPACT GLYPH VARIANT (Minimized Dynamic Island Capsule Pill)
       ========================================================================= -->
  <div
    class="weather-art-compact {className}"
    class:variant-mini={variant === 'mini'}
    class:monochrome
  >
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
    contain: layout style paint;
  }

  .weather-art-hero.is-paused :global(*) {
    animation-play-state: paused !important;
  }

  .hero-svg {
    width: 100%;
    height: 100%;
    overflow: visible;
    object-fit: contain;
  }

  /* =========================================================================
     HERO MASCOT & WEATHER ANIMATIONS
     ========================================================================= */

  /* Gentle Floating Motions */
  .anim-float-calm {
    animation: floatCalm 4s ease-in-out infinite;
  }
  .anim-float-hot {
    animation: floatHot 2.6s ease-in-out infinite;
  }
  .anim-float-sleep {
    animation: floatSleep 5s ease-in-out infinite;
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
  @keyframes floatHot {
    0%,
    100% {
      transform: translateY(0px);
    }
    50% {
      transform: translateY(2.5px);
    }
  }
  @keyframes floatSleep {
    0%,
    100% {
      transform: translateY(0px) rotate(-1deg);
    }
    50% {
      transform: translateY(-3px) rotate(1deg);
    }
  }

  /* Sun & Heat Animations */
  .anim-spin-sun {
    animation: spinSun 24s linear infinite;
  }
  .anim-pulse-aura {
    animation: pulseAura 3s ease-in-out infinite;
    transform-origin: 60px 40px;
  }
  .anim-sweat-1 {
    animation: sweatDrop 2.2s ease-in infinite;
  }
  .anim-sweat-2 {
    animation: sweatDrop 2.2s ease-in 1.1s infinite;
  }

  @keyframes spinSun {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
  @keyframes pulseAura {
    0%,
    100% {
      transform: scale(0.95);
      opacity: 0.8;
    }
    50% {
      transform: scale(1.08);
      opacity: 1;
    }
  }
  @keyframes sweatDrop {
    0% {
      transform: translateY(0px);
      opacity: 0;
    }
    20% {
      opacity: 1;
    }
    80% {
      opacity: 1;
    }
    100% {
      transform: translateY(11px);
      opacity: 0;
    }
  }

  /* Night & Sleep Animations */
  .anim-moon-rock {
    animation: moonRock 6s ease-in-out infinite;
  }
  .anim-sleep-mouth {
    animation: sleepBreath 5s ease-in-out infinite;
  }
  .anim-twinkle-1 {
    animation: twinkle 2.5s ease-in-out infinite;
  }
  .anim-twinkle-2 {
    animation: twinkle 3.2s ease-in-out 1.2s infinite;
  }
  .anim-zzz-1 {
    animation: floatZzz 3.6s ease-out infinite;
  }
  .anim-zzz-2 {
    animation: floatZzz 3.6s ease-out 1.2s infinite;
  }
  .anim-zzz-3 {
    animation: floatZzz 3.6s ease-out 2.4s infinite;
  }

  @keyframes moonRock {
    0%,
    100% {
      transform: rotate(-4deg);
    }
    50% {
      transform: rotate(4deg);
    }
  }
  @keyframes sleepBreath {
    0%,
    100% {
      transform: scale(0.85);
      opacity: 0.7;
    }
    50% {
      transform: scale(1.35);
      opacity: 1;
    }
  }
  @keyframes twinkle {
    0%,
    100% {
      opacity: 0.25;
      transform: scale(0.8);
    }
    50% {
      opacity: 1;
      transform: scale(1.2);
    }
  }
  @keyframes floatZzz {
    0% {
      transform: translate(0px, 4px) scale(0.7);
      opacity: 0;
    }
    30% {
      opacity: 0.95;
    }
    100% {
      transform: translate(-8px, -10px) scale(1.1);
      opacity: 0;
    }
  }

  /* Cloud & Eye Glancing Animations */
  .anim-cloud-drift {
    animation: cloudDrift 5s ease-in-out infinite;
  }
  .anim-cloud-drift-slow {
    animation: cloudDrift 7s ease-in-out reverse infinite;
  }
  .anim-look-around {
    animation: lookAround 6s ease-in-out infinite;
  }

  @keyframes cloudDrift {
    0%,
    100% {
      transform: translateX(-3px);
    }
    50% {
      transform: translateX(4px);
    }
  }
  @keyframes lookAround {
    0%,
    100% {
      transform: translate(0px, 0px);
    }
    30% {
      transform: translate(2.5px, -1.5px);
    }
    70% {
      transform: translate(-2px, -1px);
    }
  }

  /* Rain & Umbrella Animations */
  .anim-rain-1 {
    animation: rainFall 1.1s linear infinite;
  }
  .anim-rain-2 {
    animation: rainFall 1.1s linear 0.35s infinite;
  }
  .anim-rain-3 {
    animation: rainFall 1.1s linear 0.7s infinite;
  }
  .anim-umbrella-tilt {
    animation: umbrellaTilt 3.5s ease-in-out infinite;
  }

  @keyframes rainFall {
    0% {
      transform: translate(0px, -4px);
      opacity: 0;
    }
    30% {
      opacity: 0.9;
    }
    100% {
      transform: translate(-4px, 14px);
      opacity: 0;
    }
  }
  @keyframes umbrellaTilt {
    0%,
    100% {
      transform: rotate(-2deg);
    }
    50% {
      transform: rotate(2.5deg);
    }
  }

  /* Thunder & Shock Animations */
  .anim-lightning-1 {
    animation: flashBolt 3.2s infinite;
  }
  .anim-lightning-2 {
    animation: flashBolt 3.2s 1.6s infinite;
  }
  .anim-shock-shake {
    animation: shockShake 1.6s ease-in-out infinite;
  }

  @keyframes flashBolt {
    0%,
    78%,
    100% {
      opacity: 0;
    }
    80%,
    86% {
      opacity: 1;
    }
    83% {
      opacity: 0.2;
    }
  }
  @keyframes shockShake {
    0%,
    70%,
    100% {
      transform: translate(0, 0);
    }
    75% {
      transform: translate(-2px, 1px);
    }
    80% {
      transform: translate(2px, -1px);
    }
    85% {
      transform: translate(-2px, -1px);
    }
    90% {
      transform: translate(2px, 1px);
    }
  }

  /* Snow & Shivering Animations */
  .anim-snow-1 {
    animation: snowDrift 3.2s linear infinite;
  }
  .anim-snow-2 {
    animation: snowDrift 3.2s linear 1.1s infinite;
  }
  .anim-snow-3 {
    animation: snowDrift 3.2s linear 2.2s infinite;
  }
  .anim-shiver-cold {
    animation: shiverCold 0.45s ease-in-out infinite;
  }

  @keyframes snowDrift {
    0% {
      transform: translate(0px, -6px);
      opacity: 0;
    }
    25% {
      opacity: 0.95;
    }
    75% {
      opacity: 0.95;
    }
    100% {
      transform: translate(4px, 24px);
      opacity: 0;
    }
  }
  @keyframes shiverCold {
    0%,
    100% {
      transform: translateX(-0.7px);
    }
    50% {
      transform: translateX(0.7px);
    }
  }

  /* Fog & Scanning Animations */
  .anim-fog-back {
    animation: fogSlide 6s ease-in-out infinite;
  }
  .anim-fog-front-1 {
    animation: fogSlide 4.5s ease-in-out reverse infinite;
  }
  .anim-fog-front-2 {
    animation: fogSlide 5.5s ease-in-out 1s infinite;
  }
  .anim-scan-eyes {
    animation: scanEyes 4.5s ease-in-out infinite;
  }

  @keyframes fogSlide {
    0%,
    100% {
      transform: translateX(-6px);
    }
    50% {
      transform: translateX(6px);
    }
  }
  @keyframes scanEyes {
    0%,
    100% {
      transform: translateX(-3px);
    }
    50% {
      transform: translateX(3px);
    }
  }

  /* =========================================================================
     COMPACT GLYPH STYLES
     ========================================================================= */
  .weather-art-compact {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    pointer-events: none;
    user-select: none;
  }

  .weather-art-compact.variant-mini {
    width: 22px;
    height: 22px;
  }

  .compact-svg {
    width: 100%;
    height: 100%;
  }

  .weather-art-compact.monochrome :global(svg) {
    stroke: #ffffff !important;
  }

  .weather-art-compact.monochrome :global(svg path),
  .weather-art-compact.monochrome :global(svg line),
  .weather-art-compact.monochrome :global(svg circle) {
    stroke: #ffffff !important;
  }
</style>
