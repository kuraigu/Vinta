<script lang="ts">
  /**
   * WeatherArt
   *
   * Reusable vector weather illustration supporting animated mascot hero artwork
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
</script>

{#if variant === 'hero'}
  <!-- =========================================================================
       HERO ANIMATED MASCOT ARTWORK (Banner Cards and Full Weather Page)
       ========================================================================= -->
  <div class="weather-art-hero {className}">
    <svg class="hero-svg" viewBox="0 0 120 100" aria-label="{type} weather">
      <defs>
        <!-- Shared Mascot Gradients & Filters -->
        <linearGradient id="botChassis" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#ffffff" />
          <stop offset="65%" stop-color="#e2e8f0" />
          <stop offset="100%" stop-color="#94a3b8" />
        </linearGradient>

        <linearGradient id="botChassisNight" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#e2e8f0" />
          <stop offset="60%" stop-color="#94a3b8" />
          <stop offset="100%" stop-color="#475569" />
        </linearGradient>

        <linearGradient id="botVisor" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#1e293b" />
          <stop offset="100%" stop-color="#090d16" />
        </linearGradient>

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

        <linearGradient id="cloudDarkGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#64748b" />
          <stop offset="55%" stop-color="#334155" />
          <stop offset="100%" stop-color="#1e293b" />
        </linearGradient>

        <!-- High-Contrast Moonlit Nocturnal & Storm Cloud Gradient -->
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

        <filter id="cyanGlow" x="-30%" y="-30%" width="160%" height="160%">
          <feDropShadow
            dx="0"
            dy="0"
            stdDeviation="1.8"
            flood-color="#38bdf8"
            flood-opacity="0.85"
          />
        </filter>

        <filter id="amberGlow" x="-30%" y="-30%" width="160%" height="160%">
          <feDropShadow
            dx="0"
            dy="0"
            stdDeviation="2"
            flood-color="#facc15"
            flood-opacity="0.9"
          />
        </filter>

        <filter id="softShadow" x="-20%" y="-20%" width="140%" height="140%">
          <feDropShadow
            dx="0"
            dy="5"
            stdDeviation="5"
            flood-color="#0f172a"
            flood-opacity="0.32"
          />
        </filter>
      </defs>

      {#if type === 'sunny'}
        <!-- ===============================================================
             SUNNY: Blazing Sun Behind + Sweating Overheated Bot
             =============================================================== -->
        <!-- Ambient Heat Aura -->
        <circle
          class="anim-pulse-aura"
          cx="60"
          cy="42"
          r="46"
          fill="url(#sunAuraGrad)"
        />

        <!-- Rotating Sun at the Back -->
        <g class="anim-spin-sun" style="transform-origin: 60px 36px;">
          <g
            stroke="#fde047"
            stroke-width="2.8"
            stroke-linecap="round"
            opacity="0.85"
          >
            <line x1="60" y1="4" x2="60" y2="10" />
            <line x1="60" y1="62" x2="60" y2="68" />
            <line x1="28" y1="36" x2="34" y2="36" />
            <line x1="86" y1="36" x2="92" y2="36" />
            <line x1="37" y1="13" x2="42" y2="18" />
            <line x1="78" y1="54" x2="83" y2="59" />
            <line x1="37" y1="59" x2="42" y2="54" />
            <line x1="78" y1="18" x2="83" y2="13" />
          </g>
          <circle cx="60" cy="36" r="20" fill="url(#sunCoreGrad)" />
        </g>

        <!-- Mascot Character -->
        <g class="anim-float-hot" filter="url(#softShadow)">
          <!-- Antenna (Drooping slightly from heat) -->
          <path
            d="M 60 43 Q 63 36 67 37"
            fill="none"
            stroke="#cbd5e1"
            stroke-width="2.5"
            stroke-linecap="round"
          />
          <circle cx="68" cy="37" r="3" fill="#fb923c" />

          <!-- Side Ear Pods -->
          <rect x="29" y="57" width="6" height="14" rx="3" fill="#94a3b8" />
          <rect x="85" y="57" width="6" height="14" rx="3" fill="#94a3b8" />

          <!-- Main Head Chassis -->
          <rect
            x="33"
            y="42"
            width="54"
            height="42"
            rx="20"
            fill="url(#botChassis)"
          />

          <!-- Glossy Visor Screen -->
          <rect
            x="38"
            y="48"
            width="44"
            height="30"
            rx="14"
            fill="url(#botVisor)"
          />
          <!-- Visor Specular Highlight -->
          <path
            d="M 44 51 L 56 51"
            stroke="#ffffff"
            stroke-width="1.5"
            stroke-linecap="round"
            opacity="0.22"
          />

          <!-- Heat Flush Cheeks -->
          <ellipse
            cx="45"
            cy="68"
            rx="3.5"
            ry="2"
            fill="#fb7185"
            opacity="0.55"
          />
          <ellipse
            cx="75"
            cy="68"
            rx="3.5"
            ry="2"
            fill="#fb7185"
            opacity="0.55"
          />

          <!-- Strained Squinting Eyes & Sweating Wavy Mouth -->
          <g
            filter="url(#cyanGlow)"
            stroke="#38bdf8"
            stroke-width="2.4"
            stroke-linecap="round"
            stroke-linejoin="round"
            fill="none"
          >
            <!-- Left Squint Eye (> ) -->
            <path d="M 46 57 L 52 60 L 46 63" />
            <!-- Right Squint Eye ( <) -->
            <path d="M 74 57 L 68 60 L 74 63" />
            <!-- Overheated Wavy Mouth -->
            <path
              d="M 51 70 Q 54 67 57 70 T 63 70 T 69 70"
              stroke-width="2.1"
            />
          </g>

          <!-- Animated Sweat Drops -->
          <g class="anim-sweat-1">
            <path
              d="M 82 47 C 82 47 86 52 86 54.5 A 2.5 2.5 0 1 1 81 54.5 C 81 52 82 47 82 47 Z"
              fill="url(#sweatGrad)"
            />
          </g>
          <g class="anim-sweat-2">
            <path
              d="M 37 51 C 37 51 40 55 40 57 A 2 2 0 1 1 36 57 C 36 55 37 51 37 51 Z"
              fill="url(#sweatGrad)"
            />
          </g>
        </g>
      {:else if type === 'night'}
        <!-- ===============================================================
             NIGHT: Crescent Moon, Twinkling Stars + Sleeping Bot with Zzz
             =============================================================== -->
        <!-- Twinkling Starfield -->
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
          filter="url(#amberGlow)"
        />

        <!-- Sleeping Mascot -->
        <g class="anim-float-sleep" filter="url(#softShadow)">
          <!-- Antenna with Dimmed Night Bulb -->
          <line
            x1="60"
            y1="36"
            x2="60"
            y2="43"
            stroke="#64748b"
            stroke-width="2.5"
            stroke-linecap="round"
          />
          <circle cx="60" cy="35" r="3" fill="#818cf8" opacity="0.8" />

          <!-- Side Ear Pods -->
          <rect x="29" y="57" width="6" height="14" rx="3" fill="#475569" />
          <rect x="85" y="57" width="6" height="14" rx="3" fill="#475569" />

          <!-- Main Head Chassis -->
          <rect
            x="33"
            y="42"
            width="54"
            height="42"
            rx="20"
            fill="url(#botChassisNight)"
          />

          <!-- Glossy Visor Screen -->
          <rect
            x="38"
            y="48"
            width="44"
            height="30"
            rx="14"
            fill="url(#botVisor)"
          />
          <path
            d="M 44 51 L 56 51"
            stroke="#ffffff"
            stroke-width="1.5"
            stroke-linecap="round"
            opacity="0.15"
          />

          <!-- Peaceful Sleeping Eyes & Breathing Mouth -->
          <g
            filter="url(#cyanGlow)"
            stroke="#818cf8"
            stroke-width="2.3"
            stroke-linecap="round"
            fill="none"
          >
            <path d="M 45 59 Q 49 64 53 59" />
            <path d="M 67 59 Q 71 64 75 59" />
            <circle
              class="anim-sleep-mouth"
              style="transform-origin: 60px 69px;"
              cx="60"
              cy="69"
              r="2.2"
              fill="#818cf8"
              stroke="none"
            />
          </g>
        </g>

        <!-- Floating Zzz Particles -->
        <g
          fill="#c7d2fe"
          font-family="system-ui, sans-serif"
          font-weight="800"
          text-anchor="middle"
        >
          <text class="anim-zzz-1" x="34" y="40" font-size="9">Z</text>
          <text class="anim-zzz-2" x="26" y="30" font-size="7">z</text>
          <text class="anim-zzz-3" x="20" y="22" font-size="5.5">z</text>
        </g>
      {:else if type === 'cloudy'}
        <!-- ===============================================================
             CLOUDY: Peeking Sun + Drifting Fluffy Clouds + Curious Bot
             =============================================================== -->
        <!-- Peeking Warm Sun at Upper Left -->
        <circle
          cx="34"
          cy="26"
          r="14"
          fill="url(#sunCoreGrad)"
          filter="url(#amberGlow)"
        />

        <!-- Background Cloud Puff -->
        <path
          class="anim-cloud-drift-slow"
          d="M 22 44 A 9 9 0 0 1 15 35 A 10 10 0 0 1 26 25 A 12 12 0 0 1 46 23 A 11 11 0 0 1 56 34 A 9 9 0 0 1 56 44 Z"
          fill="url(#cloudLightGrad)"
          opacity="0.88"
        />

        <!-- Mascot Looking Up at the Clouds -->
        <g class="anim-float-calm" filter="url(#softShadow)">
          <line
            x1="60"
            y1="36"
            x2="60"
            y2="43"
            stroke="#cbd5e1"
            stroke-width="2.5"
            stroke-linecap="round"
          />
          <circle cx="60" cy="35" r="3" fill="#38bdf8" />

          <rect x="29" y="57" width="6" height="14" rx="3" fill="#94a3b8" />
          <rect x="85" y="57" width="6" height="14" rx="3" fill="#94a3b8" />

          <rect
            x="33"
            y="42"
            width="54"
            height="42"
            rx="20"
            fill="url(#botChassis)"
          />
          <rect
            x="38"
            y="48"
            width="44"
            height="30"
            rx="14"
            fill="url(#botVisor)"
          />
          <path
            d="M 44 51 L 56 51"
            stroke="#ffffff"
            stroke-width="1.5"
            stroke-linecap="round"
            opacity="0.22"
          />

          <!-- Happy Eyes Glancing Upward + Gentle Smile -->
          <g class="anim-look-around" filter="url(#cyanGlow)">
            <ellipse cx="49" cy="58" rx="3.2" ry="4" fill="#38bdf8" />
            <ellipse cx="71" cy="58" rx="3.2" ry="4" fill="#38bdf8" />
            <circle cx="50" cy="56.5" r="1" fill="#ffffff" />
            <circle cx="72" cy="56.5" r="1" fill="#ffffff" />
            <path
              d="M 54 68 Q 60 73 66 68"
              fill="none"
              stroke="#38bdf8"
              stroke-width="2.3"
              stroke-linecap="round"
            />
          </g>
        </g>

        <!-- Foreground Fluffy Cloud Overlapping Top-Right of Bot -->
        <path
          class="anim-cloud-drift"
          d="M 62 49 A 8 8 0 0 1 55 41 A 9 9 0 0 1 65 32 A 11 11 0 0 1 84 30 A 10 10 0 0 1 95 40 A 8 8 0 0 1 95 49 Z"
          fill="url(#cloudLightGrad)"
          filter="url(#softShadow)"
        />
      {:else if type === 'cloudy-night'}
        <!-- ===============================================================
             CLOUDY NIGHT: Moon + Nocturnal Clouds + Cozy Night Bot
             =============================================================== -->
        <g fill="#ffffff">
          <circle class="anim-twinkle-1" cx="22" cy="20" r="1.4" />
          <circle class="anim-twinkle-2" cx="98" cy="18" r="1.5" />
        </g>

        <!-- Crescent Moon Behind Cloud -->
        <path
          d="M 42 12 C 33 12 26 19 26 28 C 26 34 29 39 34 42 C 31 39 29 35 29 31 C 29 23 35 17 43 17 C 46 17 48 18 50 19 C 48 14 45 12 42 12 Z"
          fill="url(#moonGrad)"
          filter="url(#amberGlow)"
        />

        <!-- Mascot in Night Palette -->
        <g class="anim-float-calm" filter="url(#softShadow)">
          <line
            x1="60"
            y1="36"
            x2="60"
            y2="43"
            stroke="#64748b"
            stroke-width="2.5"
            stroke-linecap="round"
          />
          <circle cx="60" cy="35" r="3" fill="#818cf8" />

          <rect x="29" y="57" width="6" height="14" rx="3" fill="#475569" />
          <rect x="85" y="57" width="6" height="14" rx="3" fill="#475569" />

          <rect
            x="33"
            y="42"
            width="54"
            height="42"
            rx="20"
            fill="url(#botChassisNight)"
          />
          <rect
            x="38"
            y="48"
            width="44"
            height="30"
            rx="14"
            fill="url(#botVisor)"
          />
          <path
            d="M 44 51 L 56 51"
            stroke="#ffffff"
            stroke-width="1.5"
            stroke-linecap="round"
            opacity="0.15"
          />

          <!-- Calm Nocturnal Eyes & Cozy Smile -->
          <g
            filter="url(#cyanGlow)"
            stroke="#a5b4fc"
            stroke-width="2.3"
            stroke-linecap="round"
            fill="none"
          >
            <path d="M 45 60 Q 49 55 53 60" />
            <path d="M 67 60 Q 71 55 75 60" />
            <path d="M 55 69 Q 60 72 65 69" />
          </g>
        </g>

        <!-- Foreground Nocturnal Cloud (Moonlit Slate with Rim Highlight) -->
        <path
          class="anim-cloud-drift"
          d="M 58 48 A 8 8 0 0 1 51 40 A 9 9 0 0 1 61 31 A 11 11 0 0 1 80 29 A 10 10 0 0 1 91 39 A 8 8 0 0 1 91 48 Z"
          fill="url(#cloudNightLitGrad)"
          stroke="rgba(226, 232, 240, 0.38)"
          stroke-width="0.8"
          filter="url(#softShadow)"
        />
      {:else if type === 'rain'}
        <!-- ===============================================================
             RAIN: Storm Cloud + Falling Raindrops + Bot Under Umbrella
             =============================================================== -->
        <!-- Upper Rain Cloud -->
        <path
          class="anim-cloud-drift-slow"
          d="M 32 32 A 10 10 0 0 1 23 22 A 12 12 0 0 1 36 10 A 15 15 0 0 1 62 8 A 14 14 0 0 1 80 20 A 10 10 0 0 1 80 32 Z"
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
          <line class="anim-rain-1" x1="26" y1="36" x2="23" y2="45" />
          <line class="anim-rain-2" x1="42" y1="34" x2="39" y2="43" />
          <line class="anim-rain-3" x1="60" y1="36" x2="57" y2="45" />
          <line class="anim-rain-1" x1="78" y1="35" x2="75" y2="44" />
          <line class="anim-rain-2" x1="92" y1="42" x2="89" y2="51" />
        </g>

        <!-- Mascot Sheltering Under Umbrella -->
        <g class="anim-float-calm" filter="url(#softShadow)">
          <rect x="29" y="61" width="6" height="13" rx="3" fill="#64748b" />
          <rect x="85" y="61" width="6" height="13" rx="3" fill="#64748b" />

          <rect
            x="33"
            y="47"
            width="54"
            height="40"
            rx="19"
            fill="url(#botChassis)"
          />
          <rect
            x="38"
            y="53"
            width="44"
            height="28"
            rx="13"
            fill="url(#botVisor)"
          />

          <!-- Pensive / Gloomy Rainy Expression -->
          <g
            filter="url(#cyanGlow)"
            stroke="#38bdf8"
            stroke-width="2.2"
            stroke-linecap="round"
            fill="none"
          >
            <!-- Slightly Droopy Eyes -->
            <line x1="45" y1="63" x2="53" y2="61" />
            <line x1="75" y1="63" x2="67" y2="61" />
            <circle cx="49" cy="65" r="2" fill="#38bdf8" stroke="none" />
            <circle cx="71" cy="65" r="2" fill="#38bdf8" stroke="none" />
            <!-- Small Pout Mouth -->
            <path d="M 55 74 Q 60 70 65 74" />
          </g>

          <!-- Sleek Cyber Umbrella Canopy Over Bot -->
          <g class="anim-umbrella-tilt" style="transform-origin: 60px 44px;">
            <!-- Umbrella Handle Shaft -->
            <path
              d="M 60 32 L 60 48"
              stroke="#94a3b8"
              stroke-width="2.5"
              stroke-linecap="round"
            />
            <!-- Glowing Umbrella Canopy -->
            <path
              d="M 24 44 C 24 24 96 24 96 44 Q 84 40 72 44 Q 60 40 48 44 Q 36 40 24 44 Z"
              fill="url(#umbrellaGrad)"
            />
            <circle cx="60" cy="28" r="2.5" fill="#38bdf8" />
          </g>
        </g>
      {:else if type === 'thunder'}
        <!-- ===============================================================
             THUNDER: Flashing Lightning + Startled Shivering Bot
             =============================================================== -->
        <!-- Dark Thunderhead at Top -->
        <path
          d="M 26 34 A 11 11 0 0 1 16 23 A 13 13 0 0 1 30 10 A 16 16 0 0 1 62 7 A 15 15 0 0 1 88 20 A 11 11 0 0 1 88 34 Z"
          fill="url(#cloudNightLitGrad)"
          stroke="rgba(226, 232, 240, 0.32)"
          stroke-width="0.8"
          filter="url(#softShadow)"
        />

        <!-- Flashing Lightning Bolts Left & Right -->
        <polygon
          class="anim-lightning-1"
          points="28,30 18,48 27,48 21,66 37,44 27,44"
          fill="#facc15"
          filter="url(#amberGlow)"
        />
        <polygon
          class="anim-lightning-2"
          points="88,28 79,45 87,45 82,61 96,41 87,41"
          fill="#fef08a"
          filter="url(#amberGlow)"
        />

        <!-- Startled Jittering Mascot -->
        <g class="anim-shock-shake" filter="url(#softShadow)">
          <!-- Electrified Antenna -->
          <line
            x1="60"
            y1="36"
            x2="60"
            y2="45"
            stroke="#fde047"
            stroke-width="2.5"
            stroke-linecap="round"
          />
          <circle
            cx="60"
            cy="35"
            r="3.8"
            fill="#facc15"
            filter="url(#amberGlow)"
          />

          <rect x="29" y="59" width="6" height="14" rx="3" fill="#64748b" />
          <rect x="85" y="59" width="6" height="14" rx="3" fill="#64748b" />

          <rect
            x="33"
            y="44"
            width="54"
            height="42"
            rx="20"
            fill="url(#botChassis)"
          />
          <rect
            x="38"
            y="50"
            width="44"
            height="30"
            rx="14"
            fill="url(#botVisor)"
          />

          <!-- Shocked Ring Eyes & Zigzag / 'O' Mouth -->
          <g filter="url(#amberGlow)">
            <circle
              cx="49"
              cy="61"
              r="4.5"
              fill="none"
              stroke="#facc15"
              stroke-width="2.2"
            />
            <circle cx="49" cy="61" r="1.5" fill="#ffffff" />
            <circle
              cx="71"
              cy="61"
              r="4.5"
              fill="none"
              stroke="#facc15"
              stroke-width="2.2"
            />
            <circle cx="71" cy="61" r="1.5" fill="#ffffff" />
            <!-- Surprised 'O' Mouth -->
            <ellipse
              cx="60"
              cy="72"
              rx="4"
              ry="3.2"
              fill="none"
              stroke="#facc15"
              stroke-width="2.2"
            />
          </g>
        </g>
      {:else if type === 'snow'}
        <!-- ===============================================================
             SNOW: Falling Snowflakes + Chilly Bot in Cozy Winter Beanie
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

        <!-- Chilly Shivering Mascot with Knit Beanie -->
        <g class="anim-shiver-cold" filter="url(#softShadow)">
          <rect x="29" y="59" width="6" height="14" rx="3" fill="#94a3b8" />
          <rect x="85" y="59" width="6" height="14" rx="3" fill="#94a3b8" />

          <!-- Main Head Chassis -->
          <rect
            x="33"
            y="44"
            width="54"
            height="42"
            rx="20"
            fill="url(#botChassis)"
          />
          <rect
            x="38"
            y="50"
            width="44"
            height="30"
            rx="14"
            fill="url(#botVisor)"
          />

          <!-- Rosy Cold Cheeks -->
          <ellipse
            cx="44"
            cy="69"
            rx="3.5"
            ry="2.2"
            fill="#f43f5e"
            opacity="0.6"
          />
          <ellipse
            cx="76"
            cy="69"
            rx="3.5"
            ry="2.2"
            fill="#f43f5e"
            opacity="0.6"
          />

          <!-- Chilly Wide Eyes & Shivering Wavy Mouth -->
          <g filter="url(#cyanGlow)">
            <circle cx="49" cy="61" r="3.4" fill="#7dd3fc" />
            <circle cx="71" cy="61" r="3.4" fill="#7dd3fc" />
            <circle cx="50" cy="59.8" r="1.1" fill="#ffffff" />
            <circle cx="72" cy="59.8" r="1.1" fill="#ffffff" />
            <path
              d="M 51 72 Q 53.2 69.5 55.5 72 T 60 72 T 64.5 72 T 69 72"
              fill="none"
              stroke="#7dd3fc"
              stroke-width="2"
              stroke-linecap="round"
            />
          </g>

          <!-- Cozy Winter Beanie Hat Sitting on Top of Head -->
          <g>
            <!-- Fluffy Pom-Pom -->
            <circle cx="60" cy="25" r="6" fill="#ffffff" />
            <!-- Beanie Dome -->
            <path d="M 35 46 C 35 27 85 27 85 46 Z" fill="url(#beanieGrad)" />
            <!-- Ribbed Beanie Cuff -->
            <rect x="32" y="42" width="56" height="9" rx="4.5" fill="#f8fafc" />
            <line
              x1="44"
              y1="43"
              x2="44"
              y2="50"
              stroke="#cbd5e1"
              stroke-width="1.5"
              stroke-linecap="round"
            />
            <line
              x1="54"
              y1="43"
              x2="54"
              y2="50"
              stroke="#cbd5e1"
              stroke-width="1.5"
              stroke-linecap="round"
            />
            <line
              x1="66"
              y1="43"
              x2="66"
              y2="50"
              stroke="#cbd5e1"
              stroke-width="1.5"
              stroke-linecap="round"
            />
            <line
              x1="76"
              y1="43"
              x2="76"
              y2="50"
              stroke="#cbd5e1"
              stroke-width="1.5"
              stroke-linecap="round"
            />
          </g>
        </g>
      {:else if type === 'fog'}
        <!-- ===============================================================
             FOG: Drifting Mist Layers + Bot Peering Side-to-Side
             =============================================================== -->
        <!-- Back Fog Ribbon -->
        <rect
          class="anim-fog-back"
          x="10"
          y="34"
          width="90"
          height="7"
          rx="3.5"
          fill="url(#fogBandGrad)"
          opacity="0.65"
        />

        <!-- Mascot Peering Through the Mist -->
        <g class="anim-float-calm" filter="url(#softShadow)">
          <line
            x1="60"
            y1="34"
            x2="60"
            y2="42"
            stroke="#cbd5e1"
            stroke-width="2.5"
            stroke-linecap="round"
          />
          <circle
            cx="60"
            cy="33"
            r="3.2"
            fill="#fde047"
            filter="url(#amberGlow)"
          />

          <rect x="29" y="56" width="6" height="14" rx="3" fill="#94a3b8" />
          <rect x="85" y="56" width="6" height="14" rx="3" fill="#94a3b8" />

          <rect
            x="33"
            y="41"
            width="54"
            height="42"
            rx="20"
            fill="url(#botChassis)"
          />
          <rect
            x="38"
            y="47"
            width="44"
            height="30"
            rx="14"
            fill="url(#botVisor)"
          />

          <!-- Searching / Squinting Eyes Scanning Left and Right -->
          <g class="anim-scan-eyes" filter="url(#cyanGlow)">
            <!-- Narrowed Scanning Eyes -->
            <rect
              x="44"
              y="56"
              width="9"
              height="4.5"
              rx="2.2"
              fill="#38bdf8"
            />
            <rect
              x="65"
              y="56"
              width="9"
              height="4.5"
              rx="2.2"
              fill="#38bdf8"
            />
            <!-- Uncertain Crooked Mouth -->
            <path
              d="M 54 69 L 64 67"
              stroke="#38bdf8"
              stroke-width="2.2"
              stroke-linecap="round"
            />
          </g>
        </g>

        <!-- Foreground Drifting Fog Layers Partially Obscuring Bot -->
        <rect
          class="anim-fog-front-1"
          x="14"
          y="64"
          width="92"
          height="6.5"
          rx="3.25"
          fill="url(#fogBandGrad)"
        />
        <rect
          class="anim-fog-front-2"
          x="22"
          y="76"
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
    transform-origin: 60px 42px;
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
