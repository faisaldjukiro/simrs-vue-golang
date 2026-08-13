<script setup>
import InputPencarian from './InputPencarian.vue'
import { cariDokterPenanganan } from '../../lib/faisal/api'

const model = defineModel({ default: () => ({}) })
const props = defineProps({
  token: { type: String, required: true },
  disabled: Boolean,
  required: Boolean,
})

const cariDokter = (kataKunci) => cariDokterPenanganan(props.token, kataKunci)
</script>

<template>
  <InputPencarian
    v-model="model"
    label="Dokter"
    placeholder="Cari kode, nama, atau spesialis dokter..."
    :search="cariDokter"
    description-field="spesialis"
    :disabled="disabled"
    :required="required"
  />
</template>
