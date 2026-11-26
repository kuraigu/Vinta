/**
 * Vinta CSS Token Reference Checker
 *
 * Scans frontend/src for all var(--vinta-*) usages and verifies that every
 * referenced custom property is defined in the canonical base theme
 * (themes/presets/dark-gold/style.css) or an inline/component CSS/TS assignment.
 *
 * Exit code: 0 on success, 1 if any undefined --vinta-* token is referenced.
 */

const fs = require('fs')
const path = require('path')
const { walkFiles, printHelpIfRequested } = require('./lib/lint-utils')

printHelpIfRequested(
  [
    'Usage: node scripts/check-tokens.js [--help]',
    '',
    'Scans frontend/src for var(--vinta-*) usages and verifies every token is defined.'
  ].join('\n')
)

const SRC_DIR = path.resolve(__dirname, '../frontend/src')
const PRESETS_DIR = path.resolve(SRC_DIR, 'themes/presets')
const CANONICAL_BASE_PRESET = path.resolve(PRESETS_DIR, 'dark-gold')
const EXTENSIONS = new Set(['.css', '.svelte', '.ts'])
const SKIP_DIRS = new Set(['bindings', 'node_modules'])

function isNonBaseThemePreset(filePath) {
  const resolved = path.resolve(filePath)
  return (
    resolved.startsWith(PRESETS_DIR + path.sep) &&
    !resolved.startsWith(CANONICAL_BASE_PRESET + path.sep)
  )
}

function walk(dir, out = []) {
  return walkFiles(dir, { extensions: EXTENSIONS, skipDirs: SKIP_DIRS }, out)
}

function run() {
  const files = walk(SRC_DIR)
  const defined = new Set()

  const defCssRegex = /(--vinta-[a-z0-9-]+)\s*:/g
  const defJsRegex = /setProperty\(\s*['"](--vinta-[a-z0-9-]+)['"]/g
  // Flag undefined var(--vinta-*) references that have no fallback value.
  const useRegex = /var\(\s*(--vinta-[a-z0-9-]+)\s*\)/g

  const contents = new Map()
  for (const file of files) {
    const text = fs.readFileSync(file, 'utf8')
    contents.set(file, text)

    // Only collect canonical theme tokens from dark-gold/style.css (plus component/runtime vars)
    // so tokens defined solely in a secondary preset do not mask missing base tokens.
    if (isNonBaseThemePreset(file)) {
      continue
    }

    let m
    defCssRegex.lastIndex = 0
    while ((m = defCssRegex.exec(text)) !== null) {
      defined.add(m[1])
    }
    defJsRegex.lastIndex = 0
    while ((m = defJsRegex.exec(text)) !== null) {
      defined.add(m[1])
    }
  }

  const errors = []
  for (const [file, text] of contents) {
    const lines = text.split(/\r?\n/)
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]
      let m
      useRegex.lastIndex = 0
      while ((m = useRegex.exec(line)) !== null) {
        const token = m[1]
        if (!defined.has(token)) {
          errors.push({
            file: path.relative(path.resolve(__dirname, '..'), file),
            line: i + 1,
            token
          })
        }
      }
    }
  }

  if (errors.length > 0) {
    console.error(`\n✖ Found ${errors.length} undefined --vinta-* token reference(s):\n`)
    for (const e of errors) {
      console.error(`  ${e.file}:${e.line} -> ${e.token}`)
    }
    console.error('')
    process.exit(1)
  }

  console.log(`✔ All --vinta-* token references valid (${defined.size} tokens defined)`)
}

run()
