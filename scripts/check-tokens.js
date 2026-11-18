/**
 * Vinta CSS Token Reference Checker
 *
 * Scans frontend/src for all var(--vinta-*) usages and verifies that every
 * referenced custom property is defined in themes/tokens.css, a theme preset,
 * or an inline/component CSS/TS assignment.
 *
 * Exit code: 0 on success, 1 if any undefined --vinta-* token is referenced.
 */

const fs = require('fs')
const path = require('path')

const SRC_DIR = path.resolve(__dirname, '../frontend/src')
const EXTENSIONS = new Set(['.css', '.svelte', '.ts'])
const SKIP_DIRS = new Set(['bindings', 'node_modules'])

function walk(dir, out = []) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.isDirectory()) {
      if (SKIP_DIRS.has(entry.name)) continue
      walk(path.join(dir, entry.name), out)
    } else if (EXTENSIONS.has(path.extname(entry.name))) {
      out.push(path.join(dir, entry.name))
    }
  }
  return out
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
