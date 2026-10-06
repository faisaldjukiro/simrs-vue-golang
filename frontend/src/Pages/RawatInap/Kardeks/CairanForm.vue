<script setup lang="ts">
import FormInput from '../../../Components/Ui/FormInput.vue'
import type { PropsKardeks } from '../../../types/kardeks'
import { useCairan } from './useCairan'

const props = defineProps<PropsKardeks>()
const emit = defineEmits<{ berubah: [] }>()
const { form, saving, error, kategori, simpan } = useCairan(props, () => emit('berubah'))
</script>

<template>
  <article class="clinical-form-card">
    <header class="clinical-section-header">
      <div>
        <h3>Catat Cairan Masuk / Keluar</h3>
        <p>Volume aktual selama periode pengukuran, bukan jumlah kumulatif atau kecepatan infus.</p>
        <p>Petugas mengikuti akun login yang terhubung dengan NIP petugas SIMRS.</p>
      </div>
    </header>
    <form class="kardeks-filter" @submit.prevent="simpan">
      <FormInput v-model="form.waktu_mulai" label="Mulai Pengukuran (WITA)" type="datetime-local" required :disabled="saving" />
      <FormInput v-model="form.waktu_selesai" label="Selesai Pengukuran (WITA)" type="datetime-local" required :disabled="saving" />
      <FormInput v-model="form.jenis" label="Jenis" jenis="select" :options="[{ label: 'Masuk', value: 'Masuk' }, { label: 'Keluar', value: 'Keluar' }]" :disabled="saving" />
      <FormInput v-model="form.kategori" label="Kategori" jenis="select" :options="kategori" :disabled="saving" />
      <FormInput v-model="form.rincian" label="Nama Cairan / Lokasi Drain" :maxlength="100" :disabled="saving" />
      <FormInput v-model="form.volume_ml" label="Volume Aktual (mL)" type="number" min="0" max="9999999999.99" step="0.01" required :disabled="saving" />
      <FormInput v-model="form.catatan" label="Catatan" jenis="textarea" :maxlength="2000" :disabled="saving" />
      <button class="clinical-button primary" type="submit" :disabled="saving">
        {{ saving ? 'Menyimpan...' : 'Simpan Cairan' }}
      </button>
    </form>
    <p v-if="error" class="patient-error" role="alert">{{ error }}</p>
  </article>
</template>

<style src="./kardeks.css" scoped></style>
