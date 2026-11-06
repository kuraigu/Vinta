/**
 * Comprehensive global weather presets and timezone anchor coordinates.
 */

export interface WeatherLocationPreset {
  name: string
  lat: number
  lon: number
}

/**
 * Curated list of major world cities and metropolitan hubs across all continents.
 */
export const GLOBAL_WEATHER_PRESETS: WeatherLocationPreset[] = [
  { name: 'Auto-detect (System Local)', lat: 14.5995, lon: 120.9842 },

  // Southeast Asia & Pacific
  { name: 'Manila, Philippines', lat: 14.5995, lon: 120.9842 },
  { name: 'Quezon City, Philippines', lat: 14.676, lon: 121.0437 },
  { name: 'Cebu City, Philippines', lat: 10.3157, lon: 123.8854 },
  { name: 'Davao City, Philippines', lat: 7.1907, lon: 125.4553 },
  { name: 'Baguio, Philippines', lat: 16.4023, lon: 120.596 },
  { name: 'Singapore', lat: 1.3521, lon: 103.8198 },
  { name: 'Bangkok, Thailand', lat: 13.7563, lon: 100.5018 },
  { name: 'Kuala Lumpur, Malaysia', lat: 3.139, lon: 101.6869 },
  { name: 'Jakarta, Indonesia', lat: -6.2088, lon: 106.8456 },
  { name: 'Hanoi, Vietnam', lat: 21.0285, lon: 105.8542 },
  { name: 'Ho Chi Minh City, Vietnam', lat: 10.8231, lon: 106.6297 },
  { name: 'Sydney, Australia', lat: -33.8688, lon: 151.2093 },
  { name: 'Melbourne, Australia', lat: -37.8136, lon: 144.9631 },
  { name: 'Brisbane, Australia', lat: -27.4698, lon: 153.0251 },
  { name: 'Auckland, New Zealand', lat: -36.8485, lon: 174.7633 },

  // East Asia
  { name: 'Tokyo, Japan', lat: 35.6762, lon: 139.6503 },
  { name: 'Osaka, Japan', lat: 34.6937, lon: 135.5023 },
  { name: 'Seoul, South Korea', lat: 37.5665, lon: 126.978 },
  { name: 'Beijing, China', lat: 39.9042, lon: 116.4074 },
  { name: 'Shanghai, China', lat: 31.2304, lon: 121.4737 },
  { name: 'Guangzhou, China', lat: 23.1291, lon: 113.2644 },
  { name: 'Hong Kong', lat: 22.3193, lon: 114.1694 },
  { name: 'Taipei, Taiwan', lat: 25.033, lon: 121.5654 },

  // South & Central Asia
  { name: 'New Delhi, India', lat: 28.6139, lon: 77.209 },
  { name: 'Mumbai, India', lat: 19.076, lon: 72.8777 },
  { name: 'Bengaluru, India', lat: 12.9716, lon: 77.5946 },
  { name: 'Dhaka, Bangladesh', lat: 23.8103, lon: 90.4125 },
  { name: 'Karachi, Pakistan', lat: 24.8607, lon: 67.0011 },
  { name: 'Colombo, Sri Lanka', lat: 6.9271, lon: 79.8612 },
  { name: 'Almaty, Kazakhstan', lat: 43.222, lon: 76.8512 },

  // Western & Northern Europe
  { name: 'London, United Kingdom', lat: 51.5074, lon: -0.1278 },
  { name: 'Manchester, United Kingdom', lat: 53.4808, lon: -2.2426 },
  { name: 'Edinburgh, United Kingdom', lat: 55.9533, lon: -3.1883 },
  { name: 'Paris, France', lat: 48.8566, lon: 2.3522 },
  { name: 'Berlin, Germany', lat: 52.52, lon: 13.405 },
  { name: 'Munich, Germany', lat: 48.1351, lon: 11.582 },
  { name: 'Frankfurt, Germany', lat: 50.1109, lon: 8.6821 },
  { name: 'Amsterdam, Netherlands', lat: 52.3676, lon: 4.9041 },
  { name: 'Brussels, Belgium', lat: 50.8503, lon: 4.3517 },
  { name: 'Dublin, Ireland', lat: 53.3498, lon: -6.2603 },
  { name: 'Zurich, Switzerland', lat: 47.3769, lon: 8.5417 },
  { name: 'Geneva, Switzerland', lat: 46.2044, lon: 6.1432 },
  { name: 'Vienna, Austria', lat: 48.2082, lon: 16.3738 },
  { name: 'Stockholm, Sweden', lat: 59.3293, lon: 18.0686 },
  { name: 'Oslo, Norway', lat: 59.9139, lon: 10.7522 },
  { name: 'Copenhagen, Denmark', lat: 55.6761, lon: 12.5683 },
  { name: 'Helsinki, Finland', lat: 60.1699, lon: 24.9384 },

  // Southern Europe
  { name: 'Madrid, Spain', lat: 40.4168, lon: -3.7038 },
  { name: 'Barcelona, Spain', lat: 41.3879, lon: 2.1699 },
  { name: 'Rome, Italy', lat: 41.9028, lon: 12.4964 },
  { name: 'Milan, Italy', lat: 45.4642, lon: 9.19 },
  { name: 'Lisbon, Portugal', lat: 38.7223, lon: -9.1393 },
  { name: 'Athens, Greece', lat: 37.9838, lon: 23.7275 },

  // Central & Eastern Europe
  { name: 'Warsaw, Poland', lat: 52.2297, lon: 21.0122 },
  { name: 'Prague, Czech Republic', lat: 50.0755, lon: 14.4378 },
  { name: 'Budapest, Hungary', lat: 47.4979, lon: 19.0402 },
  { name: 'Bucharest, Romania', lat: 44.4268, lon: 26.1025 },
  { name: 'Kyiv, Ukraine', lat: 50.4501, lon: 30.5234 },
  { name: 'Istanbul, Turkey', lat: 41.0082, lon: 28.9784 },

  // North America
  { name: 'New York, United States', lat: 40.7128, lon: -74.006 },
  { name: 'Los Angeles, United States', lat: 34.0522, lon: -118.2437 },
  { name: 'Chicago, United States', lat: 41.8781, lon: -87.6298 },
  { name: 'Houston, United States', lat: 29.7604, lon: -95.3698 },
  { name: 'Phoenix, United States', lat: 33.4484, lon: -112.074 },
  { name: 'Philadelphia, United States', lat: 39.9526, lon: -75.1652 },
  { name: 'San Francisco, United States', lat: 37.7749, lon: -122.4194 },
  { name: 'Seattle, United States', lat: 47.6062, lon: -122.3321 },
  { name: 'Miami, United States', lat: 25.7617, lon: -80.1918 },
  { name: 'Boston, United States', lat: 42.3601, lon: -71.0589 },
  { name: 'Atlanta, United States', lat: 33.749, lon: -84.388 },
  { name: 'Denver, United States', lat: 39.7392, lon: -104.9903 },
  { name: 'Toronto, Canada', lat: 43.6532, lon: -79.3832 },
  { name: 'Vancouver, Canada', lat: 49.2827, lon: -123.1207 },
  { name: 'Montreal, Canada', lat: 45.5017, lon: -73.5673 },
  { name: 'Calgary, Canada', lat: 51.0447, lon: -114.0719 },
  { name: 'Mexico City, Mexico', lat: 19.4326, lon: -99.1332 },
  { name: 'Guadalajara, Mexico', lat: 20.6597, lon: -103.3496 },

  // Latin America
  { name: 'São Paulo, Brazil', lat: -23.5505, lon: -46.6333 },
  { name: 'Rio de Janeiro, Brazil', lat: -22.9068, lon: -43.1729 },
  { name: 'Buenos Aires, Argentina', lat: -34.6037, lon: -58.3816 },
  { name: 'Santiago, Chile', lat: -33.4489, lon: -70.6693 },
  { name: 'Bogota, Colombia', lat: 4.711, lon: -74.0721 },
  { name: 'Lima, Peru', lat: -12.0464, lon: -77.0428 },

  // Middle East & Africa
  { name: 'Dubai, United Arab Emirates', lat: 25.2048, lon: 55.2708 },
  { name: 'Abu Dhabi, United Arab Emirates', lat: 24.4539, lon: 54.3773 },
  { name: 'Riyadh, Saudi Arabia', lat: 24.7136, lon: 46.6753 },
  { name: 'Doha, Qatar', lat: 25.2854, lon: 51.531 },
  { name: 'Tel Aviv, Israel', lat: 32.0853, lon: 34.7818 },
  { name: 'Cairo, Egypt', lat: 30.0444, lon: 31.2357 },
  { name: 'Johannesburg, South Africa', lat: -26.2041, lon: 28.0473 },
  { name: 'Cape Town, South Africa', lat: -33.9249, lon: 18.4241 },
  { name: 'Nairobi, Kenya', lat: -1.2921, lon: 36.8219 },
  { name: 'Lagos, Nigeria', lat: 6.5244, lon: 3.3792 },
  { name: 'Casablanca, Morocco', lat: 33.5731, lon: -7.5898 }
]

/**
 * Maps browser IANA timezone identifiers to approximate GPS coordinates.
 *
 * Browsers only provide a timezone name string (e.g. "Asia/Tokyo") without latitude or
 * longitude. This lookup table translates that string into coordinates so weather can
 * be fetched without network calls.
 */
const TIMEZONE_TO_COORDINATES_MAP: Record<string, WeatherLocationPreset> = {
  'Asia/Manila': { name: 'Manila, Philippines', lat: 14.5995, lon: 120.9842 },
  'Asia/Tokyo': { name: 'Tokyo, Japan', lat: 35.6762, lon: 139.6503 },
  'Asia/Seoul': { name: 'Seoul, South Korea', lat: 37.5665, lon: 126.978 },
  'Asia/Singapore': { name: 'Singapore', lat: 1.3521, lon: 103.8198 },
  'Asia/Hong_Kong': { name: 'Hong Kong', lat: 22.3193, lon: 114.1694 },
  'Asia/Shanghai': { name: 'Shanghai, China', lat: 31.2304, lon: 121.4737 },
  'Asia/Taipei': { name: 'Taipei, Taiwan', lat: 25.033, lon: 121.5654 },
  'Asia/Bangkok': { name: 'Bangkok, Thailand', lat: 13.7563, lon: 100.5018 },
  'Asia/Jakarta': { name: 'Jakarta, Indonesia', lat: -6.2088, lon: 106.8456 },
  'Asia/Kuala_Lumpur': {
    name: 'Kuala Lumpur, Malaysia',
    lat: 3.139,
    lon: 101.6869
  },
  'Asia/Kolkata': { name: 'New Delhi, India', lat: 28.6139, lon: 77.209 },
  'Asia/Dubai': { name: 'Dubai, UAE', lat: 25.2048, lon: 55.2708 },
  'Europe/London': {
    name: 'London, United Kingdom',
    lat: 51.5074,
    lon: -0.1278
  },
  'Europe/Berlin': { name: 'Berlin, Germany', lat: 52.52, lon: 13.405 },
  'Europe/Paris': { name: 'Paris, France', lat: 48.8566, lon: 2.3522 },
  'Europe/Rome': { name: 'Rome, Italy', lat: 41.9028, lon: 12.4964 },
  'Europe/Madrid': { name: 'Madrid, Spain', lat: 40.4168, lon: -3.7038 },
  'Europe/Amsterdam': {
    name: 'Amsterdam, Netherlands',
    lat: 52.3676,
    lon: 4.9041
  },
  'Europe/Brussels': { name: 'Brussels, Belgium', lat: 50.8503, lon: 4.3517 },
  'Europe/Zurich': { name: 'Zurich, Switzerland', lat: 47.3769, lon: 8.5417 },
  'Europe/Vienna': { name: 'Vienna, Austria', lat: 48.2082, lon: 16.3738 },
  'Europe/Warsaw': { name: 'Warsaw, Poland', lat: 52.2297, lon: 21.0122 },
  'Europe/Prague': {
    name: 'Prague, Czech Republic',
    lat: 50.0755,
    lon: 14.4378
  },
  'Europe/Stockholm': { name: 'Stockholm, Sweden', lat: 59.3293, lon: 18.0686 },
  'Europe/Athens': { name: 'Athens, Greece', lat: 37.9838, lon: 23.7275 },
  'Europe/Kyiv': { name: 'Kyiv, Ukraine', lat: 50.4501, lon: 30.5234 },
  'Europe/Istanbul': { name: 'Istanbul, Turkey', lat: 41.0082, lon: 28.9784 },
  'America/New_York': {
    name: 'New York, United States',
    lat: 40.7128,
    lon: -74.006
  },
  'America/Chicago': {
    name: 'Chicago, United States',
    lat: 41.8781,
    lon: -87.6298
  },
  'America/Los_Angeles': {
    name: 'Los Angeles, United States',
    lat: 34.0522,
    lon: -118.2437
  },
  'America/Denver': {
    name: 'Denver, United States',
    lat: 39.7392,
    lon: -104.9903
  },
  'America/Phoenix': {
    name: 'Phoenix, United States',
    lat: 33.4484,
    lon: -112.074
  },
  'America/Toronto': { name: 'Toronto, Canada', lat: 43.6532, lon: -79.3832 },
  'America/Vancouver': {
    name: 'Vancouver, Canada',
    lat: 49.2827,
    lon: -123.1207
  },
  'America/Mexico_City': {
    name: 'Mexico City, Mexico',
    lat: 19.4326,
    lon: -99.1332
  },
  'America/Sao_Paulo': {
    name: 'São Paulo, Brazil',
    lat: -23.5505,
    lon: -46.6333
  },
  'America/Buenos_Aires': {
    name: 'Buenos Aires, Argentina',
    lat: -34.6037,
    lon: -58.3816
  },
  'America/Santiago': { name: 'Santiago, Chile', lat: -33.4489, lon: -70.6693 },
  'America/Bogota': { name: 'Bogota, Colombia', lat: 4.711, lon: -74.0721 },
  'Australia/Sydney': {
    name: 'Sydney, Australia',
    lat: -33.8688,
    lon: 151.2093
  },
  'Australia/Melbourne': {
    name: 'Melbourne, Australia',
    lat: -37.8136,
    lon: 144.9631
  },
  'Pacific/Auckland': {
    name: 'Auckland, New Zealand',
    lat: -36.8485,
    lon: 174.7633
  },
  'Africa/Cairo': { name: 'Cairo, Egypt', lat: 30.0444, lon: 31.2357 },
  'Africa/Johannesburg': {
    name: 'Johannesburg, South Africa',
    lat: -26.2041,
    lon: 28.0473
  }
}

/**
 * Resolves local coordinates and city from the user's system timezone.
 *
 * @param customTz - Optional timezone identifier. Defaults to system timezone.
 * @returns Resolved anchor location.
 */
export function resolveLocationFromTimezone(
  customTz?: string
): WeatherLocationPreset {
  let tz = customTz
  if (!tz && typeof Intl !== 'undefined' && Intl.DateTimeFormat) {
    try {
      tz = Intl.DateTimeFormat().resolvedOptions().timeZone
    } catch {
      // Ignore
    }
  }

  if (tz && TIMEZONE_TO_COORDINATES_MAP[tz]) {
    return TIMEZONE_TO_COORDINATES_MAP[tz]
  }

  // Fallback search by city name in timezone identifier
  if (tz && tz.includes('/')) {
    const rawCity = tz.split('/')[1]?.replace(/_/g, ' ').toLowerCase()
    if (rawCity) {
      const match = GLOBAL_WEATHER_PRESETS.find((p) =>
        p.name.toLowerCase().includes(rawCity)
      )
      if (match) return match
    }
  }

  return { name: 'Manila, Philippines', lat: 14.5995, lon: 120.9842 }
}

/**
 * Multi-provider IP and timezone auto-detection.
 *
 * @returns Resolved location with coordinates.
 */
export async function detectAutoWeatherLocation(): Promise<WeatherLocationPreset> {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 3000)

  // Primary: FreeIPAPI
  try {
    const res = await fetch('https://freeipapi.com/api/json', {
      signal: controller.signal
    })
    clearTimeout(timeoutId)
    if (res.ok) {
      const data = await res.json()
      if (
        data.cityName &&
        typeof data.latitude === 'number' &&
        typeof data.longitude === 'number'
      ) {
        const country = data.countryName || ''
        const name = country ? `${data.cityName}, ${country}` : data.cityName
        return { name, lat: data.latitude, lon: data.longitude }
      }
    }
  } catch {
    // Continue to secondary
  }

  // Secondary: ipapi.co
  try {
    const secondaryController = new AbortController()
    const secondaryTimeout = setTimeout(() => secondaryController.abort(), 2500)
    const res = await fetch('https://ipapi.co/json/', {
      signal: secondaryController.signal
    })
    clearTimeout(secondaryTimeout)
    if (res.ok) {
      const data = await res.json()
      if (data.city && data.latitude && data.longitude) {
        const country = data.country_name || ''
        const name = country ? `${data.city}, ${country}` : data.city
        return { name, lat: data.latitude, lon: data.longitude }
      }
    }
  } catch {
    // Continue to timezone anchor
  }

  // Tertiary fallback: Instant zero-latency system timezone anchor
  return resolveLocationFromTimezone()
}
