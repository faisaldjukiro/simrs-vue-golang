import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { request } from '../../../lib/shared/http'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import { bidangEdukasi } from '../../../types/edukasiPasien'
import type { CatatanEdukasi, HasilEdukasi, PropsEdukasi } from '../../../types/edukasiPasien'

export function useEdukasiPasien(props: PropsEdukasi) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const errorSimpan = ref('')
  const keyword = ref('')
  const mulai = ref('')
  const selesai = ref('')
  const formVisible = ref(true)
  const records = ref<CatatanEdukasi[]>([])
  const editing = ref<CatatanEdukasi | null>(null)
  const detail = ref<CatatanEdukasi | null>(null)
  const hapusTarget = ref<CatatanEdukasi | null>(null)
  const petugas = ref<Record<string, string>>({})
  const ruangan = ref<Record<string, string>>({})
  const petugasLogin = ref<Record<string, string>>({})
  const bolehPilihPetugas = ref(false)
  const form = reactive<Record<string, string>>({})
  const foto = ref<File | null>(null)
  const fotoLokal = ref('')
  const kunciFoto = ref(0)
  const fotoGagal = ref(false)
  const detailFotoGagal = ref(false)
  const pratinjauFoto = computed(() => fotoLokal.value || editing.value?.foto_url || '')
  watch(pratinjauFoto, () => { fotoGagal.value = false })
  watch(detail, () => { detailFotoGagal.value = false })

  function batalFoto() {
    if (fotoLokal.value) URL.revokeObjectURL(fotoLokal.value)
    fotoLokal.value = ''
    foto.value = null
    kunciFoto.value++
  }

  function pilihFoto(event: Event) {
    const file = (event.target as HTMLInputElement).files?.[0]
    batalFoto()
    if (!file) return
    if (!['image/jpeg', 'image/png'].includes(file.type) || !/\.(jpe?g|png)$/i.test(file.name)) {
      errorSimpan.value = 'Foto harus berupa JPG atau PNG.'
      return
    }
    if (!file.size || file.size > 10 * 1024 * 1024) {
      errorSimpan.value = 'Ukuran foto maksimal 10 MB.'
      return
    }
    errorSimpan.value = ''
    foto.value = file
    fotoLokal.value = URL.createObjectURL(file)
  }
  let generasi = 0
  let urutan = 0
  const errorFilter = computed(() => mulai.value && selesai.value && mulai.value > selesai.value
    ? 'Tanggal mulai tidak boleh melewati tanggal selesai.' : '')
  const rows = computed(() => records.value.filter(r =>
    !errorFilter.value &&
    (!mulai.value || r.data.tgl_perawatan >= mulai.value) &&
    (!selesai.value || r.data.tgl_perawatan <= selesai.value) &&
    [r.nama_petugas, r.nama_ruangan, r.sumber, ...Object.values(r.data)].join(' ').toLocaleLowerCase()
      .includes(keyword.value.trim().toLocaleLowerCase()),
  ).map(r => ({ ...r, kunci: r.sumber + ':' + r.data.tgl_perawatan + ' ' + r.data.jam_rawat })))
  const terkunci = computed(() => loading.value || saving.value || !!error.value)

  function api<T>(path = '', options: RequestInit = {}) {
    return request<T>('/api/edukasi-pasien' + path, {
      ...options,
      headers: { Authorization: 'Bearer ' + props.token },
    })
  }

  function waktuSekarang() {
    const sekarang = new Date(Date.now() + 8 * 3600000).toISOString()
    form.tgl_perawatan = sekarang.slice(0, 10)
    form.jam_rawat = sekarang.slice(11, 19)
  }

  function reset() {
    batalFoto()
    editing.value = null
    errorSimpan.value = ''
    if (!petugas.value.kode) petugas.value = { ...petugasLogin.value }
    const metode = form.metode || 'Audio'
    Object.keys(form).forEach(k => delete form[k])
    for (const b of bidangEdukasi) form[b.key] = ''
    form.foto = ''
    form.metode = metode
    waktuSekarang()
  }

  async function muat() {
    const id = ++urutan
    const konteks = generasi
    loading.value = true
    error.value = ''
    try {
      const hasil = await api<HasilEdukasi>('?' + new URLSearchParams({ no_rawat: props.patient.no_rawat }))
      if (konteks !== generasi || id !== urutan) return
      records.value = hasil.catatan
      petugasLogin.value = hasil.petugas_login
      bolehPilihPetugas.value = hasil.boleh_pilih_petugas
      if (!editing.value && !petugas.value.kode) petugas.value = { ...hasil.petugas_login }
    } catch (e) {
      if (konteks === generasi && id === urutan) error.value = e instanceof Error ? e.message : 'Gagal memuat edukasi.'
    } finally {
      if (konteks === generasi && id === urutan) loading.value = false
    }
  }

  function cariReferensi(jenis: 'petugas' | 'ruangan', q: string) {
    return api<Record<string, string>[]>('/referensi?' + new URLSearchParams({ jenis, q }))
  }

  function cetak() {
    if (loading.value || error.value || !rows.value.length) return
    const popup = window.open('', '_blank', 'width=1100,height=750')
    if (!popup) {
      notifikasi.peringatan('Izinkan jendela cetak pada browser terlebih dahulu.')
      return
    }
    const doc = popup.document
    doc.title = 'Catatan Edukasi Pasien'
    doc.documentElement.lang = 'id'
    const style = doc.createElement('style')
    style.textContent = '@page { size: A4 landscape; margin: 12mm; } body { font: 12px Arial, sans-serif; color: #111; } h1 { font-size: 20px; } table { width: 100%; border-collapse: collapse; } th, td { border: 1px solid #aaa; padding: 8px; text-align: left; overflow-wrap: anywhere; } th { background: #eee; } thead { display: table-header-group; } tr { break-inside: avoid; }'
    doc.head.append(style)
    const title = doc.createElement('h1')
    title.textContent = 'Catatan Edukasi Pasien'
    const identity = doc.createElement('p')
    identity.textContent = `${props.patient.nm_pasien || '-'} | RM: ${props.patient.no_rkm_medis || '-'} | No. Rawat: ${props.patient.no_rawat}`
    const period = doc.createElement('p')
    period.textContent = `Periode: ${mulai.value || 'Awal kunjungan'} s.d. ${selesai.value || 'Terakhir'} | Pencarian: ${keyword.value || '-'} | ${rows.value.length} catatan | Waktu WITA`
    const table = doc.createElement('table')
    const head = table.createTHead().insertRow()
    for (const label of ['Tanggal', 'Jam', ...bidangEdukasi.map(b => b.label), 'Petugas', 'Ruangan', 'Sumber']) {
      const th = doc.createElement('th')
      th.textContent = label
      head.append(th)
    }
    const body = table.createTBody()
    for (const r of rows.value) {
      const tr = body.insertRow()
      for (const value of [r.data.tgl_perawatan, r.data.jam_rawat, ...bidangEdukasi.map(b => r.data[b.key]), `${r.nama_petugas || '-'} (${r.data.nip})`, r.nama_ruangan || r.data.kd_ruangan, r.sumber]) {
        tr.insertCell().textContent = value || '-'
      }
    }
    doc.body.append(title, identity, period, table)
    popup.focus()
    popup.requestAnimationFrame(() => popup.print())
  }

  function edit(row: CatatanEdukasi) {
    if (terkunci.value || !row.bisa_ubah) return
    reset()
    editing.value = { ...row, data: { ...row.data } }
    formVisible.value = true
    Object.assign(form, row.data)
    petugas.value = { kode: row.data.nip, nama: row.nama_petugas }
    ruangan.value = { kode: row.data.kd_ruangan, nama: row.nama_ruangan }
  }

  async function mutasi(hapus = false) {
    if (terkunci.value || (hapus && !hapusTarget.value)) return
    errorSimpan.value = ''
    if (!hapus && (!petugas.value.kode || !form.tgl_perawatan || !form.jam_rawat)) {
      errorSimpan.value = 'Lengkapi tanggal, jam, dan petugas.'
      return
    }
    const konteks = generasi
    saving.value = true
    try {
      const payload = JSON.stringify({
        no_rawat: props.patient.no_rawat,
        sumber: hapus ? hapusTarget.value?.sumber : editing.value?.sumber,
        data: hapus ? undefined : { ...form, nip: petugas.value.kode, kd_ruangan: ruangan.value.kode || '' },
        asli: hapus ? hapusTarget.value?.data : editing.value?.data,
      })
      let body: string | FormData = payload
      if (!hapus && foto.value) {
        body = new FormData()
        body.append('payload', payload)
        body.append('foto', foto.value)
      }
      const hasil = await api<{ pesan: string }>('', {
        method: hapus ? 'DELETE' : editing.value ? 'PUT' : 'POST',
        body,
      })
      if (konteks !== generasi) return
      hapusTarget.value = null
      reset()
      notifikasi.sukses(hasil.pesan)
      await muat()
    } catch (e) {
      if (konteks === generasi) errorSimpan.value = e instanceof Error ? e.message : 'Edukasi gagal disimpan.'
    } finally {
      if (konteks === generasi) saving.value = false
    }
  }

  watch(() => [props.token, props.patient.no_rawat], () => {
    generasi++
    saving.value = false
    records.value = []
    petugasLogin.value = {}
    petugas.value = {}
    ruangan.value = {}
    form.metode = 'Audio'
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
  onBeforeUnmount(() => { generasi++; batalFoto() })

  return {
    loading, saving, error, errorSimpan, keyword, rows, records, formVisible, editing, detail,
    hapusTarget, petugas, ruangan, petugasLogin, bolehPilihPetugas, form, terkunci,
    mulai, selesai, errorFilter, cetak,
    waktuSekarang, reset, muat, cariReferensi, edit, mutasi,
    foto, kunciFoto, pratinjauFoto, fotoGagal, detailFotoGagal, pilihFoto, batalFoto,
  }
}
