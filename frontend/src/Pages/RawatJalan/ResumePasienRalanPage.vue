<script setup>
import { FileCheck2, LoaderCircle, Save, Trash2, X } from '@lucide/vue'
import { computed, reactive, ref, watch } from 'vue'
import CariDokter from '../../Components/Ui/CariDokter.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import InputPencarian from '../../Components/Ui/InputPencarian.vue'
import {
  cariCodingResumePasien,
  hapusResumePasien,
  resumePasienData,
  simpanResumePasien,
  ubahResumePasien,
  validasiCodingResumePasien,
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
const pilihan = reactive({ kondisi_pulang: ['Hidup', 'Meninggal'] })
const form = reactive(formKosong())
const statusCoding = reactive({})

const pilihanSelect = computed(() => ({
  kondisi_pulang: pilihan.kondisi_pulang.map((value) => ({ label: value, value })),
}))

const catatanResume = [
  { key: 'keluhan_utama', label: 'Keluhan Utama / Riwayat Penyakit', required: true },
  { key: 'jalannya_penyakit', label: 'Jalannya Penyakit', required: true },
  { key: 'pemeriksaan_penunjang', label: 'Pemeriksaan Penunjang' },
  { key: 'hasil_laborat', label: 'Hasil Laboratorium' },
  { key: 'obat_pulang', label: 'Obat Pulang' },
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

function formKosong() {
  return {
    no_rawat: props.patient?.no_rawat || '',
    kode_dokter: '',
    kondisi_pulang: 'Hidup',
    keluhan_utama: '',
    jalannya_penyakit: '',
    pemeriksaan_penunjang: '',
    hasil_laborat: '',
    diagnosa_utama: '',
    kd_diagnosa_utama: '',
    diagnosa_sekunder: '',
    kd_diagnosa_sekunder: '',
    diagnosa_sekunder2: '',
    kd_diagnosa_sekunder2: '',
    diagnosa_sekunder3: '',
    kd_diagnosa_sekunder3: '',
    diagnosa_sekunder4: '',
    kd_diagnosa_sekunder4: '',
    prosedur_utama: '',
    kd_prosedur_utama: '',
    prosedur_sekunder: '',
    kd_prosedur_sekunder: '',
    prosedur_sekunder2: '',
    kd_prosedur_sekunder2: '',
    prosedur_sekunder3: '',
    kd_prosedur_sekunder3: '',
    obat_pulang: '',
  }
}

function resetStatusCoding() {
  Object.keys(statusCoding).forEach((key) => delete statusCoding[key])
}

function dariServer(value = {}) {
  resetStatusCoding()
  Object.assign(form, formKosong(), value, { no_rawat: props.patient.no_rawat })
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
    const data = await validasiCodingResumePasien(props.token, {
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
    statusCoding[item.kode] = { valid: true, hint: data.pesan || 'Kode valid', im: data.im === '1' }
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
  return cariCodingResumePasien(props.token, {
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
  statusCoding[item.kode] = { valid: true, hint: pilihanCoding.pesan || 'Kode valid', im: pilihanCoding.im === '1' }
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
    const data = await resumePasienData(props.token, props.patient.no_rawat)
    tersedia.value = Boolean(data?.tersedia)
    billingLocked.value = Boolean(data?.billing_terkunci)
    Object.assign(pilihan, data?.pilihan || {})
    dariServer(data?.resume)
    doctor.value = data?.dokter || {}
  } catch (error) {
    notifikasi.gagal(error.message || 'Resume pasien rawat jalan tidak dapat dibaca.')
  } finally {
    loading.value = false
  }
}

function payload() {
  return {
    ...form,
    no_rawat: props.patient.no_rawat,
    kode_dokter: doctor.value?.kode || '',
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
      ? await ubahResumePasien(props.token, payload())
      : await simpanResumePasien(props.token, payload())
    notifikasi.sukses(response?.pesan || 'Resume pasien rawat jalan berhasil disimpan.')
    await muat()
  } catch (error) {
    notifikasi.gagal(error.message || 'Resume pasien rawat jalan gagal disimpan.')
  } finally {
    saving.value = false
  }
}

async function hapus() {
  deleting.value = true
  try {
    const response = await hapusResumePasien(props.token, props.patient.no_rawat)
    notifikasi.sukses(response?.pesan || 'Resume pasien rawat jalan berhasil dihapus.')
    hapusTerbuka.value = false
    await muat()
  } catch (error) {
    notifikasi.gagal(error.message || 'Resume pasien rawat jalan gagal dihapus.')
  } finally {
    deleting.value = false
  }
}

watch(() => props.patient.no_rawat, muat, { immediate: true })
</script>

<template>
  <section class="resume-ranap-page resume-ralan-page">
    <article class="resume-card">
      <header class="resume-header">
        <div>
          <span>REKAM MEDIS RAWAT JALAN</span>
          <h3>Resume Pasien Rawat Jalan</h3>
        </div>
        <div class="resume-status" :class="{ available: tersedia }">
          <FileCheck2 :size="17" /> {{ tersedia ? 'Resume Tersimpan' : 'Resume Baru' }}
        </div>
      </header>

      <div v-if="loading" class="resume-state">
        <LoaderCircle class="spin" :size="27" />
        <strong>Menarik data resume pasien...</strong>
      </div>

      <form v-else class="resume-form" @submit.prevent="simpan">
        <div v-if="billingLocked" class="resume-billing-lock">
          Billing sudah terverifikasi atau kunjungan dibatalkan. Resume hanya dapat dilihat dan tidak bisa ditambah, diedit, atau dihapus.
        </div>
        <fieldset :disabled="saving || deleting || billingLocked">
          <section class="resume-section">
            <h4>Informasi Resume</h4>
            <div class="resume-grid identity">
              <CariDokter v-model="doctor" :token="token" required />
              <FormInput
                v-model="form.kondisi_pulang"
                label="Kondisi Pasien Pulang"
                jenis="select"
                :options="pilihanSelect.kondisi_pulang"
                required
              />
            </div>
            <div class="resume-grid notes">
              <FormInput
                v-for="field in catatanResume"
                :key="field.key"
                v-model="form[field.key]"
                :label="field.label"
                jenis="textarea"
                :rows="3"
                :required="field.required"
                maxlength="2000"
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

          <footer class="resume-actions">
            <button v-if="tersedia && !billingLocked" type="button" class="resume-button danger" @click="hapusTerbuka = true">
              <Trash2 :size="15" /> Hapus
            </button>
            <button v-if="!billingLocked" type="submit" class="resume-button primary">
              <LoaderCircle v-if="saving" class="spin" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : tersedia ? 'Simpan Perubahan' : 'Simpan Resume' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <div v-if="hapusTerbuka" class="resume-dialog-backdrop" @click.self="hapusTerbuka = false">
      <section class="resume-dialog">
        <h3>Hapus Resume Pasien?</h3>
        <p>Resume untuk nomor rawat {{ patient.no_rawat }} akan dihapus dari SIMRS Khanza.</p>
        <div>
          <button type="button" class="resume-button secondary" @click="hapusTerbuka = false">
            <X :size="15" /> Batal
          </button>
          <button type="button" class="resume-button danger" @click="hapus">
            <LoaderCircle v-if="deleting" class="spin" :size="15" />
            <Trash2 v-else :size="15" /> Hapus
          </button>
        </div>
      </section>
    </div>
  </section>
</template>
