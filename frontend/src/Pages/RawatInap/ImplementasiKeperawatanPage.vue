<script setup>
import { ChevronDown, ChevronUp, LoaderCircle, Pencil, Save, Trash2, X } from '@lucide/vue'
import PrimeColumn from 'primevue/column'
import { computed, nextTick, reactive, ref, watch } from 'vue'
import CariPetugas from '../../Components/Ui/CariPetugas.vue'
import DataTable from '../../Components/Ui/DataTable.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import TableSearch from '../../Components/Ui/TableSearch.vue'
import {
  hapusImplementasiKeperawatan,
  implementasiKeperawatanData,
  simpanImplementasiKeperawatan,
  ubahImplementasiKeperawatan,
} from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
})

const notifikasi = useNotifikasi()
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const error = ref('')
const records = ref([])
const petugas = ref({ nip: '', nama: '', jabatan: '' })
const selectedOfficer = ref({ nip: '', nama: '', jabatan: '' })
const canChooseOfficer = ref(false)
const billingLocked = ref(false)
const editingKey = ref(null)
const deleteTarget = ref(null)
const formVisible = ref(false)
const keyword = ref('')
const form = reactive(emptyForm())

const tableRows = computed(() => records.value.map((item) => ({ ...item, _key: keyFor(item) })))
const filteredRows = computed(() => {
  const query = keyword.value.trim().toLowerCase()
  if (!query) return tableRows.value
  return tableRows.value.filter((item) => [
    item.tanggal, item.jam, item.uraian, item.nip, item.nama_petugas, item.jabatan,
  ].join(' ').toLowerCase().includes(query))
})

function currentDate() {
  const now = new Date()
  const offset = now.getTimezoneOffset() * 60000
  return new Date(now.getTime() - offset).toISOString().slice(0, 10)
}

function currentTime() {
  return [new Date().getHours(), new Date().getMinutes(), new Date().getSeconds()]
    .map((value) => String(value).padStart(2, '0')).join(':')
}

function emptyForm() {
  return {
    no_rawat: props.patient?.no_rawat || '',
    tanggal: currentDate(),
    jam: currentTime(),
    uraian: '',
    nip: '',
  }
}

function resetForm() {
  Object.assign(form, emptyForm(), { no_rawat: props.patient?.no_rawat || '', nip: petugas.value.nip || '' })
  selectedOfficer.value = { ...petugas.value }
  editingKey.value = null
}

function keyFor(item) {
  return `${item.no_rawat}|${item.tanggal}|${item.jam}`
}

function recordKey(item) {
  return { no_rawat: item.no_rawat, tanggal: item.tanggal, jam: item.jam }
}

function formatDate(value) {
  if (!value) return '-'
  const [year, month, day] = value.split('-')
  return `${day}/${month}/${year}`
}

async function loadRecords() {
  if (!props.token || !props.patient?.no_rawat) return
  loading.value = true
  error.value = ''
  try {
    const data = await implementasiKeperawatanData(props.token, props.patient.no_rawat)
    records.value = data?.catatan || []
    petugas.value = data?.petugas || { nip: '', nama: '', jabatan: '' }
    canChooseOfficer.value = Boolean(data?.bisa_memilih_petugas)
    billingLocked.value = Boolean(data?.billing_terkunci)
    resetForm()
  } catch (err) {
    error.value = err.message || 'Data implementasi keperawatan tidak dapat dibaca.'
    notifikasi.gagal(error.value)
  } finally {
    loading.value = false
  }
}

function editRecord(item) {
  if (billingLocked.value || !item.bisa_diubah) {
    notifikasi.peringatan(billingLocked.value
      ? 'Kunjungan sudah masuk billing. Catatan hanya dapat dilihat.'
      : 'Catatan ini hanya dapat diedit oleh petugas yang membuatnya.')
    return
  }
  editingKey.value = recordKey(item)
  Object.assign(form, {
    no_rawat: item.no_rawat,
    tanggal: item.tanggal,
    jam: item.jam,
    uraian: item.uraian,
    nip: item.nip,
  })
  selectedOfficer.value = { nip: item.nip, nama: item.nama_petugas || item.nip, jabatan: item.jabatan || '' }
  formVisible.value = true
  nextTick(() => document.querySelector('.implementasi-page .cppt-form-card')?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
}

function payload() {
  return {
    kunci_lama: editingKey.value,
    no_rawat: String(props.patient?.no_rawat || '').trim(),
    tanggal: String(form.tanggal || '').trim(),
    jam: String(form.jam || '').trim(),
    uraian: String(form.uraian || '').trim(),
    nip: selectedOfficer.value?.nip || form.nip || '',
  }
}

async function saveRecord() {
  if (billingLocked.value) return notifikasi.peringatan('Kunjungan sudah masuk billing. Catatan tidak dapat disimpan.')
  if (!String(form.uraian || '').trim()) {
    notifikasi.peringatan('Uraian implementasi keperawatan wajib diisi.')
    return nextTick(() => document.querySelector('.implementasi-uraian textarea')?.focus())
  }
  saving.value = true
  try {
    const response = editingKey.value
      ? await ubahImplementasiKeperawatan(props.token, payload())
      : await simpanImplementasiKeperawatan(props.token, payload())
    notifikasi.sukses(response?.pesan || 'Implementasi keperawatan berhasil disimpan.')
    await loadRecords()
    formVisible.value = false
  } catch (err) {
    notifikasi.gagal(err.message || 'Implementasi keperawatan gagal disimpan.')
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    const response = await hapusImplementasiKeperawatan(props.token, recordKey(deleteTarget.value))
    notifikasi.sukses(response?.pesan || 'Catatan berhasil dihapus.')
    deleteTarget.value = null
    await loadRecords()
  } catch (err) {
    notifikasi.gagal(err.message || 'Catatan gagal dihapus.')
  } finally {
    deleting.value = false
  }
}

watch(() => props.patient?.no_rawat, () => {
  formVisible.value = false
  keyword.value = ''
  loadRecords()
}, { immediate: true })
</script>

<template>
  <section class="implementasi-page cppt-page">
    <article class="cppt-form-card">
      <header class="cppt-section-header">
        <div>
          <span>Catatan Keperawatan Rawat Inap</span>
          <h3>{{ editingKey ? 'Edit Implementasi Keperawatan' : 'Input Implementasi Keperawatan' }}</h3>
          <p>Petugas: {{ petugas.nama || '-' }}<small v-if="petugas.jabatan"> · {{ petugas.jabatan }}</small></p>
        </div>
        <div class="cppt-section-tools">
          <button v-if="editingKey" type="button" class="cppt-button secondary" @click="resetForm"><X :size="15" /> Batal Edit</button>
          <button type="button" class="cppt-button toggle icon-only" :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'" @click="formVisible = !formVisible">
            <ChevronUp v-if="formVisible" :size="16" /><ChevronDown v-else :size="16" />
          </button>
        </div>
      </header>

      <div v-if="billingLocked" class="billing-lock">Kunjungan sudah masuk billing atau dibatalkan. Data hanya dapat dilihat.</div>
      <form v-show="formVisible" class="cppt-form form-compact implementasi-form" @submit.prevent="saveRecord">
        <fieldset :disabled="saving || billingLocked">
          <div class="form-grid">
            <FormInput v-model="form.tanggal" label="Tanggal Perawatan" type="date" required />
            <FormInput v-model="form.jam" label="Jam Rawat" type="time" step="1" required />
            <CariPetugas v-model="selectedOfficer" class="petugas-field" sumber="implementasi_keperawatan" :token="token" :disabled="!canChooseOfficer" required />
          </div>
          <FormInput v-model="form.uraian" class="implementasi-uraian" label="Uraian Implementasi Keperawatan" jenis="textarea" :rows="6" maxlength="1000" required placeholder="Tuliskan tindakan dan implementasi keperawatan yang telah dilakukan..." />
          <footer class="cppt-form-actions">
            <button type="button" class="cppt-button secondary" @click="resetForm"><X :size="15" /> Batal / Reset</button>
            <button type="submit" class="cppt-button primary">
              <LoaderCircle v-if="saving" class="spin" :size="15" /><Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editingKey ? 'Simpan Perubahan' : 'Simpan Implementasi' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="cppt-history-card implementasi-history-card">
      <header class="cppt-section-header">
        <div><span>Riwayat Keperawatan</span><h3>Implementasi Keperawatan</h3><p>{{ filteredRows.length }} dari {{ records.length }} catatan ditampilkan</p></div>
        <TableSearch v-if="records.length" v-model="keyword" placeholder="Cari tanggal, petugas, atau uraian..." :total="records.length" :filtered="filteredRows.length" />
      </header>

      <div v-if="loading" class="cppt-state"><LoaderCircle class="spin" :size="25" /><strong>Menarik implementasi keperawatan...</strong></div>
      <div v-else-if="error" class="cppt-state error"><strong>{{ error }}</strong><button type="button" @click="loadRecords">Coba Lagi</button></div>
      <div v-else-if="!records.length" class="cppt-state"><strong>Belum ada implementasi keperawatan untuk kunjungan ini.</strong></div>
      <DataTable v-else :rows="filteredRows" data-key="_key" empty-message="Implementasi keperawatan tidak ditemukan.">
        <PrimeColumn header="TANGGAL / JAM" style="min-width:150px"><template #body="{ data }"><div class="cppt-table-main"><strong>{{ formatDate(data.tanggal) }}</strong><span>{{ data.jam }}</span></div></template></PrimeColumn>
        <PrimeColumn header="PETUGAS" style="min-width:240px"><template #body="{ data }"><div class="cppt-table-main"><strong>{{ data.nama_petugas || data.nip }}</strong><span>{{ data.jabatan || data.nip }}</span></div></template></PrimeColumn>
        <PrimeColumn header="URAIAN IMPLEMENTASI" style="min-width:520px"><template #body="{ data }"><p class="cppt-table-note implementasi-table-note">{{ data.uraian || '-' }}</p></template></PrimeColumn>
        <PrimeColumn header="AKSI" frozen align-frozen="right" style="min-width:140px">
          <template #body="{ data }">
            <div v-if="data.bisa_diubah && !billingLocked" class="cppt-table-actions">
              <button type="button" @click="editRecord(data)"><Pencil :size="14" /> Edit</button>
              <button type="button" class="danger" @click="deleteTarget = data"><Trash2 :size="14" /> Hapus</button>
            </div>
            <span v-else class="cppt-owner-note">{{ billingLocked ? 'Terkunci billing' : 'Bukan milik Anda' }}</span>
          </template>
        </PrimeColumn>
      </DataTable>
    </article>

    <div v-if="deleteTarget" class="cppt-confirm-backdrop" @click.self="deleteTarget = null">
      <section class="cppt-confirm-dialog" role="dialog" aria-modal="true">
        <h3>Hapus Implementasi Keperawatan?</h3>
        <p>Catatan tanggal {{ formatDate(deleteTarget.tanggal) }} pukul {{ deleteTarget.jam }} akan dihapus.</p>
        <div><button type="button" class="cppt-button secondary" :disabled="deleting" @click="deleteTarget = null">Batal</button><button type="button" class="cppt-button danger" :disabled="deleting" @click="confirmDelete"><Trash2 :size="15" /> {{ deleting ? 'Menghapus...' : 'Hapus' }}</button></div>
      </section>
    </div>
  </section>
</template>

<style scoped>
.implementasi-page { color:var(--text); }
.form-grid { display:grid; grid-template-columns:220px 180px minmax(300px,1fr); gap:12px; margin-bottom:14px; }
.implementasi-uraian { width:100%; }
.billing-lock { border-bottom:1px solid rgba(245,158,11,.24); padding:11px 20px; color:#b45309; background:rgba(245,158,11,.09); font-size:12px; }
.implementasi-table-note { width:min(680px,58vw); }
.implementasi-history-card :deep(.p-datatable-table) { min-width:1050px; }
.implementasi-page :deep(.staff-search-results) { z-index:1000; }
@media (max-width:850px) { .form-grid { grid-template-columns:1fr 1fr; }.petugas-field { grid-column:1/-1; }.implementasi-history-card :deep(.table-search) { width:100%; } }
@media (max-width:540px) { .form-grid { grid-template-columns:1fr; }.petugas-field { grid-column:auto; }.implementasi-table-note { width:300px; } }
</style>
