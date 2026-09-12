<script setup lang="ts">
// @ts-nocheck -- komponen lama, tipe props akan dimigrasikan bertahap.
import { computed, reactive, watch } from 'vue'
import { ListFilter, LoaderCircle } from '@lucide/vue'
import FormInput from '../Ui/FormInput.vue'

const props = defineProps({
  filter: { type: Object, required: true },
  daftarPoliklinik: { type: Array, default: () => [] },
  daftarDokter: { type: Array, default: () => [] },
  daftarStatusPeriksa: { type: Array, default: () => [] },
  daftarStatusRawatInap: { type: Array, default: () => [] },
  daftarStatusBayar: { type: Array, default: () => [] },
  sedangMemuat: Boolean,
  modeGelap: Boolean,
  tipe: { type: String, default: '' },
})

const emit = defineEmits(['terapkan'])
const filterLokal = reactive({ ...props.filter })
const daftarStatus = computed(() => props.tipe === 'Rawat Inap' ? props.daftarStatusRawatInap : props.daftarStatusPeriksa)
const tampilkanPoliklinik = computed(() => props.tipe === 'Rawat Jalan')
const tampilkanBelumPulang = computed(() => props.tipe === 'Rawat Inap')
const opsiStatus = computed(() => [
  { label: 'Semua Status', value: '' },
  ...daftarStatus.value.map((status) => ({ label: status, value: status })),
])
const opsiStatusBayar = computed(() => [
  { label: 'Semua Status Bayar', value: '' },
  ...props.daftarStatusBayar.map((status) => ({ label: status, value: status })),
])
const opsiDokter = computed(() => [
  { label: props.tipe === 'Rawat Inap' ? 'Semua Dokter / DPJP' : 'Semua Dokter Spesialis', value: '' },
  ...props.daftarDokter.map((dokter) => {
    const kode = dokter.kode || dokter.kd_dokter || ''
    const nama = dokter.nama || dokter.nm_dokter || kode
    const spesialis = dokter.kode_spesialis || dokter.kd_sps || ''

    return {
      label: spesialis ? `${nama} (${spesialis})` : nama,
      value: kode,
    }
  }),
])
const opsiPoliklinik = computed(() => [
  { label: 'Semua Poliklinik', value: '' },
  ...props.daftarPoliklinik.map((poliklinik) => ({
    label: poliklinik.nama || poliklinik.name,
    value: poliklinik.kode || poliklinik.code,
  })),
])
const placeholderPencarian = computed(() => props.tipe === 'Rawat Inap'
  ? 'Cari pasien, No. RM, No. Rawat, kamar, DPJP'
  : 'Cari pasien, No. RM, No. Rawat')
const tanggalMulai = computed({
  get: () => parseTanggal(filterLokal.date_from),
  set: (tanggal) => {
    filterLokal.date_from = formatTanggal(tanggal)
    if (!filterLokal.date_to || filterLokal.date_to < filterLokal.date_from) {
      filterLokal.date_to = filterLokal.date_from
    }
    tanggalDipilih()
  },
})
const tanggalSelesai = computed({
  get: () => parseTanggal(filterLokal.date_to || filterLokal.date_from),
  set: (tanggal) => {
    filterLokal.date_to = formatTanggal(tanggal)
    if (!filterLokal.date_from) {
      filterLokal.date_from = filterLokal.date_to
    }
    if (filterLokal.date_to < filterLokal.date_from) {
      filterLokal.date_from = filterLokal.date_to
    }
    tanggalDipilih()
  },
})

// Gunakan normalisasi rentang tanggal yang sama dengan filter modul lainnya.
const tanggalMulaiInput = computed({
  get: () => filterLokal.date_from || '',
  set: (nilai) => {
    tanggalMulai.value = parseTanggal(nilai)
  },
})
const tanggalSelesaiInput = computed({
  get: () => filterLokal.date_to || filterLokal.date_from || '',
  set: (nilai) => {
    tanggalSelesai.value = parseTanggal(nilai)
  },
})

watch(
  () => props.filter,
  (filterBaru) => Object.assign(filterLokal, filterBaru),
  { deep: true },
)

function tanggalDipilih() {
  if (tampilkanBelumPulang.value) filterLokal.belum_pulang = false
}

function parseTanggal(nilai) {
  if (!nilai) return null
  const tanggal = new Date(`${nilai}T00:00:00`)
  return Number.isNaN(tanggal.getTime()) ? null : tanggal
}

function formatTanggal(nilai) {
  if (!nilai) return ''
  const tanggal = nilai instanceof Date ? nilai : new Date(nilai)
  if (Number.isNaN(tanggal.getTime())) return ''
  const tahun = tanggal.getFullYear()
  const bulan = String(tanggal.getMonth() + 1).padStart(2, '0')
  const hari = String(tanggal.getDate()).padStart(2, '0')
  return `${tahun}-${bulan}-${hari}`
}

function terapkanFilter() {
  if (props.sedangMemuat) return
  if (!filterLokal.date_to || filterLokal.date_to < filterLokal.date_from) {
    filterLokal.date_to = filterLokal.date_from
  }
  emit('terapkan', { ...filterLokal })
}

</script>

<template>
  <form class="service-filters" @submit.prevent="terapkanFilter">
    <FormInput
      v-model="tanggalMulaiInput"
      label="Tanggal mulai"
      type="date"
      :disabled="sedangMemuat"
    />
    <FormInput
      v-model="tanggalSelesaiInput"
      label="Sampai dengan"
      type="date"
      :disabled="sedangMemuat"
    />
    <FormInput
      v-model="filterLokal.dokter"
      class="service-filter-reference"
      :class="{ 'service-filter-wide': !tampilkanPoliklinik && !tampilkanBelumPulang }"
      :label="tampilkanBelumPulang ? 'Dokter / DPJP' : 'Dokter'"
      :placeholder="tampilkanBelumPulang ? 'Semua Dokter / DPJP' : 'Semua Dokter Spesialis'"
      jenis="select"
      :options="opsiDokter"
      filter
      :filter-placeholder="tampilkanBelumPulang ? 'Cari dokter / DPJP' : 'Cari dokter'"
      :disabled="sedangMemuat"
    />
    <FormInput
      v-if="tampilkanPoliklinik"
      v-model="filterLokal.poly"
      class="service-filter-reference"
      label="Poliklinik"
      placeholder="Semua Poliklinik"
      jenis="select"
      :options="opsiPoliklinik"
      filter
      filter-placeholder="Cari poliklinik"
      :disabled="sedangMemuat"
    />
    <label v-if="tampilkanBelumPulang" class="service-filter-check">
      <span>Cakupan data</span>
      <span class="service-filter-check-control">
        <input
          v-model="filterLokal.belum_pulang"
          type="checkbox"
          :disabled="sedangMemuat"
        />
        Belum Pulang
      </span>
    </label>
    <FormInput
      v-model="filterLokal.status"
      :label="tampilkanBelumPulang ? 'Status pulang' : 'Status periksa'"
      placeholder="Semua Status"
      jenis="select"
      :options="opsiStatus"
      :disabled="sedangMemuat"
    />
    <FormInput
      v-model="filterLokal.status_bayar"
      label="Status bayar"
      placeholder="Semua Status Bayar"
      jenis="select"
      :options="opsiStatusBayar"
      :disabled="sedangMemuat"
    />
    <FormInput
      v-model="filterLokal.search"
      class="service-filter-search"
      label="Pencarian pasien"
      type="search"
      :placeholder="placeholderPencarian"
      :disabled="sedangMemuat"
    />
    <button type="submit" class="service-filter-submit" :disabled="sedangMemuat">
      <LoaderCircle v-if="sedangMemuat" class="spin" :size="17" />
      <ListFilter v-else :size="17" />
      {{ sedangMemuat ? 'Memuat...' : 'Tampilkan' }}
    </button>
  </form>
</template>

<style scoped>
.service-filters {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  align-items: end;
  gap: 12px;
  margin-top: 16px;
  padding: 16px;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: var(--surface-soft);
}

.service-filter-reference,
.service-filter-check {
  grid-column: span 2;
}

.service-filter-wide {
  grid-column: span 4;
}

.service-filter-check {
  display: grid;
  gap: 6px;
  margin: 0;
  color: var(--muted);
  font-size: 12px;
}

.service-filter-check-control {
  display: flex;
  min-height: 42px;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  border: 1px solid var(--line);
  border-radius: 10px;
  background: var(--surface);
  color: var(--text);
  font-size: 13px;
}

.service-filter-check input {
  width: 16px;
  height: 16px;
  margin: 0;
  accent-color: #0d9488;
}

.service-filter-check input:focus-visible {
  outline: 2px solid #0d9488;
  outline-offset: 3px;
}

.service-filter-check input:disabled {
  cursor: not-allowed;
}

/* Ukuran teks filter; warna dan interaksi tetap dari komponen bersama. */
.service-filters :deep(.form-input-label) {
  font-size: 12px;
  letter-spacing: normal;
  text-transform: none;
}

.service-filters :deep(.form-input-element),
.service-filters :deep(.p-select-label) {
  font-size: 13px;
}

.service-filters :deep(.p-select-option) {
  font-size: 12px;
}

/* Saat loading, pertahankan palet filter; bawaan disabled PrimeVue lebih terang. */
.service-filters :deep(.ui-select.p-disabled) {
  border-color: var(--line);
  background: var(--surface);
  opacity: 1;
  cursor: not-allowed;
}

.service-filters :deep(.ui-select.p-disabled .p-select-label),
.service-filters :deep(.ui-select.p-disabled .p-select-dropdown) {
  color: var(--muted);
}

.service-filter-search {
  grid-column: span 3;
}

.service-filter-submit {
  display: flex;
  min-height: 42px;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 14px;
  border: 1px solid #0f766e;
  border-radius: 10px;
  color: #fff;
  background: #0f766e;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.service-filter-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.service-filter-submit:focus-visible {
  outline: 2px solid #0d9488;
  outline-offset: 3px;
}

@media (max-width: 1100px) {
  .service-filters {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .service-filter-reference,
  .service-filter-wide,
  .service-filter-check,
  .service-filter-search {
    grid-column: auto;
  }
}

@media (max-width: 600px) {
  .service-filters {
    grid-template-columns: minmax(0, 1fr);
    padding: 12px;
  }
}
</style>
