// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { computed, reactive, ref, watch } from "vue"
import { riwayatPerawatanData } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useRiwayatPerawatan(props) {
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
  return {
    loading,
    error,
    data,
    mode,
    tanggalMulai,
    tanggalSelesai,
    nomorRawat,
    terbuka,
    pilihanTerbuka,
    pilihanMode,
    pilihanKelompok,
    kelompok,
    casemixDipilih,
    semuaKelompokDipilih,
    kunjungan,
    tanggalIndonesia,
    rupiah,
    jumlahIsi,
    aturSemuaKelompok,
    aturCasemix,
    catatanTerpilih,
    tindakanTerpilih,
    obatTerpilih,
    biayaTerpilih,
    dokumenKlinisTerpilih,
    labelKolom,
    berkasPDF,
    toggleKunjungan,
    muatData,
  }
}
