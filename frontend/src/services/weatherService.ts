import {
  Sun,
  CloudSun,
  Cloud,
  CloudRain,
  CloudSnow,
  CloudLightning
} from '@lucide/svelte'

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
  condition: string
  high: number
  low: number
  humidity: number
  windSpeed: number
  uvIndex: string
  weatherCode: number
  forecast: ForecastDay[]
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

export function parseWmoCode(code: number): {
  condition: string
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  icon: any
  theme: WeatherTheme
} {
  if (code === 0) {
    return {
      condition: 'Clear Sky',
      icon: Sun,
      theme: {
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
      icon: CloudSun,
      theme: {
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

    const url = `https://api.open-meteo.com/v1/forecast?latitude=${lat}&longitude=${lon}&current=temperature_2m,relative_humidity_2m,weather_code,wind_speed_10m&daily=weather_code,temperature_2m_max,temperature_2m_min,uv_index_max&timezone=auto`

    const response = await fetch(url, { signal: controller.signal })
    clearTimeout(timeoutId)

    if (!response.ok) {
      throw new Error(`Weather API error: ${response.statusText}`)
    }

    const data = await response.json()
    const current = data.current
    const daily = data.daily

    const parsedCondition = parseWmoCode(current.weather_code)

    const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
    const forecastList: ForecastDay[] = []

    if (daily && daily.time) {
      for (let i = 0; i < Math.min(daily.time.length, 4); i++) {
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
    const uvLabel =
      uvMax <= 2
        ? `${uvMax} (Low)`
        : uvMax <= 5
          ? `${uvMax} (Mod)`
          : `${uvMax} (High)`

    const freshData: WeatherData = {
      location: locationName,
      temp: Math.round(current.temperature_2m),
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
      uvIndex: uvLabel,
      weatherCode: current.weather_code,
      forecast: forecastList,
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

export async function searchCities(query: string): Promise<GeocodingResult[]> {
  if (!query || query.trim().length < 2) return []

  try {
    const url = `https://geocoding-api.open-meteo.com/v1/search?name=${encodeURIComponent(query.trim())}&count=5&language=en&format=json`
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
