<script lang="ts" module>
  import type { IslandConfig } from '../store'

  export const config: IslandConfig = {
    minimal: { w: 10, h: 10, r: 5 },
    expanded: { w: 340, h: 480, r: 28 },
    borderColor: 'var(--c-border)',
    glowColor: 'transparent'
  }
</script>

<script lang="ts">
  import { onMount, getContext } from 'svelte'
  import { ArrowLeft, Delete, History } from 'lucide-svelte'
  import AppsPage, { config as appsConfig } from './AppsPage.svelte'
  import {
    evaluateExpression,
    formatNumberDisplay,
    isCalculationError
  } from '../utils/calculatorEngine'

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  let expression = $state<string>('')
  let displayValue = $state<string>('0')
  let historyExpr = $state<string>('')
  let historyResult = $state<string>('')
  let isResultEvaluated = $state<boolean>(false)

  function handleGoBack() {
    if (navigate) {
      navigate(AppsPage, appsConfig, true)
    }
  }

  function handleInput(char: string) {
    if (isResultEvaluated) {
      if ('0123456789.'.includes(char)) {
        expression = char === '.' ? '0.' : char
        displayValue = expression
        isResultEvaluated = false
        return
      } else {
        // Continue operating on the previous result
        expression = displayValue
        isResultEvaluated = false
      }
    }

    if (char === '.') {
      const parts = expression.split(/[+\-×÷%()]/)
      const currentSegment = parts[parts.length - 1]
      if (currentSegment.includes('.')) return
      if (!currentSegment || currentSegment === '') {
        expression += '0.'
        displayValue = expression
        return
      }
    }

    expression += char
    displayValue = expression
  }

  function handleOperator(op: string) {
    if (isResultEvaluated) {
      expression = displayValue
      isResultEvaluated = false
    }

    if (!expression) {
      if (op === '−') {
        expression = '−'
        displayValue = '−'
      }
      return
    }

    const lastChar = expression[expression.length - 1]
    if (['+', '−', '×', '÷'].includes(lastChar)) {
      expression = expression.slice(0, -1) + op
    } else {
      expression += op
    }
    displayValue = expression
  }

  function handleToggleSign() {
    if (isResultEvaluated) {
      const num = parseFloat(displayValue.replace(/,/g, ''))
      if (!isNaN(num)) {
        const toggled = -num
        displayValue = formatNumberDisplay(toggled)
        expression = toggled.toString()
        return
      }
    }

    if (!expression) {
      expression = '−'
      displayValue = '−'
      return
    }

    if (expression.startsWith('−')) {
      expression = expression.slice(1)
    } else if (expression.startsWith('-')) {
      expression = expression.slice(1)
    } else {
      expression = '−' + expression
    }
    displayValue = expression
  }

  function handleBackspace() {
    if (isResultEvaluated) {
      expression = ''
      displayValue = '0'
      isResultEvaluated = false
      return
    }

    if (expression.length > 0) {
      expression = expression.slice(0, -1)
      displayValue = expression.length > 0 ? expression : '0'
    }
  }

  function handleClear() {
    if (expression || displayValue !== '0') {
      expression = ''
      displayValue = '0'
      isResultEvaluated = false
    } else {
      historyExpr = ''
      historyResult = ''
    }
  }

  function handleCalculate() {
    if (!expression.trim()) return

    const outcome = evaluateExpression(expression)
    if (isCalculationError(outcome)) {
      displayValue = outcome.error
      isResultEvaluated = true
      return
    }

    historyExpr = expression
    historyResult = outcome.display
    displayValue = outcome.display
    expression = outcome.value.toString()
    isResultEvaluated = true
  }

  function handleRecallHistory() {
    if (historyResult) {
      expression = historyResult.replace(/,/g, '')
      displayValue = historyResult
      isResultEvaluated = false
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (
      e.target instanceof HTMLInputElement ||
      e.target instanceof HTMLTextAreaElement
    ) {
      return
    }

    const key = e.key

    if (key >= '0' && key <= '9') {
      handleInput(key)
      e.preventDefault()
    } else if (key === '.') {
      handleInput('.')
      e.preventDefault()
    } else if (key === '+') {
      handleOperator('+')
      e.preventDefault()
    } else if (key === '-') {
      handleOperator('−')
      e.preventDefault()
    } else if (key === '*') {
      handleOperator('×')
      e.preventDefault()
    } else if (key === '/') {
      handleOperator('÷')
      e.preventDefault()
    } else if (key === '%') {
      handleOperator('%')
      e.preventDefault()
    } else if (key === '(' || key === ')') {
      handleInput(key)
      e.preventDefault()
    } else if (key === 'Enter' || key === '=') {
      handleCalculate()
      e.preventDefault()
    } else if (key === 'Backspace') {
      handleBackspace()
      e.preventDefault()
    } else if (key === 'Escape') {
      handleClear()
      e.preventDefault()
    }
  }

  onMount(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
    }
  })
</script>

<div class="calc-page">
  <!-- Header Bar -->
  <div class="calc-header">
    <button class="icon-btn" onclick={handleGoBack} title="Back to Apps">
      <ArrowLeft size={16} />
    </button>
    <div class="calc-title">Calculator</div>
    <button
      class="icon-btn"
      onclick={handleClear}
      title="Clear Expression (Escape)"
    >
      <Delete size={15} />
    </button>
  </div>

  <!-- Display Section -->
  <div class="calc-screen">
    {#if historyExpr}
      <button
        class="calc-history"
        onclick={handleRecallHistory}
        title="Recall previous calculation"
      >
        <History size={11} />
        <span>{historyExpr} = {historyResult}</span>
      </button>
    {:else}
      <div class="calc-history-placeholder"></div>
    {/if}
    <div class="calc-expression">{expression || ' '}</div>
    <div class="calc-result" title={displayValue}>{displayValue}</div>
  </div>

  <!-- Keypad Grid -->
  <div class="calc-keypad">
    <button class="key action-key" onclick={handleClear}>C</button>
    <button class="key action-key" onclick={() => handleInput('(')}>(</button>
    <button class="key action-key" onclick={() => handleInput(')')}>)</button>
    <button class="key op-key" onclick={() => handleOperator('÷')}>÷</button>

    <button class="key num-key" onclick={() => handleInput('7')}>7</button>
    <button class="key num-key" onclick={() => handleInput('8')}>8</button>
    <button class="key num-key" onclick={() => handleInput('9')}>9</button>
    <button class="key op-key" onclick={() => handleOperator('×')}>×</button>

    <button class="key num-key" onclick={() => handleInput('4')}>4</button>
    <button class="key num-key" onclick={() => handleInput('5')}>5</button>
    <button class="key num-key" onclick={() => handleInput('6')}>6</button>
    <button class="key op-key" onclick={() => handleOperator('−')}>−</button>

    <button class="key num-key" onclick={() => handleInput('1')}>1</button>
    <button class="key num-key" onclick={() => handleInput('2')}>2</button>
    <button class="key num-key" onclick={() => handleInput('3')}>3</button>
    <button class="key op-key" onclick={() => handleOperator('+')}>+</button>

    <button class="key action-key" onclick={handleToggleSign}>±</button>
    <button class="key num-key" onclick={() => handleInput('0')}>0</button>
    <button class="key num-key" onclick={() => handleInput('.')}>.</button>
    <button class="key equals-key" onclick={handleCalculate}>=</button>
  </div>
</div>

<style src="./styles/CalculatorPage.css"></style>
