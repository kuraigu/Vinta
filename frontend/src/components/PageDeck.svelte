<!--
  @component PageDeck
  @description Keep-alive LRU page pool that preserves mounted page DOM across navigation
  switches so changing pages does not destroy or remount Svelte component trees.

  IMPORTANT ARCHITECTURE NOTE:
  Do NOT remove this keep-alive pool in favor of conditional `{#if}` mounting. Pages such as
  HomePage, SettingsPage, CalendarPage, and NotesPage are intentionally kept warm in the DOM
  (hidden via `display: none` and `inert`) for instant 0ms navigation switches and preserved
  scroll/form state. Memory stays bounded via the LRU eviction cap (`MAX_WARM_PAGES`).
-->
<script lang="ts">
  import { untrack, type Component } from 'svelte'
  import Notification from './Notification.svelte'
  import OnboardingPage from '../pages/OnboardingPage.svelte'
  import HomePage from '../pages/HomePage.svelte'

  interface Props {
    activePage?: Component
    expanded?: boolean
  }

  let { activePage, expanded = false }: Props = $props()

  const MAX_WARM_PAGES = 6
  let mountedPages = $state<Component[]>([])
  let lruOrder: Component[] = []

  function isTransientPage(page: Component | undefined): boolean {
    return !page || page === Notification || page === OnboardingPage
  }

  function syncMountedPagePoolEffect() {
    const current = activePage
    if (!current || isTransientPage(current)) return
    untrack(() => {
      lruOrder = [...lruOrder.filter((p) => p !== current), current]
      if (!mountedPages.includes(current)) {
        const nextMounted = [...mountedPages, current]
        while (nextMounted.length > MAX_WARM_PAGES) {
          const victim = lruOrder.find((p) => p !== current && p !== HomePage)
          if (!victim) break
          lruOrder = lruOrder.filter((p) => p !== victim)
          const idx = nextMounted.indexOf(victim)
          if (idx >= 0) nextMounted.splice(idx, 1)
        }
        mountedPages = nextMounted
      }
    })
  }

  $effect(syncMountedPagePoolEffect)
</script>

{#each mountedPages as PageComponent (PageComponent)}
  {@const isCurrent = activePage === PageComponent}
  <div
    class="page-deck-slot"
    style:display={isCurrent ? 'contents' : 'none'}
    aria-hidden={!isCurrent}
    inert={!isCurrent}
  >
    <PageComponent expanded={isCurrent ? expanded : false} />
  </div>
{/each}

{#if activePage && isTransientPage(activePage)}
  {@const TransientPage = activePage}
  <TransientPage {expanded} />
{/if}
