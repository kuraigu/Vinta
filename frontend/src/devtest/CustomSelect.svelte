<script lang="ts">
  import CustomSelect from '../components/CustomSelect.svelte'

  let {
    options,
    value = $bindable(),
    placeholder = 'Select...'
  } = $props<{
    options: { value: string; label: string }[]
    value: string
    placeholder?: string
  }>()
</script>

<CustomSelect {options} bind:value {placeholder} />
