<script setup>
import { FileCheck2, LoaderCircle, Save, Trash2, X } from '@lucide/vue'
import { computed, reactive, ref, watch } from 'vue'
import CariDokter from '../../Components/Ui/CariDokter.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import InputPencarian from '../../Components/Ui/InputPencarian.vue'
import {
  cariCodingResumePasienRanap,
  hapusResumePasienRanap,
  resumePasienRanapData,
  simpanResumePasienRanap,
  ubahResumePasienRanap,
  validasiCodingResumePasienRanap,
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
const tersedia = ref(false)
const billingLocked = ref(false)
const hapusTerbuka = ref(false)
const doctor = ref({})
const pilihan = reactive({ cara_keluar: [], keadaan: [], dilanjutkan: [] })
const form = reactive(formKosong())
const statusCoding = reactive({})

const pilihanSelect = computed(() => ({
  cara_keluar: pilihan.cara_keluar.map((value) => ({ label: value, value })),
  keadaan: pilihan.keadaan.map((value) => ({ label: value, value })),
  dilanjutkan: pilihan.dilanjutkan.map((value) => ({ label: value, value })),
}))

const catatanPerawatan = [
  { key: 'keluhan_utama', label: 'Keluhan Utama / Riwayat Penyakit', required: true },
  { key: 'pemeriksaan_fisik', label: 'Pemeriksaan Fisik' },
  { key: 'jalannya_penyakit', label: 'Jalannya Penyakit Selama Perawatan', required: true },
  { key: 'pemeriksaan_penunjang', label: 'Pemeriksaan Penunjang Radiologi Terpenting' },
  { key: 'hasil_laborat', label: 'Pemeriksaan Penunjang Laboratorium Terpenting' },
  { key: 'tindakan_dan_operasi', label: 'Tindakan / Operasi Selama Perawatan' },
  { key: 'obat_di_rs', label: 'Obat-obatan Selama Perawatan' },
]

const diagnosa = [
  { kode: 'kd_diagnosa_utama', nama: 'diagnosa_utama', label: 'Diagnosa Utama', required: true },
  { kode: 'kd_diagnosa_sekunder', nama: 'diagnosa_sekunder', label: 'Diagnosa Sekunder 1' },
  { kode: 'kd_diagnosa_sekunder2', nama: 'diagnosa_sekunder2', label: 'Diagnosa Sekunder 2' },
  { kode: 'kd_diagnosa_sekunder3', nama: 'diagnosa_sekunder3', label: 'Diagnosa Sekunder 3' },
  { kode: 'kd_diagnosa_sekunder4', nama: 'diagnosa_sekunder4', label: 'Diagnosa Sekunder 4' },
]

const prosedur = [
  { kode: 'kd_prosedur_utama', nama: 'prosedur_utama', label: 'Prosedur Utama' },
  { kode: 'kd_prosedur_sekunder', nama: 'prosedur_sekunder', label: 'Prosedur Sekunder 1' },
  { kode: 'kd_prosedur_sekunder2', nama: 'prosedur_sekunder2', label: 'Prosedur Sekunder 2' },
  { kode: 'kd_prosedur_sekunder3', nama: 'prosedur_sekunder3', label: 'Prosedur Sekunder 3' },
]

function sekarang() {
  const date = new Date(Date.now() - new Date().getTimezoneOffset() * 60000)
  return date.toISOString().slice(0, 16)
}

function formKosong() {
  return {
    no_rawat: props.patient?.no_rawat || '', kode_dokter: '', diagnosa_awal: '', alasan: '',
    keluhan_utama: '', pemeriksaan_fisik: '', jalannya_penyakit: '', pemeriksaan_penunjang: '',
    hasil_laborat: '', tindakan_dan_operasi: '', obat_di_rs: '', diagnosa_utama: '',
    kd_diagnosa_utama: '', diagnosa_sekunder: '', kd_diagnosa_sekunder: '', diagnosa_sekunder2: '',
    kd_diagnosa_sekunder2: '', diagnosa_sekunder3: '', kd_diagnosa_sekunder3: '', diagnosa_sekunder4: '',
    kd_diagnosa_sekunder4: '', prosedur_utama: '', kd_prosedur_utama: '', prosedur_sekunder: '',
    kd_prosedur_sekunder: '', prosedur_sekunder2: '', kd_prosedur_sekunder2: '', prosedur_sekunder3: '',
    kd_prosedur_sekunder3: '', alergi: '', diet: '', lab_belum: '', edukasi: '',
    cara_keluar: 'Atas Izin Dokter', ket_keluar: '', keadaan: 'Membaik', ket_keadaan: '',
    dilanjutkan: 'Kembali Ke RS', ket_dilanjutkan: '', kontrol: sekarang(), obat_pulang: '',
  }
}

function dariServer(value = {}) {
	resetStatusCoding()
  Object.assign(form, formKosong(), value, {
    no_rawat: props.patient.no_rawat,
    kontrol: String(value.kontrol || sekarang()).replace(' ', 'T').slice(0, 16),
  })
}

function resetStatusCoding() {
  Object.keys(statusCoding).forEach((key) => delete statusCoding[key])
}

function daftarCoding(jenis) {
  return jenis === 'diagnosa' ? diagnosa : prosedur
}

function kodeUtama(jenis, item) {
  const daftar = daftarCoding(jenis)
  const posisi = daftar.findIndex((entry) => entry.kode === item.kode)
  return posisi <= 0 || !daftar.slice(0, posisi).some((entry) => String(form[entry.kode] || '').trim())
}

function kodeDuplikat(jenis, item, kode) {
  return daftarCoding(jenis).some((entry) => entry.kode !== item.kode && String(form[entry.kode] || '').trim().toUpperCase() === kode)
}

async function periksaCoding(item, jenis) {
  const kode = String(form[item.kode] || '').trim().toUpperCase()
  if (!kode) {
    form[item.nama] = ''
    delete statusCoding[item.kode]
    return true
  }
  if (kodeDuplikat(jenis, item, kode)) {
    form[item.nama] = ''
    statusCoding[item.kode] = { valid: false, error: `Kode ${kode} sudah digunakan.` }
    return false
  }

  statusCoding[item.kode] = { loading: true, hint: 'Memeriksa kode ke master SIMRS...' }
  try {
    const data = await validasiCodingResumePasienRanap(props.token, {
      jenis,
      kode,
      utama: kodeUtama(jenis, item) ? '1' : '0',
    })
    if (String(form[item.kode] || '').trim().toUpperCase() !== kode) return false
    if (!data?.valid) {
      form[item.nama] = ''
      statusCoding[item.kode] = { valid: false, error: data?.pesan || `Kode ${kode} tidak valid.` }
      return false
    }
    form[item.kode] = data.kode
    form[item.nama] = data.nama
    statusCoding[item.kode] = {
      valid: true,
      hint: data.pesan || 'Kode valid',
      im: data.im === '1',
    }
    return true
  } catch (error) {
    if (String(form[item.kode] || '').trim().toUpperCase() !== kode) return false
    form[item.nama] = ''
    statusCoding[item.kode] = { valid: false, error: error.message || 'Kode tidak dapat diperiksa.' }
    return false
  }
}

function nilaiCoding(item) {
  if (!form[item.kode]) return {}
  return {
    kode: form[item.kode],
    nama: form[item.nama],
    penanda: statusCoding[item.kode]?.im ? 'IM' : '',
  }
}

async function cariPilihanCoding(item, jenis, kataKunci) {
  return cariCodingResumePasienRanap(props.token, {
    jenis,
    q: kataKunci,
    utama: kodeUtama(jenis, item) ? '1' : '0',
  })
}

function pilihCoding(item, jenis, pilihanCoding) {
  if (!pilihanCoding?.kode) {
    form[item.kode] = ''
    form[item.nama] = ''
    delete statusCoding[item.kode]
    return
  }
  const kode = String(pilihanCoding.kode).trim().toUpperCase()
  if (!pilihanCoding.valid) {
    form[item.kode] = kode
    form[item.nama] = pilihanCoding.nama || ''
    statusCoding[item.kode] = { valid: false, error: pilihanCoding.pesan || `Kode ${kode} tidak valid.` }
    notifikasi.peringatan(pilihanCoding.pesan || `Kode ${kode} tidak valid.`)
    return
  }
  if (kodeDuplikat(jenis, item, kode)) {
    form[item.kode] = kode
    form[item.nama] = pilihanCoding.nama || ''
    statusCoding[item.kode] = { valid: false, error: `Kode ${kode} sudah digunakan.` }
    notifikasi.peringatan(`Kode ${kode} sudah digunakan.`)
    return
  }
  form[item.kode] = kode
  form[item.nama] = pilihanCoding.nama || ''
  statusCoding[item.kode] = {
    valid: true,
    hint: pilihanCoding.pesan || 'Kode valid',
    im: pilihanCoding.im === '1',
  }
}

async function periksaSemuaCoding() {
  const daftar = [
    ...diagnosa.filter((item) => String(form[item.kode] || '').trim()).map((item) => [item, 'diagnosa']),
    ...prosedur.filter((item) => String(form[item.kode] || '').trim()).map((item) => [item, 'prosedur']),
  ]
  const hasil = await Promise.all(daftar.map(([item, jenis]) => periksaCoding(item, jenis)))
  return hasil.every(Boolean)
}

async function muat() {
  if (!props.patient.no_rawat) return
  loading.value = true
  try {
    const data = await resumePasienRanapData(props.token, props.patient.no_rawat)
    tersedia.value = Boolean(data?.tersedia)
    billingLocked.value = Boolean(data?.billing_terkunci)
    Object.assign(pilihan, data?.pilihan || {})
    dariServer(data?.resume)
    doctor.value = data?.dokter || {}
  } catch (error) {
    notifikasi.gagal(error.message || 'Resume pasien rawat inap tidak dapat dibaca.')
  } finally {
    loading.value = false
  }
}

function payload() {
  return {
    ...form,
    no_rawat: props.patient.no_rawat,
    kode_dokter: doctor.value?.kode || '',
    kontrol: String(form.kontrol || '').replace('T', ' '),
  }
}

async function simpan() {
  if (billingLocked.value) {
    notifikasi.peringatan('Billing sudah terverifikasi. Resume hanya dapat dilihat.')
    return
  }
  if (!doctor.value?.kode || !form.keluhan_utama.trim() || !form.jalannya_penyakit.trim() || !form.kd_diagnosa_utama.trim()) {
    notifikasi.peringatan('Dokter, keluhan utama, jalannya penyakit, dan diagnosa utama wajib diisi.')
    return
  }
  if (!await periksaSemuaCoding()) {
    notifikasi.peringatan('Periksa kembali kode diagnosa atau prosedur yang ditandai.')
    return
  }
  saving.value = true
  try {
    const response = tersedia.value
      ? await ubahResumePasienRanap(props.token, payload())
      : await simpanResumePasienRanap(props.token, payload())
    notifikasi.sukses(response?.pesan || 'Resume pasien rawat inap berhasil disimpan.')
    await muat()
  } catch (error) {
    notifikasi.gagal(error.message || 'Resume pasien rawat inap gagal disimpan.')
  } finally {
    saving.value = false
  }
}

async function hapus() {
  deleting.value = true
  try {
    const response = await hapusResumePasienRanap(props.token, props.patient.no_rawat)
    notifikasi.sukses(response?.pesan || 'Resume pasien rawat inap berhasil dihapus.')
    hapusTerbuka.value = false
    await muat()
  } catch (error) {
    notifikasi.gagal(error.message || 'Resume pasien rawat inap gagal dihapus.')
  } finally {
    deleting.value = false
  }
}

watch(() => props.patient.no_rawat, muat, { immediate: true })
</script>

<template>
  <section class="resume-ranap-page">
    <article class="resume-card">
      <header class="resume-header">
        <div>
          <span>REKAM MEDIS RAWAT INAP</span>
          <h3>Resume Pasien Ranap</h3>
          <p>Ringkasan pelayanan pasien selama menjalani perawatan.</p>
        </div>
        <div class="resume-status" :class="{ available: tersedia }">
          <FileCheck2 :size="17" /> {{ tersedia ? 'Resume Tersimpan' : 'Resume Baru' }}
        </div>
      </header>

      <div v-if="loading" class="resume-state"><LoaderCircle class="spin" :size="27" /><strong>Menarik data resume pasien...</strong></div>

      <form v-else class="resume-form" @submit.prevent="simpan">
        <div v-if="billingLocked" class="resume-billing-lock">Billing sudah terverifikasi atau kunjungan dibatalkan. Resume hanya dapat dilihat dan tidak bisa ditambah, diedit, atau dihapus.</div>
        <fieldset :disabled="saving || deleting || billingLocked">
          <section class="resume-section">
            <h4>Informasi Perawatan</h4>
            <div class="resume-grid identity">
              <CariDokter v-model="doctor" :token="token" required />
              <FormInput v-model="form.diagnosa_awal" label="Diagnosa Awal" maxlength="70" />
              <FormInput v-model="form.alasan" label="Alasan Dirawat" maxlength="70" />
            </div>
            <div class="resume-grid notes">
              <FormInput
                v-for="field in catatanPerawatan" :key="field.key"
                v-model="form[field.key]" :label="field.label" jenis="textarea" :rows="3"
                :required="field.required" maxlength="2000"
              />
            </div>
          </section>

          <section class="resume-section coding-section">
            <h4>Diagnosa</h4> 
            <div v-for="item in diagnosa" :key="item.kode" class="resume-code-row">
              <div class="resume-coding-picker" :class="{ 'coding-valid': statusCoding[item.kode]?.valid, invalid: statusCoding[item.kode]?.error }">
                <InputPencarian
                  :model-value="nilaiCoding(item)"
                  :label="`${item.label} (ICD-10)`"
                  placeholder="Ketik kode atau nama diagnosa..."
                  :search="(kataKunci) => cariPilihanCoding(item, 'diagnosa', kataKunci)"
                  right-field="penanda"
                  :required="item.required"
                  :disabled="saving || deleting || billingLocked"
                  @update:model-value="pilihCoding(item, 'diagnosa', $event)"
                />
                <small v-if="statusCoding[item.kode]?.error" class="form-input-message error">{{ statusCoding[item.kode].error }}</small>
                <small v-else-if="statusCoding[item.kode]?.hint" class="form-input-message">{{ statusCoding[item.kode].hint }}</small>
              </div>
            </div>
          </section>

          <section class="resume-section coding-section">
            <h4>Prosedur / Tindakan</h4>
            <div v-for="item in prosedur" :key="item.kode" class="resume-code-row">
              <div class="resume-coding-picker" :class="{ 'coding-valid': statusCoding[item.kode]?.valid, invalid: statusCoding[item.kode]?.error }">
                <InputPencarian
                  :model-value="nilaiCoding(item)"
                  :label="`${item.label} (ICD-9-CM)`"
                  placeholder="Ketik kode atau nama prosedur..."
                  :search="(kataKunci) => cariPilihanCoding(item, 'prosedur', kataKunci)"
                  right-field="penanda"
                  :disabled="saving || deleting || billingLocked"
                  @update:model-value="pilihCoding(item, 'prosedur', $event)"
                />
                <small v-if="statusCoding[item.kode]?.error" class="form-input-message error">{{ statusCoding[item.kode].error }}</small>
                <small v-else-if="statusCoding[item.kode]?.hint" class="form-input-message">{{ statusCoding[item.kode].hint }}</small>
              </div>
            </div>
          </section>

          <section class="resume-section">
            <h4>Kondisi dan Rencana Pulang</h4>
            <div class="resume-grid discharge">
              <FormInput v-model="form.cara_keluar" label="Cara Keluar" jenis="select" :options="pilihanSelect.cara_keluar" required />
              <FormInput v-model="form.ket_keluar" label="Keterangan Cara Keluar" maxlength="50" />
              <FormInput v-model="form.keadaan" label="Keadaan Pulang" jenis="select" :options="pilihanSelect.keadaan" required />
              <FormInput v-model="form.ket_keadaan" label="Keterangan Keadaan Pulang" maxlength="50" />
              <FormInput v-model="form.dilanjutkan" label="Perawatan Dilanjutkan" jenis="select" :options="pilihanSelect.dilanjutkan" required />
              <FormInput v-model="form.ket_dilanjutkan" label="Keterangan Lanjutan" maxlength="50" />
              <FormInput v-model="form.kontrol" label="Tanggal dan Jam Kontrol" type="datetime-local" required />
              <FormInput v-model="form.alergi" label="Alergi Obat" maxlength="100" />
            </div>
            <div class="resume-grid notes discharge-notes">
              <FormInput v-model="form.diet" label="Diet" jenis="textarea" :rows="3" maxlength="2000" />
              <FormInput v-model="form.lab_belum" label="Hasil Laboratorium Belum Selesai" jenis="textarea" :rows="3" maxlength="2000" />
              <FormInput v-model="form.edukasi" label="Instruksi / Anjuran dan Edukasi" jenis="textarea" :rows="3" maxlength="2000" />
              <FormInput v-model="form.obat_pulang" label="Obat Pulang" jenis="textarea" :rows="3" maxlength="2000" />
            </div>
          </section>

          <footer class="resume-actions">
            <button v-if="tersedia && !billingLocked" type="button" class="resume-button danger" @click="hapusTerbuka = true"><Trash2 :size="15" /> Hapus</button>
            <button v-if="!billingLocked" type="submit" class="resume-button primary"><LoaderCircle v-if="saving" class="spin" :size="15" /><Save v-else :size="15" />{{ saving ? 'Menyimpan...' : tersedia ? 'Simpan Perubahan' : 'Simpan Resume' }}</button>
          </footer>
        </fieldset>
      </form>
    </article>

    <div v-if="hapusTerbuka" class="resume-dialog-backdrop" @click.self="hapusTerbuka = false">
      <section class="resume-dialog">
        <h3>Hapus Resume Pasien?</h3>
        <p>Resume untuk nomor rawat {{ patient.no_rawat }} akan dihapus dari SIMRS Khanza.</p>
        <div><button type="button" class="resume-button secondary" @click="hapusTerbuka = false"><X :size="15" /> Batal</button><button type="button" class="resume-button danger" @click="hapus"><LoaderCircle v-if="deleting" class="spin" :size="15" /><Trash2 v-else :size="15" /> Hapus</button></div>
      </section>
    </div>
  </section>
</template>

<style>
.resume-ranap-page{margin-top:15px}.resume-card{overflow:hidden;border:1px solid var(--line);border-radius:16px;background:var(--surface-soft)}
.resume-header{display:flex;align-items:center;justify-content:space-between;gap:15px;padding:17px 19px;border-bottom:1px solid var(--line);background:var(--surface)}
.resume-header span{color:#0d9488;font-size:9px;font-weight:800;letter-spacing:.12em}.resume-header h3{margin:5px 0 0;font-size:18px}.resume-header p{margin:5px 0 0;color:var(--muted);font-size:11px}
.resume-status{display:flex;align-items:center;gap:7px;padding:9px 11px;border-radius:10px;color:#64748b;background:var(--surface-soft);font-size:11px;font-weight:700}.resume-status.available{color:#0f766e;background:rgba(20,184,166,.12)}
.resume-form{padding:15px;background:var(--surface-soft)}.resume-form fieldset{display:grid;gap:14px;margin:0;padding:0;border:0}.resume-section{padding:14px;border:1px solid var(--line);border-radius:13px;background:var(--surface)}.resume-section h4{margin:0 0 13px;padding-bottom:10px;border-bottom:1px solid var(--line);color:var(--text);font-size:14px}
.resume-grid{display:grid;column-gap:14px;row-gap:17px}.resume-grid.identity{grid-template-columns:1.2fr 1fr 1fr}.resume-grid.notes{grid-template-columns:1fr 1fr;margin-top:17px}.resume-grid.discharge{grid-template-columns:repeat(4,minmax(0,1fr))}.discharge-notes{margin-top:18px}.resume-code-row{position:relative;margin-top:15px}.resume-code-row:first-of-type{margin-top:0}.resume-coding-picker{position:relative}.resume-coding-picker .staff-search-results{z-index:120}.resume-coding-picker>.form-input-message{display:block;margin-top:5px}.resume-coding-picker.coding-valid .staff-search-selected{border-color:#14b8a6;box-shadow:0 0 0 2px rgba(20,184,166,.08)}.resume-coding-picker.coding-valid>.form-input-message{color:#0f766e}.resume-coding-picker.invalid .staff-search-box,.resume-coding-picker.invalid .staff-search-selected{border-color:#dc2626}
.resume-form .form-input-field,.resume-form .staff-search{gap:9px}.resume-form .form-input-label{line-height:1.35}.resume-form .staff-search-selected,.resume-form .staff-search-box{margin-top:1px}
.resume-billing-lock{margin-bottom:14px;padding:11px 13px;border:1px solid rgba(245,158,11,.3);border-radius:10px;color:#92400e;background:rgba(245,158,11,.1);font-size:11px;font-weight:700;line-height:1.55}
.resume-actions{display:flex;justify-content:flex-end;gap:9px;padding-top:2px}.resume-button{display:inline-flex;height:40px;align-items:center;justify-content:center;gap:7px;border:1px solid var(--line);border-radius:10px;padding:0 14px;color:var(--text);background:var(--surface);font-size:11px;font-weight:800;cursor:pointer}.resume-button.primary{border-color:#0d9488;color:white;background:#0d9488}.resume-button.danger{border-color:rgba(220,38,38,.22);color:#dc2626;background:rgba(220,38,38,.08)}
.resume-state{display:grid;min-height:280px;place-items:center;align-content:center;gap:10px;color:var(--muted)}.resume-dialog-backdrop{position:fixed;z-index:1000;inset:0;display:grid;place-items:center;padding:20px;background:rgba(2,6,23,.58)}.resume-dialog{width:min(430px,100%);padding:20px;border:1px solid var(--line);border-radius:16px;color:var(--text);background:var(--surface)}.resume-dialog h3{margin:0}.resume-dialog p{color:var(--muted);font-size:12px;line-height:1.6}.resume-dialog>div{display:flex;justify-content:flex-end;gap:8px}
.theme-dark .resume-card,.sirava-dark .resume-card{border-color:#304258;background:#0c1625}.theme-dark .resume-form,.sirava-dark .resume-form{background:#0c1625}.theme-dark .resume-section,.sirava-dark .resume-section{border-color:#304258;background:#142132}.theme-dark .resume-section h4,.sirava-dark .resume-section h4{border-color:#304258}
.theme-dark .resume-form .form-input-element,.theme-dark .resume-form .form-input-select.ui-select,.theme-dark .resume-form .staff-search-selected,.theme-dark .resume-form .staff-search-box,.sirava-dark .resume-form .form-input-element,.sirava-dark .resume-form .form-input-select.ui-select,.sirava-dark .resume-form .staff-search-selected,.sirava-dark .resume-form .staff-search-box{border-color:#3b4d64;color:#f8fafc;background:#1c293b;color-scheme:dark}
.theme-dark .resume-form .form-input-label,.sirava-dark .resume-form .form-input-label{color:#9fb2ca}.theme-dark .resume-form .form-input-element::placeholder,.sirava-dark .resume-form .form-input-element::placeholder{color:#71849d}
.theme-dark .resume-status.available,.sirava-dark .resume-status.available{color:#5eead4;background:rgba(20,184,166,.14)}
.theme-dark .resume-coding-picker.coding-valid>.form-input-message,.sirava-dark .resume-coding-picker.coding-valid>.form-input-message{color:#5eead4}
.theme-dark .resume-billing-lock,.sirava-dark .resume-billing-lock{color:#fcd34d;background:rgba(120,53,15,.24)}
@media(max-width:1200px){.resume-grid.identity,.resume-grid.discharge{grid-template-columns:1fr 1fr}}
@media(max-width:760px){.resume-header{align-items:flex-start;flex-direction:column}.resume-grid.identity,.resume-grid.notes,.resume-grid.discharge,.resume-code-row{grid-template-columns:1fr}}
</style>
