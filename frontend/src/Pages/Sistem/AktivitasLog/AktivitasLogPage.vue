<script setup lang="ts">
import { Eye, Filter, ListChecks, Search, ShieldCheck, ShieldX } from "@lucide/vue"
import Column from "primevue/column"
import Dialog from "primevue/dialog"
import DataTable from "../../../Components/Ui/DataTable.vue"
import DatePicker from "../../../Components/Ui/DatePicker.vue"
import Select from "../../../Components/Ui/Select.vue"
import { useAktivitasLog } from "./useAktivitasLog"

const props = defineProps({ token: { type: String, default: '' } })

const {
  loading,
  error,
  hasil,
  detail,
  filter,
  pilihanAksi,
  pilihanHasil,
  jumlahBerhasil,
  jumlahGagal,
  waktu,
  muat,
  gantiHalaman,
  jsonRapi,
  ubahDialogDetail,
} = useAktivitasLog(props)
</script>

<template>
  <section class="activity-module">
    <header class="activity-header">
      <div>
        <span>Audit Trail SIRAPI</span>
        <h1>Log Aktivitas</h1>
        <p>Riwayat login, akses data, perubahan, kegagalan, alamat IP, dan waktu proses.</p>
      </div>
      <ListChecks :size="27" />
    </header>

    <form class="activity-filter" @submit.prevent="muat(1)">
      <label class="activity-date">
        <span>Tanggal Mulai</span>
        <DatePicker v-model="filter.tanggal_mulai" placeholder="Tanggal mulai" />
      </label>
      <label class="activity-date">
        <span>Sampai Dengan</span>
        <DatePicker v-model="filter.tanggal_selesai" placeholder="Sampai dengan" />
      </label>
      <label class="activity-select">
        <span>Aktivitas</span>
        <Select v-model="filter.aksi" :options="pilihanAksi" placeholder="Semua Aktivitas" append-to="body" overlay-class="activity-select-overlay" />
      </label>
      <label class="activity-select">
        <span>Hasil</span>
        <Select v-model="filter.berhasil" :options="pilihanHasil" placeholder="Semua Hasil" append-to="body" overlay-class="activity-select-overlay" />
      </label>
      <label class="activity-search activity-keyword">
        <span>Pencarian Data</span>
        <div>
          <Search :size="17" />
          <input v-model="filter.kata_kunci" placeholder="Cari No. Rawat, No. RM, No. SEP, target, endpoint..." />
        </div>
      </label>
      <label class="activity-search activity-user">
        <span>User</span>
        <div>
          <Search :size="17" />
          <input v-model="filter.username" placeholder="Cari username..." />
        </div>
      </label>
      <label class="activity-search">
        <span>Nama Modul</span>
        <div>
          <Search :size="17" />
          <input v-model="filter.modul" placeholder="Contoh: cppt" />
        </div>
      </label>
      <button type="submit" :disabled="loading">
        <Filter :size="17" /> Tampilkan
      </button>
    </form>

    <div class="activity-summary">
      <article><ListChecks :size="19" /><div><span>Total Hasil</span><strong>{{ hasil.total }}</strong></div></article>
      <article><ShieldCheck :size="19" /><div><span>Berhasil di Halaman Ini</span><strong>{{ jumlahBerhasil }}</strong></div></article>
      <article><ShieldX :size="19" /><div><span>Gagal di Halaman Ini</span><strong>{{ jumlahGagal }}</strong></div></article>
    </div>

    <div v-if="error" class="patient-error">{{ error }}</div>
    <DataTable
      v-else
      class="activity-table"
      :rows="hasil.data"
      data-key="id"
      :loading="loading"
      paginator
      lazy
      :first="(hasil.halaman - 1) * hasil.batas"
      :rows-per-page="hasil.batas"
      :rows-per-page-options="[10, 25, 50, 100]"
      :total-records="hasil.total"
      empty-message="Belum ada aktivitas pada filter ini."
      @page="gantiHalaman"
    >
      <Column header="Waktu / User" style="min-width: 240px">
        <template #body="{ data }">
          <strong>{{ waktu(data.created_at) }}</strong>
          <span class="activity-user-name">{{ data.nama_user || data.username }}</span>
          <small>{{ data.username }} | {{ data.ip_address || '-' }}</small>
        </template>
      </Column>
      <Column header="Aktivitas" style="width: 120px">
        <template #body="{ data }"><span class="activity-badge" :class="`aksi-${data.aksi.toLowerCase()}`">{{ data.aksi }}</span></template>
      </Column>
      <Column header="Modul / Endpoint" style="min-width: 250px">
        <template #body="{ data }"><strong>{{ data.modul }}</strong><small>{{ data.method }} {{ data.endpoint }}</small></template>
      </Column>
      <Column header="Target" style="min-width: 165px">
        <template #body="{ data }"><strong>{{ data.target_id || '-' }}</strong><small>{{ data.tabel_target || 'Tidak ada tabel target' }}</small></template>
      </Column>
      <Column header="Hasil" style="width: 145px">
        <template #body="{ data }">
          <span class="result-badge" :class="{ gagal: !data.berhasil }">{{ data.berhasil ? 'Berhasil' : 'Gagal' }}</span>
          <small>HTTP {{ data.response_status }} | {{ data.durasi_ms }} ms</small>
        </template>
      </Column>
      <Column header="Detail" style="width: 80px; text-align: center">
        <template #body="{ data }"><button class="detail-button" title="Lihat detail" @click="detail = data"><Eye :size="17" /></button></template>
      </Column>
    </DataTable>
  </section>

  <Dialog :visible="Boolean(detail)" modal header="Detail Log Aktivitas" class="activity-dialog" :style="{ width: 'min(920px, 94vw)' }" @update:visible="ubahDialogDetail">
    <div v-if="detail" class="activity-detail">
      <dl>
        <div><dt>Request ID</dt><dd>{{ detail.request_id }}</dd></div>
        <div><dt>Nama Pegawai</dt><dd>{{ detail.nama_user || detail.username }}</dd></div>
        <div><dt>NIK / Username</dt><dd>{{ detail.username }}</dd></div>
        <div><dt>Endpoint</dt><dd>{{ detail.method }} {{ detail.endpoint }}</dd></div>
        <div><dt>Hasil</dt><dd>{{ detail.berhasil ? 'Berhasil' : detail.pesan_error || 'Gagal' }}</dd></div>
      </dl>
      <section><h3>Data Sebelum</h3><pre>{{ jsonRapi(detail.data_sebelum) }}</pre></section>
      <section><h3>Data Sesudah</h3><pre>{{ jsonRapi(detail.data_sesudah) }}</pre></section>
      <section><h3>Data Permintaan</h3><pre>{{ jsonRapi(detail.request_data) }}</pre></section>
      <section><h3>Parameter URL</h3><pre>{{ jsonRapi(detail.query_params) }}</pre></section>
    </div>
  </Dialog>
</template>

<style src="./aktivitas-log.css" scoped></style>
