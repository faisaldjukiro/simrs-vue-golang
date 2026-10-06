<script setup lang="ts">
import { BarChart3, Filter, Printer } from '@lucide/vue'
import Column from 'primevue/column'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import ObservasiRanapPage from '../ObservasiRanap/ObservasiRanapPage.vue'
import { parameterKardeks } from '../../../types/kardeks'
import type { PropsKardeks } from '../../../types/kardeks'
import { useKardeks } from './useKardeks'
import { computed } from 'vue'
import CairanForm from './CairanForm.vue'
import { rekapCairan } from './useCairan'

const props = defineProps<PropsKardeks>()
const {
  loading, printing, inputVisible, observasiBerubah, error, tanggal, jam, hasil, tab, sumber, grafikVisible, metrik,
  perluMuat, adaError, observasi, slots, bagianAktif, judulGrafik, grafik, muat, cetak,
} = useKardeks(props)
const cairan = computed(() => hasil.value ? rekapCairan(hasil.value) : null)
const tabs = [
  { kode: 'observasi', nama: 'Observasi & Grafik' },
  { kode: 'terapi', nama: 'Resep & Transaksi Obat' },
  { kode: 'instruksi', nama: 'Instruksi & Catatan' },
  { kode: 'penunjang', nama: 'Laboratorium & Radiologi' },
  { kode: 'ventilator', nama: 'Ventilator' },
  { kode: 'cairan', nama: 'Cairan & Pemantauan Lain' },
]
const sumberOptions = [
  { label: 'Catatan Observasi Rawat Inap', value: 'observasi' },
  { label: 'Pemeriksaan Rawat Inap / CPPT', value: 'pemeriksaan' },
  { label: 'Monitoring Ventilator', value: 'monitoring_ventilator' },
]
const metrikOptions = parameterKardeks.filter(p => !['td', 'gcs'].includes(p.key))
  .map(p => ({ label: p.label + ' (' + p.unit + ')', value: p.key }))
</script>

<template>
  <section class="clinical-page kardeks-page">
    <div class="kardeks-filter">
      <button type="button" class="clinical-button primary" :aria-expanded="inputVisible"
        aria-controls="kardeks-input-observasi" @click="inputVisible = !inputVisible">
        {{ inputVisible ? 'Tutup Input Observasi' : 'Input / Edit Observasi' }}
      </button>
    </div>
    <section v-if="inputVisible" id="kardeks-input-observasi">
      <p>Form menggunakan Catatan Observasi Ranap, bukan CPPT. Pastikan tanggal dan jam catatan benar.
        Ringkasan tetap mengikuti periode kardeks yang dipilih.</p>
      <ObservasiRanapPage :token="token" :patient="patient" @berubah="observasiBerubah" />
    </section>
    <article class="clinical-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Pelayanan Rawat Inap · Ringkasan Kardeks</span>
          <h3>Kardeks ICU</h3>
          <p>Observasi dapat diisi melalui tombol di atas. Resep, instruksi, dan hasil penunjang tetap mengikuti modul asal.</p>
        </div>
        <button class="clinical-button secondary" type="button"
          :disabled="!hasil || loading || printing || perluMuat || adaError" @click="cetak">
          <Printer :size="16" /> {{ printing ? 'Menyiapkan...' : 'Cetak Kardeks' }}
        </button>
      </header>
      <form class="kardeks-filter" @submit.prevent="muat">
        <FormInput v-model="tanggal" label="Tanggal Mulai" type="date" required />
        <FormInput v-model="jam" label="Jam Mulai (WITA)" type="time" required />
        <button class="clinical-button primary" type="submit" :disabled="loading">
          <Filter :size="16" /> {{ loading ? 'Memuat...' : 'Tampilkan 24 Jam' }}
        </button>
      </form>
      <div class="kardeks-content">
        <p v-if="error" class="patient-error" role="alert">{{ error }}</p>
        <p v-if="loading" role="status">Menarik data kardeks...</p>
        <template v-if="hasil && !loading">
          <p>Periode: {{ hasil.mulai }} sampai sebelum {{ hasil.selesai }} WITA.</p>
          <p v-if="perluMuat" class="patient-error">Filter berubah. Tekan Tampilkan untuk memuat periode baru.</p>
          <p v-if="adaError" class="patient-error" role="alert">Sebagian sumber gagal dimuat. Cetak dinonaktifkan agar hasil tidak dianggap lengkap.</p>
          <nav class="visit-report-tabs" aria-label="Bagian kardeks">
            <button v-for="t in tabs" :key="t.kode" type="button" :class="{ active: tab === t.kode }" @click="tab = t.kode">{{ t.nama }}</button>
          </nav>
          <section v-if="tab === 'observasi'">
            <div class="kardeks-filter">
              <FormInput v-model="sumber" label="Sumber Observasi" jenis="select" :options="sumberOptions" />
              <button class="clinical-button secondary" type="button" :aria-expanded="grafikVisible" @click="grafikVisible = !grafikVisible">
                <BarChart3 :size="16" /> {{ grafikVisible ? 'Tutup Grafik' : 'Lihat Grafik' }}
              </button>
            </div>
            <p>Semua pencatatan pada setiap interval ditampilkan dengan jam asli. Tanda — berarti belum tercatat, bukan nol.</p>
            <p v-if="observasi?.error" class="patient-error">{{ observasi.error }}</p>
            <template v-else>
              <section v-if="grafikVisible" class="kardeks-grafik">
                <FormInput v-model="metrik" label="Parameter Grafik" jenis="select" :options="metrikOptions" />
                <p>{{ judulGrafik?.label }} ({{ judulGrafik?.unit }}) · {{ observasi?.nama }}. Titik sesuai waktu pencatatan, tanpa interpolasi.</p>
                <svg v-if="grafik.titik.length" viewBox="0 0 800 220" role="img" :aria-label="'Grafik ' + judulGrafik?.label">
                  <line x1="55" y1="25" x2="55" y2="180" stroke="currentColor" />
                  <line x1="55" y1="180" x2="745" y2="180" stroke="currentColor" />
                  <text x="5" y="38">{{ grafik.max }}</text>
                  <text x="5" y="175">{{ grafik.min }}</text>
                  <text x="55" y="207">{{ hasil.mulai.slice(11, 16) }}</text>
                  <text x="650" y="207">+24 jam</text>
                  <circle v-for="(t, i) in grafik.titik" :key="i" :cx="t.x" :cy="t.y" r="4" fill="currentColor">
                    <title>{{ t.waktu }} · {{ t.nilai }} {{ judulGrafik?.unit }}</title>
                  </circle>
                </svg>
                <p v-else>Belum ada nilai numerik untuk grafik. Nilai asli tetap tersedia di tabel.</p>
              </section>
              <div class="kardeks-matrix" tabindex="0" aria-label="Tabel observasi 24 jam, geser untuk melihat interval lainnya">
                <table>
                  <thead>
                    <tr>
                      <th>Parameter</th>
                      <th v-for="(s, i) in slots" :key="i">{{ s.label }}<small>{{ s.tanggal }}</small></th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="p in parameterKardeks" :key="p.key">
                      <th>{{ p.label }}<small>{{ p.unit }}</small></th>
                      <td v-for="(s, i) in slots" :key="i">
                        <template v-if="s.baris.length">
                          <div v-for="(r, n) in s.baris" :key="n" :title="'Petugas: ' + r.petugas">
                            <small>{{ r.waktu.slice(11) }}</small>
                            <strong>{{ r[p.key] || '—' }}</strong>
                          </div>
                        </template>
                        <span v-else>—</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </template>
          </section>
          <section v-if="tab === 'cairan'">
            <CairanForm :token="token" :patient="patient" @berubah="muat" />
            <h4>Rekap Cairan Tercatat — 24 Jam</h4>
            <p>Hanya catatan yang seluruh periode pengukurannya berada dalam periode Kardeks. Belum dicatat bukan nol. Rekap bukan konfirmasi kelengkapan pencatatan.</p>
            <p v-if="cairan">
              Masuk: <strong>{{ cairan.masuk ?? '—' }} mL</strong> ·
              Keluar: <strong>{{ cairan.keluar ?? '—' }} mL</strong> ·
              Selisih tercatat (tanpa IWL): <strong>{{ cairan.balance ?? '—' }} mL</strong>
            </p>
            <p v-if="cairan?.lintas" class="patient-error">
              {{ cairan.lintas }} catatan melintasi batas periode dan tidak dijumlahkan. Volume tidak dibagi rata otomatis. Lihat periode pengukuran pada tabel.
            </p>
            <p>CVP dan ICP belum tersedia. Rekap per shift belum diaktifkan karena jam pergantian shift perlu ditetapkan.</p>
          </section>
          <section v-for="b in bagianAktif" :key="b.kode" class="kardeks-bagian">
            <h4>{{ b.nama }}</h4>
            <p v-if="b.kode === 'obat'">Jumlah merupakan jumlah transaksi, bukan dosis. Jam transaksi belum dikonfirmasi sebagai jam pemberian ke pasien.</p>
            <p v-if="b.kode === 'resep'">Menampilkan resep nonracikan yang ditulis dalam periode ini, bukan seluruh terapi aktif atau bukti pemberian.</p>
            <p v-if="b.error" class="patient-error">{{ b.error }}</p>
            <DataTable v-else :rows="b.baris" paginator :rows-per-page="10" empty-message="Belum ada catatan pada periode ini." class="visit-report-table">
              <Column v-for="k in Object.keys(b.baris[0] || {})" :key="k" :field="k" :header="k.replaceAll('_', ' ')" style="min-width:150px">
                <template #body="{ data: r }"><span class="kardeks-teks">{{ r[k] || '—' }}</span></template>
              </Column>
            </DataTable>
          </section>
        </template>
      </div>
    </article>
  </section>
</template>

<style src="../../../Components/Ui/report.css" scoped></style>
<style src="./kardeks.css" scoped></style>
