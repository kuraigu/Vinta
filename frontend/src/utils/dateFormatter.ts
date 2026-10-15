/**
 * Date formatting options supported across Vinta widgets and pages.
 */
export type DateFormatOption =
  | 'short'
  | 'full'
  | 'us_numeric'
  | 'eu_numeric'
  | 'us_long'
  | 'eu_long'
  | 'iso'
  | 'default'
  | 'us'
  | 'eu'

/**
 * Normalizes legacy or shorthand date format options to modern keys.
 *
 * @param format Stored format key.
 * @returns Standardized DateFormatOption.
 */
export function normalizeDateFormat(format?: string | null): DateFormatOption {
  if (!format || format === 'default') return 'short'
  if (format === 'us') return 'us_numeric'
  if (format === 'eu') return 'eu_numeric'
  return format as DateFormatOption
}

/**
 * Formats a Date object or ISO date string based on the user's selected format preference.
 *
 * @param input Target date instance or YYYY-MM-DD date key.
 * @param format Chosen date format option.
 * @param options Optional configuration (e.g. includeWeekday).
 * @returns Formatted human-readable date string.
 */
export function formatUserDate(
  input: Date | string,
  format: DateFormatOption = 'short',
  options?: { includeWeekday?: boolean }
): string {
  const normFormat = normalizeDateFormat(format)
  let date: Date

  if (typeof input === 'string') {
    if (input.includes('T')) {
      date = new Date(input)
    } else {
      const parts = input.split('-').map(Number)
      if (
        parts.length === 3 &&
        !isNaN(parts[0]) &&
        !isNaN(parts[1]) &&
        !isNaN(parts[2])
      ) {
        date = new Date(parts[0], parts[1] - 1, parts[2])
      } else {
        date = new Date(input)
      }
    }
  } else {
    date = input
  }

  if (isNaN(date.getTime())) {
    return typeof input === 'string' ? input : ''
  }

  const year = date.getFullYear()
  const month = date.getMonth() + 1
  const day = date.getDate()
  const mm = String(month).padStart(2, '0')
  const dd = String(day).padStart(2, '0')

  const monthNames = [
    'January',
    'February',
    'March',
    'April',
    'May',
    'June',
    'July',
    'August',
    'September',
    'October',
    'November',
    'December'
  ]
  const monthShort = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'May',
    'Jun',
    'Jul',
    'Aug',
    'Sep',
    'Oct',
    'Nov',
    'Dec'
  ]
  const weekdayNames = [
    'Sunday',
    'Monday',
    'Tuesday',
    'Wednesday',
    'Thursday',
    'Friday',
    'Saturday'
  ]
  const weekdayShort = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

  const wkFull = weekdayNames[date.getDay()]
  const wkShort = weekdayShort[date.getDay()]
  const mFull = monthNames[date.getMonth()]
  const mShort = monthShort[date.getMonth()]

  switch (normFormat) {
    case 'us_numeric':
      return options?.includeWeekday
        ? `${wkShort}, ${mm}/${dd}/${year}`
        : `${mm}/${dd}/${year}`
    case 'eu_numeric':
      return options?.includeWeekday
        ? `${wkShort}, ${dd}/${mm}/${year}`
        : `${dd}/${mm}/${year}`
    case 'iso':
      return options?.includeWeekday
        ? `${wkShort}, ${year}-${mm}-${dd}`
        : `${year}-${mm}-${dd}`
    case 'us_long':
      return options?.includeWeekday
        ? `${wkFull}, ${mFull} ${day}, ${year}`
        : `${mFull} ${day}, ${year}`
    case 'eu_long':
      return options?.includeWeekday
        ? `${wkFull}, ${day} ${mFull} ${year}`
        : `${day} ${mFull} ${year}`
    case 'full':
      return `${wkFull}, ${mFull} ${day}`
    case 'short':
    default:
      return `${wkShort}, ${mShort} ${day}`
  }
}
