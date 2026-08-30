<script setup>
import { ChevronDown, ChevronUp, Eye, LoaderCircle, Pencil, Save, Trash2, X } from "@lucide/vue"
import PrimeColumn from "primevue/column"
import CariPetugas from "../../Components/Ui/CariPetugas.vue"
import DataTable from "../../Components/Ui/DataTable.vue"
import FormInput from "../../Components/Ui/FormInput.vue"
import TableSearch from "../../Components/Ui/TableSearch.vue"
import scoreNyeriImage from "../../assets/simrs/score-nyeri.png"
import { useEwsRanap } from "./EwsRanap/useEwsRanap.js"

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
})

const {
  loading,
  saving,
  deleting,
  error,
  records,
  petugas,
  selectedOfficer,
  canChooseOfficer,
  billingLocked,
  editingKey,
  deleteTarget,
  formVisible,
  kataKunciRiwayat,
  pilihanAlat,
  pilihanKesadaran,
  pilihanSkalaNyeri,
  form,
  validationErrors,
  tableRecordsTampil,
  resetForm,
  loadRecords,
  editRecord,
  viewRecord,
  submitForm,
  deleteRecord,
  klasifikasi,
  scoreClass,
} = useEwsRanap(props)
</script>

<template>
  <section class="clinical-page ews-ranap-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Early Warning Score Ranap</span>
          <h3>{{ billingLocked && records.length ? 'Detail EWS Ranap' : editingKey ? 'Edit EWS Ranap' : 'Input EWS Ranap' }}</h3>
          <p>
            <template v-if="billingLocked">Kunjungan sudah masuk billing. Data hanya dapat dilihat.</template>
          </p>
        </div>
        <div class="clinical-section-tools">
          <button v-if="editingKey && !billingLocked" type="button" class="clinical-button secondary" @click="resetForm">
            <X :size="15" /> Batal Edit
          </button>
          <button
            type="button"
            class="clinical-button toggle icon-only"
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

      <form v-show="formVisible" class="clinical-form ews-form" novalidate @submit.prevent="submitForm">
        <fieldset class="form-compact" :disabled="saving || billingLocked">
          <div v-if="error" class="ews-alert">{{ error }}</div>

          <div class="clinical-time-grid ews-time-grid">
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

            <div class="clinical-vital-grid ews-vital-grid">
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

          <div class="clinical-soap-grid ews-soap-grid">
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

            <div class="clinical-vital-grid ews-fluid-grid">
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

          <footer v-if="!billingLocked" class="clinical-form-actions">
            <button type="button" class="clinical-button secondary" :disabled="saving" @click="resetForm">
              <X :size="15" /> Batal / Reset
            </button>
            <button type="submit" class="clinical-button primary" :disabled="saving">
              <LoaderCircle v-if="saving" :size="15" class="spin" />
              <Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editingKey ? 'Simpan Perubahan' : 'Simpan EWS' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card">
      <header class="clinical-section-header">
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

      <div v-if="loading" class="clinical-state">
        <LoaderCircle class="spin" :size="25" />
        <strong>Menarik data EWS Ranap...</strong>
      </div>
      <div v-else-if="error" class="clinical-state error">
        <strong>{{ error }}</strong>
        <button type="button" @click="loadRecords">Coba Lagi</button>
      </div>
      <div v-else-if="records.length === 0" class="clinical-state">
        <strong>Belum ada EWS Ranap pada kunjungan ini.</strong>
      </div>

      <DataTable v-else :rows="tableRecordsTampil" data-key="_key" empty-message="EWS Ranap tidak ditemukan.">
        <PrimeColumn header="TANGGAL / JAM" style="min-width:140px">
          <template #body="{ data }">
            <div class="clinical-table-main">
              <strong>{{ data.tanggal }}</strong>
              <span>{{ data.jam }}</span>
            </div>
          </template>
        </PrimeColumn>
        <PrimeColumn header="PETUGAS" style="min-width:210px">
          <template #body="{ data }">
            <div class="clinical-table-main">
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
            <div class="clinical-table-vitals">
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
            <div class="clinical-table-extra">
              <p><b>Respon:</b> {{ data.respon || '-' }}</p>
              <p><b>Tindakan:</b> {{ data.tindakan || '-' }}</p>
              <p><b>Frekuensi:</b> {{ data.frekuensi || '-' }}</p>
            </div>
          </template>
        </PrimeColumn>
        <PrimeColumn header="BALANCE CAIRAN" style="min-width:180px">
          <template #body="{ data }">
            <div class="clinical-table-main">
              <strong>{{ data.bc || '0' }}</strong>
              <span>Masuk {{ data.jumlahmasuk || '0' }} / Keluar {{ data.jumlahkeluar || '0' }}</span>
            </div>
          </template>
        </PrimeColumn>
        <PrimeColumn header="AKSI" frozen align-frozen="right" style="min-width:145px">
          <template #body="{ data }">
            <div class="clinical-table-actions">
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

    <div v-if="deleteTarget" class="clinical-confirm-backdrop" @click.self="deleteTarget = null">
      <section class="clinical-confirm-dialog" role="dialog" aria-modal="true">
        <h3>Hapus EWS Ranap?</h3>
        <p>Catatan tanggal {{ deleteTarget.tanggal }} pukul {{ deleteTarget.jam }} akan dihapus dari SIMRS Khanza.</p>
        <div>
          <button type="button" class="clinical-button secondary" :disabled="deleting" @click="deleteTarget = null">Batal</button>
          <button type="button" class="clinical-button danger" :disabled="deleting" @click="deleteRecord(deleteTarget)">
            <LoaderCircle v-if="deleting" class="spin" :size="15" />
            <Trash2 v-else :size="15" />
            {{ deleting ? 'Menghapus...' : 'Hapus' }}
          </button>
        </div>
      </section>
    </div>
  </section>
</template>

<style src="./EwsRanap/ews-ranap.css" scoped></style>
