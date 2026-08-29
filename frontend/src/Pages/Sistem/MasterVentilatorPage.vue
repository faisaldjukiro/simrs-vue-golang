<script setup>
import { CircleCheck, Pencil, Plus, RefreshCw, Save, Trash2, TriangleAlert, Wind, Wrench, X } from '@lucide/vue'
import Column from 'primevue/column'
import { computed, onMounted, reactive, ref } from 'vue'
import DataTable from '../../Components/Ui/DataTable.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import TableSearch from '../../Components/Ui/TableSearch.vue'
import { hapusMasterVentilator, masterVentilatorData, simpanMasterVentilator, ubahMasterVentilator } from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({ token: { type: String, required: true } })
const notifikasi = useNotifikasi()
const records = ref([]), loading = ref(false), saving = ref(false), keyword = ref(''), editing = ref(''), formTerbuka = ref(false)
const statuses = ['Tersedia', 'Digunakan', 'Pemeliharaan', 'Rusak', 'Nonaktif'].map((value) => ({ label: value, value }))
const empty = () => ({ kode_ventilator: '', nama: '', merk: '', model: '', nomor_seri: '', ruangan: '', status: 'Tersedia', tanggal_maintenance_terakhir: '', tanggal_maintenance_berikutnya: '', keterangan: '' })
const form = reactive(empty())
const rows = computed(() => { const q = keyword.value.toLowerCase().trim(); return q ? records.value.filter((item) => Object.values(item).join(' ').toLowerCase().includes(q)) : records.value })
const jumlahTersedia = computed(() => records.value.filter((item) => item.status === 'Tersedia').length)
const jumlahDigunakan = computed(() => records.value.filter((item) => item.status === 'Digunakan').length)
const jumlahPerhatian = computed(() => records.value.filter((item) => ['Pemeliharaan', 'Rusak', 'Nonaktif'].includes(item.status)).length)

function statusClass(status) { if (status === 'Tersedia') return 'success'; if (status === 'Digunakan') return 'active'; if (['Rusak', 'Nonaktif'].includes(status)) return 'danger'; return 'pending' }
function reset() { Object.assign(form, empty()); editing.value = ''; formTerbuka.value = false }
function tambah() { Object.assign(form, empty()); editing.value = ''; formTerbuka.value = true }
function edit(item) {
  editing.value = item.kode_ventilator
  Object.assign(form, empty(), item, { tanggal_maintenance_terakhir: item.tanggal_maintenance_terakhir || '', tanggal_maintenance_berikutnya: item.tanggal_maintenance_berikutnya || '' })
  formTerbuka.value = true
  window.scrollTo({ top: 0, behavior: 'smooth' })
}
async function load() { loading.value = true; try { records.value = await masterVentilatorData(props.token) || [] } catch (error) { notifikasi.gagal(error.message) } finally { loading.value = false } }
async function save() {
  if (saving.value) return
  if (!form.nama.trim()) return notifikasi.peringatan('Nama ventilator wajib diisi.')
  saving.value = true
  try {
    const hasil = editing.value ? await ubahMasterVentilator(props.token, editing.value, form) : await simpanMasterVentilator(props.token, form)
    notifikasi.sukses(hasil.pesan); reset(); await load()
  } catch (error) { notifikasi.gagal(error.message) } finally { saving.value = false }
}
async function remove(item) {
  if (!window.confirm(`Hapus ventilator ${item.nama}?`)) return
  try { const hasil = await hapusMasterVentilator(props.token, item.kode_ventilator); notifikasi.sukses(hasil.pesan); await load() } catch (error) { notifikasi.gagal(error.message) }
}
onMounted(load)
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

<style scoped>
.master-page{position:relative;z-index:1;display:grid;width:calc(100% - 40px);box-sizing:border-box;gap:16px;margin:20px;color:var(--text)}
.module-header,.module-card{overflow:hidden;border:1px solid var(--line);border-radius:16px;background:var(--surface)}.module-header{display:flex;align-items:center;justify-content:space-between;gap:20px;padding:20px 22px}
.module-identity,.header-actions{display:flex;align-items:center;gap:14px}.module-icon{display:grid;width:50px;height:50px;place-items:center;flex:0 0 auto;border-radius:14px;background:rgba(13,148,136,.12);color:#0d9488}.eyebrow{font-size:10px;font-weight:700;letter-spacing:.12em;color:#0d9488}
.module-header h1,.card-header h2{margin:3px 0;color:var(--text)}.module-header h1{font-size:24px}.card-header h2{font-size:18px}.module-header p,.card-header p{margin:0;color:var(--muted);font-size:13px;line-height:1.5}.configuration-status{display:inline-flex;align-items:center;gap:7px;white-space:nowrap;color:var(--muted);font-size:12px}.configuration-status i{width:8px;height:8px;border-radius:50%;background:#10b981}
.summary-grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:12px}.summary-grid article{position:relative;display:grid;gap:5px;padding:16px 18px;border:1px solid var(--line);border-radius:14px;background:var(--surface)}.summary-grid span{color:var(--muted);font-size:12px}.summary-grid strong{font-size:23px}.summary-grid svg{position:absolute;right:17px;top:50%;transform:translateY(-50%);color:#0d9488}
.card-header{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:17px 20px;border-bottom:1px solid var(--line)}.form-area{padding:20px;background:var(--surface-soft)}fieldset{min-width:0;margin:0;padding:0;border:0}.master-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:13px}.span-all{grid-column:1/-1}.form-actions{display:flex;justify-content:flex-end;gap:9px;margin-top:18px;padding-top:16px;border-top:1px solid var(--line)}
.primary-button,.secondary-button,.icon-button,.table-actions button{display:inline-flex;align-items:center;justify-content:center;gap:7px;min-height:40px;border:1px solid var(--line);border-radius:10px;padding:0 14px;font:inherit;font-size:13px;font-weight:700;cursor:pointer}.primary-button{border-color:#0d9488;background:#0d9488;color:#fff}.secondary-button,.icon-button,.table-actions button{background:var(--surface-soft);color:var(--text)}.icon-button{width:40px;padding:0}.secondary-button:hover,.icon-button:hover,.table-actions button:hover{border-color:#0d9488;color:#0d9488}button:disabled{opacity:.5;cursor:not-allowed}
.table-toolbar{display:flex;align-items:center;justify-content:space-between;gap:12px;padding:14px 14px 0}.table-toolbar :deep(.table-search){width:min(100%,520px)}.record-count{flex:0 0 auto;border-radius:999px;padding:6px 10px;background:var(--surface-soft);color:var(--muted);font-size:11px}.history-card{position:relative}.history-card :deep(.data-table-wrap){position:relative;margin:14px}.table-loading{display:flex;min-height:210px;flex-direction:column;align-items:center;justify-content:center;gap:8px;margin:14px;border:1px dashed var(--line);border-radius:13px;color:var(--muted);background:var(--surface-soft);text-align:center}.table-loading svg{color:#0d9488}.table-loading strong{color:var(--text);font-size:14px}.table-loading span{font-size:11px}.table-primary{display:grid;gap:3px;min-width:145px}.table-primary strong{color:var(--text)}.table-primary span{color:var(--muted);font-size:11px}.table-actions{display:flex;gap:7px}.table-actions button{min-height:34px;padding:0 10px;font-size:11px}.table-actions .delete-button{color:#dc2626}
.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-table){width:100%;min-width:980px;table-layout:fixed}
.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-thead>tr>th){padding:13px 14px;color:#fff;background:#0f6367;text-align:left}
.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-thead>tr>th:nth-child(1)),.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-tbody>tr>td:nth-child(1)){width:25%;text-align:left}
.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-thead>tr>th:nth-child(2)),.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-tbody>tr>td:nth-child(2)){width:18%;text-align:left}
.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-thead>tr>th:nth-child(3)),.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-tbody>tr>td:nth-child(3)){width:15%;text-align:left}
.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-thead>tr>th:nth-child(4)),.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-tbody>tr>td:nth-child(4)){width:12%;text-align:left}
.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-thead>tr>th:nth-child(5)),.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-tbody>tr>td:nth-child(5)){width:19%;text-align:left}
.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-thead>tr>th:nth-child(6)),.history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-tbody>tr>td:nth-child(6)){width:11%;text-align:left}
:global(.theme-dark) .history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-thead>tr>th),:global(.sirapi-dark) .history-card :deep(.master-ventilator-table.data-table-wrap .p-datatable-thead>tr>th){border-color:rgba(148,163,184,.14);background:#0d4f55}
.status-pill{display:inline-flex;width:max-content;border-radius:999px;padding:5px 9px;font-size:10px;font-style:normal;font-weight:700}.status-pill.success{background:rgba(13,148,136,.13);color:#0d9488}.status-pill.active{background:rgba(59,130,246,.14);color:#2563eb}.status-pill.pending{background:rgba(245,158,11,.14);color:#d97706}.status-pill.danger{background:rgba(239,68,68,.13);color:#dc2626}.spin{animation:spin 1s linear infinite}@keyframes spin{to{transform:rotate(360deg)}}
@media(max-width:1050px){.summary-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.master-grid{grid-template-columns:repeat(2,minmax(0,1fr))}}
@media(max-width:760px){.module-header{align-items:flex-start}.header-actions{align-items:flex-end;flex-direction:column}.master-grid{grid-template-columns:1fr}.span-all{grid-column:auto}.history-header{align-items:flex-start}.table-toolbar{align-items:stretch;flex-direction:column}.record-count{align-self:flex-start}}
@media(max-width:560px){.master-page{width:calc(100% - 20px);margin:10px}.module-header{display:grid;padding:16px}.header-actions{align-items:stretch}.configuration-status{justify-content:center}.module-header .secondary-button,.history-header .primary-button{width:100%}.summary-grid{grid-template-columns:1fr}.card-header,.form-area{padding-left:14px;padding-right:14px}.history-header{display:grid}.form-actions{flex-direction:column-reverse}.form-actions button{width:100%}.table-actions{flex-direction:column}}
</style>
