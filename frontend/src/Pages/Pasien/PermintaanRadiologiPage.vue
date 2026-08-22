<script setup>
import { ChevronDown, ChevronUp, LoaderCircle, Pencil, Save, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import { computed, nextTick, reactive, ref, watch } from 'vue'
import CariDokter from '../../Components/Ui/CariDokter.vue'
import CariTindakanRadiologi from '../../Components/Ui/CariTindakanRadiologi.vue'
import DataTable from '../../Components/Ui/DataTable.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import TableSearch from '../../Components/Ui/TableSearch.vue'
import {
  hapusPermintaanRadiologi,
  permintaanRadiologiData,
  simpanPermintaanRadiologi,
  ubahPermintaanRadiologi,
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
const formVisible = ref(true)
const requests = ref([])
const doctor = ref({})
const defaultDoctor = ref({})
const treatments = ref([])
const billingLocked = ref(false)
const scope = ref({})
const deleteTarget = ref(null)
const editingNumber = ref('')
const kataKunciPermintaan = ref('')
const form = reactive(emptyForm())
const tableRows = computed(() => requests.value.map((item) => ({ ...item, _key: item.nomor })))
const tableRowsTampil = computed(() => {
  const keyword = kataKunciPermintaan.value.trim().toLowerCase()
  if (!keyword) return tableRows.value
  return tableRows.value.filter((item) => teksPermintaan(item).includes(keyword))
})

function today() {
  const date = new Date()
  return new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 10)
}
function now() { return new Date().toTimeString().slice(0, 8) }
function emptyForm() {
  return {
    no_rawat: props.patient?.no_rawat || '',
    tanggal: today(),
    jam: now(),
    informasi_tambahan: '',
    diagnosis_klinis: '',
  }
}
function rupiah(value) {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(Number(value || 0))
}
function formatDate(value) {
  if (!value) return '-'
  const [year, month, day] = value.split('-')
  return `${day}/${month}/${year}`
}
function teksPermintaan(item) {
  const pemeriksaan = (item.pemeriksaan || []).flatMap((pemeriksaan) => [
    pemeriksaan.kode,
    pemeriksaan.nama,
    pemeriksaan.kelas,
    pemeriksaan.total,
  ])
  return [
    item.nomor,
    item.tanggal,
    item.jam,
    item.kode_dokter,
    item.nama_dokter,
    item.informasi_tambahan,
    item.diagnosis_klinis,
    item.status_pemeriksaan,
    item.status_bayar,
    item.total,
    ...pemeriksaan,
  ].join(' ').toLowerCase()
}
function waktuStatus(item) {
  if (item.status_pemeriksaan === 'Selesai') {
    return `${formatDate(item.tanggal_hasil)} · ${item.jam_hasil || '-'}`
  }
  if (item.status_pemeriksaan === 'Sedang Dikerjakan') {
    return `${formatDate(item.tanggal_diterima)} · ${item.jam_diterima || '-'}`
  }
  return 'Belum diterima petugas'
}
function kelasStatusPemeriksaan(status) {
  if (status === 'Selesai') return 'completed'
  if (status === 'Sedang Dikerjakan') return 'processing'
  return 'waiting'
}
function kelasStatusBayar(status) {
  if (status === 'Sudah Bayar') return 'paid'
  if (status === 'Sebagian Dibayar') return 'partial'
  return 'unpaid'
}
function resetForm() {
  Object.assign(form, emptyForm())
  doctor.value = { ...defaultDoctor.value }
  treatments.value = []
  editingNumber.value = ''
}

function editRequest(item) {
  if (billingLocked.value || !item.dapat_diubah) return
  editingNumber.value = item.nomor
  Object.assign(form, {
    no_rawat: item.no_rawat,
    tanggal: item.tanggal,
    jam: item.jam,
    informasi_tambahan: item.informasi_tambahan,
    diagnosis_klinis: item.diagnosis_klinis,
  })
  doctor.value = { kode: item.kode_dokter, nama: item.nama_dokter }
  treatments.value = (item.pemeriksaan || []).map((tindakan) => ({
    kode: tindakan.kode,
    nama: tindakan.nama,
    kode_cara_bayar: tindakan.kode_cara_bayar,
    kelas: tindakan.kelas,
    total: tindakan.total,
  }))
  formVisible.value = true
  nextTick(() => document.querySelector('.radiology-form-card')?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
}

async function loadData() {
  if (!props.patient.no_rawat) return
  loading.value = true
  error.value = ''
  try {
    const data = await permintaanRadiologiData(props.token, props.patient.no_rawat)
    requests.value = data?.permintaan || []
    billingLocked.value = Boolean(data?.billing_terkunci)
    defaultDoctor.value = data?.dokter_perujuk || {}
    scope.value = {
      status: data?.status_rawat || '',
      kodeCaraBayar: data?.kode_cara_bayar || '',
      kelas: data?.kelas_pasien || '',
      filterCaraBayar: Boolean(data?.filter_cara_bayar),
      filterKelas: Boolean(data?.filter_kelas),
    }
    if (!doctor.value?.kode) doctor.value = { ...defaultDoctor.value }
  } catch (err) {
    error.value = err.message
    notifikasi.gagal(err.message)
  } finally {
    loading.value = false
  }
}

async function saveRequest() {
  if (!doctor.value?.kode || treatments.value.length === 0) {
    notifikasi.peringatan('Dokter perujuk dan minimal satu tindakan radiologi wajib dipilih.')
    return
  }
  saving.value = true
  try {
    const payload = {
      ...form,
      kode_dokter: doctor.value.kode,
      kode_tindakan: treatments.value.map((item) => item.kode),
    }
    const response = editingNumber.value
      ? await ubahPermintaanRadiologi(props.token, editingNumber.value, payload)
      : await simpanPermintaanRadiologi(props.token, payload)
    notifikasi.sukses(`${response?.pesan || 'Permintaan radiologi berhasil disimpan.'} Nomor: ${response?.nomor || '-'}`)
    resetForm()
    await loadData()
    formVisible.value = false
  } catch (err) {
    notifikasi.gagal(err.message)
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    const response = await hapusPermintaanRadiologi(props.token, props.patient.no_rawat, deleteTarget.value.nomor)
    notifikasi.sukses(response?.pesan || 'Permintaan radiologi berhasil dihapus.')
    deleteTarget.value = null
    await loadData()
  } catch (err) {
    notifikasi.gagal(err.message)
  } finally {
    deleting.value = false
  }
}

watch(() => props.patient.no_rawat, () => {
  kataKunciPermintaan.value = ''
  resetForm()
  loadData()
}, { immediate: true })
</script>

<template>
  <section class="cppt-page radiology-request-page">
    <article class="cppt-form-card radiology-form-card">
      <header class="cppt-section-header">
        <div>
          <span>Penunjang Medis</span>
          <h3>{{ editingNumber ? 'Edit Permintaan Radiologi' : 'Input Permintaan Radiologi' }}</h3>
          <p>Tindakan dapat dipilih lebih dari satu dalam satu nomor permintaan.</p>
        </div>
        <div class="cppt-section-tools">
          <button v-if="editingNumber" type="button" class="cppt-button secondary" @click="resetForm"><X :size="15" /> Batal Edit</button>
          <button type="button" class="cppt-button toggle icon-only" :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'" @click="formVisible = !formVisible">
            <ChevronUp v-if="formVisible" :size="15" />
            <ChevronDown v-else :size="15" />
          </button>
        </div>
      </header>

      <div v-if="billingLocked" class="handling-lock">Billing sudah terverifikasi atau kunjungan dibatalkan. Permintaan baru tidak dapat dibuat.</div>

      <form v-show="formVisible" class="cppt-form radiology-form" @submit.prevent="saveRequest">
        <fieldset class="form-compact" :disabled="saving || billingLocked">
          <div class="radiology-form-main">
            <FormInput v-model="form.tanggal" label="Tanggal Permintaan" type="date" required />
            <FormInput v-model="form.jam" label="Jam Permintaan" type="time" step="1" required />
            <CariDokter v-model="doctor" class="radiology-referrer" :token="token" sumber="radiologi" required />
          </div>

          <div class="radiology-notes-grid">
            <FormInput v-model="form.informasi_tambahan" label="Informasi Tambahan" jenis="textarea" :rows="2" maxlength="60" placeholder="Informasi tambahan untuk petugas radiologi" required />
            <FormInput v-model="form.diagnosis_klinis" label="Diagnosis Klinis" jenis="textarea" :rows="2" maxlength="80" placeholder="Diagnosis atau alasan klinis pemeriksaan" required />
          </div>

          <CariTindakanRadiologi v-model="treatments" :token="token" :no-rawat="patient.no_rawat" required />

          <div class="radiology-filter-note">
            <span>{{ scope.status === 'ranap' ? 'Rawat Inap' : 'Rawat Jalan' }}</span>
            <span v-if="scope.filterCaraBayar">Cara bayar: {{ scope.kodeCaraBayar || '-' }}</span>
            <span v-if="scope.filterKelas">Kelas: {{ scope.kelas || '-' }}</span>
          </div>

          <footer class="cppt-form-actions">
            <button type="button" class="cppt-button secondary" @click="resetForm"><X :size="15" /> Batal / Reset</button>
            <button type="submit" class="cppt-button primary">
              <LoaderCircle v-if="saving" class="spin" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editingNumber ? 'Simpan Perubahan' : `Simpan ${treatments.length || ''} Tindakan` }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="cppt-history-card radiology-history-card">
      <header class="cppt-section-header">
        <div>
          <span>Riwayat Pasien</span>
          <h3>Permintaan Radiologi</h3>
          <p>
            <template v-if="kataKunciPermintaan">{{ tableRowsTampil.length }} dari {{ requests.length }} permintaan ditampilkan</template>
            <template v-else>{{ requests.length }} permintaan ditemukan</template>
          </p>
        </div>
        <TableSearch
          v-if="requests.length > 0"
          v-model="kataKunciPermintaan"
          placeholder="Cari permintaan, dokter, atau tindakan..."
          :total="requests.length"
          :filtered="tableRowsTampil.length"
        />
      </header>
      <div v-if="loading" class="cppt-state"><LoaderCircle class="spin" :size="25" /><strong>Menarik permintaan radiologi...</strong></div>
      <div v-else-if="error" class="cppt-state error"><strong>{{ error }}</strong><button type="button" @click="loadData">Coba Lagi</button></div>
      <div v-else-if="requests.length === 0" class="cppt-state"><strong>Belum ada permintaan radiologi.</strong></div>
      <DataTable v-else :rows="tableRowsTampil" data-key="_key" empty-message="Permintaan radiologi tidak ditemukan.">
        <Column header="NO. PERMINTAAN" style="min-width:175px"><template #body="{ data }"><div class="cppt-table-main"><strong>{{ data.nomor }}</strong><span>{{ formatDate(data.tanggal) }} · {{ data.jam }}</span></div></template></Column>
        <Column header="DOKTER PERUJUK" style="min-width:230px"><template #body="{ data }"><div class="cppt-table-main"><strong>{{ data.nama_dokter }}</strong><span>{{ data.kode_dokter }}</span></div></template></Column>
        <Column header="INFORMASI / DIAGNOSIS" style="min-width:300px"><template #body="{ data }"><div class="radiology-clinical"><p><b>Informasi:</b> {{ data.informasi_tambahan }}</p><p><b>Diagnosis:</b> {{ data.diagnosis_klinis }}</p></div></template></Column>
        <Column header="TINDAKAN RADIOLOGI" style="min-width:360px"><template #body="{ data }"><div class="radiology-exam-list"><article v-for="item in data.pemeriksaan" :key="item.kode"><span><strong>{{ item.nama }}</strong><small>{{ item.kode }} · {{ item.kelas }}</small></span><b>{{ rupiah(item.total) }}</b></article></div></template></Column>
        <Column header="TOTAL" style="min-width:130px"><template #body="{ data }"><strong>{{ rupiah(data.total) }}</strong></template></Column>
        <Column header="STATUS PEMERIKSAAN" style="min-width:180px"><template #body="{ data }"><div class="radiology-status-cell"><span class="radiology-status-badge" :class="kelasStatusPemeriksaan(data.status_pemeriksaan)">{{ data.status_pemeriksaan }}</span><small>{{ waktuStatus(data) }}</small></div></template></Column>
        <Column header="STATUS BAYAR" style="min-width:145px"><template #body="{ data }"><span class="radiology-status-badge" :class="kelasStatusBayar(data.status_bayar)">{{ data.status_bayar }}</span></template></Column>
        <Column header="AKSI" frozen align-frozen="right" style="min-width:145px">
          <template #body="{ data }">
            <div v-if="!billingLocked && (data.dapat_diubah || data.dapat_dihapus)" class="cppt-table-actions">
              <button v-if="data.dapat_diubah" type="button" @click="editRequest(data)"><Pencil :size="14" /> Edit</button>
              <button v-if="data.dapat_dihapus" type="button" class="danger" @click="deleteTarget = data"><Trash2 :size="14" /> Hapus</button>
            </div>
            <span v-else class="radiology-locked" :title="billingLocked || data.status_bayar !== 'Belum Bayar' ? 'Sudah masuk billing/pembayaran' : 'Sudah diterima petugas radiologi'">Terkunci</span>
          </template>
        </Column>
      </DataTable>
    </article>

    <div v-if="deleteTarget" class="cppt-confirm-backdrop" @click.self="deleteTarget = null">
      <section class="cppt-confirm-dialog">
        <h3>Hapus Permintaan Radiologi?</h3>
        <p>Permintaan {{ deleteTarget.nomor }} beserta {{ deleteTarget.pemeriksaan.length }} tindakan akan dihapus.</p>
        <div><button type="button" class="cppt-button secondary" @click="deleteTarget = null">Batal</button><button type="button" class="cppt-button danger" :disabled="deleting" @click="confirmDelete"><LoaderCircle v-if="deleting" class="spin" :size="15" /><Trash2 v-else :size="15" />{{ deleting ? 'Menghapus...' : 'Hapus' }}</button></div>
      </section>
    </div>
  </section>
</template>
