/**
 * Vinta Hardcoded UI & Design Guard Checker
 *
 * Audits frontend/src/components and frontend/src/pages (excluding devtest/) for:
 *  - Hardcoded hex/rgb/hsl colors outside CSS var(--vinta-*, fallback)
 *  - Native <select>, <input type="date|time">, native title="..." tooltips, and JS CSS imports
 *  - Hardcoded English UI strings in attributes (aria-label, placeholder, label, description)
 *
 * Exits 0 with warnings by default. Pass --strict to exit 1 on any finding,
 * or --all to print every match instead of a truncated preview.
 */

const fs = require('fs')
const path = require('path')

const ROOT_DIR = path.resolve(__dirname, '..')
const TARGET_DIRS = [
  path.join(ROOT_DIR, 'frontend/src/components'),
  path.join(ROOT_DIR, 'frontend/src/pages')
]

const STRICT = process.argv.includes('--strict')
const SHOW_ALL = process.argv.includes('--all')

function walkDir(dir, exts, out = []) {
  if (!fs.existsSync(dir)) return out
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      if (entry.name === 'node_modules' || entry.name === 'devtest') continue
      walkDir(full, exts, out)
    } else if (exts.some((ext) => entry.name.endsWith(ext))) {
      out.push(full)
    }
  }
  return out
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
      findings.nativeControls.push({ file: rel, line: lineNum, kind: 'native title="..."' })
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

  // Check JS CSS imports in script block
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
  }
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

function main() {
  const files = []
  for (const dir of TARGET_DIRS) {
    walkDir(dir, ['.css', '.svelte'], files)
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
