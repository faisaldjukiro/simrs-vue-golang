<script setup lang="ts">
import { CircleCheck, Pencil, Plus, RefreshCw, Save, Trash2, TriangleAlert, Wind, Wrench, X } from "@lucide/vue"
import Column from "primevue/column"
import DataTable from "../../../Components/Ui/DataTable.vue"
import FormInput from "../../../Components/Ui/FormInput.vue"
import TableSearch from "../../../Components/Ui/TableSearch.vue"
import { useMasterVentilator } from "./useMasterVentilator"

const props = defineProps({ token: { type: String, required: true } })

const {
  records,
  loading,
  saving,
  keyword,
  editing,
  formTerbuka,
  statuses,
  empty,
  form,
  rows,
  jumlahTersedia,
  jumlahDigunakan,
  jumlahPerhatian,
  statusClass,
  reset,
  tambah,
  edit,
  load,
  save,
  remove,
} = useMasterVentilator(props)
</script>

<template>
  <section class="master-page">
    <header class="module-header">
      <div class="module-identity"><div class="module-icon"><Wind :size="25" /></div><div><span class="eyebrow">MASTER SETTING</span><h1>Master Ventilator</h1><p>Kelola perangkat, lokasi, status, dan jadwal pemeliharaan ventilator.</p></div></div>
      <div class="header-actions"><span class="configuration-status"><i /> {{ records.length }} perangkat terdaftar</span><button type="button" class="secondary-button" :disabled="loading" @click="load"><RefreshCw :size="16" :class="{ spin: loading }" /> {{ loading ? 'Memuat...' : 'Muat Data' }}</button></div>
    </header>

    <div class="summary-grid">
      <article><span>Total Perangkat</span><strong>{{ records.length }}</strong><Wind :size="21" /></article>
      <article><span>Ventilator Tersedia</span><strong>{{ jumlahTersedia }}</strong><CircleCheck :size="21" /></article>
      <article><span>Sedang Digunakan</span><strong>{{ jumlahDigunakan }}</strong><Wrench :size="21" /></article>
      <article><span>Perlu Perhatian</span><strong>{{ jumlahPerhatian }}</strong><TriangleAlert :size="21" /></article>
    </div>

    <section v-if="formTerbuka" class="module-card form-card">
      <div class="card-header"><div><span class="eyebrow">DATA PERANGKAT</span><h2>{{ editing ? 'Edit Ventilator' : 'Tambah Ventilator' }}</h2><p>Lengkapi identitas perangkat dan informasi pemeliharaannya.</p></div><button type="button" class="icon-button" title="Tutup form" @click="reset"><X :size="18" /></button></div>
      <form class="form-area" @submit.prevent="save"><fieldset :disabled="saving"><div class="master-grid">
        <FormInput v-model="form.kode_ventilator" label="Kode Ventilator" placeholder="Dibuat otomatis saat disimpan" disabled />
        <FormInput v-model="form.nama" label="Nama Ventilator" placeholder="Nama perangkat" required />
        <FormInput v-model="form.status" label="Status" jenis="select" :options="statuses" required />
        <FormInput v-model="form.merk" label="Merk" placeholder="Merk ventilator" /><FormInput v-model="form.model" label="Model" placeholder="Model perangkat" /><FormInput v-model="form.nomor_seri" label="Nomor Seri" placeholder="Nomor seri perangkat" />
        <FormInput v-model="form.ruangan" label="Ruangan" placeholder="Lokasi ventilator" /><FormInput v-model="form.tanggal_maintenance_terakhir" label="Pemeliharaan Terakhir" type="date" /><FormInput v-model="form.tanggal_maintenance_berikutnya" label="Pemeliharaan Berikutnya" type="date" />
        <FormInput v-model="form.keterangan" class="span-all" label="Keterangan" jenis="textarea" :rows="3" placeholder="Catatan kondisi atau pemeliharaan perangkat" />
      </div><div class="form-actions"><button type="button" class="secondary-button" @click="reset"><X :size="16" /> Batal / Reset</button><button type="submit" class="primary-button"><Save :size="16" /> {{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan Ventilator' }}</button></div></fieldset></form>
    </section>

    <section class="module-card history-card">
      <div class="card-header history-header"><div><span class="eyebrow">DAFTAR PERANGKAT</span><h2>Ventilator Rumah Sakit</h2><p>Lihat lokasi, status, dan jadwal pemeliharaan setiap perangkat.</p></div><button type="button" class="primary-button" @click="tambah"><Plus :size="16" /> Tambah Ventilator</button></div>
      <div class="table-toolbar"><TableSearch v-model="keyword" placeholder="Cari kode, nama, seri, merk, atau ruangan..." /><span class="record-count">{{ rows.length }} perangkat</span></div>
      <div v-if="loading" class="table-loading"><RefreshCw :size="22" class="spin" /><strong>Memuat master ventilator...</strong><span>Mohon tunggu, data perangkat sedang dibaca.</span></div>
      <DataTable v-else class="master-ventilator-table" :rows="rows" data-key="kode_ventilator" empty-message="Master ventilator belum tersedia." paginator :rows-per-page="10" :rows-per-page-options="[10, 25]">
        <Column header="PERANGKAT"><template #body="{ data }"><div class="table-primary"><strong>{{ data.nama }}</strong><span>{{ data.kode_ventilator }} · {{ data.nomor_seri || 'Tanpa nomor seri' }}</span></div></template></Column>
        <Column header="MERK / MODEL"><template #body="{ data }"><div class="table-primary"><strong>{{ data.merk || '-' }}</strong><span>{{ data.model || '-' }}</span></div></template></Column>
        <Column header="RUANGAN"><template #body="{ data }"><strong>{{ data.ruangan || '-' }}</strong></template></Column>
        <Column header="STATUS"><template #body="{ data }"><em class="status-pill" :class="statusClass(data.status)">{{ data.status || '-' }}</em></template></Column>
        <Column header="PEMELIHARAAN"><template #body="{ data }"><div class="table-primary"><strong>Berikutnya {{ data.tanggal_maintenance_berikutnya || '-' }}</strong><span>Terakhir {{ data.tanggal_maintenance_terakhir || '-' }}</span></div></template></Column>
        <Column header="AKSI"><template #body="{ data }"><div class="table-actions"><button type="button" @click="edit(data)"><Pencil :size="14" /> Edit</button><button type="button" class="delete-button" @click="remove(data)"><Trash2 :size="14" /> Hapus</button></div></template></Column>
      </DataTable>
    </section>
  </section>
</template>

<style src="./master-ventilator.css" scoped></style>
