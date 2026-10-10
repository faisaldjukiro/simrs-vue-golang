<script setup lang="ts">
import { ChevronDown, ChevronUp, Eye, Image, LoaderCircle, Pencil, Printer, RefreshCw, Save, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import InputPencarian from '../../../Components/Ui/InputPencarian.vue'
import TableSearch from '../../../Components/Ui/TableSearch.vue'
import { bidangEdukasi, bidangAsesmen, pilihanPemahaman, pilihanStatusVerifikasi, type PropsEdukasi } from '../../../types/edukasiPasien'
import { useEdukasiPasien } from './useEdukasiPasien'

const props = defineProps<PropsEdukasi>()
const {
  loading, saving, error, errorSimpan, keyword, rows, records, formVisible, editing,
  hapusTarget, petugas, ruangan, bolehPilihPetugas, form, terkunci, mulai, selesai, errorFilter,
  waktuSekarang, reset, muat, cariReferensi, edit, mutasi, cetak, printing, cetakAktif,
  foto, kunciFoto, pratinjauFoto, fotoGagal, detail, detailFotoGagal, pilihFoto, batalFoto,
  detailCatatan, bidangDetail, verifikator, terverifikasi, ubahStatusVerifikasi, waktuVerifikasiSekarang,
} = useEdukasiPasien(props)
</script>

<template>
  <section class="clinical-page edukasi-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Pelayanan Rawat Inap</span>
          <h3>{{ editing ? 'Edit Catatan Edukasi' : 'Catatan Edukasi Pasien' }}</h3>
          <p>Catat pelaksanaan edukasi, kebutuhan belajar, dan hasil verifikasi pemahaman pasien atau keluarga.</p>
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
          <h4 class="edukasi-subjudul">Pelaksanaan Edukasi</h4>
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
            <FormInput v-model="form.materi" label="Materi Edukasi" jenis="textarea" :rows="5" :maxlength="65535"
              :disabled="terkunci" />
            <FormInput v-model="form.keterangan" label="Keterangan" jenis="textarea" :rows="5" :maxlength="255"
              :disabled="terkunci" />
          </div>
          <section class="edukasi-bagian" aria-labelledby="asesmen-edukasi-title">
            <h4 id="asesmen-edukasi-title" class="edukasi-subjudul">Asesmen Kebutuhan Belajar</h4>
            <p class="edukasi-petunjuk">Isi berdasarkan penerima edukasi. Biarkan kosong bila belum dikaji.</p>
            <div class="edukasi-identitas-grid">
              <FormInput v-for="bidang in bidangAsesmen" :key="bidang.key" v-model="form[bidang.key]"
                :label="bidang.label" :jenis="bidang.pilihan ? 'select' : bidang.jenis || 'input'"
                :options="[{ label: 'Belum dikaji', value: '' }, ...(bidang.pilihan || []).map(v => ({ label: v, value: v }))]"
                :rows="3" :maxlength="bidang.maxlength" :disabled="terkunci" />
            </div>
          </section>
          <section class="edukasi-bagian" aria-labelledby="verifikasi-edukasi-title">
            <h4 id="verifikasi-edukasi-title" class="edukasi-subjudul">Pemahaman dan Verifikasi</h4>
            <div class="edukasi-identitas-grid">
              <FormInput v-model="form.tingkat_pemahaman" label="Tingkat Pemahaman" jenis="select"
                :options="[{ label: 'Belum dinilai', value: '' }, ...pilihanPemahaman.map(v => ({ label: v, value: v }))]"
                :required="terverifikasi" :disabled="terkunci" />
              <FormInput :model-value="form.status_verifikasi" label="Status Verifikasi" jenis="select"
                :options="[{ label: 'Belum dicatat', value: '' }, ...pilihanStatusVerifikasi.map(v => ({ label: v, value: v }))]"
                :disabled="terkunci" @update:model-value="ubahStatusVerifikasi" />
              <FormInput v-model="form.tanggal_verifikasi" label="Waktu Verifikasi (WITA)" type="datetime-local" step="1"
                :required="terverifikasi" :disabled="terkunci || !terverifikasi" />
              <InputPencarian v-model="verifikator" label="Petugas Verifikator" :search="q => cariReferensi('petugas', q)"
                :required="terverifikasi" :disabled="terkunci || !terverifikasi || !bolehPilihPetugas" />
            </div>
            <FormInput class="edukasi-catatan-verifikasi" v-model="form.catatan_verifikasi" label="Catatan Hasil Verifikasi"
              jenis="textarea" :rows="4" :maxlength="65535" :disabled="terkunci"
              hint="Catat respons penerima edukasi, pemahaman yang dinilai, serta kebutuhan penjelasan ulang bila ada." />
            <button v-if="terverifikasi" type="button" class="clinical-button secondary" :disabled="terkunci" @click="waktuVerifikasiSekarang">
              <RefreshCw :size="15" /> Waktu Verifikasi Sekarang
            </button>
            <p class="edukasi-petunjuk">Verifikator mengikuti petugas login. Administrator dapat memilih petugas verifikator.</p>
          </section>
          <div class="edukasi-foto">
            <FormInput :key="kunciFoto" label="Foto Edukasi (Opsional)" type="file"
              :file-name="foto?.name || ''"
              accept=".jpg,.jpeg,.png,image/jpeg,image/png" :disabled="terkunci"
              hint="Satu foto JPG/PNG, maksimal 10 MB dan 40 megapiksel. Foto dikirim saat menyimpan catatan."
              @change="pilihFoto" />
            <template v-if="pratinjauFoto">
              <img v-if="!fotoGagal" :src="pratinjauFoto" alt="Pratinjau foto edukasi" class="edukasi-foto-preview"
                @error="fotoGagal = true" />
              <p v-else class="patient-error" role="alert">Foto tidak dapat ditampilkan. Periksa file atau koneksi server berkas.</p>
              <button v-if="foto" type="button" class="clinical-button secondary" :disabled="terkunci" @click="batalFoto">
                <X :size="15" /> Batalkan Pilihan Foto
              </button>
            </template>
            <p v-if="editing?.data.foto && !foto" class="edukasi-catatan">Foto tersimpan tetap digunakan. Pilih file untuk menggantinya.</p>
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
          <button type="button" class="clinical-button secondary" :disabled="terkunci || printing || !rows.length"
            title="Cetak semua catatan yang sesuai filter" @click="cetak()">
            <LoaderCircle v-if="cetakAktif === 'semua'" class="spin" :size="15" />
            <Printer v-else :size="15" /> {{ cetakAktif === 'semua' ? 'Menyiapkan...' : 'Cetak Hasil Filter' }}
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
      <DataTable v-else class="edukasi-table" :rows="rows" data-key="kunci" paginator :rows-per-page="10"
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
          <template #body="{ data: r }"><div class="edukasi-teks-ringkas">{{ r.data[bidang.key] || '—' }}</div></template>
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
        <Column header="Pemahaman / Verifikasi" style="min-width:200px">
          <template #body="{ data: r }">
            <div class="clinical-table-main">
              <strong>{{ r.data.status_verifikasi || 'Belum dicatat' }}</strong>
              <span>{{ r.data.tingkat_pemahaman || 'Pemahaman belum dinilai' }}</span>
              <span v-if="r.data.tanggal_verifikasi">{{ r.data.tanggal_verifikasi }} WITA</span>
              <span v-if="r.data.nip_verifikator">{{ r.nama_verifikator || r.data.nip_verifikator }}</span>
            </div>
          </template>
        </Column>
        <Column header="Aksi" style="min-width:210px">
          <template #body="{ data: r }">
            <div class="clinical-table-actions edukasi-actions" role="group"
              :aria-label="`Aksi edukasi ${r.data.tgl_perawatan} ${r.data.jam_rawat}`">
              <button type="button" title="Lihat detail edukasi" @click="detailCatatan = r">
                <Eye :size="15" aria-hidden="true" /> Detail
              </button>
              <button type="button" :disabled="!r.foto_url"
                :title="r.foto_url ? 'Lihat foto edukasi' : r.data.foto ? 'Foto tidak tersedia' : 'Belum ada foto'"
                @click="detail = r">
                <Image :size="15" aria-hidden="true" /> Foto
              </button>
              <button v-if="r.bisa_ubah" type="button" title="Edit catatan edukasi" :disabled="terkunci" @click="edit(r)">
                <Pencil :size="15" aria-hidden="true" /> Edit
              </button>
              <button v-if="r.bisa_ubah" type="button" class="danger" title="Hapus catatan edukasi" :disabled="terkunci"
                @click="hapusTarget = r; errorSimpan = ''">
                <Trash2 :size="15" aria-hidden="true" /> Hapus
              </button>
              <button type="button" class="edukasi-cetak" title="Cetak catatan edukasi ini / simpan sebagai PDF"
                :disabled="terkunci || printing" @click="cetak(r)">
                <LoaderCircle v-if="cetakAktif === r.kunci" class="spin" :size="15" aria-hidden="true" />
                <Printer v-else :size="15" aria-hidden="true" />
                {{ cetakAktif === r.kunci ? 'Menyiapkan...' : 'Cetak' }}
              </button>
            </div>
            <span v-if="!r.bisa_ubah" class="clinical-owner-note edukasi-action-note">Hanya baca</span>
            <span v-if="!r.foto_url" class="edukasi-action-note">{{ r.data.foto ? 'Foto tidak tersedia' : 'Tanpa foto' }}</span>
          </template>
        </Column>
      </DataTable>
    </article>

    <Dialog :visible="!!detailCatatan" modal header="Detail Edukasi Pasien" :style="{ width: '860px', maxWidth: '95vw' }"
      @update:visible="!$event && (detailCatatan = null)">
      <template v-if="detailCatatan">
        <p>Petugas: {{ detailCatatan.nama_petugas || detailCatatan.data.nip }} · Ruangan: {{ detailCatatan.nama_ruangan || detailCatatan.data.kd_ruangan || '-' }}</p>
        <dl class="edukasi-detail-grid">
          <div v-for="bidang in bidangDetail" :key="bidang.key">
            <dt>{{ bidang.label }}</dt>
            <dd>{{ detailCatatan.data[bidang.key] || 'Belum dicatat' }}</dd>
          </div>
          <div><dt>Nama Verifikator</dt><dd>{{ detailCatatan.nama_verifikator || 'Belum dicatat' }}</dd></div>
        </dl>
      </template>
    </Dialog>

    <Dialog :visible="!!detail" modal header="Foto Edukasi Pasien" :style="{ width: '760px', maxWidth: '95vw' }"
      @update:visible="!$event && (detail = null)">
      <p>{{ detail?.data.tgl_perawatan }} {{ detail?.data.jam_rawat }} WITA · {{ detail?.nama_petugas }}</p>
      <img v-if="detail?.foto_url && !detailFotoGagal" :src="detail.foto_url" alt="Foto dokumentasi edukasi pasien"
        class="edukasi-foto-detail" @error="detailFotoGagal = true" />
      <p v-else role="alert">Foto tidak dapat dimuat. Periksa koneksi server berkas.</p>
    </Dialog>

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
