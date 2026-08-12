<script setup>
import { computed, reactive, watch } from 'vue'
import { ListFilter, LoaderCircle, Search } from '@lucide/vue'
import DatePicker from '../Ui/DatePicker.vue'
import Select from '../Ui/Select.vue'

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
  <form class="patient-filter" @submit.prevent="terapkanFilter">
    <div class="filter-date-range">
      <label class="filter-date-field">
        <span>Tanggal mulai</span>
        <DatePicker
          v-model="tanggalMulai"
          placeholder="Tanggal mulai"
          :disabled="sedangMemuat"
        />
      </label>

      <label class="filter-date-field">
        <span>Sampai dengan</span>
        <DatePicker
          v-model="tanggalSelesai"
          placeholder="Sampai dengan"
          :disabled="sedangMemuat"
        />
      </label>
    </div>

    <Select
      v-model="filterLokal.status"
      class="filter-status"
      :options="opsiStatus"
      placeholder="Semua Status"
      empty-message="Status tidak ditemukan"
      :disabled="sedangMemuat"
    />

    <Select
      v-model="filterLokal.status_bayar"
      class="filter-status-bayar"
      :options="opsiStatusBayar"
      placeholder="Semua Status Bayar"
      empty-message="Status bayar tidak ditemukan"
      :disabled="sedangMemuat"
    />

    <Select
      v-model="filterLokal.dokter"
      class="filter-dokter"
      :options="opsiDokter"
      :placeholder="tipe === 'Rawat Inap' ? 'Semua Dokter / DPJP' : 'Semua Dokter Spesialis'"
      :empty-message="tipe === 'Rawat Inap' ? 'Dokter / DPJP tidak ditemukan' : 'Dokter tidak ditemukan'"
      filter
      :filter-placeholder="tipe === 'Rawat Inap' ? 'Cari dokter / DPJP' : 'Cari dokter'"
      :disabled="sedangMemuat"
    />

    <Select
      v-if="tampilkanPoliklinik"
      v-model="filterLokal.poly"
      class="filter-poliklinik"
      :options="opsiPoliklinik"
      placeholder="Semua Poliklinik"
      empty-message="Poliklinik tidak ditemukan"
      filter
      filter-placeholder="Cari poliklinik"
      :disabled="sedangMemuat"
    />

    <label class="filter-search">
      <Search :size="17" />
      <input
        v-model="filterLokal.search"
        type="search"
        :disabled="sedangMemuat"
        :placeholder="placeholderPencarian"
      />
    </label>

    <label v-if="tampilkanBelumPulang" class="filter-check">
      <input v-model="filterLokal.belum_pulang" type="checkbox" :disabled="sedangMemuat" />
      Belum Pulang
    </label>

    <button class="filter-submit" type="submit" :disabled="sedangMemuat">
      <LoaderCircle v-if="sedangMemuat" class="spin" :size="17" />
      <ListFilter v-else :size="17" />
      {{ sedangMemuat ? 'Memuat' : 'Filter' }}
    </button>
  </form>
</template>
