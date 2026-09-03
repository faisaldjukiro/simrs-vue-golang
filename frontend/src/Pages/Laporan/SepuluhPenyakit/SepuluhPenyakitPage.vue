<script setup lang="ts">
import { BarChart3, Download, Filter, Printer, RotateCcw, X } from "@lucide/vue";
import Column from "primevue/column";
import DataTable from "../../../Components/Ui/DataTable.vue";
import FormInput from "../../../Components/Ui/FormInput.vue";
import { useSepuluhPenyakit } from "./useSepuluhPenyakit";

const props = defineProps({
  token: { type: String, required: true },
});

const {
  loading,
  error,
  data,
  filter,
  ringkasan,
  grafik,
  grafikVisible,
  muat,
  reset,
  excel,
  cetak,
} = useSepuluhPenyakit(props);
</script>

<template>
  <section class="visit-report-module disease-report-module">
    <header class="visit-report-header">
      <div>
        <span>Laporan Pelayanan</span>
        <h1>Laporan 10 Penyakit</h1>
        <p>Adaptasi laporan Dlg10Penyakit SIMRS Khanza.</p>
      </div>
      <BarChart3 :size="30" />
    </header>

    <form class="visit-report-filter disease-report-filter" @submit.prevent="muat">
      <FormInput
        v-model="filter.tanggal_mulai"
        label="Tanggal Mulai"
        type="date"
        required
      />
      <FormInput
        v-model="filter.tanggal_selesai"
        label="Sampai Dengan"
        type="date"
        required
      />
      <FormInput
        v-model="filter.status"
        label="Status Pelayanan"
        jenis="select"
        :options="[
          { label: 'Semua', value: 'Semua' },
          { label: 'Rawat Jalan', value: 'Ralan' },
          { label: 'Rawat Inap', value: 'Ranap' },
        ]"
      />
      <FormInput
        v-model="filter.q"
        label="Kata Kunci"
        placeholder="Kode atau nama penyakit..."
      />
      <button type="submit" :disabled="loading">
        <Filter :size="16" />
        {{ loading ? "Memuat..." : "Tampilkan" }}
      </button>
      <button type="button" class="secondary-button" @click="reset">
        <RotateCcw :size="16" /> Reset
      </button>
    </form>

    <div class="visit-report-summary">
      <article>
        <span>Jumlah Penyakit</span>
        <strong>{{ ringkasan.jumlah_penyakit }}</strong>
      </article>
      <article>
        <span>Jumlah Diagnosis</span>
        <strong>{{ ringkasan.jumlah_diagnosa }}</strong>
      </article>
      <article>
        <span>Laki-laki</span>
        <strong>{{ ringkasan.laki_laki }}</strong>
      </article>
      <article>
        <span>Perempuan</span>
        <strong>{{ ringkasan.perempuan }}</strong>
      </article>
      <article>
        <span>Meninggal</span>
        <strong>{{ ringkasan.meninggal }}</strong>
      </article>
    </div>

    <div class="visit-report-tools">
      <span>{{ data.length }} penyakit ditampilkan</span>
      <button class="chart-button" @click="grafikVisible = !grafikVisible">
        <X v-if="grafikVisible" :size="15" />
        <BarChart3 v-else :size="15" />
        {{ grafikVisible ? "Tutup Grafik" : "Lihat Grafik" }}
      </button>
      <button @click="excel"><Download :size="15" /> Excel</button>
      <button @click="cetak"><Printer :size="15" /> Cetak</button>
    </div>

    <section v-if="grafikVisible" class="visit-analytics">
      <header>
        <div>
          <span>Visualisasi Data</span>
          <h2>10 Penyakit Terbanyak</h2>
          <p>Grafik mengikuti periode dan status pelayanan yang dipilih.</p>
        </div>
        <BarChart3 :size="25" />
      </header>
      <article class="visit-chart-card disease-chart-card">
        <div v-if="grafik.length" class="horizontal-chart">
          <div v-for="item in grafik" :key="item.kode">
            <label>
              <span :title="item.nama">{{ item.kode }} · {{ item.nama }}</span>
              <b>{{ item.jumlah }}</b>
            </label>
            <i><u :style="{ width: `${item.persen}%` }"></u></i>
          </div>
        </div>
        <em v-else>Belum ada data untuk divisualisasikan.</em>
      </article>
    </section>

    <div v-if="error" class="patient-error">{{ error }}</div>
    <DataTable
      v-else
      class="visit-report-table"
      :rows="data"
      data-key="kode"
      :loading="loading"
      empty-message="Data penyakit tidak ditemukan."
    >
      <Column header="No." style="width: 60px; text-align: center">
        <template #body="{ index }"><strong>{{ index + 1 }}</strong></template>
      </Column>
      <Column field="kode" header="Kode" style="min-width: 100px" />
      <Column field="nama" header="Nama Penyakit" style="min-width: 360px" />
      <Column field="diagnosa_lain" header="Diagnosa Lain" />
      <Column field="laki_hidup" header="Lk2 (Hidup)" />
      <Column field="perempuan_hidup" header="Pr (Hidup)" />
      <Column field="laki_meninggal" header="Lk2 (Mati)" />
      <Column field="perempuan_meninggal" header="Pr (Mati)" />
      <Column field="jumlah" header="Jumlah">
        <template #body="{ data: item }"><strong>{{ item.jumlah }}</strong></template>
      </Column>
    </DataTable>
  </section>
</template>

<style src="../KunjunganRalan/kunjungan-ralan.css" scoped></style>
<style src="./sepuluh-penyakit.css" scoped></style>
