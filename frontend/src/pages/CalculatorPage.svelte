<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 220, h: 36, r: 16 },
    expanded: { w: 340, h: 470, r: 28 }
  }
</script>

<script lang="ts">
  import { Delete } from 'lucide-svelte'
  import BackButton from '../components/BackButton.svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import {
    evaluateExpression,
    isCalculationError
  } from '../utils/calculatorEngine'

  let expression = $state('0')
  let historyFormula = $state('')
  let errorMessage = $state('')
  let isExponentMode = $state(false)

  let livePreview = $derived.by(() => {
    if (!expression || expression === '0' || errorMessage) return ''
    const res = evaluateExpression(expression)
    if (!isCalculationError(res) && res.display !== expression) {
      return `= ${res.display}`
    }
    return ''
  })

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

  function toSuperscript(numStr: string): string {
    return numStr
      .split('')
      .map((c) => SUPERSCRIPT_MAP[c] || c)
      .join('')
  }

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

  function handleDecimal() {
    errorMessage = ''
    // Check if current active number chunk already contains decimal
    const parts = expression.split(/[\s+−×÷^()]+/)
    const currentPart = parts[parts.length - 1] || ''
    if (!currentPart.includes('.')) {
      expression += '.'
    }
  }

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

  function handleToggleSign() {
    errorMessage = ''
    if (expression === '0') return
    if (expression.startsWith('-') || expression.startsWith('−')) {
      expression = expression.slice(1)
    } else {
      expression = `−${expression}`
    }
  }

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

  function setupKeyboardListener() {
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
  <!-- Uniform Reusable Header -->
  <div class="calc-header">
    <BackButton />
    <span class="header-title">Calculator</span>
    <div class="header-spacer"></div>
  </div>

  <!-- Display Panel -->
  <div class="calc-display-card">
    <div class="display-history-row">
      <span class="history-text" title={historyFormula || ''}>
        {historyFormula}
      </span>
      {#if livePreview}
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
          title={expression}
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
      use:islandTooltip={'All Clear'}
    >
      C
    </button>
    <button
      class="calc-btn action-btn"
      onclick={() => handleParenthesis('(')}
      use:islandTooltip={'Open Parenthesis'}
    >
      (
    </button>
    <button
      class="calc-btn action-btn"
      onclick={() => handleParenthesis(')')}
      use:islandTooltip={'Close Parenthesis'}
    >
      )
    </button>
    <button
      class="calc-btn func-btn backspace-btn"
      onclick={handleBackspace}
      use:islandTooltip={'Backspace'}
    >
      <Delete size={17} />
    </button>

    <!-- Row 2: Math functions & Divide -->
    <button
      class="calc-btn func-btn"
      onclick={handleSquareRoot}
      use:islandTooltip={'Square root'}
    >
      √
    </button>
    <button
      class="calc-btn func-btn"
      onclick={handleRoot}
      use:islandTooltip={'Nth root (e.g. 3√27)'}
    >
      ⁿ√
    </button>
    <button
      class="calc-btn func-btn"
      onclick={handlePower}
      use:islandTooltip={'Power (xʸ)'}
    >
      xʸ
    </button>
    <button
      class="calc-btn op-btn"
      onclick={() => handleOperator('÷')}
      use:islandTooltip={'Divide'}
    >
      ÷
    </button>

    <!-- Row 3: 7, 8, 9, Multiply -->
    <button class="calc-btn num-btn" onclick={() => handleDigit('7')}>7</button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('8')}>8</button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('9')}>9</button>
    <button
      class="calc-btn op-btn"
      onclick={() => handleOperator('×')}
      use:islandTooltip={'Multiply'}
    >
      ×
    </button>

    <!-- Row 4: 4, 5, 6, Subtract -->
    <button class="calc-btn num-btn" onclick={() => handleDigit('4')}>4</button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('5')}>5</button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('6')}>6</button>
    <button
      class="calc-btn op-btn"
      onclick={() => handleOperator('−')}
      use:islandTooltip={'Subtract'}
    >
      −
    </button>

    <!-- Row 5: 1, 2, 3, Add -->
    <button class="calc-btn num-btn" onclick={() => handleDigit('1')}>1</button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('2')}>2</button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('3')}>3</button>
    <button
      class="calc-btn op-btn"
      onclick={() => handleOperator('+')}
      use:islandTooltip={'Add'}
    >
      +
    </button>

    <!-- Row 6: ±, 0, ., Equals -->
    <button
      class="calc-btn num-btn"
      onclick={handleToggleSign}
      use:islandTooltip={'Toggle sign'}
    >
      ±
    </button>
    <button class="calc-btn num-btn" onclick={() => handleDigit('0')}>0</button>
    <button class="calc-btn num-btn" onclick={handleDecimal}>.</button>
    <button
      class="calc-btn equals-btn"
      onclick={handleEvaluate}
      use:islandTooltip={'Calculate'}
    >
      =
    </button>
  </div>
</div>

<style src="./styles/CalculatorPage.css"></style>
