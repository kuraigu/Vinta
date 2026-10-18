<script lang="ts">
  import CustomSelect from '../components/CustomSelect.svelte'

  interface Props {
    options: { value: string; label: string }[]
    value: string
    placeholder?: string
  }

  let {
    options,
    value = $bindable(),
    placeholder = 'Select...'
  }: Props = $props()
</script>

<CustomSelect {options} bind:value {placeholder} />
