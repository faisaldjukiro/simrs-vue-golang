import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { cetakEdukasiPasien } from './cetakEdukasiPasien'
import type { KopPemulangan } from '../../../types/perencanaanPemulangan'
import { request } from '../../../lib/shared/http'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import { bidangEdukasi, bidangAsesmen, bidangVerifikasi } from '../../../types/edukasiPasien'
import type { CatatanEdukasi, HasilEdukasi, PropsEdukasi } from '../../../types/edukasiPasien'

export function useEdukasiPasien(props: PropsEdukasi) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const printing = ref(false)
  const cetakAktif = ref('')
  const error = ref('')
  const errorSimpan = ref('')
  const keyword = ref('')
  const mulai = ref('')
  const selesai = ref('')
  const formVisible = ref(true)
  const records = ref<CatatanEdukasi[]>([])
  const editing = ref<CatatanEdukasi | null>(null)
  const detail = ref<CatatanEdukasi | null>(null)
  const detailCatatan = ref<CatatanEdukasi | null>(null)
  const verifikator = ref<Record<string, string>>({})
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
    [r.nama_petugas, r.nama_ruangan, r.nama_verifikator, r.sumber, ...Object.values(r.data)].join(' ').toLocaleLowerCase()
      .includes(keyword.value.trim().toLocaleLowerCase()),
  ).map(r => ({ ...r, kunci: r.sumber + ':' + r.data.tgl_perawatan + ' ' + r.data.jam_rawat })))
  const terkunci = computed(() => loading.value || saving.value || !!error.value)
  const terverifikasi = computed(() => form.status_verifikasi === 'Terverifikasi')

  function waktuVerifikasiSekarang() {
    form.tanggal_verifikasi = new Date(Date.now() + 8 * 3600000).toISOString().slice(0, 19)
  }

  function ubahStatusVerifikasi(value: string) {
    form.status_verifikasi = value
    if (value === 'Terverifikasi') {
      if (!form.tanggal_verifikasi) waktuVerifikasiSekarang()
      if (!verifikator.value.kode) verifikator.value = { ...petugasLogin.value }
    } else {
      form.tanggal_verifikasi = ''
      verifikator.value = {}
      form.nip_verifikator = ''
    }
  }

  const bidangDetail = [
    { key: 'tgl_perawatan', label: 'Tanggal Edukasi' }, { key: 'jam_rawat', label: 'Jam Edukasi (WITA)' },
    ...bidangEdukasi, ...bidangAsesmen, ...bidangVerifikasi,
  ]

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
    for (const b of bidangDetail) form[b.key] = ''
    form.status_verifikasi = 'Belum diverifikasi'
    verifikator.value = {}
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

  async function cetak(row?: CatatanEdukasi) {
    if (terkunci.value || printing.value || (!row && !rows.value.length)) return
    const popup = window.open('', '_blank', 'width=1100,height=750')
    if (!popup) {
      notifikasi.peringatan('Izinkan jendela cetak pada browser terlebih dahulu.')
      return
    }
    const pasien = { ...props.patient }
    const catatan = (row ? [row] : rows.value).map(r => ({ ...r, data: { ...r.data } }))
    const konteks = generasi
    printing.value = true
    cetakAktif.value = row ? row.sumber + ':' + row.data.tgl_perawatan + ' ' + row.data.jam_rawat : 'semua'
    popup.document.body.textContent = 'Menyiapkan edukasi dan kop rumah sakit...'
    try {
      const kop = await request<KopPemulangan>('/api/perencanaan-pemulangan/kop', {
        headers: { Authorization: 'Bearer ' + props.token },
      })
      if (popup.closed) return
      if (konteks !== generasi) { popup.close(); return }
      await cetakEdukasiPasien(popup, pasien, catatan, kop, () => konteks === generasi)
    } catch (e) {
      if (!popup.closed) popup.close()
      if (konteks === generasi) notifikasi.peringatan(e instanceof Error ? e.message : 'Cetak edukasi gagal.')
    } finally {
      if (konteks === generasi) {
        printing.value = false
        cetakAktif.value = ''
      }
    }
  }

  function edit(row: CatatanEdukasi) {
    if (terkunci.value || !row.bisa_ubah) return
    reset()
    editing.value = { ...row, data: { ...row.data } }
    formVisible.value = true
    Object.assign(form, row.data)
    form.tanggal_verifikasi = (row.data.tanggal_verifikasi || '').replace(' ', 'T')
    verifikator.value = { kode: row.data.nip_verifikator, nama: row.nama_verifikator }
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
    if (!hapus && terverifikasi.value) {
      if (!form.tingkat_pemahaman || !form.tanggal_verifikasi || !verifikator.value.kode) {
        errorSimpan.value = 'Lengkapi tingkat pemahaman, waktu, dan petugas verifikasi.'
        return
      }
      const waktu = form.tanggal_verifikasi.replace('T', ' ')
      if ((waktu.length === 16 ? waktu + ':00' : waktu) < `${form.tgl_perawatan} ${form.jam_rawat.length === 5 ? form.jam_rawat + ':00' : form.jam_rawat}`) {
        errorSimpan.value = 'Waktu verifikasi tidak boleh sebelum waktu edukasi.'
        return
      }
    }
    const konteks = generasi
    saving.value = true
    try {
      const payload = JSON.stringify({
        no_rawat: props.patient.no_rawat,
        sumber: hapus ? hapusTarget.value?.sumber : editing.value?.sumber,
        data: hapus ? undefined : { ...form, nip: petugas.value.kode, kd_ruangan: ruangan.value.kode || '',
          tanggal_verifikasi: terverifikasi.value ? form.tanggal_verifikasi.replace('T', ' ') : '',
          nip_verifikator: terverifikasi.value ? verifikator.value.kode : '' },
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
    printing.value = false
    cetakAktif.value = ''
    saving.value = false
    records.value = []
    petugasLogin.value = {}
    petugas.value = {}
    ruangan.value = {}
    form.metode = 'Audio'
    bolehPilihPetugas.value = false
    detail.value = null
    detailCatatan.value = null
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
    mulai, selesai, errorFilter, cetak, printing, cetakAktif,
    waktuSekarang, reset, muat, cariReferensi, edit, mutasi,
    foto, kunciFoto, pratinjauFoto, fotoGagal, detailFotoGagal, pilihFoto, batalFoto,
    detailCatatan, bidangDetail, verifikator, terverifikasi, ubahStatusVerifikasi, waktuVerifikasiSekarang,
  }
}
