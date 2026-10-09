<script setup lang="ts">
import {
  Activity, ArrowUpRight, BookOpen, ChevronDown, ClipboardList,
  FileText, FlaskConical, FolderOpen, House, Info, Pill,
  RefreshCw, ScanLine, ShieldAlert, X,
} from '@lucide/vue'
import Column from 'primevue/column'
import DataTable from '../../../Components/Ui/DataTable.vue'
import type { PropsRingkasanPasien } from '../../../types/ringkasanPasien'
import { useRingkasan } from './useRingkasan'

const props = defineProps<PropsRingkasanPasien>()
const emit = defineEmits<{ buka: [kode: string] }>()
const {
  loading, error, data, diperbarui, sep, alergi, vital, kelompokKartu, hasilAktif,
  detailHasil, kolomHasil, detailPermintaan, kolomPermintaan, muat, bolehBuka, tampilNilai,
  penunjangAktif, judulDetail, detailObat, bukaDetail, tutupDetail,
} = useRingkasan(props)
const ikonKartu = {
  cppt: ClipboardList,
  diagnosis: Activity,
  resume: FileText,
  berkas: FolderOpen,
  lab: FlaskConical,
  radiologi: ScanLine,
  resep: FileText,
  obat: Pill,
  edukasi: BookOpen,
  pemulangan: House,
}
const parameterVital = [
  { kode: 'tensi', nama: 'Tekanan darah', satuan: 'mmHg' },
  { kode: 'nadi', nama: 'Nadi', satuan: '/menit' },
  { kode: 'respirasi', nama: 'Respirasi', satuan: '/menit' },
  { kode: 'suhu', nama: 'Suhu', satuan: '°C' },
  { kode: 'spo2', nama: 'SpO₂', satuan: '%' },
  { kode: 'gcs', nama: 'GCS', satuan: '' },
]
</script>

<template>
  <section class="ringkasan-pasien clinical-page" :aria-busy="loading" aria-label="Monitoring catatan dan berkas kunjungan">
    <article class="clinical-history-card ringkasan-intro">
      <header class="clinical-section-header">
        <div>
          <span>Ringkasan kunjungan · {{ moduleName }}</span>
          <h3>Catatan &amp; layanan pasien</h3>
          <p>Dokumentasi, hasil pemeriksaan, dan obat dalam satu tampilan.</p>
        </div>
        <button class="clinical-button secondary" type="button" :disabled="loading" @click="muat">
          <RefreshCw :size="16" /> {{ loading ? 'Memuat...' : 'Muat Ulang' }}
        </button>
      </header>
      <p v-if="loading" class="ringkasan-state" role="status">Memeriksa catatan kunjungan...</p>
      <p v-else-if="error" class="patient-error ringkasan-state" role="alert">{{ error }}</p>
      <div v-else-if="data" class="ringkasan-meta">
        <p v-if="sep?.error" class="patient-error" role="alert">SEP: {{ sep.error }}</p>
        <p v-else><span>No. SEP</span> {{ sep?.baris.length ? sep.baris.map(r => r.no_sep).join(' · ') : 'Belum tercatat' }}</p>
        <small v-if="diperbarui">Diperbarui {{ diperbarui }} WITA</small>
      </div>
    </article>

    <template v-if="data && !loading">
      <p class="ringkasan-petunjuk">
        <Info :size="16" aria-hidden="true" />
        Ada catatan belum berarti lengkap atau tervalidasi. Kebutuhan berkas mengikuti layanan pasien.
      </p>
      <section v-for="kelompok in kelompokKartu" :key="kelompok.kode" class="ringkasan-kelompok" :aria-labelledby="'ringkasan-' + kelompok.kode">
        <header class="ringkasan-kelompok-header">
          <h3 :id="'ringkasan-' + kelompok.kode">{{ kelompok.judul }}</h3>
          <p>{{ kelompok.deskripsi }}</p>
        </header>
        <div class="ringkasan-monitoring" :class="{ 'ringkasan-enam-kartu': kelompok.kartu.length === 6 }">
          <article
            v-for="k in kelompok.kartu"
            :key="k.kode"
            class="clinical-history-card ringkasan-card ringkasan-kartu"
            :class="{ 'ringkasan-kartu-aktif': hasilAktif === k.kode }"
          >
            <div class="ringkasan-kartu-header">
              <span class="ringkasan-ikon"><component :is="ikonKartu[k.kode]" :size="20" aria-hidden="true" /></span>
              <h4>{{ k.judul }}</h4>
            </div>
            <div class="ringkasan-metrik">
              <p class="ringkasan-jumlah"><strong>{{ k.jumlah ?? '—' }}</strong><span>{{ k.unit }}</span></p>
              <span class="ringkasan-status" :class="{ tersedia: k.jumlah !== null && k.jumlah > 0 }">{{ k.status }}</span>
            </div>
            <p v-if="k.error" class="patient-error" role="alert">{{ k.error }}</p>
            <div v-if="['lab', 'radiologi'].includes(k.kode)" class="ringkasan-progres">
              <template v-if="k.progres">
                <strong>{{ k.progres.total }} permintaan</strong>
                <div class="ringkasan-progres-item">
                  <span><b>{{ k.progres.menunggu }}</b> Menunggu</span>
                  <span><b>{{ k.progres.diproses }}</b> Diproses</span>
                  <span><b>{{ k.progres.hasilDicatat }}</b> Tanggal hasil tercatat</span>
                </div>
              </template>
              <span v-else>Progres permintaan belum dapat diperiksa.</span>
            </div>
            <details v-if="!k.error && k.sumber" class="ringkasan-info-sumber">
              <summary>Informasi data</summary>
              <p>{{ k.sumber }}</p>
            </details>
            <button
              v-if="k.menu"
              class="clinical-button secondary"
              type="button"
              :disabled="!bolehBuka(k.menu)"
              :title="bolehBuka(k.menu) ? k.tombol : 'Menu belum tersedia pada sidebar layanan ini'"
              @click="emit('buka', k.menu)"
            >{{ k.tombol }} <ArrowUpRight :size="15" aria-hidden="true" /></button>
            <button
              v-else
              :id="'ringkasan-buka-' + k.kode"
              class="clinical-button secondary"
              type="button"
              :aria-expanded="hasilAktif === k.kode"
              aria-controls="ringkasan-panel-hasil"
              @click="bukaDetail(k.kode)"
            >{{ hasilAktif === k.kode ? 'Tutup Detail' : k.tombol }} <ChevronDown :size="15" aria-hidden="true" /></button>
          </article>
        </div>
      </section>

      <article v-if="hasilAktif" id="ringkasan-panel-hasil" class="clinical-history-card ringkasan-card ringkasan-detail" tabindex="-1" aria-labelledby="ringkasan-judul-detail">
        <div class="ringkasan-judul">
          <h3 id="ringkasan-judul-detail">{{ judulDetail }}</h3>
          <button class="clinical-button secondary" type="button" @click="tutupDetail"><X :size="16" /> Tutup Detail</button>
        </div>
        <template v-if="penunjangAktif">
          <h4>Progres Permintaan</h4>
          <p>{{ detailPermintaan?.sumber }}</p>
          <p v-if="!detailPermintaan || detailPermintaan.error" class="patient-error" role="alert">
            {{ detailPermintaan?.error || 'Data permintaan belum tersedia. Muat ulang atau periksa versi backend.' }}
          </p>
          <DataTable
            v-else
            :key="'permintaan-' + hasilAktif"
            :rows="detailPermintaan.baris"
            data-key="noorder"
            class="visit-report-table"
            paginator
            :rows-per-page="5"
            :rows-per-page-options="[5, 10, 25]"
            empty-message="Belum ada permintaan tercatat pada sumber ini."
          >
            <Column
              v-for="kolom in kolomPermintaan"
              :key="kolom.field"
              :field="kolom.field"
              :header="kolom.header"
              style="min-width:150px"
            >
              <template #body="{ data: row }">
                <span class="ringkasan-teks">{{ row[kolom.field] || '—' }}</span>
              </template>
            </Column>
          </DataTable>
          <p class="ringkasan-catatan">
            Progres mengikuti tanggal pada permintaan. Tanggal hasil tercatat belum menjamin seluruh
            pemeriksaan selesai atau tervalidasi. Hasil di bawah ditampilkan per kunjungan, tidak
            otomatis dipasangkan ke setiap permintaan.
          </p>
          <h4>Hasil yang Tersedia</h4>
          <p>{{ detailHasil?.sumber }}</p>
          <p v-if="!detailHasil || detailHasil.error" class="patient-error" role="alert">
            {{ detailHasil?.error || 'Data hasil belum dapat diperiksa.' }}
          </p>
          <DataTable
            v-else
            :key="hasilAktif"
            :rows="detailHasil?.baris || []"
            data-key=""
            class="visit-report-table"
            paginator
            :rows-per-page="10"
            :rows-per-page-options="[10, 25, 50]"
            empty-message="Belum ada hasil tercatat pada kunjungan ini."
          >
            <Column v-for="kolom in kolomHasil" :key="kolom.field" :field="kolom.field" :header="kolom.header" style="min-width:150px">
              <template #body="{ data: row }">
                <span class="ringkasan-teks">{{ row[kolom.field] || '—' }}</span>
              </template>
            </Column>
          </DataTable>
          <p>Nilai dan keterangan ditampilkan sesuai catatan SIMRS, tanpa interpretasi otomatis.</p>
        </template>
        <template v-else>
          <p class="ringkasan-catatan">
            Resep dan transaksi obat ditampilkan terpisah, bukan daftar terapi aktif atau bukti
            obat telah diberikan ke pasien. Jumlah tidak dijumlahkan antarobat atau antarsatuan.
          </p>
          <section v-for="detail in detailObat" :key="detail.kode">
            <h4>{{ detail.judul }}</h4>
            <p>{{ detail.bagian?.sumber }}</p>
            <p v-if="!detail.bagian || detail.bagian.error" class="patient-error" role="alert">
              {{ detail.bagian?.error || 'Sumber obat belum dapat diperiksa. Muat ulang atau periksa versi backend.' }}
            </p>
            <DataTable
              v-else
              :rows="detail.bagian.baris"
              data-key=""
              class="visit-report-table"
              paginator
              :rows-per-page="10"
              :rows-per-page-options="[10, 25, 50]"
              empty-message="Belum ada catatan pada sumber ini untuk kunjungan ini."
            >
              <Column
                v-for="kolom in detail.kolom"
                :key="kolom.field"
                :field="kolom.field"
                :header="kolom.header"
                :header-class="kolom.field === 'jumlah' ? 'right-align-header' : ''"
                :body-class="kolom.field === 'jumlah' ? 'right-align-body' : ''"
                style="min-width:150px"
              >
                <template #body="{ data: row }">
                  <span class="ringkasan-teks">{{ row[kolom.field] || '—' }}</span>
                </template>
              </Column>
            </DataTable>
          </section>
        </template>
      </article>

      <details class="clinical-history-card ringkasan-detail ringkasan-card">
        <summary>Informasi klinis tambahan · alergi &amp; tanda vital</summary>
        <div class="ringkasan-grid">
          <section>
            <h3><ShieldAlert :size="20" /> Catatan Alergi</h3>
            <p class="ringkasan-sumber">{{ alergi?.sumber }}. Bukan daftar alergi lengkap lintas kunjungan.</p>
            <p v-if="alergi?.error" class="patient-error" role="alert">{{ alergi.error }}</p>
            <ul v-else-if="alergi?.baris.length" class="ringkasan-list ringkasan-scroll">
              <li v-for="(r, i) in alergi.baris" :key="i">
                <strong class="ringkasan-teks">{{ r.alergi }}</strong>
                <small>{{ r.waktu }} WITA · {{ r.nama_petugas || r.petugas }}</small>
              </li>
            </ul>
            <p v-else>Belum tercatat pada sumber ini; bukan berarti pasien tidak memiliki alergi.</p>
          </section>
          <section>
            <h3><Activity :size="20" /> Tanda Vital Terakhir di CPPT</h3>
            <p class="ringkasan-sumber">{{ vital?.sumber }}. Tidak digabung dengan catatan sebelumnya atau modul observasi.</p>
            <p v-if="vital?.error" class="patient-error" role="alert">{{ vital.error }}</p>
            <template v-else-if="vital?.baris.length">
              <section v-for="(r, i) in vital.baris" :key="i" class="ringkasan-vital">
                <p><strong>{{ r.waktu }} WITA</strong><br /><small>{{ r.nama_petugas || r.petugas }}</small></p>
                <dl class="ringkasan-angka">
                  <div v-for="p in parameterVital" :key="p.kode">
                    <dt>{{ p.nama }} <small>{{ p.satuan }}</small></dt>
                    <dd>{{ tampilNilai(r[p.kode]) }}</dd>
                  </div>
                </dl>
                <p>Kesadaran: {{ tampilNilai(r.kesadaran) }}</p>
              </section>
            </template>
            <p v-else>Belum tercatat pada CPPT layanan ini.</p>
          </section>
        </div>
      </details>
    </template>
  </section>
</template>

<style src="../../../Components/Ui/report.css" scoped></style>
<style src="./ringkasan.css" scoped></style>
