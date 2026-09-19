/**
 * Safe, zero-eval arithmetic parser and evaluator for Vinta Calculator.
 * Implements tokenization and recursive-descent parsing with standard operator precedence.
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
  RightParen
}

interface Token {
  type: TokenType
  value: string
}

function tokenize(expression: string): Token[] | string {
  const sanitized = expression
    .replace(/×/g, '*')
    .replace(/÷/g, '/')
    .replace(/−/g, '-')
  const tokens: Token[] = []
  let i = 0

  while (i < sanitized.length) {
    const char = sanitized[i]

    if (char === ' ' || char === '\t') {
      i++
      continue
    }

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

      tokens.push({ type: TokenType.Number, value: numStr })
      continue
    }

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

    return `Unexpected character: ${char}`
  }

  return tokens
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

  function parseTerm(): number {
    let left = parseFactor()

    while (index < tokens.length) {
      const next = peek()
      if (
        next &&
        next.type === TokenType.Operator &&
        (next.value === '*' || next.value === '/' || next.value === '%')
      ) {
        consume()
        const right = parseFactor()
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

  function parseFactor(): number {
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
      const factor = parseFactor()
      return next.value === '-' ? -factor : factor
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

    throw new Error('Unexpected token')
  }

  try {
    const rawVal = parseExpression()
    if (index < tokens.length) {
      return { error: 'Syntax error' }
    }

    // Normalize floating point quirks e.g. 0.1 + 0.2 = 0.30000000000000004
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
