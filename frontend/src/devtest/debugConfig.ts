import { writable } from 'svelte/store'
import { IsTestMode } from '../../bindings/Vinta/backend/app.js'

function getInitialDebugFlag(key: string, defaultValue: boolean): boolean {
  if (typeof window === 'undefined') return defaultValue
  try {
    const val = localStorage.getItem(key)
    return val !== null ? JSON.parse(val) : defaultValue
  } catch {
    return defaultValue
  }
}

function createDebugFlag(key: string, defaultValue: boolean) {
  const store = writable<boolean>(getInitialDebugFlag(key, defaultValue))
  if (typeof window !== 'undefined') {
    store.subscribe((val) => {
      try {
        localStorage.setItem(key, JSON.stringify(val))
      } catch {
        // Ignore storage error
      }
    })
  }
  return store
}

export const debugBoundsOverlay = createDebugFlag(
  'vinta_debug_bounds_overlay',
  false
)
export const debugLogs = createDebugFlag('vinta_debug_logs', false)
export const debugClassInspector = createDebugFlag(
  'vinta_debug_class_inspector',
  false
)

if (typeof window !== 'undefined') {
  const isViteDev = Boolean(import.meta.env?.DEV)

  function initDebug() {
    ;(window as unknown as Record<string, unknown>).__VINTA_DEBUG__ = {
      showBounds: (enable = true) => {
        debugBoundsOverlay.set(enable)
        console.log(`[Vinta:Debug] Bounds overlay: ${enable ? 'ON' : 'OFF'}`)
      },
      enableLogs: (enable = true) => {
        debugLogs.set(enable)
        console.log(`[Vinta:Debug] Verbose logs: ${enable ? 'ON' : 'OFF'}`)
      },
      showInspector: (enable = true) => {
        debugClassInspector.set(enable)
        console.log(`[Vinta:Debug] Class inspector: ${enable ? 'ON' : 'OFF'}`)
      },
      toggleBounds: () => {
        let current = false
        debugBoundsOverlay.subscribe((v) => (current = v))()
        const next = !current
        debugBoundsOverlay.set(next)
        console.log(`[Vinta:Debug] Bounds overlay: ${next ? 'ON' : 'OFF'}`)
        return next
      },
      toggleLogs: () => {
        let current = false
        debugLogs.subscribe((v) => (current = v))()
        const next = !current
        debugLogs.set(next)
        console.log(`[Vinta:Debug] Verbose logs: ${next ? 'ON' : 'OFF'}`)
        return next
      },
      toggleInspector: () => {
        let current = false
        debugClassInspector.subscribe((v) => (current = v))()
        const next = !current
        debugClassInspector.set(next)
        console.log(`[Vinta:Debug] Class inspector: ${next ? 'ON' : 'OFF'}`)
        return next
      },
      status: () => {
        let b = false
        let l = false
        let i = false
        debugBoundsOverlay.subscribe((v) => (b = v))()
        debugLogs.subscribe((v) => (l = v))()
        debugClassInspector.subscribe((v) => (i = v))()
        console.table({
          'Bounds Overlay': b,
          'Verbose Logs': l,
          'Class Inspector': i
        })
      }
    }

    // Real-time keyboard shortcuts
    window.addEventListener('keydown', (e: KeyboardEvent) => {
      // Ctrl + Shift + D or Ctrl + Alt + B -> Real-time Bounds Overlay Toggle
      if (
        (e.ctrlKey && e.shiftKey && e.code === 'KeyD') ||
        (e.ctrlKey && e.altKey && e.code === 'KeyB')
      ) {
        e.preventDefault()
        let current = false
        debugBoundsOverlay.subscribe((v) => (current = v))()
        debugBoundsOverlay.set(!current)
      }

      // Ctrl + Shift + L or Ctrl + Alt + L -> Real-time Verbose Logs Toggle
      if (
        (e.ctrlKey && e.shiftKey && e.code === 'KeyL') ||
        (e.ctrlKey && e.altKey && e.code === 'KeyL')
      ) {
        e.preventDefault()
        let current = false
        debugLogs.subscribe((v) => (current = v))()
        debugLogs.set(!current)
      }

      // Ctrl + Shift + I or Ctrl + Alt + I -> Real-time Class Inspector Toggle
      if (
        (e.ctrlKey && e.shiftKey && e.code === 'KeyI') ||
        (e.ctrlKey && e.altKey && e.code === 'KeyI')
      ) {
        e.preventDefault()
        let current = false
        debugClassInspector.subscribe((v) => (current = v))()
        debugClassInspector.set(!current)
      }
    })
  }

  if (isViteDev) {
    initDebug()
  } else if (typeof IsTestMode === 'function') {
    IsTestMode()
      .then((testMode) => {
        if (testMode) {
          initDebug()
        }
      })
      .catch(() => {})
  }
}
