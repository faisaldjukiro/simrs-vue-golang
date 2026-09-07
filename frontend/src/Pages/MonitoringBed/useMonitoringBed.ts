import { computed, ref, onMounted } from 'vue'
import { monitoringBedData } from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'
import type { MonitoringBed } from '../../types/monitoringBed'

function bacaTanggal(nilai: string) {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(nilai) || nilai === '0000-00-00') return null
  const [tahun, bulan, hari] = nilai.split('-').map(Number)
  const tanggal = new Date(tahun, bulan - 1, hari)
  if (
    tanggal.getFullYear() !== tahun ||
    tanggal.getMonth() !== bulan - 1 ||
    tanggal.getDate() !== hari
  ) return null
  return tanggal
}

const formatTanggal = new Intl.DateTimeFormat('id-ID', {
  day: '2-digit',
  month: 'short',
  year: 'numeric',
})

export function useMonitoringBed(token: string) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const error = ref('')
  const grafikVisible = ref(false)
  const firstRow = ref(0)
  const waktuDimuat = ref('')
  const filterDimuat = ref({ status: '', q: '' })
  const data = ref<MonitoringBed[]>([])
  const filter = ref({
    status: '',
    q: ''
  })

  const daftarBed = computed(() => data.value.map((bed) => {
    const masuk = bacaTanggal(bed.tanggal_masuk)
    const terakhir = bacaTanggal(bed.terakhir_dipakai)
    const sekarang = new Date()
    const hariIni = Date.UTC(sekarang.getFullYear(), sekarang.getMonth(), sekarang.getDate())
    const hariTerakhir = terakhir
      ? Date.UTC(terakhir.getFullYear(), terakhir.getMonth(), terakhir.getDate())
      : null
    const hariKosong = hariTerakhir !== null && hariTerakhir <= hariIni
      ? Math.floor((hariIni - hariTerakhir) / 86400000)
      : null
    const lama = bed.lama_digunakan_hari

    return {
      ...bed,
      statusLabel: labelStatus(bed.status_kamar),
      hariKosong,
      tanggalMasukLabel: masuk ? formatTanggal.format(masuk) : 'Belum tersedia',
      terakhirDipakaiLabel: terakhir ? formatTanggal.format(terakhir) : '',
      lamaPenggunaanLabel: masuk && Number.isFinite(lama) && lama >= 0
        ? (lama === 0 ? 'Masuk hari ini' : `${lama} hari digunakan`)
        : '',
      lamaKosongLabel: hariKosong !== null
        ? (hariKosong === 0 ? 'Kosong sejak hari ini' : `${hariKosong} hari kosong`)
        : '',
    }
  }))

  function labelStatus(status: string) {
    return ({ ISI: 'Terisi', KOSONG: 'Kosong', DIBERSIHKAN: 'Dibersihkan' })[status]
      || status || 'Belum diketahui'
  }

  const ringkasan = computed(() => {
    const jumlah = { total: data.value.length, isi: 0, kosong: 0, dibersihkan: 0, lainnya: 0 }
    for (const bed of data.value) {
      if (bed.status_kamar === 'ISI') jumlah.isi++
      else if (bed.status_kamar === 'KOSONG') jumlah.kosong++
      else if (bed.status_kamar === 'DIBERSIHKAN') jumlah.dibersihkan++
      else jumlah.lainnya++
    }
    return jumlah
  })

  function kelompokkan(field: 'status_kamar' | 'kelas' | 'bangsal') {
    const kelompok = new Map<string, number>()
    for (const bed of data.value) {
      const label = field === 'status_kamar'
        ? labelStatus(bed.status_kamar)
        : bed[field]?.trim() || 'Belum diketahui'
      kelompok.set(label, (kelompok.get(label) || 0) + 1)
    }
    const terbesar = Math.max(1, ...kelompok.values())
    return Array.from(kelompok, ([label, nilai]) => ({
      label,
      nilai,
      persen: nilai / terbesar * 100,
    })).sort((a, b) => b.nilai - a.nilai || a.label.localeCompare(b.label, 'id'))
  }

  const grafik = computed(() => [
    {
      judul: 'Komposisi Status Bed',
      keterangan: 'Jumlah bed pada setiap status dalam hasil filter.',
      data: kelompokkan('status_kamar'),
      lebar: false,
    },
    {
      judul: 'Bed berdasarkan Kelas',
      keterangan: 'Distribusi seluruh bed berdasarkan kelas perawatan.',
      data: kelompokkan('kelas'),
      lebar: false,
    },
    {
      judul: 'Bed per Bangsal / Ruangan',
      keterangan: 'Seluruh bangsal pada hasil filter, diurutkan berdasarkan jumlah bed.',
      data: kelompokkan('bangsal'),
      lebar: true,
    },
  ])

  function gantiHalaman(event: { first: number }) {
    firstRow.value = event.first
  }

  function excel() {
    if (loading.value) return
    if (!daftarBed.value.length) {
      notifikasi.peringatan('Tidak ada data bed untuk diekspor.')
      return
    }

    const escape = (value: unknown) => String(value ?? '')
      .replaceAll('&', '&amp;')
      .replaceAll('<', '&lt;')
      .replaceAll('>', '&gt;')
      .replaceAll('"', '&quot;')
      .replaceAll("'", '&#39;')
    const cell = (value: string | number) => {
      if (typeof value === 'number') return `<td>${value}</td>`
      const aman = /^[\s]*[=+@-]/.test(value) ? "'" + value : value
      return `<td style='mso-number-format:"\\@"'>${escape(aman)}</td>`
    }
    const headers = [
      'No.', 'Kamar / Bed', 'Kelas', 'Bangsal / Ruangan', 'Status',
      'Nama Pasien', 'No. Rawat', 'Tanggal Masuk', 'Lama Digunakan (Hari)',
      'Terakhir Digunakan', 'Lama Kosong (Hari)',
    ]
    const rows = daftarBed.value.map((bed, index) => {
      const terisi = bed.status_kamar === 'ISI'
      const kosong = bed.status_kamar === 'KOSONG'
      const values = [
        index + 1, bed.kamar, bed.kelas, bed.bangsal, bed.statusLabel,
        terisi ? bed.nama_pasien : '',
        terisi ? bed.no_rawat : '',
        terisi && bacaTanggal(bed.tanggal_masuk) ? bed.tanggal_masuk : '',
        terisi && bed.lamaPenggunaanLabel ? bed.lama_digunakan_hari : '',
        kosong && bed.terakhirDipakaiLabel ? bed.terakhir_dipakai : '',
        kosong && bed.hariKosong !== null ? bed.hariKosong : '',
      ]
      return `<tr>${values.map(cell).join('')}</tr>`
    }).join('\n')
    const html = `<!DOCTYPE html>
      <html><head><meta charset="UTF-8"></head><body>
        <h2>Monitoring Ketersediaan Bed</h2>
        <p>Data dimuat: ${escape(waktuDimuat.value)}</p>
        <p>Status: ${escape(filterDimuat.value.status ? labelStatus(filterDimuat.value.status) : 'Semua status')}</p>
        <p>Pencarian: ${escape(filterDimuat.value.q || '-')}</p>
        <table border="1">
          <tr>${headers.map((label) => `<th>${escape(label)}</th>`).join('')}</tr>
          ${rows}
        </table>
      </body></html>`
    const url = URL.createObjectURL(new Blob(['\ufeff', html], {
      type: 'application/vnd.ms-excel;charset=utf-8',
    }))
    const link = document.createElement('a')
    link.href = url
    link.download = `monitoring-bed-${new Date().toLocaleDateString('en-CA')}.xls`
    link.click()
    window.setTimeout(() => URL.revokeObjectURL(url), 1000)
  }

  async function muatData() {
    if (loading.value) return
    loading.value = true
    error.value = ''
    const parameter = { ...filter.value }
    try {
      const response = await monitoringBedData(token, parameter)
      data.value = response.data || []
      firstRow.value = 0
      filterDimuat.value = parameter
      waktuDimuat.value = new Date().toLocaleString('id-ID')
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Gagal memuat data monitoring bed.'
      notifikasi.gagal(error.value)
      data.value = []
      waktuDimuat.value = ''
    } finally {
      loading.value = false
    }
  }

  function resetFilter() {
    filter.value = { status: '', q: '' }
    muatData()
  }

  onMounted(() => {
    if (token) muatData()
  })

  return {
    loading,
    error,
    grafikVisible,
    firstRow,
    waktuDimuat,
    ringkasan,
    grafik,
    gantiHalaman,
    excel,
    data,
    daftarBed,
    filter,
    muatData,
    resetFilter
  }
}
