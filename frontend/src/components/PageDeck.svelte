<!--
  @component PageDeck
  @description Keep-alive page pool that preserves mounted page DOM across navigation
  switches so changing pages does not destroy or remount Svelte component trees.
-->
<script lang="ts">
  import { untrack, type Component } from 'svelte'
  import Notification from './Notification.svelte'
  import OnboardingPage from '../pages/OnboardingPage.svelte'

  interface Props {
    activePage?: Component
    expanded?: boolean
  }

  let { activePage, expanded = false }: Props = $props()

  let mountedPages = $state<Component[]>([])

  function isTransientPage(page: Component | undefined): boolean {
    return !page || page === Notification || page === OnboardingPage
  }

  function syncMountedPagePoolEffect() {
    const current = activePage
    if (!current || isTransientPage(current)) return
    untrack(() => {
      if (!mountedPages.includes(current)) {
        mountedPages = [...mountedPages, current]
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
  >
    <PageComponent expanded={isCurrent ? expanded : false} />
  </div>
{/each}

{#if activePage && isTransientPage(activePage)}
  {@const TransientPage = activePage}
  <TransientPage {expanded} />
{/if}
