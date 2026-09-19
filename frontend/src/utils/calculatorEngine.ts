/**
 * Safe, zero-eval arithmetic parser and evaluator for Vinta Calculator.
 * Supports standard arithmetic, superscript powers (e.g. 9⁵, 2³), square roots (√16),
 * editable nth-roots (e.g. ³√27, ⁴√16), parentheses, and floating-point precision stabilization.
 */

export interface CalculationResult {
  value: number
  display: string
}

export interface CalculationError {
  error: string
}

export type EvalOutcome = CalculationResult | CalculationError

export function isCalculationError(
  outcome: EvalOutcome
): outcome is CalculationError {
  return 'error' in outcome
}

enum TokenType {
  Number,
  Operator,
  LeftParen,
  RightParen,
  Comma,
  Sqrt,
  NthRoot
}

interface Token {
  type: TokenType
  value: string
}

function fromSuperscript(str: string): string {
  const map: Record<string, string> = {
    '⁰': '0',
    '¹': '1',
    '²': '2',
    '³': '3',
    '⁴': '4',
    '⁵': '5',
    '⁶': '6',
    '⁷': '7',
    '⁸': '8',
    '⁹': '9'
  }
  return str
    .split('')
    .map((c) => map[c] || c)
    .join('')
}

function sanitizeExpression(expr: string): string {
  let s = expr
    .replace(/×/g, '*')
    .replace(/÷/g, '/')
    .replace(/−/g, '-')
    .replace(/xʸ/g, '^')
    .replace(/ʸ√/g, '√')
    .replace(/ⁿ√/g, '√')
    .replace(/ʸ/g, '')

  // 1. Convert root indices: ³√27 -> 3√27
  s = s.replace(/([⁰¹²³⁴⁵⁶⁷⁸⁹]+)√/g, (_, p1) => {
    return fromSuperscript(p1) + '√'
  })

  // 2. Convert remaining superscripts into powers: 9⁵ -> 9 ^ 5
  s = s.replace(/([⁰¹²³⁴⁵⁶⁷⁸⁹]+)/g, (_, p1) => {
    return ' ^ ' + fromSuperscript(p1)
  })

  return s
}

function tokenize(expression: string): Token[] | string {
  const sanitized = sanitizeExpression(expression)
  const tokens: Token[] = []
  let i = 0

  while (i < sanitized.length) {
    const char = sanitized[i]

    if (char === ' ' || char === '\t') {
      i++
      continue
    }

    // Number literals (including decimals and attached nth-root like 3√27)
    if ((char >= '0' && char <= '9') || char === '.') {
      let numStr = ''
      let hasDot = false

      while (i < sanitized.length) {
        const c = sanitized[i]
        if (c === '.') {
          if (hasDot) return 'Invalid number format'
          hasDot = true
          numStr += c
        } else if (c >= '0' && c <= '9') {
          numStr += c
        } else {
          break
        }
        i++
      }

      // Check if this number is directly followed by a root radical, e.g. 3√27
      if (i < sanitized.length && sanitized[i] === '√') {
        tokens.push({ type: TokenType.NthRoot, value: numStr })
        i++ // consume the radical symbol
        continue
      }

      tokens.push({ type: TokenType.Number, value: numStr })
      continue
    }

    // Standalone radical symbol: square root
    if (char === '√') {
      tokens.push({ type: TokenType.Sqrt, value: '2' })
      i++
      continue
    }

    // Cube root radical symbol
    if (char === '∛') {
      tokens.push({ type: TokenType.NthRoot, value: '3' })
      i++
      continue
    }

    // Exponentiation operator
    if (char === '^') {
      tokens.push({ type: TokenType.Operator, value: '^' })
      i++
      continue
    }

    // Standard arithmetic operators
    if (
      char === '+' ||
      char === '-' ||
      char === '*' ||
      char === '/' ||
      char === '%'
    ) {
      tokens.push({ type: TokenType.Operator, value: char })
      i++
      continue
    }

    // Parentheses
    if (char === '(') {
      tokens.push({ type: TokenType.LeftParen, value: char })
      i++
      continue
    }

    if (char === ')') {
      tokens.push({ type: TokenType.RightParen, value: char })
      i++
      continue
    }

    if (char === ',') {
      tokens.push({ type: TokenType.Comma, value: char })
      i++
      continue
    }

    return `Unexpected character: ${char}`
  }

  return tokens
}

/**
 * Calculates the y-th root of x safely.
 */
function calculateNthRoot(base: number, rootIndex: number): number {
  if (rootIndex === 0) {
    throw new Error('Cannot take 0th root')
  }
  if (base === 0) {
    return 0
  }
  if (base < 0) {
    // Check if rootIndex is an odd integer
    if (Math.abs(rootIndex % 2) === 1) {
      return -Math.pow(-base, 1 / rootIndex)
    }
    throw new Error('Invalid root of negative number')
  }
  return Math.pow(base, 1 / rootIndex)
}

export function evaluateExpression(expr: string): EvalOutcome {
  const trimmed = expr.trim()
  if (!trimmed) {
    return { value: 0, display: '0' }
  }

  const tokenResult = tokenize(trimmed)
  if (typeof tokenResult === 'string') {
    return { error: tokenResult }
  }

  const tokens = tokenResult
  if (tokens.length === 0) {
    return { value: 0, display: '0' }
  }

  let index = 0

  function peek(): Token | null {
    return index < tokens.length ? tokens[index] : null
  }

  function consume(): Token {
    return tokens[index++]
  }

  // Lowest precedence: Addition and Subtraction
  function parseExpression(): number {
    let left = parseTerm()

    while (index < tokens.length) {
      const next = peek()
      if (
        next &&
        next.type === TokenType.Operator &&
        (next.value === '+' || next.value === '-')
      ) {
        consume()
        const right = parseTerm()
        if (next.value === '+') {
          left = left + right
        } else {
          left = left - right
        }
      } else {
        break
      }
    }

    return left
  }

  // Medium precedence: Multiplication, Division, Modulo
  function parseTerm(): number {
    let left = parsePower()

    while (index < tokens.length) {
      const next = peek()
      if (
        next &&
        next.type === TokenType.Operator &&
        (next.value === '*' || next.value === '/' || next.value === '%')
      ) {
        consume()
        const right = parsePower()
        if (next.value === '*') {
          left = left * right
        } else if (next.value === '/') {
          if (right === 0) {
            throw new Error('Cannot divide by zero')
          }
          left = left / right
        } else if (next.value === '%') {
          if (right === 0) {
            throw new Error('Cannot divide by zero')
          }
          left = left % right
        }
      } else {
        break
      }
    }

    return left
  }

  // Higher precedence: Power (^)
  function parsePower(): number {
    let left = parseUnary()

    while (index < tokens.length) {
      const next = peek()
      if (next && next.type === TokenType.Operator && next.value === '^') {
        consume()
        const right = parseUnary()
        left = Math.pow(left, right)
      } else {
        break
      }
    }

    return left
  }

  // Highest operator precedence: Unary signs, Roots (√ and 3√)
  function parseUnary(): number {
    const next = peek()
    if (!next) {
      throw new Error('Incomplete expression')
    }

    // Unary plus or minus
    if (
      next.type === TokenType.Operator &&
      (next.value === '+' || next.value === '-')
    ) {
      consume()
      const factor = parseUnary()
      return next.value === '-' ? -factor : factor
    }

    // Square root radical: √16 or √(16)
    if (next.type === TokenType.Sqrt) {
      consume()
      const radicand = parsePrimary()
      if (radicand < 0) {
        throw new Error('Cannot take square root of negative number')
      }
      return Math.sqrt(radicand)
    }

    // Nth-root radical: 3√27 or 4√16
    if (next.type === TokenType.NthRoot) {
      const rootToken = consume()
      const rootIdx = parseFloat(rootToken.value)
      const radicand = parsePrimary()
      return calculateNthRoot(radicand, rootIdx)
    }

    // Primary: Parentheses and Numbers
    return parsePrimary()
  }

  function parsePrimary(): number {
    const next = peek()
    if (!next) {
      throw new Error('Incomplete expression')
    }

    // Parentheses
    if (next.type === TokenType.LeftParen) {
      consume()
      const val = parseExpression()
      const closing = peek()
      if (!closing || closing.type !== TokenType.RightParen) {
        throw new Error('Missing closing parenthesis')
      }
      consume()
      return val
    }

    // Number literal
    if (next.type === TokenType.Number) {
      consume()
      const num = parseFloat(next.value)
      if (isNaN(num)) {
        throw new Error('Invalid number')
      }
      return num
    }

    // If another root immediately follows, parse as nested unary
    if (next.type === TokenType.Sqrt || next.type === TokenType.NthRoot) {
      return parseUnary()
    }

    throw new Error('Unexpected token')
  }

  try {
    const rawVal = parseExpression()
    if (index < tokens.length) {
      return { error: 'Syntax error' }
    }

    // Normalize floating point quirks e.g. 0.1 + 0.2 = 0.3
    const rounded = parseFloat(rawVal.toPrecision(12))
    return {
      value: rounded,
      display: formatNumberDisplay(rounded)
    }
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : 'Invalid calculation'
    return { error: message }
  }
}

export function formatNumberDisplay(num: number): string {
  if (!isFinite(num)) {
    return 'Error'
  }

  const parts = num.toString().split('.')
  const integerPart = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  if (parts.length > 1) {
    return `${integerPart}.${parts[1]}`
  }
  return integerPart
}
