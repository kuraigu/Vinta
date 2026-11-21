/**
 * Vinta Settings & Defaults Validator
 *
 * Audits persistentWritable() stores in frontend/src/stores/settings.ts:
 *  - Verifies unique vinta_* storage keys (fatal on duplicates)
 *  - Verifies every setting token declares a non-undefined default value
 *  - Verifies CURRENT_SETTINGS_SCHEMA_VERSION matches migrations
 *  - Audits direct localStorage usage outside settings.ts (non-fatal)
 *
 * Options:
 *  defaults / --defaults   Print every persistent setting key and its default value
 */

const fs = require('fs')
const path = require('path')
const { walkFiles, hasFlag, printHelpIfRequested } = require('./lib/lint-utils')

printHelpIfRequested(
  [
    'Usage: node scripts/check-settings.js [options]',
    '',
    'Options:',
    '  --defaults  Print every persistent setting key and its default value',
    '  --help      Show this help message'
  ].join('\n')
)

const ROOT_DIR = path.resolve(__dirname, '..')
const SETTINGS_PATH = path.join(ROOT_DIR, 'frontend/src/stores/settings.ts')
const SRC_DIR = path.join(ROOT_DIR, 'frontend/src')
const SHOW_DEFAULTS =
  hasFlag('--defaults') ||
  hasFlag('--default') ||
  process.argv.includes('defaults') ||
  process.argv.includes('default')

function walkDir(dir, exts, results = []) {
  return walkFiles(
    dir,
    {
      extensions: new Set(exts),
      skipDirs: new Set(['node_modules', 'bindings'])
    },
    results
  )
}

function extractBalanced(src, startIdx, openChar = '(', closeChar = ')') {
  let depth = 0
  let inSingle = false
  let inDouble = false
  let inTemplate = false
  let escaped = false
  let start = -1

  for (let i = startIdx; i < src.length; i++) {
    const ch = src[i]
    if (escaped) {
      escaped = false
      continue
    }
    if (ch === '\\') {
      escaped = true
      continue
    }
    if (inSingle) {
      if (ch === "'") inSingle = false
      continue
    }
    if (inDouble) {
      if (ch === '"') inDouble = false
      continue
    }
    if (inTemplate) {
      if (ch === '`') inTemplate = false
      continue
    }
    if (ch === "'") {
      inSingle = true
      continue
    }
    if (ch === '"') {
      inDouble = true
      continue
    }
    if (ch === '`') {
      inTemplate = true
      continue
    }
    if (ch === openChar) {
      if (depth === 0) start = i + 1
      depth++
    } else if (ch === closeChar) {
      depth--
      if (depth === 0) {
        return { content: src.slice(start, i), endIdx: i }
      }
    }
  }
  return null
}

function splitTopLevelArgs(argsStr) {
  const args = []
  let depth = 0
  let inSingle = false
  let inDouble = false
  let inTemplate = false
  let escaped = false
  let start = 0

  for (let i = 0; i < argsStr.length; i++) {
    const ch = argsStr[i]
    if (escaped) {
      escaped = false
      continue
    }
    if (ch === '\\') {
      escaped = true
      continue
    }
    if (inSingle) {
      if (ch === "'") inSingle = false
      continue
    }
    if (inDouble) {
      if (ch === '"') inDouble = false
      continue
    }
    if (inTemplate) {
      if (ch === '`') inTemplate = false
      continue
    }
    if (ch === "'") inSingle = true
    else if (ch === '"') inDouble = true
    else if (ch === '`') inTemplate = true
    else if (ch === '(' || ch === '{' || ch === '[') depth++
    else if (ch === ')' || ch === '}' || ch === ']') depth--
    else if (ch === ',' && depth === 0) {
      args.push(argsStr.slice(start, i).trim())
      start = i + 1
    }
  }
  const tail = argsStr.slice(start).trim()
  if (tail) args.push(tail)
  return args
}

function parseNamedDefaultConsts(src) {
  const map = new Map()
  const constRe = /export\s+const\s+(default[A-Za-z0-9_]+)\s*(?::[^=]+)?=\s*\{/g
  let m
  while ((m = constRe.exec(src)) !== null) {
    const braceIdx = src.indexOf('{', m.index)
    const block = extractBalanced(src, braceIdx, '{', '}')
    if (block) {
      map.set(m[1], `{ ${block.content.replace(/\s+/g, ' ').trim()} }`)
    }
  }
  return map
}

function scanPersistentStores(src) {
  const stores = []
  const callRe = /persistentWritable(?:<[^>]*>)?\s*\(/g
  let m
  while ((m = callRe.exec(src)) !== null) {
    // Skip the function declaration itself
    const prefix = src.slice(Math.max(0, m.index - 20), m.index)
    if (/\bfunction\s+$/.test(prefix)) continue

    const openParen = src.indexOf('(', m.index)
    const block = extractBalanced(src, openParen, '(', ')')
    if (!block) continue
    const args = splitTopLevelArgs(block.content)
    const rawKey = args[0] || ''
    const keyMatch = rawKey.match(/^['"`]([^'"`]+)['"`]$/)
    const key = keyMatch ? keyMatch[1] : rawKey
    const defaultExpr = (args[1] || '').replace(/\s+/g, ' ').trim()
    const line = src.slice(0, m.index).split(/\r?\n/).length
    stores.push({ key, defaultExpr, line })
  }
  return stores
}

function checkSchemaMigrations(src) {
  const issues = []
  const versionMatch = src.match(/export\s+const\s+CURRENT_SETTINGS_SCHEMA_VERSION\s*=\s*(\d+)/)
  if (!versionMatch) {
    issues.push('Missing CURRENT_SETTINGS_SCHEMA_VERSION constant')
    return issues
  }
  const currentVersion = Number(versionMatch[1])
  const migIdx = src.indexOf('const migrations')
  if (migIdx === -1) {
    issues.push('Missing migrations map in settings.ts')
    return issues
  }
  const braceIdx = src.indexOf('{', migIdx)
  const block = extractBalanced(src, braceIdx, '{', '}')
  if (!block) {
    issues.push('Could not parse migrations map in settings.ts')
    return issues
  }
  for (let v = 1; v <= currentVersion; v++) {
    const re = new RegExp(`(?:^|[,{\\s])${v}\\s*:`)
    if (!re.test(block.content)) {
      issues.push(`Missing migration handler for schema version ${v}`)
    }
  }
  return issues
}

function scanDirectLocalStorage() {
  const files = walkDir(SRC_DIR, ['.ts', '.svelte'])
  const hits = []
  const lsRe = /localStorage\.(getItem|setItem|removeItem)\s*\(\s*([^,)\r\n]+)/g
  const constRe = /\b(?:const|let)\s+([A-Za-z0-9_]+)\s*=\s*['"`]([^'"`]+)['"`]/g

  for (const file of files) {
    const rel = path.relative(ROOT_DIR, file).replace(/\\/g, '/')
    if (rel === 'frontend/src/stores/settings.ts' || rel.startsWith('frontend/src/devtest/')) {
      continue
    }
    const content = fs.readFileSync(file, 'utf8')
    const constMap = new Map()
    let cm
    while ((cm = constRe.exec(content)) !== null) {
      constMap.set(cm[1], `'${cm[2]}'`)
    }

    const lines = content.split(/\r?\n/)
    lines.forEach((lineText, idx) => {
      let m
      lsRe.lastIndex = 0
      while ((m = lsRe.exec(lineText)) !== null) {
        const rawArg = m[2].trim()
        const resolved =
          constMap.get(rawArg) ||
          (/^['"`]/.test(rawArg) ? rawArg : `<dynamic: ${rawArg}>`)
        hits.push({
          file: rel,
          line: idx + 1,
          op: m[1],
          keyExpr: resolved
        })
      }
    })
  }
  return hits
}

function main() {
  if (!fs.existsSync(SETTINGS_PATH)) {
    console.error(`Error: settings.ts not found at ${SETTINGS_PATH}`)
    process.exit(1)
  }

  const src = fs.readFileSync(SETTINGS_PATH, 'utf8')
  const namedDefaults = parseNamedDefaultConsts(src)
  const stores = scanPersistentStores(src)
  const schemaIssues = checkSchemaMigrations(src)

  const seenKeys = new Map()
  const errors = [...schemaIssues]
  const warnings = []

  for (const s of stores) {
    if (!s.key) {
      errors.push(`Line ${s.line}: persistentWritable() missing storage key`)
      continue
    }
    if (seenKeys.has(s.key)) {
      errors.push(
        `Duplicate setting key "${s.key}" at line ${s.line} (first declared at line ${seenKeys.get(s.key)})`
      )
    } else {
      seenKeys.set(s.key, s.line)
    }
    if (!s.defaultExpr || s.defaultExpr === 'undefined') {
      warnings.push(`Setting "${s.key}" (line ${s.line}) has no default value`)
    }
  }

  if (SHOW_DEFAULTS) {
    console.log(`\nVinta Persistent Settings Defaults (${stores.length} stores):\n`)
    for (const s of stores) {
      const resolved = namedDefaults.get(s.defaultExpr) || s.defaultExpr || '<missing>'
      console.log(`  ${s.key.padEnd(36)} = ${resolved}`)
    }
    console.log('')
  }

  const directHits = scanDirectLocalStorage()
  const uniqueFiles = new Set(directHits.map((h) => h.file))

  if (warnings.length > 0) {
    console.log(`\n⚠ ${warnings.length} setting default warning(s):`)
    for (const w of warnings) console.log(`  - ${w}`)
  }

  if (directHits.length > 0) {
    console.log(
      `\nℹ Found ${directHits.length} direct localStorage call(s) across ${uniqueFiles.size} non-settings file(s) (audit only):`
    )
    for (const f of [...uniqueFiles].sort()) {
      const keys = [...new Set(directHits.filter((h) => h.file === f).map((h) => h.keyExpr))]
      console.log(`  ${f}: ${keys.join(', ')}`)
    }
  }

  if (errors.length > 0) {
    console.error(`\n✖ Settings validation failed with ${errors.length} error(s):`)
    for (const e of errors) console.error(`  - ${e}`)
    process.exit(1)
  }

  console.log(`\n✔ Settings validation passed (${stores.length} stores, 0 duplicate keys).`)
}

main()
