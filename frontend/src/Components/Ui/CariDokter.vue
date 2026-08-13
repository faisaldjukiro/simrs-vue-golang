<script setup>
import InputPencarian from './InputPencarian.vue'
import { cariDokterPenanganan, cariDokterRadiologi } from '../../lib/faisal/api'

const model = defineModel({ default: () => ({}) })
const props = defineProps({
  token: { type: String, required: true },
  sumber: { type: String, default: 'penanganan' },
  disabled: Boolean,
  required: Boolean,
})

const cariDokter = (kataKunci) => props.sumber === 'radiologi'
  ? cariDokterRadiologi(props.token, kataKunci)
  : cariDokterPenanganan(props.token, kataKunci)
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
