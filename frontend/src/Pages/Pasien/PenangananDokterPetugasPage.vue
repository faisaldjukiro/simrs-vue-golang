<script setup>
import { ChevronDown, ChevronUp, LoaderCircle, Pencil, Save, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import { computed, nextTick, reactive, ref, watch } from 'vue'
import CariDokter from '../../Components/Ui/CariDokter.vue'
import CariPetugas from '../../Components/Ui/CariPetugas.vue'
import CariTindakanRanap from '../../Components/Ui/CariTindakanRanap.vue'
import DataTable from '../../Components/Ui/DataTable.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import {
  hapusPenangananDokterPetugas, penangananDokterPetugasData,
  simpanBanyakPenangananDokterPetugas, ubahPenangananDokterPetugas,
} from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
  jenisRawat: { type: String, default: 'ranap' },
  namaModul: { type: String, default: 'Rawat Inap' },
})
const notifikasi = useNotifikasi()
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const error = ref('')
const records = ref([])
const billingLocked = ref(false)
const formVisible = ref(true)
const editingKey = ref(null)
const deleteTarget = ref(null)
const doctor = ref({})
const officer = ref({})
const treatments = ref([])
const form = reactive(emptyForm())
const labelJenisRawat = computed(() => props.jenisRawat === 'ralan' ? props.namaModul : 'Rawat Inap')
const tableRecords = computed(() => records.value.map((item) => ({
  ...item,
  _key: [item.no_rawat, item.kode_tindakan, item.kode_dokter, item.kode_petugas, item.tanggal, item.jam].join('|'),
})))

function today() { const d = new Date(); return new Date(d.getTime() - d.getTimezoneOffset() * 60000).toISOString().slice(0, 10) }
function now() { return new Date().toTimeString().slice(0, 8) }
function emptyForm() { return { no_rawat: props.patient?.no_rawat || '', tanggal: today(), jam: now() } }
function keyFor(item) { return { jenis_rawat: item.jenis_rawat || props.jenisRawat, no_rawat: item.no_rawat, kode_tindakan: item.kode_tindakan, kode_dokter: item.kode_dokter, kode_petugas: item.kode_petugas, tanggal: item.tanggal, jam: item.jam } }
function rupiah(value) { return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(Number(value || 0)) }
function formatDate(value) { if (!value) return '-'; const [y, m, d] = value.split('-'); return `${d}/${m}/${y}` }

function resetForm(defaults = true) {
  Object.assign(form, emptyForm())
  editingKey.value = null
  treatments.value = []
  if (!defaults) { doctor.value = {}; officer.value = {} }
}

async function loadRecords() {
  if (!props.patient.no_rawat) return
  loading.value = true; error.value = ''
  try {
    const data = await penangananDokterPetugasData(props.token, props.patient.no_rawat, props.jenisRawat)
    records.value = data?.catatan || []
    billingLocked.value = Boolean(data?.billing_terkunci)
    if (!editingKey.value) {
      doctor.value = data?.dokter_dpjp || {}
      officer.value = data?.petugas_login || {}
    }
  } catch (err) { error.value = err.message; notifikasi.gagal(err.message) }
  finally { loading.value = false }
}

function editRecord(item) {
  if (billingLocked.value) return
  editingKey.value = keyFor(item)
  Object.assign(form, { no_rawat: item.no_rawat, tanggal: item.tanggal, jam: item.jam })
  doctor.value = { kode: item.kode_dokter, nama: item.nama_dokter }
  officer.value = { kode: item.kode_petugas, nama: item.nama_petugas }
  treatments.value = [{ kode: item.kode_tindakan, nama: item.nama_tindakan, kelas: item.kelas, material: item.material, bhp: item.bhp, tarif_dokter: item.tarif_dokter, tarif_petugas: item.tarif_petugas, kso: item.kso, manajemen: item.manajemen, total: item.total }]
  formVisible.value = true
  nextTick(() => document.querySelector('.handling-form-card')?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
}

function catatan(tindakan) { return { jenis_rawat: props.jenisRawat, no_rawat: form.no_rawat, kode_tindakan: tindakan.kode || '', kode_dokter: doctor.value.kode || '', kode_petugas: officer.value.kode || '', tanggal: form.tanggal, jam: form.jam } }
function payload() { return { kunci_lama: editingKey.value, ...catatan(treatments.value[0] || {}) } }
async function saveRecord() {
  if (!doctor.value.kode || !officer.value.kode || treatments.value.length === 0) { notifikasi.peringatan('Dokter, petugas, dan minimal satu tindakan wajib dipilih.'); return }
  saving.value = true
  try {
    const response = editingKey.value
      ? await ubahPenangananDokterPetugas(props.token, payload())
      : await simpanBanyakPenangananDokterPetugas(props.token, { catatan: treatments.value.map(catatan) })
    notifikasi.sukses(response?.pesan || 'Penanganan berhasil disimpan.')
    resetForm(false); await loadRecords(); formVisible.value = false
  } catch (err) { notifikasi.gagal(err.message) }
  finally { saving.value = false }
}
async function confirmDelete() {
  deleting.value = true
  try { const response = await hapusPenangananDokterPetugas(props.token, keyFor(deleteTarget.value)); notifikasi.sukses(response?.pesan || 'Data berhasil dihapus.'); deleteTarget.value = null; await loadRecords() }
  catch (err) { notifikasi.gagal(err.message) }
  finally { deleting.value = false }
}

watch([() => props.patient.no_rawat, () => props.jenisRawat], () => { resetForm(false); loadRecords() }, { immediate: true })
</script>

<template>
  <section class="cppt-page handling-page">
    <article class="cppt-form-card handling-form-card">
      <header class="cppt-section-header">
        <div><span>Tindakan {{ labelJenisRawat }}</span><h3>{{ editingKey ? 'Edit Penanganan Dokter & Petugas' : 'Input Penanganan Dokter & Petugas' }}</h3><p>Hanya tindakan gabungan dokter dan petugas.</p></div>
        <div class="cppt-section-tools">
          <button v-if="editingKey" type="button" class="cppt-button secondary" @click="resetForm(false)"><X :size="15" /> Batal Edit</button>
          <button type="button" class="cppt-button toggle icon-only" :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'" @click="formVisible = !formVisible"><ChevronUp v-if="formVisible" :size="15" /><ChevronDown v-else :size="15" /></button>
        </div>
      </header>
      <div v-if="billingLocked" class="handling-lock">Billing sudah terverifikasi atau kunjungan dibatalkan. Data hanya dapat dilihat.</div>
      <form v-show="formVisible" class="cppt-form handling-form" @submit.prevent="saveRecord">
        <fieldset class="form-compact" :disabled="saving || billingLocked">
          <div class="handling-main-grid">
            <FormInput v-model="form.tanggal" label="Tanggal Perawatan" type="date" required />
            <FormInput v-model="form.jam" label="Jam Rawat" type="time" step="1" required />
            <CariDokter v-model="doctor" :token="token" required />
            <CariPetugas v-model="officer" :token="token" sumber="penanganan" required />
            <CariTindakanRanap v-model="treatments" :token="token" :no-rawat="patient.no_rawat" :jenis-rawat="jenisRawat" :multiple="!editingKey" required />
          </div>
          <footer class="cppt-form-actions"><button type="button" class="cppt-button secondary" @click="resetForm(false)"><X :size="15" /> Batal / Reset</button><button type="submit" class="cppt-button primary"><LoaderCircle v-if="saving" class="spin" :size="15" /><Save v-else :size="15" />{{ saving ? 'Menyimpan...' : editingKey ? 'Simpan Perubahan' : `Simpan ${treatments.length || ''} Penanganan` }}</button></footer>
        </fieldset>
      </form>
    </article>

    <article class="cppt-history-card handling-history-card">
      <header class="cppt-section-header"><div><span>Riwayat Tindakan</span><h3>Penanganan Dokter & Petugas</h3><p>{{ records.length }} tindakan ditemukan</p></div></header>
      <div v-if="loading" class="cppt-state"><LoaderCircle class="spin" :size="25" /><strong>Menarik data penanganan...</strong></div>
      <div v-else-if="error" class="cppt-state error"><strong>{{ error }}</strong><button type="button" @click="loadRecords">Coba Lagi</button></div>
      <div v-else-if="records.length === 0" class="cppt-state"><strong>Belum ada penanganan dokter dan petugas.</strong></div>
      <DataTable v-else :rows="tableRecords" data-key="_key">
        <Column header="TANGGAL / JAM" style="min-width:130px"><template #body="{ data }"><div class="cppt-table-main"><strong>{{ formatDate(data.tanggal) }}</strong><span>{{ data.jam }}</span></div></template></Column>
        <Column header="TINDAKAN / TAGIHAN" style="min-width:260px"><template #body="{ data }"><div class="cppt-table-main"><strong>{{ data.nama_tindakan }}</strong><span>{{ data.kode_tindakan }}</span></div></template></Column>
        <Column header="DOKTER" style="min-width:220px"><template #body="{ data }"><div class="cppt-table-main"><strong>{{ data.nama_dokter }}</strong><span>{{ data.kode_dokter }}</span></div></template></Column>
        <Column header="PETUGAS" style="min-width:220px"><template #body="{ data }"><div class="cppt-table-main"><strong>{{ data.nama_petugas }}</strong><span>{{ data.kode_petugas }}</span></div></template></Column>
        <Column header="TOTAL" style="min-width:140px"><template #body="{ data }"><strong>{{ rupiah(data.total) }}</strong></template></Column>
        <Column header="AKSI" frozen align-frozen="right" style="min-width:135px"><template #body="{ data }"><div v-if="!billingLocked" class="cppt-table-actions"><button type="button" @click="editRecord(data)"><Pencil :size="14" /> Edit</button><button type="button" class="danger" @click="deleteTarget = data"><Trash2 :size="14" /> Hapus</button></div><span v-else>Terkunci</span></template></Column>
      </DataTable>
    </article>

    <div v-if="deleteTarget" class="cppt-confirm-backdrop" @click.self="deleteTarget = null"><section class="cppt-confirm-dialog"><h3>Hapus Penanganan?</h3><p>{{ deleteTarget.nama_tindakan }} tanggal {{ formatDate(deleteTarget.tanggal) }} pukul {{ deleteTarget.jam }} akan dihapus.</p><div><button type="button" class="cppt-button secondary" @click="deleteTarget = null">Batal</button><button type="button" class="cppt-button danger" :disabled="deleting" @click="confirmDelete"><LoaderCircle v-if="deleting" class="spin" :size="15" /><Trash2 v-else :size="15" />{{ deleting ? 'Menghapus...' : 'Hapus' }}</button></div></section></div>
  </section>
</template>
