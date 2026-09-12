<script setup lang="ts">
import { BarChart3, BedDouble, Download, Filter, Printer, X } from "@lucide/vue";
import Column from "primevue/column";
import DataTable from "../../../Components/Ui/DataTable.vue";
import FormInput from "../../../Components/Ui/FormInput.vue";
import InputPencarian from "../../../Components/Ui/InputPencarian.vue";
import { usePenggunaanBed } from "./usePenggunaanBed";
const props = defineProps({ token: { type: String, required: true } });
const x = usePenggunaanBed(props);
</script>
<template>
  <section class="visit-report-module bed-report">
    <header class="visit-report-header">
      <div>
        <span>Laporan Rawat Inap</span>
        <h1>Penggunaan Bed & Frekuensi</h1>
      </div>
      <BedDouble :size="30" />
    </header>
    <form class="visit-report-filter bed-filter" @submit.prevent="x.muat">
      <FormInput v-model="x.filter.tanggal_mulai" label="Tanggal Keluar Mulai" type="date" required />
      <FormInput v-model="x.filter.tanggal_selesai" label="Tanggal Keluar Sampai" type="date" required />
      <InputPencarian v-model="x.bangsal.value" label="Bangsal" placeholder="Semua bangsal" :search="x.cariBangsal" />
      <FormInput v-model="x.pencarian.value" label="Pencarian Tabel" placeholder="Kode atau nama bangsal..." />
      <button :disabled="x.loading.value">
        <Filter :size="16" />
        {{ x.loading.value ? "Memuat..." : "Tampilkan" }}
      </button>
    </form>
    <div class="visit-report-summary">
      <article>
        <span>Jumlah Bangsal</span>
        <strong>{{ x.ringkasan.value.jumlah_bangsal || 0 }}</strong>
      </article>
      <article>
        <span>Total Bed Tersedia</span>
        <strong>{{ x.ringkasan.value.total_bed || 0 }}</strong>
      </article>
      <article>
        <span>Pasien Keluar</span>
        <strong>{{ x.ringkasan.value.total_pasien_keluar || 0 }}</strong>
      </article>
      <article>
        <span>Rata-rata Frekuensi</span>
        <strong>
          {{ Number(x.ringkasan.value.rata_rata_frekuensi || 0).toFixed(2) }}
        </strong>
      </article>
    </div>
    <div class="visit-report-tools">
      <span>
        {{ x.dataTersaring.value.length }} dari {{ x.data.value.length }} bangsal
      </span>
      <button
        type="button"
        class="chart-button"
        :class="{ active: x.grafikVisible.value }"
        :aria-expanded="x.grafikVisible.value"
        aria-controls="grafik-penggunaan-bed"
        @click="x.grafikVisible.value = !x.grafikVisible.value"
      >
        <X v-if="x.grafikVisible.value" :size="16" />
        <BarChart3 v-else :size="16" />
        {{ x.grafikVisible.value ? "Tutup Grafik" : "Lihat Grafik" }}
      </button>
      <button :disabled="x.loading.value || !x.data.value.length" @click="x.excel">
        <Download :size="15" />
        Excel
      </button>
      <button @click="x.cetak">
        <Printer :size="15" />
        Cetak
      </button>
    </div>
    <section
      v-if="x.grafikVisible.value"
      id="grafik-penggunaan-bed"
      class="visit-analytics"
    >
      <header>
        <div>
          <span>Visualisasi</span>
          <h2>Frekuensi Penggunaan Bed</h2>
          <p>Frekuensi rata-rata penggunaan bed pada setiap bangsal.</p>
        </div>
        <BedDouble />
      </header>
      <article class="visit-chart-card bed-chart">
        <div v-if="x.grafik.value.length" class="horizontal-chart">
          <div v-for="i in x.grafik.value" :key="i.kode_bangsal">
            <label>
              <span>{{ i.nama_bangsal }}</span>
              <b>{{ Number(i.frekuensi).toFixed(2) }}</b>
            </label>
            <i>
              <u :style="{ width: `${i.persen}%` }"></u>
            </i>
          </div>
        </div>
        <em v-else>Belum ada data untuk divisualisasikan.</em>
      </article>
    </section>
    <div v-if="x.error.value" class="patient-error">{{ x.error.value }}</div>
    <DataTable v-else class="visit-report-table bed-usage-table" :rows="x.dataTersaring.value" data-key="kode_bangsal"
      :loading="x.loading.value" empty-message="Data penggunaan bed tidak ditemukan.">
      <Column header="No." header-class="center-align-header" body-class="center-align-body"
        style="width: 70px; min-width: 70px">
        <template #body="{ index }">
          <strong>{{ index + 1 }}</strong>
        </template>
      </Column>
      <Column field="kode_bangsal" header="Kode Bangsal" style="min-width: 140px" />
      <Column field="nama_bangsal" header="Nama Bangsal" style="min-width: 280px" />
      <Column field="total_bed" header="Total Bed Tersedia" header-class="right-align-header"
        body-class="right-align-body" style="width: 150px; min-width: 150px" />
      <Column field="pasien_keluar" header="Pasien Keluar" header-class="right-align-header"
        body-class="right-align-body" style="width: 140px; min-width: 140px" />
      <Column header="Frekuensi" header-class="right-align-header" body-class="right-align-body"
        style="width: 130px; min-width: 130px">
        <template #body="{ data: r }">
          <strong class="bed-number">{{ Number(r.frekuensi).toFixed(2) }}</strong>
        </template>
      </Column>
    </DataTable>
  </section>
</template>
<style src="../../../Components/Ui/report.css" scoped></style>
<style src="./penggunaan-bed.css" scoped></style>
