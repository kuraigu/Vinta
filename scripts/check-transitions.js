/**
 * Vinta CSS Transition Compositor Checker
 *
 * Scans frontend/src (.css and .svelte) for `transition: all` and non-composited
 * transition properties that force main-thread Paint/Layout and CrGpuMain GPUTask
 * spikes over transparent WebView2 surfaces.
 *
 * By default exits 0 with warnings. Pass --strict to exit 1 on any warning,
 * and --all to print every warning without truncation.
 */

const fs = require('fs')
const path = require('path')
const {
  walkFiles,
  hasFlag,
  printHelpIfRequested,
  hasVintaIgnore
} = require('./lib/lint-utils')

const ROOT_DIR = path.resolve(__dirname, '..')
const SRC_DIR = path.resolve(ROOT_DIR, 'frontend/src')
const APP_CSS_PATH = path.resolve(SRC_DIR, 'app.css')
const EXTENSIONS = new Set(['.css', '.svelte'])
const SKIP_DIRS = new Set(['bindings', 'node_modules', 'presets', 'devtest'])

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

function hasGlobalReducedMotionRule() {
  if (!fs.existsSync(APP_CSS_PATH)) return false
  const appCss = fs.readFileSync(APP_CSS_PATH, 'utf8')
  return /@media\s*\(\s*prefers-reduced-motion\s*:\s*reduce\s*\)/i.test(appCss)
}

function run() {
  printHelpIfRequested(
    [
      'Usage: node scripts/check-transitions.js [options]',
      '',
      'Options:',
      '  --all     Print every warning without truncating at 25 entries',
      '  --strict  Exit with code 1 if any warnings are found',
      '  --help    Show this help message'
    ].join('\n')
  )

  const strict = hasFlag('--strict')
  const showAll = hasFlag('--all')
  const files = walkFiles(SRC_DIR, { extensions: EXTENSIONS, skipDirs: SKIP_DIRS })
  const warnings = []
  const reducedMotionWarnings = []
  const hasGlobalReducedMotion = hasGlobalReducedMotionRule()

  const transitionDeclRegex = /transition(?:-property)?\s*:\s*([^;}{]+);/gi
  const keyframeRegex = /@keyframes\s+[A-Za-z0-9_-]+/g

  for (const file of files) {
    const text = fs.readFileSync(file, 'utf8')
    const rel = path.relative(ROOT_DIR, file)
    let m
    transitionDeclRegex.lastIndex = 0
    while ((m = transitionDeclRegex.exec(text)) !== null) {
      if (hasVintaIgnore(text, m.index, 'transition')) continue
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

    if (
      !hasGlobalReducedMotion &&
      keyframeRegex.test(text) &&
      !/prefers-reduced-motion\s*:/.test(text)
    ) {
      reducedMotionWarnings.push(rel)
    }
  }

  if (warnings.length > 0) {
    console.warn(
      `\n⚠ Found ${warnings.length} non-composited CSS transition(s) (prefer transform/opacity only):\n`
    )
    const displayList = showAll ? warnings : warnings.slice(0, 25)
    for (const w of displayList) {
      console.warn(`  ${w.file}:${w.line} -> ${w.props}`)
    }
    if (!showAll && warnings.length > 25) {
      console.warn(`  ... and ${warnings.length - 25} more (pass --all to view all)`)
    }
    console.warn('')
  } else {
    console.log('✔ All CSS transitions use compositor-safe properties')
  }

  if (reducedMotionWarnings.length > 0) {
    console.warn(
      `⚠ Found ${reducedMotionWarnings.length} file(s) with @keyframes missing prefers-reduced-motion:`
    )
    for (const f of reducedMotionWarnings) {
      console.warn(`  ${f}`)
    }
    console.warn('')
  } else {
    console.log('✔ All production @keyframes respect prefers-reduced-motion')
  }

  if (strict && (warnings.length > 0 || reducedMotionWarnings.length > 0)) {
    process.exit(1)
  }
}

run()
