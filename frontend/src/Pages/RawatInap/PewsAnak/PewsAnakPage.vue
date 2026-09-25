<script setup lang="ts">
import { Activity, ChevronDown, ChevronUp, Eye, LoaderCircle, Pencil, Printer, RefreshCw, Save, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import InputPencarian from '../../../Components/Ui/InputPencarian.vue'
import TableSearch from '../../../Components/Ui/TableSearch.vue'
import type { PropsPewsAnak } from '../../../types/pewsAnak'
import { usePewsAnak } from './usePewsAnak'

const props = defineProps<PropsPewsAnak>()
const {
  loading, saving, error, errorSimpan, keyword, rows, records, formVisible, editing, detail,
  hapusTarget, petugas, petugasLogin, bolehPilihPetugas, form, terkunci, bidang, panduan,
  skor, lengkap, jumlahTerisi, total, parameter, warnaSkor, warnaParameter, mulai, selesai, errorFilter,
  waktuSekarang, reset, muat, cariPetugas, edit, mutasi, cetak,
} = usePewsAnak(props)
</script>

<template>
  <section class="clinical-page pews-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Pelayanan Rawat Inap Anak</span>
          <h3>{{ editing ? 'Edit Pemantauan PEWS Anak' : 'Pemantauan PEWS Anak' }}</h3>
          <p>Penilaian mengikuti RMPemantauanPEWS Khanza. Data disimpan langsung ke Khanza.</p>
        </div>
        <div class="clinical-section-tools">
          <button v-if="editing" type="button" class="clinical-button secondary" :disabled="saving" @click="reset">
            <X :size="15" /> Batal Edit
          </button>
          <button
            type="button"
            class="clinical-button toggle icon-only"
            :aria-expanded="formVisible"
            aria-controls="pews-form"
            :aria-label="formVisible ? 'Sembunyikan Form' : 'Tampilkan Form'"
            :title="formVisible ? 'Sembunyikan Form' : 'Tampilkan Form'"
            @click="formVisible = !formVisible"
          >
            <ChevronUp v-if="formVisible" :size="15" />
            <ChevronDown v-else :size="15" />
          </button>
        </div>
      </header>
      <form v-show="formVisible" id="pews-form" class="clinical-form" @submit.prevent="mutasi()">
        <fieldset class="form-compact" :disabled="terkunci">
          <div class="pews-identitas-grid">
            <FormInput v-model="form.tanggal" label="Tanggal Penilaian" type="date" required :disabled="terkunci" />
            <FormInput v-model="form.jam" label="Jam (WITA)" type="time" step="1" required :disabled="terkunci" />
            <InputPencarian
              v-model="petugas"
              label="Petugas"
              :search="cariPetugas"
              required
              :disabled="terkunci || !bolehPilihPetugas"
            />
          </div>
          <p v-if="!loading && !error && !bolehPilihPetugas && !petugasLogin.kode" class="patient-error" role="alert">
            Akun login belum terhubung ke data petugas Khanza. Riwayat dapat dilihat, tetapi penilaian belum dapat disimpan.
          </p>

          <section class="pews-summary" :class="lengkap ? warnaSkor(total) : 'pending'" aria-live="polite">
            <div class="pews-summary-score">
              <Activity :size="24" aria-hidden="true" />
              <div>
                <span>{{ lengkap ? 'Total Skor PEWS' : 'Skor Sementara PEWS' }}</span>
                <strong>{{ total ?? '—' }} <small>/ 9</small></strong>
                <span class="pews-progress">{{ jumlahTerisi }} / 3 parameter dipilih</span>
              </div>
            </div>
            <div class="pews-summary-guidance">
              <span>Parameter / Tindak Lanjut</span>
              <p>{{ parameter || 'Pilih ketiga parameter sesuai hasil pemeriksaan untuk menampilkan skor dan panduan.' }}</p>
            </div>
          </section>

          <div class="pews-parameter-grid">
            <section v-for="(b, index) in bidang" :key="b.key" class="pews-parameter" :class="warnaParameter(skor[index])">
              <header>
                <h4>{{ index + 1 }}. {{ b.label }}</h4>
                <span class="pews-score-badge" :class="warnaParameter(skor[index])">Skor {{ skor[index] >= 0 ? skor[index] : '—' }}</span>
              </header>
              <FormInput
                v-model="form[b.key]"
                :label="'Hasil ' + b.label"
                jenis="select"
                :options="b.pilihan.map((value, i) => ({ label: `${value} — Skor ${i}`, value }))"
                :disabled="terkunci"
                required
              />
              <p class="pews-selected">{{ form[b.key] || 'Belum dipilih' }}</p>
            </section>
          </div>

          <div class="pews-color-legend" aria-label="Warna skor tiap parameter">
            <span>Keterangan skor parameter:</span>
            <span class="pews-score-badge score-green">0 · Hijau</span>
            <span class="pews-score-badge score-yellow">1 · Kuning</span>
            <span class="pews-score-badge score-orange">2 · Oranye</span>
            <span class="pews-score-badge score-red">3 · Merah</span>
          </div>

          <details class="pews-guidance">
            <summary>Panduan skor dari SIMRS Khanza</summary>
            <dl>
              <template v-for="p in panduan" :key="p.minimal">
                <dt>Skor {{ p.minimal === p.maksimal ? p.minimal : `${p.minimal}–${p.maksimal}` }}</dt>
                <dd>{{ p.parameter }}</dd>
              </template>
            </dl>
          </details>
          <p v-if="errorSimpan && !hapusTarget" class="patient-error" role="alert">{{ errorSimpan }}</p>
          <footer class="clinical-form-actions">
            <button type="button" class="clinical-button secondary" :disabled="terkunci" @click="waktuSekarang">
              <RefreshCw :size="15" /> Waktu Sekarang
            </button>
            <button type="button" class="clinical-button secondary" :disabled="terkunci" @click="reset">
              <X :size="15" /> Batal / Reset
            </button>
            <button type="submit" class="clinical-button primary" :disabled="terkunci || !lengkap || !petugas.kode">
              <LoaderCircle v-if="saving" class="spin" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan PEWS' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Kunjungan {{ patient.no_rawat }}</span>
          <h3>Riwayat Pemantauan PEWS</h3>
          <p v-if="loading">Sedang memuat catatan...</p>
          <p v-else-if="!error">{{ rows.length }} dari {{ records.length }} catatan ditampilkan.</p>
        </div>
        <div class="clinical-section-tools">
          <TableSearch v-model="keyword" placeholder="Cari petugas, skor, atau hasil..." :total="records.length" :filtered="rows.length" />
          <button type="button" class="clinical-button secondary" :disabled="loading || saving" @click="muat">
            <RefreshCw :size="15" /> Muat Ulang
          </button>
          <button type="button" class="clinical-button secondary" :disabled="terkunci || !rows.length" @click="cetak">
            <Printer :size="15" /> Cetak
          </button>
        </div>
      </header>
      <div class="pews-filter">
        <FormInput v-model="mulai" label="Dari Tanggal" type="date" />
        <FormInput v-model="selesai" label="Sampai Tanggal" type="date" />
        <button type="button" class="clinical-button secondary" @click="mulai = ''; selesai = ''; keyword = ''">Semua Catatan</button>
        <p v-if="errorFilter" class="patient-error" role="alert">{{ errorFilter }}</p>
      </div>
      <div v-if="loading" class="clinical-state" role="status">
        <LoaderCircle class="spin" :size="25" />
        <strong>Menarik pemantauan PEWS Anak...</strong>
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
        empty-message="Pemantauan PEWS tidak ditemukan."
      >
        <Column header="Tanggal / Jam" style="min-width:150px">
          <template #body="{ data: r }">
            <div class="clinical-table-main">
              <strong>{{ r.data.tanggal.slice(0, 10) }}</strong>
              <span>{{ r.data.tanggal.slice(11) }} WITA</span>
            </div>
          </template>
        </Column>
        <Column v-for="b in bidang" :key="b.key" :header="b.label" style="min-width:200px">
          <template #body="{ data: r }">
            <div class="clinical-table-main">
              <strong>{{ r.data[b.key] || '—' }}</strong>
              <span class="pews-score-badge" :class="warnaParameter(r.data[b.skor])">Skor {{ r.data[b.skor] || '—' }}</span>
            </div>
          </template>
        </Column>
        <Column header="Total" style="min-width:75px">
          <template #body="{ data: r }">
            <strong class="pews-total-badge" :class="warnaSkor(r.data.skor_total)">{{ r.data.skor_total }}</strong>
          </template>
        </Column>
        <Column header="Parameter / Tindak Lanjut" style="min-width:320px">
          <template #body="{ data: r }">{{ r.data.parameter_total }}</template>
        </Column>
        <Column header="Petugas" style="min-width:180px">
          <template #body="{ data: r }">
            <div class="clinical-table-main">
              <strong>{{ r.nama_petugas || r.data.nip }}</strong>
              <span>{{ r.data.nip }}</span>
            </div>
          </template>
        </Column>
        <Column header="Aksi" style="min-width:230px">
          <template #body="{ data: r }">
            <div class="clinical-table-actions">
              <button type="button" :disabled="terkunci" @click="detail = r"><Eye :size="14" /> Lihat</button>
              <template v-if="r.bisa_ubah">
                <button type="button" :disabled="terkunci" @click="edit(r)"><Pencil :size="14" /> Edit</button>
                <button type="button" class="danger" :disabled="terkunci" @click="hapusTarget = r; errorSimpan = ''"><Trash2 :size="14" /> Hapus</button>
              </template>
            </div>
          </template>
        </Column>
      </DataTable>
    </article>

    <Dialog :visible="!!detail" modal header="Detail Pemantauan PEWS Anak" :style="{ width: '700px', maxWidth: '95vw' }" @update:visible="detail = null">
      <template v-if="detail">
        <p>{{ detail.data.tanggal }} WITA · {{ detail.nama_petugas || detail.data.nip }}</p>
        <dl class="pews-detail">
          <template v-for="b in bidang" :key="b.key">
            <dt>{{ b.label }}</dt>
            <dd>{{ detail.data[b.key] }} — Skor {{ detail.data[b.skor] }}</dd>
          </template>
          <dt>Total Skor</dt>
          <dd>{{ detail.data.skor_total }}</dd>
          <dt>Parameter / Tindak Lanjut</dt>
          <dd>{{ detail.data.parameter_total }}</dd>
        </dl>
      </template>
    </Dialog>
    <Dialog
      :visible="!!hapusTarget"
      modal
      header="Hapus Pemantauan PEWS?"
      :closable="!saving"
      :close-on-escape="!saving"
      :style="{ width: '480px', maxWidth: '95vw' }"
      @update:visible="!saving && (hapusTarget = null)"
    >
      <p>Catatan {{ hapusTarget?.data.tanggal }} akan dihapus langsung dari Khanza. Tindakan ini tidak dapat dibatalkan.</p>
      <p v-if="errorSimpan" role="alert">{{ errorSimpan }}</p>
      <template #footer>
        <button type="button" class="clinical-button secondary" :disabled="saving" @click="hapusTarget = null">Batal</button>
        <button type="button" class="clinical-button danger" :disabled="terkunci" @click="mutasi(true)">{{ saving ? 'Menghapus...' : 'Ya, Hapus' }}</button>
      </template>
    </Dialog>
  </section>
</template>

<style src="@/Pages/RawatInap/PewsAnak/pews-anak.css" scoped></style>
