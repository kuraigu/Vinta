<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 220, h: 36, r: 16 },
    expanded: { w: 340, h: 434, r: 28 }
  }
</script>

<script lang="ts">
  /**
   * @fileoverview Floating desktop calculator with scientific operations.
   *
   * Handles standard arithmetic, powers, roots, and parenthetical expressions with
   * physical keyboard and numpad support.
   */

  import { Delete } from 'lucide-svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import { tStore } from '../store'
  import {
    evaluateExpression,
    isCalculationError
  } from '../utils/calculatorEngine'

  let expression = $state('0')
  let historyFormula = $state('')
  let errorMessage = $state('')
  let isExponentMode = $state(false)
  let copiedIndicator = $state(false)
  let copiedTimer: ReturnType<typeof setTimeout> | null = null

  /**
   * Copies the current calculator expression or result to the system clipboard.
   */
  function handleCopyResult() {
    if (errorMessage) return
    void navigator.clipboard?.writeText(expression).catch(() => {})
    copiedIndicator = true
    if (copiedTimer) clearTimeout(copiedTimer)
    copiedTimer = setTimeout(() => {
      copiedIndicator = false
    }, 1200)
  }

  let livePreview = $derived.by(() => {
    if (!expression || expression === '0' || errorMessage) return ''
    const res = evaluateExpression(expression)
    if (!isCalculationError(res) && res.display !== expression) {
      return `= ${res.display}`
    }
    return ''
  })

  /**
   * Resets accumulator, current expression buffer, and calculation history to zero.
   */
  function handleClear() {
    expression = '0'
    historyFormula = ''
    errorMessage = ''
    isExponentMode = false
  }

  const SUPERSCRIPT_MAP: Record<string, string> = {
    '0': '⁰',
    '1': '¹',
    '2': '²',
    '3': '³',
    '4': '⁴',
    '5': '⁵',
    '6': '⁶',
    '7': '⁷',
    '8': '⁸',
    '9': '⁹'
  }

  /**
   * Converts standard numeric characters into unicode superscript characters for exponent display.
   *
   * @param numStr - Numeric string to convert (e.g. "2" -> "²").
   * @returns Superscript formatted string.
   */
  function toSuperscript(numStr: string): string {
    return numStr
      .split('')
      .map((c) => SUPERSCRIPT_MAP[c] || c)
      .join('')
  }

  /**
   * Deletes the trailing character or operator token from the current expression buffer.
   */
  function handleBackspace() {
    errorMessage = ''
    if (expression.endsWith('ʸ')) {
      expression = expression.slice(0, -1)
      isExponentMode = false
    } else if (expression.endsWith('ⁿ√')) {
      expression = expression.slice(0, -2)
      if (!expression) expression = '0'
    } else if (/[⁰¹²³⁴⁵⁶⁷⁸⁹]+√$/.test(expression)) {
      // Revert root index from e.g. ³√ back to ⁿ√
      expression = expression.replace(/[⁰¹²³⁴⁵⁶⁷⁸⁹]+√$/, 'ⁿ√')
    } else if (/[⁰¹²³⁴⁵⁶⁷⁸⁹]$/.test(expression)) {
      expression = expression.slice(0, -1)
      if (!/[⁰¹²³⁴⁵⁶⁷⁸⁹]$/.test(expression)) {
        isExponentMode = false
      }
    } else if (expression.length > 1) {
      expression = expression.slice(0, -1)
    } else {
      expression = '0'
      isExponentMode = false
    }
  }

  /**
   * Appends a numeric digit (0-9) to the active operand buffer.
   *
   * @param d - Single digit character.
   */
  function handleDigit(d: string) {
    errorMessage = ''
    if (expression === '0') {
      expression = d
      return
    }

    // If expression currently ends with ⁿ√, replace ⁿ with the typed digit in superscript
    if (expression.endsWith('ⁿ√')) {
      expression = expression.slice(0, -2) + toSuperscript(d) + '√'
      return
    }

    // If in exponent mode and expression ends with ʸ placeholder, replace ʸ with superscript digit
    if (expression.endsWith('ʸ')) {
      expression = expression.slice(0, -1) + toSuperscript(d)
      return
    }

    // If previous character was a superscript digit and we are still in exponent mode, continue superscript
    if (isExponentMode && /[⁰¹²³⁴⁵⁶⁷⁸⁹]$/.test(expression)) {
      expression += toSuperscript(d)
      return
    }

    isExponentMode = false
    expression += d
  }

  /**
   * Appends a decimal separator to the active operand if one is not already present.
   */
  function handleDecimal() {
    errorMessage = ''
    // Check if current active number chunk already contains decimal
    const parts = expression.split(/[\s+−×÷^()]+/)
    const currentPart = parts[parts.length - 1] || ''
    if (!currentPart.includes('.')) {
      expression += '.'
    }
  }

  /**
   * Appends a binary arithmetic operator (+, -, *, /) to the expression buffer.
   *
   * @param op - Operator symbol character.
   */
  function handleOperator(op: string) {
    errorMessage = ''
    isExponentMode = false

    if (expression.endsWith('ʸ')) {
      expression = expression.slice(0, -1)
    }

    if (expression === '0' && op === '−') {
      expression = '−'
      return
    }

    const trimmed = expression.trim()
    const lastChar = trimmed[trimmed.length - 1]
    if (['+', '−', '×', '÷', '^'].includes(lastChar)) {
      expression = trimmed.slice(0, -1) + op
    } else {
      expression += ` ${op} `
    }
  }

  /**
   * Applies immediate square root calculation to the active operand.
   */
  function handleSquareRoot() {
    errorMessage = ''
    if (expression === '0') {
      expression = '√'
    } else {
      const trimmed = expression.trim()
      const lastChar = trimmed[trimmed.length - 1]
      if (['+', '−', '×', '÷', '^', '('].includes(lastChar)) {
        expression += '√'
      } else {
        expression = `√(${expression})`
      }
    }
  }

  /**
   * Inserts an nth-root radical operator into the expression.
   */
  function handleRoot() {
    errorMessage = ''
    if (expression === '0') {
      // Waiting for root index (e.g. user types 3 -> ³√ or 4 -> ⁴√)
      expression = 'ⁿ√'
      return
    }

    const trimmed = expression.trim()
    const lastChar = trimmed[trimmed.length - 1]

    // If last token is a number, convert it into the superscript root index
    if (lastChar >= '0' && lastChar <= '9') {
      const match = trimmed.match(/(\d+)$/)
      if (match) {
        const num = match[1]
        const before = trimmed.slice(0, -num.length)
        expression = `${before}${toSuperscript(num)}√`
        return
      }
    }

    if (['+', '−', '×', '÷', '^', '('].includes(lastChar)) {
      expression += 'ⁿ√'
    } else {
      expression += ' × ⁿ√'
    }
  }

  /**
   * Inserts an exponentiation operator (xʸ) for power calculations.
   */
  function handlePower() {
    errorMessage = ''
    if (expression === '0') return
    const trimmed = expression.trim()
    const lastChar = trimmed[trimmed.length - 1]
    if (
      (lastChar >= '0' && lastChar <= '9') ||
      lastChar in SUPERSCRIPT_MAP ||
      lastChar === ')'
    ) {
      isExponentMode = true
      expression += 'ʸ'
    }
  }

  /**
   * Inserts a balanced opening or closing parenthesis into the expression buffer.
   *
   * @param p - Parenthesis character '(' or ')'.
   */
  function handleParenthesis(p: '(' | ')') {
    errorMessage = ''
    if (p === '(') {
      if (expression === '0') {
        expression = '('
      } else {
        const trimmed = expression.trim()
        const lastChar = trimmed[trimmed.length - 1]
        if (lastChar >= '0' && lastChar <= '9') {
          expression += ' × ('
        } else {
          expression += '('
        }
      }
    } else {
      const openCount = (expression.match(/\(/g) || []).length
      const closeCount = (expression.match(/\)/g) || []).length
      if (openCount > closeCount) {
        expression += ')'
      }
    }
  }

  /**
   * Toggles the algebraic sign (positive/negative) of the current active operand.
   */
  function handleToggleSign() {
    errorMessage = ''
    if (expression === '0') return
    if (expression.startsWith('-') || expression.startsWith('−')) {
      expression = expression.slice(1)
    } else {
      expression = `−${expression}`
    }
  }

  /**
   * Evaluates the full mathematical expression using operator precedence rules and displays result.
   */
  function handleEvaluate() {
    errorMessage = ''
    isExponentMode = false
    if (expression.endsWith('ʸ')) {
      expression = expression.slice(0, -1)
    }
    const outcome = evaluateExpression(expression)
    if (isCalculationError(outcome)) {
      errorMessage = outcome.error
    } else {
      historyFormula = `${expression} =`
      expression = outcome.value.toString()
    }
  }

  /**
   * Registers global keyboard listeners for numpad digits, operators, Enter, and Backspace.
   */
  function setupKeyboardListener() {
    /**
     * Processes intercepted keyboard events, routing keys to their corresponding calculator actions.
     *
     * @param e - Keyboard event from the window listener.
     */
    function onKeyDown(e: KeyboardEvent) {
      if (e.defaultPrevented) return

      if (e.key >= '0' && e.key <= '9') {
        e.preventDefault()
        handleDigit(e.key)
      } else if (e.key === '.') {
        e.preventDefault()
        handleDecimal()
      } else if (e.key === '+') {
        e.preventDefault()
        handleOperator('+')
      } else if (e.key === '-') {
        e.preventDefault()
        handleOperator('−')
      } else if (e.key === '*') {
        e.preventDefault()
        handleOperator('×')
      } else if (e.key === '/') {
        e.preventDefault()
        handleOperator('÷')
      } else if (e.key === '^') {
        e.preventDefault()
        handlePower()
      } else if (e.key === '(' || e.key === ')') {
        e.preventDefault()
        handleParenthesis(e.key)
      } else if (e.key === 'Enter' || e.key === '=') {
        e.preventDefault()
        handleEvaluate()
      } else if (e.key === 'Backspace') {
        e.preventDefault()
        handleBackspace()
      } else if (e.key === 'Escape') {
        e.preventDefault()
        handleClear()
      }
    }

    window.addEventListener('keydown', onKeyDown)
    return () => {
      window.removeEventListener('keydown', onKeyDown)
    }
  }

  $effect(setupKeyboardListener)
</script>

<div data-vinta="calculator-page" class="calculator-page">
  <!-- Display Panel -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="calc-display-card"
    onclick={handleCopyResult}
    use:islandTooltip={copiedIndicator
      ? $tStore('calculator.copied')
      : $tStore('calculator.click_to_copy')}
  >
    <div class="display-history-row">
      <span class="history-text">
        {historyFormula}
      </span>
      {#if copiedIndicator}
        <span class="live-preview-text">{$tStore('calculator.copied')}</span>
      {:else if livePreview}
        <span class="live-preview-text">{livePreview}</span>
      {/if}
    </div>
    <div class="display-main-row">
      {#if errorMessage}
        <span class="error-text">{errorMessage}</span>
      {:else}
        <span
          class="main-expression"
          class:scaled-down={expression.length > 12}
        >
          {expression}
        </span>
      {/if}
    </div>
  </div>

  <!-- Keypad: 6 rows x 4 columns -->
  <div class="calc-keypad">
    <!-- Row 1: Clear, Grouping, Backspace -->
    <button
      class="calc-btn action-btn"
      onclick={handleClear}
      use:islandTooltip={$tStore('calculator.all_clear')}
    >
      C
    </button>
    <button class="calc-btn action-btn" onclick={() => handleParenthesis('(')}>
      (
    </button>
    <button class="calc-btn action-btn" onclick={() => handleParenthesis(')')}>
      )
    </button>
    <button
      class="calc-btn func-btn backspace-btn"
      onclick={handleBackspace}
      use:islandTooltip={$tStore('calculator.backspace')}
    >
      <Delete size={17} />
    </button>

    <!-- Row 2: Math functions & Divide -->
    <button
      class="calc-btn func-btn"
      onclick={handleSquareRoot}
      use:islandTooltip={$tStore('calculator.square_root')}
    >
      √
    </button>
    <button
      class="calc-btn func-btn"
      onclick={handleRoot}
      use:islandTooltip={$tStore('calculator.nth_root')}
    >
      ⁿ√
    </button>
    <button
      class="calc-btn func-btn"
      onclick={handlePower}
      use:islandTooltip={$tStore('calculator.power')}
    >
      xʸ
    </button>
    <button class="calc-btn op-btn" onclick={() => handleOperator('÷')}>
      ÷
    </button>

    <!-- Row 3: 7, 8, 9, Multiply -->
    <button class="calc-btn num-btn" onclick={() => handleDigit('7')}>7</button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('8')}>8</button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('9')}>9</button>
    <button class="calc-btn op-btn" onclick={() => handleOperator('×')}>
      ×
    </button>

    <!-- Row 4: 4, 5, 6, Subtract -->
    <button class="calc-btn num-btn" onclick={() => handleDigit('4')}>4</button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('5')}>5</button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('6')}>6</button>
    <button class="calc-btn op-btn" onclick={() => handleOperator('−')}>
      −
    </button>

    <!-- Row 5: 1, 2, 3, Add -->
    <button class="calc-btn num-btn" onclick={() => handleDigit('1')}>1</button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('2')}>2</button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('3')}>3</button>
    <button class="calc-btn op-btn" onclick={() => handleOperator('+')}>
      +
    </button>

    <!-- Row 6: ±, 0, ., Equals -->
    <button class="calc-btn num-btn" onclick={handleToggleSign}> ± </button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('0')}>0</button>
    <button class="calc-btn num-btn" onclick={handleDecimal}>.</button>
    <button class="calc-btn equals-btn" onclick={handleEvaluate}> = </button>
  </div>
</div>

<style src="./styles/CalculatorPage.css"></style>
