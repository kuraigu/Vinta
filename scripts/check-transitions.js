/**
 * Vinta CSS Transition Compositor Checker
 *
 * Scans frontend/src (.css and .svelte) for `transition: all` and non-composited
 * transition properties that force main-thread Paint/Layout and CrGpuMain GPUTask
 * spikes over transparent WebView2 surfaces.
 *
 * By default exits 0 with warnings (similar to check-locales.js). Pass --strict
 * to exit 1 when any non-composited transitions are found.
 */

const fs = require('fs')
const path = require('path')

const SRC_DIR = path.resolve(__dirname, '../frontend/src')
const EXTENSIONS = new Set(['.css', '.svelte'])
const SKIP_DIRS = new Set(['bindings', 'node_modules', 'presets'])

const NON_COMPOSITED_PROPS = [
  'all',
  'box-shadow',
  'background',
  'background-color',
  'border-color',
  'border-top-color',
  'border-right-color',
  'border-bottom-color',
  'border-left-color',
  'color',
  'width',
  'height',
  'top',
  'left',
  'right',
  'bottom',
  'margin',
  'padding',
  'filter',
  'backdrop-filter',
  'stroke-dashoffset'
]

const PROP_REGEX = new RegExp(`\\b(${NON_COMPOSITED_PROPS.join('|')})\\b`, 'gi')

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
  const strict = process.argv.includes('--strict')
  const files = walk(SRC_DIR)
  const warnings = []

  const transitionDeclRegex = /transition(?:-property)?\s*:\s*([^;}{]+);/gi

  for (const file of files) {
    const text = fs.readFileSync(file, 'utf8')
    const rel = path.relative(path.resolve(__dirname, '..'), file)
    let m
    transitionDeclRegex.lastIndex = 0
    while ((m = transitionDeclRegex.exec(text)) !== null) {
      const value = m[1]
      if (/^\s*none\b/i.test(value)) continue
      const matched = new Set()
      let pm
      PROP_REGEX.lastIndex = 0
      while ((pm = PROP_REGEX.exec(value)) !== null) {
        matched.add(pm[1].toLowerCase())
      }
      if (matched.size > 0) {
        const lineNum = text.slice(0, m.index).split(/\r?\n/).length
        warnings.push({
          file: rel,
          line: lineNum,
          props: Array.from(matched).join(', ')
        })
      }
    }
  }

  if (warnings.length > 0) {
    console.warn(
      `\n⚠ Found ${warnings.length} non-composited CSS transition(s) (prefer transform/opacity only):\n`
    )
    for (const w of warnings.slice(0, 25)) {
      console.warn(`  ${w.file}:${w.line} -> ${w.props}`)
    }
    if (warnings.length > 25) {
      console.warn(`  ... and ${warnings.length - 25} more`)
    }
    console.warn('')
    if (strict) process.exit(1)
    return
  }

  console.log('✔ All CSS transitions use compositor-safe properties')
}

run()
