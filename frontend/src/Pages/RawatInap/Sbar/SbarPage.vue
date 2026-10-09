<script setup lang="ts">
import { ChevronDown, ChevronUp, Eye, LoaderCircle, Pencil, RefreshCw, Save, ShieldCheck, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import InputPencarian from '../../../Components/Ui/InputPencarian.vue'
import TableSearch from '../../../Components/Ui/TableSearch.vue'
import { bidangSbar } from '../../../types/sbar'
import type { PropsSbar } from '../../../types/sbar'
import SbarIsi from './SbarIsi.vue'
import { useSbar } from './useSbar'

const props = defineProps<PropsSbar>()
const {
  loading, saving, error, errorSimpan, terbuka, keyword, statusFilter, hasil,
  editing, detail, target, aksi, petugas, validator, form, terkunci, rows, ringkasan,
  muat, reset, waktuSekarang, cariPetugas, cariDokter, edit, konfirmasi, simpan,
} = useSbar(props)
const pilihanStatus = [
  { label: 'Semua status', value: '' },
  { label: 'Belum diverifikasi', value: 'belum' },
  { label: 'Sudah diverifikasi', value: 'sudah' },
]
</script>

<template>
  <section class="clinical-page sbar-page" :aria-busy="loading || saving">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Komunikasi SBAR &amp; Verifikasi</span>
          <h3>{{ editing ? 'Edit SBAR' : 'Input SBAR' }}</h3>
          <p>
            Petugas: {{ petugas.nama || '-' }}
            <small v-if="petugas.jabatan">· {{ petugas.jabatan }}</small>
          </p>
        </div>
        <div class="clinical-section-tools">
          <button v-if="editing" class="clinical-button secondary" type="button" :disabled="saving" @click="reset">
            <X :size="15" /> Batal Edit
          </button>
          <button
            class="clinical-button toggle icon-only"
            type="button"
            :title="terbuka ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'"
            :aria-label="terbuka ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'"
            :aria-expanded="terbuka"
            aria-controls="sbar-form"
            @click="terbuka = !terbuka"
          >
            <ChevronUp v-if="terbuka" :size="15" />
            <ChevronDown v-else :size="15" />
          </button>
        </div>
      </header>
      <form v-show="terbuka" id="sbar-form" class="clinical-form" @submit.prevent="simpan()">
        <fieldset class="form-compact" :disabled="terkunci">
          <p v-if="editing" class="sbar-note" role="status">
            Mengedit catatan {{ editing.data.tgl_perawatan }} {{ editing.data.jam_rawat }} WITA. Waktu dan pencatat tetap.
          </p>
          <div class="clinical-time-grid">
            <FormInput v-model="form.tgl_perawatan" label="Tanggal" type="date" required :disabled="!!editing || terkunci" />
            <FormInput v-model="form.jam_rawat" label="Jam (WITA)" type="time" step="1" required :disabled="!!editing || terkunci" />
            <FormInput :model-value="petugas.jabatan || 'Belum tersedia'" class="wide" label="Profesi / Jabatan" readonly />
            <InputPencarian
              v-model="petugas"
              class="petugas"
              label="Dokter / Petugas Pencatat"
              :search="cariPetugas"
              description-field="jabatan"
              :disabled="!hasil?.boleh_pilih_petugas || !!editing || terkunci"
              required
            />
          </div>
          <p v-if="hasil && !hasil.boleh_pilih_petugas && !hasil.petugas_login.kode" class="patient-error">
            Akun belum cocok dengan NIK pegawai SIMRS. Hubungi administrator untuk pemetaan akun.
          </p>
          <div class="clinical-soap-grid">
            <FormInput
              v-for="bidang in bidangSbar"
              :key="bidang.key"
              v-model="form[bidang.key]"
              :label="`${bidang.huruf} — ${bidang.label}`"
              jenis="textarea"
              :rows="3"
              :maxlength="2000"
              :placeholder="bidang.petunjuk"
              :disabled="terkunci"
            />
            <FormInput
              v-model="form.instruksi"
              class="sbar-instruksi"
              label="Instruksi"
              jenis="textarea"
              :rows="2"
              :maxlength="hasil?.batas_instruksi || 2000"
              :hint="hasil ? `Maksimal ${hasil.batas_instruksi} karakter.` : ''"
              placeholder="Instruksi pelayanan dan tindak lanjut"
              :disabled="terkunci"
            />
          </div>
          <p class="sbar-note">Isi minimal salah satu S/B/A/R. Untuk verifikasi, Situation dan Background wajib terisi. Simpan tidak otomatis memverifikasi.</p>
          <p v-if="hasil?.dpjp.length" class="sbar-profesi">DPJP kunjungan: {{ hasil.dpjp.map(d => d.nama).join(' · ') }}</p>
          <p v-if="errorSimpan && !target" class="patient-error" role="alert">{{ errorSimpan }}</p>
          <footer class="clinical-form-actions">
            <button class="clinical-button secondary" type="button" :disabled="terkunci || !!editing" @click="waktuSekarang">
              <RefreshCw :size="15" /> Waktu Sekarang
            </button>
            <button class="clinical-button secondary" type="button" :disabled="saving" @click="reset">
              <X :size="15" /> Batal / Reset
            </button>
            <button class="clinical-button primary" type="submit" :disabled="terkunci">
              <LoaderCircle v-if="saving" class="spin" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan SBAR' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Riwayat Kunjungan</span>
          <h3>Catatan SBAR</h3>
          <p v-if="hasil && !error && !loading">{{ ringkasan.total }} catatan · {{ ringkasan.validasi }} sudah diverifikasi</p>
        </div>
        <div class="clinical-section-tools">
          <TableSearch
            v-if="ringkasan.total > 0"
            v-model="keyword"
            placeholder="Cari petugas, isi SBAR, instruksi..."
            :total="ringkasan.total"
            :filtered="rows.length"
          />
          <button class="clinical-button secondary" type="button" :disabled="loading || saving" @click="muat">
            <RefreshCw :size="15" /> Muat Ulang
          </button>
        </div>
      </header>
      <div v-if="loading" class="clinical-state" role="status">
        <LoaderCircle class="spin" :size="25" />
        <strong>Menarik catatan SBAR...</strong>
      </div>
      <div v-else-if="error" class="clinical-state error" role="alert">
        <strong>{{ error }}</strong>
        <button type="button" :disabled="saving" @click="muat">Coba Lagi</button>
      </div>
      <div v-else-if="!ringkasan.total" class="clinical-state">
        <strong>Belum ada catatan SBAR untuk kunjungan ini.</strong>
      </div>
      <template v-else>
        <div class="sbar-history-tools form-compact">
          <FormInput v-model="statusFilter" label="Status Verifikasi" jenis="select" :options="pilihanStatus" />
          <p class="sbar-note">
            Edit/hapus hanya untuk pencatat atau administrator sebelum verifikasi.
            Verifikasi dilakukan oleh dokter; administrator dapat memilih dokter validator.
          </p>
        </div>
        <DataTable
          :rows="rows"
          data-key="revisi"
          paginator
          :rows-per-page="10"
          :rows-per-page-options="[10, 25, 50]"
          empty-message="Catatan SBAR tidak ditemukan."
        >
          <Column header="TANGGAL / JAM" style="min-width:140px">
            <template #body="{ data: r }">
              <div class="clinical-table-main">
                <strong>{{ r.data.tgl_perawatan }}</strong>
                <span>{{ r.data.jam_rawat }} WITA</span>
              </div>
            </template>
          </Column>
          <Column header="PETUGAS" style="min-width:190px">
            <template #body="{ data: r }">
              <div class="clinical-table-main">
                <strong>{{ r.nama_petugas || r.data.nip }}</strong>
                <span>{{ r.jabatan || r.data.nip }}</span>
              </div>
            </template>
          </Column>
          <Column
            v-for="bidang in bidangSbar"
            :key="bidang.key"
            :header="`${bidang.huruf} — ${bidang.label.toUpperCase()}`"
            style="min-width:230px"
          >
            <template #body="{ data: r }">
              <p class="clinical-table-note">{{ r.data[bidang.key] || '-' }}</p>
            </template>
          </Column>
          <Column header="INSTRUKSI" style="min-width:230px">
            <template #body="{ data: r }">
              <p class="clinical-table-note">{{ r.data.instruksi || '-' }}</p>
            </template>
          </Column>
          <Column header="VERIFIKASI" style="min-width:220px">
            <template #body="{ data: r }">
              <div class="clinical-table-main">
                <strong>{{ r.status || (r.terkunci ? 'Terkunci' : 'Belum diverifikasi') }}</strong>
                <span v-if="r.validator">{{ r.nama_validator || r.validator }}</span>
                <span v-if="r.tanggal_validasi">{{ r.tanggal_validasi }} {{ r.jam_validasi }} WITA</span>
              </div>
            </template>
          </Column>
          <Column header="AKSI" frozen align-frozen="right" style="min-width:170px">
            <template #body="{ data: r }">
              <div class="clinical-table-actions sbar-table-actions">
                <button type="button" @click="detail = r">
                  <Eye :size="13" /> Rincian
                </button>
                <button v-if="r.bisa_ubah" type="button" :disabled="terkunci" @click="edit(r)">
                  <Pencil :size="13" /> Edit
                </button>
                <button v-if="r.bisa_ubah" class="danger" type="button" :disabled="terkunci" @click="konfirmasi(r, 'hapus')">
                  <Trash2 :size="13" /> Hapus
                </button>
                <button v-if="r.bisa_verifikasi" type="button" :disabled="terkunci" @click="konfirmasi(r, 'verifikasi')">
                  <ShieldCheck :size="13" /> Verifikasi
                </button>
              </div>
            </template>
          </Column>
        </DataTable>
      </template>
    </article>

    <Dialog :visible="!!detail" modal header="Rincian SBAR" :style="{ width: '760px', maxWidth: '95vw' }" @update:visible="detail = null">
      <SbarIsi v-if="detail" :catatan="detail" />
    </Dialog>
    <Dialog
      :visible="!!target"
      modal
      :header="aksi === 'verifikasi' ? 'Verifikasi SBAR' : 'Hapus SBAR?'"
      :closable="!saving"
      :close-on-escape="!saving"
      :style="{ width: '760px', maxWidth: '95vw' }"
      @update:visible="!saving && (target = null)"
    >
      <SbarIsi v-if="target" :catatan="target" />
      <template v-if="aksi === 'verifikasi'">
        <InputPencarian v-model="validator" label="Dokter Validator" :search="cariDokter" :disabled="!hasil?.boleh_pilih_petugas || saving" required />
        <p>Dengan memverifikasi, isi SBAR yang ditampilkan disahkan dan dikunci dari edit/hapus. Waktu verifikasi menggunakan waktu server WITA.</p>
      </template>
      <p v-else>Catatan SBAR dan instruksi terkait akan dihapus dari SIMRS. Tindakan ini tidak dapat dibatalkan lewat halaman ini.</p>
      <p v-if="errorSimpan" class="patient-error" role="alert">{{ errorSimpan }}</p>
      <template #footer>
        <button class="clinical-button secondary" :disabled="saving" @click="target = null">Batal</button>
        <button class="clinical-button" :class="aksi === 'verifikasi' ? 'primary' : 'danger'" :disabled="terkunci" @click="simpan(true)">
          {{ saving ? 'Memproses...' : aksi === 'verifikasi' ? 'Ya, Verifikasi SBAR' : 'Ya, Hapus SBAR' }}
        </button>
      </template>
    </Dialog>
  </section>
</template>

<style src="./sbar.css" scoped></style>
