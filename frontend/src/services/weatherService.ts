/**
 * @fileoverview Weather data fetching and Open-Meteo integration.
 *
 * Downloads forecasts, humidity, wind, and UV data, caches responses for 15 minutes,
 * translates weather codes to condition titles and icons, and handles city search.
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

/**
 * Resolves the i18n translation key corresponding to a WMO weather code or string.
 *
 * @param code - Numeric WMO code.
 * @param isNight - Whether current local time is nighttime.
 * @param fallbackCondition - Optional literal condition string.
 * @returns Translation catalog key for `$tStore()`.
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
 * Maps a WMO weather code to a condition name, Lucide icon, and color theme.
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
      theme: isNight
        ? {
            color: '#818cf8',
            glow: 'rgba(129, 140, 248, 0.25)',
            badgeBg: 'rgba(129, 140, 248, 0.15)'
          }
        : {
            color: '#f59e0b',
            glow: 'rgba(245, 158, 11, 0.22)',
            badgeBg: 'rgba(245, 158, 11, 0.15)'
          }
    }
  }
  if (code >= 1 && code <= 3) {
    return {
      condition:
        code === 1 ? 'Mainly Clear' : code === 2 ? 'Partly Cloudy' : 'Overcast',
      icon: isNight ? CloudMoon : CloudSun,
      theme: isNight
        ? {
            color: '#818cf8',
            glow: 'rgba(129, 140, 248, 0.25)',
            badgeBg: 'rgba(129, 140, 248, 0.15)'
          }
        : {
            color: '#38bdf8',
            glow: 'rgba(56, 189, 248, 0.2)',
            badgeBg: 'rgba(56, 189, 248, 0.15)'
          }
    }
  }
  if (code === 45 || code === 48) {
    return {
      condition: 'Foggy',
      icon: Cloud,
      theme: {
        color: '#94a3b8',
        glow: 'rgba(148, 163, 184, 0.2)',
        badgeBg: 'rgba(148, 163, 184, 0.15)'
      }
    }
  }
  if ((code >= 51 && code <= 67) || (code >= 80 && code <= 82)) {
    return {
      condition: 'Rainy',
      icon: CloudRain,
      theme: {
        color: '#3b82f6',
        glow: 'rgba(59, 130, 246, 0.25)',
        badgeBg: 'rgba(59, 130, 246, 0.15)'
      }
    }
  }
  if (code >= 71 && code <= 77) {
    return {
      condition: 'Snowy',
      icon: CloudSnow,
      theme: {
        color: '#06b6d4',
        glow: 'rgba(6, 182, 212, 0.25)',
        badgeBg: 'rgba(6, 182, 212, 0.15)'
      }
    }
  }
  if (code >= 95 && code <= 99) {
    return {
      condition: 'Thunderstorm',
      icon: CloudLightning,
      theme: {
        color: '#a855f7',
        glow: 'rgba(168, 85, 247, 0.3)',
        badgeBg: 'rgba(168, 85, 247, 0.18)'
      }
    }
  }
  return {
    condition: 'Partly Cloudy',
    icon: CloudSun,
    theme: {
      color: '#38bdf8',
      glow: 'rgba(56, 189, 248, 0.2)',
      badgeBg: 'rgba(56, 189, 248, 0.15)'
    }
  }
}

const CACHE_DURATION_MS = 15 * 60 * 1000 // 15 minutes

/**
 * Fetches current weather and 7-day forecast from Open-Meteo, using cache when fresh.
 */
export async function fetchWeatherData(
  locationName: string,
  lat: number,
  lon: number,
  simulateOffline = false
): Promise<WeatherData | null> {
  const cacheKey = `vinta_weather_cache_${lat.toFixed(4)}_${lon.toFixed(4)}`
  const cachedStr = localStorage.getItem(cacheKey)

  if (cachedStr) {
    try {
      const cachedData: WeatherData = JSON.parse(cachedStr)
      const isExpired = Date.now() - cachedData.lastUpdated > CACHE_DURATION_MS

      if (!isExpired && !simulateOffline) {
        return { ...cachedData, isCached: true, isOffline: false }
      }
    } catch {
      localStorage.removeItem(cacheKey)
    }
  }

  if (simulateOffline) {
    if (cachedStr) {
      try {
        const cachedData: WeatherData = JSON.parse(cachedStr)
        return { ...cachedData, isCached: true, isOffline: true }
      } catch {
        return null
      }
    }
    return null
  }

  try {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 6000)

    const url = `https://api.open-meteo.com/v1/forecast?latitude=${lat}&longitude=${lon}&current=temperature_2m,relative_humidity_2m,apparent_temperature,is_day,weather_code,surface_pressure,wind_speed_10m,wind_direction_10m&daily=weather_code,temperature_2m_max,temperature_2m_min,apparent_temperature_max,apparent_temperature_min,uv_index_max,precipitation_probability_max,sunrise,sunset&timezone=auto`

    const response = await fetch(url, { signal: controller.signal })
    clearTimeout(timeoutId)

    if (!response.ok) {
      throw new Error(`Weather API error: ${response.statusText}`)
    }

    const data = await response.json()
    const current = data.current
    const daily = data.daily

    const isNightNow =
      typeof current.is_day === 'number'
        ? current.is_day === 0
        : (() => {
            const now = new Date()
            if (daily?.sunrise?.[0] && daily?.sunset?.[0]) {
              const sr = new Date(daily.sunrise[0])
              const ss = new Date(daily.sunset[0])
              return now < sr || now >= ss
            }
            const hr = now.getHours()
            return hr < 6 || hr >= 18
          })()

    const parsedCondition = parseWmoCode(current.weather_code, isNightNow)

    const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
    const forecastList: ForecastDay[] = []

    if (daily && daily.time) {
      for (let i = 0; i < Math.min(daily.time.length, 7); i++) {
        const dateObj = new Date(daily.time[i])
        const dayLabel = i === 0 ? 'Today' : dayNames[dateObj.getDay()] || 'Day'
        const wCode = daily.weather_code[i]
        const wInfo = parseWmoCode(wCode)

        forecastList.push({
          day: dayLabel,
          tempMax: Math.round(daily.temperature_2m_max[i]),
          tempMin: Math.round(daily.temperature_2m_min[i]),
          weatherCode: wCode,
          icon: wInfo.icon
        })
      }
    }

    const uvMax =
      daily && daily.uv_index_max ? Math.round(daily.uv_index_max[0]) : 3
    const uvLabel = String(uvMax)

    const windDirections = ['N', 'NE', 'E', 'SE', 'S', 'SW', 'W', 'NW']
    const windDeg = current.wind_direction_10m ?? 0
    const windIdx = Math.round((((windDeg % 360) + 360) % 360) / 45) % 8
    const windDirLabel = windDirections[windIdx] || 'N'

    const feelsLikeTemp =
      typeof current.apparent_temperature === 'number'
        ? Math.round(current.apparent_temperature)
        : Math.round(current.temperature_2m)

    const precipProb =
      daily && daily.precipitation_probability_max
        ? Math.round(daily.precipitation_probability_max[0])
        : 0

    const pressureVal =
      typeof current.surface_pressure === 'number'
        ? Math.round(current.surface_pressure)
        : 1013

    const freshData: WeatherData = {
      location: locationName,
      temp: Math.round(current.temperature_2m),
      feelsLike: feelsLikeTemp,
      condition: parsedCondition.condition,
      high:
        daily && daily.temperature_2m_max
          ? Math.round(daily.temperature_2m_max[0])
          : Math.round(current.temperature_2m),
      low:
        daily && daily.temperature_2m_min
          ? Math.round(daily.temperature_2m_min[0])
          : Math.round(current.temperature_2m),
      humidity: Math.round(current.relative_humidity_2m),
      windSpeed: Math.round(current.wind_speed_10m),
      windDirection: windDeg,
      windDirectionLabel: windDirLabel,
      pressure: pressureVal,
      precipitationProbability: precipProb,
      uvIndex: uvLabel,
      weatherCode: current.weather_code,
      forecast: forecastList,
      sunrise: daily?.sunrise?.[0],
      sunset: daily?.sunset?.[0],
      isNight: isNightNow,
      isCached: false,
      isOffline: false,
      lastUpdated: Date.now()
    }

    localStorage.setItem(cacheKey, JSON.stringify(freshData))
    return freshData
  } catch {
    if (cachedStr) {
      try {
        const cachedData: WeatherData = JSON.parse(cachedStr)
        return { ...cachedData, isCached: true, isOffline: true }
      } catch {
        return null
      }
    }
    return null
  }
}

/**
 * Formats a geocoding result into a human-readable display string with country and administrative division.
 *
 * @param item - Geocoding result.
 * @returns Disambiguated location string (e.g. "Cambridge, England, United Kingdom").
 */
export function formatLocationName(item: {
  name: string
  country?: string
  admin1?: string
}): string {
  const parts = [item.name]
  if (item.admin1 && item.admin1 !== item.name) {
    parts.push(item.admin1)
  }
  if (item.country && item.country !== item.name) {
    parts.push(item.country)
  }
  return parts.join(', ')
}

/**
 * Searches global cities by name using Open-Meteo geocoding.
 */
export async function searchCities(query: string): Promise<GeocodingResult[]> {
  if (!query || query.trim().length < 2) return []

  try {
    const url = `https://geocoding-api.open-meteo.com/v1/search?name=${encodeURIComponent(query.trim())}&count=10&language=en&format=json`
    const res = await fetch(url)
    if (!res.ok) return []
    const data = await res.json()
    if (!data.results) return []

    return data.results.map(
      (item: {
        name: string
        country?: string
        admin1?: string
        latitude: number
        longitude: number
        timezone?: string
      }) => ({
        name: item.name,
        country: item.country,
        admin1: item.admin1,
        latitude: item.latitude,
        longitude: item.longitude,
        timezone: item.timezone
      })
    )
  } catch {
    return []
  }
}

export type TempFormatted = {
  val: number
  symbol: string
  str: string
}

/**
 * Converts Celsius temperatures to Fahrenheit or Kelvin.
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
