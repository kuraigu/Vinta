<!--
  @component SidebarLayout
  @description Responsive two-column split layout for pages like Settings and Onboarding,
  featuring a BentoCard sidebar with ScrollView overflow indicators and a flexible content pane.
-->
<script lang="ts" module>
  import type { ComponentType } from 'svelte'
  import type { IconBadgeColor } from './IconBadge.svelte'

  export interface SidebarNavItem {
    id: string
    label: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon: ComponentType<any>
    badge?: IconBadgeColor
  }
</script>

<script lang="ts">
  import type { Snippet } from 'svelte'
  import BackButton from './BackButton.svelte'
  import BentoCard from './BentoCard.svelte'
  import ScrollView from './ScrollView.svelte'
  import IconBadge from './IconBadge.svelte'
  import { tStore } from '../store'

  interface Props {
    title?: string
    showBackButton?: boolean
    onback?: () => void
    sidebarWidth?: string
    noDockClearance?: boolean
    dataVinta?: string
    items?: SidebarNavItem[]
    activeId?: string
    onSelect?: (id: string) => void
    sidebar?: Snippet
    sidebarFooter?: Snippet
    content?: Snippet
    children?: Snippet
  }

  let {
    title = '',
    showBackButton = false,
    onback,
    sidebarWidth = '168px',
    noDockClearance = false,
    dataVinta,
    items,
    activeId,
    onSelect,
    sidebar,
    sidebarFooter,
    content,
    children
  }: Props = $props()
</script>

<div
  data-vinta={dataVinta}
  class="sidebar-page-layout"
  class:no-dock-clearance={noDockClearance}
>
  {#if title}
    <div class="page-header" class:has-back={showBackButton}>
      {#if showBackButton}
        <BackButton {onback} />
      {/if}
      <div class="brand-page-title-group">
        <span class="page-title">{title}</span>
      </div>
      {#if showBackButton}
        <div class="header-spacer"></div>
      {/if}
    </div>
  {/if}

  <div class="sidebar-body">
    <BentoCard
      variant="sidebar"
      class="sidebar-pane"
      style="width: {sidebarWidth}"
    >
      <ScrollView
        showScrollIndicators={true}
        hideScrollbar={true}
        scrollIndicatorLabel={$tStore('home.scroll_cue_more')}
      >
        {#if items && items.length > 0}
          <div class="sidebar-nav-list" role="tablist">
            {#each items as item (item.id)}
              {@const isActive = activeId === item.id}
              <button
                type="button"
                role="tab"
                aria-selected={isActive}
                class="sidebar-nav-btn"
                class:active={isActive}
                onclick={() => onSelect?.(item.id)}
              >
                <IconBadge
                  icon={item.icon}
                  color={item.badge ?? 'blue'}
                  size="sm"
                />
                <span class="sidebar-nav-label">{item.label}</span>
              </button>
            {/each}
          </div>
        {:else}
          {@render sidebar?.()}
        {/if}
      </ScrollView>

      {#if sidebarFooter}
        <div class="sidebar-footer-slot">
          {@render sidebarFooter()}
        </div>
      {/if}
    </BentoCard>

    <div class="content-pane">
      {@render (content || children)?.()}
    </div>
  </div>
</div>

<style src="./styles/SidebarLayout.css"></style>
