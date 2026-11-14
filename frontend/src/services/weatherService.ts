/**
 * @fileoverview Open-Meteo client, caching, and WMO weather code helpers.
 */

import {
  Sun,
  Moon,
  CloudSun,
  CloudMoon,
  Cloud,
  CloudRain,
  CloudSnow,
  CloudLightning
} from 'lucide-svelte'

export type ForecastDay = {
  day: string
  tempMax: number
  tempMin: number
  weatherCode: number
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  icon: any
}

export type WeatherData = {
  location: string
  temp: number
  feelsLike: number
  condition: string
  high: number
  low: number
  humidity: number
  windSpeed: number
  windDirection: number
  windDirectionLabel: string
  pressure: number
  precipitationProbability: number
  uvIndex: string
  weatherCode: number
  forecast: ForecastDay[]
  sunrise?: string
  sunset?: string
  isNight?: boolean
  isCached?: boolean
  isOffline?: boolean
  lastUpdated: number
}

export type GeocodingResult = {
  name: string
  country?: string
  admin1?: string
  latitude: number
  longitude: number
  timezone?: string
}

export type WeatherTheme = {
  color: string
  glow: string
  badgeBg: string
}

export type TempFormatted = {
  val: number
  symbol: string
  str: string
}

const CACHE_TTL = 15 * 60 * 1000
const DAY_NAMES = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
const WIND_DIRS = ['N', 'NE', 'E', 'SE', 'S', 'SW', 'W', 'NW']

const CONDITION_STRING_MAP: Record<string, string> = {
  'clear sky': 'weather.conditions.clear_sky',
  'clear night': 'weather.conditions.clear_night',
  sunny: 'weather.conditions.sunny',
  'mainly clear': 'weather.conditions.mainly_clear',
  'partly cloudy': 'weather.conditions.partly_cloudy',
  'cloudy night': 'weather.conditions.cloudy_night',
  overcast: 'weather.conditions.overcast',
  foggy: 'weather.conditions.foggy',
  rainy: 'weather.conditions.rainy',
  snowy: 'weather.conditions.snowy',
  thunderstorm: 'weather.conditions.thunderstorm'
}

const THEMES: Record<string, WeatherTheme> = {
  night: {
    color: '#818cf8',
    glow: 'rgba(129, 140, 248, 0.25)',
    badgeBg: 'rgba(129, 140, 248, 0.15)'
  },
  clear: {
    color: '#f59e0b',
    glow: 'rgba(245, 158, 11, 0.22)',
    badgeBg: 'rgba(245, 158, 11, 0.15)'
  },
  cloudy: {
    color: '#38bdf8',
    glow: 'rgba(56, 189, 248, 0.2)',
    badgeBg: 'rgba(56, 189, 248, 0.15)'
  },
  fog: {
    color: '#94a3b8',
    glow: 'rgba(148, 163, 184, 0.2)',
    badgeBg: 'rgba(148, 163, 184, 0.15)'
  },
  rain: {
    color: '#3b82f6',
    glow: 'rgba(59, 130, 246, 0.25)',
    badgeBg: 'rgba(59, 130, 246, 0.15)'
  },
  snow: {
    color: '#06b6d4',
    glow: 'rgba(6, 182, 212, 0.25)',
    badgeBg: 'rgba(6, 182, 212, 0.15)'
  },
  thunder: {
    color: '#a855f7',
    glow: 'rgba(168, 85, 247, 0.3)',
    badgeBg: 'rgba(168, 85, 247, 0.18)'
  }
}

const memoryCache = new Map<string, WeatherData>()
const inflightRequests = new Map<string, Promise<WeatherData | null>>()

function getCacheKey(lat: number, lon: number): string {
  return `vinta_weather_cache_${lat.toFixed(4)}_${lon.toFixed(4)}`
}

/**
 * Re-attaches Lucide icon components lost during JSON serialization.
 */
function hydrateWeather(data: WeatherData): WeatherData {
  if (Array.isArray(data.forecast)) {
    data.forecast = data.forecast.map((f) => ({
      ...f,
      icon: parseWmoCode(f.weatherCode).icon
    }))
  }
  return data
}

/**
 * Reads cached weather from memory or localStorage without waiting for async ticks.
 */
export function getCachedWeatherData(
  lat: number,
  lon: number,
  allowExpired = false
): WeatherData | null {
  const key = getCacheKey(lat, lon)
  const mem = memoryCache.get(key)

  if (mem) {
    if (allowExpired || Date.now() - mem.lastUpdated <= CACHE_TTL) {
      return { ...mem, isCached: true }
    }
  }

  try {
    const raw = localStorage.getItem(key)
    if (!raw) return null
    const parsed = hydrateWeather(JSON.parse(raw) as WeatherData)
    memoryCache.set(key, parsed)

    if (allowExpired || Date.now() - parsed.lastUpdated <= CACHE_TTL) {
      return { ...parsed, isCached: true }
    }
  } catch {
    localStorage.removeItem(key)
  }

  return null
}

/**
 * Maps a WMO code or condition label to its i18n key.
 */
export function getWmoConditionKey(
  code: number,
  isNight: boolean = false,
  fallbackCondition?: string
): string {
  if (fallbackCondition) {
    const mapped = CONDITION_STRING_MAP[fallbackCondition.toLowerCase().trim()]
    if (mapped) return mapped
  }
  if (code === 0) {
    return isNight
      ? 'weather.conditions.clear_night'
      : 'weather.conditions.clear_sky'
  }
  if (code === 1) return 'weather.conditions.mainly_clear'
  if (code === 2) {
    return isNight
      ? 'weather.conditions.cloudy_night'
      : 'weather.conditions.partly_cloudy'
  }
  if (code === 3) return 'weather.conditions.overcast'
  if (code === 45 || code === 48) return 'weather.conditions.foggy'
  if ((code >= 51 && code <= 67) || (code >= 80 && code <= 82)) {
    return 'weather.conditions.rainy'
  }
  if (code >= 71 && code <= 77) return 'weather.conditions.snowy'
  if (code >= 95 && code <= 99) return 'weather.conditions.thunderstorm'
  return isNight
    ? 'weather.conditions.cloudy_night'
    : 'weather.conditions.partly_cloudy'
}

/**
 * Returns display label, Lucide icon, and color tokens for a WMO code.
 */
export function parseWmoCode(
  code: number,
  isNight: boolean = false
): {
  condition: string
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  icon: any
  theme: WeatherTheme
} {
  if (code === 0) {
    return {
      condition: isNight ? 'Clear Night' : 'Clear Sky',
      icon: isNight ? Moon : Sun,
      theme: isNight ? THEMES.night : THEMES.clear
    }
  }
  if (code >= 1 && code <= 3) {
    return {
      condition:
        code === 1 ? 'Mainly Clear' : code === 2 ? 'Partly Cloudy' : 'Overcast',
      icon: isNight ? CloudMoon : CloudSun,
      theme: isNight ? THEMES.night : THEMES.cloudy
    }
  }
  if (code === 45 || code === 48) {
    return { condition: 'Foggy', icon: Cloud, theme: THEMES.fog }
  }
  if ((code >= 51 && code <= 67) || (code >= 80 && code <= 82)) {
    return { condition: 'Rainy', icon: CloudRain, theme: THEMES.rain }
  }
  if (code >= 71 && code <= 77) {
    return { condition: 'Snowy', icon: CloudSnow, theme: THEMES.snow }
  }
  if (code >= 95 && code <= 99) {
    return {
      condition: 'Thunderstorm',
      icon: CloudLightning,
      theme: THEMES.thunder
    }
  }
  return { condition: 'Partly Cloudy', icon: CloudSun, theme: THEMES.cloudy }
}

/**
 * Loads current conditions and 7-day forecast from Open-Meteo.
 * Deduplicates concurrent calls and serves cached data when valid.
 */
export async function fetchWeatherData(
  locationName: string,
  lat: number,
  lon: number,
  simulateOffline = false,
  forceRefresh = false
): Promise<WeatherData | null> {
  const cacheKey = getCacheKey(lat, lon)

  if (simulateOffline) {
    const fallback = getCachedWeatherData(lat, lon, true)
    return fallback ? { ...fallback, isCached: true, isOffline: true } : null
  }

  if (!forceRefresh) {
    const freshCache = getCachedWeatherData(lat, lon, false)
    if (freshCache) {
      return { ...freshCache, location: locationName, isOffline: false }
    }
  }

  const existingRequest = inflightRequests.get(cacheKey)
  if (existingRequest) return existingRequest

  const requestPromise = (async (): Promise<WeatherData | null> => {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 6000)

    try {
      const url = `https://api.open-meteo.com/v1/forecast?latitude=${lat}&longitude=${lon}&current=temperature_2m,relative_humidity_2m,apparent_temperature,is_day,weather_code,surface_pressure,wind_speed_10m,wind_direction_10m&daily=weather_code,temperature_2m_max,temperature_2m_min,apparent_temperature_max,apparent_temperature_min,uv_index_max,precipitation_probability_max,sunrise,sunset&timezone=auto`

      const response = await fetch(url, { signal: controller.signal })
      if (!response.ok) {
        throw new Error(`Weather API error: ${response.statusText}`)
      }

      const { current, daily } = await response.json()

      const isNightNow =
        typeof current.is_day === 'number'
          ? current.is_day === 0
          : (() => {
              const now = new Date()
              if (daily?.sunrise?.[0] && daily?.sunset?.[0]) {
                return (
                  now < new Date(daily.sunrise[0]) ||
                  now >= new Date(daily.sunset[0])
                )
              }
              const hr = now.getHours()
              return hr < 6 || hr >= 18
            })()

      const parsedCondition = parseWmoCode(current.weather_code, isNightNow)
      const forecastList: ForecastDay[] = []

      if (daily?.time) {
        const count = Math.min(daily.time.length, 7)
        for (let i = 0; i < count; i++) {
          const dateObj = new Date(`${daily.time[i]}T00:00:00`)
          const wCode = daily.weather_code[i]
          forecastList.push({
            day: i === 0 ? 'Today' : DAY_NAMES[dateObj.getDay()] || 'Day',
            tempMax: Math.round(daily.temperature_2m_max[i]),
            tempMin: Math.round(daily.temperature_2m_min[i]),
            weatherCode: wCode,
            icon: parseWmoCode(wCode).icon
          })
        }
      }

      const windDeg = current.wind_direction_10m ?? 0
      const windIdx = Math.round((((windDeg % 360) + 360) % 360) / 45) % 8
      const baseTemp = Math.round(current.temperature_2m)

      const freshData: WeatherData = {
        location: locationName,
        temp: baseTemp,
        feelsLike:
          typeof current.apparent_temperature === 'number'
            ? Math.round(current.apparent_temperature)
            : baseTemp,
        condition: parsedCondition.condition,
        high: daily?.temperature_2m_max
          ? Math.round(daily.temperature_2m_max[0])
          : baseTemp,
        low: daily?.temperature_2m_min
          ? Math.round(daily.temperature_2m_min[0])
          : baseTemp,
        humidity: Math.round(current.relative_humidity_2m),
        windSpeed: Math.round(current.wind_speed_10m),
        windDirection: windDeg,
        windDirectionLabel: WIND_DIRS[windIdx] || 'N',
        pressure:
          typeof current.surface_pressure === 'number'
            ? Math.round(current.surface_pressure)
            : 1013,
        precipitationProbability: daily?.precipitation_probability_max
          ? Math.round(daily.precipitation_probability_max[0])
          : 0,
        uvIndex: String(
          daily?.uv_index_max ? Math.round(daily.uv_index_max[0]) : 3
        ),
        weatherCode: current.weather_code,
        forecast: forecastList,
        sunrise: daily?.sunrise?.[0],
        sunset: daily?.sunset?.[0],
        isNight: isNightNow,
        isCached: false,
        isOffline: false,
        lastUpdated: Date.now()
      }

      memoryCache.set(cacheKey, freshData)
      try {
        localStorage.setItem(cacheKey, JSON.stringify(freshData))
      } catch {
        // Ignore storage quota errors
      }

      return freshData
    } catch {
      const stale = getCachedWeatherData(lat, lon, true)
      return stale ? { ...stale, isCached: true, isOffline: true } : null
    } finally {
      clearTimeout(timeoutId)
      inflightRequests.delete(cacheKey)
    }
  })()

  inflightRequests.set(cacheKey, requestPromise)
  return requestPromise
}

/**
 * Formats a geocoding entry into "City, Region, Country".
 */
export function formatLocationName(item: {
  name: string
  country?: string
  admin1?: string
}): string {
  const parts = [item.name]
  if (item.admin1 && item.admin1 !== item.name) parts.push(item.admin1)
  if (item.country && item.country !== item.name) parts.push(item.country)
  return parts.join(', ')
}

let searchAbortController: AbortController | null = null

/**
 * Queries Open-Meteo geocoding API for matching city names, aborting stale queries.
 */
export async function searchCities(query: string): Promise<GeocodingResult[]> {
  const trimmed = query?.trim()
  if (!trimmed || trimmed.length < 2) return []

  if (searchAbortController) searchAbortController.abort()
  searchAbortController = new AbortController()

  try {
    const url = `https://geocoding-api.open-meteo.com/v1/search?name=${encodeURIComponent(trimmed)}&count=10&language=en&format=json`
    const res = await fetch(url, { signal: searchAbortController.signal })
    if (!res.ok) return []
    const data = await res.json()
    if (!data.results) return []

    return data.results.map(
      ({
        name,
        country,
        admin1,
        latitude,
        longitude,
        timezone
      }: GeocodingResult) => ({
        name,
        country,
        admin1,
        latitude,
        longitude,
        timezone
      })
    )
  } catch {
    return []
  }
}

/**
 * Converts Celsius to Fahrenheit or Kelvin.
 */
export function convertTemp(
  celsius: number,
  unit: 'celsius' | 'fahrenheit' | 'kelvin'
): TempFormatted {
  if (unit === 'fahrenheit') {
    const val = Math.round((celsius * 9) / 5 + 32)
    return { val, symbol: '°F', str: `${val}°F` }
  }
  if (unit === 'kelvin') {
    const val = Math.round(celsius + 273.15)
    return { val, symbol: 'K', str: `${val}K` }
  }
  return { val: celsius, symbol: '°C', str: `${celsius}°C` }
}
