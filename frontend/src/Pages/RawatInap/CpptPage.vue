<script setup>
import { ChevronDown, ChevronUp, LoaderCircle, Pencil, Save, Trash2, X } from "@lucide/vue"
import PrimeColumn from "primevue/column"
import DataTable from "../../Components/Ui/DataTable.vue"
import CariPetugas from "../../Components/Ui/CariPetugas.vue"
import FormInput from "../../Components/Ui/FormInput.vue"
import TableSearch from "../../Components/Ui/TableSearch.vue"
import { useCppt } from "./Cppt/useCppt.js"

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
  jenisRawat: { type: String, default: 'ranap' },
  namaModul: { type: String, default: 'Rawat Inap' },
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
  awarenessOptions,
  editingKey,
  deleteTarget,
  formVisible,
  kataKunciCatatan,
  judulCatatan,
  form,
  tableRecordsTampil,
  resetForm,
  loadRecords,
  editRecord,
  saveRecord,
  confirmDelete,
  formatDate,
} = useCppt(props)
</script>

<template>
  <section class="clinical-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Catatan Perkembangan Pasien Terintegrasi</span>
          <h3>{{ editingKey ? `Edit ${judulCatatan}` : `Input ${judulCatatan}` }}</h3>
          <p>Petugas: {{ petugas.nama || '-' }} <small v-if="petugas.jabatan">· {{ petugas.jabatan }}</small></p>
        </div>
        <div class="clinical-section-tools">
          <button v-if="editingKey" type="button" class="clinical-button secondary" @click="resetForm">
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

      <div v-if="billingLocked" class="handling-lock">Billing sudah terverifikasi atau kunjungan dibatalkan. Form {{ judulCatatan }} hanya dapat dilihat.</div>
      <form v-show="formVisible" class="clinical-form" @submit.prevent="saveRecord">
        <fieldset class="form-compact" :disabled="saving || billingLocked">
          <div class="clinical-time-grid">
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

          <div class="clinical-vital-grid">
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

          <div class="clinical-soap-grid">
            <FormInput v-model="form.subjek" label="S — Subjek" jenis="textarea" :rows="3" maxlength="2000" placeholder="Keluhan dan informasi subjektif pasien" />
            <FormInput v-model="form.objek" label="O — Objek" jenis="textarea" :rows="3" maxlength="2000" placeholder="Hasil pemeriksaan objektif" />
            <FormInput v-model="form.asesmen" label="A — Asesmen" jenis="textarea" :rows="3" maxlength="2000" placeholder="Penilaian klinis" />
            <FormInput v-model="form.plan" label="P — Plan" jenis="textarea" :rows="3" maxlength="2000" placeholder="Rencana tindak lanjut" />
            <FormInput v-model="form.instruksi" label="Instruksi" jenis="textarea" :rows="2" maxlength="2000" placeholder="Instruksi pelayanan" />
            <FormInput v-model="form.evaluasi" label="Evaluasi" jenis="textarea" :rows="2" maxlength="2000" placeholder="Evaluasi perkembangan pasien" />
          </div>

          <footer class="clinical-form-actions">
            <button type="button" class="clinical-button secondary" @click="resetForm"><X :size="15" /> Batal / Reset</button>
            <button type="submit" class="clinical-button primary">
              <LoaderCircle v-if="saving" class="spin" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editingKey ? 'Simpan Perubahan' : `Simpan ${judulCatatan}` }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Riwayat Pasien</span>
          <h3>Catatan {{ judulCatatan }}</h3>
          <p>
            <template v-if="kataKunciCatatan">{{ tableRecordsTampil.length }} dari {{ records.length }} catatan ditampilkan</template>
            <template v-else>{{ records.length }} catatan ditemukan</template>
          </p>
        </div>
        <TableSearch
          v-if="records.length > 0"
          v-model="kataKunciCatatan"
          :placeholder="`Cari catatan ${judulCatatan}...`"
          :total="records.length"
          :filtered="tableRecordsTampil.length"
        />
      </header>

      <div v-if="loading" class="clinical-state"><LoaderCircle class="spin" :size="25" /><strong>Menarik catatan CPPT/SOAP...</strong></div>
      <div v-else-if="error" class="clinical-state error"><strong>{{ error }}</strong><button type="button" @click="loadRecords">Coba Lagi</button></div>
      <div v-else-if="records.length === 0" class="clinical-state"><strong>Belum ada catatan CPPT/SOAP untuk pasien ini.</strong></div>

      <DataTable v-else :rows="tableRecordsTampil" data-key="_key" empty-message="Catatan CPPT/SOAP tidak ditemukan.">
        <PrimeColumn header="TANGGAL / JAM" style="min-width:140px">
          <template #body="{ data }"><div class="clinical-table-main"><strong>{{ formatDate(data.tgl_perawatan) }}</strong><span>{{ data.jam_rawat }}</span></div></template>
        </PrimeColumn>
        <PrimeColumn header="PETUGAS" style="min-width:190px">
          <template #body="{ data }"><div class="clinical-table-main"><strong>{{ data.nama_petugas || data.nip }}</strong><span>{{ data.jabatan || data.nip }}</span></div></template>
        </PrimeColumn>
        <PrimeColumn header="TANDA VITAL" style="min-width:220px">
          <template #body="{ data }">
            <div class="clinical-table-vitals">
              <span>Kesadaran <b>{{ data.kesadaran || '-' }}</b></span><span>Suhu <b>{{ data.suhu_tubuh || '-' }}</b></span>
              <span>Tensi <b>{{ data.tensi || '-' }}</b></span><span>Nadi <b>{{ data.nadi || '-' }}</b></span>
              <span>RR <b>{{ data.respirasi || '-' }}</b></span><span>SpO2 <b>{{ data.spo2 || '-' }}</b></span>
              <span>GCS <b>{{ data.gcs || '-' }}</b></span><span>TB/BB <b>{{ data.tinggi || '-' }}/{{ data.berat || '-' }}</b></span>
              <span v-if="jenisRawat === 'ralan'">Lingkar Perut <b>{{ data.lingkar_perut || '-' }}</b></span>
            </div>
          </template>
        </PrimeColumn>
        <PrimeColumn header="SUBJEK" style="min-width:230px"><template #body="{ data }"><p class="clinical-table-note">{{ data.subjek || '-' }}</p></template></PrimeColumn>
        <PrimeColumn header="OBJEK" style="min-width:230px"><template #body="{ data }"><p class="clinical-table-note">{{ data.objek || '-' }}</p></template></PrimeColumn>
        <PrimeColumn header="ASESMEN" style="min-width:230px"><template #body="{ data }"><p class="clinical-table-note">{{ data.asesmen || '-' }}</p></template></PrimeColumn>
        <PrimeColumn header="PLAN" style="min-width:230px"><template #body="{ data }"><p class="clinical-table-note">{{ data.plan || '-' }}</p></template></PrimeColumn>
        <PrimeColumn header="INSTRUKSI / EVALUASI" style="min-width:250px">
          <template #body="{ data }"><div class="clinical-table-extra"><p><b>Instruksi:</b> {{ data.instruksi || '-' }}</p><p><b>Evaluasi:</b> {{ data.evaluasi || '-' }}</p><p><b>Alergi:</b> {{ data.alergi || '-' }}</p></div></template>
        </PrimeColumn>
        <PrimeColumn header="AKSI" frozen align-frozen="right" style="min-width:135px">
          <template #body="{ data }">
            <div v-if="data.bisa_diubah && !billingLocked" class="clinical-table-actions">
              <button type="button" title="Edit catatan" @click="editRecord(data)"><Pencil :size="14" /> Edit</button>
              <button type="button" class="danger" title="Hapus catatan" @click="deleteTarget = data"><Trash2 :size="14" /> Hapus</button>
            </div>
            <span v-else class="clinical-owner-note">{{ billingLocked ? 'Terkunci billing' : 'Bukan milik Anda' }}</span>
          </template>
        </PrimeColumn>
      </DataTable>
    </article>

    <div v-if="deleteTarget" class="clinical-confirm-backdrop" @click.self="deleteTarget = null">
      <section class="clinical-confirm-dialog" role="dialog" aria-modal="true">
        <h3>Hapus Catatan CPPT?</h3>
        <p>Catatan tanggal {{ formatDate(deleteTarget.tgl_perawatan) }} pukul {{ deleteTarget.jam_rawat }} akan dihapus dari SIMRS Khanza.</p>
        <div>
          <button type="button" class="clinical-button secondary" :disabled="deleting" @click="deleteTarget = null">Batal</button>
          <button type="button" class="clinical-button danger" :disabled="deleting" @click="confirmDelete">
            <LoaderCircle v-if="deleting" class="spin" :size="15" /><Trash2 v-else :size="15" />
            {{ deleting ? 'Menghapus...' : 'Hapus' }}
          </button>
        </div>
      </section>
    </div>
  </section>
</template>
