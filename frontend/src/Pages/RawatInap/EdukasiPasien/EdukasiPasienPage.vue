<script setup lang="ts">
import { ChevronDown, ChevronUp, LoaderCircle, Pencil, Printer, RefreshCw, Save, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import InputPencarian from '../../../Components/Ui/InputPencarian.vue'
import TableSearch from '../../../Components/Ui/TableSearch.vue'
import { bidangEdukasi, type PropsEdukasi } from '../../../types/edukasiPasien'
import { useEdukasiPasien } from './useEdukasiPasien'

const props = defineProps<PropsEdukasi>()
const {
  loading, saving, error, errorSimpan, keyword, rows, records, formVisible, editing,
  hapusTarget, petugas, ruangan, bolehPilihPetugas, form, terkunci, mulai, selesai, errorFilter,
  waktuSekarang, reset, muat, cariReferensi, edit, mutasi, cetak,
} = useEdukasiPasien(props)
</script>

<template>
  <section class="clinical-page edukasi-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Pelayanan Rawat Inap</span>
          <h3>{{ editing ? 'Edit Catatan Edukasi' : 'Catatan Edukasi Pasien' }}</h3>
          <p>Catatan edukasi dibaca dan disimpan langsung pada tabel SIMRS.</p>
        </div>
        <div class="clinical-section-tools">
          <button v-if="editing" type="button" class="clinical-button secondary" :disabled="saving" @click="reset">
            <X :size="15" /> Batal Edit
          </button>
          <button type="button" class="clinical-button toggle icon-only" :aria-expanded="formVisible"
            aria-controls="edukasi-form" :aria-label="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'"
            :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'"
            @click="formVisible = !formVisible">
            <ChevronUp v-if="formVisible" :size="15" />
            <ChevronDown v-else :size="15" />
          </button>
        </div>
      </header>
      <form v-show="formVisible" id="edukasi-form" class="clinical-form" @submit.prevent="mutasi()">
        <fieldset class="form-compact" :disabled="saving">
          <div class="edukasi-identitas-grid">
            <FormInput v-model="form.tgl_perawatan" label="Tanggal Edukasi" type="date" required :disabled="terkunci" />
            <FormInput v-model="form.jam_rawat" label="Jam Edukasi (WITA)" type="time" step="1" required
              :disabled="terkunci" />
            <InputPencarian v-model="petugas" label="Petugas" :search="q => cariReferensi('petugas', q)"
              :disabled="terkunci || !bolehPilihPetugas" required />
            <InputPencarian v-model="ruangan" label="Ruangan" :search="q => cariReferensi('ruangan', q)"
              :disabled="terkunci" />
            <FormInput v-model="form.metode" label="Metode" jenis="select"
              :options="['Audio', 'Demonstrasi', 'Lisan', 'Tulisan', 'Visual'].map(v => ({ label: v, value: v }))"
              required :disabled="terkunci" />
            <FormInput v-model="form.durasi" label="Durasi" :maxlength="30" :disabled="terkunci" />
            <FormInput v-model="form.penerima" label="Penerima Edukasi" :maxlength="30" :disabled="terkunci" />
          </div>
          <div class="edukasi-materi-grid">
            <FormInput v-model="form.materi" label="Materi Edukasi" jenis="textarea" :rows="5" :maxlength="50"
              :disabled="terkunci" />
            <FormInput v-model="form.keterangan" label="Keterangan" jenis="textarea" :rows="5" :maxlength="255"
              :disabled="terkunci" />
          </div>
          <p v-if="!loading && !error && !bolehPilihPetugas && !petugas.kode" class="patient-error" role="alert">
            Akun login belum terhubung dengan data petugas. Hubungi administrator untuk melengkapi pemetaan akun.
          </p>
          <p class="edukasi-catatan">Petugas mengikuti akun login. Administrator dapat memilih petugas.</p>
          <p v-if="errorSimpan" class="patient-error" role="alert">{{ errorSimpan }}</p>
          <footer class="clinical-form-actions">
            <button type="button" class="clinical-button secondary" :disabled="terkunci" @click="waktuSekarang">
              <RefreshCw :size="15" /> Waktu Sekarang
            </button>
            <button type="button" class="clinical-button secondary" :disabled="saving" @click="reset">
              <X :size="15" /> Batal / Reset
            </button>
            <button type="submit" class="clinical-button primary" :disabled="terkunci || !petugas.kode">
              <LoaderCircle v-if="saving" class="spin" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan Edukasi' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Kunjungan {{ patient.no_rawat }}</span>
          <h3>Riwayat Edukasi</h3>
          <p v-if="loading">Sedang memuat catatan...</p>
          <p v-else-if="!error">{{ rows.length }} dari {{ records.length }} catatan ditampilkan.</p>
        </div>
        <div class="clinical-section-tools">
          <TableSearch v-model="keyword" placeholder="Cari tanggal, petugas, atau hasil..." :total="records.length"
            :filtered="rows.length" aria-label="Cari riwayat edukasi" />
          <button type="button" class="clinical-button secondary" :disabled="loading || saving" @click="muat">
            <RefreshCw :size="15" /> Muat Ulang
          </button>
          <button type="button" class="clinical-button secondary" :disabled="loading || !!error || !rows.length"
            @click="cetak">
            <Printer :size="15" /> Cetak
          </button>
        </div>
      </header>
      <div class="edukasi-filter">
        <FormInput v-model="mulai" label="Dari Tanggal" type="date" />
        <FormInput v-model="selesai" label="Sampai Tanggal" type="date" />
        <button type="button" class="clinical-button secondary" @click="mulai = ''; selesai = ''; keyword = ''">Semua
          Catatan</button>
        <p v-if="errorFilter" class="patient-error" role="alert">{{ errorFilter }}</p>
      </div>
      <div v-if="loading" class="clinical-state" role="status">
        <LoaderCircle class="spin" :size="25" />
        <strong>Menarik catatan edukasi...</strong>
      </div>
      <div v-else-if="error" class="clinical-state error" role="alert">
        <strong>{{ error }}</strong>
        <button type="button" @click="muat">Coba Lagi</button>
      </div>
      <DataTable v-else :rows="rows" data-key="kunci" paginator :rows-per-page="10"
        :rows-per-page-options="[10, 25, 50]" empty-message="Catatan edukasi tidak ditemukan.">
        <Column header="Tanggal / Jam" style="min-width:150px">
          <template #body="{ data: r }">
            <div class="clinical-table-main">
              <strong>{{ r.data.tgl_perawatan }}</strong>
              <span>{{ r.data.jam_rawat }} WITA</span>
            </div>
          </template>
        </Column>
        <Column v-for="bidang in bidangEdukasi" :key="bidang.key" :header="bidang.label" :style="{
          minWidth: bidang.key === 'materi' || bidang.key === 'keterangan' ? '250px' : '130px',
          whiteSpace: 'pre-wrap',
        }">
          <template #body="{ data: r }">{{ r.data[bidang.key] || '—' }}</template>
        </Column>
        <Column header="Ruangan" style="min-width:180px">
          <template #body="{ data: r }">{{ r.nama_ruangan || r.data.kd_ruangan }}</template>
        </Column>
        <Column field="sumber" header="Sumber" style="min-width:120px" />
        <Column header="Petugas" style="min-width:180px">
          <template #body="{ data: r }">
            <div class="clinical-table-main">
              <strong>{{ r.nama_petugas || r.data.nip }}</strong>
              <span>{{ r.data.nip }}</span>
            </div>
          </template>
        </Column>
        <Column header="Aksi" style="min-width:140px">
          <template #body="{ data: r }">
            <div v-if="r.bisa_ubah" class="clinical-table-actions">
              <button type="button" :disabled="terkunci" @click="edit(r)">
                <Pencil :size="14" /> Edit
              </button>
              <button type="button" class="danger" :disabled="terkunci" @click="hapusTarget = r; errorSimpan = ''">
                <Trash2 :size="14" /> Hapus
              </button>
            </div>
            <span v-else class="clinical-owner-note">Hanya baca</span>
          </template>
        </Column>
      </DataTable>
    </article>

    <Dialog :visible="!!hapusTarget" modal header="Hapus Catatan Edukasi?" :closable="!saving"
      :style="{ width: '480px', maxWidth: '95vw' }" @update:visible="!saving && (hapusTarget = null)">
      <p>Catatan {{ hapusTarget?.data.tgl_perawatan }} pukul {{ hapusTarget?.data.jam_rawat }} akan dihapus dari SIMRS.
        Tindakan ini tidak dapat dibatalkan.</p>
      <p v-if="errorSimpan" role="alert">{{ errorSimpan }}</p>
      <template #footer>
        <button type="button" class="clinical-button secondary" :disabled="saving"
          @click="hapusTarget = null">Batal</button>
        <button type="button" class="clinical-button danger" :disabled="saving" @click="mutasi(true)">{{ saving ?
          'Menghapus...' : 'Ya, Hapus' }}</button>
      </template>
    </Dialog>
  </section>
</template>

<style src="@/Pages/RawatInap/EdukasiPasien/edukasi-pasien.css" scoped></style>
