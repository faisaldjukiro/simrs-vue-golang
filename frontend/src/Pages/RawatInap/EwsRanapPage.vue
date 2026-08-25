<script setup>
import { ChevronDown, ChevronUp, Eye, LoaderCircle, Pencil, Save, Trash2, X } from '@lucide/vue'
import PrimeColumn from 'primevue/column'
import { computed, nextTick, reactive, ref, watch } from 'vue'
import CariPetugas from '../../Components/Ui/CariPetugas.vue'
import DataTable from '../../Components/Ui/DataTable.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import TableSearch from '../../Components/Ui/TableSearch.vue'
import scoreNyeriImage from '../../assets/simrs/score-nyeri.png'
import { ewsRanapData, hapusEwsRanap, simpanEwsRanap, ubahEwsRanap } from '../../lib/faisal/api'
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
const formVisible = ref(true)
const kataKunciRiwayat = ref('')

const pilihanAlat = ref(toSelectOptions(['Ya', 'Tidak']))
const pilihanKesadaran = ref(toSelectOptions(['A', 'P-V-U']))
const pilihanSkalaNyeri = ref(toSelectOptions(['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '10']))
const form = reactive(emptyForm())
const validationErrors = reactive({})

const tableRecords = computed(() => records.value.map((item) => ({ ...item, _key: keyFor(item) })))
const tableRecordsTampil = computed(() => {
  const kata = kataKunciRiwayat.value.trim().toLowerCase()
  if (!kata) return tableRecords.value
  return tableRecords.value.filter((item) => teksRiwayat(item).includes(kata))
})

const totalMasuk = computed(() => toNumber(form.masuk1) + toNumber(form.masuk2))
const totalKeluar = computed(() => (
  toNumber(form.keluar1) + toNumber(form.keluar2) + toNumber(form.keluar3) + toNumber(form.keluar4) + toNumber(form.keluar5)
))
const balanceCairan = computed(() => totalMasuk.value - totalKeluar.value)

watch(
  () => [
    form.pernafasan, form.saturasi, form.alat, form.suhu, form.denyut, form.tekanan,
    form.kesadaran, form.masuk1, form.masuk2, form.keluar1, form.keluar2, form.keluar3,
    form.keluar4, form.keluar5,
  ],
  hitungOtomatis,
  { immediate: true },
)

watch(() => selectedOfficer.value, (value) => {
  form.nip = value?.nip || value?.kode || ''
  if (form.nip) delete validationErrors.nip
}, { deep: true })

watch(
  () => daftarFieldWajib().map((field) => field.nilai()),
  () => {
    for (const field of daftarFieldWajib()) {
      if (!fieldKosong(field.nilai())) delete validationErrors[field.nama]
    }
  },
)

watch(() => props.patient?.no_rawat, () => loadRecords(), { immediate: true })

function toSelectOptions(options = []) {
  return options.map((option) => ({ label: String(option), value: String(option) }))
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
    no_rawat: props.patient?.no_rawat || '',
    tanggal: currentDate(),
    jam: currentTime(),
    nip: '',
    pernafasan: '',
    score_pernafasan: '',
    saturasi: '',
    score_saturasi: '',
    alat: 'Tidak',
    score_alat: '0',
    suhu: '',
    score_suhu: '',
    denyut: '',
    score_denyut: '',
    tekanan: '',
    diastol: '',
    score_tekanan: '',
    kesadaran: 'A',
    score_kesadaran: '0',
    total_score: '0',
    klasifikasi: 'Sangat Rendah',
    respon: 'Dilakukan monitoring',
    tindakan: 'Melanjutkan monitoring',
    frekuensi: 'Minimal 12 Jam',
    skala_nyeri: '0',
    bb: '',
    tb: '',
    lk: '',
    lp: '',
    masuk1: '0',
    masuk2: '0',
    jumlahmasuk: '0',
    keluar1: '0',
    keluar2: '0',
    keluar3: '0',
    keluar4: '0',
    keluar5: '0',
    jumlahkeluar: '0',
    bc: '0',
  }
}

function resetForm() {
  Object.assign(form, emptyForm(), {
    no_rawat: props.patient.no_rawat,
    nip: petugas.value.nip || '',
  })
  editingKey.value = null
  selectedOfficer.value = { ...petugas.value }
  hapusSemuaErrorValidasi()
  hitungOtomatis()
}

function keyFor(item) {
  return `${item.no_rawat}|${item.tanggal}|${item.jam}`
}

function recordKey(item) {
  return {
    no_rawat: nilaiPayload(item.no_rawat || props.patient.no_rawat),
    tanggal: nilaiPayload(item.tanggal),
    jam: nilaiPayload(item.jam),
  }
}

function teksRiwayat(item) {
  return [
    item.tanggal, item.jam, item.nip, item.nama_petugas, item.jabatan, item.total_score,
    item.klasifikasi, item.respon, item.tindakan, item.frekuensi, item.kesadaran,
    item.pernafasan, item.saturasi, item.suhu, item.denyut, item.tekanan, item.diastol,
  ].join(' ').toLowerCase()
}

async function loadRecords() {
  if (!props.token || !props.patient?.no_rawat) return
  loading.value = true
  error.value = ''
  try {
    const data = await ewsRanapData(props.token, props.patient.no_rawat)
    records.value = data?.catatan || []
    petugas.value = data?.petugas || { nip: '', nama: '', jabatan: '' }
    canChooseOfficer.value = Boolean(data?.bisa_memilih_petugas)
    billingLocked.value = Boolean(data?.billing_terkunci)
    pilihanAlat.value = toSelectOptions(data?.pilihan_alat?.length ? data.pilihan_alat : ['Ya', 'Tidak'])
    pilihanKesadaran.value = toSelectOptions(data?.pilihan_kesadaran?.length ? data.pilihan_kesadaran : ['A', 'P-V-U'])
    pilihanSkalaNyeri.value = toSelectOptions(data?.pilihan_skala_nyeri?.length ? data.pilihan_skala_nyeri : ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '10'])
    resetForm()
    if (billingLocked.value && records.value.length) {
      fillForm(records.value[0], false)
      formVisible.value = true
    }
  } catch (err) {
    error.value = err.message
    notifikasi.gagal(err.message || 'EWS Ranap tidak dapat dibaca.')
  } finally {
    loading.value = false
  }
}

function editRecord(item) {
  if (billingLocked.value) {
    notifikasi.peringatan('Kunjungan sudah masuk billing. EWS Ranap hanya dapat dilihat.')
    return
  }
  if (!item.bisa_diubah) {
    notifikasi.peringatan('EWS ini hanya dapat diedit oleh petugas yang membuatnya.')
    return
  }
  editingKey.value = recordKey(item)
  formVisible.value = true
  fillForm(item, true)
  nextTick(() => document.querySelector('.ews-ranap-page .cppt-form-card')?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
}

function viewRecord(item) {
  editingKey.value = null
  formVisible.value = true
  fillForm(item, false)
  nextTick(() => document.querySelector('.ews-ranap-page .cppt-form-card')?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
}

function fillForm(item, sebagaiEdit = false) {
  if (sebagaiEdit) editingKey.value = recordKey(item)
  hapusSemuaErrorValidasi()
  selectedOfficer.value = {
    nip: item.nip || '',
    nama: item.nama_petugas || item.nip || '',
    jabatan: item.jabatan || '',
  }
  Object.assign(form, {
    ...emptyForm(),
    ...item,
    no_rawat: props.patient.no_rawat,
    nip: item.nip,
  })
  hitungOtomatis()
}

function payload() {
  hitungOtomatis()
  const body = {
    no_rawat: props.patient.no_rawat,
    nip: selectedOfficer.value?.nip || selectedOfficer.value?.kode || form.nip || petugas.value.nip,
    tanggal: form.tanggal,
    jam: form.jam,
    pernafasan: form.pernafasan,
    score_pernafasan: form.score_pernafasan,
    saturasi: form.saturasi,
    score_saturasi: form.score_saturasi,
    alat: form.alat,
    score_alat: form.score_alat,
    suhu: form.suhu,
    score_suhu: form.score_suhu,
    denyut: form.denyut,
    score_denyut: form.score_denyut,
    tekanan: form.tekanan,
    diastol: form.diastol,
    score_tekanan: form.score_tekanan,
    kesadaran: form.kesadaran,
    score_kesadaran: form.score_kesadaran,
    total_score: form.total_score,
    klasifikasi: form.klasifikasi,
    respon: form.respon,
    tindakan: form.tindakan,
    frekuensi: form.frekuensi,
    skala_nyeri: form.skala_nyeri,
    bb: form.bb,
    tb: form.tb,
    lk: form.lk,
    lp: form.lp,
    masuk1: form.masuk1,
    masuk2: form.masuk2,
    jumlahmasuk: form.jumlahmasuk,
    keluar1: form.keluar1,
    keluar2: form.keluar2,
    keluar3: form.keluar3,
    keluar4: form.keluar4,
    keluar5: form.keluar5,
    jumlahkeluar: form.jumlahkeluar,
    bc: form.bc,
  }
  return Object.fromEntries(Object.entries(body).map(([key, value]) => [key, nilaiPayload(value)]))
}

function nilaiPayload(value) {
  if (value === undefined || value === null) return ''
  return String(value).trim()
}

async function submitForm() {
  if (billingLocked.value) {
    notifikasi.peringatan('Kunjungan sudah masuk billing. EWS Ranap tidak dapat disimpan.')
    return
  }
  if (!validasiFormEws()) return
  saving.value = true
  try {
    const body = { ...payload() }
    if (editingKey.value) await ubahEwsRanap(props.token, { kunci_lama: editingKey.value, ...body })
    else await simpanEwsRanap(props.token, body)
    notifikasi.sukses(editingKey.value ? 'EWS Ranap berhasil diperbarui.' : 'EWS Ranap berhasil disimpan.')
    await loadRecords()
    formVisible.value = true
  } catch (err) {
    fokusDariPesanBackend(err.message)
    notifikasi.gagal(err.message || 'EWS Ranap tidak dapat disimpan.')
  } finally {
    saving.value = false
  }
}

function daftarFieldWajib() {
  return [
    { nama: 'tanggal', label: 'Tanggal', nilai: () => form.tanggal },
    { nama: 'jam', label: 'Jam', nilai: () => form.jam },
    { nama: 'nip', label: 'Petugas', nilai: () => selectedOfficer.value?.nip || selectedOfficer.value?.kode || form.nip },
    { nama: 'pernafasan', label: 'Pernafasan', nilai: () => form.pernafasan },
    { nama: 'saturasi', label: 'Saturasi O2', nilai: () => form.saturasi },
    { nama: 'alat', label: 'Alat Bantu O2', nilai: () => form.alat },
    { nama: 'suhu', label: 'Suhu', nilai: () => form.suhu },
    { nama: 'denyut', label: 'Denyut Jantung', nilai: () => form.denyut },
    { nama: 'tekanan', label: 'Sistolik', nilai: () => form.tekanan },
    { nama: 'diastol', label: 'Diastolik', nilai: () => form.diastol },
    { nama: 'kesadaran', label: 'Kesadaran', nilai: () => form.kesadaran },
    { nama: 'respon', label: 'Respon Klinis', nilai: () => form.respon },
    { nama: 'tindakan', label: 'Tindakan', nilai: () => form.tindakan },
    { nama: 'frekuensi', label: 'Frekuensi Monitoring', nilai: () => form.frekuensi },
  ]
}

function fieldKosong(value) {
  return value === undefined || value === null || String(value).trim() === ''
}

function hapusSemuaErrorValidasi() {
  Object.keys(validationErrors).forEach((key) => delete validationErrors[key])
}

function validasiFormEws() {
  hapusSemuaErrorValidasi()
  hitungOtomatis()
  const fieldKosongPertama = daftarFieldWajib().find((field) => fieldKosong(field.nilai()))
  if (!fieldKosongPertama) return true

  validationErrors[fieldKosongPertama.nama] = `${fieldKosongPertama.label} wajib diisi`
  notifikasi.peringatan(`${fieldKosongPertama.label} wajib diisi.`)
  fokusKeField(fieldKosongPertama.nama)
  return false
}

function fokusKeField(namaField) {
  const selector = namaField === 'nip'
    ? '.ews-ranap-page .ews-field-petugas input, .ews-ranap-page .ews-field-petugas button'
    : `#ews-${namaField}`

  formVisible.value = true
  nextTick(() => {
    const element = document.querySelector(selector)
    const target = element?.matches?.('input, textarea, select, button, [tabindex]')
      ? element
      : element?.querySelector?.('input, textarea, select, button, [tabindex]')
    target?.scrollIntoView({ behavior: 'smooth', block: 'center' })
    window.setTimeout(() => {
      target?.focus?.({ preventScroll: true })
      target?.select?.()
    }, 250)
  })
}

function fokusDariPesanBackend(pesan = '') {
  const teks = String(pesan).toLowerCase()
  const peta = [
    ['petugas', 'nip'],
    ['tanggal', 'tanggal'],
    ['jam', 'jam'],
    ['pernafasan', 'pernafasan'],
    ['saturasi', 'saturasi'],
    ['alat bantu', 'alat'],
    ['suhu', 'suhu'],
    ['denyut', 'denyut'],
    ['sistolik', 'tekanan'],
    ['diastolik', 'diastol'],
    ['tekanan', 'tekanan'],
    ['kesadaran', 'kesadaran'],
    ['respon', 'respon'],
    ['tindakan', 'tindakan'],
    ['frekuensi', 'frekuensi'],
  ]
  const cocok = peta.find(([kata]) => teks.includes(kata))
  if (!cocok) return
  const field = daftarFieldWajib().find((item) => item.nama === cocok[1])
  if (!field) return
  validationErrors[field.nama] = pesan
  fokusKeField(field.nama)
}

async function deleteRecord(item) {
  if (billingLocked.value) {
    notifikasi.peringatan('Kunjungan sudah masuk billing. EWS Ranap tidak dapat dihapus.')
    return
  }
  if (!item.bisa_diubah) {
    notifikasi.peringatan('EWS ini hanya dapat dihapus oleh petugas yang membuatnya.')
    return
  }
  deleting.value = true
  try {
    await hapusEwsRanap(props.token, recordKey(item))
    notifikasi.sukses('EWS Ranap berhasil dihapus.')
    deleteTarget.value = null
    await loadRecords()
  } catch (err) {
    notifikasi.gagal(err.message || 'EWS Ranap tidak dapat dihapus.')
  } finally {
    deleting.value = false
  }
}

function hitungOtomatis() {
  form.score_pernafasan = skorPernafasan(form.pernafasan)
  form.score_saturasi = skorSaturasi(form.saturasi)
  form.score_alat = form.alat === 'Ya' ? '2' : '0'
  form.score_suhu = skorSuhu(form.suhu)
  form.score_denyut = skorDenyut(form.denyut)
  form.score_tekanan = skorTekanan(form.tekanan)
  form.score_kesadaran = form.kesadaran === 'P-V-U' ? '3' : '0'
  const total = [
    form.score_pernafasan, form.score_saturasi, form.score_alat, form.score_suhu,
    form.score_denyut, form.score_tekanan, form.score_kesadaran,
  ].reduce((sum, nilai) => sum + toNumber(nilai), 0)
  form.total_score = String(total)
  form.klasifikasi = klasifikasi(total)
  form.respon = responKlinis(total)
  form.tindakan = tindakanKlinis(total)
  form.frekuensi = frekuensiMonitoring(total)
  form.jumlahmasuk = String(totalMasuk.value)
  form.jumlahkeluar = String(totalKeluar.value)
  form.bc = String(balanceCairan.value)
}

function toNumber(value) {
  const parsed = Number.parseFloat(String(value ?? '').replace(',', '.'))
  return Number.isFinite(parsed) ? parsed : 0
}

function skorPernafasan(value) {
  const n = toNumber(value)
  if (!n) return ''
  if (n <= 8) return '3'
  if (n <= 11) return '1'
  if (n <= 20) return '0'
  if (n <= 24) return '2'
  return '3'
}

function skorSaturasi(value) {
  const n = toNumber(value)
  if (!n) return ''
  if (n <= 91) return '3'
  if (n <= 93) return '2'
  if (n <= 95) return '1'
  return '0'
}

function skorSuhu(value) {
  const n = toNumber(value)
  if (!n) return ''
  if (n <= 35) return '3'
  if (n <= 36) return '1'
  if (n <= 38) return '0'
  if (n <= 39) return '1'
  return '2'
}

function skorDenyut(value) {
  const n = toNumber(value)
  if (!n) return ''
  if (n <= 40) return '3'
  if (n <= 50) return '1'
  if (n <= 90) return '0'
  if (n <= 110) return '1'
  if (n <= 130) return '2'
  return '3'
}

function skorTekanan(value) {
  const n = toNumber(value)
  if (!n) return ''
  if (n <= 90) return '3'
  if (n <= 100) return '2'
  if (n <= 110) return '1'
  if (n <= 219) return '0'
  return '3'
}

function klasifikasi(total) {
  if (total === 0) return 'Sangat Rendah'
  if (total <= 4) return 'Rendah'
  if (total <= 6) return 'Sedang'
  return 'Tinggi'
}

function responKlinis(total) {
  if (total === 0) return 'Dilakukan monitoring'
  if (total <= 4) return 'Harus segera dievaluasi oleh perawat terdaftar yang kompeten, harus memutuskan apakah perubahan frekuensi pemantauan klinis atau wajib eskalasi perawatan klinis'
  if (total <= 6) return 'Harus segera melakukan tinjauan mendesak oleh klinis yang terampil dengan kompetensi dalam penilaian penyakit akut di bangsal, biasanya oleh dokter atau perawat dengan mempertimbangkan apakah eskalasi perawatan ke tim perawatan kritis diperlukan'
  return 'Harus segera memberikan penilaian darurat secara klinis oleh tim critical care outreach atau code blue dengan kompetensi penanganan pasien kritis dan biasanya terjadi transfer pasien ke area perawatan dengan alat bantu'
}

function tindakanKlinis(total) {
  if (total === 0) return 'Melanjutkan monitoring'
  if (total <= 4) return 'Perawat melakukan assessment atau meningkatkan frekuensi monitor'
  if (total <= 6) return 'Perawat berkolaborasi dengan tim/pemberian assessment kegawatan/meningkatkan perawatan dengan fasilitas monitor yang lengkap'
  return 'Berkolaborasi dengan tim medis/pemberian assessment kegawatan/pindah ruang HCU/ICU'
}

function frekuensiMonitoring(total) {
  if (total === 0) return 'Minimal 12 Jam'
  if (total <= 4) return 'Minimal 4-6 Jam'
  if (total <= 6) return 'Minimal 1 Jam'
  return 'Bed side monitor/every time'
}

function scoreClass(total) {
  const n = toNumber(total)
  if (n >= 7) return 'danger'
  if (n >= 5) return 'warning'
  if (n >= 1) return 'info'
  return 'safe'
}
</script>

<template>
  <section class="cppt-page ews-ranap-page">
    <article class="cppt-form-card">
      <header class="cppt-section-header">
        <div>
          <span>Early Warning Score Ranap</span>
          <h3>{{ billingLocked && records.length ? 'Detail EWS Ranap' : editingKey ? 'Edit EWS Ranap' : 'Input EWS Ranap' }}</h3>
          <p>
            <template v-if="billingLocked">Kunjungan sudah masuk billing. Data hanya dapat dilihat.</template>
          </p>
        </div>
        <div class="cppt-section-tools">
          <button v-if="editingKey && !billingLocked" type="button" class="cppt-button secondary" @click="resetForm">
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

      <div v-if="billingLocked" class="handling-lock">Billing sudah terverifikasi atau kunjungan dibatalkan. Form EWS Ranap hanya dapat dilihat.</div>

      <form v-show="formVisible" class="cppt-form ews-form" novalidate @submit.prevent="submitForm">
        <fieldset class="form-compact" :disabled="saving || billingLocked">
          <div v-if="error" class="ews-alert">{{ error }}</div>

          <div class="cppt-time-grid ews-time-grid">
            <FormInput v-model="form.tanggal" id="ews-tanggal" label="Tanggal" type="date" :error="validationErrors.tanggal" required />
            <FormInput v-model="form.jam" id="ews-jam" label="Jam" type="time" step="1" :error="validationErrors.jam" required />
            <CariPetugas
              v-model="selectedOfficer"
              class="petugas ews-field-petugas"
              sumber="ews_ranap"
              :token="token"
              label="Petugas"
              required
              :disabled="billingLocked || !canChooseOfficer"
              :error="validationErrors.nip"
            />
          </div>

          <section class="ews-sheet">
            <header class="ews-subheader">
              <div>
                <span>Parameter EWS</span>
                <h4>Tanda Vital & Kesadaran</h4>
              </div>
              <div :class="['ews-score-chip', scoreClass(form.total_score)]">
                <strong>{{ form.total_score }}</strong>
                <span>{{ form.klasifikasi }}</span>
              </div>
            </header>

            <div class="cppt-vital-grid ews-vital-grid">
              <FormInput v-model="form.pernafasan" id="ews-pernafasan" label="Pernafasan (/mnt)" type="number" inputmode="numeric" :error="validationErrors.pernafasan" required />
              <FormInput v-model="form.saturasi" id="ews-saturasi" label="Saturasi O2 (%)" type="number" inputmode="numeric" :error="validationErrors.saturasi" required />
              <FormInput v-model="form.alat" id="ews-alat" label="Alat Bantu O2" jenis="select" :options="pilihanAlat" append-to="body" :error="validationErrors.alat" required />
              <FormInput v-model="form.suhu" id="ews-suhu" label="Suhu (°C)" type="number" step="0.1" :error="validationErrors.suhu" required />
              <FormInput v-model="form.denyut" id="ews-denyut" label="Denyut Jantung" type="number" inputmode="numeric" :error="validationErrors.denyut" required />
              <FormInput v-model="form.tekanan" id="ews-tekanan" label="Sistolik" type="number" inputmode="numeric" :error="validationErrors.tekanan" required />
              <FormInput v-model="form.diastol" id="ews-diastol" label="Diastolik" type="number" inputmode="numeric" :error="validationErrors.diastol" required />
              <FormInput v-model="form.kesadaran" id="ews-kesadaran" label="Kesadaran" jenis="select" :options="pilihanKesadaran" append-to="body" :error="validationErrors.kesadaran" required />
            </div>

            <div class="ews-score-list">
              <span>RR <b>{{ form.score_pernafasan || '-' }}</b></span>
              <span>SpO2 <b>{{ form.score_saturasi || '-' }}</b></span>
              <span>O2 <b>{{ form.score_alat || '-' }}</b></span>
              <span>Suhu <b>{{ form.score_suhu || '-' }}</b></span>
              <span>Denyut <b>{{ form.score_denyut || '-' }}</b></span>
              <span>TD <b>{{ form.score_tekanan || '-' }}</b></span>
              <span>Sadar <b>{{ form.score_kesadaran || '-' }}</b></span>
            </div>
          </section>

          <div class="cppt-soap-grid ews-soap-grid">
            <FormInput v-model="form.respon" id="ews-respon" label="Respon Klinis" jenis="textarea" :rows="3" :error="validationErrors.respon" required />
            <FormInput v-model="form.tindakan" id="ews-tindakan" label="Tindakan" jenis="textarea" :rows="3" :error="validationErrors.tindakan" required />
            <FormInput v-model="form.frekuensi" id="ews-frekuensi" label="Frekuensi Monitoring" :error="validationErrors.frekuensi" required />
          </div>

          <section class="ews-sheet ews-pain-sheet">
            <header class="ews-subheader">
              <div>
                <span>Score Nyeri</span>
                <h4>Skala Nyeri & Antropometri</h4>
              </div>
            </header>

            <div class="ews-pain-grid">
              <figure class="ews-pain-image">
                <img :src="scoreNyeriImage" alt="Wong Baker Faces Pain Rating Scale" />
              </figure>

              <div class="ews-pain-fields">
                <FormInput
                  v-model="form.skala_nyeri"
                  class="pain-scale"
                  label="Skala Nyeri"
                  jenis="select"
                  :options="pilihanSkalaNyeri"
                  append-to="body"
                />
                <FormInput v-model="form.bb" label="BB (Kg)" type="number" step="0.1" />
                <FormInput v-model="form.tb" label="TB (Cm)" type="number" step="0.1" />
                <FormInput v-model="form.lk" label="Lingkar Kepala" type="number" step="0.1" />
                <FormInput v-model="form.lp" label="Lingkar Perut" type="number" step="0.1" />
              </div>
            </div>
          </section>

          <section class="ews-sheet">
            <header class="ews-subheader">
              <div>
                <span>Data Tambahan</span>
                <h4>Balance Cairan</h4>
              </div>
              <div class="ews-balance">
                <span>Balance Cairan</span>
                <strong>{{ form.bc || 0 }}</strong>
              </div>
            </header>

            <div class="cppt-vital-grid ews-fluid-grid">
              <FormInput v-model="form.masuk1" label="Masuk Peroral / NGT" type="number" />
              <FormInput v-model="form.masuk2" label="Masuk Parenteral / Transfusi" type="number" />
              <FormInput v-model="form.jumlahmasuk" label="Jumlah Masuk" disabled />
              <FormInput v-model="form.keluar1" label="Keluar Feses" type="number" />
              <FormInput v-model="form.keluar2" label="Keluar Urine" type="number" />
              <FormInput v-model="form.keluar3" label="Keluar Muntah / NGT" type="number" />
              <FormInput v-model="form.keluar4" label="Keluar Drain / Darah" type="number" />
              <FormInput v-model="form.keluar5" label="Keluar IWL" type="number" />
              <FormInput v-model="form.jumlahkeluar" label="Jumlah Keluar" disabled />
              <FormInput v-model="form.bc" label="Balance Cairan" disabled />
            </div>
          </section>

          <footer v-if="!billingLocked" class="cppt-form-actions">
            <button type="button" class="cppt-button secondary" :disabled="saving" @click="resetForm">
              <X :size="15" /> Batal / Reset
            </button>
            <button type="submit" class="cppt-button primary" :disabled="saving">
              <LoaderCircle v-if="saving" :size="15" class="spin" />
              <Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editingKey ? 'Simpan Perubahan' : 'Simpan EWS' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="cppt-history-card">
      <header class="cppt-section-header">
        <div>
          <span>Riwayat Monitoring</span>
          <h3>EWS Ranap</h3>
          <p>
            <template v-if="kataKunciRiwayat">{{ tableRecordsTampil.length }} dari {{ records.length }} catatan ditampilkan</template>
            <template v-else>{{ records.length }} catatan ditemukan</template>
          </p>
        </div>
        <TableSearch
          v-if="records.length > 0"
          v-model="kataKunciRiwayat"
          placeholder="Cari tanggal, petugas, skor, atau klasifikasi..."
          :total="records.length"
          :filtered="tableRecordsTampil.length"
        />
      </header>

      <div v-if="loading" class="cppt-state">
        <LoaderCircle class="spin" :size="25" />
        <strong>Menarik data EWS Ranap...</strong>
      </div>
      <div v-else-if="error" class="cppt-state error">
        <strong>{{ error }}</strong>
        <button type="button" @click="loadRecords">Coba Lagi</button>
      </div>
      <div v-else-if="records.length === 0" class="cppt-state">
        <strong>Belum ada EWS Ranap pada kunjungan ini.</strong>
      </div>

      <DataTable v-else :rows="tableRecordsTampil" data-key="_key" empty-message="EWS Ranap tidak ditemukan.">
        <PrimeColumn header="TANGGAL / JAM" style="min-width:140px">
          <template #body="{ data }">
            <div class="cppt-table-main">
              <strong>{{ data.tanggal }}</strong>
              <span>{{ data.jam }}</span>
            </div>
          </template>
        </PrimeColumn>
        <PrimeColumn header="PETUGAS" style="min-width:210px">
          <template #body="{ data }">
            <div class="cppt-table-main">
              <strong>{{ data.nama_petugas || data.nip }}</strong>
              <span>{{ data.jabatan || data.nip }}</span>
            </div>
          </template>
        </PrimeColumn>
        <PrimeColumn header="SKOR" style="min-width:150px">
          <template #body="{ data }">
            <div class="ews-table-score">
              <span :class="['ews-score-pill', scoreClass(data.total_score)]">{{ data.total_score }}</span>
              <strong>{{ data.klasifikasi }}</strong>
            </div>
          </template>
        </PrimeColumn>
        <PrimeColumn header="TANDA VITAL" style="min-width:270px">
          <template #body="{ data }">
            <div class="cppt-table-vitals">
              <span>RR <b>{{ data.pernafasan || '-' }}</b></span>
              <span>SpO2 <b>{{ data.saturasi || '-' }}</b></span>
              <span>Suhu <b>{{ data.suhu || '-' }}</b></span>
              <span>TD <b>{{ data.tekanan || '-' }}/{{ data.diastol || '-' }}</b></span>
              <span>Nadi <b>{{ data.denyut || '-' }}</b></span>
              <span>Sadar <b>{{ data.kesadaran || '-' }}</b></span>
            </div>
          </template>
        </PrimeColumn>
        <PrimeColumn header="RESPON / TINDAKAN" style="min-width:320px">
          <template #body="{ data }">
            <div class="cppt-table-extra">
              <p><b>Respon:</b> {{ data.respon || '-' }}</p>
              <p><b>Tindakan:</b> {{ data.tindakan || '-' }}</p>
              <p><b>Frekuensi:</b> {{ data.frekuensi || '-' }}</p>
            </div>
          </template>
        </PrimeColumn>
        <PrimeColumn header="BALANCE CAIRAN" style="min-width:180px">
          <template #body="{ data }">
            <div class="cppt-table-main">
              <strong>{{ data.bc || '0' }}</strong>
              <span>Masuk {{ data.jumlahmasuk || '0' }} / Keluar {{ data.jumlahkeluar || '0' }}</span>
            </div>
          </template>
        </PrimeColumn>
        <PrimeColumn header="AKSI" frozen align-frozen="right" style="min-width:145px">
          <template #body="{ data }">
            <div class="cppt-table-actions">
              <button
                type="button"
                :disabled="!billingLocked && !data.bisa_diubah"
                :title="billingLocked ? 'Lihat EWS' : 'Edit EWS'"
                @click="billingLocked ? viewRecord(data) : editRecord(data)"
              >
                <Eye v-if="billingLocked" :size="14" />
                <Pencil v-else :size="14" />
                {{ billingLocked ? 'Lihat' : 'Edit' }}
              </button>
              <button
                v-if="!billingLocked"
                type="button"
                class="danger"
                :disabled="!data.bisa_diubah"
                title="Hapus EWS"
                @click="deleteTarget = data"
              >
                <Trash2 :size="14" /> Hapus
              </button>
            </div>
          </template>
        </PrimeColumn>
      </DataTable>
    </article>

    <div v-if="deleteTarget" class="cppt-confirm-backdrop" @click.self="deleteTarget = null">
      <section class="cppt-confirm-dialog" role="dialog" aria-modal="true">
        <h3>Hapus EWS Ranap?</h3>
        <p>Catatan tanggal {{ deleteTarget.tanggal }} pukul {{ deleteTarget.jam }} akan dihapus dari SIMRS Khanza.</p>
        <div>
          <button type="button" class="cppt-button secondary" :disabled="deleting" @click="deleteTarget = null">Batal</button>
          <button type="button" class="cppt-button danger" :disabled="deleting" @click="deleteRecord(deleteTarget)">
            <LoaderCircle v-if="deleting" class="spin" :size="15" />
            <Trash2 v-else :size="15" />
            {{ deleting ? 'Menghapus...' : 'Hapus' }}
          </button>
        </div>
      </section>
    </div>
  </section>
</template>

<style scoped>
.ews-time-grid {
  grid-template-columns: minmax(150px, 180px) minmax(140px, 170px) minmax(320px, 1fr);
}

.ews-sheet {
  display: grid;
  gap: 12px;
  margin-top: 14px;
  padding: 12px;
  border: 1px solid #b8ddd7;
  border-radius: 12px;
  background: rgba(255, 255, 255, .58);
}

.ews-subheader {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--line);
}

.ews-subheader span {
  color: #0d9488;
  font-size: 9px;
  font-weight: 850;
  letter-spacing: .11em;
  text-transform: uppercase;
}

.ews-subheader h4 {
  margin: 3px 0 0;
  color: var(--text);
  font-size: 14px;
}

.ews-vital-grid,
.ews-fluid-grid {
  grid-template-columns: repeat(4, minmax(130px, 1fr));
  margin-top: 0 !important;
}

.ews-fluid-grid {
  grid-template-columns: repeat(5, minmax(125px, 1fr));
}

.ews-soap-grid {
  margin-top: 14px;
}

.ews-pain-grid {
  display: grid;
  grid-template-columns: minmax(280px, 430px) minmax(320px, 1fr);
  gap: 14px;
  align-items: center;
}

.ews-pain-image {
  display: grid;
  place-items: center;
  min-height: 150px;
  margin: 0;
  padding: 12px;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: #fff;
}

.ews-pain-image img {
  display: block;
  width: 100%;
  max-width: 410px;
  height: auto;
  object-fit: contain;
}

.ews-pain-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(130px, 1fr));
  gap: 12px;
  align-items: end;
}

.ews-pain-fields .pain-scale {
  grid-column: 1 / -1;
  max-width: 240px;
}

.ews-score-chip,
.ews-balance {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 6px 12px;
  border: 1px solid rgba(13, 148, 136, .24);
  border-radius: 999px;
  background: rgba(13, 148, 136, .10);
  color: #0d9488;
}

.ews-score-chip strong,
.ews-balance strong {
  font-size: 18px;
  line-height: 1;
}

.ews-score-chip span,
.ews-balance span {
  color: inherit;
  font-size: 10px;
  letter-spacing: .02em;
  text-transform: none;
}

.ews-score-chip.info,
.ews-score-pill.info {
  border-color: rgba(59, 130, 246, .26);
  background: rgba(59, 130, 246, .12);
  color: #3b82f6;
}

.ews-score-chip.warning,
.ews-score-pill.warning {
  border-color: rgba(245, 158, 11, .28);
  background: rgba(245, 158, 11, .13);
  color: #f59e0b;
}

.ews-score-chip.danger,
.ews-score-pill.danger {
  border-color: rgba(239, 68, 68, .28);
  background: rgba(239, 68, 68, .13);
  color: #ef4444;
}

.ews-score-list {
  display: grid;
  grid-template-columns: repeat(7, minmax(75px, 1fr));
  gap: 8px;
}

.ews-score-list span {
  min-height: 34px;
  padding: 8px 10px;
  border: 1px solid var(--line);
  border-radius: 10px;
  background: var(--surface);
  color: var(--muted);
  font-size: 11px;
}

.ews-score-list b {
  color: var(--text);
}

.ews-alert {
  padding: 11px 13px;
  border: 1px solid color-mix(in srgb, #ef4444 36%, var(--line));
  border-radius: 10px;
  background: color-mix(in srgb, #ef4444 10%, var(--surface));
  color: var(--text);
  font-size: 12px;
}

.ews-table-score {
  display: grid;
  gap: 6px;
  justify-items: start;
}

.ews-score-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 34px;
  min-height: 26px;
  border-radius: 999px;
  background: rgba(13, 148, 136, .12);
  color: #0d9488;
  font-weight: 800;
}

:global(.theme-dark .ews-sheet),
:global(.sirapi-dark .ews-sheet) {
  border-color: #304258;
  background: #142132;
}

:global(.theme-dark .ews-pain-image),
:global(.sirapi-dark .ews-pain-image) {
  border-color: #3b4b61;
  background: #f8fafc;
}

:global(.theme-dark .ews-score-list span),
:global(.sirapi-dark .ews-score-list span) {
  background: #1c283a;
}

.ews-form :deep(.staff-search-results) {
  z-index: 1100;
}

@media (max-width: 1180px) {
  .ews-time-grid,
  .ews-vital-grid,
  .ews-fluid-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .ews-score-list {
    grid-template-columns: repeat(4, minmax(75px, 1fr));
  }

  .ews-pain-grid {
    grid-template-columns: 1fr;
  }

  .ews-pain-image {
    justify-items: start;
  }
}

@media (max-width: 760px) {
  .ews-time-grid,
  .ews-vital-grid,
  .ews-fluid-grid,
  .ews-soap-grid,
  .ews-pain-fields {
    grid-template-columns: 1fr;
  }

  .ews-pain-fields .pain-scale {
    max-width: none;
  }

  .ews-subheader {
    align-items: flex-start;
    flex-direction: column;
  }

  .ews-score-list {
    grid-template-columns: repeat(2, minmax(75px, 1fr));
  }
}
</style>
