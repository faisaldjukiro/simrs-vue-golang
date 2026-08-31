<script setup lang="ts">
import { computed, useAttrs } from 'vue'
import PrimeSelect from 'primevue/select'

const model = defineModel<any>({ default: '' })
const attrs = useAttrs()

defineOptions({ inheritAttrs: false })

withDefaults(defineProps<{
  options?: unknown[]
  optionLabel?: string
  optionValue?: string
  placeholder?: string
  emptyMessage?: string
  filterPlaceholder?: string
  appendTo?: 'self' | 'body' | HTMLElement
  overlayClass?: string | Record<string, boolean>
  filter?: boolean
  disabled?: boolean
}>(), {
  options: () => [],
  optionLabel: 'label',
  optionValue: 'value',
  placeholder: 'Pilih Data',
  emptyMessage: 'Data tidak ditemukan',
  filterPlaceholder: 'Cari data',
  appendTo: 'self',
  overlayClass: undefined,
  filter: false,
  disabled: false,
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
