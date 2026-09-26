<script setup lang="ts">
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import { ChevronDown, ChevronUp, Pencil, Printer, Save, Trash2 } from '@lucide/vue'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import InputPencarian from '../../../Components/Ui/InputPencarian.vue'
import { useRiwayatPendukung, type PropsRiwayatPendukung } from './useRiwayatPendukung'

const props = defineProps<PropsRiwayatPendukung>()
const {
  loading, error, q, hasil, detail, baris, kolomUtama, label, muat,
  bidang, form, pilihan, editing, hapusTarget, saving, errorSimpan, formVisible,
  terkunci, nilaiSkor, reset, edit, cari, mutasi, cetak,
} = useRiwayatPendukung(props)
</script>

<template>
  <section class="clinical-page operasi-riwayat">
    <article v-if="bidang.length" class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>{{ editing ? 'Edit Catatan SIMRS' : 'Input Catatan Baru' }}</span>
          <h3>{{ judul }}</h3>
          <p>Simpan, edit, dan hapus langsung berlaku di SIMRS.</p>
        </div>
        <button class="clinical-button toggle" :aria-expanded="formVisible" @click="formVisible = !formVisible">
          <ChevronUp v-if="formVisible" :size="16" />
          <ChevronDown v-else :size="16" />
          {{ formVisible ? 'Tutup Form' : 'Buka Form' }}
        </button>
      </header>
      <form v-show="formVisible" class="operasi-riwayat-body" @submit.prevent="mutasi()">
        <p v-if="hasil.pesan_kunci" class="patient-error" role="status">{{ hasil.pesan_kunci }}</p>
        <p v-if="jenis.startsWith('skor_')" class="operasi-tab-help">
          Isi sesuai hasil pemeriksaan. Nilai skor dihitung dari pilihan penilaian, bukan keputusan otomatis untuk memindahkan pasien.
        </p>
        <p v-if="jenis === 'transfer_pasien_antar_ruang'" class="operasi-tab-help">
          Form serah-terima klinis. Pencatatan ini tidak memindahkan bed atau mengubah status kamar pasien.
        </p>
        <div class="operasi-input-grid">
          <template v-for="b in bidang" :key="b.kode">
            <InputPencarian
              v-if="b.jenis === 'dokter' || b.jenis === 'petugas'"
              v-model="pilihan[b.kode]"
              :label="b.label"
              :required="b.wajib"
              :disabled="terkunci"
              :search="q => cari(b.jenis, q)"
            />
            <FormInput
              v-else-if="b.jenis === 'computed'"
              :model-value="nilaiSkor[b.kode] || ''"
              :label="b.label"
              disabled
              hint="Dihitung otomatis dari skala yang dipilih."
            />
            <FormInput
              v-else
              v-model="form[b.kode]"
              :class="{ 'operasi-input-wide': b.jenis === 'textarea' }"
              :label="b.label"
              :jenis="b.jenis === 'select' || b.jenis === 'textarea' ? b.jenis : 'input'"
              :type="b.jenis === 'date' || b.jenis === 'datetime-local' ? b.jenis : 'text'"
              :options="(b.pilihan || []).map(value => ({ label: value, value }))"
              :maxlength="b.batas"
              :required="b.wajib"
              :disabled="terkunci || (b.kode === 'tanggal_pemasangan_kateter' && form.kateter_urine !== 'Ada')"
              :rows="b.kode === 'laporan_operasi' ? 12 : 3"
              step="1"
            />
          </template>
        </div>
        <p v-if="errorSimpan" class="patient-error" role="alert">{{ errorSimpan }}</p>
        <footer class="operasi-form-actions">
          <button type="button" class="clinical-button secondary" :disabled="saving" @click="reset">
            {{ editing ? 'Batal Edit' : 'Reset' }}
          </button>
          <button type="submit" class="clinical-button primary" :disabled="terkunci">
            <Save :size="16" /> {{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan ke SIMRS' }}
          </button>
        </footer>
      </form>
    </article>
    <article class="clinical-history-card">
    <header class="clinical-section-header">
      <div>
        <span>Riwayat SIMRS · {{ patient.no_rawat }}</span>
        <h3>{{ judul }}</h3>
        <p v-if="jenis === 'laporan_operasi'">Laporan pada tanggal jadwal {{ jadwalOperasi.tanggal }}.</p>
        <p v-else>Riwayat seluruh kunjungan ini, termasuk catatan sebelum tanggal operasi.</p>
      </div>
      <button class="clinical-button secondary" :disabled="loading || saving" @click="muat">Muat Ulang</button>
    </header>
    <div class="operasi-riwayat-body">
      <p v-if="!bidang.length && !loading && !error" class="operasi-tab-help" role="status">
        Tab ini menampilkan data yang sudah tersimpan di SIMRS. Form tambah, edit, dan hapus modul ini belum tersedia di SIRAPI.
      </p>
      <FormInput v-model="q" type="search" label="Cari dalam riwayat" placeholder="Cari isi catatan..." />
      <p v-if="error" class="patient-error" role="alert">{{ error }}</p>
      <DataTable
        v-else
        class="visit-report-table"
        :rows="baris"
        data-key="_id"
        :loading="loading"
        paginator
        :rows-per-page="10"
        :rows-per-page-options="[10, 25, 50]"
        empty-message="Belum ada catatan yang sesuai."
      >
        <Column
          v-for="kolom in kolomUtama"
          :key="kolom"
          :header="label(kolom)"
          style="min-width: 150px"
        >
          <template #body="{ data: r }">
            <span class="operasi-ringkas">{{ r[kolom] || '—' }}</span>
          </template>
        </Column>
        <Column header="Aksi" style="min-width: 250px">
          <template #body="{ data: r }">
            <div class="clinical-table-actions">
              <button @click="detail = r">Detail</button>
              <button @click="cetak(r)"><Printer :size="14" /> Cetak</button>
              <template v-if="bidang.length && r._bisa_ubah === 'true'">
                <button :disabled="terkunci" @click="edit(r)"><Pencil :size="14" /> Edit</button>
                <button class="danger" :disabled="terkunci" @click="hapusTarget = r">
                  <Trash2 :size="14" /> Hapus
                </button>
              </template>
            </div>
          </template>
        </Column>
      </DataTable>
    </div>
    </article>
    <Dialog
      :visible="!!detail"
      :header="judul"
      modal
      :style="{ width: '800px', maxWidth: '95vw' }"
      @update:visible="value => { if (!value) detail = null }"
    >
      <dl v-if="detail" class="operasi-detail">
        <div v-for="kolom in hasil.kolom" :key="kolom">
          <dt>{{ label(kolom) }}</dt>
          <dd>{{ detail[kolom] || '—' }}</dd>
        </div>
      </dl>
    </Dialog>
    <Dialog
      :visible="!!hapusTarget"
      header="Hapus Catatan SIMRS?"
      modal
      :closable="!saving"
      :close-on-escape="!saving"
      :style="{ width: '480px', maxWidth: '95vw' }"
      @update:visible="value => { if (!value && !saving) hapusTarget = null }"
    >
      <p>{{ judul }} · {{ hapusTarget?.tanggal || hapusTarget?.tanggal_masuk }}</p>
      <p>Catatan dihapus permanen dari SIMRS dan tidak dapat dipulihkan melalui halaman ini.</p>
      <p v-if="errorSimpan" class="patient-error" role="alert">{{ errorSimpan }}</p>
      <template #footer>
        <button class="clinical-button secondary" :disabled="saving" @click="hapusTarget = null">Batal</button>
        <button class="clinical-button danger" :disabled="terkunci" @click="mutasi(true)">
          {{ saving ? 'Memproses...' : 'Ya, Hapus' }}
        </button>
      </template>
    </Dialog>
  </section>
</template>

<style src="@/Components/Ui/report.css" scoped></style>
<style src="@/Pages/Pasien/JadwalOperasi/riwayat-pendukung.css" scoped></style>
