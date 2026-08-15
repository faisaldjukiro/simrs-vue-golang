<script setup>
import { ChevronDown, ChevronUp, LoaderCircle, Pencil, Save, Trash2, X } from '@lucide/vue'
import PrimeColumn from 'primevue/column'
import { computed, nextTick, reactive, ref, watch } from 'vue'
import DataTable from '../../Components/Ui/DataTable.vue'
import CariPetugas from '../../Components/Ui/CariPetugas.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import { cpptData, hapusCppt, simpanCppt, ubahCppt } from '../../lib/faisal/api'
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
const petugas = ref({ nip: '', nama: '', jabatan: '' })
const selectedOfficer = ref({ nip: '', nama: '', jabatan: '' })
const canChooseOfficer = ref(false)
const billingLocked = ref(false)
const defaultAwareness = ['Compos Mentis', 'Apatis', 'Somnolence', 'Sopor', 'Coma']
const awarenessOptions = ref(toSelectOptions(defaultAwareness))
const editingKey = ref(null)
const deleteTarget = ref(null)
const formVisible = ref(false)
const judulCatatan = computed(() => props.jenisRawat === 'ralan' ? 'Pemeriksaan & SOAP' : 'CPPT & SOAP')

const form = reactive(emptyForm())
const tableRecords = computed(() => records.value.map((item) => ({
  ...item,
  _key: keyFor(item),
})))

function toSelectOptions(options = []) {
  return options.map((option) => {
    if (typeof option === 'object' && option !== null) return option
    return { label: String(option), value: String(option) }
  })
}

function currentDate() {
  const date = new Date()
  const offset = date.getTimezoneOffset() * 60000
  return new Date(date.getTime() - offset).toISOString().slice(0, 10)
}

function currentTime() {
  const date = new Date()
  return [date.getHours(), date.getMinutes(), date.getSeconds()]
    .map((value) => String(value).padStart(2, '0'))
    .join(':')
}

function emptyForm() {
  return {
    jenis_rawat: props.jenisRawat,
    no_rawat: props.patient?.no_rawat || '',
    tgl_perawatan: currentDate(),
    jam_rawat: currentTime(),
    suhu_tubuh: '',
    tensi: '',
    nadi: '',
    respirasi: '',
    tinggi: '',
    berat: '',
    spo2: '',
    gcs: '',
    kesadaran: 'Compos Mentis',
    subjek: '',
    objek: '',
    alergi: '',
    lingkar_perut: '',
    asesmen: '',
    plan: '',
    instruksi: '',
    evaluasi: '',
    nip: '',
  }
}

function resetForm() {
  Object.assign(form, emptyForm(), {
    no_rawat: props.patient.no_rawat,
    nip: petugas.value.nip || '',
  })
  editingKey.value = null
  selectedOfficer.value = { ...petugas.value }
}

function keyFor(item) {
  return `${item.no_rawat}|${item.tgl_perawatan}|${item.jam_rawat}`
}

function recordKey(item) {
  return {
    jenis_rawat: item.jenis_rawat || props.jenisRawat,
    no_rawat: String(item?.no_rawat || props.patient?.no_rawat || '').trim(),
    tgl_perawatan: item.tgl_perawatan,
    jam_rawat: item.jam_rawat,
  }
}

async function loadRecords() {
  if (!props.token || !props.patient.no_rawat) return
  loading.value = true
  error.value = ''
  try {
    const data = await cpptData(props.token, props.patient.no_rawat, props.jenisRawat)
    records.value = data?.catatan || []
    petugas.value = data?.petugas || { nip: '', nama: '', jabatan: '' }
    canChooseOfficer.value = Boolean(data?.bisa_memilih_petugas)
    billingLocked.value = Boolean(data?.billing_terkunci)
    awarenessOptions.value = toSelectOptions(
      data?.pilihan_kesadaran?.length ? data.pilihan_kesadaran : defaultAwareness,
    )
    resetForm()
  } catch (err) {
    error.value = err.message
    notifikasi.gagal(err.message || 'Catatan CPPT/SOAP tidak dapat dibaca.')
  } finally {
    loading.value = false
  }
}

function editRecord(item) {
  if (billingLocked.value) {
    notifikasi.peringatan('Kunjungan sudah masuk billing. Catatan hanya dapat dilihat.')
    return
  }
  if (!item.bisa_diubah) {
    notifikasi.peringatan('Catatan ini hanya dapat diedit oleh petugas yang membuatnya.')
    return
  }
  editingKey.value = recordKey(item)
  formVisible.value = true
  selectedOfficer.value = {
    nip: item.nip || '',
    nama: item.nama_petugas || item.nip || '',
    jabatan: item.jabatan || '',
  }
  Object.assign(form, {
    ...emptyForm(),
    no_rawat: item.no_rawat,
    tgl_perawatan: item.tgl_perawatan,
    jam_rawat: item.jam_rawat,
    suhu_tubuh: item.suhu_tubuh,
    tensi: item.tensi,
    nadi: item.nadi,
    respirasi: item.respirasi,
    tinggi: item.tinggi,
    berat: item.berat,
    spo2: item.spo2,
    gcs: item.gcs,
    kesadaran: item.kesadaran,
    subjek: item.subjek,
    objek: item.objek,
    alergi: item.alergi,
    lingkar_perut: item.lingkar_perut,
    asesmen: item.asesmen,
    plan: item.plan,
    instruksi: item.instruksi,
    evaluasi: item.evaluasi,
    nip: item.nip,
    no_rawat: props.patient.no_rawat,
  })
  nextTick(() => {
    document.querySelector('.cppt-form-card')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  })
}

function hasClinicalContent() {
  return [
    form.suhu_tubuh, form.tensi, form.nadi, form.respirasi, form.tinggi,
    form.berat, form.spo2, form.gcs, form.subjek, form.objek, form.alergi,
    form.lingkar_perut, form.asesmen, form.plan, form.instruksi, form.evaluasi,
  ].some((value) => String(value || '').trim())
}

function payload() {
  return {
    kunci_lama: editingKey.value,
    jenis_rawat: props.jenisRawat,
    no_rawat: String(form.no_rawat || '').trim(),
    tgl_perawatan: String(form.tgl_perawatan || '').trim(),
    jam_rawat: String(form.jam_rawat || '').trim(),
    suhu_tubuh: String(form.suhu_tubuh || '').trim(),
    tensi: String(form.tensi || '').trim(),
    nadi: String(form.nadi || '').trim(),
    respirasi: String(form.respirasi || '').trim(),
    tinggi: String(form.tinggi || '').trim(),
    berat: String(form.berat || '').trim(),
    spo2: String(form.spo2 || '').trim(),
    gcs: String(form.gcs || '').trim(),
    kesadaran: String(form.kesadaran || '').trim(),
    subjek: String(form.subjek || '').trim(),
    objek: String(form.objek || '').trim(),
    alergi: String(form.alergi || '').trim(),
    lingkar_perut: String(form.lingkar_perut || '').trim(),
    asesmen: String(form.asesmen || '').trim(),
    plan: String(form.plan || '').trim(),
    instruksi: String(form.instruksi || '').trim(),
    evaluasi: String(form.evaluasi || '').trim(),
    nip: selectedOfficer.value?.nip || form.nip || '',
  }
}

async function saveRecord() {
  if (billingLocked.value) {
    notifikasi.peringatan('Kunjungan sudah masuk billing. Catatan tidak dapat disimpan.')
    return
  }
  if (!hasClinicalContent()) {
    notifikasi.peringatan('Isi minimal satu pemeriksaan atau catatan SOAP.')
    return
  }
  saving.value = true
  try {
    const response = editingKey.value
      ? await ubahCppt(props.token, payload())
      : await simpanCppt(props.token, payload())
    notifikasi.sukses(response?.pesan || (editingKey.value ? 'Catatan berhasil diperbarui.' : 'Catatan berhasil disimpan.'))
    await loadRecords()
    formVisible.value = false
  } catch (err) {
    notifikasi.gagal(err.message || 'Catatan CPPT/SOAP gagal disimpan.')
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  if (billingLocked.value) {
    deleteTarget.value = null
    notifikasi.peringatan('Kunjungan sudah masuk billing. Catatan tidak dapat dihapus.')
    return
  }
  deleting.value = true
  try {
    const response = await hapusCppt(props.token, recordKey(deleteTarget.value))
    notifikasi.sukses(response?.pesan || 'Catatan berhasil dihapus.')
    deleteTarget.value = null
    await loadRecords()
  } catch (err) {
    notifikasi.gagal(err.message || 'Catatan CPPT/SOAP gagal dihapus.')
  } finally {
    deleting.value = false
  }
}

function formatDate(value) {
  if (!value) return '-'
  const [year, month, day] = value.split('-')
  return `${day}/${month}/${year}`
}

watch([() => props.patient.no_rawat, () => props.jenisRawat], () => {
  formVisible.value = false
  loadRecords()
}, { immediate: true })
</script>

<template>
  <section class="cppt-page">
    <article class="cppt-form-card">
      <header class="cppt-section-header">
        <div>
          <span>Catatan Perkembangan Pasien Terintegrasi</span>
          <h3>{{ editingKey ? `Edit ${judulCatatan}` : `Input ${judulCatatan}` }}</h3>
          <p>Petugas: {{ petugas.nama || '-' }} <small v-if="petugas.jabatan">· {{ petugas.jabatan }}</small></p>
        </div>
        <div class="cppt-section-tools">
          <button v-if="editingKey" type="button" class="cppt-button secondary" @click="resetForm">
            <X :size="15" /> Batal Edit
          </button>
          <button
            type="button"
            class="cppt-button toggle icon-only"
            :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'"
            :aria-label="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'"
            @click="formVisible = !formVisible"
          >
            <ChevronUp v-if="formVisible" :size="15" />
            <ChevronDown v-else :size="15" />
          </button>
        </div>
      </header>

      <div v-if="billingLocked" class="handling-lock">Billing sudah terverifikasi atau kunjungan dibatalkan. Form {{ judulCatatan }} hanya dapat dilihat.</div>
      <form v-show="formVisible" class="cppt-form" @submit.prevent="saveRecord">
        <fieldset class="form-compact" :disabled="saving || billingLocked">
          <div class="cppt-time-grid">
            <FormInput v-model="form.tgl_perawatan" label="Tanggal Perawatan" type="date" required />
            <FormInput v-model="form.jam_rawat" label="Jam Rawat" type="time" step="1" required />
            <FormInput v-model="form.kesadaran" class="wide" label="Kesadaran" jenis="select" :options="awarenessOptions" required />
            <CariPetugas
              v-model="selectedOfficer"
              class="petugas"
              :token="token"
              :disabled="!canChooseOfficer"
              required
            />
          </div>

          <div class="cppt-vital-grid">
            <FormInput v-model="form.suhu_tubuh" label="Suhu (°C)" maxlength="5" placeholder="36.5" />
            <FormInput v-model="form.tensi" label="Tensi (mmHg)" maxlength="8" placeholder="120/80" />
            <FormInput v-model="form.nadi" label="Nadi (/mnt)" maxlength="3" inputmode="numeric" />
            <FormInput v-model="form.respirasi" label="Respirasi (/mnt)" maxlength="3" inputmode="numeric" />
            <FormInput v-model="form.spo2" label="SpO2 (%)" maxlength="3" inputmode="numeric" />
            <FormInput v-model="form.gcs" label="GCS" maxlength="10" />
            <FormInput v-model="form.tinggi" label="Tinggi (cm)" maxlength="5" inputmode="decimal" />
            <FormInput v-model="form.berat" label="Berat (kg)" maxlength="5" inputmode="decimal" />
            <FormInput v-model="form.alergi" class="allergy" label="Alergi" maxlength="50" placeholder="Tuliskan alergi pasien" />
            <FormInput v-if="jenisRawat === 'ralan'" v-model="form.lingkar_perut" label="Lingkar Perut (cm)" maxlength="5" inputmode="decimal" />
          </div>

          <div class="cppt-soap-grid">
            <FormInput v-model="form.subjek" label="S — Subjek" jenis="textarea" :rows="3" maxlength="2000" placeholder="Keluhan dan informasi subjektif pasien" />
            <FormInput v-model="form.objek" label="O — Objek" jenis="textarea" :rows="3" maxlength="2000" placeholder="Hasil pemeriksaan objektif" />
            <FormInput v-model="form.asesmen" label="A — Asesmen" jenis="textarea" :rows="3" maxlength="2000" placeholder="Penilaian klinis" />
            <FormInput v-model="form.plan" label="P — Plan" jenis="textarea" :rows="3" maxlength="2000" placeholder="Rencana tindak lanjut" />
            <FormInput v-model="form.instruksi" label="Instruksi" jenis="textarea" :rows="2" maxlength="2000" placeholder="Instruksi pelayanan" />
            <FormInput v-model="form.evaluasi" label="Evaluasi" jenis="textarea" :rows="2" maxlength="2000" placeholder="Evaluasi perkembangan pasien" />
          </div>

          <footer class="cppt-form-actions">
            <button type="button" class="cppt-button secondary" @click="resetForm"><X :size="15" /> Batal / Reset</button>
            <button type="submit" class="cppt-button primary">
              <LoaderCircle v-if="saving" class="spin" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editingKey ? 'Simpan Perubahan' : `Simpan ${judulCatatan}` }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="cppt-history-card">
      <header class="cppt-section-header">
        <div>
          <span>Riwayat Pasien</span>
          <h3>Catatan {{ judulCatatan }}</h3>
          <p>{{ records.length }} catatan ditemukan</p>
        </div>
      </header>

      <div v-if="loading" class="cppt-state"><LoaderCircle class="spin" :size="25" /><strong>Menarik catatan CPPT/SOAP...</strong></div>
      <div v-else-if="error" class="cppt-state error"><strong>{{ error }}</strong><button type="button" @click="loadRecords">Coba Lagi</button></div>
      <div v-else-if="records.length === 0" class="cppt-state"><strong>Belum ada catatan CPPT/SOAP untuk pasien ini.</strong></div>

      <DataTable v-else :rows="tableRecords" data-key="_key" empty-message="Belum ada catatan CPPT/SOAP.">
        <PrimeColumn header="TANGGAL / JAM" style="min-width:140px">
          <template #body="{ data }"><div class="cppt-table-main"><strong>{{ formatDate(data.tgl_perawatan) }}</strong><span>{{ data.jam_rawat }}</span></div></template>
        </PrimeColumn>
        <PrimeColumn header="PETUGAS" style="min-width:190px">
          <template #body="{ data }"><div class="cppt-table-main"><strong>{{ data.nama_petugas || data.nip }}</strong><span>{{ data.jabatan || data.nip }}</span></div></template>
        </PrimeColumn>
        <PrimeColumn header="TANDA VITAL" style="min-width:220px">
          <template #body="{ data }">
            <div class="cppt-table-vitals">
              <span>Kesadaran <b>{{ data.kesadaran || '-' }}</b></span><span>Suhu <b>{{ data.suhu_tubuh || '-' }}</b></span>
              <span>Tensi <b>{{ data.tensi || '-' }}</b></span><span>Nadi <b>{{ data.nadi || '-' }}</b></span>
              <span>RR <b>{{ data.respirasi || '-' }}</b></span><span>SpO2 <b>{{ data.spo2 || '-' }}</b></span>
              <span>GCS <b>{{ data.gcs || '-' }}</b></span><span>TB/BB <b>{{ data.tinggi || '-' }}/{{ data.berat || '-' }}</b></span>
              <span v-if="jenisRawat === 'ralan'">Lingkar Perut <b>{{ data.lingkar_perut || '-' }}</b></span>
            </div>
          </template>
        </PrimeColumn>
        <PrimeColumn header="SUBJEK" style="min-width:230px"><template #body="{ data }"><p class="cppt-table-note">{{ data.subjek || '-' }}</p></template></PrimeColumn>
        <PrimeColumn header="OBJEK" style="min-width:230px"><template #body="{ data }"><p class="cppt-table-note">{{ data.objek || '-' }}</p></template></PrimeColumn>
        <PrimeColumn header="ASESMEN" style="min-width:230px"><template #body="{ data }"><p class="cppt-table-note">{{ data.asesmen || '-' }}</p></template></PrimeColumn>
        <PrimeColumn header="PLAN" style="min-width:230px"><template #body="{ data }"><p class="cppt-table-note">{{ data.plan || '-' }}</p></template></PrimeColumn>
        <PrimeColumn header="INSTRUKSI / EVALUASI" style="min-width:250px">
          <template #body="{ data }"><div class="cppt-table-extra"><p><b>Instruksi:</b> {{ data.instruksi || '-' }}</p><p><b>Evaluasi:</b> {{ data.evaluasi || '-' }}</p><p><b>Alergi:</b> {{ data.alergi || '-' }}</p></div></template>
        </PrimeColumn>
        <PrimeColumn header="AKSI" frozen align-frozen="right" style="min-width:135px">
          <template #body="{ data }">
            <div v-if="data.bisa_diubah && !billingLocked" class="cppt-table-actions">
              <button type="button" title="Edit catatan" @click="editRecord(data)"><Pencil :size="14" /> Edit</button>
              <button type="button" class="danger" title="Hapus catatan" @click="deleteTarget = data"><Trash2 :size="14" /> Hapus</button>
            </div>
            <span v-else class="cppt-owner-note">{{ billingLocked ? 'Terkunci billing' : 'Bukan milik Anda' }}</span>
          </template>
        </PrimeColumn>
      </DataTable>
    </article>

    <div v-if="deleteTarget" class="cppt-confirm-backdrop" @click.self="deleteTarget = null">
      <section class="cppt-confirm-dialog" role="dialog" aria-modal="true">
        <h3>Hapus Catatan CPPT?</h3>
        <p>Catatan tanggal {{ formatDate(deleteTarget.tgl_perawatan) }} pukul {{ deleteTarget.jam_rawat }} akan dihapus dari SIMRS Khanza.</p>
        <div>
          <button type="button" class="cppt-button secondary" :disabled="deleting" @click="deleteTarget = null">Batal</button>
          <button type="button" class="cppt-button danger" :disabled="deleting" @click="confirmDelete">
            <LoaderCircle v-if="deleting" class="spin" :size="15" /><Trash2 v-else :size="15" />
            {{ deleting ? 'Menghapus...' : 'Hapus' }}
          </button>
        </div>
      </section>
    </div>
  </section>
</template>
