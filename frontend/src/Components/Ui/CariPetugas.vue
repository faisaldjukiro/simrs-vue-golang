<script setup>
import { computed } from 'vue'
import InputPencarian from './InputPencarian.vue'
import { cariPetugasCppt, cariPetugasPenanganan } from '../../lib/faisal/api'

const model = defineModel({ default: () => ({}) })
const props = defineProps({
  token: { type: String, required: true },
  sumber: { type: String, default: 'cppt' },
  label: { type: String, default: 'Petugas' },
  placeholder: { type: String, default: 'Cari NIP, nama, atau jabatan petugas...' },
  disabled: Boolean,
  required: Boolean,
})

const kodeField = computed(() => props.sumber === 'penanganan' ? 'kode' : 'nip')
const cariPetugas = (kataKunci) => props.sumber === 'penanganan'
  ? cariPetugasPenanganan(props.token, kataKunci)
  : cariPetugasCppt(props.token, kataKunci)
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
    />
    <small v-if="disabled" class="staff-search-hint">Petugas mengikuti akun yang sedang login.</small>
  </div>
</template>
