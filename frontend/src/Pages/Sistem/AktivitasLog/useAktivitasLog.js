import { computed, onMounted, reactive, ref } from "vue"
import { aktivitasLogData } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useAktivitasLog(props) {
  const notifikasi = useNotifikasi()
  const hariIni = new Date()
  const tujuhHariLalu = new Date(hariIni)
  tujuhHariLalu.setDate(tujuhHariLalu.getDate() - 6)
  
  const loading = ref(false)
  const error = ref('')
  const hasil = ref({ data: [], halaman: 1, batas: 25, total: 0, total_halaman: 0 })
  const detail = ref(null)
  const filter = reactive({
    tanggal_mulai: tujuhHariLalu,
    tanggal_selesai: hariIni,
    kata_kunci: '',
    username: '',
    aksi: '',
    modul: '',
    berhasil: '',
  })
  
  const pilihanAksi = [
    { label: 'Semua Aktivitas', value: '' },
    ...['LOGIN', 'LOGOUT', 'LIHAT', 'TAMBAH', 'UBAH', 'HAPUS', 'PROSES'].map((value) => ({ label: value, value })),
  ]
  const pilihanHasil = [
    { label: 'Semua Hasil', value: '' },
    { label: 'Berhasil', value: '1' },
    { label: 'Gagal', value: '0' },
  ]
  
  const jumlahBerhasil = computed(() => hasil.value.data.filter((item) => item.berhasil).length)
  const jumlahGagal = computed(() => hasil.value.data.filter((item) => !item.berhasil).length)
  
  function tanggalAPI(value) {
    if (!value) return ''
    const date = value instanceof Date ? value : new Date(value)
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  }
  
  function waktu(value) {
    if (!value) return '-'
    return new Intl.DateTimeFormat('id-ID', {
      dateStyle: 'medium',
      timeStyle: 'medium',
      timeZone: 'Asia/Makassar',
    }).format(new Date(value))
  }
  
  function parameter(halaman = 1, batas = hasil.value.batas || 25) {
    return {
      halaman,
      batas,
      tanggal_mulai: tanggalAPI(filter.tanggal_mulai),
      tanggal_selesai: tanggalAPI(filter.tanggal_selesai),
      kata_kunci: filter.kata_kunci.trim(),
      username: filter.username.trim(),
      aksi: filter.aksi,
      modul: filter.modul.trim(),
      berhasil: filter.berhasil,
    }
  }
  
  async function muat(halaman = 1, batas = hasil.value.batas || 25) {
    if (loading.value || !props.token) return
    loading.value = true
    error.value = ''
    try {
      hasil.value = await aktivitasLogData(props.token, parameter(halaman, batas))
    } catch (err) {
      error.value = err.message
      notifikasi.gagal(err.message || 'Log aktivitas tidak dapat dibaca.')
    } finally {
      loading.value = false
    }
  }
  
  function gantiHalaman(event) {
    muat((event.page ?? 0) + 1, event.rows ?? hasil.value.batas)
  }
  
  function jsonRapi(value) {
    if (!value) return 'Tidak ada data.'
    return JSON.stringify(value, null, 2)
  }
  
  function ubahDialogDetail(terbuka) {
    if (!terbuka) detail.value = null
  }
  
  onMounted(() => muat())
  return {
    loading,
    error,
    hasil,
    detail,
    filter,
    pilihanAksi,
    pilihanHasil,
    jumlahBerhasil,
    jumlahGagal,
    waktu,
    muat,
    gantiHalaman,
    jsonRapi,
    ubahDialogDetail,
  }
}
