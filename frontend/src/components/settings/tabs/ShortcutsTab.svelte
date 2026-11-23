<!--
  @component ShortcutsTab
  @description Global hotkey manager. Allows users to record and bind keyboard shortcuts
  for toggling the island, opening action center, switching media, and navigating pages.
-->
<script lang="ts">
  import {
    Keyboard,
    RotateCcw,
    X,
    Check,
    ShieldCheck,
    FileText,
    Sliders,
    ArrowLeftRight,
    CornerDownLeft,
    Layers
  } from 'lucide-svelte'
  import BentoCard from '../../BentoCard.svelte'
  import SettingItem from '../SettingItem.svelte'
  import SettingSectionHeader from '../SettingSectionHeader.svelte'
  import SettingBanner from '../SettingBanner.svelte'
  import { islandTooltip } from '../../../services/islandTooltip'
  import {
    RegisterNamedGlobalShortcut,
    UnregisterNamedGlobalShortcut
  } from '../../../../bindings/Vinta/backend/app.js'
  import {
    globalHotkeyStore,
    quickNotesHotkeyStore,
    modeSwitchHotkeyStore,
    tStore
  } from '../../../store'

  type ShortcutDef = {
    id: 'toggle' | 'quick_notes' | 'mode_switch'
    label: string
    description: string
    defaultKey: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon: any
  }

  const SHORTCUT_DEFS: ShortcutDef[] = $derived([
    {
      id: 'toggle',
      label: $tStore('settings.shortcut_toggle_label'),
      description: $tStore('settings.shortcut_toggle_desc'),
      defaultKey: 'Alt+Shift+V',
      icon: Keyboard
    },
    {
      id: 'quick_notes',
      label: $tStore('settings.shortcut_notes_label'),
      description: $tStore('settings.shortcut_notes_desc'),
      defaultKey: 'Alt+Shift+N',
      icon: FileText
    },
    {
      id: 'mode_switch',
      label: $tStore('settings.shortcut_mode_label'),
      description: $tStore('settings.shortcut_mode_desc'),
      defaultKey: 'Alt+Shift+D',
      icon: Sliders
    }
  ])

  const SYSTEM_RESERVED_SHORTCUTS: Record<string, string> = {
    'SUPER+L': 'Lock Computer',
    'CTRL+ALT+DELETE': 'Windows Security',
    'SUPER+D': 'Show Desktop',
    'ALT+TAB': 'Task Switcher',
    'ALT+F4': 'Close Application',
    'CTRL+SHIFT+ESCAPE': 'Task Manager',
    'SUPER+TAB': 'Task View',
    'SUPER+X': 'Windows Power Menu',
    'SUPER+E': 'File Explorer',
    'SUPER+R': 'Run Dialog'
  }

  const COMMON_APP_COLLISIONS: Record<string, string> = {
    'CTRL+SHIFT+N': 'Chrome/Edge Incognito & Explorer New Folder',
    'CTRL+SHIFT+V': 'Paste Plain Text in browsers, Slack, & editors',
    'CTRL+SHIFT+D': 'Bookmark All Tabs in Chrome & Edge',
    'CTRL+SHIFT+T': 'Reopen Closed Tab in browsers',
    'CTRL+SHIFT+W': 'Close Window in browsers',
    'CTRL+SHIFT+I': 'Developer Tools in browsers',
    'CTRL+SHIFT+J': 'Console / Downloads in Chrome',
    'CTRL+SHIFT+P': 'Command Palette in VS Code & browsers',
    'CTRL+SHIFT+F': 'Global Search in VS Code & editors',
    'CTRL+SHIFT+B': 'Toggle Bookmarks Bar in Chrome',
    'CTRL+SHIFT+O': 'Bookmark Manager in Chrome',
    'CTRL+SHIFT+DELETE': 'Clear Browsing Data in browsers'
  }

  let listeningId = $state<string | null>(null)
  let activeModifiers = $state<string[]>([])
  let statusMessages = $state<Record<string, string>>({})
  let statusTypes = $state<
    Record<string, 'error' | 'warning' | 'success' | 'info'>
  >({})
  let savedFlashes = $state<Record<string, boolean>>({})

  /**
   * Helper to fetch the current value of a shortcut by ID.
   *
   * @param id - Shortcut identifier.
   * @returns Current shortcut accelerator string.
   */
  function getShortcutValue(id: string): string {
    if (id === 'toggle') return $globalHotkeyStore
    if (id === 'quick_notes') return $quickNotesHotkeyStore
    if (id === 'mode_switch') return $modeSwitchHotkeyStore
    return ''
  }

  /**
   * Helper to set the store value of a shortcut by ID.
   *
   * @param id - Shortcut identifier.
   * @param val - New accelerator value.
   */
  function setShortcutValue(id: string, val: string) {
    if (id === 'toggle') $globalHotkeyStore = val
    else if (id === 'quick_notes') $quickNotesHotkeyStore = val
    else if (id === 'mode_switch') $modeSwitchHotkeyStore = val
  }

  /**
   * Validates shortcut combinations against Windows system keys and other Vinta bindings.
   *
   * @param accel - Tested accelerator string.
   * @param currentId - The action ID being configured.
   * @returns Collision message or null if clear.
   */
  function checkCollision(accel: string, currentId: string): string | null {
    if (!accel.trim()) return null
    const normalized = accel.toUpperCase()

    if (SYSTEM_RESERVED_SHORTCUTS[normalized]) {
      return $tStore('settings.shortcut_warn_windows', {
        name: SYSTEM_RESERVED_SHORTCUTS[normalized]
      })
    }

    if (COMMON_APP_COLLISIONS[normalized]) {
      return $tStore('settings.shortcut_warn_app', {
        name: COMMON_APP_COLLISIONS[normalized]
      })
    }

    if (
      currentId !== 'toggle' &&
      $globalHotkeyStore &&
      $globalHotkeyStore.toUpperCase() === normalized
    ) {
      return $tStore('settings.shortcut_warn_assigned', {
        name: $tStore('settings.shortcut_toggle_label')
      })
    }
    if (
      currentId !== 'quick_notes' &&
      $quickNotesHotkeyStore &&
      $quickNotesHotkeyStore.toUpperCase() === normalized
    ) {
      return $tStore('settings.shortcut_warn_assigned', {
        name: $tStore('settings.shortcut_notes_label')
      })
    }
    if (
      currentId !== 'mode_switch' &&
      $modeSwitchHotkeyStore &&
      $modeSwitchHotkeyStore.toUpperCase() === normalized
    ) {
      return $tStore('settings.shortcut_warn_assigned', {
        name: $tStore('settings.shortcut_mode_label')
      })
    }

    return null
  }

  /**
   * Registers or removes a named shortcut with the backend.
   *
   * @param id - Shortcut identifier.
   * @param accel - Target accelerator string.
   */
  async function applyShortcut(id: string, accel: string) {
    try {
      const trimmed = accel.trim()

      if (!trimmed) {
        if (typeof UnregisterNamedGlobalShortcut === 'function') {
          await UnregisterNamedGlobalShortcut(id)
        }
        setShortcutValue(id, '')
        statusMessages = {
          ...statusMessages,
          [id]: $tStore('settings.shortcut_msg_unassigned')
        }
        statusTypes = { ...statusTypes, [id]: 'info' }
        return
      }

      // Check collision
      const collisionWarning = checkCollision(trimmed, id)
      if (collisionWarning) {
        statusMessages = { ...statusMessages, [id]: collisionWarning }
        statusTypes = { ...statusTypes, [id]: 'warning' }
      }

      if (typeof RegisterNamedGlobalShortcut === 'function') {
        const ok = await RegisterNamedGlobalShortcut(id, trimmed)
        if (ok) {
          setShortcutValue(id, trimmed)
          savedFlashes = { ...savedFlashes, [id]: true }
          if (!collisionWarning) {
            statusMessages = {
              ...statusMessages,
              [id]: $tStore('settings.shortcut_msg_registered')
            }
            statusTypes = { ...statusTypes, [id]: 'success' }
          }
          setTimeout(() => {
            savedFlashes = { ...savedFlashes, [id]: false }
          }, 1200)
        } else {
          statusMessages = {
            ...statusMessages,
            [id]: $tStore('settings.shortcut_msg_rejected')
          }
          statusTypes = { ...statusTypes, [id]: 'error' }
        }
      } else {
        setShortcutValue(id, trimmed)
      }
    } catch (err: unknown) {
      statusMessages = {
        ...statusMessages,
        [id]: err instanceof Error ? err.message : String(err)
      }
      statusTypes = { ...statusTypes, [id]: 'error' }
    }
  }

  function startListening(id: string) {
    listeningId = id
    activeModifiers = []
    statusMessages = {
      ...statusMessages,
      [id]: 'Press shortcut keys (Esc to cancel, Backspace to clear)'
    }
    statusTypes = { ...statusTypes, [id]: 'info' }
  }

  function stopListening() {
    if (listeningId) {
      const activeId = listeningId
      if (statusMessages[activeId]?.includes('Press shortcut keys')) {
        statusMessages = { ...statusMessages, [activeId]: '' }
      }
    }
    listeningId = null
    activeModifiers = []
  }

  /**
   * Captures keyboard input at the window level during shortcut assignment.
   * Intercepts keys during the capture phase to prevent conflicts with
   * island navigation failsafes and parent containers.
   */
  function setupShortcutCapture() {
    if (!listeningId) return

    const currentTargetId = listeningId

    function onWindowKeyDown(e: KeyboardEvent) {
      e.preventDefault()
      e.stopPropagation()
      e.stopImmediatePropagation()

      if (e.key === 'Escape') {
        stopListening()
        return
      }

      if (e.key === 'Backspace' || e.key === 'Delete') {
        void applyShortcut(currentTargetId, '')
        stopListening()
        return
      }

      const mods: string[] = []
      if (e.ctrlKey) mods.push('Ctrl')
      if (e.altKey) mods.push('Alt')
      if (e.shiftKey) mods.push('Shift')
      if (e.metaKey) mods.push('Super')

      const isModifierKey = ['Control', 'Alt', 'Shift', 'Meta'].includes(e.key)
      if (isModifierKey) {
        activeModifiers = mods
        return
      }

      let keyName = e.key.toUpperCase()
      if (keyName === ' ') keyName = 'Space'

      if (mods.length === 0) {
        statusMessages = {
          ...statusMessages,
          [currentTargetId]: $tStore('settings.shortcut_err_modifier')
        }
        statusTypes = { ...statusTypes, [currentTargetId]: 'error' }
        return
      }

      const combination = [...mods, keyName].join('+')
      void applyShortcut(currentTargetId, combination)
      stopListening()
    }

    function onWindowKeyUp(e: KeyboardEvent) {
      e.preventDefault()
      e.stopPropagation()
      e.stopImmediatePropagation()

      const mods: string[] = []
      if (e.ctrlKey) mods.push('Ctrl')
      if (e.altKey) mods.push('Alt')
      if (e.shiftKey) mods.push('Shift')
      if (e.metaKey) mods.push('Super')
      activeModifiers = mods
    }

    function onWindowPointerDown(e: PointerEvent) {
      const target = e.target as HTMLElement | null
      if (!target?.closest('.keybind-binder-btn')) {
        stopListening()
      }
    }

    window.addEventListener('keydown', onWindowKeyDown, { capture: true })
    window.addEventListener('keyup', onWindowKeyUp, { capture: true })
    window.addEventListener('pointerdown', onWindowPointerDown, {
      capture: true
    })

    return () => {
      window.removeEventListener('keydown', onWindowKeyDown, { capture: true })
      window.removeEventListener('keyup', onWindowKeyUp, { capture: true })
      window.removeEventListener('pointerdown', onWindowPointerDown, {
        capture: true
      })
    }
  }

  $effect(setupShortcutCapture)
</script>

<div class="shortcuts-tab">
  <!-- System Isolation Notice -->
  <SettingBanner
    variant="blue"
    icon={ShieldCheck}
    title={$tStore('settings.shortcut_isolation_title')}
    description={$tStore('settings.shortcut_isolation_desc')}
  />

  <div class="divider"></div>

  <!-- Global Shortcuts Section -->
  <SettingSectionHeader title={$tStore('settings.shortcut_global_section')} />

  <BentoCard variant="group">
    {#each SHORTCUT_DEFS as sc (sc.id)}
      {@const currentVal = getShortcutValue(sc.id)}
      {@const keyCaps =
        currentVal && currentVal.trim() ? currentVal.split('+') : []}
      {@const isCurrentListening = listeningId === sc.id}
      {@const isFlash = savedFlashes[sc.id]}
      {@const statusMsg = statusMessages[sc.id]}
      {@const statusType = statusTypes[sc.id]}

      <SettingItem
        label={sc.label}
        description={sc.description}
        icon={sc.icon}
        flexCol={true}
      >
        <div class="keybind-control-row">
          <button
            type="button"
            class="keybind-binder-btn"
            class:is-listening={isCurrentListening}
            class:is-saved={isFlash}
            onclick={() => {
              if (isCurrentListening) {
                stopListening()
              } else {
                startListening(sc.id)
              }
            }}
            aria-label={$tStore('settings.shortcut_rebind_aria', {
              label: sc.label
            })}
          >
            {#if isCurrentListening}
              <span class="listening-indicator">
                <span class="pulse-dot"></span>
                {#if activeModifiers.length > 0}
                  <span class="key-caps-list">
                    {#each activeModifiers as mod, idx (mod)}
                      <kbd class="key-cap active">{mod}</kbd>
                      {#if idx < activeModifiers.length - 1}
                        <span class="key-plus">+</span>
                      {/if}
                    {/each}
                    <span class="key-plus">+</span>
                    <span>{$tStore('settings.shortcut_press_key')}</span>
                  </span>
                {:else}
                  <span>{$tStore('settings.shortcut_press_keys')}</span>
                {/if}
              </span>
            {:else if isFlash}
              <span class="listening-indicator" style="color: #4ade80;">
                <Check size={13} />
                <span>{$tStore('settings.shortcut_saved')}</span>
              </span>
            {:else if keyCaps.length > 0}
              <span class="key-caps-list">
                {#each keyCaps as key, idx (key)}
                  <kbd class="key-cap">{key}</kbd>
                  {#if idx < keyCaps.length - 1}
                    <span class="key-plus">+</span>
                  {/if}
                {/each}
              </span>
            {:else}
              <span class="unassigned-label"
                >{$tStore('settings.shortcut_not_bound')}</span
              >
            {/if}
          </button>

          {#if currentVal}
            <button
              type="button"
              class="keybind-action-btn danger"
              use:islandTooltip={{
                text: $tStore('settings.shortcut_unbind'),
                position: 'top'
              }}
              onclick={() => void applyShortcut(sc.id, '')}
              aria-label={$tStore('settings.shortcut_unbind_aria', {
                label: sc.label
              })}
            >
              <X size={14} />
            </button>
          {/if}

          {#if currentVal !== sc.defaultKey}
            <button
              type="button"
              class="keybind-action-btn"
              use:islandTooltip={{
                text: $tStore('settings.shortcut_reset_to', {
                  key: sc.defaultKey
                }),
                position: 'top'
              }}
              onclick={() => void applyShortcut(sc.id, sc.defaultKey)}
              aria-label={$tStore('settings.shortcut_reset_aria', {
                label: sc.label
              })}
            >
              <RotateCcw size={13} />
            </button>
          {/if}
        </div>

        {#if statusMsg}
          <div
            class="status-message"
            class:is-error={statusType === 'error'}
            class:is-warning={statusType === 'warning'}
            class:is-success={statusType === 'success'}
          >
            {statusMsg}
          </div>
        {/if}
      </SettingItem>
    {/each}
  </BentoCard>

  <div class="divider"></div>

  <!-- Island Navigation Section -->
  <SettingSectionHeader title={$tStore('settings.shortcut_nav_section')} />

  <BentoCard variant="group">
    <SettingItem
      label={$tStore('settings.shortcut_nav_dismiss')}
      description={$tStore('settings.shortcut_nav_dismiss_desc')}
      icon={Keyboard}
    >
      <div class="shortcut-keys">
        <kbd class="key-badge">Esc</kbd>
      </div>
    </SettingItem>

    <SettingItem
      label={$tStore('settings.shortcut_nav_cycle_widgets')}
      description={$tStore('settings.shortcut_nav_cycle_widgets_desc')}
      icon={Layers}
    >
      <div class="shortcut-keys">
        <kbd class="key-badge">Tab</kbd>
        <span class="key-separator">/</span>
        <kbd class="key-badge">Shift</kbd>
        <kbd class="key-badge">Tab</kbd>
      </div>
    </SettingItem>

    <SettingItem
      label={$tStore('settings.shortcut_nav_activate')}
      description={$tStore('settings.shortcut_nav_activate_desc')}
      icon={CornerDownLeft}
    >
      <div class="shortcut-keys">
        <kbd class="key-badge">Space</kbd>
        <span class="key-separator">{$tStore('settings.shortcut_nav_or')}</span>
        <kbd class="key-badge">Enter</kbd>
      </div>
    </SettingItem>

    <SettingItem
      label={$tStore('settings.shortcut_nav_gallery')}
      description={$tStore('settings.shortcut_nav_gallery_desc')}
      icon={ArrowLeftRight}
    >
      <div class="shortcut-keys">
        <kbd class="key-badge">←</kbd>
        <kbd class="key-badge">→</kbd>
      </div>
    </SettingItem>
  </BentoCard>
</div>

<style src="../../styles/settings/ShortcutsTab.css"></style>
