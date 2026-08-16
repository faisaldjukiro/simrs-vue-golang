<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { Eye, Filter, ListChecks, Search, ShieldCheck, ShieldX } from '@lucide/vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import DataTable from '../../Components/Ui/DataTable.vue'
import DatePicker from '../../Components/Ui/DatePicker.vue'
import Select from '../../Components/Ui/Select.vue'
import { aktivitasLogData } from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({ token: { type: String, default: '' } })
const notifikasi = useNotifikasi()
const hariIni = new Date()
const tujuhHariLalu = new Date(hariIni)
tujuhHariLalu.setDate(tujuhHariLalu.getDate() - 6)

const loading = ref(false)
const error = ref('')
const hasil = ref({ data: [], halaman: 1, batas: 25, total: 0, total_halaman: 0 })
const detail = ref(null)
const filter = reactive({
  tanggal_mulai: tujuhHariLalu,
  tanggal_selesai: hariIni,
  kata_kunci: '',
  username: '',
  aksi: '',
  modul: '',
  berhasil: '',
})

const pilihanAksi = [
  { label: 'Semua Aktivitas', value: '' },
  ...['LOGIN', 'LOGOUT', 'LIHAT', 'TAMBAH', 'UBAH', 'HAPUS', 'PROSES'].map((value) => ({ label: value, value })),
]
const pilihanHasil = [
  { label: 'Semua Hasil', value: '' },
  { label: 'Berhasil', value: '1' },
  { label: 'Gagal', value: '0' },
]

const jumlahBerhasil = computed(() => hasil.value.data.filter((item) => item.berhasil).length)
const jumlahGagal = computed(() => hasil.value.data.filter((item) => !item.berhasil).length)

function tanggalAPI(value) {
  if (!value) return ''
  const date = value instanceof Date ? value : new Date(value)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function waktu(value) {
  if (!value) return '-'
  return new Intl.DateTimeFormat('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'medium',
    timeZone: 'Asia/Makassar',
  }).format(new Date(value))
}

function parameter(halaman = 1, batas = hasil.value.batas || 25) {
  return {
    halaman,
    batas,
    tanggal_mulai: tanggalAPI(filter.tanggal_mulai),
    tanggal_selesai: tanggalAPI(filter.tanggal_selesai),
    kata_kunci: filter.kata_kunci.trim(),
    username: filter.username.trim(),
    aksi: filter.aksi,
    modul: filter.modul.trim(),
    berhasil: filter.berhasil,
  }
}

async function muat(halaman = 1, batas = hasil.value.batas || 25) {
  if (loading.value || !props.token) return
  loading.value = true
  error.value = ''
  try {
    hasil.value = await aktivitasLogData(props.token, parameter(halaman, batas))
  } catch (err) {
    error.value = err.message
    notifikasi.gagal(err.message || 'Log aktivitas tidak dapat dibaca.')
  } finally {
    loading.value = false
  }
}

function gantiHalaman(event) {
  muat((event.page ?? 0) + 1, event.rows ?? hasil.value.batas)
}

function jsonRapi(value) {
  if (!value) return 'Tidak ada data.'
  return JSON.stringify(value, null, 2)
}

function ubahDialogDetail(terbuka) {
  if (!terbuka) detail.value = null
}

onMounted(() => muat())
</script>

<template>
  <section class="activity-module">
    <header class="activity-header">
      <div>
        <span>Audit Trail SIRAVA</span>
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

<style scoped>
.activity-module{position:relative;z-index:1;display:flex;width:calc(100% - 40px);min-height:520px;flex:1;flex-direction:column;margin:20px;padding:22px;border:1px solid var(--line);border-radius:24px;color:var(--text);background:var(--surface);box-shadow:var(--shadow)}
.activity-header{display:flex;align-items:center;justify-content:space-between;gap:18px;padding-bottom:18px;border-bottom:1px solid var(--line)}
.activity-header span{color:#0d9488;font-size:10px;font-weight:750;letter-spacing:.15em;text-transform:uppercase}
.activity-header h1{margin:5px 0 0;font-size:27px;letter-spacing:-.02em}
.activity-header p{margin:6px 0 0;color:var(--muted);font-size:12px;line-height:1.6}
.activity-header>svg{box-sizing:content-box;flex:0 0 auto;padding:10px;border-radius:13px;color:#0d9488;background:rgba(20,184,166,.1)}
.activity-filter{display:grid;grid-template-columns:repeat(12,minmax(0,1fr));gap:12px;margin-top:16px;padding:15px;border:1px solid var(--line);border-radius:16px;background:var(--surface-soft)}
.activity-filter label{display:flex;min-width:0;flex-direction:column;gap:7px}
.activity-filter label>span{color:var(--muted);font-size:10px;font-weight:750;letter-spacing:.07em;text-transform:uppercase}
.activity-date,.activity-select{grid-column:span 3}
.activity-keyword{grid-column:span 5}
.activity-user{grid-column:span 3}
.activity-search:not(.activity-keyword):not(.activity-user){grid-column:span 2}
.activity-search>div{display:flex;height:44px;align-items:center;gap:9px;padding:0 13px;border:1px solid var(--line);border-radius:10px;background:var(--surface);transition:border-color .16s,box-shadow .16s}
.activity-search>div:focus-within{border-color:#0d9488;box-shadow:0 0 0 3px rgba(13,148,136,.12)}
.activity-search svg{flex:0 0 auto;color:var(--muted)}
.activity-search input{min-width:0;width:100%;height:100%;border:0;outline:0;color:var(--text);background:transparent;font-size:12px}
.activity-search input::placeholder{color:var(--muted)}
.activity-filter>button{grid-column:span 2;display:flex;height:44px;align-self:end;align-items:center;justify-content:center;gap:8px;border:0;border-radius:10px;color:white;background:#0d9488;font-size:12px;font-weight:750;cursor:pointer;transition:background .16s,transform .16s}
.activity-filter>button:hover:not(:disabled){background:#0f766e;transform:translateY(-1px)}
.activity-filter>button:disabled{cursor:wait;opacity:.6}
:deep(.activity-filter .ui-date-picker){display:flex!important;width:100%!important;height:44px!important;overflow:hidden;border:1px solid var(--line)!important;border-radius:10px!important;background:var(--surface)!important;box-shadow:none!important}
:deep(.activity-filter .ui-date-picker .p-datepicker-input){min-width:0!important;width:calc(100% - 43px)!important;height:42px!important;flex:1 1 auto!important;border:0!important;border-radius:0!important;padding:0 13px!important;color:var(--text)!important;background:transparent!important;font-size:12px!important;box-shadow:none!important}
:deep(.activity-filter .ui-date-picker .p-datepicker-dropdown){width:43px!important;height:42px!important;flex:0 0 43px!important;border:0!important;border-left:1px solid var(--line)!important;border-radius:0!important;color:var(--muted)!important;background:transparent!important}
:deep(.activity-filter .ui-select){width:100%!important;height:44px!important;border:1px solid var(--line)!important;border-radius:10px!important;color:var(--text)!important;background:var(--surface)!important;box-shadow:none!important}
:deep(.activity-filter .ui-select .p-select-label){display:flex!important;align-items:center!important;padding:0 13px!important;color:var(--text)!important;font-size:12px!important}
:deep(.activity-filter .ui-select .p-select-label.p-placeholder){color:var(--muted)!important}
:deep(.activity-filter .ui-select .p-select-dropdown){color:var(--muted)!important}
:deep(.activity-filter .ui-date-picker:hover),:deep(.activity-filter .ui-select:hover){border-color:#7897a9!important}
:deep(.activity-filter .ui-date-picker:focus-within),:deep(.activity-filter .ui-select.p-focus){border-color:#0d9488!important;box-shadow:0 0 0 3px rgba(13,148,136,.12)!important}
:deep(.activity-filter .p-select-overlay){z-index:40!important;margin-top:5px!important;overflow:hidden;border:1px solid var(--line)!important;border-radius:11px!important;color:var(--text)!important;background:var(--surface)!important;box-shadow:0 18px 44px rgba(15,23,42,.22)!important}
:deep(.activity-filter .p-select-list){padding:5px!important;background:transparent!important}
:deep(.activity-filter .p-select-option){border-radius:7px!important;padding:9px 10px!important;color:var(--text)!important;background:transparent!important;font-size:11px!important}
:deep(.activity-filter .p-select-option:not(.p-disabled).p-focus),:deep(.activity-filter .p-select-option:not(.p-disabled):hover){color:#0f766e!important;background:rgba(20,184,166,.11)!important}
:deep(.activity-filter .p-select-option.p-select-option-selected){color:white!important;background:#0d9488!important}
.activity-summary{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:12px;margin-top:16px}
.activity-summary article{display:flex;min-width:0;align-items:center;gap:12px;padding:13px 15px;border:1px solid var(--line);border-radius:14px;background:var(--surface-soft)}
.activity-summary svg{flex:0 0 auto;color:#0d9488}.activity-summary span{display:block;color:var(--muted);font-size:9px;font-weight:700;letter-spacing:.05em;text-transform:uppercase}.activity-summary strong{display:block;margin-top:3px;font-size:20px}
:deep(.activity-table.data-table-wrap){overflow:auto;margin-top:16px}
:deep(.activity-table.data-table-wrap .p-datatable-table){width:100%!important;min-width:900px!important;table-layout:fixed!important}
:deep(.activity-table.data-table-wrap .p-datatable-thead>tr>th){padding:13px 14px!important;color:white!important;background:#0f6367!important;text-align:left!important}
:deep(.activity-table.data-table-wrap .p-datatable-thead>tr>th:nth-child(1)),:deep(.activity-table.data-table-wrap .p-datatable-tbody>tr>td:nth-child(1)){width:240px!important;text-align:left!important}
:deep(.activity-table.data-table-wrap .p-datatable-thead>tr>th:nth-child(2)),:deep(.activity-table.data-table-wrap .p-datatable-tbody>tr>td:nth-child(2)){width:100px!important;text-align:left!important}
:deep(.activity-table.data-table-wrap .p-datatable-thead>tr>th:nth-child(3)),:deep(.activity-table.data-table-wrap .p-datatable-tbody>tr>td:nth-child(3)){width:30%!important;text-align:left!important}
:deep(.activity-table.data-table-wrap .p-datatable-thead>tr>th:nth-child(4)),:deep(.activity-table.data-table-wrap .p-datatable-tbody>tr>td:nth-child(4)){width:22%!important;text-align:left!important}
:deep(.activity-table.data-table-wrap .p-datatable-thead>tr>th:nth-child(5)),:deep(.activity-table.data-table-wrap .p-datatable-tbody>tr>td:nth-child(5)){width:145px!important;text-align:left!important}
:deep(.activity-table.data-table-wrap .p-datatable-thead>tr>th:nth-child(6)),:deep(.activity-table.data-table-wrap .p-datatable-tbody>tr>td:nth-child(6)){width:74px!important;text-align:center!important}
:deep(.activity-table.data-table-wrap .p-datatable-tbody>tr>td){height:72px!important;padding:10px 14px!important}
:deep(.activity-table td small){display:block;overflow:hidden;margin-top:5px;color:var(--muted);font-size:10px;text-overflow:ellipsis;white-space:nowrap}
:deep(.activity-table td strong){display:block;overflow:hidden;color:var(--text);font-size:11px;font-weight:750;text-overflow:ellipsis;white-space:nowrap}
.activity-user-name{display:block;overflow:hidden;margin-top:4px;color:var(--text);font-size:11px;font-weight:600;text-overflow:ellipsis;white-space:nowrap}
.activity-badge,.result-badge{display:inline-flex!important;width:max-content;padding:5px 8px;border-radius:999px;color:#0f766e!important;background:rgba(20,184,166,.13);font-size:9px!important;font-weight:750;line-height:1}
.aksi-hapus,.result-badge.gagal{color:#e11d48!important;background:rgba(244,63,94,.12)}.aksi-ubah{color:#b45309!important;background:rgba(245,158,11,.13)}.aksi-login,.aksi-logout{color:#2563eb!important;background:rgba(59,130,246,.12)}
.detail-button{display:inline-grid;width:34px;height:34px;place-items:center;border:1px solid var(--line);border-radius:9px;color:#0d9488;background:var(--surface-soft);cursor:pointer}.detail-button:hover{border-color:#0d9488;background:rgba(20,184,166,.1)}
.activity-detail dl{display:grid;grid-template-columns:1fr 1fr;gap:10px;margin:0 0 14px}.activity-detail dl>div,.activity-detail section{padding:12px;border:1px solid var(--line);border-radius:12px;background:var(--surface-soft)}.activity-detail dt{color:var(--muted);font-size:9px;text-transform:uppercase}.activity-detail dd{margin:5px 0 0;overflow-wrap:anywhere}.activity-detail section{margin-top:10px}.activity-detail h3{margin:0 0 9px;font-size:13px}.activity-detail pre{max-height:220px;overflow:auto;margin:0;padding:11px;border-radius:9px;color:var(--text);background:var(--surface);font-size:11px;white-space:pre-wrap}
:global(.theme-dark) .activity-module :deep(.activity-table.data-table-wrap .p-datatable-thead>tr>th){border-color:rgba(148,163,184,.14)!important;background:#0d4f55!important}
:global(.theme-dark) .activity-module :deep(.activity-filter .ui-date-picker),:global(.theme-dark) .activity-module :deep(.activity-filter .ui-select),:global(.theme-dark) .activity-module .activity-search>div{border-color:#405269!important;background:#1c283a!important}
:global(.theme-dark) .activity-module :deep(.activity-filter .p-select-overlay){border-color:#405269!important;background:#111c2c!important;box-shadow:0 22px 55px rgba(0,0,0,.42)!important}
:global(.theme-dark) .activity-module :deep(.activity-filter .p-select-option:not(.p-disabled).p-focus),:global(.theme-dark) .activity-module :deep(.activity-filter .p-select-option:not(.p-disabled):hover){color:#5eead4!important;background:#26364b!important}
:global(.sirava-dark .activity-dialog){border-color:rgba(148,163,184,.2)!important;color:#f8fafc!important;background:#0f172a!important}
:global(.activity-select-overlay){z-index:10050!important;overflow:hidden!important;border:1px solid #cbd5e1!important;border-radius:11px!important;color:#0f172a!important;background:#fff!important;box-shadow:0 20px 48px rgba(15,23,42,.25)!important}
:global(.activity-select-overlay .p-select-list){padding:5px!important;background:#fff!important}
:global(.activity-select-overlay .p-select-option){border-radius:7px!important;color:#0f172a!important;background:#fff!important;font-size:11px!important}
:global(.activity-select-overlay .p-select-option:not(.p-disabled).p-focus),:global(.activity-select-overlay .p-select-option:not(.p-disabled):hover){color:#0f766e!important;background:#e8f7f5!important}
:global(.activity-select-overlay .p-select-option.p-select-option-selected){color:#fff!important;background:#0d9488!important}
:global(body.sirava-dark .activity-select-overlay){border-color:#405269!important;color:#f8fafc!important;background:#111c2c!important;box-shadow:0 24px 58px rgba(0,0,0,.5)!important}
:global(body.sirava-dark .activity-select-overlay .p-select-list){background:#111c2c!important}
:global(body.sirava-dark .activity-select-overlay .p-select-option){color:#e2e8f0!important;background:#111c2c!important}
:global(body.sirava-dark .activity-select-overlay .p-select-option:not(.p-disabled).p-focus),:global(body.sirava-dark .activity-select-overlay .p-select-option:not(.p-disabled):hover){color:#5eead4!important;background:#26364b!important}
:global(body.sirava-dark .activity-select-overlay .p-select-option.p-select-option-selected){color:#fff!important;background:#0f766e!important}
@media(max-width:1100px){.activity-filter{grid-template-columns:repeat(2,minmax(0,1fr))}.activity-filter :where(.activity-date,.activity-select,.activity-keyword,.activity-user,.activity-search),.activity-filter>button{grid-column:auto}.activity-keyword{grid-column:span 2!important}.activity-summary{grid-template-columns:1fr 1fr 1fr}}
@media(max-width:720px){.activity-module{width:calc(100% - 20px);margin:10px;padding:15px}.activity-filter,.activity-summary,.activity-detail dl{grid-template-columns:1fr}.activity-filter :where(.activity-date,.activity-select,.activity-keyword,.activity-user,.activity-search),.activity-filter>button{grid-column:auto!important}.activity-header h1{font-size:24px}}
</style>
