<script setup>
import { ChevronDown, ChevronUp, LoaderCircle, Pencil, Save, Trash2, X } from "@lucide/vue"
import PrimeColumn from "primevue/column"
import CariPetugas from "../../Components/Ui/CariPetugas.vue"
import DataTable from "../../Components/Ui/DataTable.vue"
import FormInput from "../../Components/Ui/FormInput.vue"
import TableSearch from "../../Components/Ui/TableSearch.vue"
import { useImplementasiKeperawatan } from "./ImplementasiKeperawatan/useImplementasiKeperawatan.js"

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
  keyword,
  form,
  filteredRows,
  resetForm,
  formatDate,
  loadRecords,
  editRecord,
  saveRecord,
  confirmDelete,
} = useImplementasiKeperawatan(props)
</script>

<template>
  <section class="implementasi-page clinical-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Catatan Keperawatan Rawat Inap</span>
          <h3>{{ editingKey ? 'Edit Implementasi Keperawatan' : 'Input Implementasi Keperawatan' }}</h3>
          <p>Petugas: {{ petugas.nama || '-' }}<small v-if="petugas.jabatan"> · {{ petugas.jabatan }}</small></p>
        </div>
        <div class="clinical-section-tools">
          <button v-if="editingKey" type="button" class="clinical-button secondary" @click="resetForm"><X :size="15" /> Batal Edit</button>
          <button type="button" class="clinical-button toggle icon-only" :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'" @click="formVisible = !formVisible">
            <ChevronUp v-if="formVisible" :size="16" /><ChevronDown v-else :size="16" />
          </button>
        </div>
      </header>

      <div v-if="billingLocked" class="billing-lock">Kunjungan sudah masuk billing atau dibatalkan. Data hanya dapat dilihat.</div>
      <form v-show="formVisible" class="clinical-form form-compact implementasi-form" @submit.prevent="saveRecord">
        <fieldset :disabled="saving || billingLocked">
          <div class="form-grid">
            <FormInput v-model="form.tanggal" label="Tanggal Perawatan" type="date" required />
            <FormInput v-model="form.jam" label="Jam Rawat" type="time" step="1" required />
            <CariPetugas v-model="selectedOfficer" class="petugas-field" sumber="implementasi_keperawatan" :token="token" :disabled="!canChooseOfficer" required />
          </div>
          <FormInput v-model="form.uraian" class="implementasi-uraian" label="Uraian Implementasi Keperawatan" jenis="textarea" :rows="6" maxlength="1000" required placeholder="Tuliskan tindakan dan implementasi keperawatan yang telah dilakukan..." />
          <footer class="clinical-form-actions">
            <button type="button" class="clinical-button secondary" @click="resetForm"><X :size="15" /> Batal / Reset</button>
            <button type="submit" class="clinical-button primary">
              <LoaderCircle v-if="saving" class="spin" :size="15" /><Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editingKey ? 'Simpan Perubahan' : 'Simpan Implementasi' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card implementasi-history-card">
      <header class="clinical-section-header">
        <div><span>Riwayat Keperawatan</span><h3>Implementasi Keperawatan</h3><p>{{ filteredRows.length }} dari {{ records.length }} catatan ditampilkan</p></div>
        <TableSearch v-if="records.length" v-model="keyword" placeholder="Cari tanggal, petugas, atau uraian..." :total="records.length" :filtered="filteredRows.length" />
      </header>

      <div v-if="loading" class="clinical-state"><LoaderCircle class="spin" :size="25" /><strong>Menarik implementasi keperawatan...</strong></div>
      <div v-else-if="error" class="clinical-state error"><strong>{{ error }}</strong><button type="button" @click="loadRecords">Coba Lagi</button></div>
      <div v-else-if="!records.length" class="clinical-state"><strong>Belum ada implementasi keperawatan untuk kunjungan ini.</strong></div>
      <DataTable v-else :rows="filteredRows" data-key="_key" empty-message="Implementasi keperawatan tidak ditemukan.">
        <PrimeColumn header="TANGGAL / JAM" style="min-width:150px"><template #body="{ data }"><div class="clinical-table-main"><strong>{{ formatDate(data.tanggal) }}</strong><span>{{ data.jam }}</span></div></template></PrimeColumn>
        <PrimeColumn header="PETUGAS" style="min-width:240px"><template #body="{ data }"><div class="clinical-table-main"><strong>{{ data.nama_petugas || data.nip }}</strong><span>{{ data.jabatan || data.nip }}</span></div></template></PrimeColumn>
        <PrimeColumn header="URAIAN IMPLEMENTASI" style="min-width:520px"><template #body="{ data }"><p class="clinical-table-note implementasi-table-note">{{ data.uraian || '-' }}</p></template></PrimeColumn>
        <PrimeColumn header="AKSI" frozen align-frozen="right" style="min-width:140px">
          <template #body="{ data }">
            <div v-if="data.bisa_diubah && !billingLocked" class="clinical-table-actions">
              <button type="button" @click="editRecord(data)"><Pencil :size="14" /> Edit</button>
              <button type="button" class="danger" @click="deleteTarget = data"><Trash2 :size="14" /> Hapus</button>
            </div>
            <span v-else class="clinical-owner-note">{{ billingLocked ? 'Terkunci billing' : 'Bukan milik Anda' }}</span>
          </template>
        </PrimeColumn>
      </DataTable>
    </article>

    <div v-if="deleteTarget" class="clinical-confirm-backdrop" @click.self="deleteTarget = null">
      <section class="clinical-confirm-dialog" role="dialog" aria-modal="true">
        <h3>Hapus Implementasi Keperawatan?</h3>
        <p>Catatan tanggal {{ formatDate(deleteTarget.tanggal) }} pukul {{ deleteTarget.jam }} akan dihapus.</p>
        <div><button type="button" class="clinical-button secondary" :disabled="deleting" @click="deleteTarget = null">Batal</button><button type="button" class="clinical-button danger" :disabled="deleting" @click="confirmDelete"><Trash2 :size="15" /> {{ deleting ? 'Menghapus...' : 'Hapus' }}</button></div>
      </section>
    </div>
  </section>
</template>

<style src="./ImplementasiKeperawatan/implementasi-keperawatan.css" scoped></style>
