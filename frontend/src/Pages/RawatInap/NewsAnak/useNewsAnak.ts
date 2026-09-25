import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { request } from '../../../lib/shared/http'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import type { BidangNews, CatatanNews, HasilNews, PanduanNews, PropsNewsAnak } from '../../../types/newsAnak'

export function useNewsAnak(props: PropsNewsAnak) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const errorSimpan = ref('')
  const keyword = ref('')
  const mulai = ref('')
  const selesai = ref('')
  const formVisible = ref(true)
  const records = ref<CatatanNews[]>([])
  const bidang = ref<BidangNews[]>([])
  const panduan = ref<PanduanNews[]>([])
  const editing = ref<CatatanNews | null>(null)
  const detail = ref<CatatanNews | null>(null)
  const hapusTarget = ref<CatatanNews | null>(null)
  const petugas = ref<Record<string, string>>({})
  const petugasLogin = ref<Record<string, string>>({})
  const bolehPilihPetugas = ref(false)
  const form = reactive<Record<string, string>>({})
  let generasi = 0
  let urutan = 0

  const errorFilter = computed(() => mulai.value && selesai.value && mulai.value > selesai.value
    ? 'Tanggal mulai tidak boleh melewati tanggal selesai.' : '')
  const rows = computed(() => records.value.filter(r => {
    const tanggal = r.data.tanggal.slice(0, 10)
    return !errorFilter.value
      && (!mulai.value || tanggal >= mulai.value)
      && (!selesai.value || tanggal <= selesai.value)
      && [r.nama_petugas, ...Object.values(r.data)].join(' ').toLocaleLowerCase()
        .includes(keyword.value.trim().toLocaleLowerCase())
  }).map(r => ({ ...r, kunci: r.data.tanggal })))
  const terkunci = computed(() => loading.value || saving.value || !!error.value)
  const skor = computed(() => bidang.value.map(b => b.nilai[b.pilihan.indexOf(form[b.key] || '')] ?? -1))
  const lengkap = computed(() => skor.value.length === 5 && skor.value.every(s => s >= 0))
  const jumlahTerisi = computed(() => skor.value.filter(s => s >= 0).length)
  const total = computed(() => jumlahTerisi.value ? skor.value.reduce((sum, nilai) => sum + Math.max(nilai, 0), 0) : null)
  const kodeRespon = computed(() => {
    if (!lengkap.value) return ''
    return kategori(skor.value)
  })
  const hasilPanduan = computed(() => panduan.value.find(p => p.kode === kodeRespon.value))
  const parameter = computed(() => hasilPanduan.value?.respon || '')
  const eskalasi = computed(() => hasilPanduan.value?.eskalasi || '')
  const perluVerifikasi = computed(() => lengkap.value && skor.value.every(n => n === 1))
  function kategori(nilai: number[]) {
    const kuning = nilai.filter(n => n === 1).length
    if (nilai.includes(3) || nilai.reduce((a, b) => a + b, 0) > 5) return 'merah'
    if ((kuning >= 3 && kuning <= 4) || nilai.includes(2)) return 'oranye'
    if (kuning >= 1 && kuning <= 2) return 'kuning'
    return 'hijau'
  }

  function warnaParameter(nilai: string | number | undefined) {
    if (nilai === undefined || nilai === '') return 'pending'
    const n = Number(nilai)
    return ['score-green', 'score-yellow', 'score-orange', 'score-red'][n] || 'pending'
  }

  const warnaKategori: Record<string, string> = {
    merah: 'danger', oranye: 'warning', kuning: 'info', hijau: 'safe',
  }
  function warnaSkor(data: Record<string, string>) {
    const nilai = bidang.value.map(b => Number(data[b.skor]))
    if (nilai.length !== 5 || bidang.value.some(b => !data[b.skor]) || nilai.some(n => !Number.isInteger(n) || n < 0 || n > 3) || nilai.every(n => n === 1)) return 'pending'
    return warnaKategori[kategori(nilai)]
  }
  const warnaTotal = computed(() => perluVerifikasi.value ? 'pending' : warnaKategori[kodeRespon.value] || 'pending')

  function verifikasiCatatan(data: Record<string, string>) {
    return bidang.value.length === 5 && bidang.value.every(b => data[b.skor] === '1')
  }

  function api<T>(path = '', options: RequestInit = {}) {
    return request<T>('/api/news-anak' + path, {
      ...options,
      headers: { Authorization: 'Bearer ' + props.token },
    })
  }

  function waktuSekarang() {
    const sekarang = new Date(Date.now() + 8 * 3600000).toISOString()
    form.tanggal = sekarang.slice(0, 10)
    form.jam = sekarang.slice(11, 19)
  }

  function reset() {
    editing.value = null
    errorSimpan.value = ''
    petugas.value = { ...petugasLogin.value }
    Object.keys(form).forEach(k => delete form[k])
    waktuSekarang()
  }

  async function muat() {
    const id = ++urutan
    const konteks = generasi
    loading.value = true
    error.value = ''
    try {
      const hasil = await api<HasilNews>('?' + new URLSearchParams({ no_rawat: props.patient.no_rawat }))
      if (konteks !== generasi || id !== urutan) return
      records.value = hasil.catatan
      bidang.value = hasil.bidang
      panduan.value = hasil.panduan
      petugasLogin.value = hasil.petugas_login
      bolehPilihPetugas.value = hasil.boleh_pilih_petugas
      if (!petugas.value.kode) petugas.value = { ...hasil.petugas_login }
    } catch (e) {
      if (konteks === generasi && id === urutan) {
        error.value = e instanceof Error ? e.message : 'Gagal memuat NEWS Anak.'
      }
    } finally {
      if (konteks === generasi && id === urutan) loading.value = false
    }
  }

  function cariPetugas(q: string) {
    return api<Record<string, string>[]>('/petugas?' + new URLSearchParams({ q }))
  }

  function edit(row: CatatanNews) {
    if (terkunci.value || !row.bisa_ubah) return
    reset()
    editing.value = { ...row, data: { ...row.data } }
    formVisible.value = true
    Object.assign(form, row.data, {
      tanggal: row.data.tanggal.slice(0, 10),
      jam: row.data.tanggal.slice(11, 19),
    })
    petugas.value = { kode: row.data.nip, nama: row.nama_petugas }
    void nextTick(() => document.getElementById('news-form')?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
  }

  async function mutasi(hapus = false) {
    if (terkunci.value || (hapus && !hapusTarget.value)) return
    errorSimpan.value = ''
    if (!hapus && (!petugas.value.kode || !form.tanggal || !form.jam || !lengkap.value)) {
      errorSimpan.value = 'Lengkapi tanggal, jam, petugas, dan kelima parameter penilaian.'
      return
    }
    const konteks = generasi
    saving.value = true
    try {
      const jam = form.jam?.length === 5 ? form.jam + ':00' : form.jam
      const hasil = await api<{ pesan: string }>('', {
        method: hapus ? 'DELETE' : editing.value ? 'PUT' : 'POST',
        body: JSON.stringify({
          no_rawat: props.patient.no_rawat,
          data: hapus ? undefined : { ...form, tanggal: `${form.tanggal} ${jam}`, nip: petugas.value.kode },
          asli: hapus ? hapusTarget.value?.data : editing.value?.data,
        }),
      })
      if (konteks !== generasi) return
      hapusTarget.value = null
      reset()
      notifikasi.sukses(hasil.pesan)
      await muat()
    } catch (e) {
      if (konteks === generasi) errorSimpan.value = e instanceof Error ? e.message : 'NEWS Anak gagal disimpan.'
    } finally {
      if (konteks === generasi) saving.value = false
    }
  }

  function cetak() {
    if (terkunci.value || !rows.value.length) return
    const popup = window.open('', '_blank', 'width=1100,height=750')
    if (!popup) {
      notifikasi.peringatan('Izinkan jendela cetak pada browser terlebih dahulu.')
      return
    }
    const doc = popup.document
    doc.title = 'Lembar Pemantauan NEWS Anak'
    doc.documentElement.lang = 'id'
    const style = doc.createElement('style')
    style.textContent = '@page { size: A4 landscape; margin: 12mm; } body { font: 12px Arial, sans-serif; color: #111; } h1 { font-size: 20px; } table { width: 100%; border-collapse: collapse; } th, td { border: 1px solid #aaa; padding: 8px; text-align: left; overflow-wrap: anywhere; } th { background: #eee; } thead { display: table-header-group; } tr { break-inside: avoid; }'
    doc.head.append(style)
    const title = doc.createElement('h1')
    title.textContent = doc.title
    const identity = doc.createElement('p')
    identity.textContent = `${props.patient.nm_pasien || '-'} | RM: ${props.patient.no_rkm_medis || '-'} | No. Rawat: ${props.patient.no_rawat}`
    const period = doc.createElement('p')
    period.textContent = `Periode: ${mulai.value || 'Awal kunjungan'} s.d. ${selesai.value || 'Terakhir'} | Pencarian: ${keyword.value || '-'} | ${rows.value.length} catatan | WITA`
    const table = doc.createElement('table')
    const head = table.createTHead().insertRow()
    for (const label of ['Tanggal / Jam', ...bidang.value.map(b => b.label + ' (Skor)'), 'Total', 'Respon', 'Proses Eskalasi', 'Petugas']) {
      const th = doc.createElement('th')
      th.textContent = label
      head.append(th)
    }
    const body = table.createTBody()
    for (const r of rows.value) {
      const tr = body.insertRow()
      const responCetak = (verifikasiCatatan(r.data) ? '[PERLU VERIFIKASI KLINIS: lima skor 1 menggunakan respon bawaan legacy] ' : '') + r.data.respon
      for (const value of [r.data.tanggal, ...bidang.value.map(b => `${r.data[b.key]} (${r.data[b.skor]})`), r.data.skor_total, responCetak, r.data.proses_eskalasi, `${r.nama_petugas || '-'} (${r.data.nip})`]) {
        tr.insertCell().textContent = value || '-'
      }
    }
    doc.body.append(title, identity, period, table)
    popup.focus()
    popup.requestAnimationFrame(() => popup.print())
  }

  watch(() => [props.token, props.patient.no_rawat], () => {
    generasi++
    saving.value = false
    records.value = []
    bidang.value = []
    panduan.value = []
    petugasLogin.value = {}
    bolehPilihPetugas.value = false
    detail.value = null
    hapusTarget.value = null
    keyword.value = ''
    mulai.value = ''
    selesai.value = ''
    formVisible.value = true
    reset()
    void muat()
  }, { immediate: true })
  onBeforeUnmount(() => { generasi++ })

  return {
    loading, saving, error, errorSimpan, keyword, rows, records, formVisible, editing, detail,
    hapusTarget, petugas, petugasLogin, bolehPilihPetugas, form, terkunci, bidang, panduan,
    eskalasi, perluVerifikasi, warnaTotal, skor, lengkap, jumlahTerisi, total, parameter, warnaSkor, warnaParameter, mulai, selesai, errorFilter, cetak,
    waktuSekarang, reset, muat, cariPetugas, edit, mutasi, verifikasiCatatan,
  }
}
