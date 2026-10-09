import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { request } from '../../../lib/shared/http'
import type { PropsRingkasanPasien, RingkasanPasien } from '../../../types/ringkasanPasien'

export function useRingkasan(props: PropsRingkasanPasien) {
  const loading = ref(false)
  const error = ref('')
  const data = ref<RingkasanPasien | null>(null)
  const diperbarui = ref('')
  let urutan = 0
  let controller: AbortController | undefined
  const bagian = (kode: string) => data.value?.bagian.find(b => b.kode === kode)
  const sep = computed(() => bagian('sep'))
  const diagnosis = computed(() => bagian('diagnosis'))
  const alergi = computed(() => bagian('alergi'))
  const vital = computed(() => bagian('vital'))
  const hasilAktif = ref('')
  const kartu = computed(() => [
    { kode: 'cppt', judul: 'CPPT / SOAP', menu: 'cppt_soap', unit: 'catatan', tombol: 'Buka CPPT' },
    { kode: 'diagnosis', judul: 'Diagnosis', menu: 'diagnosa', unit: 'diagnosis terkode', tombol: 'Buka Diagnosis' },
    { kode: 'resume', judul: 'Resume Medis', menu: 'resume_pasien', unit: 'resume', tombol: 'Buka Resume' },
    { kode: 'berkas', judul: 'Berkas Digital', menu: 'berkas_digital', unit: 'tautan berkas', tombol: 'Lihat Berkas' },
    { kode: 'lab', judul: 'Laboratorium PK', menu: '', unit: 'parameter hasil', tombol: 'Lihat Permintaan & Hasil' },
    { kode: 'radiologi', judul: 'Radiologi', menu: '', unit: 'laporan teks', tombol: 'Lihat Permintaan & Hasil' },
    { kode: 'resep', judul: 'Resep Dokter', menu: '', unit: 'nomor resep', tombol: 'Lihat Resep & Racikan' },
    { kode: 'obat', judul: 'Transaksi Obat', menu: '', unit: 'baris transaksi', tombol: 'Lihat Transaksi Obat' },
    { kode: 'edukasi', judul: 'Edukasi Pasien', menu: 'edukasi_pasien', unit: 'catatan', tombol: 'Buka Edukasi' },
    { kode: 'pemulangan', judul: 'Rencana Pemulangan', menu: 'perencanaan_pemulangan', unit: 'rencana', tombol: 'Buka Rencana Pemulangan' },
  ].filter(k => props.moduleName === 'Rawat Inap' || !['edukasi', 'pemulangan'].includes(k.kode)).map(k => {
    const b = bagian(k.kode)
    const jumlah = !b || b.error ? null : ['cppt', 'resume', 'berkas', 'edukasi', 'pemulangan'].includes(k.kode)
      ? Number(b.baris[0]?.jumlah ?? 0) : b.baris.length
    const permintaan = ['lab', 'radiologi'].includes(k.kode) ? bagian('permintaan_' + k.kode) : undefined
    const progres = permintaan && !permintaan.error ? {
      total: permintaan.baris.length,
      menunggu: permintaan.baris.filter(r => r.tahap === 'menunggu').length,
      diproses: permintaan.baris.filter(r => r.tahap === 'diproses').length,
      hasilDicatat: permintaan.baris.filter(r => r.tahap === 'hasil_dicatat').length,
    } : null
    return { ...k, jumlah, progres, sumber: b?.sumber, error: b?.error,
      status: jumlah === null ? 'Gagal diperiksa' : jumlah > 0 ? 'Ada catatan' : 'Belum tercatat' }
  }))
  const kelompokKartu = computed(() => [
    {
      kode: 'catatan',
      judul: 'Catatan & berkas',
      deskripsi: 'Dokumentasi pelayanan pada kunjungan ini.',
      daftar: ['cppt', 'diagnosis', 'resume', 'berkas', 'edukasi', 'pemulangan'],
    },
    {
      kode: 'pemeriksaan',
      judul: 'Pemeriksaan penunjang',
      deskripsi: 'Pantau permintaan dan lihat hasil yang tersedia.',
      daftar: ['lab', 'radiologi'],
    },
    {
      kode: 'farmasi',
      judul: 'Resep & obat',
      deskripsi: 'Resep dan transaksi dicatat terpisah; bukan bukti pemberian obat.',
      daftar: ['resep', 'obat'],
    },
  ].map(kelompok => ({
    ...kelompok,
    kartu: kartu.value.filter(k => kelompok.daftar.includes(k.kode)),
  })))
  async function bukaDetail(kode: string) {
    if (hasilAktif.value === kode) {
      hasilAktif.value = ''
      return
    }
    hasilAktif.value = kode
    await nextTick()
    if (typeof document !== 'undefined') {
      document.getElementById('ringkasan-panel-hasil')?.focus()
    }
  }
  async function tutupDetail() {
    const kode = hasilAktif.value
    hasilAktif.value = ''
    await nextTick()
    if (typeof document !== 'undefined') {
      document.getElementById('ringkasan-buka-' + kode)?.focus()
    }
  }
  const detailHasil = computed(() => bagian(hasilAktif.value))
  const penunjangAktif = computed(() => ['lab', 'radiologi'].includes(hasilAktif.value))
  const judulDetail = computed(() => kartu.value.find(k => k.kode === hasilAktif.value)?.judul || '')
  const kolomObat: Record<string, { field: string; header: string }[]> = {
    resep: [
      { field: 'no_resep', header: 'No. Resep' },
      { field: 'waktu', header: 'Waktu Peresepan (WITA)' },
      { field: 'dokter', header: 'Dokter Peresep' },
      { field: 'layanan', header: 'Layanan' },
    ],
    resep_item: [
      { field: 'no_resep', header: 'No. Resep' },
      { field: 'kode_obat', header: 'Kode Obat' },
      { field: 'obat', header: 'Obat Nonracikan' },
      { field: 'jumlah', header: 'Jumlah Resep' },
      { field: 'satuan', header: 'Satuan' },
      { field: 'aturan_pakai', header: 'Aturan Pakai' },
    ],
    resep_racikan: [
      { field: 'no_resep', header: 'No. Resep' },
      { field: 'no_racik', header: 'No. Racikan' },
      { field: 'nama_racik', header: 'Nama Racikan' },
      { field: 'jumlah', header: 'Jumlah Sediaan' },
      { field: 'aturan_pakai', header: 'Aturan Pakai' },
      { field: 'keterangan', header: 'Keterangan' },
    ],
    obat: [
      { field: 'waktu', header: 'Waktu Transaksi (WITA)' },
      { field: 'kode_obat', header: 'Kode Obat' },
      { field: 'obat', header: 'Nama Obat' },
      { field: 'jumlah', header: 'Jumlah Transaksi' },
      { field: 'satuan', header: 'Satuan' },
      { field: 'layanan', header: 'Layanan' },
      { field: 'kd_bangsal', header: 'Kode Depo / Bangsal' },
      { field: 'no_batch', header: 'Batch' },
      { field: 'no_faktur', header: 'Faktur' },
    ],
  }
  const detailObat = computed(() => {
    const daftar = hasilAktif.value === 'resep'
      ? [{ kode: 'resep', judul: 'Daftar Resep' },
         { kode: 'resep_item', judul: 'Obat Nonracikan' },
         { kode: 'resep_racikan', judul: 'Racikan' }]
      : [{ kode: 'obat', judul: 'Riwayat Transaksi Obat' }]
    return daftar.map(item => ({ ...item, bagian: bagian(item.kode), kolom: kolomObat[item.kode] || [] }))
  })
  const detailPermintaan = computed(() => bagian('permintaan_' + hasilAktif.value))
  const kolomPermintaan = computed(() => [
    { field: 'noorder', header: 'No. Permintaan' },
    { field: 'waktu_permintaan', header: 'Diminta (WITA)' },
    { field: 'dokter', header: 'Dokter Perujuk' },
    { field: 'layanan', header: 'Layanan' },
    { field: 'status_progres', header: 'Progres Permintaan' },
    { field: 'waktu_proses', header: hasilAktif.value === 'lab' ? 'Sampel Diterima (WITA)' : 'Diterima / Diproses (WITA)' },
    { field: 'waktu_hasil', header: 'Tanggal Hasil pada Order (WITA)' },
  ])
  const kolomHasil = computed(() => hasilAktif.value === 'lab' ? [
    { field: 'waktu', header: 'Tanggal / Jam (WITA)' },
    { field: 'kode_pemeriksaan', header: 'Kode Pemeriksaan' },
    { field: 'pemeriksaan', header: 'Parameter' },
    { field: 'nilai', header: 'Hasil' },
    { field: 'satuan', header: 'Satuan' },
    { field: 'nilai_rujukan', header: 'Nilai Rujukan' },
    { field: 'keterangan', header: 'Keterangan' },
  ] : [
    { field: 'waktu', header: 'Tanggal / Jam Pemeriksaan (WITA)' },
    { field: 'hasil', header: 'Laporan Hasil Radiologi' },
  ])
  async function muat() {
    const id = ++urutan
    controller?.abort()
    controller = new AbortController()
    loading.value = true
    data.value = null
    error.value = ''
    diperbarui.value = ''
    hasilAktif.value = ''
    try {
      const params = new URLSearchParams({ no_rawat: props.patient.no_rawat, modul: props.moduleName })
      const hasil = await request<RingkasanPasien>('/api/ringkasan-pasien?' + params, {
        headers: { Authorization: 'Bearer ' + props.token }, signal: controller.signal,
      })
      if (id !== urutan) return
      data.value = hasil
      diperbarui.value = new Intl.DateTimeFormat('id-ID', {
        timeZone: 'Asia/Makassar', dateStyle: 'medium', timeStyle: 'medium',
      }).format(new Date())
    } catch (e) {
      if (id === urutan) error.value = e instanceof Error ? e.message : 'Ringkasan gagal dimuat.'
    } finally {
      if (id === urutan) loading.value = false
    }
  }
  watch(() => [props.token, props.patient.no_rawat, props.moduleName], () => { void muat() }, { immediate: true })
  onBeforeUnmount(() => { urutan++; controller?.abort() })
  const bolehBuka = (kode: string) => props.kodeSidebar.includes(kode)
  const tampilNilai = (nilai?: string) => !nilai?.trim() || nilai.trim() === '-' ? 'Belum tercatat' : nilai
  return { loading, error, data, diperbarui, sep, diagnosis, alergi, vital,
    kartu, kelompokKartu, bukaDetail, tutupDetail, hasilAktif, detailHasil, kolomHasil, detailPermintaan, kolomPermintaan,
    penunjangAktif, judulDetail, detailObat,
    muat, bolehBuka, tampilNilai }
}
