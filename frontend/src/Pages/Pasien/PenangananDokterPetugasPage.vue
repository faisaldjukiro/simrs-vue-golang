<script setup>
import { ChevronDown, ChevronUp, LoaderCircle, Pencil, Save, Trash2, X } from "@lucide/vue"
import Column from "primevue/column"
import CariDokter from "../../Components/Ui/CariDokter.vue"
import CariPetugas from "../../Components/Ui/CariPetugas.vue"
import CariTindakanRanap from "../../Components/Ui/CariTindakanRanap.vue"
import DataTable from "../../Components/Ui/DataTable.vue"
import FormInput from "../../Components/Ui/FormInput.vue"
import TableSearch from "../../Components/Ui/TableSearch.vue"
import { usePenangananDokterPetugas } from "./PenangananDokterPetugas/usePenangananDokterPetugas.js"

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
  billingLocked,
  formVisible,
  editingKey,
  deleteTarget,
  kataKunciTindakan,
  doctor,
  officer,
  treatments,
  form,
  labelJenisRawat,
  tableRecordsTampil,
  rupiah,
  formatDate,
  resetForm,
  loadRecords,
  editRecord,
  saveRecord,
  confirmDelete,
} = usePenangananDokterPetugas(props)
</script>

<template>
  <section class="clinical-page handling-page">
    <article class="clinical-form-card handling-form-card">
      <header class="clinical-section-header">
        <div><span>Tindakan {{ labelJenisRawat }}</span><h3>{{ editingKey ? 'Edit Penanganan Dokter & Petugas' : 'Input Penanganan Dokter & Petugas' }}</h3><p>Hanya tindakan gabungan dokter dan petugas.</p></div>
        <div class="clinical-section-tools">
          <button v-if="editingKey" type="button" class="clinical-button secondary" @click="resetForm(false)"><X :size="15" /> Batal Edit</button>
          <button type="button" class="clinical-button toggle icon-only" :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'" @click="formVisible = !formVisible"><ChevronUp v-if="formVisible" :size="15" /><ChevronDown v-else :size="15" /></button>
        </div>
      </header>
      <div v-if="billingLocked" class="handling-lock">Billing sudah terverifikasi atau kunjungan dibatalkan. Data hanya dapat dilihat.</div>
      <form v-show="formVisible" class="clinical-form handling-form" @submit.prevent="saveRecord">
        <fieldset class="form-compact" :disabled="saving || billingLocked">
          <div class="handling-main-grid">
            <FormInput v-model="form.tanggal" label="Tanggal Perawatan" type="date" required />
            <FormInput v-model="form.jam" label="Jam Rawat" type="time" step="1" required />
            <CariDokter v-model="doctor" :token="token" required />
            <CariPetugas v-model="officer" :token="token" sumber="penanganan" required />
            <CariTindakanRanap v-model="treatments" :token="token" :no-rawat="patient.no_rawat" :jenis-rawat="jenisRawat" :multiple="!editingKey" required />
          </div>
          <footer class="clinical-form-actions"><button type="button" class="clinical-button secondary" @click="resetForm(false)"><X :size="15" /> Batal / Reset</button><button type="submit" class="clinical-button primary"><LoaderCircle v-if="saving" class="spin" :size="15" /><Save v-else :size="15" />{{ saving ? 'Menyimpan...' : editingKey ? 'Simpan Perubahan' : `Simpan ${treatments.length || ''} Penanganan` }}</button></footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card handling-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Riwayat Tindakan</span>
          <h3>Penanganan Dokter & Petugas</h3>
          <p>
            <template v-if="kataKunciTindakan">{{ tableRecordsTampil.length }} dari {{ records.length }} tindakan ditampilkan</template>
            <template v-else>{{ records.length }} tindakan ditemukan</template>
          </p>
        </div>
        <TableSearch
          v-if="records.length > 0"
          v-model="kataKunciTindakan"
          placeholder="Cari tindakan, dokter, atau petugas..."
          :total="records.length"
          :filtered="tableRecordsTampil.length"
        />
      </header>
      <div v-if="loading" class="clinical-state"><LoaderCircle class="spin" :size="25" /><strong>Menarik data penanganan...</strong></div>
      <div v-else-if="error" class="clinical-state error"><strong>{{ error }}</strong><button type="button" @click="loadRecords">Coba Lagi</button></div>
      <div v-else-if="records.length === 0" class="clinical-state"><strong>Belum ada penanganan dokter dan petugas.</strong></div>
      <DataTable v-else class="handling-table" :rows="tableRecordsTampil" data-key="_key" empty-message="Penanganan tidak ditemukan.">
        <Column header="TANGGAL / JAM" style="min-width:130px"><template #body="{ data }"><div class="clinical-table-main"><strong>{{ formatDate(data.tanggal) }}</strong><span>{{ data.jam }}</span></div></template></Column>
        <Column header="TINDAKAN / TAGIHAN" style="min-width:260px"><template #body="{ data }"><div class="clinical-table-main"><strong>{{ data.nama_tindakan }}</strong><span>{{ data.kode_tindakan }}</span></div></template></Column>
        <Column header="DOKTER" style="min-width:220px"><template #body="{ data }"><div class="clinical-table-main"><strong>{{ data.nama_dokter }}</strong><span>{{ data.kode_dokter }}</span></div></template></Column>
        <Column header="PETUGAS" style="min-width:220px"><template #body="{ data }"><div class="clinical-table-main"><strong>{{ data.nama_petugas }}</strong><span>{{ data.kode_petugas }}</span></div></template></Column>
        <Column header="TOTAL" style="min-width:140px"><template #body="{ data }"><strong>{{ rupiah(data.total) }}</strong></template></Column>
        <Column header="AKSI" style="min-width:135px"><template #body="{ data }"><div v-if="!billingLocked" class="clinical-table-actions"><button type="button" @click="editRecord(data)"><Pencil :size="14" /> Edit</button><button type="button" class="danger" @click="deleteTarget = data"><Trash2 :size="14" /> Hapus</button></div><span v-else>Terkunci</span></template></Column>
      </DataTable>
    </article>

    <div v-if="deleteTarget" class="clinical-confirm-backdrop" @click.self="deleteTarget = null"><section class="clinical-confirm-dialog"><h3>Hapus Penanganan?</h3><p>{{ deleteTarget.nama_tindakan }} tanggal {{ formatDate(deleteTarget.tanggal) }} pukul {{ deleteTarget.jam }} akan dihapus.</p><div><button type="button" class="clinical-button secondary" @click="deleteTarget = null">Batal</button><button type="button" class="clinical-button danger" :disabled="deleting" @click="confirmDelete"><LoaderCircle v-if="deleting" class="spin" :size="15" /><Trash2 v-else :size="15" />{{ deleting ? 'Menghapus...' : 'Hapus' }}</button></div></section></div>
  </section>
</template>
