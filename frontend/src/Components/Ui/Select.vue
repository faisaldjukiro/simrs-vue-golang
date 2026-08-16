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
  appendTo: { type: [String, Object], default: 'self' },
  overlayClass: { type: [String, Object], default: undefined },
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
    :append-to="appendTo"
    :overlay-class="overlayClass"
    :disabled="disabled"
    v-bind="atributTurunan"
  />
</template>
