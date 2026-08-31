import { computed, onMounted, ref, watch } from "vue"
import { resepDaftarCopy, resepDetail, resepInfoPasien } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useCopyResep(props, emit) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const loadingDetail = ref(false)
  const locked = ref(false)
  const recipes = ref([])
  const selected = ref(null)
  const detail = ref(null)
  
  const noRkmMedis = computed(() => props.patient?.no_rekam_medis || props.patient?.no_rkm_medis || props.patient?.no_rm || '')
  const jumlahIsi = computed(() => (detail.value?.obat?.length || 0) + (detail.value?.racikan?.length || 0))
  
  function rupiah(value) {
    return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(Number(value || 0))
  }
  function tanggalPendek(value) {
    if (!value) return '-'
    const clean = String(value).slice(0, 10)
    const [y, m, d] = clean.split('-')
    return y && m && d ? `${d}/${m}/${y}` : value
  }
  function jamPendek(value) {
    return value ? String(value).slice(0, 5) : '-'
  }
  function labelStatus(value) {
    const status = String(value || '').toLowerCase()
    if (status === 'ranap') return 'Rawat Inap'
    if (status === 'ralan') return 'Rawat Jalan'
    return value || '-'
  }
  function angka(value, fallback = 0) {
    const n = Number(value)
    return Number.isFinite(n) ? n : fallback
  }
  function normalisasiObat(item) {
    return {
      ...item,
      jml: angka(item.jml ?? item.jumlah ?? item.jumlah_obat, 1),
      harga: angka(item.harga, 0),
      aturan_pakai: item.aturan_pakai || '',
    }
  }
  function normalisasiRacikan(item) {
    return {
      ...item,
      no_racik: '',
      jml_dr: angka(item.jml_dr, 1),
      aturan_pakai: item.aturan_pakai || '',
      keterangan: item.keterangan || '',
      detail: (item.detail || []).map((obat) => ({
        ...obat,
        jml: angka(obat.jml ?? obat.jumlah, 1),
        p1: angka(obat.p1, 1),
        p2: angka(obat.p2, 1),
        kandungan: obat.kandungan || '',
      })),
    }
  }
  
  async function load() {
    if (!props.patient?.no_rawat || !noRkmMedis.value) return
    loading.value = true
    selected.value = null
    detail.value = null
    try {
      const [info, daftar] = await Promise.all([
        resepInfoPasien(props.token, props.patient.no_rawat),
        resepDaftarCopy(props.token, noRkmMedis.value),
      ])
      locked.value = Boolean(info?.billing_terkunci)
      recipes.value = Array.isArray(daftar) ? daftar : []
    } catch (error) {
      notifikasi.gagal(error.message)
    } finally {
      loading.value = false
    }
  }
  
  async function pilihResep(row) {
    selected.value = row
    detail.value = null
    loadingDetail.value = true
    try {
      detail.value = await resepDetail(props.token, row.no_resep)
    } catch (error) {
      notifikasi.gagal(error.message)
    } finally {
      loadingDetail.value = false
    }
  }
  
  function salinResep() {
    if (locked.value) {
      notifikasi.peringatan('Billing sudah diproses. Resep tidak dapat dicopy.')
      return
    }
    if (!detail.value?.header) {
      notifikasi.peringatan('Pilih resep yang ingin disalin dulu.')
      return
    }
    const obat = (detail.value.obat || []).map(normalisasiObat)
    const racikan = (detail.value.racikan || []).map(normalisasiRacikan)
    if (!obat.length && !racikan.length) {
      notifikasi.peringatan('Resep lama tidak memiliki obat atau racikan.')
      return
    }
    emit('copy-resep', {
      sumber: detail.value.header,
      obat,
      racikan,
    })
  }
  
  watch(() => props.patient?.no_rawat, load)
  onMounted(load)
  return {
    loading,
    loadingDetail,
    locked,
    recipes,
    selected,
    detail,
    noRkmMedis,
    jumlahIsi,
    rupiah,
    tanggalPendek,
    jamPendek,
    labelStatus,
    load,
    pilihResep,
    salinResep,
  }
}
