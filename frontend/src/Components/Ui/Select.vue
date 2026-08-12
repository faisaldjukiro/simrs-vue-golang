<script setup>
import { computed, useAttrs } from 'vue'
import PrimeSelect from 'primevue/select'

const model = defineModel({ default: '' })
const attrs = useAttrs()

defineOptions({ inheritAttrs: false })

defineProps({
  options: { type: Array, default: () => [] },
  optionLabel: { type: String, default: 'label' },
  optionValue: { type: String, default: 'value' },
  placeholder: { type: String, default: 'Pilih Data' },
  emptyMessage: { type: String, default: 'Data tidak ditemukan' },
  filterPlaceholder: { type: String, default: 'Cari data' },
  filter: Boolean,
  disabled: Boolean,
})

const atributTurunan = computed(() => {
  const { class: _class, ...atribut } = attrs
  return atribut
})
</script>

<template>
  <PrimeSelect
    v-model="model"
    :class="['ui-select', attrs.class]"
    :options="options"
    :option-label="optionLabel"
    :option-value="optionValue"
    :placeholder="placeholder"
    :empty-message="emptyMessage"
    :filter="filter"
    :filter-placeholder="filterPlaceholder"
    :disabled="disabled"
    append-to="self"
    v-bind="atributTurunan"
  />
</template>
