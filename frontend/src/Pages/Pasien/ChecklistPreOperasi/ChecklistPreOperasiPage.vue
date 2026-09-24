<script setup lang="ts">
import { ChevronDown, ChevronUp, Eye, Pencil, Printer, RefreshCw, Save, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import InputPencarian from '../../../Components/Ui/InputPencarian.vue'
import { kelompokChecklist, type PropsChecklist } from '../../../types/checklistPreOperasi'
import { useChecklistPreOperasi } from './useChecklistPreOperasi'

const props = defineProps<PropsChecklist>()
const {
  loading, saving, error, errorSimpan, peringatan, formVisible, keyword, filteredRows,
  editing, detail, hapusTarget, tanggal, form, pilihan,
  waktuSekarang, reset, muat, cariReferensi, edit, simpan, hapus, cetak,
} = useChecklistPreOperasi(props)
</script>

<template>
  <section class="pre-operasi-page clinical-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Rekam Medis</span>
          <h3>Checklist Pre Operasi</h3>
          <p>Konfirmasi persiapan pasien sebelum operasi.</p>
        </div>
        <button
          type="button"
          class="clinical-button toggle"
          :aria-expanded="formVisible"
          @click="formVisible = !formVisible"
        >
          <ChevronUp v-if="formVisible" :size="16" />
          <ChevronDown v-else :size="16" />
          {{ formVisible ? 'Tutup Form' : 'Buka Form' }}
        </button>
      </header>
      <p class="pre-info">
        Checklist baru langsung disimpan ke Khanza.
        Edit dan hapus langsung berlaku pada sumber data baris yang dipilih.
      </p>
      <form :key="String(patient.no_rawat)" v-show="formVisible" class="pre-form" @submit.prevent="simpan">
        <p v-if="editing" class="pre-info" role="status">
          Mengedit checklist {{ editing.sumber }} tanggal {{ editing.tanggal }}.
        </p>
        <FormInput
          v-model="tanggal"
          label="Tanggal / Jam Checklist (WITA)"
          type="datetime-local"
          step="1"
          :disabled="saving"
          required
        />
        <section v-for="kelompok in kelompokChecklist" :key="kelompok.judul" class="pre-section">
          <h4>{{ kelompok.judul }}</h4>
          <div class="pre-fields">
            <template v-for="bidang in kelompok.bidang" :key="bidang.kode">
              <InputPencarian
                v-if="bidang.jenis"
                v-model="pilihan[bidang.kode]"
                :label="bidang.label"
                :search="q => cariReferensi(bidang.jenis!, q)"
                :disabled="saving"
                required
              />
              <FormInput
                v-else
                v-model="form[bidang.kode]"
                :label="bidang.label"
                :jenis="bidang.pilihan ? 'select' : 'input'"
                :options="bidang.pilihan?.map(value => ({ label: value, value })) || []"
                :maxlength="bidang.maxlength"
                :required="!bidang.kode.startsWith('keterangan_')"
                :disabled="saving"
              />
            </template>
          </div>
        </section>
        <p v-if="errorSimpan" class="patient-error" role="alert">{{ errorSimpan }}</p>
        <div class="pre-actions">
          <button type="button" class="clinical-button secondary" :disabled="saving" @click="waktuSekarang">
            <RefreshCw :size="16" /> Waktu Sekarang
          </button>
          <button type="button" class="clinical-button secondary" :disabled="saving" @click="reset">
            <X :size="16" /> {{ editing ? 'Batal Edit' : 'Reset' }}
          </button>
          <button type="submit" class="clinical-button primary" :disabled="saving">
            <Save :size="16" />
            {{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan ke Khanza' }}
          </button>
        </div>
      </form>
    </article>

    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <h3>Riwayat Checklist Pre Operasi</h3>
          <p>{{ filteredRows.length }} catatan pada kunjungan ini.</p>
        </div>
        <button class="clinical-button secondary" :disabled="loading || saving" @click="muat">
          <RefreshCw :size="16" /> Muat Ulang
        </button>
      </header>
      <div class="pre-history">
        <FormInput v-model="keyword" label="Cari Riwayat" type="search" placeholder="Tanggal, tindakan, dokter, petugas..." />
        <p v-if="peringatan" class="pre-info" role="status">{{ peringatan }}</p>
        <p v-if="error" class="patient-error" role="alert">{{ error }}</p>
        <DataTable
          v-else
          class="visit-report-table"
          :rows="filteredRows"
          data-key="kunci"
          :loading="loading"
          paginator
          :rows-per-page="10"
          :rows-per-page-options="[10, 25, 50]"
          empty-message="Belum ada checklist pre operasi pada kunjungan ini."
        >
          <Column field="tanggal" header="Tanggal / Jam" style="min-width: 160px" />
          <Column header="Tindakan / SN/CN" style="min-width: 190px">
            <template #body="{ data: r }">
              <strong>{{ r.data.tindakan }}</strong>
              <small>SN/CN: {{ r.data.sncn }}</small>
            </template>
          </Column>
          <Column header="Dokter Bedah / Anestesi" style="min-width: 220px">
            <template #body="{ data: r }">
              <strong>{{ r.data.kd_dokter_bedah_nama || r.data.kd_dokter_bedah }}</strong>
              <small>{{ r.data.kd_dokter_anestesi_nama || r.data.kd_dokter_anestesi }}</small>
            </template>
          </Column>
          <Column field="sumber" header="Sumber" style="min-width: 90px" />
          <Column header="Aksi" style="min-width: 225px">
            <template #body="{ data: r }">
              <div class="pre-row-actions">
                <button class="clinical-button secondary" @click="detail = r">
                  <Eye :size="15" /> Detail
                </button>
                <button v-if="r.bisa_ubah" class="clinical-button secondary" :disabled="saving" @click="edit(r)">
                  <Pencil :size="15" /> Edit
                </button>
                <button v-if="r.bisa_ubah" class="clinical-button danger" :disabled="saving" @click="hapusTarget = r">
                  <Trash2 :size="15" /> Hapus
                </button>
              </div>
              <small v-if="!r.bisa_ubah">Hanya baca</small>
            </template>
          </Column>
        </DataTable>
      </div>
    </article>

    <Dialog
      :visible="!!detail"
      modal
      header="Detail Checklist Pre Operasi"
      :style="{ width: '760px', maxWidth: '95vw' }"
      @update:visible="value => { if (!value) detail = null }"
    >
      <template v-if="detail">
        <p>{{ detail.no_rawat }} · {{ detail.tanggal }} · {{ detail.sumber }}</p>
        <section v-for="kelompok in kelompokChecklist" :key="kelompok.judul" class="pre-section">
          <h4>{{ kelompok.judul }}</h4>
          <dl class="pre-detail">
            <div v-for="bidang in kelompok.bidang" :key="bidang.kode">
              <dt>{{ bidang.label }}</dt>
              <dd>{{ detail.data[bidang.kode + '_nama'] || detail.data[bidang.kode] || '-' }}</dd>
            </div>
          </dl>
        </section>
      </template>
      <template #footer>
        <button class="clinical-button secondary" @click="detail = null">Tutup</button>
        <button class="clinical-button primary" @click="cetak">
          <Printer :size="16" /> Cetak
        </button>
      </template>
    </Dialog>
    <Dialog
      :visible="!!hapusTarget"
      modal
      header="Hapus checklist?"
      :closable="!saving"
      :close-on-escape="!saving"
      :style="{ width: '460px', maxWidth: '95vw' }"
      @update:visible="value => { if (!value && !saving) hapusTarget = null }"
    >
      <p>Hapus checklist tanggal {{ hapusTarget?.tanggal }} dari {{ hapusTarget?.sumber }}?</p>
      <p v-if="hapusTarget?.sumber === 'Khanza'">Data akan dihapus langsung dari Khanza dan tidak dapat dipulihkan melalui halaman ini.</p>
      <template #footer>
        <button class="clinical-button secondary" :disabled="saving" @click="hapusTarget = null">Batal</button>
        <button class="clinical-button danger" :disabled="saving" @click="hapus">
          {{ saving ? 'Memproses...' : 'Ya, Hapus' }}
        </button>
      </template>
    </Dialog>
  </section>
</template>

<style src="@/Components/Ui/report.css" scoped></style>
<style src="@/Pages/Pasien/ChecklistPreOperasi/checklist-pre-operasi.css" scoped></style>
