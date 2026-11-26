/**
 * Vinta Hardcoded UI & Design Guard Checker
 *
 * Audits frontend/src/components and frontend/src/pages (excluding devtest/) for:
 *  - Hardcoded hex/rgb/hsl colors outside CSS var(--vinta-*, fallback)
 *  - Native <select>, <input type="date|time">, native title="..." tooltips, and JS CSS imports
 *  - Hardcoded English UI strings in attributes (aria-label, placeholder, label, description)
 *
 * Exits 0 with warnings by default. Pass --strict to exit 1 on any finding,
 * --all to print every match instead of a truncated preview, --changed to scan
 * only modified/untracked files against origin/main, or explicit file/dir paths.
 */

const { execSync } = require('child_process')
const fs = require('fs')
const path = require('path')
const { walkFiles, hasFlag, printHelpIfRequested } = require('./lib/lint-utils')

printHelpIfRequested(
  [
    'Usage: node scripts/check-hardcoded-ui.js [options] [paths...]',
    '',
    'Options:',
    '  --all      Print every finding instead of a truncated preview',
    '  --changed  Scan only modified/untracked files against origin/main',
    '  --strict   Exit with code 1 if any findings are reported',
    '  --help     Show this help message'
  ].join('\n')
)

const ROOT_DIR = path.resolve(__dirname, '..')
const TARGET_DIRS = [
  path.join(ROOT_DIR, 'frontend/src/components'),
  path.join(ROOT_DIR, 'frontend/src/pages')
]

const ARGS = process.argv.slice(2)
const STRICT = hasFlag('--strict')
const SHOW_ALL = hasFlag('--all')
const CHANGED_ONLY = hasFlag('--changed')
const PATH_ARGS = ARGS.filter((a) => !a.startsWith('-'))

function walkDir(dir, exts, out = []) {
  if (!fs.existsSync(dir)) return out
  return walkFiles(
    dir,
    {
      extensions: new Set(exts),
      skipDirs: new Set(['node_modules', 'devtest'])
    },
    out
  )
}

function stripCssComments(text) {
  return text.replace(/\/\*[\s\S]*?\*\//g, (m) => m.replace(/[^\r\n]/g, ' '))
}

function stripVarFallbacks(line) {
  // Remove var(--vinta-..., fallback) expressions so intentional fallbacks are not flagged
  return line.replace(/var\(\s*--vinta-[^,)]+,\s*[^)]+\)/g, 'var(--vinta-token)')
}

function scanColors(file, rel, text, findings) {
  let cssText = text
  if (file.endsWith('.svelte')) {
    const styleMatch = text.match(/<style[^>]*>([\s\S]*?)<\/style>/i)
    if (!styleMatch) return
    const offsetLines = text.slice(0, styleMatch.index).split(/\r?\n/).length - 1
    cssText = '\n'.repeat(offsetLines) + styleMatch[1]
  }

  const clean = stripCssComments(cssText)
  const lines = clean.split(/\r?\n/)
  const colorRe = /(#[0-9a-fA-F]{3,8}\b|\b(?:rgba?|hsla?)\s*\([^)]+\))/g

  lines.forEach((rawLine, idx) => {
    const line = stripVarFallbacks(rawLine)
    if (/^\s*--vinta-/.test(line)) return
    let m
    colorRe.lastIndex = 0
    while ((m = colorRe.exec(line)) !== null) {
      findings.colors.push({
        file: rel,
        line: idx + 1,
        match: m[1]
      })
    }
  })
}

function scanSvelteTemplate(rel, text, findings) {
  // Strip <script> and <style> blocks while preserving line numbers
  const templateOnly = text
    .replace(/<script[\s\S]*?<\/script>/gi, (m) => m.replace(/[^\r\n]/g, ' '))
    .replace(/<style[\s\S]*?<\/style>/gi, (m) => m.replace(/[^\r\n]/g, ' '))
    .replace(/<!--[\s\S]*?-->/g, (m) => m.replace(/[^\r\n]/g, ' '))

  const lines = templateOnly.split(/\r?\n/)
  const nativeSelectRe = /<select\b/i
  const nativeDateRe = /<input\b[^>]*\btype\s*=\s*["'](?:date|time)["']/i
  const nativeTitleRe = /\btitle\s*=\s*["'][^"']+["']/i
  const attrStringRe = /\b(aria-label|placeholder|label|description)\s*=\s*["']([^"'{]+)["']/g

  lines.forEach((line, idx) => {
    const lineNum = idx + 1
    if (nativeSelectRe.test(line)) {
      findings.nativeControls.push({ file: rel, line: lineNum, kind: '<select>' })
    }
    if (nativeDateRe.test(line)) {
      findings.nativeControls.push({
        file: rel,
        line: lineNum,
        kind: '<input type="date|time">'
      })
    }
    if (nativeTitleRe.test(line) && !/<iframe\b/i.test(line)) {
      // Ignore title="..." props passed to PascalCase Svelte components (e.g. <DrawerPanel title="...">)
      const tagContext = lines.slice(Math.max(0, idx - 8), idx + 1).join('\n')
      const openTagMatches = [...tagContext.matchAll(/<([A-Za-z][A-Za-z0-9_-]*)\b/g)]
      const lastTag = openTagMatches.length > 0 ? openTagMatches[openTagMatches.length - 1][1] : ''
      const isSvelteComponent = /^[A-Z]/.test(lastTag)
      if (!isSvelteComponent) {
        findings.nativeControls.push({ file: rel, line: lineNum, kind: 'native title="..."' })
      }
    }

    let m
    attrStringRe.lastIndex = 0
    while ((m = attrStringRe.exec(line)) !== null) {
      const val = m[2].trim()
      if (!val || !/[A-Za-z]{2,}/.test(val)) continue
      findings.strings.push({
        file: rel,
        line: lineNum,
        attr: m[1],
        text: val
      })
    }
  })

  // Check JS CSS imports and banned Sparkle/Sparkles icon imports in script blocks
  const scriptMatches = text.match(/<script[\s\S]*?<\/script>/gi) || []
  for (const block of scriptMatches) {
    const importRe = /import\s+['"][^'"]+\.css['"]/g
    if (importRe.test(block)) {
      const lineNum = text.slice(0, text.indexOf(block)).split(/\r?\n/).length
      findings.nativeControls.push({
        file: rel,
        line: lineNum,
        kind: 'JS CSS import (use <style src="...">)'
      })
    }
    const sparkleRe = /import\s+\{[^}]*\bSparkles?\b[^}]*\}\s+from\s+['"]lucide-svelte['"]/g
    if (sparkleRe.test(block)) {
      const lineNum = text.slice(0, text.indexOf(block)).split(/\r?\n/).length
      findings.nativeControls.push({
        file: rel,
        line: lineNum,
        kind: 'Banned Sparkle/Sparkles icon from lucide-svelte'
      })
    }
  }

  // Check icon size hygiene (even >= 12px, respecting <!-- vinta-ignore icon-size -->)
  const rawLines = text.split(/\r?\n/)
  const iconSizeRe = /<[A-Z][A-Za-z0-9_]*\b[^>]*\bsize=\{(\d+(?:\.\d+)?)\}/g
  rawLines.forEach((rawLine, idx) => {
    const prevLine = idx > 0 ? rawLines[idx - 1] : ''
    if (rawLine.includes('vinta-ignore icon-size') || prevLine.includes('vinta-ignore icon-size')) {
      return
    }
    iconSizeRe.lastIndex = 0
    let match
    while ((match = iconSizeRe.exec(rawLine)) !== null) {
      const n = Number(match[1])
      if (Number.isFinite(n) && (n < 12 || n % 2 !== 0)) {
        findings.nativeControls.push({
          file: rel,
          line: idx + 1,
          kind: `Icon size={${match[1]}} must be even and >= 12px`
        })
      }
    }
  })
}

function printGroup(title, items, formatItem) {
  if (items.length === 0) return
  console.warn(`\n⚠ ${title} (${items.length}):`)
  const limit = SHOW_ALL ? items.length : 15
  for (const item of items.slice(0, limit)) {
    console.warn(`  ${formatItem(item)}`)
  }
  if (!SHOW_ALL && items.length > limit) {
    console.warn(`  ... and ${items.length - limit} more (pass --all to view all)`)
  }
}

function isTargetFile(rel) {
  return (
    (rel.startsWith('frontend/src/components/') || rel.startsWith('frontend/src/pages/')) &&
    !rel.includes('/devtest/') &&
    (rel.endsWith('.css') || rel.endsWith('.svelte'))
  )
}

function runGitLines(cmd) {
  try {
    return execSync(cmd, { cwd: ROOT_DIR, encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'] })
      .split(/\r?\n/)
      .map((l) => l.trim().replace(/\\/g, '/'))
      .filter(Boolean)
  } catch {
    return []
  }
}

function getChangedFiles() {
  const candidates = new Set([
    ...runGitLines('git diff --name-only --diff-filter=ACMRTUXB origin/main...HEAD'),
    ...runGitLines('git diff --name-only --diff-filter=ACMRTUXB'),
    ...runGitLines('git diff --cached --name-only --diff-filter=ACMRTUXB'),
    ...runGitLines('git ls-files --others --exclude-standard')
  ])

  return [...candidates]
    .filter(isTargetFile)
    .map((rel) => path.join(ROOT_DIR, rel))
    .filter((full) => fs.existsSync(full))
}

function getExplicitFiles() {
  const files = []
  for (const rawArg of PATH_ARGS) {
    const resolved = fs.existsSync(path.resolve(process.cwd(), rawArg))
      ? path.resolve(process.cwd(), rawArg)
      : path.resolve(ROOT_DIR, rawArg)
    if (!fs.existsSync(resolved)) continue
    const stat = fs.statSync(resolved)
    if (stat.isDirectory()) {
      walkDir(resolved, ['.css', '.svelte'], files)
    } else if (resolved.endsWith('.css') || resolved.endsWith('.svelte')) {
      files.push(resolved)
    }
  }
  return [...new Set(files)]
}

function main() {
  let files = []
  if (PATH_ARGS.length > 0) {
    files = getExplicitFiles()
  } else if (CHANGED_ONLY) {
    files = getChangedFiles()
    if (files.length === 0) {
      console.log('✔ No changed UI component/page files to check.')
      return
    }
  } else {
    for (const dir of TARGET_DIRS) {
      walkDir(dir, ['.css', '.svelte'], files)
    }
  }

  const findings = {
    colors: [],
    nativeControls: [],
    strings: []
  }

  for (const file of files) {
    const rel = path.relative(ROOT_DIR, file).replace(/\\/g, '/')
    const text = fs.readFileSync(file, 'utf8')
    scanColors(file, rel, text, findings)
    if (file.endsWith('.svelte')) {
      scanSvelteTemplate(rel, text, findings)
    }
  }

  const total =
    findings.colors.length + findings.nativeControls.length + findings.strings.length

  if (total === 0) {
    console.log('✔ No hardcoded UI colors, native controls, or hardcoded attribute strings found.')
    return
  }

  printGroup(
    'Hardcoded colors (prefer --vinta-* tokens)',
    findings.colors,
    (i) => `${i.file}:${i.line} -> ${i.match}`
  )
  printGroup(
    'Disallowed native controls / JS CSS imports',
    findings.nativeControls,
    (i) => `${i.file}:${i.line} -> ${i.kind}`
  )
  printGroup(
    'Hardcoded attribute strings (prefer $tStore(...))',
    findings.strings,
    (i) => `${i.file}:${i.line} [${i.attr}] -> "${i.text}"`
  )
  console.warn('')

  if (STRICT) {
    process.exit(1)
  }
}

main()
