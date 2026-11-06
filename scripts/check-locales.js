/**
 * Vinta Locale Parity & Validation Checker
 *
 * Compares all locale catalogs in frontend/src/locales against the base
 * en.json catalog to verify:
 *  - Missing keys (present in en.json but absent in target locale)
 *  - Extra / obsolete keys (present in target locale but absent in en.json)
 *  - Structural type mismatches (object vs primitive)
 *  - Empty translation values
 *
 * Exit code: 0 on success, 1 on discrepancies.
 */

const fs = require('fs')
const path = require('path')

const LOCALES_DIR = path.resolve(__dirname, '../frontend/src/locales')
const BASE_LOCALE = 'en.json'

function flattenKeys(obj, prefix = '') {
  const result = new Map()
  for (const [key, val] of Object.entries(obj)) {
    const fullPath = prefix ? `${prefix}.${key}` : key
    if (val !== null && typeof val === 'object' && !Array.isArray(val)) {
      const nested = flattenKeys(val, fullPath)
      for (const [nestedKey, nestedVal] of nested) {
        result.set(nestedKey, nestedVal)
      }
    } else {
      result.set(fullPath, val)
    }
  }
  return result
}

function run() {
  if (!fs.existsSync(LOCALES_DIR)) {
    console.error(`Error: Locales directory not found at: ${LOCALES_DIR}`)
    process.exit(1)
  }

  const basePath = path.join(LOCALES_DIR, BASE_LOCALE)
  if (!fs.existsSync(basePath)) {
    console.error(`Error: Base locale file not found at: ${basePath}`)
    process.exit(1)
  }

  let baseData
  try {
    baseData = JSON.parse(fs.readFileSync(basePath, 'utf8'))
  } catch (err) {
    console.error(`Error reading base locale ${BASE_LOCALE}:`, err.message)
    process.exit(1)
  }

  const baseMap = flattenKeys(baseData)
  const files = fs
    .readdirSync(LOCALES_DIR)
    .filter((f) => f.endsWith('.json') && f !== BASE_LOCALE)

  console.log(`\n🔍 Checking Vinta Locale Catalogs against base [${BASE_LOCALE}] (${baseMap.size} keys)...\n`)

  let hasErrors = false

  for (const file of files) {
    const filePath = path.join(LOCALES_DIR, file)
    let localeData
    try {
      localeData = JSON.parse(fs.readFileSync(filePath, 'utf8'))
    } catch (err) {
      console.error(`✖ ${file}: Failed to parse JSON - ${err.message}`)
      hasErrors = true
      continue
    }

    const localeMap = flattenKeys(localeData)
    const missingKeys = []
    const emptyKeys = []
    const extraKeys = []

    for (const [key, baseVal] of baseMap) {
      if (!localeMap.has(key)) {
        missingKeys.push(key)
      } else {
        const val = localeMap.get(key)
        if (typeof val === 'string' && val.trim() === '') {
          emptyKeys.push(key)
        }
      }
    }

    for (const [key] of localeMap) {
      if (!baseMap.has(key)) {
        extraKeys.push(key)
      }
    }

    const fileErrors = missingKeys.length + emptyKeys.length + extraKeys.length

    if (fileErrors === 0) {
      console.log(`✔ ${file} - 100% in sync (${localeMap.size} keys)`)
    } else {
      hasErrors = true
      console.log(`✖ ${file} - ${fileErrors} issue(s) found:`)

      if (missingKeys.length > 0) {
        console.log(`   Missing keys (${missingKeys.length}):`)
        missingKeys.slice(0, 15).forEach((k) => console.log(`     - ${k}`))
        if (missingKeys.length > 15) {
          console.log(`     ... and ${missingKeys.length - 15} more`)
        }
      }

      if (emptyKeys.length > 0) {
        console.log(`   Empty translation values (${emptyKeys.length}):`)
        emptyKeys.slice(0, 10).forEach((k) => console.log(`     - ${k}`))
        if (emptyKeys.length > 10) {
          console.log(`     ... and ${emptyKeys.length - 10} more`)
        }
      }

      if (extraKeys.length > 0) {
        console.log(`   Extra / obsolete keys (${extraKeys.length}):`)
        extraKeys.slice(0, 10).forEach((k) => console.log(`     + ${k}`))
        if (extraKeys.length > 10) {
          console.log(`     ... and ${extraKeys.length - 10} more`)
        }
      }
      console.log('')
    }
  }

  if (hasErrors) {
    console.log(`\n❌ Locale check failed. Please resolve discrepancies before committing.\n`)
    process.exit(1)
  } else {
    console.log(`\n✨ All locale catalogs are synchronized with zero discrepancies.\n`)
    process.exit(0)
  }
}

run()

