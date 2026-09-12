<script setup lang="ts">
import { BarChart3, BedDouble, Clock3, Download, Filter, X } from "@lucide/vue";
import Column from "primevue/column";
import DataTable from "../../Components/Ui/DataTable.vue";
import FormInput from "../../Components/Ui/FormInput.vue";
import { useMonitoringBed } from "./useMonitoringBed";

const props = defineProps({ token: { type: String, required: true } });
const {
  loading,
  error,
  daftarBed,
  filter,
  muatData,
  ringkasan,
  waktuDimuat,
  grafikVisible,
  grafik,
  firstRow,
  gantiHalaman,
  excel,
} = useMonitoringBed(props.token);

const opsiStatus = [
  { value: '', label: 'Semua Status' },
  { value: 'KOSONG', label: 'Kosong' },
  { value: 'ISI', label: 'Terisi' },
  { value: 'DIBERSIHKAN', label: 'Dibersihkan' },
];

</script>

<template>
  <section class="visit-report-module monitoring-bed">
    <header class="visit-report-header">
      <div>
        <span>Modul BED</span>
        <h1>Monitoring Ketersediaan Bed</h1>
        <p>Status kamar dan informasi penggunaan bed saat ini.</p>
      </div>
      <BedDouble :size="30" />
    </header>

    <form class="visit-report-filter monitoring-bed-filter" @submit.prevent="muatData">
      <FormInput
        v-model="filter.q"
        label="Pencarian"
        placeholder="Nama bed, bangsal, atau pasien..."
      />
      <FormInput
        v-model="filter.status"
        jenis="select"
        label="Status Kamar"
        :options="opsiStatus"
      />
      <button :disabled="loading">
        <Filter :size="16" />
        {{ loading ? "Mencari..." : "Tampilkan" }}
      </button>
    </form>

    <div class="monitoring-summary" aria-label="Ringkasan bed pada hasil filter">
      <article>
        <span>Total Bed</span>
        <strong>{{ ringkasan.total }}</strong>
        <small>Pada hasil filter</small>
      </article>
      <article>
        <span><i class="status-dot isi"></i>Terisi</span>
        <strong>{{ ringkasan.isi }}</strong>
        <small>Sedang digunakan</small>
      </article>
      <article>
        <span><i class="status-dot kosong"></i>Kosong</span>
        <strong>{{ ringkasan.kosong }}</strong>
        <small>Tidak sedang digunakan</small>
      </article>
      <article>
        <span><i class="status-dot dibersihkan"></i>Dibersihkan</span>
        <strong>{{ ringkasan.dibersihkan }}</strong>
        <small>Dalam persiapan</small>
      </article>
      <article v-if="ringkasan.lainnya">
        <span>Status Lainnya</span>
        <strong>{{ ringkasan.lainnya }}</strong>
        <small>Termasuk status belum diketahui</small>
      </article>
    </div>

    <div class="visit-report-tools monitoring-tools">
      <div class="monitoring-table-heading">
        <h2>Daftar Ketersediaan Bed</h2>
        <span v-if="waktuDimuat">{{ daftarBed.length }} bed · Diperbarui {{ waktuDimuat }}</span>
        <span v-else>{{ loading ? 'Memuat data bed...' : 'Belum ada data yang dimuat.' }}</span>
      </div>
      <button
        class="chart-button"
        :class="{ active: grafikVisible }"
        :aria-expanded="grafikVisible"
        aria-controls="grafik-monitoring-bed"
        @click="grafikVisible = !grafikVisible"
      >
        <X v-if="grafikVisible" :size="16" />
        <BarChart3 v-else :size="16" />
        {{ grafikVisible ? 'Tutup Grafik' : 'Lihat Grafik' }}
      </button>
      <button :disabled="loading || !daftarBed.length" @click="excel">
        <Download :size="16" />
        Export Excel
      </button>
    </div>

    <section
      v-if="grafikVisible"
      id="grafik-monitoring-bed"
      class="visit-analytics"
    >
      <header>
        <div>
          <span>Visualisasi Data</span>
          <h2>Grafik Ketersediaan Bed</h2>
          <p>Grafik mengikuti seluruh hasil filter yang telah ditampilkan, termasuk halaman tabel lainnya.</p>
        </div>
        <BarChart3 :size="25" />
      </header>
      <div class="visit-chart-grid">
        <article
          v-for="bagian in grafik"
          :key="bagian.judul"
          class="visit-chart-card rank-chart"
          :class="{ 'monitoring-chart-wide': bagian.lebar }"
        >
          <h3>{{ bagian.judul }}</h3>
          <p>{{ bagian.keterangan }}</p>
          <div v-if="bagian.data.length" class="horizontal-chart">
            <div v-for="item in bagian.data" :key="item.label">
              <label>
                <span :title="item.label">{{ item.label }}</span>
                <b>{{ item.nilai }} bed</b>
              </label>
              <i aria-hidden="true">
                <u :style="{ width: `${item.persen}%` }"></u>
              </i>
            </div>
          </div>
          <em v-else>Belum ada data untuk divisualisasikan.</em>
        </article>
      </div>
    </section>

    <div v-if="error" class="patient-error" role="alert">{{ error }}</div>
    <DataTable
      v-else
      class="visit-report-table monitoring-bed-table"
      :rows="daftarBed"
      data-key="kamar"
      :loading="loading"
      paginator
      :first="firstRow"
      :rows-per-page="25"
      :rows-per-page-options="[25, 50, 100]"
      empty-message="Data bed tidak ditemukan."
      @page="gantiHalaman"
    >
      <Column
        header="No."
        header-class="monitoring-center"
        body-class="monitoring-center"
        style="width: 70px; min-width: 70px"
      >
        <template #body="{ index }">
          <span class="bed-row-number">{{ firstRow + index + 1 }}</span>
        </template>
      </Column>
      <Column field="kamar" header="Kamar / Kelas" style="width: 155px">
        <template #body="{ data: r }">
          <div class="bed-cell-stack">
            <span class="bed-cell-primary bed-code">{{ r.kamar }}</span>
            <span class="bed-cell-secondary">{{ r.kelas || '-' }}</span>
          </div>
        </template>
      </Column>
      <Column field="bangsal" header="Bangsal / Ruangan" style="width: 19%" />
      <Column
        field="status_kamar"
        header="Status"
        header-class="monitoring-center"
        body-class="monitoring-center"
        style="width: 130px"
      >
        <template #body="{ data: r }">
          <span class="bed-status-badge" :class="r.status_kamar?.toLowerCase()">
            <i aria-hidden="true"></i>
            {{ r.statusLabel }}
          </span>
        </template>
      </Column>
      <Column header="Pasien / No. Rawat">
        <template #body="{ data: r }">
          <div v-if="r.status_kamar === 'ISI'" class="bed-cell-stack">
            <span class="bed-cell-primary">
              {{ r.nama_pasien || 'Data pasien belum tersedia' }}
            </span>
            <span class="bed-cell-secondary">
              No. rawat
              <span class="bed-record-number">{{ r.no_rawat || 'belum tersedia' }}</span>
            </span>
          </div>
          <div v-else-if="r.status_kamar === 'KOSONG'" class="bed-cell-stack">
            <span class="bed-cell-secondary">Tidak ada pasien</span>
          </div>
          <div v-else class="bed-cell-stack">
            <span class="bed-cell-primary">
              {{ r.status_kamar === 'DIBERSIHKAN' ? 'Dalam proses pembersihan' : 'Informasi belum tersedia' }}
            </span>
          </div>
        </template>
      </Column>
      <Column header="Tanggal Penggunaan" style="width: 160px">
        <template #body="{ data: r }">
          <div v-if="r.status_kamar === 'ISI'" class="bed-cell-stack">
            <span class="bed-cell-date">{{ r.tanggalMasukLabel }}</span>
            <span class="bed-cell-secondary">Tanggal masuk</span>
          </div>
          <div v-else-if="r.status_kamar === 'KOSONG'" class="bed-cell-stack">
            <span class="bed-cell-date">
              {{ r.terakhirDipakaiLabel || 'Belum ada riwayat' }}
            </span>
            <span v-if="r.terakhirDipakaiLabel" class="bed-cell-secondary">Terakhir digunakan</span>
          </div>
          <span v-else class="bed-cell-secondary">—</span>
        </template>
      </Column>
      <Column header="Durasi" style="width: 175px">
        <template #body="{ data: r }">
          <span v-if="r.status_kamar === 'ISI' && r.lamaPenggunaanLabel" class="bed-duration">
            <Clock3 :size="14" aria-hidden="true" />
            {{ r.lamaPenggunaanLabel }}
          </span>
          <span v-else-if="r.status_kamar === 'KOSONG' && r.lamaKosongLabel" class="bed-duration">
            <Clock3 :size="14" aria-hidden="true" />
            {{ r.lamaKosongLabel }}
          </span>
          <span v-else class="bed-cell-secondary">-</span>
        </template>
      </Column>
    </DataTable>
  </section>
</template>

<style src="../../Components/Ui/report.css" scoped></style>
<style src="./monitoring-bed.css" scoped></style>
