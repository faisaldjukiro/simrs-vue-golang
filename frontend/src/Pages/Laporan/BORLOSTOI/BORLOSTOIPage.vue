<script setup lang="ts">
import { BedDouble, Download, Filter, Printer } from "@lucide/vue";
import Column from "primevue/column";
import DataTable from "../../../Components/Ui/DataTable.vue";
import FormInput from "../../../Components/Ui/FormInput.vue";
import { useBORLOSTOI } from "./useBORLOSTOI";

const props = defineProps({ token: { type: String, required: true } });
const x = useBORLOSTOI(props);
</script>

<template>
  <section class="visit-report-module indicator-report">
    <header class="visit-report-header">
      <div>
        <span>Laporan Rawat Inap</span>
        <h1>BOR, LOS & TOI</h1>
        <p>Indikator pemanfaatan tempat tidur rawat inap per bulan.</p>
      </div>
      <BedDouble :size="30" />
    </header>

    <form class="visit-report-filter indicator-filter" @submit.prevent="x.muat">
      <FormInput
        v-model="x.filter.bulan_mulai"
        label="Bulan Mulai"
        type="month"
        required
      />
      <FormInput
        v-model="x.filter.bulan_selesai"
        label="Sampai Bulan"
        type="month"
        required
      />
      <FormInput
        v-model="x.pencarian.value"
        label="Pencarian Tabel"
        placeholder="Cari periode..."
      />
      <button :disabled="x.loading.value">
        <Filter :size="16" />
        {{ x.loading.value ? "Memuat..." : "Tampilkan" }}
      </button>
    </form>

    <div class="visit-report-summary">
      <article>
        <span>Jumlah Periode</span>
        <strong>{{ x.ringkasan.value.jumlah_periode || 0 }}</strong>
      </article>
      <article>
        <span>Rata-rata BOR</span>
        <strong>{{ x.ringkasan.value.rata_bor || 0 }}%</strong>
      </article>
      <article>
        <span>Rata-rata LOS</span>
        <strong>{{ x.ringkasan.value.rata_los || 0 }} hari</strong>
      </article>
      <article>
        <span>Rata-rata TOI</span>
        <strong>{{ x.ringkasan.value.rata_toi || 0 }} hari</strong>
      </article>
    </div>

    <div class="visit-report-tools">
      <span>
        {{ x.dataTersaring.value.length }} dari {{ x.data.value.length }} periode
      </span>
      <button @click="x.excel">
        <Download :size="15" />
        Excel
      </button>
      <button @click="x.cetak">
        <Printer :size="15" />
        Cetak
      </button>
    </div>

    <div v-if="x.error.value" class="patient-error">{{ x.error.value }}</div>
    <DataTable
      v-else
      class="visit-report-table indicator-table"
      :rows="x.dataTersaring.value"
      data-key="periode"
      :loading="x.loading.value"
      empty-message="Data indikator tidak ditemukan."
    >
      <Column
        header="No."
        header-class="center-column"
        body-class="center-column"
        style="width: 65px; min-width: 65px"
      >
        <template #body="{ index }">
          <strong>{{ index + 1 }}</strong>
        </template>
      </Column>
      <Column field="periode" header="Periode" style="min-width: 160px" />
      <Column field="total_tempat_tidur" header="Tempat Tidur" />
      <Column field="jumlah_hari" header="Jumlah Hari" />
      <Column field="total_hari_perawatan" header="Hari Perawatan" />
      <Column field="pasien_keluar" header="Pasien Keluar" />
      <Column header="BOR">
        <template #body="{ data: item }">
          <strong>{{ item.bor }}%</strong>
        </template>
      </Column>
      <Column header="LOS">
        <template #body="{ data: item }">
          <strong>{{ item.los }} hari</strong>
        </template>
      </Column>
      <Column header="TOI">
        <template #body="{ data: item }">
          <strong>{{ item.toi }} hari</strong>
        </template>
      </Column>
    </DataTable>
  </section>
</template>

<style src="../../../Components/Ui/report.css" scoped></style>
<style src="./bor-los-toi.css" scoped></style>
