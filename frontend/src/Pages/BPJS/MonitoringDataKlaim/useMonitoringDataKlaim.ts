import { computed, onBeforeUnmount, ref, watch } from "vue"
import { monitoringDataKlaim } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"
import type { FilterKlaim, HasilMonitoringKlaim } from '../../../types/monitoringDataKlaim'
import { unduhExcelKlaim } from './excelKlaim'

export function useMonitoringDataKlaim(props: { token: string }) {
  const notifikasi = useNotifikasi()
  const hariIni = new Date()
  const tanggalMulai = ref(new Date(hariIni.getFullYear(), hariIni.getMonth(), hariIni.getDate()))
  const tanggalSelesai = ref(new Date(hariIni.getFullYear(), hariIni.getMonth(), hariIni.getDate()))
  const jenisPelayanan = ref('semua')
  const statusKlaim = ref('3')
  const pencarian = ref('')
  const sedangMemuat = ref(false)
  const sedangEkspor = ref(false)
  const hasil = ref<HasilMonitoringKlaim | null>(null)
  const filterHasil = ref<FilterKlaim | null>(null)
  let generasi = 0
  watch(() => props.token, () => {
    generasi++
    hasil.value = null
    filterHasil.value = null
    sedangMemuat.value = false
    sedangEkspor.value = false
    pesanError.value = ''
  })
  onBeforeUnmount(() => { generasi++ })
  const pesanError = ref('')
  
  const pilihanJenisPelayanan = [
    { label: 'Rawat Inap & Rawat Jalan', value: 'semua' },
    { label: 'Rawat Inap', value: '1' },
    { label: 'Rawat Jalan', value: '2' },
  ]
  
  const pilihanStatusKlaim = [
    { label: 'Proses', value: '1' },
    { label: 'Pending', value: '2' },
    { label: 'Terbayar', value: '3' },
  ]
  
  const daftarKlaim = computed(() => {
    const klaim = hasil.value?.response?.klaim
    return Array.isArray(klaim) ? klaim : []
  })
  
  const daftarTampil = computed(() => {
    const kataKunci = pencarian.value.trim().toLowerCase()
    if (!kataKunci) return daftarKlaim.value
  
    return daftarKlaim.value.filter((item) => [
      item?.noSEP,
      item?.noFPK,
      item?.peserta?.noMR,
      item?.peserta?.noKartu,
      item?.peserta?.nama,
      item?.poli,
      item?.jenisPelayanan,
      item?.Inacbg?.kode,
      item?.Inacbg?.nama,
      item?.status,
      item?.dokter_simrs?.nama_dokter,
      item?.dokter_simrs?.no_rawat,
      item?.dokter_simrs?.keterangan,
    ].some((nilai) => String(nilai || '').toLowerCase().includes(kataKunci)))
  })
  
  const totalPengajuan = computed(() => jumlahBiaya('byPengajuan'))
  const totalDisetujui = computed(() => jumlahBiaya('bySetujui'))
  const totalTarifRS = computed(() => jumlahBiaya('byTarifRS'))
  const periode = computed(() => hasil.value?.periode || null)
  const tanggalTanpaData = computed(() => hasil.value?.tanggal_tanpa_data || [])
  const tanggalGagal = computed(() => hasil.value?.tanggal_gagal || [])
  
  function jumlahBiaya(field) {
    return daftarKlaim.value.reduce((total, item) => total + angka(item?.biaya?.[field]), 0)
  }
  
  function angka(nilai) {
    const hasilAngka = Number(String(nilai ?? '0').replace(/[^0-9.-]/g, ''))
    return Number.isFinite(hasilAngka) ? hasilAngka : 0
  }
  
  function rupiah(nilai) {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      maximumFractionDigits: 0,
    }).format(angka(nilai))
  }
  
  function tanggalApi(nilai) {
    if (!(nilai instanceof Date) || Number.isNaN(nilai.getTime())) return ''
    const tahun = nilai.getFullYear()
    const bulan = String(nilai.getMonth() + 1).padStart(2, '0')
    const tanggal = String(nilai.getDate()).padStart(2, '0')
    return `${tahun}-${bulan}-${tanggal}`
  }
  
  function selisihHari(mulai, selesai) {
    const awal = Date.UTC(mulai.getFullYear(), mulai.getMonth(), mulai.getDate())
    const akhir = Date.UTC(selesai.getFullYear(), selesai.getMonth(), selesai.getDate())
    return Math.floor((akhir - awal) / 86400000) + 1
  }
  
  async function tampilkanData() {
    if (sedangMemuat.value) return
    if (!(tanggalMulai.value instanceof Date) || !(tanggalSelesai.value instanceof Date)) {
      notifikasi.peringatan('Tanggal mulai dan tanggal selesai wajib dipilih.')
      return
    }
  
    const jumlahHari = selisihHari(tanggalMulai.value, tanggalSelesai.value)
    if (jumlahHari < 1) {
      notifikasi.peringatan('Tanggal selesai tidak boleh sebelum tanggal mulai.')
      return
    }
    if (jumlahHari > 31) {
      notifikasi.peringatan('Rentang tanggal maksimal 31 hari.')
      return
    }
  
    sedangMemuat.value = true
    const konteks = generasi
    hasil.value = null
    filterHasil.value = null
    pesanError.value = ''
    try {
      const filter: FilterKlaim = {
        tanggal_mulai: tanggalApi(tanggalMulai.value),
        tanggal_selesai: tanggalApi(tanggalSelesai.value),
        jenis_pelayanan: jenisPelayanan.value,
        status_klaim: statusKlaim.value,
      }
      const data = await monitoringDataKlaim(props.token, filter)
      if (konteks !== generasi) return
      hasil.value = data
      filterHasil.value = filter
  
      if (daftarKlaim.value.length === 0) {
        notifikasi.peringatan('Data klaim tidak ditemukan pada filter yang dipilih.', 'Data Kosong')
      } else if (tanggalGagal.value.length > 0) {
        notifikasi.peringatan(`${daftarKlaim.value.length} klaim ditemukan, tetapi ${tanggalGagal.value.length} tanggal gagal diproses BPJS.`, 'Selesai dengan Peringatan')
      } else if (hasil.value.peringatan_simrs) {
        notifikasi.peringatan(hasil.value.peringatan_simrs)
      } else {
        notifikasi.sukses(`${daftarKlaim.value.length} data klaim berhasil ditampilkan.`)
      }
    } catch (error) {
      if (konteks !== generasi) return
      hasil.value = null
      pesanError.value = error.message || 'Data klaim BPJS gagal dimuat.'
      notifikasi.gagal(pesanError.value)
    } finally {
      if (konteks === generasi) sedangMemuat.value = false
    }
  }
  async function excel() {
    if (sedangMemuat.value || sedangEkspor.value || !hasil.value || !filterHasil.value || !daftarTampil.value.length) return
    const konteks = generasi
    sedangEkspor.value = true
    try {
      await unduhExcelKlaim(daftarTampil.value, hasil.value, filterHasil.value, pencarian.value, () => konteks === generasi)
    } catch (e) {
      if (konteks === generasi) notifikasi.gagal(e instanceof Error ? e.message : 'Ekspor gagal.')
    } finally {
      if (konteks === generasi) sedangEkspor.value = false
    }
  }
  return {
    excel,
    sedangEkspor,
    tanggalMulai,
    tanggalSelesai,
    jenisPelayanan,
    statusKlaim,
    pencarian,
    sedangMemuat,
    hasil,
    pesanError,
    pilihanJenisPelayanan,
    pilihanStatusKlaim,
    daftarKlaim,
    daftarTampil,
    totalPengajuan,
    totalDisetujui,
    totalTarifRS,
    periode,
    tanggalTanpaData,
    tanggalGagal,
    rupiah,
    tampilkanData,
  }
}
