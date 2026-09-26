<script setup lang="ts">
import { Activity, ChevronDown, ChevronUp, Eye, LoaderCircle, Pencil, Printer, RefreshCw, Save, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import InputPencarian from '../../../Components/Ui/InputPencarian.vue'
import TableSearch from '../../../Components/Ui/TableSearch.vue'
import type { PropsNewsAnak } from '../../../types/newsAnak'
import { useNewsAnak } from './useNewsAnak'

const props = defineProps<PropsNewsAnak>()
const {
  loading, saving, error, errorSimpan, keyword, rows, records, formVisible, editing, detail,
  hapusTarget, petugas, petugasLogin, bolehPilihPetugas, form, terkunci, bidang, panduan,
  eskalasi, perluVerifikasi, warnaTotal, skor, lengkap, jumlahTerisi, total, parameter, warnaSkor, warnaParameter, mulai, selesai, errorFilter,
  waktuSekarang, reset, muat, cariPetugas, edit, mutasi, cetak, verifikasiCatatan,
} = useNewsAnak(props)
</script>

<template>
  <section class="clinical-page news-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Pelayanan Rawat Inap Anak</span>
          <h3>{{ editing ? 'Edit Pemantauan NEWS Anak' : 'Pemantauan NEWS Anak' }}</h3>
          <p>Penilaian mengikuti RMPemantauanNEWS SIMRS. Data disimpan langsung ke SIMRS.</p>
        </div>
        <div class="clinical-section-tools">
          <button v-if="editing" type="button" class="clinical-button secondary" :disabled="saving" @click="reset">
            <X :size="15" /> Batal Edit
          </button>
          <button
            type="button"
            class="clinical-button toggle icon-only"
            :aria-expanded="formVisible"
            aria-controls="news-form"
            :aria-label="formVisible ? 'Sembunyikan Form' : 'Tampilkan Form'"
            :title="formVisible ? 'Sembunyikan Form' : 'Tampilkan Form'"
            @click="formVisible = !formVisible"
          >
            <ChevronUp v-if="formVisible" :size="15" />
            <ChevronDown v-else :size="15" />
          </button>
        </div>
      </header>
      <form v-show="formVisible" id="news-form" class="clinical-form" @submit.prevent="mutasi()">
        <fieldset class="form-compact" :disabled="terkunci">
          <div class="news-identitas-grid">
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
            Akun login belum terhubung ke data petugas SIMRS. Riwayat dapat dilihat, tetapi penilaian belum dapat disimpan.
          </p>

          <section class="news-summary" :class="warnaTotal" aria-live="polite">
            <div class="news-summary-score">
              <Activity :size="24" aria-hidden="true" />
              <div>
                <span>{{ lengkap ? 'Total Skor NEWS' : 'Skor Sementara NEWS' }}</span>
                <strong>{{ total ?? '—' }} <small>/ 14</small></strong>
                <span class="news-progress">{{ jumlahTerisi }} / 5 parameter dipilih</span>
              </div>
            </div>
            <div class="news-summary-guidance">
              <span>Respon</span>
              <p class="news-response">{{ parameter || 'Pilih kelima parameter sesuai hasil pemeriksaan untuk menampilkan skor dan panduan.' }}</p>
            </div>
          </section>
          <p v-if="perluVerifikasi" class="patient-error" role="alert">
            Perlu verifikasi klinis: lima parameter berskor 1 masuk respon bawaan pada kode SIMRS.
            Konfirmasikan kepada penanggung jawab klinis; jangan menganggapnya sebagai kondisi normal.
          </p>
          <section v-if="lengkap" class="news-guidance">
            <strong>Proses Eskalasi</strong>
            <p class="news-response">{{ eskalasi }}</p>
          </section>

          <div class="news-parameter-grid">
            <section v-for="(b, index) in bidang" :key="b.key" class="news-parameter" :class="warnaParameter(skor[index])">
              <header>
                <h4>{{ index + 1 }}. {{ b.label }}</h4>
                <span class="news-score-badge" :class="warnaParameter(skor[index])">Skor {{ skor[index] >= 0 ? skor[index] : '—' }}</span>
              </header>
              <FormInput
                v-model="form[b.key]"
                :label="'Hasil ' + b.label"
                jenis="select"
                :options="b.pilihan.map((value, i) => ({ label: `${value.trim() || 'Kosong (opsi legacy)'} — Skor ${b.nilai[i]}`, value }))"
                :disabled="terkunci"
                required
              />
              <p class="news-selected">{{ form[b.key] || 'Belum dipilih' }}</p>
            </section>
          </div>

          <div class="news-color-legend" aria-label="Warna skor tiap parameter">
            <span>Keterangan skor parameter:</span>
            <span class="news-score-badge score-green">0 · Hijau</span>
            <span class="news-score-badge score-yellow">1 · Kuning</span>
            <span class="news-score-badge score-orange">2 · Oranye</span>
            <span class="news-score-badge score-red">3 · Merah</span>
          </div>

          <details class="news-guidance">
            <summary>Panduan skor dari SIMRS</summary>
            <dl>
              <template v-for="p in panduan" :key="p.kode">
                <dt>{{ p.syarat }}</dt>
                <dd class="news-response"><p>{{ p.respon }}</p><p>{{ p.eskalasi }}</p></dd>
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
              {{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan NEWS' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Kunjungan {{ patient.no_rawat }}</span>
          <h3>Riwayat Pemantauan NEWS</h3>
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
      <div class="news-filter">
        <FormInput v-model="mulai" label="Dari Tanggal" type="date" />
        <FormInput v-model="selesai" label="Sampai Tanggal" type="date" />
        <button type="button" class="clinical-button secondary" @click="mulai = ''; selesai = ''; keyword = ''">Semua Catatan</button>
        <p v-if="errorFilter" class="patient-error" role="alert">{{ errorFilter }}</p>
      </div>
      <div v-if="loading" class="clinical-state" role="status">
        <LoaderCircle class="spin" :size="25" />
        <strong>Menarik pemantauan NEWS Anak...</strong>
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
        empty-message="Pemantauan NEWS tidak ditemukan."
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
              <span class="news-score-badge" :class="warnaParameter(r.data[b.skor])">Skor {{ r.data[b.skor] || '—' }}</span>
            </div>
          </template>
        </Column>
        <Column header="Total" style="min-width:75px">
          <template #body="{ data: r }">
            <strong class="news-total-badge" :class="warnaSkor(r.data)">{{ r.data.skor_total }}</strong>
          </template>
        </Column>
        <Column header="Respon" style="min-width:320px">
          <template #body="{ data: r }">
            <span class="news-response">{{ r.data.respon }}</span>
            <p v-if="verifikasiCatatan(r.data)" class="patient-error">Perlu verifikasi klinis: lima skor 1 menggunakan respon bawaan legacy.</p>
          </template>
        </Column>
        <Column header="Proses Eskalasi" style="min-width:320px">
          <template #body="{ data: r }"><span class="news-response">{{ r.data.proses_eskalasi }}</span></template>
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

    <Dialog :visible="!!detail" modal header="Detail Pemantauan NEWS Anak" :style="{ width: '700px', maxWidth: '95vw' }" @update:visible="detail = null">
      <template v-if="detail">
        <p>{{ detail.data.tanggal }} WITA · {{ detail.nama_petugas || detail.data.nip }}</p>
        <p v-if="verifikasiCatatan(detail.data)" class="patient-error">Perlu verifikasi klinis: lima skor 1 menggunakan respon bawaan legacy.</p>
        <dl class="news-detail">
          <template v-for="b in bidang" :key="b.key">
            <dt>{{ b.label }}</dt>
            <dd>{{ detail.data[b.key] }} — Skor {{ detail.data[b.skor] }}</dd>
          </template>
          <dt>Total Skor</dt>
          <dd>{{ detail.data.skor_total }}</dd>
          <dt>Respon</dt>
          <dd class="news-response">{{ detail.data.respon }}</dd>
          <dt>Proses Eskalasi</dt>
          <dd class="news-response">{{ detail.data.proses_eskalasi }}</dd>
        </dl>
      </template>
    </Dialog>
    <Dialog
      :visible="!!hapusTarget"
      modal
      header="Hapus Pemantauan NEWS?"
      :closable="!saving"
      :close-on-escape="!saving"
      :style="{ width: '480px', maxWidth: '95vw' }"
      @update:visible="!saving && (hapusTarget = null)"
    >
      <p>Catatan {{ hapusTarget?.data.tanggal }} akan dihapus langsung dari SIMRS. Tindakan ini tidak dapat dibatalkan.</p>
      <p v-if="errorSimpan" role="alert">{{ errorSimpan }}</p>
      <template #footer>
        <button type="button" class="clinical-button secondary" :disabled="saving" @click="hapusTarget = null">Batal</button>
        <button type="button" class="clinical-button danger" :disabled="terkunci" @click="mutasi(true)">{{ saving ? 'Menghapus...' : 'Ya, Hapus' }}</button>
      </template>
    </Dialog>
  </section>
</template>

<style src="@/Pages/RawatInap/NewsAnak/news-anak.css" scoped></style>
