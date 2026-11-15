<!--
  @component WeatherGlance
  @description Minimized weather preview showing current temperature, condition icon, and city name.
-->
<script lang="ts">
  import { onMount, getContext } from 'svelte'
  import { MapPin } from 'lucide-svelte'
  import {
    weatherLocationStore,
    weatherUnitStore,
    simulateOfflineStore,
    simulateWeatherStore,
    tStore,
    type IslandConfig
  } from '../store'
  import {
    fetchWeatherData,
    convertTemp,
    getWmoConditionKey,
    type WeatherData
  } from '../services/weatherService'
  import WeatherPage, {
    config as weatherPageConfig
  } from '../pages/WeatherPage.svelte'
  import WeatherAtmosphere from './WeatherAtmosphere.svelte'
  import WeatherArt, {
    type WeatherConditionType
  } from '../illustrations/weather/WeatherArt.svelte'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let {
    salutation = '',
    userName = '',
    timeText = '',
    amPmText = ''
  }: {
    salutation?: string
    userName?: string
    timeText?: string
    amPmText?: string
  } = $props()

  let weather = $state<WeatherData | null>(null)
  let _loading = $state<boolean>(true)
  let _error = $state<boolean>(false)
  let lastFetchKey = ''

  /**
   * Fetches weather metrics for the active location.
   */
  async function loadWeather(
    name: string,
    lat: number,
    lon: number,
    offline: boolean
  ) {
    _loading = true
    _error = false

    try {
      const data = await fetchWeatherData(name, lat, lon, offline)
      if (data) {
        weather = data
        _error = false
      } else {
        _error = true
      }
    } catch {
      _error = true
    } finally {
      _loading = false
    }
  }

  onMount(() => {
    function checkAndFetch() {
      const isOffline = Boolean($simulateOfflineStore)
      const loc = $weatherLocationStore
      const key = `${loc.lat.toFixed(4)}_${loc.lon.toFixed(4)}_${isOffline}`
      if (key === lastFetchKey) return
      lastFetchKey = key
      loadWeather(loc.name, loc.lat, loc.lon, isOffline)
    }

    checkAndFetch()

    const unsubLoc = weatherLocationStore.subscribe(checkAndFetch)
    const unsubOff = simulateOfflineStore.subscribe(checkAndFetch)
    const unsubMock = simulateWeatherStore.subscribe((mock) => {
      if (mock) {
        weather = mock
        _loading = false
        _error = false
      } else {
        lastFetchKey = ''
        checkAndFetch()
      }
    })

    return () => {
      unsubLoc()
      unsubOff()
      unsubMock()
    }
  })

  function computeDisplayTemp() {
    if (!weather) return { val: 0, symbol: '°', str: '--' }
    return convertTemp(weather.temp, $weatherUnitStore)
  }
  let displayTemp = $derived.by(computeDisplayTemp)

  function computeUnitLetter(): string {
    if ($weatherUnitStore === 'fahrenheit') return 'F'
    if ($weatherUnitStore === 'kelvin') return 'K'
    return 'C'
  }
  let unitLetter = $derived.by(computeUnitLetter)

  function openWeatherPage() {
    if (navigate) {
      navigate(WeatherPage, weatherPageConfig, true)
    }
  }

  function computeIsNight() {
    if (typeof weather?.isNight === 'boolean') {
      return weather.isNight
    }
    const hour = new Date().getHours()
    return hour < 6 || hour >= 18
  }
  let isNight = $derived.by(computeIsNight)

  function computeIllustrationType(): WeatherConditionType {
    if (!weather) return isNight ? 'cloudy-night' : 'cloudy'
    const code = weather.weatherCode
    if (code >= 95) return 'thunder'
    if ((code >= 51 && code <= 67) || (code >= 80 && code <= 82)) return 'rain'
    if (code >= 71 && code <= 77) return 'snow'
    if (code === 0) return isNight ? 'night' : 'sunny'
    if (code >= 1 && code <= 3) return isNight ? 'cloudy-night' : 'cloudy'
    if (code === 45 || code === 48) return 'fog'
    return isNight ? 'cloudy-night' : 'cloudy'
  }
  let illustrationType = $derived.by(computeIllustrationType)

  function computeBannerConditionTitle() {
    if (!weather) {
      return isNight
        ? $tStore('weather.conditions.clear_night')
        : $tStore('weather.conditions.sunny')
    }
    const key = getWmoConditionKey(
      weather.weatherCode ?? 0,
      isNight,
      weather.condition
    )
    return $tStore(key)
  }
  let bannerConditionTitle = $derived.by(computeBannerConditionTitle)

  function computeShortCityName() {
    const raw = $weatherLocationStore.name || 'Auto'
    return raw.split(',')[0].trim()
  }
  let shortCityName = $derived.by(computeShortCityName)
</script>

<div class="weather-glance-hero {illustrationType}">
  <WeatherAtmosphere condition={illustrationType} seamless={true} />

  <div class="home-hero-layout">
    <div class="home-hero-left">
      {#if salutation}
        <span class="hero-banner-salutation">{salutation},</span>
      {/if}
      {#if userName}
        <h1 class="hero-banner-username">{userName}</h1>
      {/if}
      <div class="hero-banner-info-stack">
        <div class="hero-banner-weather-row">
          <span class="hero-weather-temp">
            {weather ? `${displayTemp.val}°${unitLetter}` : `--°${unitLetter}`}
          </span>
          <span class="sub-divider">•</span>
          <span class="hero-weather-cond">
            {weather ? bannerConditionTitle : $tStore('common.loading')}
          </span>
        </div>
        <div class="hero-banner-loc-row">
          {#if $weatherLocationStore.name || shortCityName}
            <MapPin size={10} class="loc-pin" />
          {/if}
          <span>{$weatherLocationStore.name || shortCityName || '\u00A0'}</span>
        </div>

        {#if timeText}
          <div class="hero-banner-time-row">
            <span class="hero-banner-time">{timeText}</span>
            {#if amPmText}
              <span class="hero-banner-ampm">{amPmText}</span>
            {/if}
          </div>
        {/if}
      </div>
    </div>

    <div class="home-hero-art">
      <div class="weather-art-wrap">
        <div
          class="weather-art-link"
          data-weather-mascot
          role="button"
          tabindex="0"
          aria-label="Open weather page"
          onclick={(e) => {
            e.stopPropagation()
            openWeatherPage()
          }}
          onkeydown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault()
              openWeatherPage()
            }
          }}
        >
          <WeatherArt type={illustrationType} variant="hero" />
        </div>
      </div>
    </div>
  </div>
</div>

<style src="./styles/WeatherGlance.css"></style>
