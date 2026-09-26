<script setup lang="ts">
import { ChevronDown, ChevronUp, LoaderCircle, Pencil, Printer, RefreshCw, Save, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import TableSearch from '../../../Components/Ui/TableSearch.vue'
import { bidangHais, type PropsHais } from '../../../types/dataHais'
import { useDataHais } from './useDataHais'

const props = defineProps<PropsHais>()
const {
  loading, saving, error, errorSimpan, keyword, rows, records, formVisible, editing,
  hapusTarget, kamar, form, terkunci, mulai, selesai, errorFilter,
  waktuSekarang, reset, muat, edit, mutasi, cetak,
} = useDataHais(props)
</script>

<template>
  <section class="clinical-page hais-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Pelayanan Rawat Inap</span>
          <h3>{{ editing ? 'Edit Data HAIs' : 'Data HAIs' }}</h3>
          <p>Pencatatan HAIs pasien rawat inap. Data baru disimpan di SIRAPI; riwayat lama hanya dapat dibaca.</p>
        </div>
        <div class="clinical-section-tools">
          <button v-if="editing" type="button" class="clinical-button secondary" :disabled="saving" @click="reset">
            <X :size="15" /> Batal Edit
          </button>
          <button
            type="button"
            class="clinical-button toggle icon-only"
            :aria-expanded="formVisible"
            aria-controls="hais-form"
            :aria-label="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'"
            :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'"
            @click="formVisible = !formVisible"
          >
            <ChevronUp v-if="formVisible" :size="15" />
            <ChevronDown v-else :size="15" />
          </button>
        </div>
      </header>
      <form v-show="formVisible" id="hais-form" class="clinical-form" @submit.prevent="mutasi()">
        <fieldset class="form-compact" :disabled="saving">
          <div class="hais-identitas-grid">
            <FormInput v-model="form.tanggal" label="Tanggal" type="date" required :disabled="terkunci" />
            <FormInput :model-value="form.kd_kamar || kamar" label="Kamar / Bed" disabled />
            <FormInput v-model="form.DEKU" label="Dekubitus" jenis="select"
              :options="[{ label: 'Tidak', value: 'TIDAK' }, { label: 'Iya', value: 'IYA' }]"
              :disabled="terkunci" required />
          </div>
          <section v-for="grup in ['Hari Pemasangan Alat', 'Infeksi RS', 'Kultur dan Antibiotik']" :key="grup" class="hais-group">
            <h4>{{ grup }}</h4>
            <div class="hais-vital-grid">
              <FormInput
                v-for="bidang in bidangHais.filter(b => b.grup === grup)"
                :key="bidang.key"
                v-model="form[bidang.key]"
                :label="bidang.label"
                :type="bidang.angka ? 'number' : 'text'"
                :min="bidang.angka ? 0 : undefined"
                :max="bidang.angka ? 99 : undefined"
                :step="bidang.angka ? 1 : undefined"
                :maxlength="bidang.angka ? 2 : 200"
                :required="bidang.angka"
                :disabled="terkunci"
              />
            </div>
          </section>
          <p class="hais-catatan">Kamar mengikuti kamar terakhir pasien, termasuk kamar ibu pada rawat gabung. Nilai angka diisi 0–99 sesuai pencatatan.</p>
          <p v-if="errorSimpan" class="patient-error" role="alert">{{ errorSimpan }}</p>
          <footer class="clinical-form-actions">
            <button type="button" class="clinical-button secondary" :disabled="terkunci" @click="waktuSekarang">
              <RefreshCw :size="15" /> Waktu Sekarang
            </button>
            <button type="button" class="clinical-button secondary" :disabled="saving" @click="reset">
              <X :size="15" /> Batal / Reset
            </button>
            <button type="submit" class="clinical-button primary" :disabled="terkunci">
              <LoaderCircle v-if="saving" class="spin" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan HAIs' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Kunjungan {{ patient.no_rawat }}</span>
          <h3>Riwayat HAIs</h3>
          <p v-if="loading">Sedang memuat catatan...</p>
          <p v-else-if="!error">{{ rows.length }} dari {{ records.length }} catatan ditampilkan.</p>
        </div>
        <div class="clinical-section-tools">
          <TableSearch v-model="keyword" placeholder="Cari tanggal, kamar, kultur, atau hasil..." :total="records.length" :filtered="rows.length" aria-label="Cari riwayat hais" />
          <button type="button" class="clinical-button secondary" :disabled="loading || saving" @click="muat">
            <RefreshCw :size="15" /> Muat Ulang
          </button>
          <button type="button" class="clinical-button secondary" :disabled="loading || !!error || !rows.length" @click="cetak">
            <Printer :size="15" /> Cetak
          </button>
        </div>
      </header>
      <div class="hais-filter">
        <FormInput v-model="mulai" label="Dari Tanggal" type="date" />
        <FormInput v-model="selesai" label="Sampai Tanggal" type="date" />
        <button type="button" class="clinical-button secondary" @click="mulai = ''; selesai = ''; keyword = ''">Semua Catatan</button>
        <p v-if="errorFilter" class="patient-error" role="alert">{{ errorFilter }}</p>
      </div>
      <div v-if="loading" class="clinical-state" role="status">
        <LoaderCircle class="spin" :size="25" />
        <strong>Menarik catatan hais...</strong>
      </div>
      <div v-else-if="error" class="clinical-state error" role="alert">
        <strong>{{ error }}</strong>
        <button type="button" @click="muat">Coba Lagi</button>
      </div>
      <DataTable
        v-else
        :rows="rows"
        data-key="kunci"
        paginator
        :rows-per-page="10"
        :rows-per-page-options="[10, 25, 50]"
        empty-message="Catatan hais tidak ditemukan."
      >
        <Column header="Tanggal" style="min-width:150px">
          <template #body="{ data: r }">
            <div class="clinical-table-main">
              <strong>{{ r.data.tanggal }}</strong>
            </div>
          </template>
        </Column>
        <Column v-for="bidang in bidangHais" :key="bidang.key" :header="bidang.label" style="min-width:105px">
          <template #body="{ data: r }">{{ r.data[bidang.key] || '—' }}</template>
        </Column>
        <Column header="Dekubitus">
          <template #body="{ data: r }">{{ r.data.DEKU }}</template>
        </Column>
        <Column header="Kamar / Bed" style="min-width:140px">
          <template #body="{ data: r }">{{ r.data.kd_kamar }}</template>
        </Column>
        <Column field="sumber" header="Sumber" style="min-width:140px" />
        <Column header="Aksi" style="min-width:140px">
          <template #body="{ data: r }">
            <div v-if="r.bisa_ubah" class="clinical-table-actions">
              <button type="button" :disabled="terkunci" @click="edit(r)"><Pencil :size="14" /> Edit</button>
              <button type="button" class="danger" :disabled="terkunci" @click="hapusTarget = r; errorSimpan = ''"><Trash2 :size="14" /> Hapus</button>
            </div>
            <span v-else class="clinical-owner-note">Hanya baca</span>
          </template>
        </Column>
      </DataTable>
    </article>

    <Dialog :visible="!!hapusTarget" modal header="Hapus Data HAIs?" :closable="!saving" :style="{ width: '480px', maxWidth: '95vw' }" @update:visible="!saving && (hapusTarget = null)">
      <p>Catatan {{ hapusTarget?.data.tanggal }} akan dihapus dari SIRAPI. Tindakan ini tidak dapat dibatalkan.</p>
      <p v-if="errorSimpan" role="alert">{{ errorSimpan }}</p>
      <template #footer>
        <button type="button" class="clinical-button secondary" :disabled="saving" @click="hapusTarget = null">Batal</button>
        <button type="button" class="clinical-button danger" :disabled="saving" @click="mutasi(true)">{{ saving ? 'Menghapus...' : 'Ya, Hapus' }}</button>
      </template>
    </Dialog>
  </section>
</template>

<style src="@/Pages/RawatInap/DataHais/data-hais.css" scoped></style>
