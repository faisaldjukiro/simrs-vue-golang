import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { useBuktiEdukasi } from './useBuktiEdukasi'
import { cetakEdukasiPasien } from './cetakEdukasiPasien'
import type { KopPemulangan } from '../../../types/perencanaanPemulangan'
import { request } from '../../../lib/shared/http'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import { bidangEdukasi, bidangAsesmen, bidangVerifikasi } from '../../../types/edukasiPasien'
import type { BidangAsesmen, CatatanEdukasi, HasilEdukasi, PropsEdukasi } from '../../../types/edukasiPasien'

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
  const parafTersedia = ref(false)
  const parafDibatalkan = ref(false)
  const parafPenerimaTersedia = ref(false)
  const fotoPenerimaTersedia = ref(false)
  const namaPenerimaTersedia = ref(false)
  let mengisiForm = false
  const form = reactive<Record<string, string>>({})
  const isianLainnya = reactive<Record<string, boolean>>({})
  function pilihanBidang(bidang: BidangAsesmen) {
    return [
      { label: 'Belum dicatat', value: '' },
      ...(bidang.pilihan || []).map(value => ({ label: value, value })),
      { label: 'Lainnya / tulis sendiri', value: '__lainnya__' },
    ]
  }
  function nilaiPilihan(bidang: BidangAsesmen) {
    const nilai = form[bidang.key] || ''
    return isianLainnya[bidang.key] || (nilai && !bidang.pilihan?.includes(nilai)) ? '__lainnya__' : nilai
  }
  function pilihBidang(bidang: BidangAsesmen, nilai: string) {
    if (terkunci.value) return
    isianLainnya[bidang.key] = nilai === '__lainnya__'
    form[bidang.key] = nilai === '__lainnya__' ? '' : nilai
  }
  const materiTerpilih = computed(() => (form.materi || '').split('\n').map(v => v.trim()).filter(Boolean))
  function pilihMateri(materi: string, dipilih: boolean) {
    if (terkunci.value) return
    const baris = (form.materi || '').split('\n')
    const nilai = dipilih
      ? materiTerpilih.value.includes(materi) ? form.materi : [form.materi, materi].filter(Boolean).join('\n')
      : baris.filter(v => v.trim() !== materi).join('\n')
    if (nilai.length > 65535) {
      notifikasi.peringatan('Materi terlalu panjang. Ringkas isian sebelum menambahkan pilihan.')
      return
    }
    form.materi = nilai
  }
  const bukti = useBuktiEdukasi(form,
    key => (key === 'foto' ? editing.value?.foto_url : editing.value?.foto_penerima_url) || '',
    () => terkunci.value, pesan => { errorSimpan.value = pesan })
  const { jenis: jenisBukti, fotoPetugas, fotoPenerima, kameraAktif, buktiGanda,
    batalFoto, pilihJenis: pilihJenisBukti } = bukti
  const detailFotoGagal = ref(false)
  const detailFotoPenerimaGagal = ref(false)
  watch(detail, () => { detailFotoGagal.value = false; detailFotoPenerimaGagal.value = false })
  watch(formVisible, value => { if (!value) kameraAktif.value = '' })
  let generasi = 0
  let urutan = 0
  const errorFilter = computed(() => mulai.value && selesai.value && mulai.value > selesai.value
    ? 'Tanggal mulai tidak boleh melewati tanggal selesai.' : '')
  const rows = computed(() => records.value.filter(r =>
    !errorFilter.value &&
    (!mulai.value || r.data.tgl_perawatan >= mulai.value) &&
    (!selesai.value || r.data.tgl_perawatan <= selesai.value) &&
    [r.nama_petugas, r.nama_ruangan, r.nama_verifikator, r.sumber,
      ...Object.entries(r.data).filter(([k]) => !k.startsWith('paraf_')).map(([, v]) => v)].join(' ').toLocaleLowerCase()
      .includes(keyword.value.trim().toLocaleLowerCase()),
  ).map(r => ({ ...r, kunci: r.sumber + ':' + r.data.tgl_perawatan + ' ' + r.data.jam_rawat })))
  const terkunci = computed(() => loading.value || saving.value || !!error.value)
  const terverifikasi = computed(() => form.status_verifikasi === 'Terverifikasi')
  const kekuranganVerifikasi = computed(() => [
    ...bidangAsesmen,
    { key: 'materi', label: 'Materi Edukasi' }, { key: 'penerima', label: 'Penerima Edukasi' },
    { key: 'catatan_verifikasi', label: 'Catatan Hasil Verifikasi' },
    { key: 'tingkat_pemahaman', label: 'Tingkat Pemahaman' },
    { key: 'tanggal_verifikasi', label: 'Waktu Verifikasi' },
    { key: 'nama_penerima', label: 'Nama Penerima' },
  ].filter(b => !form[b.key]?.trim()).map(b => b.label)
    .concat(jenisBukti.value === 'foto' ? [
      ...(!fotoPetugas.ada ? ['Foto Petugas'] : []),
      ...(!fotoPenerima.ada ? ['Foto Penerima'] : []),
    ] : [
      ...(!form.paraf_petugas ? ['Paraf Petugas'] : []),
      ...(!form.paraf_penerima ? ['Paraf Penerima'] : []),
    ]))
  watch(() => [JSON.stringify(Object.keys(form).filter(k => !['paraf_petugas', 'paraf_penerima', 'foto', 'foto_penerima'].includes(k)).map(k => [k, form[k]])),
    petugas.value.kode, ruangan.value.kode, verifikator.value.kode], () => {
    if (!mengisiForm && (form.paraf_petugas || form.paraf_penerima)) {
      form.paraf_petugas = ''
      form.paraf_penerima = ''
      parafDibatalkan.value = true
    }
  }, { flush: 'sync' })

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
    { key: 'nama_penerima', label: 'Nama Penerima' },
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
    mengisiForm = true
    bukti.reset()
    Object.keys(isianLainnya).forEach(k => delete isianLainnya[k])
    editing.value = null
    errorSimpan.value = ''
    if (!petugas.value.kode) petugas.value = { ...petugasLogin.value }
    const metode = form.metode || 'Audio'
    Object.keys(form).forEach(k => delete form[k])
    for (const b of bidangDetail) form[b.key] = ''
    form.status_verifikasi = 'Belum diverifikasi'
    verifikator.value = {}
    form.foto = ''
    form.foto_penerima = ''
    form.paraf_petugas = ''
    form.paraf_penerima = ''
    parafDibatalkan.value = false
    jenisBukti.value = ''
    form.metode = metode
    waktuSekarang()
    mengisiForm = false
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
      parafTersedia.value = hasil.paraf_tersedia === true
      parafPenerimaTersedia.value = hasil.paraf_penerima_tersedia === true
      fotoPenerimaTersedia.value = hasil.foto_penerima_tersedia === true
      namaPenerimaTersedia.value = hasil.nama_penerima_tersedia === true
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
    mengisiForm = true
    editing.value = { ...row, data: { ...row.data } }
    formVisible.value = true
    Object.assign(form, row.data)
    form.tanggal_verifikasi = (row.data.tanggal_verifikasi || '').replace(' ', 'T')
    verifikator.value = { kode: row.data.nip_verifikator, nama: row.nama_verifikator }
    petugas.value = { kode: row.data.nip, nama: row.nama_petugas }
    ruangan.value = { kode: row.data.kd_ruangan, nama: row.nama_ruangan }
    bukti.muatJenis()
    mengisiForm = false
  }

  async function mutasi(hapus = false) {
    if (terkunci.value || (hapus && !hapusTarget.value)) return
    errorSimpan.value = ''
    if (!hapus && buktiGanda.value) {
      errorSimpan.value = 'Pilih dua paraf atau dua foto (petugas dan penerima), tidak boleh bersamaan.'
      return
    }
    if (!hapus && (!petugas.value.kode || !form.tgl_perawatan || !form.jam_rawat)) {
      errorSimpan.value = 'Lengkapi tanggal, jam, dan petugas.'
      return
    }
    if (!hapus && terverifikasi.value) {
      if (jenisBukti.value === 'foto' && (!fotoPenerimaTersedia.value || !namaPenerimaTersedia.value)) {
        errorSimpan.value = 'Penyimpanan foto penerima belum tersedia. Hubungi administrator untuk melengkapi kolom foto_penerima dan nama_penerima.'
        return
      }
      if (jenisBukti.value === 'paraf' && (!parafTersedia.value || !parafPenerimaTersedia.value)) {
        errorSimpan.value = 'Penyimpanan paraf belum tersedia. Pilih foto atau hubungi administrator.'
        return
      }
      if (kekuranganVerifikasi.value.length) {
        errorSimpan.value = 'Lengkapi sebelum verifikasi: ' + kekuranganVerifikasi.value.join(', ') + '.'
        return
      }
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
      if (!hapus && (fotoPetugas.file || fotoPenerima.file)) {
        body = new FormData()
        body.append('payload', payload)
        if (fotoPetugas.file) body.append('foto', fotoPetugas.file)
        if (fotoPenerima.file) body.append('foto_penerima', fotoPenerima.file)
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
    parafTersedia.value = false
    parafPenerimaTersedia.value = false
    fotoPenerimaTersedia.value = false
    namaPenerimaTersedia.value = false
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
    fotoPetugas, fotoPenerima, kameraAktif, detailFotoGagal, detailFotoPenerimaGagal,
    fotoPenerimaTersedia, namaPenerimaTersedia,
    detailCatatan, bidangDetail, verifikator, terverifikasi, ubahStatusVerifikasi, waktuVerifikasiSekarang,
    pilihanBidang, nilaiPilihan, pilihBidang, materiTerpilih, pilihMateri,
    parafTersedia, parafDibatalkan, kekuranganVerifikasi,
    jenisBukti, pilihJenisBukti, buktiGanda, parafPenerimaTersedia,
  }
}
