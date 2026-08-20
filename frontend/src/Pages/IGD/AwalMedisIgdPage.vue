<script setup>
import { ChevronDown, ChevronUp, LoaderCircle, RotateCcw, Save, Trash2 } from '@lucide/vue'
import Dialog from 'primevue/dialog'
import { computed, reactive, ref, watch } from 'vue'
import CariDokter from '../../Components/Ui/CariDokter.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import { awalMedisIgdData, hapusAwalMedisIgd, simpanAwalMedisIgd, ubahAwalMedisIgd } from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({ token: { type: String, required: true }, patient: { type: Object, required: true } })
const notifikasi = useNotifikasi()
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const formOpen = ref(true)
const confirmDelete = ref(false)
const data = reactive({ tersedia: false, billing_terkunci: false, penilaian: null, dokter: null, pilihan: {} })
const dokter = ref({})
const form = reactive(formKosong())

const editing = computed(() => Boolean(data.tersedia))
const billingLocked = computed(() => Boolean(data.billing_terkunci))
const pilihan = computed(() => ({
  anamnesis: list(data.pilihan?.anamnesis || ['Autoanamnesis', 'Alloanamnesis']),
  keadaan: list(data.pilihan?.keadaan || ['Sehat', 'Sakit Ringan', 'Sakit Sedang', 'Sakit Berat']),
  kesadaran: list(data.pilihan?.kesadaran || ['Compos Mentis', 'Apatis', 'Somnolen', 'Sopor', 'Koma']),
  pemeriksaan: list(data.pilihan?.pemeriksaan || ['Normal', 'Abnormal', 'Tidak Diperiksa']),
}))

function list(items) { return items.map((value) => ({ label: value, value })) }
function sekarang() { const d = new Date(Date.now() - new Date().getTimezoneOffset() * 60000); return d.toISOString().slice(0, 19) }
function tanggalInput(value) { return String(value || '').replace(' ', 'T').slice(0, 19) }
function formKosong() {
  return {
    no_rawat: props.patient?.no_rawat || '', tanggal: sekarang(), kode_dokter: '', anamnesis: 'Autoanamnesis', hubungan: '',
    keluhan_utama: '', rps: '', rpd: '', rpk: '', rpo: '', alergi: '', keadaan: 'Sehat', gcs: '', kesadaran: 'Compos Mentis',
    td: '', nadi: '', rr: '', suhu: '', spo2: '', bb: '', tb: '', kepala: 'Normal', mata: 'Normal', gigi: 'Normal', leher: 'Normal',
    thoraks: 'Normal',   abdomen: 'Normal', genital: 'Normal', ekstremitas: 'Normal', 
    ket_fisik: '', ket_lokalis: '', lab: '', rad: '', ekg: '', diagnosis: '', tata: '', 
  }
}

function isiForm() {
  const record = data.penilaian || {}
  Object.assign(form, formKosong(), record, {
    no_rawat: props.patient.no_rawat,
    tanggal: tanggalInput(record.tanggal) || sekarang(),
  })
  dokter.value = data.dokter ? { ...data.dokter } : {}
}

async function loadData() {
  if (!props.token || !props.patient?.no_rawat) return
  loading.value = true
  try {
    const response = await awalMedisIgdData(props.token, props.patient.no_rawat)
    Object.assign(data, response || {})
    isiForm()
    formOpen.value = true
  } catch (error) {
    notifikasi.gagal(error.message || 'Penilaian Awal Medis IGD tidak dapat dibaca.')
  } finally { loading.value = false }
}

function payload() {
  return { ...form, tanggal: form.tanggal.replace('T', ' '), kode_dokter: dokter.value?.kode || '' }
}

async function save() {
  if (billingLocked.value) { notifikasi.peringatan('Kunjungan sudah masuk billing. Penilaian hanya dapat dilihat.'); return }
  if (!dokter.value?.kode) { notifikasi.peringatan('Pilih dokter yang melakukan penilaian.'); return }
  saving.value = true
  try {
    const response = editing.value ? await ubahAwalMedisIgd(props.token, payload()) : await simpanAwalMedisIgd(props.token, payload())
    notifikasi.sukses(response?.pesan || 'Penilaian Awal Medis IGD berhasil disimpan.')
    await loadData()
  } catch (error) {
    notifikasi.gagal(error.message || 'Penilaian Awal Medis IGD gagal disimpan.')
  } finally { saving.value = false }
}

async function remove() {
  deleting.value = true
  try {
    const response = await hapusAwalMedisIgd(props.token, props.patient.no_rawat)
    confirmDelete.value = false
    notifikasi.sukses(response?.pesan || 'Penilaian Awal Medis IGD berhasil dihapus.')
    await loadData()
  } catch (error) {
    notifikasi.gagal(error.message || 'Penilaian Awal Medis IGD gagal dihapus.')
  } finally { deleting.value = false }
}

watch(() => props.patient.no_rawat, loadData, { immediate: true })
</script>

<template>
  <section class="medical-page">
    <div v-if="loading" class="medical-state"><LoaderCircle class="spin" :size="28"/><strong>Menarik penilaian Awal Medis IGD...</strong></div>
    <form v-else class="medical-form cppt-form-card cppt-form" @submit.prevent="save">
      <header class="cppt-section-header">
        <div><span>PENILAIAN MEDIS IGD</span><h3>{{ editing ? 'Edit' : 'Input' }} Awal Medis IGD</h3><p>Penilaian awal dokter pasien dewasa mengikuti form SIMRS Khanza.</p></div>
        <div class="header-actions">
          <button v-if="editing && !billingLocked" type="button" class="cppt-button danger" :disabled="deleting" @click="confirmDelete = true"><Trash2 :size="15"/>Hapus</button>
          <button type="button" class="cppt-button primary icon-only" :title="formOpen ? 'Sembunyikan form input' : 'Tampilkan form input'" @click="formOpen = !formOpen"><ChevronUp v-if="formOpen" :size="18"/><ChevronDown v-else :size="18"/></button>
        </div>
      </header>

      <div v-if="billingLocked" class="medical-lock">Billing sudah terverifikasi atau kunjungan dibatalkan. Form Awal Medis IGD hanya dapat dilihat.</div>

      <fieldset v-show="formOpen" :disabled="saving || billingLocked">
        <div class="medical-sheet">
          <section class="medical-section meta-section">
            <div class="medical-grid meta-grid">
              <FormInput v-model="form.tanggal" label="Tanggal Penilaian" type="datetime-local" step="1" required/>
              <CariDokter v-model="dokter" class="doctor-field" :token="token" sumber="awal-medis-ranap" required/>
            </div>
          </section>

          <section class="medical-section">
            <h4>I. RIWAYAT KESEHATAN</h4>
            <div class="medical-grid history-grid">
              <FormInput v-model="form.anamnesis" label="Anamnesis" jenis="select" :options="pilihan.anamnesis" required/>
              <FormInput v-model="form.hubungan" label="Hubungan" maxlength="30" placeholder="Hubungan pemberi informasi dengan pasien"/>
              <FormInput v-model="form.keluhan_utama" class="wide" label="Keluhan Utama" jenis="textarea" :rows="3" maxlength="2000" required/>
              <FormInput v-model="form.rps" label="Riwayat Penyakit Sekarang" jenis="textarea" :rows="4" maxlength="2000" required/>
              <FormInput v-model="form.rpd" label="Riwayat Penyakit Dahulu" jenis="textarea" :rows="4" maxlength="1000" required/>
              <FormInput v-model="form.rpk" label="Riwayat Penyakit Keluarga" jenis="textarea" :rows="4" maxlength="1000" required/>
              <FormInput v-model="form.rpo" label="Riwayat Penggunaan Obat" jenis="textarea" :rows="4" maxlength="1000" required/>
              <FormInput v-model="form.alergi" class="wide" label="Alergi" maxlength="50" placeholder="Obat, makanan, atau alergen lainnya"/>
            </div>
          </section>

          <section class="medical-section">
            <h4>II. PEMERIKSAAN FISIK</h4>
            <div class="medical-grid vital-grid">
              <FormInput v-model="form.keadaan" label="Keadaan Umum" jenis="select" :options="pilihan.keadaan"/>
              <FormInput v-model="form.gcs" label="GCS(E,V,M)" maxlength="10"/>
              <FormInput v-model="form.kesadaran" label="Kesadaran" jenis="select" :options="pilihan.kesadaran"/>
              <FormInput v-model="form.td" label="Tekanan Darah (mmHg)" maxlength="8"/>
              <FormInput v-model="form.nadi" label="Nadi (/menit)" maxlength="5"/>
              <FormInput v-model="form.rr" label="Respirasi (/menit)" maxlength="5"/>
              <FormInput v-model="form.suhu" label="Suhu (°C)" maxlength="5"/>
              <FormInput v-model="form.spo2" label="SpO2 (%)" maxlength="5"/>
              <FormInput v-model="form.bb" label="Berat Badan (kg)" maxlength="5"/>
              <FormInput v-model="form.tb" label="Tinggi Badan (cm)" maxlength="5"/>
            </div>
            <h5>Pemeriksaan Organ</h5>
            <div class="medical-grid organ-grid">
              <FormInput v-model="form.kepala" label="Kepala" jenis="select" :options="pilihan.pemeriksaan"/>
              <FormInput v-model="form.mata" label="Mata" jenis="select" :options="pilihan.pemeriksaan"/>
              <FormInput v-model="form.gigi" label="Gigi & Mulut" jenis="select" :options="pilihan.pemeriksaan"/>
              <FormInput v-model="form.leher" label="Leher" jenis="select" :options="pilihan.pemeriksaan"/>
              <FormInput v-model="form.thoraks" label="Thoraks" jenis="select" :options="pilihan.pemeriksaan"/>
              
              
              <FormInput v-model="form.abdomen" label="Abdomen" jenis="select" :options="pilihan.pemeriksaan"/>
              <FormInput v-model="form.genital" label="Genital & Anus" jenis="select" :options="pilihan.pemeriksaan"/>
              <FormInput v-model="form.ekstremitas" label="Ekstremitas" jenis="select" :options="pilihan.pemeriksaan"/>
              
              <FormInput v-model="form.ket_fisik" class="wide" label="Keterangan Pemeriksaan Fisik" jenis="textarea" :rows="4" maxlength="5000"/>
            </div>
          </section>

          <section class="medical-section">
            <h4>III. Status Lokalis</h4>
            <div class="lokalis-container">
              <img src="/img/lokalis.png" alt="Status Lokalis" class="lokalis-image"/>
            </div>
            <FormInput v-model="form.ket_lokalis" label="Keterangan Status Lokalis" jenis="textarea" :rows="5" maxlength="3000" placeholder="Tuliskan lokasi dan temuan pemeriksaan setempat"/>
          </section>
          <section class="medical-section conclusion-section">
            <h4>IV. PEMERIKSAAN PENUNJANG</h4>
            <div class="medical-grid" style="grid-template-columns: repeat(3, 1fr);">
              <FormInput v-model="form.ekg" label="EKG" jenis="textarea" :rows="3" maxlength="1000"/>
              <FormInput v-model="form.rad" label="Radiologi" jenis="textarea" :rows="3" maxlength="1000"/>
              <FormInput v-model="form.lab" label="Laboratorium" jenis="textarea" :rows="3" maxlength="1000"/>
            </div>
          </section>
            <section class="medical-section conclusion-section">
              <h4>V. DIAGNOSA/ASESMEN</h4>
              <div class="medical-grid" style="grid-template-columns: 1fr;">
                <FormInput v-model="form.diagnosis" class="wide" jenis="textarea" :rows="3" maxlength="500"/>
              </div>
            </section>
            <section class="medical-section conclusion-section">
              <h4>VI. TATA LAKSANA</h4>
              <div class="medical-grid" style="grid-template-columns: 1fr;">
                <FormInput v-model="form.tata" class="wide" jenis="textarea" :rows="5" maxlength="5000"/>
              </div>
            </section>
        </div>

        <footer class="cppt-form-actions">
          <button type="button" class="cppt-button secondary" @click="isiForm"><RotateCcw :size="15"/>Batal / Reset</button>
          <button type="submit" class="cppt-button primary" :disabled="saving"><LoaderCircle v-if="saving" class="spin" :size="15"/><Save v-else :size="15"/>{{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan Penilaian' }}</button>
        </footer>
      </fieldset>
    </form>

    <Dialog v-model:visible="confirmDelete" modal header="Hapus Penilaian Awal Medis IGD" class="medical-dialog" :style="{ width: 'min(440px, 92vw)' }">
      <p>Penilaian pasien ini akan dihapus dari SIMRS Khanza.</p>
      <template #footer><button type="button" class="cppt-button secondary" @click="confirmDelete = false">Batal</button><button type="button" class="cppt-button danger" :disabled="deleting" @click="remove"><LoaderCircle v-if="deleting" class="spin" :size="14"/><Trash2 v-else :size="14"/>{{ deleting ? 'Menghapus...' : 'Hapus' }}</button></template>
    </Dialog>
  </section>
</template>

<style scoped>
.medical-page{--border:var(--line);--muted:var(--text-muted);display:grid;gap:16px}.medical-state{min-height:280px;display:grid;place-content:center;justify-items:center;gap:10px;color:var(--muted)}.medical-form{overflow:visible;border:1px solid var(--border);border-radius:16px;background:var(--surface)}.medical-form>header{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:18px 20px;border-bottom:1px solid var(--border)}header span{color:var(--primary);font-size:.69rem;font-weight:700;letter-spacing:.12em}header h3{margin:4px 0;font-size:1.12rem}header p{margin:0;color:var(--muted);font-size:.8rem}.header-actions{display:flex;align-items:center;gap:8px}.medical-form fieldset{display:grid;gap:14px;margin:0;padding:20px;border:0;background:color-mix(in srgb,var(--primary) 3%,var(--surface))}.medical-sheet{overflow:hidden;border:1px solid var(--border);border-radius:13px;background:var(--surface)}.medical-section{padding:15px;border-bottom:1px solid var(--border)}.medical-section:last-child{border-bottom:0}.medical-section:nth-child(odd){background:color-mix(in srgb,var(--surface-soft) 62%,var(--surface))}.medical-section h4{display:flex;align-items:center;gap:7px;margin:0 0 13px;color:var(--text);font-size:.8rem;font-weight:700;letter-spacing:.035em;text-transform:uppercase}.medical-section h4 span{color:var(--primary)}.medical-section h5{margin:14px 0 10px;padding-top:12px;border-top:1px dashed var(--border);color:var(--muted);font-size:.7rem;letter-spacing:.08em;text-transform:uppercase}.medical-grid{display:grid;gap:11px}.meta-grid{grid-template-columns:240px minmax(330px,1fr)}.history-grid{grid-template-columns:1fr 1fr}.history-grid .wide,.organ-grid .wide{grid-column:1/-1}.vital-grid{grid-template-columns:repeat(5,minmax(125px,1fr))}.organ-grid{grid-template-columns:repeat(4,minmax(145px,1fr))}.conclusion-grid{grid-template-columns:1fr 1fr}.medical-lock{margin:14px 20px 0;padding:11px 13px;border:1px solid color-mix(in srgb,#f59e0b 36%,var(--border));border-radius:10px;background:color-mix(in srgb,#f59e0b 10%,var(--surface));color:var(--text);font-size:.82rem}.cppt-form-actions{display:flex;justify-content:flex-end;gap:9px;padding-top:2px}:global(.theme-light .medical-form fieldset){background:#f6f8fb}:global(.theme-light .medical-sheet){border-color:#d8e1eb;background:#fff}:global(.theme-light .medical-section:nth-child(odd)){background:#fbfcfd}:global(.theme-light .meta-section){background:#f5f8fb}:global(.theme-dark .medical-form fieldset),:global(.sirapi-dark .medical-form fieldset){background:linear-gradient(135deg,#111c2c 0%,#0e1b29 100%)}:global(.theme-dark .medical-sheet),:global(.sirapi-dark .medical-sheet){border-color:#304258;background:#142132}:global(.theme-dark .medical-section),:global(.sirapi-dark .medical-section),:global(.theme-dark .medical-section:nth-child(odd)),:global(.sirapi-dark .medical-section:nth-child(odd)){border-color:#304258;background:#142132}:global(.theme-dark .meta-section),:global(.sirapi-dark .meta-section){background:#111d2c}:global(.medical-dialog){color:var(--text);background:var(--surface)}@media(max-width:1200px){.vital-grid{grid-template-columns:repeat(3,minmax(130px,1fr))}.organ-grid{grid-template-columns:repeat(3,minmax(140px,1fr))}}@media(max-width:800px){.meta-grid,.history-grid,.conclusion-grid{grid-template-columns:1fr}.vital-grid,.organ-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.doctor-field,.history-grid .wide,.organ-grid .wide{grid-column:1/-1}}@media(max-width:540px){.medical-form>header{align-items:flex-start}.header-actions{flex-direction:column-reverse;align-items:flex-end}.vital-grid,.organ-grid{grid-template-columns:1fr}.doctor-field,.history-grid .wide,.organ-grid .wide{grid-column:auto}.medical-form fieldset{padding:12px}.medical-section{padding:12px}}
.lokalis-container{display:flex;justify-content:center;margin-bottom:15px;background-color:color-mix(in srgb,var(--surface-soft) 40%,var(--surface));border:1px solid var(--border);border-radius:8px;padding:10px}.lokalis-image{max-width:100%;height:auto;border-radius:4px;filter:drop-shadow(0 0 2px rgba(0,0,0,0.1));}:global(.theme-dark .lokalis-image),:global(.sirapi-dark .lokalis-image){filter:invert(0.85) hue-rotate(180deg) drop-shadow(0 0 2px rgba(0,0,0,0.5))}
</style>
