/**
 * Vinta Locale Token Synchronizer
 *
 * Propagates new keys from en.json to all other locale catalogs in
 * frontend/src/locales/, preserving existing translations, key ordering,
 * and each file's @metadata block.
 *
 * Usage:
 *   node scripts/sync-locales.js [--prune]
 */

const fs = require('fs')
const path = require('path')
const { hasFlag, printHelpIfRequested } = require('./lib/lint-utils')

printHelpIfRequested(
  [
    'Usage: node scripts/sync-locales.js [--prune] [--help]',
    '',
    'Options:',
    '  --prune  Remove obsolete keys not present in en.json',
    '  --help   Show this help message'
  ].join('\n')
)

const LOCALES_DIR = path.resolve(__dirname, '../frontend/src/locales')
const BASE_LOCALE = 'en.json'
const METADATA_KEY = '@metadata'

function isPlainObject(val) {
  return val !== null && typeof val === 'object' && !Array.isArray(val)
}

function syncObject(baseObj, targetObj, stats) {
  const result = {}

  for (const [key, baseVal] of Object.entries(baseObj)) {
    if (key === METADATA_KEY) continue

    const hasTarget = isPlainObject(targetObj) && Object.prototype.hasOwnProperty.call(targetObj, key)
    const targetVal = hasTarget ? targetObj[key] : undefined

    if (isPlainObject(baseVal)) {
      const nextTarget = isPlainObject(targetVal) ? targetVal : {}
      result[key] = syncObject(baseVal, nextTarget, stats)
    } else if (typeof targetVal === 'string' && targetVal.trim() !== '') {
      result[key] = targetVal
    } else {
      result[key] = baseVal
      stats.added++
    }
  }

  return result
}

function collectExtraKeys(baseObj, targetObj, prefix = '') {
  const extras = []
  if (!isPlainObject(targetObj)) return extras

  for (const [key, targetVal] of Object.entries(targetObj)) {
    if (!prefix && key === METADATA_KEY) continue
    const fullPath = prefix ? `${prefix}.${key}` : key
    if (!isPlainObject(baseObj) || !Object.prototype.hasOwnProperty.call(baseObj, key)) {
      extras.push(fullPath)
    } else if (isPlainObject(baseObj[key]) && isPlainObject(targetVal)) {
      extras.push(...collectExtraKeys(baseObj[key], targetVal, fullPath))
    }
  }
  return extras
}

function preserveExtras(result, targetObj, baseObj) {
  if (!isPlainObject(targetObj)) return
  for (const [key, targetVal] of Object.entries(targetObj)) {
    if (key === METADATA_KEY) continue
    if (!Object.prototype.hasOwnProperty.call(baseObj, key)) {
      result[key] = targetVal
    } else if (isPlainObject(baseObj[key]) && isPlainObject(targetVal) && isPlainObject(result[key])) {
      preserveExtras(result[key], targetVal, baseObj[key])
    }
  }
}

function run() {
  const prune = hasFlag('--prune')

  if (!fs.existsSync(LOCALES_DIR)) {
    console.error(`Error: Locales directory not found at: ${LOCALES_DIR}`)
    process.exit(1)
  }

  const basePath = path.join(LOCALES_DIR, BASE_LOCALE)
  if (!fs.existsSync(basePath)) {
    console.error(`Error: Base locale file not found at: ${basePath}`)
    process.exit(1)
  }

  const baseData = JSON.parse(fs.readFileSync(basePath, 'utf8'))
  const files = fs
    .readdirSync(LOCALES_DIR)
    .filter((f) => f.endsWith('.json') && f !== BASE_LOCALE)
    .sort()

  console.log(`\nSyncing locale catalogs from [${BASE_LOCALE}]${prune ? ' (with --prune)' : ''}...\n`)

  for (const file of files) {
    const filePath = path.join(LOCALES_DIR, file)
    let targetData
    try {
      targetData = JSON.parse(fs.readFileSync(filePath, 'utf8'))
    } catch (err) {
      console.error(`✖ ${file}: Failed to parse JSON - ${err.message}`)
      process.exitCode = 1
      continue
    }

    const stats = { added: 0 }
    const metadata = isPlainObject(targetData[METADATA_KEY])
      ? targetData[METADATA_KEY]
      : { authors: [], 'last-updated': '' }

    const syncedBody = syncObject(baseData, targetData, stats)
    const extras = collectExtraKeys(baseData, targetData)

    if (!prune && extras.length > 0) {
      preserveExtras(syncedBody, targetData, baseData)
    }

    const output = {
      [METADATA_KEY]: metadata,
      ...syncedBody
    }

    const serialized = `${JSON.stringify(output, null, 2)}\n`
    const existing = fs.readFileSync(filePath, 'utf8')

    if (serialized !== existing) {
      fs.writeFileSync(filePath, serialized, 'utf8')
      const pruneNote = prune && extras.length > 0 ? `, pruned ${extras.length}` : ''
      console.log(`✔ ${file}: +${stats.added} key(s) synced${pruneNote}`)
    } else {
      console.log(`• ${file}: already up to date`)
    }
  }

  console.log('')
}

run()
