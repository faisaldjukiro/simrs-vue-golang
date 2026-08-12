<script setup>
import { computed } from 'vue'
import { LoaderCircle } from '@lucide/vue'
import Column from 'primevue/column'
import PatientFilters from '../Common/PatientFilters.vue'
import DataTable from '../Ui/DataTable.vue'

const props = defineProps({
  currentTab: { type: String, required: true },
  patientRows: { type: Array, required: true },
  filteredPatientRows: { type: Array, required: true },
  filter: { type: Object, default: () => ({}) },
  daftarPoliklinik: { type: Array, default: () => [] },
  daftarDokter: { type: Array, default: () => [] },
  daftarStatusPeriksa: { type: Array, default: () => [] },
  daftarStatusRawatInap: { type: Array, default: () => [] },
  daftarStatusBayar: { type: Array, default: () => [] },
  pagination: { type: Object, default: () => ({ halaman: 1, batas: 10, total: 0 }) },
  isDark: Boolean,
  dashboardError: { type: String, default: '' },
  dashboardLoading: Boolean,
})

const emit = defineEmits(['apply-filters', 'page-change', 'select-patient'])
const bolehFilter = computed(() => ['Rawat Jalan', 'IGD/UGD', 'Rawat Inap'].includes(props.currentTab))
const rawatInap = computed(() => props.currentTab === 'Rawat Inap')
const jumlahDataTampil = computed(() => props.dashboardLoading ? 0 : props.filteredPatientRows.length)
const jumlahDataTotal = computed(() => props.dashboardLoading ? 0 : props.pagination.total ?? props.patientRows.length)
const firstRow = computed(() => ((props.pagination.halaman || 1) - 1) * (props.pagination.batas || 10))
const rowsPerPage = computed(() => props.pagination.batas || 10)

function gantiHalaman(event) {
  emit('page-change', {
    halaman: Math.floor(event.first / event.rows) + 1,
    batas: event.rows,
  })
}

function labelJenisKelamin(kode) {
  const nilai = String(kode || '').trim().toUpperCase()
  if (nilai === 'P') return 'Perempuan'
  if (nilai === 'L') return 'Laki-laki'
  return nilai || ''
}
</script>

<template>
  <section class="patient-module">
    <header class="patient-module-header">
      <div>
        <span>Data Pasien</span>
        <h1>{{ currentTab }}</h1>
        <p>{{ jumlahDataTampil }} dari {{ jumlahDataTotal }} data ditampilkan</p>
      </div>
    </header>

    <PatientFilters
      v-if="bolehFilter"
      :filter="filter"
      :daftar-poliklinik="daftarPoliklinik"
      :daftar-dokter="daftarDokter"
      :daftar-status-periksa="daftarStatusPeriksa"
      :daftar-status-rawat-inap="daftarStatusRawatInap"
      :daftar-status-bayar="daftarStatusBayar"
      :sedang-memuat="dashboardLoading"
      :mode-gelap="isDark"
      :tipe="currentTab"
      @terapkan="emit('apply-filters', $event)"
    />

    <div v-if="dashboardError" class="patient-error">{{ dashboardError }}</div>
    <div v-else-if="dashboardLoading" class="patient-loading-panel">
      <LoaderCircle class="spin" :size="28" />
      <strong>Sedang menarik data {{ currentTab }}...</strong>
      <span>Tabel akan tampil setelah data berhasil dibaca.</span>
    </div>
    <DataTable
      v-else
      class="patient-list-table"
      :rows="filteredPatientRows"
      data-key="no_rawat"
      empty-message="Data pasien tidak ditemukan."
      paginator
      lazy
      :first="firstRow"
      :rows-per-page="rowsPerPage"
      :rows-per-page-options="[10, 25, 50, 100]"
      :total-records="jumlahDataTotal"
      clickable
      @page="gantiHalaman"
      @row-click="emit('select-patient', $event.data)"
    >
      <template v-if="rawatInap">
        <Column header="No.">
          <template #body="{ index }">
            <strong class="table-number">{{ firstRow + index + 1 }}</strong>
          </template>
        </Column>

        <Column header="No. Rawat">
          <template #body="{ data: patient }">
            <strong><code>{{ patient.no_rawat }}</code></strong>
          </template>
        </Column>

        <Column header="No. RM">
          <template #body="{ data: patient }">
            <strong><code>{{ patient.no_rekam_medis }}</code></strong>
          </template>
        </Column>

        <Column header="Nama Pasien">
          <template #body="{ data: patient }">
            <strong>{{ patient.nama_pasien || `RM ${patient.no_rekam_medis}` }}</strong>
            <span>
              {{ patient.umur || '-' }}
              <template v-if="labelJenisKelamin(patient.jenis_kelamin)"> - {{ labelJenisKelamin(patient.jenis_kelamin) }}</template>
            </span>
          </template>
        </Column>

        <Column header="Kamar/Ruangan">
          <template #body="{ data: patient }">
            <strong>{{ patient.kamar || '-' }}</strong>
            <span v-if="patient.diagnosa_awal">{{ patient.diagnosa_awal }}</span>
          </template>
        </Column>

        <Column header="DPJP">
          <template #body="{ data: patient }"><strong>{{ patient.dokter || '-' }}</strong></template>
        </Column>

        <Column header="Status Pulang">
          <template #body="{ data: patient }">
            <span class="patient-status">{{ patient.status || '-' }}</span>
          </template>
        </Column>

        <Column header="Tgl Masuk">
          <template #body="{ data: patient }">
            <strong><code>{{ patient.tanggal_registrasi || '-' }}</code></strong>
          </template>
        </Column>

        <Column header="Jam">
          <template #body="{ data: patient }">
            <strong><code>{{ patient.jam_registrasi || '-' }}</code></strong>
          </template>
        </Column>

        <Column header="Lama Rawat">
          <template #body="{ data: patient }"><strong>{{ patient.lama_rawat || '-' }}</strong></template>
        </Column>

        <Column header="Cara Bayar">
          <template #body="{ data: patient }"><strong>{{ patient.penjamin || '-' }}</strong></template>
        </Column>

        <Column header="Status Bayar">
          <template #body="{ data: patient }"><strong>{{ patient.status_bayar || '-' }}</strong></template>
        </Column>

        <Column header="No. SEP">
          <template #body="{ data: patient }">
            <strong><code>{{ patient.no_sep || '-' }}</code></strong>
          </template>
        </Column>
      </template>

      <template v-else>
        <Column header="No.">
          <template #body="{ index }">
            <strong class="table-number">{{ firstRow + index + 1 }}</strong>
          </template>
        </Column>

        <Column header="No. Rawat">
          <template #body="{ data: patient }">
            <strong><code>{{ patient.no_rawat }}</code></strong>
          </template>
        </Column>

        <Column header="No. CM">
          <template #body="{ data: patient }">
            <strong><code>{{ patient.no_rekam_medis }}</code></strong>
          </template>
        </Column>

        <Column header="Nama Pasien">
          <template #body="{ data: patient }">
            <strong>{{ patient.nama_pasien || `RM ${patient.no_rekam_medis}` }}</strong>
            <span>{{ patient.umur || '-' }}</span>
          </template>
        </Column>

        <Column header="Nama Dokter">
          <template #body="{ data: patient }"><strong>{{ patient.dokter || '-' }}</strong></template>
        </Column>

        <Column header="Poliklinik">
          <template #body="{ data: patient }"><strong>{{ patient.poliklinik || '-' }}</strong></template>
        </Column>

        <Column header="Status">
          <template #body="{ data: patient }">
            <span class="patient-status">{{ patient.status || '-' }}</span>
          </template>
        </Column>

        <Column header="Tanggal">
          <template #body="{ data: patient }">
            <strong><code>{{ patient.tanggal_registrasi || '-' }}</code></strong>
          </template>
        </Column>

        <Column header="Jam">
          <template #body="{ data: patient }">
            <strong><code>{{ patient.jam_registrasi || '-' }}</code></strong>
          </template>
        </Column>

        <Column header="No. Reg.">
          <template #body="{ data: patient }">
            <strong><code>{{ patient.no_registrasi || '-' }}</code></strong>
          </template>
        </Column>

        <Column header="Cara Bayar">
          <template #body="{ data: patient }"><strong>{{ patient.penjamin || '-' }}</strong></template>
        </Column>

        <Column header="Status Bayar">
          <template #body="{ data: patient }"><strong>{{ patient.status_bayar || '-' }}</strong></template>
        </Column>

        <Column header="No. SEP">
          <template #body="{ data: patient }">
            <strong><code>{{ patient.no_sep || '-' }}</code></strong>
          </template>
        </Column>
      </template>
    </DataTable>
  </section>
</template>
