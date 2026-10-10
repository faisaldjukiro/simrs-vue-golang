<script setup lang="ts">
import { ChevronDown, ChevronUp, Eye, Image, LoaderCircle, Pencil, Printer, RefreshCw, Save, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import InputPencarian from '../../../Components/Ui/InputPencarian.vue'
import TableSearch from '../../../Components/Ui/TableSearch.vue'
import { bidangEdukasi, bidangAsesmen, pilihanPelaksanaan, pilihanMateri, pilihanPemahaman, pilihanStatusVerifikasi, type PropsEdukasi } from '../../../types/edukasiPasien'
import { useEdukasiPasien } from './useEdukasiPasien'
import ParafEdukasi from './ParafEdukasi.vue'
import KameraEdukasi from './KameraEdukasi.vue'

const props = defineProps<PropsEdukasi>()
const {
  loading, saving, error, errorSimpan, keyword, rows, records, formVisible, editing,
  hapusTarget, petugas, ruangan, bolehPilihPetugas, form, terkunci, mulai, selesai, errorFilter,
  waktuSekarang, reset, muat, cariReferensi, edit, mutasi, cetak, printing, cetakAktif,
  fotoPetugas, fotoPenerima, kameraAktif, detail, detailFotoGagal, detailFotoPenerimaGagal,
  fotoPenerimaTersedia, namaPenerimaTersedia,
  detailCatatan, bidangDetail, verifikator, terverifikasi, ubahStatusVerifikasi, waktuVerifikasiSekarang,
  pilihanBidang, nilaiPilihan, pilihBidang, materiTerpilih, pilihMateri,
  parafTersedia, parafPenerimaTersedia, parafDibatalkan, kekuranganVerifikasi,
  jenisBukti, pilihJenisBukti, buktiGanda,
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
            <div v-for="bidang in pilihanPelaksanaan" :key="bidang.key" class="edukasi-pilihan-bidang">
              <FormInput :model-value="nilaiPilihan(bidang)" :label="bidang.label" jenis="select"
                :options="pilihanBidang(bidang)" :disabled="terkunci"
                @update:model-value="pilihBidang(bidang, $event)" />
              <FormInput v-if="nilaiPilihan(bidang) === '__lainnya__'" v-model="form[bidang.key]"
                :label="bidang.label + ' Lainnya'" :maxlength="bidang.maxlength"
                :disabled="terkunci" />
            </div>
          </div>
          <fieldset class="edukasi-pilihan-materi" :disabled="terkunci">
            <legend>Materi Edukasi yang Diberikan</legend>
            <p class="edukasi-petunjuk">Boleh pilih lebih dari satu. Materi lain atau penjelasan tambahan dapat ditulis pada isian Materi Edukasi.</p>
            <div class="edukasi-checkbox-grid">
              <label v-for="materi in pilihanMateri" :key="materi" class="edukasi-checkbox">
                <input type="checkbox" :checked="materiTerpilih.includes(materi)" :disabled="terkunci"
                  @change="pilihMateri(materi, ($event.target as HTMLInputElement).checked)" />
                <span>{{ materi }}</span>
              </label>
            </div>
          </fieldset>
          <div class="edukasi-materi-grid">
            <FormInput v-model="form.materi" label="Materi Edukasi" jenis="textarea" :rows="5" :maxlength="65535"
              :disabled="terkunci" />
            <FormInput v-model="form.keterangan" label="Keterangan" jenis="textarea" :rows="5" :maxlength="255"
              :disabled="terkunci" />
          </div>
          <section class="edukasi-bagian" aria-labelledby="asesmen-edukasi-title">
            <h4 id="asesmen-edukasi-title" class="edukasi-subjudul">Asesmen Kebutuhan Belajar</h4>
            <p class="edukasi-petunjuk">Isi berdasarkan penerima edukasi. Asesmen wajib lengkap sebelum verifikasi. Bila belum dapat dinilai, jelaskan alasannya melalui pilihan Lainnya.</p>
            <div class="edukasi-identitas-grid">
              <div v-for="bidang in bidangAsesmen" :key="bidang.key" class="edukasi-pilihan-bidang">
                <FormInput :model-value="nilaiPilihan(bidang)" :label="bidang.label" jenis="select"
                  :options="pilihanBidang(bidang)" :disabled="terkunci"
                  @update:model-value="pilihBidang(bidang, $event)" />
                <FormInput v-if="nilaiPilihan(bidang) === '__lainnya__'" v-model="form[bidang.key]"
                  :label="bidang.label + ' Lainnya'" :jenis="bidang.jenis || 'input'"
                  :rows="3" :maxlength="bidang.maxlength" :disabled="terkunci" />
              </div>
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
              jenis="textarea" :rows="4" :maxlength="65535" :required="terverifikasi" :disabled="terkunci"
              hint="Catat cara memeriksa pemahaman, respons penerima, hasil penilaian, dan rencana edukasi ulang bila belum memahami." />
            <button v-if="terverifikasi" type="button" class="clinical-button secondary" :disabled="terkunci" @click="waktuVerifikasiSekarang">
              <RefreshCw :size="15" /> Waktu Verifikasi Sekarang
            </button>
            <p class="edukasi-petunjuk">Verifikator mengikuti petugas login. Administrator dapat memilih petugas verifikator.</p>
          </section>
          <section class="edukasi-bagian" aria-labelledby="bukti-edukasi-title">
            <h4 id="bukti-edukasi-title" class="edukasi-subjudul">Bukti Edukasi</h4>
            <FormInput :model-value="jenisBukti" label="Jenis Bukti" jenis="select" :disabled="terkunci"
              :options="[{ label: 'Belum dipilih', value: '' }, { label: 'Paraf petugas dan penerima', value: 'paraf' }, { label: 'Foto petugas dan penerima', value: 'foto' }]"
              @update:model-value="pilihJenisBukti" />
            <p class="edukasi-petunjuk">Pilih dua paraf atau dua foto: petugas dan penerima edukasi. Bukti wajib lengkap saat verifikasi. Mengganti jenis bukti mengosongkan bukti sebelumnya pada form; perubahan berlaku setelah disimpan.</p>
            <p v-if="buktiGanda" class="patient-error" role="alert">Catatan ini memuat paraf dan foto. Pilih satu jenis bukti sebelum menyimpan atau mencetak ulang.</p>
            <FormInput v-if="jenisBukti" v-model="form.nama_penerima" label="Nama Penerima Edukasi" :maxlength="100"
              :required="terverifikasi" :disabled="terkunci || !namaPenerimaTersedia"
              hint="Nama pasien atau keluarga yang menerima edukasi." />
          <div v-if="jenisBukti === 'foto'" class="edukasi-foto">
            <p v-if="!loading && !error && (!fotoPenerimaTersedia || !namaPenerimaTersedia)" class="patient-error" role="alert">
              Penyimpanan foto penerima belum tersedia. Hubungi administrator untuk melengkapi kolom foto_penerima dan nama_penerima.
            </p>
            <div class="edukasi-identitas-grid">
              <KameraEdukasi label="Foto Petugas Pemberi Edukasi" :nama="petugas.nama || ''"
                :pratinjau="fotoPetugas.pratinjau" :baru="!!fotoPetugas.file" :aktif="kameraAktif === 'petugas'"
                :disabled="terkunci || !petugas.kode" @buka="kameraAktif = 'petugas'"
                @tutup="kameraAktif === 'petugas' && (kameraAktif = '')" @tangkap="fotoPetugas.tangkap" @batal="fotoPetugas.batal" />
              <KameraEdukasi label="Foto Penerima Edukasi" :nama="form.nama_penerima || ''"
                :pratinjau="fotoPenerima.pratinjau" :baru="!!fotoPenerima.file" :aktif="kameraAktif === 'penerima'"
                :disabled="terkunci || !fotoPenerimaTersedia || !form.nama_penerima?.trim()" @buka="kameraAktif = 'penerima'"
                @tutup="kameraAktif === 'penerima' && (kameraAktif = '')" @tangkap="fotoPenerima.tangkap" @batal="fotoPenerima.batal" />
            </div>
            <p class="edukasi-petunjuk">Izinkan akses kamera saat diminta browser. Foto tersimpan tetap digunakan sampai diganti dan catatan disimpan.</p>
          </div>
          <section v-if="jenisBukti === 'paraf'" class="edukasi-bagian" aria-labelledby="paraf-edukasi-title">
            <h4 id="paraf-edukasi-title" class="edukasi-subjudul">Paraf Petugas dan Penerima Edukasi</h4>
            <p v-if="!loading && !error && (!parafTersedia || !parafPenerimaTersedia)" class="patient-error" role="alert">
              Penyimpanan paraf petugas dan penerima belum lengkap. Hubungi administrator untuk mengaktifkannya atau pilih foto sebagai bukti edukasi.
            </p>
            <p class="edukasi-petunjuk">Lengkapi nama penerima dan seluruh isian sebelum membubuhkan kedua paraf. Perubahan catatan membatalkan kedua paraf sebelumnya.</p>
            <div class="edukasi-identitas-grid">
              <div>
                <h4 class="edukasi-subjudul">Petugas Pemberi Edukasi</h4>
                <ParafEdukasi v-model="form.paraf_petugas" :nama="petugas.nama || ''" label="Kotak paraf petugas pemberi edukasi"
                  :disabled="terkunci || !parafTersedia || !petugas.kode" />
              </div>
              <div>
                <h4 class="edukasi-subjudul">Penerima Edukasi</h4>
                <ParafEdukasi v-model="form.paraf_penerima" :nama="form.nama_penerima || ''" label="Kotak paraf penerima edukasi"
                  :disabled="terkunci || !parafPenerimaTersedia || !form.nama_penerima?.trim()" />
              </div>
            </div>
            <p v-if="parafDibatalkan && (!form.paraf_petugas || !form.paraf_penerima)" class="patient-error" role="status">Isi catatan berubah. Silakan gambar ulang paraf petugas dan penerima edukasi.</p>
          </section>
            <p v-if="terverifikasi && kekuranganVerifikasi.length" class="edukasi-petunjuk" role="status">
              Belum lengkap untuk verifikasi: {{ kekuranganVerifikasi.join(', ') }}.
            </p>
          </section>
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
              <button type="button" :disabled="!r.foto_url && !r.foto_penerima_url"
                title="Lihat foto petugas dan penerima edukasi"
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
          <div><dt>Paraf Petugas</dt><dd>
            <img v-if="detailCatatan.data.paraf_petugas?.startsWith('data:image/png;base64,')"
              :src="detailCatatan.data.paraf_petugas" alt="Paraf petugas pemberi edukasi" class="edukasi-paraf-detail" />
            <span v-else>Belum dibubuhkan</span>
          </dd></div>
          <div><dt>Paraf Penerima Edukasi</dt><dd>
            <img v-if="detailCatatan.data.paraf_penerima?.startsWith('data:image/png;base64,')"
              :src="detailCatatan.data.paraf_penerima" alt="Paraf penerima edukasi" class="edukasi-paraf-detail" />
            <span v-else>Belum dibubuhkan</span>
          </dd></div>
        </dl>
      </template>
    </Dialog>

    <Dialog :visible="!!detail" modal header="Foto Edukasi Pasien" :style="{ width: '760px', maxWidth: '95vw' }"
      @update:visible="!$event && (detail = null)">
      <p>{{ detail?.data.tgl_perawatan }} {{ detail?.data.jam_rawat }} WITA · {{ detail?.nama_petugas }}</p>
      <h4>Petugas Pemberi Edukasi — {{ detail?.nama_petugas || '-' }}</h4>
      <img v-if="detail?.foto_url && !detailFotoGagal" :src="detail.foto_url" alt="Foto petugas pemberi edukasi"
        class="edukasi-foto-detail" @error="detailFotoGagal = true" />
      <p v-else role="alert">Foto tidak dapat dimuat. Periksa koneksi server berkas.</p>
      <h4>Penerima Edukasi — {{ detail?.data.nama_penerima || '-' }}</h4>
      <img v-if="detail?.foto_penerima_url && !detailFotoPenerimaGagal" :src="detail.foto_penerima_url" alt="Foto penerima edukasi"
        class="edukasi-foto-detail" @error="detailFotoPenerimaGagal = true" />
      <p v-else>{{ detail?.data.foto_penerima && detail.data.foto_penerima !== '-' ? 'Foto penerima tidak dapat dimuat. Periksa koneksi server berkas.' : 'Foto penerima belum dilampirkan.' }}</p>
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
