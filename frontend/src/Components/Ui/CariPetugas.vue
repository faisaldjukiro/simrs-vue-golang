<script setup>
import { computed } from 'vue'
import InputPencarian from './InputPencarian.vue'
import { cariPetugasCppt, cariPetugasEwsRanap, cariPetugasPenanganan } from '../../lib/faisal/api'

const model = defineModel({ default: () => ({}) })
const props = defineProps({
  token: { type: String, required: true },
  sumber: { type: String, default: 'cppt' },
  label: { type: String, default: 'Petugas' },
  placeholder: { type: String, default: 'Cari NIP, nama, atau jabatan petugas...' },
  disabled: Boolean,
  required: Boolean,
  error: { type: String, default: '' },
})

const kodeField = computed(() => props.sumber === 'penanganan' ? 'kode' : 'nip')
const cariPetugas = (kataKunci) => {
  if (props.sumber === 'penanganan') return cariPetugasPenanganan(props.token, kataKunci)
  if (props.sumber === 'ews_ranap') return cariPetugasEwsRanap(props.token, kataKunci)
  return cariPetugasCppt(props.token, kataKunci)
}
</script>

<template>
  <div class="staff-search-wrapper">
    <InputPencarian
      v-model="model"
      :label="label"
      :placeholder="placeholder"
      :search="cariPetugas"
      :code-field="kodeField"
      description-field="jabatan"
      :disabled="disabled"
      :required="required"
      :error="error"
    />
    <small v-if="disabled" class="staff-search-hint">Petugas mengikuti akun yang sedang login.</small>
  </div>
</template>
