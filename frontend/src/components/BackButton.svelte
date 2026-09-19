<script lang="ts">
  import { getContext } from 'svelte'
  import { ArrowLeft } from 'lucide-svelte'
  import AppsPage, { config as appsConfig } from '../pages/AppsPage.svelte'
  import type { IslandConfig } from '../store'

  interface Props {
    onback?: () => void
    title?: string
  }

  let { onback, title = 'Back to Apps' }: Props = $props()

  const navigate =
    getContext<
      (page: unknown, pageConfig: IslandConfig, expand?: boolean) => void
    >('navigate')

  function handleClick() {
    if (onback) {
      onback()
    } else if (typeof navigate === 'function') {
      navigate(AppsPage, appsConfig, true)
    }
  }
</script>

<button class="back-btn" onclick={handleClick} {title}>
  <ArrowLeft size={14} />
</button>

<style src="./styles/BackButton.css"></style>
