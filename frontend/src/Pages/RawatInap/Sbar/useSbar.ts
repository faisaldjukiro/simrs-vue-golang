import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { request } from '../../../lib/shared/http'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import type { PropsSbar, DataSbar, CatatanSbar, HasilSbar, PilihanSbar } from '../../../types/sbar'

export function useSbar(props: PropsSbar) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const errorSimpan = ref('')
  const terbuka = ref(true)
  const keyword = ref('')
  const statusFilter = ref('')
  const hasil = ref<HasilSbar | null>(null)
  const editing = ref<CatatanSbar | null>(null)
  const detail = ref<CatatanSbar | null>(null)
  const target = ref<CatatanSbar | null>(null)
  const aksi = ref<'hapus' | 'verifikasi'>('hapus')
  const petugas = ref<Record<string, string>>({})
  const validator = ref<Record<string, string>>({})
  const form = reactive<DataSbar>({ tgl_perawatan: '', jam_rawat: '', nip: '', situation: '', background: '', assesment: '', recommendation: '', instruksi: '' })
  let generasi = 0
  let urutan = 0
  let controller: AbortController | undefined
  const terkunci = computed(() => loading.value || saving.value || !!error.value || !hasil.value)
  const rows = computed(() => (hasil.value?.catatan || []).filter(r => {
    const status = r.status || (r.terkunci ? 'Terkunci' : 'Belum diverifikasi')
    const cocok = !statusFilter.value || (statusFilter.value === 'belum' ? !r.terkunci : r.status === 'Validasi')
    return cocok && [Object.values(r.data).join(' '), r.nama_petugas, r.nama_validator, status].join(' ').toLowerCase().includes(keyword.value.trim().toLowerCase())
  }))
  const ringkasan = computed(() => {
    const daftar = hasil.value?.catatan || []
    return { total: daftar.length, validasi: daftar.filter(r => r.status === 'Validasi').length }
  })
  function api<T>(path = '', options: RequestInit = {}) {
    return request<T>('/api/sbar' + path, {
      ...options,
      headers: { Authorization: 'Bearer ' + props.token },
    })
  }
  function waktuSekarang() {
    const waktu = new Date(Date.now() + 8 * 3600000).toISOString()
    form.tgl_perawatan = waktu.slice(0, 10)
    form.jam_rawat = waktu.slice(11, 19)
  }
  function reset() {
    editing.value = null
    errorSimpan.value = ''
    form.situation = ''
    form.background = ''
    form.assesment = ''
    form.recommendation = ''
    form.instruksi = ''
    form.nip = ''
    petugas.value = { ...hasil.value?.petugas_login }
    waktuSekarang()
  }
  async function muat() {
    const id = ++urutan
    const konteks = generasi
    controller?.abort()
    controller = new AbortController()
    loading.value = true
    error.value = ''
    try {
      const data = await api<HasilSbar>('?' + new URLSearchParams({ no_rawat: props.patient.no_rawat }), { signal: controller.signal })
      if (id !== urutan || konteks !== generasi) return
      hasil.value = data
      if (!editing.value) petugas.value = { ...data.petugas_login }
    } catch (e) {
      if (id === urutan && konteks === generasi) error.value = e instanceof Error ? e.message : 'Gagal memuat SBAR'
    } finally {
      if (id === urutan && konteks === generasi) loading.value = false
    }
  }
  function cariPetugas(q: string) {
    return api<PilihanSbar[]>('/referensi?' + new URLSearchParams({ jenis: 'petugas', q }))
  }
  function cariDokter(q: string) {
    return api<PilihanSbar[]>('/referensi?' + new URLSearchParams({ jenis: 'dokter', q }))
  }
  function edit(r: CatatanSbar) {
    if (terkunci.value || !r.bisa_ubah) return
    editing.value = { ...r, data: { ...r.data } }
    Object.assign(form, r.data)
    petugas.value = { kode: r.data.nip, nama: r.nama_petugas, jabatan: r.jabatan }
    terbuka.value = true
    errorSimpan.value = ''
  }
  function konfirmasi(r: CatatanSbar, jenis: 'hapus' | 'verifikasi') {
    if (terkunci.value || (jenis === 'hapus' ? !r.bisa_ubah : !r.bisa_verifikasi)) return
    target.value = { ...r, data: { ...r.data } }
    aksi.value = jenis
    validator.value = { ...hasil.value?.dokter_login }
    errorSimpan.value = ''
  }
  async function simpan(dariDialog = false) {
    if (terkunci.value || (dariDialog && !target.value)) return
    const verifikasi = dariDialog && aksi.value === 'verifikasi'
    if (!dariDialog && !petugas.value.kode) {
      errorSimpan.value = 'Pilih petugas. Akun petugas harus sesuai NIK pegawai SIMRS.'
      return
    }
    if (verifikasi && !validator.value.kode) {
      errorSimpan.value = 'Dokter validator wajib dipilih.'
      return
    }
    const konteks = generasi
    const catatan = dariDialog ? target.value : editing.value
    saving.value = true
    errorSimpan.value = ''
    try {
      const res = await api<{ pesan: string }>(verifikasi ? '/verifikasi' : '', {
        method: verifikasi ? 'POST' : dariDialog ? 'DELETE' : editing.value ? 'PUT' : 'POST',
        body: JSON.stringify({
          no_rawat: props.patient.no_rawat,
          data: dariDialog ? undefined : {
            ...form,
            nip: petugas.value.kode,
            jam_rawat: form.jam_rawat.length === 5 ? form.jam_rawat + ':00' : form.jam_rawat,
          },
          asli: catatan ? { tgl_perawatan: catatan.data.tgl_perawatan, jam_rawat: catatan.data.jam_rawat } : undefined,
          revisi: catatan?.revisi,
          validator: verifikasi ? validator.value.kode : undefined,
        }),
      })
      if (konteks !== generasi) return
      target.value = null
      detail.value = null
      reset()
      notifikasi.sukses(res.pesan)
      await muat()
    } catch (e) {
      if (konteks === generasi) errorSimpan.value = e instanceof Error ? e.message : 'SBAR gagal diproses'
    } finally {
      if (konteks === generasi) saving.value = false
    }
  }
  watch(() => [props.token, props.patient.no_rawat], () => {
    generasi++
    hasil.value = null
    target.value = null
    detail.value = null
    keyword.value = ''
    statusFilter.value = ''
    saving.value = false
    reset()
    void muat()
  }, { immediate: true })
  onBeforeUnmount(() => { generasi++; urutan++; controller?.abort() })
  return {
    loading, saving, error, errorSimpan, terbuka, keyword, statusFilter, hasil,
    editing, detail, target, aksi, petugas, validator, form, terkunci, rows, ringkasan,
    muat, reset, waktuSekarang, cariPetugas, cariDokter, edit, konfirmasi, simpan,
  }
}
