import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { request } from '../../../lib/shared/http'
import { cetakFormulirPemulangan } from './cetakPerencanaanPemulangan'
import type { KopPemulangan } from '../../../types/perencanaanPemulangan'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import { penilaianPemulangan, bantuanPemulangan } from '../../../types/perencanaanPemulangan'
import type { PropsPemulangan, CatatanPemulangan, HasilPemulangan } from '../../../types/perencanaanPemulangan'

export function usePerencanaanPemulangan(props: PropsPemulangan) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const errorSimpan = ref('')
  const terbuka = ref(true)
  const keyword = ref('')
  const records = ref<CatatanPemulangan[]>([])
  const editing = ref<CatatanPemulangan | null>(null)
  const hapusTarget = ref<CatatanPemulangan | null>(null)
  const petugas = ref<Record<string,string>>({})
  const petugasLogin = ref<Record<string,string>>({})
  const bolehPilih = ref(false)
  const form = reactive<Record<string,string>>({})
  let generasi = 0
  let urutan = 0
  const terkunci = computed(() => loading.value || saving.value || !!error.value)
  const rows = computed(() => records.value.filter(r => Object.values(r.data).join(' ').toLowerCase().includes(keyword.value.toLowerCase())))
  function api<T>(path = '', options: RequestInit = {}) {
    return request<T>('/api/perencanaan-pemulangan' + path, {
      ...options, headers: { Authorization: 'Bearer ' + props.token },
    })
  }
  function reset() {
    editing.value = null
    errorSimpan.value = ''
    Object.keys(form).forEach(k => delete form[k])
    form.rencana_pulang = new Date(Date.now() + 8 * 3600000).toISOString().slice(0,10)
    form.alasan_masuk = ''
    form.diagnosa_medis = ''
    form.nama_pasien_keluarga = ''
    for (const p of penilaianPemulangan) {
      form[p.key] = p.key === 'bantuan_diperlukan_dalam' ? bantuanPemulangan[0] : 'Tidak'
      form['keterangan_' + p.key] = ''
    }
    petugas.value = { ...petugasLogin.value }
  }
  async function muat() {
    const id = ++urutan
    const konteks = generasi
    loading.value = true
    error.value = ''
    try {
      const hasil = await api<HasilPemulangan>('?' + new URLSearchParams({ no_rawat: props.patient.no_rawat }))
      if (konteks !== generasi || id !== urutan) return
      records.value = hasil.catatan
      bolehPilih.value = hasil.boleh_pilih_petugas
      petugasLogin.value = hasil.petugas_login
      if (!editing.value) petugas.value = { ...hasil.petugas_login }
    } catch (e) {
      if (konteks === generasi && id === urutan) error.value = e instanceof Error ? e.message : 'Gagal memuat catatan'
    } finally {
      if (konteks === generasi && id === urutan) loading.value = false
    }
  }
  function cariPetugas(q: string) {
    return api<Record<string,string>[]>('/referensi?' + new URLSearchParams({ q }))
  }
  function edit(r: CatatanPemulangan) {
    if (terkunci.value || !r.bisa_ubah) return
    editing.value = { ...r, data: { ...r.data } }
    Object.assign(form,r.data)
    petugas.value = { kode: r.data.nip, nama: r.nama_petugas }
    terbuka.value = true
    errorSimpan.value = ''
  }
  async function simpan(hapus = false) {
    if (terkunci.value || (hapus && !hapusTarget.value)) return
    if (!hapus && !petugas.value.kode) {
      errorSimpan.value = 'Petugas wajib dipilih. Akun petugas harus sesuai NIP SIMRS.'
      return
    }
    const konteks = generasi
    saving.value = true
    errorSimpan.value = ''
    try {
      const hasil = await api<{pesan:string}>('',{
        method: hapus ? 'DELETE' : editing.value ? 'PUT' : 'POST',
        body: JSON.stringify({
          no_rawat: props.patient.no_rawat,
          data: hapus ? undefined : Object.fromEntries(Object.entries({ ...form, nip: petugas.value.kode }).map(([k,v]) => [k,String(v ?? '')])),
          asli: hapus ? hapusTarget.value?.data : editing.value?.data,
        }),
      })
      if (konteks !== generasi) return
      hapusTarget.value = null
      reset()
      notifikasi.sukses(hasil.pesan)
      await muat()
    } catch (e) {
      if (konteks === generasi) errorSimpan.value = e instanceof Error ? e.message : 'Gagal menyimpan'
    } finally {
      if (konteks === generasi) saving.value = false
    }
  }
  async function cetak(r: CatatanPemulangan) {
    if (loading.value || saving.value) return
    const win = window.open('', '_blank', 'width=1000,height=750')
    if (!win) { notifikasi.peringatan('Izinkan jendela cetak browser.'); return }
    const konteks = generasi
    const pasien = { ...props.patient }
    const catatan = { ...r, data: { ...r.data } }
    win.document.body.textContent = 'Menyiapkan formulir dan kop rumah sakit...'
    try {
      const kop = await api<KopPemulangan>('/kop')
      if (win.closed) return
      if (konteks !== generasi) { win.close(); return }
      await cetakFormulirPemulangan(win, pasien, catatan, kop)
    } catch (e) {
      if (!win.closed) win.close()
      if (konteks === generasi) notifikasi.peringatan(e instanceof Error ? e.message : 'Gagal menyiapkan cetakan.')
    }
  }
  watch(() => [props.token,props.patient.no_rawat],() => {
    generasi++
    records.value = []
    saving.value = false
    hapusTarget.value = null
    petugasLogin.value = {}
    keyword.value = ''
    reset()
    void muat()
  },{ immediate:true })
  onBeforeUnmount(() => { generasi++ })
  return { loading,saving,error,errorSimpan,terbuka,keyword,rows,editing,hapusTarget,petugas,bolehPilih,form,terkunci,reset,muat,cariPetugas,edit,simpan,cetak }
}
