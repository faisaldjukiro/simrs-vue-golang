<script setup lang="ts">
// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { ChevronDown, ChevronUp, LoaderCircle, Pencil, Save, Trash2, X } from "@lucide/vue"
import Column from "primevue/column"
import CariDokter from "../../../Components/Ui/CariDokter.vue"
import CariTindakanLaboratorium from "../../../Components/Ui/CariTindakanLaboratorium.vue"
import DataTable from "../../../Components/Ui/DataTable.vue"
import FormInput from "../../../Components/Ui/FormInput.vue"
import TableSearch from "../../../Components/Ui/TableSearch.vue"
import { usePermintaanLaboratorium } from "./usePermintaanLaboratorium"

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
})

const {
  loading,
  saving,
  deleting,
  error,
  formVisible,
  requests,
  doctor,
  treatments,
  billingLocked,
  scope,
  deleteTarget,
  editingNumber,
  kataKunciPermintaan,
  form,
  tableRowsTampil,
  rupiah,
  formatDate,
  waktuStatus,
  kelasStatusPemeriksaan,
  kelasStatusBayar,
  resetForm,
  editRequest,
  loadData,
  saveRequest,
  confirmDelete,
} = usePermintaanLaboratorium(props)
</script>

<template>
  <section class="clinical-page radiology-request-page laboratory-request-page">
    <article class="clinical-form-card radiology-form-card laboratory-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Penunjang Medis</span>
          <h3>{{ editingNumber ? 'Edit Permintaan Laboratorium' : 'Input Permintaan Laboratorium' }}</h3>
          <p>Pemeriksaan dapat dipilih lebih dari satu dalam satu nomor permintaan.</p>
        </div>
        <div class="clinical-section-tools">
          <button v-if="editingNumber" type="button" class="clinical-button secondary" @click="resetForm"><X :size="15" /> Batal Edit</button>
          <button type="button" class="clinical-button toggle icon-only" :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'" @click="formVisible = !formVisible">
            <ChevronUp v-if="formVisible" :size="15" />
            <ChevronDown v-else :size="15" />
          </button>
        </div>
      </header>

      <div v-if="billingLocked" class="handling-lock">Billing sudah terverifikasi atau kunjungan dibatalkan. Permintaan baru tidak dapat dibuat.</div>

      <form v-show="formVisible" class="clinical-form radiology-form" @submit.prevent="saveRequest">
        <fieldset class="form-compact" :disabled="saving || billingLocked">
          <div class="radiology-form-main">
            <FormInput v-model="form.tanggal" label="Tanggal Permintaan" type="date" required />
            <FormInput v-model="form.jam" label="Jam Permintaan" type="time" step="1" required />
            <CariDokter v-model="doctor" class="radiology-referrer" :token="token" sumber="laboratorium" required />
          </div>
          <div class="radiology-notes-grid">
            <FormInput v-model="form.informasi_tambahan" label="Informasi Tambahan" jenis="textarea" :rows="2" maxlength="60" placeholder="Informasi tambahan untuk petugas laboratorium" required />
            <FormInput v-model="form.diagnosis_klinis" label="Diagnosis Klinis" jenis="textarea" :rows="2" maxlength="80" placeholder="Diagnosis atau alasan klinis pemeriksaan" required />
          </div>
          <CariTindakanLaboratorium v-model="treatments" :token="token" :no-rawat="patient.no_rawat" required />
          <div class="radiology-filter-note">
            <span>{{ scope.status === 'ranap' ? 'Rawat Inap' : 'Rawat Jalan' }}</span>
            <span>Cara bayar tarif: {{ scope.kodeCaraBayar || '-' }}</span>
            <span v-if="scope.kelas">Kelas pasien: {{ scope.kelas }}</span>
          </div>
          <footer class="clinical-form-actions">
            <button type="button" class="clinical-button secondary" @click="resetForm"><X :size="15" /> Batal / Reset</button>
            <button type="submit" class="clinical-button primary">
              <LoaderCircle v-if="saving" class="spin" :size="15" /><Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editingNumber ? 'Simpan Perubahan' : `Simpan ${treatments.length || ''} Pemeriksaan` }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card radiology-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Riwayat Pasien</span>
          <h3>Permintaan Laboratorium</h3>
          <p>
            <template v-if="kataKunciPermintaan">{{ tableRowsTampil.length }} dari {{ requests.length }} permintaan ditampilkan</template>
            <template v-else>{{ requests.length }} permintaan ditemukan</template>
          </p>
        </div>
        <TableSearch
          v-if="requests.length > 0"
          v-model="kataKunciPermintaan"
          placeholder="Cari permintaan, dokter, atau pemeriksaan..."
          :total="requests.length"
          :filtered="tableRowsTampil.length"
        />
      </header>
      <div v-if="loading" class="clinical-state"><LoaderCircle class="spin" :size="25" /><strong>Menarik permintaan laboratorium...</strong></div>
      <div v-else-if="error" class="clinical-state error"><strong>{{ error }}</strong><button type="button" @click="loadData">Coba Lagi</button></div>
      <div v-else-if="requests.length === 0" class="clinical-state"><strong>Belum ada permintaan laboratorium.</strong></div>
      <DataTable v-else :rows="tableRowsTampil" data-key="_key" empty-message="Permintaan laboratorium tidak ditemukan.">
        <Column header="NO. PERMINTAAN" style="min-width:175px"><template #body="{ data }"><div class="clinical-table-main"><strong>{{ data.nomor }}</strong><span>{{ formatDate(data.tanggal) }} · {{ data.jam }}</span></div></template></Column>
        <Column header="DOKTER PERUJUK" style="min-width:230px"><template #body="{ data }"><div class="clinical-table-main"><strong>{{ data.nama_dokter }}</strong><span>{{ data.kode_dokter }}</span></div></template></Column>
        <Column header="INFORMASI / DIAGNOSIS" style="min-width:300px"><template #body="{ data }"><div class="radiology-clinical"><p><b>Informasi:</b> {{ data.informasi_tambahan }}</p><p><b>Diagnosis:</b> {{ data.diagnosis_klinis }}</p></div></template></Column>
        <Column header="PEMERIKSAAN LABORATORIUM" style="min-width:410px">
          <template #body="{ data }"><div class="radiology-exam-list"><article v-for="item in data.pemeriksaan" :key="item.kode"><span><strong>{{ item.nama }}</strong><small>{{ item.kode }} · {{ item.kelas }} · {{ item.detail?.length || 0 }} detail</small><small v-if="item.detail?.length" class="laboratory-detail-summary">{{ item.detail.map((detail) => detail.nama).join(', ') }}</small></span><b>{{ rupiah(item.total) }}</b></article></div></template>
        </Column>
        <Column header="TOTAL" style="min-width:130px"><template #body="{ data }"><strong>{{ rupiah(data.total) }}</strong></template></Column>
        <Column header="STATUS PEMERIKSAAN" style="min-width:180px"><template #body="{ data }"><div class="radiology-status-cell"><span class="radiology-status-badge" :class="kelasStatusPemeriksaan(data.status_pemeriksaan)">{{ data.status_pemeriksaan }}</span><small>{{ waktuStatus(data) }}</small></div></template></Column>
        <Column header="STATUS BAYAR" style="min-width:145px"><template #body="{ data }"><span class="radiology-status-badge" :class="kelasStatusBayar(data.status_bayar)">{{ data.status_bayar }}</span></template></Column>
        <Column header="AKSI" frozen align-frozen="right" style="min-width:145px">
          <template #body="{ data }">
            <div v-if="!billingLocked && (data.dapat_diubah || data.dapat_dihapus)" class="clinical-table-actions">
              <button v-if="data.dapat_diubah" type="button" @click="editRequest(data)"><Pencil :size="14" /> Edit</button>
              <button v-if="data.dapat_dihapus" type="button" class="danger" @click="deleteTarget = data"><Trash2 :size="14" /> Hapus</button>
            </div>
            <span v-else class="radiology-locked">Terkunci</span>
          </template>
        </Column>
      </DataTable>
    </article>

    <div v-if="deleteTarget" class="clinical-confirm-backdrop" @click.self="deleteTarget = null">
      <section class="clinical-confirm-dialog">
        <h3>Hapus Permintaan Laboratorium?</h3>
        <p>Permintaan {{ deleteTarget.nomor }} beserta {{ deleteTarget.pemeriksaan.length }} pemeriksaan akan dihapus.</p>
        <div><button type="button" class="clinical-button secondary" @click="deleteTarget = null">Batal</button><button type="button" class="clinical-button danger" :disabled="deleting" @click="confirmDelete"><LoaderCircle v-if="deleting" class="spin" :size="15" /><Trash2 v-else :size="15" />{{ deleting ? 'Menghapus...' : 'Hapus' }}</button></div>
      </section>
    </div>
  </section>
</template>
