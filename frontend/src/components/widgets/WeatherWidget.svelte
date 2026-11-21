<!--
  @component WeatherWidget
  @description Anchored Home weather hero widget with `data-vinta-widget="weather"` scoping.
-->
<script lang="ts">
  import WeatherGlance from '../WeatherGlance.svelte'

  interface Props {
    salutation?: string
    userName?: string
    timeText?: string
    amPmText?: string
    paused?: boolean
  }

  let {
    salutation = '',
    userName = '',
    timeText = '',
    amPmText = '',
    paused = false
  }: Props = $props()
</script>

<div data-vinta-widget="weather">
  <WeatherGlance {salutation} {userName} {timeText} {amPmText} {paused} />
</div>
