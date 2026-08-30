<script setup>
import { FileCheck2, LoaderCircle, Paperclip, Save, Search, Trash2, X } from "@lucide/vue"
import CariDokter from "../../Components/Ui/CariDokter.vue"
import FormInput from "../../Components/Ui/FormInput.vue"
import InputPencarian from "../../Components/Ui/InputPencarian.vue"
import { useResumePasienRanap } from "./ResumePasienRanap/useResumePasienRanap.js"

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
})

const {
  loading,
  saving,
  deleting,
  tersedia,
  billingLocked,
  hapusTerbuka,
  doctor,
  form,
  statusCoding,
  referensiTerbuka,
  referensiLoading,
  referensiSearch,
  referensiItems,
  referensiDipilih,
  referensiSemuaTerpilih,
  pilihanSelect,
  catatanPerawatan,
  catatanPulang,
  diagnosa,
  prosedur,
  nilaiCoding,
  cariPilihanCoding,
  pilihCoding,
  bukaReferensi,
  muatReferensi,
  referensiKey,
  referensiAktif,
  toggleReferensi,
  pilihSemuaReferensi,
  tambahReferensiTerpilih,
  simpan,
  hapus,
} = useResumePasienRanap(props)
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
              <div v-for="field in catatanPerawatan" :key="field.key" class="resume-reference-field">
                <div v-if="field.references?.length" class="resume-reference-actions">
                  <button
                    v-for="reference in field.references"
                    :key="reference.jenis"
                    type="button"
                    class="resume-reference-button"
                    :disabled="saving || deleting || billingLocked"
                    @click="bukaReferensi(reference.jenis)"
                  >
                    <Paperclip :size="13" /> {{ reference.label }}
                  </button>
                </div>
                <FormInput
                  v-model="form[field.key]"
                  :label="field.label"
                  jenis="textarea"
                  :rows="3"
                  :required="field.required"
                  maxlength="2000"
                />
              </div>
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
              <div v-for="field in catatanPulang" :key="field.key" class="resume-reference-field">
                <div v-if="field.references?.length" class="resume-reference-actions">
                  <button
                    v-for="reference in field.references"
                    :key="reference.jenis"
                    type="button"
                    class="resume-reference-button"
                    :disabled="saving || deleting || billingLocked"
                    @click="bukaReferensi(reference.jenis)"
                  >
                    <Paperclip :size="13" /> {{ reference.label }}
                  </button>
                </div>
                <FormInput
                  v-model="form[field.key]"
                  :label="field.label"
                  jenis="textarea"
                  :rows="3"
                  maxlength="2000"
                />
              </div>
            </div>
          </section>

          <footer class="resume-actions">
            <button v-if="tersedia && !billingLocked" type="button" class="resume-button danger" @click="hapusTerbuka = true"><Trash2 :size="15" /> Hapus</button>
            <button v-if="!billingLocked" type="submit" class="resume-button primary"><LoaderCircle v-if="saving" class="spin" :size="15" /><Save v-else :size="15" />{{ saving ? 'Menyimpan...' : tersedia ? 'Simpan Perubahan' : 'Simpan Resume' }}</button>
          </footer>
        </fieldset>
      </form>
    </article>

    <div v-if="referensiTerbuka" class="resume-dialog-backdrop" @click.self="referensiTerbuka = null">
      <section class="resume-dialog resume-reference-dialog">
        <header>
          <div>
            <span>REFERENSI RESUME RANAP</span>
            <h3>{{ referensiTerbuka.judul }}</h3>
            <p>Data diambil dari pelayanan SIMRS Khanza pada nomor rawat ini.</p>
          </div>
          <button type="button" class="resume-button secondary" @click="referensiTerbuka = null">
            <X :size="15" />
          </button>
        </header>

        <div class="resume-reference-search">
          <Search :size="17" />
          <input
            v-model="referensiSearch"
            type="search"
            placeholder="Cari tanggal atau isi referensi..."
            @keyup.enter="muatReferensi"
          >
          <button type="button" class="resume-button primary" @click="muatReferensi">Cari</button>
        </div>

        <div class="resume-reference-toolbar">
          <span>{{ referensiDipilih.length }} dari {{ referensiItems.length }} referensi dipilih</span>
          <div>
            <button
              type="button"
              class="resume-button secondary"
              :disabled="!referensiItems.length"
              @click="pilihSemuaReferensi"
            >
              {{ referensiSemuaTerpilih ? 'Batal Pilih Semua' : 'Pilih Semua' }}
            </button>
            <button
              type="button"
              class="resume-button primary"
              :disabled="!referensiDipilih.length"
              @click="tambahReferensiTerpilih"
            >
              Tambahkan Terpilih
            </button>
          </div>
        </div>

        <div class="resume-reference-list">
          <div v-if="referensiLoading" class="resume-reference-state">
            <LoaderCircle class="spin" :size="22" />
            Menarik referensi...
          </div>
          <template v-else>
            <button
              v-for="item in referensiItems"
              :key="referensiKey(item)"
              type="button"
              class="resume-reference-row"
              :class="{ active: referensiAktif(item) }"
              @click="toggleReferensi(item)"
            >
              <span class="resume-reference-check">✓</span>
              <span>
                <strong>{{ item.isi }}</strong>
                <small>{{ item.tanggal }} {{ item.jam }} · {{ item.sumber }}</small>
              </span>
            </button>
          </template>
          <div v-if="!referensiLoading && !referensiItems.length" class="resume-reference-state">
            Referensi belum ditemukan untuk nomor rawat ini.
          </div>
        </div>
      </section>
    </div>

    <div v-if="hapusTerbuka" class="resume-dialog-backdrop" @click.self="hapusTerbuka = false">
      <section class="resume-dialog">
        <h3>Hapus Resume Pasien?</h3>
        <p>Resume untuk nomor rawat {{ patient.no_rawat }} akan dihapus dari SIMRS Khanza.</p>
        <div><button type="button" class="resume-button secondary" @click="hapusTerbuka = false"><X :size="15" /> Batal</button><button type="button" class="resume-button danger" @click="hapus"><LoaderCircle v-if="deleting" class="spin" :size="15" /><Trash2 v-else :size="15" /> Hapus</button></div>
      </section>
    </div>
  </section>
</template>

<style src="./ResumePasienRanap/resume-pasien-ranap.css"></style>
