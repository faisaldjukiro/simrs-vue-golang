<script setup lang="ts">
import PrimeDatePicker from 'primevue/datepicker'

const model = defineModel({ default: null })

defineProps({
  placeholder: { type: String, default: 'Pilih tanggal' },
  disabled: Boolean,
  showIcon: { type: Boolean, default: true },
})
</script>

<template>
  <PrimeDatePicker
    v-model="model"
    class="ui-date-picker"
    :show-icon="showIcon"
    show-button-bar
    date-format="dd/mm/yy"
    :placeholder="placeholder"
    :disabled="disabled"
  />
</template>
