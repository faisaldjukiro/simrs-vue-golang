<script setup>
import { ChevronDown, ChevronLeft, ChevronRight, ChevronUp, Filter, LoaderCircle, Search } from '@lucide/vue'
import DatePicker from 'primevue/datepicker'
import { computed, reactive, ref, watch } from 'vue'
import { riwayatPerawatanData } from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'
import TandaTanganVerifikasi from '../../Components/TandaTanganVerifikasi.vue'

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
})

const notifikasi = useNotifikasi()
const loading = ref(false)
const error = ref('')
const data = ref({ kunjungan: [], jumlah: 0 })
const mode = ref('nomor')
const tanggalMulai = ref(new Date())
const tanggalSelesai = ref(new Date())
const nomorRawat = ref('')
const terbuka = ref(new Set())
const pilihanTerbuka = ref(true)
const pilihanMode = [
  { value: 'terakhir', label: '5 Riwayat Terakhir' },
  { value: 'semua', label: 'Semua Riwayat' },
  { value: 'tanggal', label: 'Rentang Tanggal' },
  { value: 'nomor', label: 'Nomor Rawat' },
]

const pilihanKelompok = [
  { value: 'diagnosis', label: 'Diagnosa Penyakit' },
  { value: 'prosedur', label: 'Prosedur Tindakan' },
  { value: 'triase', label: 'Triase IGD', tables: ['data_triase_igd', 'data_triase_igdprimer', 'data_triase_igdsekunder', 'data_triase_igddetail_skala1', 'data_triase_igddetail_skala2', 'data_triase_igddetail_skala3', 'data_triase_igddetail_skala4', 'data_triase_igddetail_skala5'] },
  { value: 'pemeriksaan_ralan', label: 'Pemeriksaan Ralan' },
  { value: 'obstetri_ralan', label: 'Pemeriksaan Obstetri Ralan', tables: ['pemeriksaan_obstetri_ralan'] },
  { value: 'ginekologi_ralan', label: 'Pemeriksaan Ginekologi Ralan', tables: ['pemeriksaan_ginekologi_ralan'] },
  { value: 'pemeriksaan_ranap', label: 'Pemeriksaan Ranap' },
  { value: 'obstetri_ranap', label: 'Pemeriksaan Obstetri Ranap', tables: ['pemeriksaan_obstetri_ranap'] },
  { value: 'ginekologi_ranap', label: 'Pemeriksaan Ginekologi Ranap', tables: ['pemeriksaan_ginekologi_ranap'] },
  { value: 'catatan_dokter', label: 'Catatan Dokter' },
  { value: 'observasi_igd', label: 'Catatan Observasi IGD', tables: ['catatan_observasi_igd'] },
  { value: 'cek_gds', label: 'Catatan Cek GDS', tables: ['catatan_cek_gds'] },
  { value: 'observasi_ranap', label: 'Catatan Observasi Ranap', tables: ['catatan_observasi_ranap'] },
  { value: 'observasi_kebidanan', label: 'Observasi Ranap Kebidanan', tables: ['catatan_observasi_ranap_kebidanan'] },
  { value: 'observasi_postpartum', label: 'Observasi Ranap Postpartum', tables: ['catatan_observasi_ranap_postpartum'] },
  { value: 'catatan_keperawatan_ranap', label: 'Catatan Keperawatan Ranap', tables: ['catatan_keperawatan_ranap'] },
  { value: 'pews', label: 'Pemantauan PEWS Anak', tables: ['pemantauan_pews_anak'] },
  { value: 'askep_igd', label: 'Asuhan Keperawatan IGD', tables: ['penilaian_awal_keperawatan_igd', 'penilaian_awal_keperawatan_igd_masalah', 'penilaian_awal_keperawatan_ralan_rencana_igd'] },
  { value: 'askep_ralan', label: 'Asuhan Keperawatan Ralan', tables: ['penilaian_awal_keperawatan_ralan', 'penilaian_awal_keperawatan_ralan_masalah', 'penilaian_awal_keperawatan_ralan_rencana'] },
  { value: 'askep_gigi', label: 'Asuhan Keperawatan Gigi', tables: ['penilaian_awal_keperawatan_gigi', 'penilaian_awal_keperawatan_gigi_masalah', 'penilaian_awal_keperawatan_ralan_rencana_gigi'] },
  { value: 'askep_bayi', label: 'Asuhan Keperawatan Bayi', tables: ['penilaian_awal_keperawatan_ralan_bayi', 'penilaian_awal_keperawatan_ralan_bayi_masalah', 'penilaian_awal_keperawatan_ralan_rencana_anak'] },
  { value: 'askep_kandungan', label: 'Asuhan Keperawatan Kandungan', tables: ['penilaian_awal_keperawatan_kebidanan'] },
  { value: 'askep_psikiatri', label: 'Asuhan Keperawatan Psikiatri', tables: ['penilaian_awal_keperawatan_ralan_psikiatri', 'penilaian_awal_keperawatan_ralan_masalah_psikiatri', 'penilaian_awal_keperawatan_ralan_rencana_psikiatri'] },
  { value: 'asuhan_fisioterapi', label: 'Asuhan Fisioterapi', tables: ['penilaian_fisioterapi'] },
  { value: 'asuhan_psikolog', label: 'Asuhan Psikolog', tables: ['penilaian_psikologi'] },
  { value: 'askep_ranap', label: 'Asuhan Keperawatan Ranap', tables: ['penilaian_awal_keperawatan_ranap', 'penilaian_awal_keperawatan_ranap_masalah', 'penilaian_awal_keperawatan_ranap_rencana'] },
  { value: 'askep_ranap_kandungan', label: 'Asuhan Keperawatan Ranap Kandungan', tables: ['penilaian_awal_keperawatan_kebidanan_ranap'] },
  { value: 'medis_igd', label: 'Asuhan Medis IGD', tables: ['penilaian_medis_igd'] },
  { value: 'medis_ralan', label: 'Asuhan Medis Ralan', tables: ['penilaian_medis_ralan'] },
  { value: 'medis_ralan_kandungan', label: 'Asuhan Medis Ralan Kandungan', tables: ['penilaian_medis_ralan_kandungan'] },
  { value: 'medis_ralan_bayi', label: 'Asuhan Medis Ralan Bayi', tables: ['penilaian_medis_ralan_anak'] },
  { value: 'medis_ralan_tht', label: 'Asuhan Medis Ralan THT', tables: ['penilaian_medis_ralan_tht'] },
  { value: 'medis_ralan_psikiatri', label: 'Asuhan Medis Ralan Psikiatri', tables: ['penilaian_medis_ralan_psikiatrik'] },
  { value: 'medis_ralan_penyakit_dalam', label: 'Asuhan Medis Penyakit Dalam', tables: ['penilaian_medis_ralan_penyakit_dalam'] },
  { value: 'medis_ralan_mata', label: 'Asuhan Medis Mata', tables: ['penilaian_medis_ralan_mata'] },
  { value: 'medis_ralan_neurologi', label: 'Asuhan Medis Neurologi', tables: ['penilaian_medis_ralan_neurologi'] },
  { value: 'medis_ralan_orthopedi', label: 'Asuhan Medis Orthopedi', tables: ['penilaian_medis_ralan_orthopedi'] },
  { value: 'medis_ralan_bedah', label: 'Asuhan Medis Bedah', tables: ['penilaian_medis_ralan_bedah'] },
  { value: 'medis_ralan_geriatri', label: 'Asuhan Medis Geriatri', tables: ['penilaian_medis_ralan_geriatri'] },
  { value: 'medis_ranap', label: 'Asuhan Medis Ranap', tables: ['penilaian_medis_ranap'] },
  { value: 'medis_ranap_kandungan', label: 'Asuhan Medis Ranap Kandungan', tables: ['penilaian_medis_ranap_kandungan'] },
  { value: 'checklist_pre_operasi', label: 'Checklist Pre Operasi', tables: ['checklist_pre_operasi'] },
  { value: 'signin_anestesi', label: 'Sign In Sebelum Anestesi', tables: ['signin_sebelum_anestesi'] },
  { value: 'timeout_insisi', label: 'Time Out Sebelum Insisi', tables: ['timeout_sebelum_insisi'] },
  { value: 'signout_luka', label: 'Sign Out Menutup Luka', tables: ['signout_sebelum_menutup_luka'] },
  { value: 'checklist_post_operasi', label: 'Checklist Post Operasi', tables: ['checklist_post_operasi'] },
  { value: 'asuhan_pre_operasi', label: 'Asuhan Pre Operasi', tables: ['penilaian_pre_operasi'] },
  { value: 'asuhan_pre_anestesi', label: 'Asuhan Pre Anestesi', tables: ['penilaian_pre_anestesi'] },
  { value: 'risiko_jatuh_dewasa', label: 'Risiko Jatuh Dewasa', tables: ['penilaian_lanjutan_resiko_jatuh_dewasa'] },
  { value: 'risiko_jatuh_anak', label: 'Risiko Jatuh Anak', tables: ['penilaian_lanjutan_resiko_jatuh_anak'] },
  { value: 'tambahan_geriatri', label: 'Asuhan Tambahan Geriatri', tables: ['penilaian_tambahan_geriatri'] },
  { value: 'hasil_usg', label: 'Hasil Pemeriksaan USG', tables: ['hasil_pemeriksaan_usg'] },
  { value: 'pemulangan', label: 'Perencanaan Pemulangan', tables: ['perencanaan_pemulangan'] },
  { value: 'uji_kfr', label: 'Uji Fungsi KFR', tables: ['uji_fungsi_kfr'] },
  { value: 'hemodialisa', label: 'Hemodialisa', tables: ['hemodialisa'] },
  { value: 'skrining_nutrisi_dewasa', label: 'Skrining Nutrisi Dewasa', tables: ['skrining_nutrisi_dewasa'] },
  { value: 'skrining_nutrisi_lansia', label: 'Skrining Nutrisi Lansia', tables: ['skrining_nutrisi_lansia'] },
  { value: 'skrining_nutrisi_anak', label: 'Skrining Nutrisi Anak', tables: ['skrining_nutrisi_anak'] },
  { value: 'skrining_gizi', label: 'Skrining Gizi Lanjut', tables: ['skrining_gizi'] },
  { value: 'asuhan_gizi', label: 'Asuhan Gizi', tables: ['asuhan_gizi'] },
  { value: 'monitoring_gizi', label: 'Monitoring Gizi', tables: ['monitoring_asuhan_gizi'] },
  { value: 'konseling_farmasi', label: 'Konseling Farmasi', tables: ['konseling_farmasi'] },
  { value: 'informasi_obat', label: 'Pelayanan Informasi Obat', tables: ['pelayanan_informasi_obat'] },
  { value: 'transfer_ruang', label: 'Transfer Antar Ruang', tables: ['transfer_pasien_antar_ruang'] },
  { value: 'rujuk_internal', label: 'Rujukan Internal Poli', tables: ['rujukan_internal_poli'] },
  { value: 'tindakan_ralan_dokter', label: 'Tindakan Ralan Dokter' },
  { value: 'tindakan_ralan_petugas', label: 'Tindakan Ralan Petugas' },
  { value: 'tindakan_ralan_gabungan', label: 'Tindakan Ralan Dokter & Petugas' },
  { value: 'tindakan_ranap_dokter', label: 'Tindakan Ranap Dokter' },
  { value: 'tindakan_ranap_petugas', label: 'Tindakan Ranap Petugas' },
  { value: 'tindakan_ranap_gabungan', label: 'Tindakan Ranap Dokter & Petugas' },
  { value: 'kamar', label: 'Penggunaan Kamar' },
  { value: 'operasi', label: 'Operasi / VK' },
  { value: 'laporan_operasi', label: 'Laporan Operasi', tables: ['laporan_operasi'] },
  { value: 'radiologi', label: 'Pemeriksaan Radiologi' },
  { value: 'laboratorium', label: 'Pemeriksaan Laboratorium' },
  { value: 'lab_pa', label: 'Pemeriksaan Patologi Anatomi', tables: ['permintaan_labpa', 'saran_kesan_lab'] },
  { value: 'obat', label: 'Pemberian Obat / BHP / Alkes' },
  { value: 'obat_operasi', label: 'Penggunaan Obat Operasi' },
  { value: 'aturan_pakai', label: 'Aturan Pakai Obat', tables: ['aturan_pakai'] },
  { value: 'resep_pulang', label: 'Resep Pulang' },
  { value: 'biaya_tambahan', label: 'Tambahan Biaya' },
  { value: 'potongan_biaya', label: 'Potongan Biaya' },
  { value: 'resume', label: 'Resume Pasien' },
  { value: 'resume_ranap', label: 'Resume Pasien Ranap', tables: ['resume_pasien_ranap'] },
  { value: 'berkas_digital', label: 'Berkas Digital Perawatan' },
  { value: 'tanda_tangan', label: 'Tanda Tangan / Verifikasi' },
  { value: 'dokumen_lain', label: 'Dokumen Klinis Lainnya' },
]

const kelompok = reactive(Object.fromEntries(pilihanKelompok.map((pilihan) => [pilihan.value, true])))
const casemixDipilih = ref(false)
const semuaKelompokDipilih = computed(() => pilihanKelompok.every((pilihan) => kelompok[pilihan.value]))

const kunjungan = computed(() => data.value?.kunjungan || [])

function tanggalAPI(value) {
  if (!(value instanceof Date) || Number.isNaN(value.getTime())) return ''
  const lokal = new Date(value.getTime() - value.getTimezoneOffset() * 60000)
  return lokal.toISOString().slice(0, 10)
}

function tanggalIndonesia(value) {
  if (!value) return '-'
  const [tahun, bulan, tanggal] = value.split('-')
  return `${tanggal}/${bulan}/${tahun}`
}

function rupiah(value) {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(Number(value || 0))
}

function jumlahIsi(item) {
  const jumlahUtama = ['diagnosis', 'prosedur', 'catatan', 'catatan_dokter', 'radiologi', 'laboratorium', 'obat', 'tindakan', 'kamar', 'operasi', 'biaya_lain', 'resep_pulang', 'berkas_digital', 'tanda_tangan']
    .reduce((jumlah, key) => jumlah + (Array.isArray(item[key]) ? item[key].length : 0), 0)
  const jumlahDokumen = (item.dokumen_klinis || []).reduce((jumlah, dokumen) => jumlah + (dokumen.data?.length || 0), 0)
  return jumlahUtama + jumlahDokumen + (item.resume ? 1 : 0)
}

function aturSemuaKelompok(event) {
  const dipilih = Boolean(event.target.checked)
  casemixDipilih.value = false
  pilihanKelompok.forEach((pilihan) => {
    kelompok[pilihan.value] = dipilih
  })
}

function aturCasemix() {
  pilihanKelompok.forEach((pilihan) => {
    kelompok[pilihan.value] = false
  })
  if (!casemixDipilih.value) return
  const pilihanCasemix = [
    'triase', 'diagnosis', 'prosedur', 'hemodialisa', 'asuhan_gizi',
    'tindakan_ralan_gabungan', 'tindakan_ranap_gabungan', 'kamar', 'operasi',
    'radiologi', 'laboratorium', 'obat', 'resep_pulang', 'resume',
    'berkas_digital', 'observasi_igd', 'catatan_keperawatan_ranap',
  ]
  pilihanCasemix.forEach((nama) => {
    kelompok[nama] = true
  })
}

function catatanTerpilih(item) {
  return (item.catatan || []).filter((catatan) => (
    catatan.jenis === 'Rawat Jalan' ? kelompok.pemeriksaan_ralan : kelompok.pemeriksaan_ranap
  ))
}

function tindakanTerpilih(item) {
  const pemetaan = {
    'Rawat Jalan Dokter': 'tindakan_ralan_dokter',
    'Rawat Jalan Petugas': 'tindakan_ralan_petugas',
    'Rawat Jalan Dokter & Petugas': 'tindakan_ralan_gabungan',
    'Rawat Inap Dokter': 'tindakan_ranap_dokter',
    'Rawat Inap Petugas': 'tindakan_ranap_petugas',
    'Rawat Inap Dokter & Petugas': 'tindakan_ranap_gabungan',
  }
  return (item.tindakan || []).filter((tindakan) => kelompok[pemetaan[tindakan.jenis]])
}

function obatTerpilih(item) {
  return (item.obat || []).filter((obat) => (
    obat.jenis === 'Obat Operasi' ? kelompok.obat_operasi : kelompok.obat
  ))
}

function biayaTerpilih(item) {
  return (item.biaya_lain || []).filter((biaya) => (
    biaya.jenis === 'Potongan' ? kelompok.potongan_biaya : kelompok.biaya_tambahan
  ))
}

function dokumenKlinisTerpilih(item) {
  return (item.dokumen_klinis || []).filter((dokumen) => {
    const pilihan = pilihanKelompok.find((opsi) => opsi.tables?.includes(dokumen.tabel))
    return pilihan ? kelompok[pilihan.value] : kelompok.dokumen_lain
  })
}

const labelKhusus = {
  nip: 'NIP/NIK Petugas', kd_dokter: 'Kode Dokter', tgl_perawatan: 'Tanggal Perawatan', jam_rawat: 'Jam Rawat',
  tgl_periksa: 'Tanggal Pemeriksaan', kd_jenis_prw: 'Kode Pemeriksaan', no_rkm_medis: 'Nomor Rekam Medis',
  gcs: 'GCS', spo2: 'SpO2', adl: 'ADL', td: 'Tekanan Darah', rr: 'Respirasi',
}

function labelKolom(nama) {
  if (labelKhusus[nama]) return labelKhusus[nama]
  return String(nama).split('_').map((kata) => kata ? kata[0].toUpperCase() + kata.slice(1) : '').join(' ')
}

function berkasPDF(berkas) {
  return /\.pdf(?:$|[?#])/i.test(berkas?.url || berkas?.lokasi_file || '')
}

function toggleKunjungan(noRawat) {
  const berikutnya = new Set(terbuka.value)
  if (berikutnya.has(noRawat)) berikutnya.delete(noRawat)
  else berikutnya.add(noRawat)
  terbuka.value = berikutnya
}

async function muatData() {
  if (!props.patient?.no_rawat) return
  loading.value = true
  error.value = ''
  try {
    const params = { no_rawat: props.patient.no_rawat, mode: mode.value }
    if (mode.value === 'tanggal') {
      params.tanggal_mulai = tanggalAPI(tanggalMulai.value)
      params.tanggal_selesai = tanggalAPI(tanggalSelesai.value)
    }
    if (mode.value === 'nomor') params.nomor_rawat = nomorRawat.value.trim()
    const response = await riwayatPerawatanData(props.token, params)
    data.value = response || { kunjungan: [], jumlah: 0 }
    const daftarKunjungan = response?.kunjungan || []
    const kunjunganAktif = daftarKunjungan.find((item) => item.no_rawat === props.patient.no_rawat) || daftarKunjungan[0]
    terbuka.value = new Set(kunjunganAktif ? [kunjunganAktif.no_rawat] : [])
    if (daftarKunjungan.length === 0) notifikasi.peringatan('Riwayat perawatan pasien tidak ditemukan.')
  } catch (err) {
    error.value = err.message
    notifikasi.gagal(err.message)
  } finally {
    loading.value = false
  }
}

watch(() => props.patient.no_rawat, () => {
  mode.value = 'nomor'
  nomorRawat.value = props.patient.no_rawat || ''
  muatData()
}, { immediate: true })
</script>

<template>
  <section class="care-history-page">
    <article class="care-history-filter">
      <header>
        <div><span>REKAM MEDIS PASIEN</span><h3>Riwayat Perawatan</h3><p>Data dibaca langsung dari SIMRS Khanza berdasarkan nomor rekam medis.</p></div>
        <button type="button" :disabled="loading" @click="muatData"><LoaderCircle v-if="loading" class="spin" :size="15" /><Filter v-else :size="15" /> Tampilkan</button>
      </header>

      <div class="care-history-modes">
        <label v-for="pilihan in pilihanMode" :key="pilihan.value" :class="{ active: mode === pilihan.value }">
          <input v-model="mode" type="radio" :value="pilihan.value" />
          <span>{{ pilihan.label }}</span>
        </label>
      </div>

      <div v-if="mode === 'tanggal'" class="care-history-range">
        <label><span>TANGGAL MULAI</span><DatePicker v-model="tanggalMulai" date-format="dd/mm/yy" show-icon /></label>
        <label><span>SAMPAI DENGAN</span><DatePicker v-model="tanggalSelesai" date-format="dd/mm/yy" show-icon /></label>
      </div>
      <label v-else-if="mode === 'nomor'" class="care-history-number">
        <span>NOMOR RAWAT</span>
        <div><Search :size="15" /><input v-model="nomorRawat" type="search" placeholder="Contoh: 2026/08/13/000001" @keyup.enter="muatData" /></div>
      </label>

    </article>

    <article class="care-history-content">
      <header class="care-history-summary">
        <div><span>RIWAYAT KUNJUNGAN PASIEN</span><h3>{{ data.nama_pasien || patient.nama_pasien }}</h3><p>RM {{ data.no_rekam_medis || patient.no_rekam_medis }} · {{ data.jumlah || 0 }} riwayat kunjungan ditampilkan</p></div>
      </header>

      <div class="care-history-report-layout" :class="{ 'options-hidden': !pilihanTerbuka }">
        <aside v-if="pilihanTerbuka" class="care-history-report-options">
          <div class="care-history-options-head">
            <strong>DATA DITAMPILKAN</strong>
            <button type="button" title="Sembunyikan pilihan data" aria-label="Sembunyikan pilihan data" @click="pilihanTerbuka = false"><ChevronLeft :size="17" /></button>
          </div>
          <label class="care-history-preset"><input v-model="casemixDipilih" type="checkbox" @change="aturCasemix" /> Casemix</label>
          <label class="care-history-select-all"><input type="checkbox" :checked="semuaKelompokDipilih" @change="aturSemuaKelompok" /> Semua</label>
          <label v-for="pilihan in pilihanKelompok" :key="pilihan.value"><input v-model="kelompok[pilihan.value]" type="checkbox" @change="casemixDipilih = false" /> {{ pilihan.label }}</label>
        </aside>
        <aside v-else class="care-history-options-rail">
          <button type="button" title="Tampilkan pilihan data" aria-label="Tampilkan pilihan data" @click="pilihanTerbuka = true"><ChevronRight :size="18" /></button>
        </aside>

        <main class="care-history-report-content">
          <div v-if="loading" class="cppt-state"><LoaderCircle class="spin" :size="26" /><strong>Menarik seluruh riwayat pasien dari SIMRS...</strong></div>
          <div v-else-if="error" class="cppt-state error"><strong>{{ error }}</strong><button type="button" @click="muatData">Coba Lagi</button></div>
          <div v-else-if="kunjungan.length === 0" class="cppt-state"><strong>Riwayat perawatan tidak ditemukan.</strong></div>

          <div v-else class="care-history-visits">
            <section v-for="(item, visitIndex) in kunjungan" :key="item.no_rawat" class="care-history-visit" :class="{ open: terbuka.has(item.no_rawat) }">
              <button type="button" class="care-history-report-caption" @click="toggleKunjungan(item.no_rawat)">
                <span>Riwayat Kunjungan {{ visitIndex + 1 }}</span>
                <small>{{ tanggalIndonesia(item.tanggal) }} {{ item.jam }} · {{ jumlahIsi(item) }} data</small>
                <ChevronUp v-if="terbuka.has(item.no_rawat)" :size="17" /><ChevronDown v-else :size="17" />
              </button>

              <div v-if="terbuka.has(item.no_rawat)" class="care-history-visit-body">
                <div class="care-history-identity">
                  <div><i>{{ visitIndex + 1 }}</i><b>No.Rawat</b><em>:</em><span>{{ item.no_rawat }}</span></div>
                  <div><i></i><b>No. Rkm Medis</b><em>:</em><span>{{ data.no_rekam_medis || patient.no_rekam_medis }}</span></div>
                  <div v-if="item.no_sep"><i></i><b>No.SEP</b><em>:</em><span>{{ item.no_sep }}</span></div>
                  <div v-if="item.kelas_rawat"><i></i><b>Kelas Rawat</b><em>:</em><span>{{ item.kelas_rawat }}</span></div>
                  <div><i></i><b>Nama Pasien</b><em>:</em><span>{{ data.nama_pasien || patient.nama_pasien }}</span></div>
                  <div><i></i><b>No.Registrasi</b><em>:</em><span>{{ item.no_registrasi || '-' }}</span></div>
                  <div><i></i><b>Tanggal Registrasi</b><em>:</em><span>{{ item.tanggal }} {{ item.jam }}</span></div>
                  <div><i></i><b>Unit/Poliklinik</b><em>:</em><span>{{ item.poliklinik || item.ruangan || '-' }}</span></div>
                  <div><i></i><b>Dokter Poli</b><em>:</em><span>{{ item.dokter || '-' }}</span></div>
                  <div v-if="item.status_rawat === 'Ranap'"><i></i><b>DPJP Ranap</b><em>:</em><span>{{ item.dpjp?.map((nama, index) => `${index + 1}. ${nama}`).join('  ') || '-' }}</span></div>
                  <div><i></i><b>Cara Bayar</b><em>:</em><span>{{ item.penjamin || '-' }}</span></div>
                  <div><i></i><b>Penanggung Jawab</b><em>:</em><span>{{ item.penanggung_jawab || '-' }}</span></div>
                  <div><i></i><b>Alamat P.J.</b><em>:</em><span>{{ item.alamat_penanggung_jawab || '-' }}</span></div>
                  <div><i></i><b>Hubungan P.J.</b><em>:</em><span>{{ item.hubungan_penanggung_jawab || '-' }}</span></div>
                  <div><i></i><b>Status</b><em>:</em><span>{{ item.status_rawat || '-' }}</span></div>
                  <div><i></i><b>Status Periksa</b><em>:</em><span>{{ item.status_periksa || '-' }}</span></div>
                  <div><i></i><b>Status Bayar</b><em>:</em><span>{{ item.status_bayar || '-' }}</span></div>
                </div>

            <section v-if="kelompok.diagnosis && item.diagnosis.length" class="care-history-section">
              <h4>Diagnosis</h4>
              <div class="care-history-code-list"><article v-for="diagnosis in item.diagnosis" :key="`${diagnosis.kode}-${diagnosis.prioritas}`"><b>{{ diagnosis.kode }}</b><span>{{ diagnosis.nama }}</span><small>{{ diagnosis.prioritas === 1 ? 'Primer' : 'Sekunder' }} · {{ diagnosis.status }}</small></article></div>
            </section>

            <section v-if="kelompok.prosedur && item.prosedur.length" class="care-history-section">
              <h4>Prosedur / Tindakan ICD-9-CM</h4>
              <div class="care-history-code-list"><article v-for="prosedur in item.prosedur" :key="`${prosedur.kode}-${prosedur.prioritas}`"><b>{{ prosedur.kode }}</b><span>{{ prosedur.nama }}</span><small>{{ prosedur.prioritas === 1 ? 'Primer' : 'Sekunder' }}</small></article></div>
            </section>

            <section v-if="catatanTerpilih(item).length" class="care-history-section">
              <h4>Catatan Perkembangan Pasien / SOAP</h4>
              <div class="care-history-soap"><article v-for="catatan in catatanTerpilih(item)" :key="`${catatan.jenis}-${catatan.tanggal}-${catatan.jam}`"><header><b>{{ tanggalIndonesia(catatan.tanggal) }} {{ catatan.jam }}</b><span>{{ catatan.jenis }} · {{ catatan.petugas }} · {{ catatan.jabatan }}</span></header><div class="care-history-vitals"><span>Kesadaran <b>{{ catatan.kesadaran || '-' }}</b></span><span>Tensi <b>{{ catatan.tensi || '-' }}</b></span><span>Suhu <b>{{ catatan.suhu || '-' }}</b></span><span>Nadi <b>{{ catatan.nadi || '-' }}</b></span><span>RR <b>{{ catatan.respirasi || '-' }}</b></span><span>SpO2 <b>{{ catatan.spo2 || '-' }}</b></span></div><dl><dt>S</dt><dd>{{ catatan.subjek || '-' }}</dd><dt>O</dt><dd>{{ catatan.objek || '-' }}</dd><dt>A</dt><dd>{{ catatan.asesmen || '-' }}</dd><dt>P</dt><dd>{{ catatan.plan || '-' }}</dd><dt>Instruksi</dt><dd>{{ catatan.instruksi || '-' }}</dd><dt>Evaluasi</dt><dd>{{ catatan.evaluasi || '-' }}</dd></dl></article></div>
            </section>

			<section v-if="kelompok.catatan_dokter && item.catatan_dokter?.length" class="care-history-section">
			  <h4>Catatan Dokter</h4>
			  <div class="care-history-doctor-notes"><article v-for="(catatan, index) in item.catatan_dokter" :key="`${catatan.tanggal}-${catatan.jam}-${index}`"><header><b>{{ tanggalIndonesia(catatan.tanggal) }} {{ catatan.jam }}</b><span>{{ catatan.dokter || '-' }}</span></header><p>{{ catatan.catatan || '-' }}</p></article></div>
			</section>

            <section v-if="kelompok.radiologi && item.radiologi.length" class="care-history-section">
              <h4>Pemeriksaan Radiologi</h4>
              <div class="care-history-table"><div class="head"><span>Tanggal</span><span>Kode / Pemeriksaan</span><span>Petugas / Dokter</span><span>Biaya</span></div><div v-for="row in item.radiologi" :key="`${row.kode}-${row.tanggal}-${row.jam}`"><span>{{ tanggalIndonesia(row.tanggal) }} {{ row.jam }}</span><span><b>{{ row.kode }}</b><small>{{ row.nama }}</small><small v-if="row.hasil" class="care-history-result">Hasil: {{ row.hasil }}</small></span><span><b>{{ row.petugas || '-' }}</b><small>{{ row.dokter || '-' }}</small></span><span>{{ rupiah(row.biaya) }}</span></div></div>
            </section>

            <section v-if="kelompok.laboratorium && item.laboratorium.length" class="care-history-section">
              <h4>Pemeriksaan Laboratorium</h4>
              <div class="care-history-table"><div class="head"><span>Tanggal</span><span>Kode / Pemeriksaan</span><span>Petugas / Dokter</span><span>Biaya</span></div><div v-for="row in item.laboratorium" :key="`${row.kode}-${row.tanggal}-${row.jam}`"><span>{{ tanggalIndonesia(row.tanggal) }} {{ row.jam }}</span><span><b>{{ row.kode }}</b><small>{{ row.nama }}</small><span v-if="row.detail?.length" class="care-history-lab-detail"><small v-for="(detail, detailIndex) in row.detail" :key="detailIndex"><b>{{ detail.pemeriksaan }}</b>: {{ detail.nilai }} {{ detail.satuan }} <i>(Rujukan {{ detail.nilai_rujukan || '-' }})</i><em v-if="detail.keterangan">{{ detail.keterangan }}</em></small></span></span><span><b>{{ row.petugas || '-' }}</b><small>{{ row.dokter || '-' }}</small></span><span>{{ rupiah(row.biaya) }}</span></div></div>
            </section>

			<section v-if="dokumenKlinisTerpilih(item).length" class="care-history-section">
			  <h4>Asesmen dan Dokumen Klinis</h4>
			  <div class="care-history-documents">
				<details v-for="dokumen in dokumenKlinisTerpilih(item)" :key="dokumen.tabel" open>
				  <summary><span>{{ dokumen.jenis }}</span><small>{{ dokumen.data.length }} data</small></summary>
				  <article v-for="(baris, index) in dokumen.data" :key="index" class="care-history-document-row">
					<div v-for="(nilai, kolom) in baris" :key="kolom" v-show="nilai !== ''"><small>{{ labelKolom(kolom) }}</small><p>{{ nilai || '-' }}</p></div>
				  </article>
				</details>
			  </div>
			</section>

            <section v-if="obatTerpilih(item).length" class="care-history-section">
              <h4>Pemberian Obat / BHP / Alkes</h4>
              <div class="care-history-table medicine"><div class="head"><span>Tanggal</span><span>Kode / Nama Obat</span><span>Jumlah</span><span>Total</span></div><div v-for="row in obatTerpilih(item)" :key="`${row.jenis}-${row.kode}-${row.tanggal}-${row.jam}`"><span>{{ tanggalIndonesia(row.tanggal) }} {{ row.jam }}</span><span><b>{{ row.kode }}</b><small>{{ row.jenis || 'Pemberian Obat' }} · {{ row.nama }}</small></span><span>{{ row.jumlah }} {{ row.satuan }}</span><span>{{ rupiah(row.total) }}</span></div></div>
            </section>

            <section v-if="tindakanTerpilih(item).length" class="care-history-section">
              <h4>Tindakan Rawat Jalan / Rawat Inap</h4>
              <div class="care-history-table"><div class="head"><span>Tanggal</span><span>Jenis / Tindakan</span><span>Pelaksana</span><span>Biaya</span></div><div v-for="row in tindakanTerpilih(item)" :key="`${row.jenis}-${row.kode}-${row.tanggal}-${row.jam}`"><span>{{ tanggalIndonesia(row.tanggal) }} {{ row.jam }}</span><span><b>{{ row.kode }} · {{ row.jenis }}</b><small>{{ row.nama }}</small></span><span>{{ row.pelaksana || '-' }}</span><span>{{ rupiah(row.biaya) }}</span></div></div>
            </section>

            <section v-if="kelompok.kamar && item.kamar.length" class="care-history-section">
              <h4>Penggunaan Kamar</h4>
              <div class="care-history-table room"><div class="head"><span>Masuk / Keluar</span><span>Kamar / Bangsal</span><span>Lama / Status Pulang</span><span>Biaya</span></div><div v-for="row in item.kamar" :key="`${row.kode_kamar}-${row.tanggal_masuk}-${row.jam_masuk}`"><span><b>{{ tanggalIndonesia(row.tanggal_masuk) }} {{ row.jam_masuk }}</b><small>{{ row.tanggal_keluar ? `${tanggalIndonesia(row.tanggal_keluar)} ${row.jam_keluar}` : 'Belum pulang' }}</small></span><span><b>{{ row.kode_kamar }}</b><small>{{ row.bangsal }}</small></span><span><b>{{ row.lama }} hari</b><small>{{ row.status_pulang || '-' }}</small></span><span>{{ rupiah(row.total) }}</span></div></div>
            </section>

            <section v-if="kelompok.operasi && item.operasi.length" class="care-history-section">
              <h4>Operasi / VK</h4>
              <div class="care-history-table"><div class="head"><span>Tanggal</span><span>Kode / Tindakan</span><span>Operator / Anestesi</span><span>Biaya</span></div><div v-for="row in item.operasi" :key="`${row.kode}-${row.tanggal}`"><span>{{ row.tanggal }}</span><span><b>{{ row.kode }}</b><small>{{ row.nama }}</small></span><span><b>{{ row.operator || '-' }}</b><small>{{ row.anestesi || '-' }}</small></span><span>{{ rupiah(row.total) }}</span></div></div>
            </section>

            <section v-if="biayaTerpilih(item).length" class="care-history-section">
              <h4>Biaya Tambahan / Potongan</h4>
              <div class="care-history-costs"><article v-for="(row, index) in biayaTerpilih(item)" :key="`${row.jenis}-${row.nama}-${index}`"><span><b>{{ row.jenis }}</b><small>{{ row.nama }}</small></span><strong :class="{ deduction: row.total < 0 }">{{ rupiah(row.total) }}</strong></article></div>
            </section>

            <section v-if="kelompok.resep_pulang && item.resep_pulang.length" class="care-history-section">
              <h4>Resep Pulang</h4>
              <div class="care-history-table prescription"><div class="head"><span>Kode</span><span>Nama Obat</span><span>Dosis / Jumlah</span><span>Total</span></div><div v-for="row in item.resep_pulang" :key="row.kode"><span>{{ row.kode }}</span><span>{{ row.nama }}</span><span><b>{{ row.dosis || '-' }}</b><small>{{ row.jumlah }} {{ row.satuan }}</small></span><span>{{ rupiah(row.total) }}</span></div></div>
            </section>

            <section v-if="kelompok.resume && item.resume" class="care-history-section care-history-resume">
              <h4>Resume Pasien</h4>
              <div class="care-history-resume-head"><span>Dokter <b>{{ item.resume.dokter || '-' }}</b></span><span>Kondisi Pulang <b>{{ item.resume.kondisi_pulang || '-' }}</b></span></div>
              <dl><dt>Keluhan Utama</dt><dd>{{ item.resume.keluhan_utama || '-' }}</dd><dt>Jalannya Penyakit</dt><dd>{{ item.resume.jalannya_penyakit || '-' }}</dd><dt>Pemeriksaan Penunjang</dt><dd>{{ item.resume.pemeriksaan_penunjang || '-' }}</dd><dt>Hasil Laboratorium</dt><dd>{{ item.resume.hasil_laboratorium || '-' }}</dd><dt>Diagnosis Utama</dt><dd>{{ item.resume.diagnosis_utama || '-' }}</dd><dt>Diagnosis Sekunder</dt><dd>{{ item.resume.diagnosis_sekunder || '-' }}</dd><dt>Prosedur Utama</dt><dd>{{ item.resume.prosedur_utama || '-' }}</dd><dt>Prosedur Sekunder</dt><dd>{{ item.resume.prosedur_sekunder || '-' }}</dd><dt>Obat Pulang</dt><dd>{{ item.resume.obat_pulang || '-' }}</dd></dl>
            </section>

			<section v-if="kelompok.berkas_digital && item.berkas_digital?.length" class="care-history-section care-history-files">
			  <h4>Foto / Berkas Digital Perawatan</h4>
			  <div class="care-history-file-list">
				<article v-for="(berkas, index) in item.berkas_digital" :key="`${berkas.lokasi_file}-${index}`">
				  <header><span><b>{{ index + 1 }}. {{ berkas.nama || berkas.jenis }}</b><small>{{ berkas.jenis }} · {{ berkas.lokasi_file }}</small></span><a v-if="berkas.url" :href="berkas.url" target="_blank" rel="noopener">Buka File Asli</a></header>
				  <iframe v-if="berkas.url && berkasPDF(berkas)" :src="berkas.url" :title="berkas.nama" loading="lazy" />
				  <a v-else-if="berkas.url" class="care-history-image-link" :href="berkas.url" target="_blank" rel="noopener"><img :src="berkas.url" :alt="berkas.nama || berkas.jenis" loading="lazy" /></a>
				  <p v-else>Alamat berkas belum tersedia.</p>
				</article>
			  </div>
			</section>

            <section v-if="kelompok.tanda_tangan && item.tanda_tangan?.length" class="care-history-section care-history-signatures">
              <h4>Tanda Tangan / Verifikasi</h4>
              <div class="care-history-signature-list">
                <TandaTanganVerifikasi v-for="tandaTangan in item.tanda_tangan" :key="`${tandaTangan.kode_dokter}-${tandaTangan.urutan}`" :data="tandaTangan" />
              </div>
            </section>
              </div>
            </section>
          </div>
        </main>
      </div>
    </article>
  </section>
</template>
