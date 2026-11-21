/**
 * @fileoverview Shared file-walking, CLI flag, and ignore-directive helpers for Vinta lint scripts.
 */

const fs = require('fs')
const path = require('path')

/**
 * Recursively collects files matching `extensions` while skipping `skipDirs`.
 *
 * @param {string} dir
 * @param {{ extensions: Set<string>, skipDirs?: Set<string> }} options
 * @param {string[]} [out]
 * @returns {string[]}
 */
function walkFiles(dir, { extensions, skipDirs = new Set() }, out = []) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.isDirectory()) {
      if (skipDirs.has(entry.name)) continue
      walkFiles(path.join(dir, entry.name), { extensions, skipDirs }, out)
    } else if (extensions.has(path.extname(entry.name))) {
      out.push(path.join(dir, entry.name))
    }
  }
  return out
}

/**
 * Checks whether a CLI flag (e.g. `--all`, `--strict`) was passed in `process.argv`.
 *
 * @param {string} flag
 * @returns {boolean}
 */
function hasFlag(flag) {
  return process.argv.slice(2).includes(flag)
}

/**
 * Prints `helpText` and exits 0 if `--help` or `-h` was passed.
 *
 * @param {string} helpText
 * @returns {boolean}
 */
function printHelpIfRequested(helpText) {
  if (hasFlag('--help') || hasFlag('-h')) {
    console.log(helpText)
    process.exit(0)
  }
  return false
}

/**
 * Checks whether a match at `index` is suppressed by a `vinta-ignore` comment
 * on the same line or the immediately preceding non-empty line.
 *
 * Supports:
 * - `/* vinta-ignore *\/`
 * - `/* vinta-ignore: <rule> *\/`
 * - `<!-- vinta-ignore -->` / `<!-- vinta-ignore: <rule> -->`
 * - `// vinta-ignore` / `// vinta-ignore: <rule>`
 *
 * @param {string} text
 * @param {number} index
 * @param {string} [rule]
 * @returns {boolean}
 */
function hasVintaIgnore(text, index, rule) {
  const before = text.slice(0, index)
  const lineStart = before.lastIndexOf('\n') + 1
  const lineEndIdx = text.indexOf('\n', index)
  const currentLine = text.slice(
    lineStart,
    lineEndIdx === -1 ? text.length : lineEndIdx
  )

  const lines = [currentLine]
  if (lineStart > 1) {
    const prevLines = before.slice(0, lineStart - 1).split(/\r?\n/)
    for (let i = prevLines.length - 1; i >= 0; i--) {
      const trimmed = prevLines[i].trim()
      if (!trimmed) continue
      lines.push(trimmed)
      break
    }
  }

  const rulePattern = rule ? `(?:\\s*:\\s*${rule})?` : '(?:\\s*:\\s*[a-z0-9_-]+)?'
  const ignoreRegex = new RegExp(
    `(?:\\/\\*|<!--|\\/\\/)\\s*vinta-ignore${rulePattern}\\b`,
    'i'
  )

  return lines.some((line) => ignoreRegex.test(line))
}

module.exports = {
  walkFiles,
  hasFlag,
  printHelpIfRequested,
  hasVintaIgnore
}
