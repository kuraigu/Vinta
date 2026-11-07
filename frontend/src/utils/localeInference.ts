/**
 * @fileoverview Regional default inference.
 *
 * Inspects system language settings on first run to pick sensible defaults for
 * Celsius vs. Fahrenheit, 12-hour vs. 24-hour time, and date order.
 */

import type { TemperatureUnit, TimeFormatOption } from '../stores/settings'
import type { DateFormatOption } from './dateFormatter'

/**
 * List of ISO alpha-2 country/territory codes that officially or predominantly use Fahrenheit.
 */
const FAHRENHEIT_REGIONS = new Set([
  'US', // United States
  'LR', // Liberia
  'MM', // Myanmar
  'BS', // Bahamas
  'BZ', // Belize
  'KY', // Cayman Islands
  'PW' // Palau
])

/**
 * Extracts the user's ISO-3166 region code from the browser/system navigator language.
 *
 * @returns Uppercase 2-letter country code (e.g. 'US', 'GB', 'DE') or empty string if undetermined.
 */
export function getDetectedRegion(): string {
  try {
    const lang = typeof navigator !== 'undefined' ? navigator.language : ''
    if (!lang) return ''
    const locale = new Intl.Locale(lang).maximize()
    return (locale.region || '').toUpperCase()
  } catch {
    return ''
  }
}

/**
 * Infers whether the user's system locale defaults to Celsius or Fahrenheit.
 *
 * @returns 'fahrenheit' if residing in an official Fahrenheit region, otherwise 'celsius'.
 */
export function inferDefaultTemperatureUnit(): TemperatureUnit {
  const region = getDetectedRegion()
  return FAHRENHEIT_REGIONS.has(region) ? 'fahrenheit' : 'celsius'
}

/**
 * Infers the preferred time format (12-hour vs 24-hour) from the system locale.
 *
 * @returns '12h' or '24h' according to the user's regional hourCycle standard.
 */
export function inferDefaultTimeFormat(): TimeFormatOption {
  try {
    const lang = typeof navigator !== 'undefined' ? navigator.language : ''
    const formatter = new Intl.DateTimeFormat(lang, { hour: 'numeric' })
    const cycle = formatter.resolvedOptions().hourCycle
    if (cycle === 'h11' || cycle === 'h12') return '12h'
    if (cycle === 'h23' || cycle === 'h24') return '24h'

    // Fallback hour cycle test using 13:00 reference time
    const sample = formatter.format(new Date(2026, 0, 1, 13, 0))
    const is12Hour =
      sample.includes('PM') ||
      sample.includes('pm') ||
      (sample.includes('1') && !sample.includes('13'))
    return is12Hour ? '12h' : '24h'
  } catch {
    return '12h'
  }
}

/**
 * Infers the preferred date display format option from the user's regional standards.
 *
 * @returns 'us_numeric' for United States locale, otherwise 'short'.
 */
export function inferDefaultDateFormat(): DateFormatOption {
  const region = getDetectedRegion()
  if (region === 'US') {
    return 'us_numeric'
  }
  return 'short'
}
