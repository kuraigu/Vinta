<script lang="ts">
  import { Music2, Check, X, Pin } from 'lucide-svelte'
  import { islandTooltip } from '../services/islandTooltip'
  import { resolveAppBrand, nativeIconsStore } from '../services/iconService'
  import { pinnedMediaSourceStore } from '../store'
  import './styles/MediaSourcesDrawer.css'

  interface MediaSession {
    id: string
    sourceApp: string
    title?: string
    artist?: string
    isPlaying?: boolean
  }

  interface Props {
    sessions: MediaSession[]
    activeSessionId?: string
    activeTitle?: string
    activeSourceApp?: string
    onSelectSession: (sessionId: string) => void
    onClose: () => void
  }

  let {
    sessions,
    activeSessionId = '',
    activeTitle = '',
    activeSourceApp = '',
    onSelectSession,
    onClose
  }: Props = $props()

  let nativeIcons = $derived($nativeIconsStore)

  function getSessionBrand(sourceApp: string, sessionId: string) {
    void nativeIcons
    return resolveAppBrand(sourceApp, sessionId)
  }

  function handleTogglePin(sourceApp: string, e: MouseEvent) {
    e.stopPropagation()
    if ($pinnedMediaSourceStore === sourceApp) {
      $pinnedMediaSourceStore = null
    } else {
      $pinnedMediaSourceStore = sourceApp
      const match = sessions.find((s) => s.sourceApp === sourceApp)
      if (match) {
        onSelectSession(match.id)
      }
    }
  }
</script>

<div class="session-drawer-panel">
  <div class="session-drawer-header">
    <div class="session-header-title-group">
      <span class="session-header-title">Media Sources</span>
      <span class="session-count-tag">{sessions.length}</span>
    </div>
    <button
      class="session-close-btn"
      use:islandTooltip={{ text: 'Close Sources', align: 'right' }}
      aria-label="Close Sources"
      onclick={(e) => {
        e.stopPropagation()
        onClose()
      }}
    >
      <X size={13} />
    </button>
  </div>

  <div class="session-drawer-list">
    {#each sessions as s (s.id)}
      {@const isSelected =
        s.id === activeSessionId ||
        (Boolean(activeTitle) && s.title === activeTitle) ||
        (sessions.length === 1 && s.sourceApp === activeSourceApp)}
      {@const sBrand = getSessionBrand(s.sourceApp, s.id)}
      <div
        class="session-menu-item"
        class:is-selected={isSelected}
        onclick={(e) => {
          e.stopPropagation()
          onSelectSession(s.id)
        }}
        onkeydown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.stopPropagation()
            onSelectSession(s.id)
          }
        }}
        role="menuitem"
        tabindex="0"
      >
        <div class="session-item-icon">
          {#if sBrand.icon}
            <img
              src={sBrand.icon}
              alt={sBrand.name}
              class="session-app-img"
              draggable="false"
              decoding="async"
              onerror={(e) => {
                const el = e.currentTarget as HTMLImageElement
                el.style.display = 'none'
              }}
            />
          {:else}
            <Music2 size={13} class="session-item-fallback-icon" />
          {/if}
        </div>
        <div class="session-item-info">
          <div class="session-item-title" title={s.title || sBrand.name}>
            {s.title || sBrand.name}
          </div>
          <div class="session-item-subtitle" title={s.artist || sBrand.name}>
            {s.artist || sBrand.name}
          </div>
        </div>
        <div class="session-item-status">
          {#if isSelected}
            <Check size={13} class="selected-check" />
          {:else if s.isPlaying}
            <div class="live-playing-dot" use:islandTooltip={'Playing'}></div>
          {/if}
        </div>
        <button
          type="button"
          class="session-pin-btn"
          class:is-pinned={$pinnedMediaSourceStore === s.sourceApp}
          use:islandTooltip={{
            text:
              $pinnedMediaSourceStore === s.sourceApp
                ? 'Pinned Source (Click to unpin)'
                : 'Pin this source',
            position: 'top'
          }}
          onclick={(e) => handleTogglePin(s.sourceApp, e)}
          aria-label={$pinnedMediaSourceStore === s.sourceApp
            ? 'Unpin source'
            : 'Pin source'}
        >
          <Pin size={11} />
        </button>
      </div>
    {/each}
  </div>
</div>
