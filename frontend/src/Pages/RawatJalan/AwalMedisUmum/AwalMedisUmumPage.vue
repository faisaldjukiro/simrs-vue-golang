<script setup lang="ts">
import { ChevronDown, ChevronUp, LoaderCircle, RotateCcw, Save, Trash2 } from "@lucide/vue"
import Dialog from "primevue/dialog"
import CariDokter from "../../../Components/Ui/CariDokter.vue"
import FormInput from "../../../Components/Ui/FormInput.vue"
import { useAwalMedisUmum } from "./useAwalMedisUmum"

const props = defineProps({ token: { type: String, required: true }, patient: { type: Object, required: true } })

const {
  loading,
  saving,
  deleting,
  formOpen,
  confirmDelete,
  dokter,
  form,
  editing,
  billingLocked,
  pilihan,
  isiForm,
  save,
  remove,
} = useAwalMedisUmum(props)
</script>

<template>
  <section class="medical-page">
    <div v-if="loading" class="medical-state"><LoaderCircle class="spin" :size="28"/><strong>Menarik penilaian awal medis umum...</strong></div>
    <form v-else class="medical-form clinical-form-card clinical-form" @submit.prevent="save">
      <header class="clinical-section-header">
        <div><span>PENILAIAN MEDIS RAWAT JALAN</span><h3>{{ editing ? 'Edit' : 'Input' }} Awal Medis Umum</h3></div>
        <div class="header-actions">
          <button v-if="editing && !billingLocked" type="button" class="clinical-button danger" :disabled="deleting" @click="confirmDelete = true"><Trash2 :size="15"/>Hapus</button>
          <button type="button" class="clinical-button primary icon-only" :title="formOpen ? 'Sembunyikan form input' : 'Tampilkan form input'" @click="formOpen = !formOpen"><ChevronUp v-if="formOpen" :size="18"/><ChevronDown v-else :size="18"/></button>
        </div>
      </header>

      <div v-if="billingLocked" class="medical-lock">Billing sudah terverifikasi atau kunjungan dibatalkan. Form Awal Medis Umum hanya dapat dilihat.</div>

      <fieldset v-show="formOpen" :disabled="saving || billingLocked">
        <div class="medical-sheet">
          <section class="medical-section meta-section">
            <div class="medical-grid meta-grid">
              <FormInput v-model="form.tanggal" label="Tanggal Penilaian" type="datetime-local" step="1" required/>
              <CariDokter v-model="dokter" class="doctor-field" :token="token" sumber="awal-medis-umum" required/>
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
              <FormInput v-model="form.gcs" label="GCS" maxlength="10"/>
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
              <FormInput v-model="form.gigi" label="Gigi" jenis="select" :options="pilihan.pemeriksaan"/>
              <FormInput v-model="form.tht" label="THT" jenis="select" :options="pilihan.pemeriksaan"/>
              <FormInput v-model="form.thoraks" label="Thoraks" jenis="select" :options="pilihan.pemeriksaan"/>
              <FormInput v-model="form.abdomen" label="Abdomen" jenis="select" :options="pilihan.pemeriksaan"/>
              <FormInput v-model="form.genital" label="Genital" jenis="select" :options="pilihan.pemeriksaan"/>
              <FormInput v-model="form.ekstremitas" label="Ekstremitas" jenis="select" :options="pilihan.pemeriksaan"/>
              <FormInput v-model="form.kulit" label="Kulit" jenis="select" :options="pilihan.pemeriksaan"/>
              <FormInput v-model="form.ket_fisik" class="wide" label="Keterangan Pemeriksaan Fisik" jenis="textarea" :rows="4" maxlength="5000"/>
            </div>
          </section>

          <section class="medical-section">
            <h4>III. STATUS LOKALIS</h4>
            <div class="lokalis-container">
              <img src="/img/lokalis.png" alt="Status Lokalis" class="lokalis-image"/>
            </div>
            <FormInput v-model="form.ket_lokalis" label="Keterangan Status Lokalis" jenis="textarea" :rows="5" maxlength="3000" placeholder="Tuliskan lokasi dan temuan pemeriksaan setempat"/>
          </section>

          <section class="medical-section conclusion-section">
            <h4>Kesimpulan & Rencana Pelayanan</h4>
            <div class="medical-grid conclusion-grid">
              <FormInput v-model="form.penunjang" label="IV. PEMERIKSAAN PENUNJANG" jenis="textarea" :rows="4" maxlength="3000"/>
              <FormInput v-model="form.diagnosis" label="V. DIAGNOSIS / ASESMEN" jenis="textarea" :rows="4" maxlength="500"/>
              <FormInput v-model="form.tata" label="VI. TATALAKSANA" jenis="textarea" :rows="5" maxlength="5000"/>
              <FormInput v-model="form.konsulrujuk" label="VII. KONSUL / RUJUK" jenis="textarea" :rows="5" maxlength="1000"/>
            </div>
          </section>
        </div>

        <footer class="clinical-form-actions">
          <button type="button" class="clinical-button secondary" @click="isiForm"><RotateCcw :size="15"/>Batal / Reset</button>
          <button type="submit" class="clinical-button primary" :disabled="saving"><LoaderCircle v-if="saving" class="spin" :size="15"/><Save v-else :size="15"/>{{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan Penilaian' }}</button>
        </footer>
      </fieldset>
    </form>

    <Dialog v-model:visible="confirmDelete" modal header="Hapus Penilaian Awal Medis Umum" class="medical-dialog" :style="{ width: 'min(440px, 92vw)' }">
      <p>Penilaian pasien ini akan dihapus dari SIMRS Khanza.</p>
      <template #footer><button type="button" class="clinical-button secondary" @click="confirmDelete = false">Batal</button><button type="button" class="clinical-button danger" :disabled="deleting" @click="remove"><LoaderCircle v-if="deleting" class="spin" :size="14"/><Trash2 v-else :size="14"/>{{ deleting ? 'Menghapus...' : 'Hapus' }}</button></template>
    </Dialog>
  </section>
</template>

<style src="./awal-medis-umum.css" scoped></style>
