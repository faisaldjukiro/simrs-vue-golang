<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  ArrowLeft,
  CheckCircle2,
  ChevronRight,
  ClipboardList,
  Database,
  Edit3,
  FileSpreadsheet,
  FileText,
  Layers,
  LoaderCircle,
  PanelLeftClose,
  PanelLeftOpen,
  RefreshCw,
  Search,
  Send,
  Settings,
  Users,
} from '@lucide/vue'
import Column from 'primevue/column'
import AutoComplete from 'primevue/autocomplete'
import DataTable from '../../Components/Ui/DataTable.vue'
import DatePicker from '../../Components/Ui/DatePicker.vue'
import Select from '../../Components/Ui/Select.vue'
import { idrgData, idrgDiagnosa, idrgProsedur, prosesIdrg } from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({
  token: { type: String, default: '' },
  user: { type: Object, default: null },
})

const notifikasi = useNotifikasi()
const hariIni = new Date().toLocaleDateString('en-CA')
const loadingDaftar = ref(false)
const loadingDetail = ref(false)
const loadingCodingIdrg = ref('')
const loadingProses = ref('')
const loadingImportInacbg = ref(false)
const loadingNewClaimOtomatis = ref(false)
const pasienTerpilih = ref(null)
const menuSidebarAktif = ref('pengajuan')
const tabAktif = ref('data-pasien')
const sidebarTertutup = ref(localStorage.getItem('idrg_sidebar_tertutup') === 'true')
const diagnosa = ref([])
const prosedur = ref([])
const diagnosaInacbg = ref([])
const prosedurInacbg = ref([])
const hasilProses = ref(null)
const hasilGroupingIdrg = ref(null)
const hasilGroupingInacbg = ref(null)
const dataKlaimEclaim = ref(null)
const waktuGroupingIdrg = ref('')
const waktuGroupingInacbg = ref('')
const daftarPasien = ref([])
const paginasi = ref({ halaman: 1, batas: 50, total: 0, total_halaman: 0 })
const filter = reactive({
  tanggal: hariIni,
  jenis_rawat: 'ranap',
  cari: '',
})
const formKlaim = reactive({
  jaminan: 'JKN',
  nomor_kartu_t: 'kartu_jkn',
  cob_cd: '0',
  jenis_rawat: '1',
  kelas_rawat: '3',
  tgl_masuk: hariIni,
  jam_masuk: '08:00',
  tgl_pulang: hariIni,
  jam_pulang: '09:00',
  cara_masuk: 'gp',
  discharge_status: '1',
  adl_sub_acute: '0',
  adl_chronic: '0',
  birth_weight: '0',
  sistole: '110',
  diastole: '60',
  kantong_darah: '0',
  alteplase_ind: '0',
  icu_indikator: false,
  icu_los: '0',
  ventilator_use_ind: false,
  ventilator_hour: '0',
  ventilator_start_dttm: '',
  ventilator_stop_dttm: '',
  upgrade_class_ind: false,
  upgrade_class_class: 'vip',
  upgrade_class_los: '0',
  upgrade_class_payor: 'peserta',
  add_payment_pct: '0',
  tarif_poli_eks: '0',
  topup_codes: '',
  nama_dokter: '',
  kode_tarif: 'BP',
  pasien_tb: false,
  nomor_register_sitb: '',
  tarif_dikonfirmasi: true,
  tampilkan_apgar: false,
  apgar: {
    menit_1: { appearance: 0, pulse: 0, grimace: 0, activity: 0, respiration: 0 },
    menit_5: { appearance: 0, pulse: 0, grimace: 0, activity: 0, respiration: 0 }
  },
  tampilkan_persalinan: false,
  persalinan: {
    usia_kehamilan: '0', gravida: '0', partus: '0', abortus: '0', onset_kontraksi: 'spontan',
    delivery: [
      { delivery_sequence: '1', delivery_method: 'vaginal', delivery_dttm: '', letak_janin: 'kepala', kondisi: 'livebirth', use_manual: '0', use_forcep: '0', use_vacuum: '0', shk_spesimen_ambil: 'tidak', shk_lokasi: 'tumit', shk_alasan: '', shk_spesimen_dttm: '' }
    ]
  }
})

const menuSidebar = [
  { id: 'pengajuan', label: 'Pengajuan Klaim', icon: ClipboardList, badge: true },
  { id: 'final', label: 'Klaim Final', icon: CheckCircle2 },
  { id: 'kirim', label: 'Pengiriman Klaim', icon: Send },
  { id: 'laporan', label: 'Laporan', icon: FileText },
  { id: 'personel', label: 'Personel (DPJP)', icon: Users },
  { id: 'konfigurasi', label: 'Konfigurasi', icon: Settings },
]

const opsiJenisRawat = [
  { label: 'Rawat Inap', value: 'ranap' },
  { label: 'Rawat Jalan', value: 'ralan' },
]

const opsiJenisRawatKlaim = [
  { label: 'Rawat Inap', value: '1' },
  { label: 'Rawat Jalan', value: '2' },
]

const opsiKelasRawat = [
  { label: 'Kelas 3', value: '3' },
  { label: 'Kelas 2', value: '2' },
  { label: 'Kelas 1', value: '1' },
]

const opsiKelasPelayanan = [
  { label: 'Kelas 3', value: 'kelas_3' },
  { label: 'Kelas 2', value: 'kelas_2' },
  { label: 'Kelas 1', value: 'kelas_1' },
  { label: 'Diatas Kelas 1 / VIP', value: 'vip' },
]

const opsiCaraMasuk = [
  { label: 'Rujukan FKTP', value: 'gp' },
  { label: 'Rujukan RS', value: 'hosp-trans' },
  { label: 'Datang Sendiri', value: 'mp' },
  { label: 'Dari Rawat Jalan', value: 'outp' },
  { label: 'Dari Rawat Inap', value: 'inp' },
  { label: 'Dari IGD', value: 'emd' },
]

const opsiCaraPulang = [
  { label: 'Atas Persetujuan Dokter', value: '1' },
  { label: 'Dirujuk', value: '2' },
  { label: 'Pulang Paksa / APS', value: '3' },
  { label: 'Meninggal', value: '4' },
  { label: 'Lain-lain', value: '5' },
]

const opsiCob = [
  { label: '-', value: '0' },
  { label: 'COB', value: '1' },
]

const opsiKodeTarif = [
  { label: 'Tarif RS Kelas B Pemerintah', value: 'BP' },
  { label: 'Tarif RS Kelas A Pemerintah', value: 'AP' },
  { label: 'Tarif RS Kelas C Pemerintah', value: 'CP' },
  { label: 'Tarif RS Kelas D Pemerintah', value: 'DP' },
]

const definisiTarifRS = [
  ['prosedur_non_bedah', 'Prosedur Non Bedah'],
  ['prosedur_bedah', 'Prosedur Bedah'],
  ['konsultasi', 'Konsultasi'],
  ['tenaga_ahli', 'Tenaga Ahli'],
  ['keperawatan', 'Keperawatan'],
  ['penunjang', 'Penunjang'],
  ['radiologi', 'Radiologi'],
  ['laboratorium', 'Laboratorium'],
  ['pelayanan_darah', 'Pelayanan Darah'],
  ['rehabilitasi', 'Rehabilitasi'],
  ['kamar', 'Kamar / Akomodasi'],
  ['rawat_intensif', 'Rawat Intensif'],
  ['obat', 'Obat'],
  ['obat_kronis', 'Obat Kronis'],
  ['obat_kemoterapi', 'Obat Kemoterapi'],
  ['alkes', 'Alkes'],
  ['bmhp', 'BMHP'],
  ['sewa_alat', 'Sewa Alat'],
]
const formTarifRS = reactive(tarifKosong())

const tanggalFilter = computed({
  get: () => parseTanggal(filter.tanggal),
  set: (tanggal) => {
    filter.tanggal = formatTanggal(tanggal)
  },
})

const daftarTarifRS = computed(() => {
  return definisiTarifRS.map(([key, label]) => ({ key, label, nilai: formTarifRS[key] ?? 0 }))
})

const responseIdrgTerakhir = computed(() => {
  return hasilGroupingIdrg.value
    || hasilProses.value?.hasil?.raw?.response_idrg
    || hasilProses.value?.hasil?.data?.response_idrg
    || hasilProses.value?.hasil?.response?.response_idrg
    || null
})

const responseInacbgTerakhir = computed(() => {
  return hasilGroupingInacbg.value
    || hasilProses.value?.hasil?.raw?.response_inacbg
    || hasilProses.value?.hasil?.data?.response_inacbg
    || hasilProses.value?.hasil?.response?.response_inacbg
    || dataKlaimEclaim.value?.grouper?.response_inacbg
    || null
})

const opsiTopupIdrg = computed(() => {
  const daftar = responseIdrgTerakhir.value?.topup_options
  return Array.isArray(daftar) ? daftar : []
})

const statusIdrgEclaim = computed(() => {
  return String(responseIdrgTerakhir.value?.status_cd || responseIdrgTerakhir.value?.status || '').toLowerCase()
})

const statusKlaimSaatIni = computed(() => normalisasiStatusKlaim(pasienTerpilih.value?.status_klaim))

const setKlaimSudahDilakukan = computed(() => {
  return !['belum', 'new_claim'].includes(statusKlaimSaatIni.value)
})

const statusSesudahFinalIdrg = new Set([
  'final_idrg',
  'import_idrg',
  'grouping_inacbg',
  'final_inacbg',
  'final_claim',
  'send_claim_individual',
])

const statusInacbgEclaim = computed(() => {
  return String(responseInacbgTerakhir.value?.status_cd || responseInacbgTerakhir.value?.status || '').toLowerCase()
})

const klaimStatusEclaim = computed(() => {
  return String(
    dataKlaimEclaim.value?.klaim_status_cd
    || hasilProses.value?.hasil?.response?.data?.klaim_status_cd
    || hasilProses.value?.hasil?.data?.data?.klaim_status_cd
    || '',
  ).toLowerCase()
})

const statusKemenkesEclaim = computed(() => {
  return String(
    dataKlaimEclaim.value?.kemenkes_dc_status_cd
    || dataKlaimEclaim.value?.kemkes_dc_status_cd
    || '',
  ).toLowerCase()
})

const statusBpjsEclaim = computed(() => {
  return String(dataKlaimEclaim.value?.bpjs_dc_status_cd || '').toLowerCase()
})

const idrgSudahFinal = computed(() => {
  return statusIdrgEclaim.value === 'final'
    || statusInacbgEclaim.value === 'final'
    || statusSesudahFinalIdrg.has(statusKlaimSaatIni.value)
})

const bolehBukaIdrg = computed(() => setKlaimSudahDilakukan.value)
const bolehBukaInacbg = computed(() => idrgSudahFinal.value)

const idrgBisaEditUlang = computed(() => {
  return idrgSudahFinal.value
})

const inacbgSudahFinal = computed(() => {
  return statusInacbgEclaim.value === 'final' || klaimStatusEclaim.value === 'final'
})

const klaimSudahTerkirim = computed(() => {
  return statusKlaimSaatIni.value === 'send_claim_individual'
    || ['sent', 'terkirim'].includes(statusKemenkesEclaim.value)
    || ['sent', 'terkirim'].includes(statusBpjsEclaim.value)
})

const klaimSudahFinal = computed(() => {
  return klaimSudahTerkirim.value
    || statusKlaimSaatIni.value === 'final_claim'
    || klaimStatusEclaim.value === 'final'
})

const inacbgBisaEditUlang = computed(() => {
  return inacbgSudahFinal.value && !klaimSudahFinal.value
})

const codingInacbgTidakValid = computed(() => {
  return [...diagnosaInacbg.value, ...prosedurInacbg.value].filter((item) => String(item?.validcode ?? '1') === '0')
})

const idrgGroupingFinal = computed(() => {
  return idrgBisaEditUlang.value || String(hasilGroupingIdrgTabel.value.status || '').toLowerCase() === 'final'
})

const inacbgGroupingFinal = computed(() => {
  return inacbgBisaEditUlang.value || String(hasilGroupingInacbgTabel.value.status || '').toLowerCase() === 'final'
})

const hasilGroupingIdrgTabel = computed(() => {
  const data = responseIdrgTerakhir.value || {}
  const scriptVersion = ambilNilai(data, ['script_version'], '')
  const logicVersion = ambilNilai(data, ['logic_version'], '')
  const version = [
    scriptVersion ? `Script V.${scriptVersion}` : '',
    logicVersion ? `Logic V.${logicVersion}` : '',
  ].filter(Boolean).join(' / ')
  const petugas = props.user?.name || props.user?.username || 'Petugas'
  const waktu = waktuGroupingIdrg.value || waktuIndoLengkap(new Date())
  const jenisRawat = `${formKlaim.jenis_rawat === '1' ? 'Rawat Inap' : 'Rawat Jalan'}${formKlaim.jenis_rawat === '1' ? ` (${losHari()})` : ''}`

  return {
    info: `${petugas} @ ${waktu}${version ? ` - ${version}` : ''}`,
    jenisRawat,
    mdcNomor: ambilNilai(data, ['mdc_number', 'mdc_no', 'mdc']),
    mdcDeskripsi: ambilNilai(data, ['mdc_description', 'mdc_desc', 'mdc_name']),
    drgKode: ambilNilai(data, ['drg_code', 'drg']),
    drgDeskripsi: ambilNilai(data, ['drg_description', 'drg_desc', 'drg_name']),
    drgCostWeight: formatBobotGrouping(ambilNilai(data, ['cost_weight', 'drg_cost_weight'])),
    kris: ambilNilai(data, ['kris', 'kris_code', 'kelas_rs'], dataKlaimEclaim.value?.kelas_rs || 'A'),
    krisCostWeight: formatBobotGrouping(ambilNilai(data, ['kris_cost_weight', 'kris_weight'])),
    nbr: ambilNilai(data, ['nbr', 'national_base_rate']),
    totalCostWeight: formatBobotGrouping(ambilNilai(data, ['total_cost_weight', 'total_weight'])),
    totalKlaim: ambilRupiahGrouping(data, ['total_tarif', 'total_claim', 'total_klaim', 'claim_amount', 'tariff']),
    status: ambilNilai(data, ['status_cd', 'status', 'status_klaim'], 'normal'),
  }
})

const hasilGroupingInacbgTabel = computed(() => {
  const data = responseInacbgTerakhir.value || {}
  const cbg = data.cbg || {}
  const specialRows = daftarSpecialCmgTampil(data)
  const subAcute = angkaDariNilai(data.sub_acute_tariff || data.sub_acute || 0)
  const chronic = angkaDariNilai(data.chronic_tariff || data.chronic || 0)
  const tarifUtama = angkaDariNilai(data.tariff || data.base_tariff || 0)
  const totalSpecial = specialRows.reduce((total, item) => total + angkaDariNilai(item.tariff), 0)
  const total = angkaDariNilai(data.total_tarif || data.total_claim || data.total_klaim || 0) || (tarifUtama + totalSpecial + subAcute + chronic)
  const petugas = props.user?.name || props.user?.username || 'Petugas'
  const waktu = waktuGroupingInacbg.value || waktuIndoLengkap(new Date())
  const kelas = String(data.kelas || dataKlaimEclaim.value?.kelas || '').replace('_', ' ').toUpperCase()

  return {
    info: `${petugas} @ ${waktu}${data.inacbg_version ? ` - Versi Grouper: ${data.inacbg_version}` : ''}${kelas ? ` • ${kelas}` : ''}`,
    kode: cbg.code || data.cbg_code || '-',
    deskripsi: cbg.description || data.cbg_description || '-',
    tarifUtama,
    subAcute,
    chronic,
    specialRows,
    total,
    status: data.status_cd || data.status || 'normal',
  }
})

onMounted(muatDaftar)
watch(() => props.token, () => {
  pasienTerpilih.value = null
  muatDaftar()
})
watch(() => formKlaim.jenis_rawat, (jenisRawat) => {
  if (jenisRawat === '2') {
    formKlaim.upgrade_class_ind = false
    formKlaim.icu_indikator = false
    formKlaim.ventilator_use_ind = false
    return
  }
  formKlaim.tarif_poli_eks = '0'
})
watch(() => formKlaim.upgrade_class_ind, (aktif) => {
  if (aktif) {
    if (!formKlaim.upgrade_class_class) formKlaim.upgrade_class_class = 'kelas_1'
    return
  }
  formKlaim.upgrade_class_class = ''
  formKlaim.upgrade_class_los = '0'
  formKlaim.upgrade_class_payor = 'peserta'
  formKlaim.add_payment_pct = '0'
})
watch(() => formKlaim.icu_indikator, (aktif) => {
  if (aktif) return
  formKlaim.icu_los = '0'
  formKlaim.ventilator_use_ind = false
})
watch(() => formKlaim.ventilator_use_ind, (aktif) => {
  if (aktif) return
  formKlaim.ventilator_hour = '0'
  formKlaim.ventilator_start_dttm = ''
  formKlaim.ventilator_stop_dttm = ''
})

async function muatDaftar() {
  if (!props.token || loadingDaftar.value) return

  loadingDaftar.value = true
  try {
    const data = await idrgData(props.token, {
      tanggal: filter.tanggal,
      jenis_rawat: filter.jenis_rawat,
      cari: filter.cari,
      page: paginasi.value.halaman,
      limit: paginasi.value.batas,
    })
    daftarPasien.value = data.data ?? []
    paginasi.value = data.paginasi ?? paginasi.value
    if (daftarPasien.value.length === 0) {
      notifikasi.peringatan('Data pasien IDRG tidak ditemukan untuk filter ini.', 'Data Kosong')
    }
  } catch (error) {
    daftarPasien.value = []
    notifikasi.gagal(error.message || 'Data IDRG tidak dapat dimuat.')
  } finally {
    loadingDaftar.value = false
  }
}

function terapkanFilter() {
  paginasi.value = { ...paginasi.value, halaman: 1 }
  pasienTerpilih.value = null
  muatDaftar()
}

function gantiHalaman(event) {
  paginasi.value = {
    ...paginasi.value,
    halaman: Math.floor(event.first / event.rows) + 1,
    batas: event.rows,
  }
  muatDaftar()
}

function toggleSidebar() {
  sidebarTertutup.value = !sidebarTertutup.value
  localStorage.setItem('idrg_sidebar_tertutup', String(sidebarTertutup.value))
}

async function pilihPasien(pasien) {
  const perluKirimNewClaim = Boolean(pasien?.no_rawat)

  pasienTerpilih.value = pasien
  isiFormKlaim(pasien)
  menuSidebarAktif.value = 'pengajuan'
  tabAktif.value = 'data-pasien'
  diagnosa.value = []
  prosedur.value = []
  diagnosaInacbg.value = []
  prosedurInacbg.value = []
  hasilProses.value = null
  hasilGroupingIdrg.value = null
  hasilGroupingInacbg.value = null
  dataKlaimEclaim.value = null
  waktuGroupingIdrg.value = ''
  waktuGroupingInacbg.value = ''
  if (perluKirimNewClaim) {
    await Promise.all([
      kirimNewClaimOtomatis(),
      muatCodingPasien(),
    ])
  } else {
    await muatCodingPasien()
  }
  if (pasien?.no_sep) {
    await muatDataKlaimEclaim({ diam: true, bukaTabJikaAda: true })
  }
}

function pilihMenuSidebar(menuId) {
  menuSidebarAktif.value = menuId
  if (menuId !== 'pengajuan') {
    pasienTerpilih.value = null
    tabAktif.value = 'data-pasien'
    hasilProses.value = null
  }
}

async function muatCodingPasien() {
  if (!props.token || !pasienTerpilih.value) return

  loadingDetail.value = true
  try {
    const [hasilDiagnosa, hasilProsedur] = await Promise.all([
      idrgDiagnosa(props.token, pasienTerpilih.value.no_rawat),
      idrgProsedur(props.token, pasienTerpilih.value.no_rawat),
    ])
    diagnosa.value = hasilDiagnosa?.data?.diagnosa ?? []
    prosedur.value = hasilProsedur?.data?.prosedur ?? []
  } catch (error) {
    notifikasi.gagal(error.message || 'Coding pasien tidak dapat dimuat.')
  } finally {
    loadingDetail.value = false
  }
}

function hapusDiagnosa(index) {
  if (idrgSudahFinal.value) return
  diagnosa.value.splice(index, 1)
  if (diagnosa.value.length > 0) diagnosa.value[0].status = 'Primary'
  activeSubstitusiDiagnosa.value = null
}
function hapusProsedur(index) {
  if (idrgSudahFinal.value) return
  prosedur.value.splice(index, 1)
  activeSubstitusiProsedur.value = null
}

const searchHasilDiagnosa = ref([])
const activeSubstitusiDiagnosa = ref(null)
const substitusiKeyword = ref('')
const debounceTimeout = ref(null)

const searchHasilProsedur = ref([])
const activeSubstitusiProsedur = ref(null)
const searchHasilDiagnosaInacbg = ref([])
const searchHasilProsedurInacbg = ref([])
const substitusiDiagnosaInacbg = ref(null)
const substitusiProsedurInacbg = ref(null)
const activeSubstitusiDiagnosaInacbg = ref(null)
const activeSubstitusiProsedurInacbg = ref(null)

async function muatCodingIdrg(jenis) {
  if (!props.token || !pasienTerpilih.value || loadingCodingIdrg.value || idrgSudahFinal.value) return

  loadingCodingIdrg.value = jenis
  try {
    if (jenis === 'diagnosa') {
      const hasil = await idrgDiagnosa(props.token, pasienTerpilih.value.no_rawat)
      diagnosa.value = hasil?.data?.diagnosa ?? []
      return
    }
    const hasil = await idrgProsedur(props.token, pasienTerpilih.value.no_rawat)
    prosedur.value = hasil?.data?.prosedur ?? []
  } catch (error) {
    notifikasi.gagal(error.message || `Coding ${jenis} pasien tidak dapat dimuat.`)
  } finally {
    loadingCodingIdrg.value = ''
  }
}

async function onSearchProsedur(event) {
  if (debounceTimeout.value) clearTimeout(debounceTimeout.value)
  debounceTimeout.value = setTimeout(async () => {
    try {
      const res = await prosesIdrg(props.token, { 
        aksi: 'search_procedures_idrg', 
        no_rawat: pasienTerpilih.value.no_rawat, 
        no_sep: pasienTerpilih.value.no_sep, 
        pasien: pasienTerpilih.value,
        klaim: { keyword: event.query }
      })
      searchHasilProsedur.value = hasilPencarianEclaim(res)
    } catch (e) {
      searchHasilProsedur.value = []
    }
  }, 400)
}

function applySubstitusiProsedur(index, event) {
  if (idrgSudahFinal.value) return
  const pilihan = event?.value
  if (!pilihan?.kode || !prosedur.value[index]) return
  prosedur.value[index].kd_prosedur = pilihan.kode
  prosedur.value[index].nm_prosedur = pilihan.deskripsi
  activeSubstitusiProsedur.value = null
}

async function onSearchDiagnosa(event) {
  if (debounceTimeout.value) clearTimeout(debounceTimeout.value)
  debounceTimeout.value = setTimeout(async () => {
    try {
      const res = await prosesIdrg(props.token, { 
        aksi: 'search_diagnosis_idrg', 
        no_rawat: pasienTerpilih.value.no_rawat, 
        no_sep: pasienTerpilih.value.no_sep, 
        pasien: pasienTerpilih.value,
        klaim: { keyword: event.query }
      })
      searchHasilDiagnosa.value = hasilPencarianEclaim(res)
    } catch (e) {
      searchHasilDiagnosa.value = []
    }
  }, 400)
}

function applySubstitusi(index, event) {
  if (idrgSudahFinal.value) return
  const pilihan = event?.value
  if (!pilihan?.kode || !diagnosa.value[index]) return
  const kodePilihan = String(pilihan.kode).trim().toUpperCase()
  const sudahAda = diagnosa.value.some((item, posisi) => {
    return posisi !== index && String(item.kd_diag || '').trim().toUpperCase() === kodePilihan
  })
  if (sudahAda) {
    substitusiKeyword.value = ''
    notifikasi.peringatan(`Diagnosis ${pilihan.kode} sudah ada pada daftar pasien.`, 'Duplikat Diagnosa')
    return
  }
  diagnosa.value[index].kd_diag = pilihan.kode
  diagnosa.value[index].nm_diag = pilihan.deskripsi
  activeSubstitusiDiagnosa.value = null
}

function setPrimaryDiagnosa(index) {
  if (idrgSudahFinal.value) return
  if (index === 0) return
  const item = diagnosa.value.splice(index, 1)[0]
  diagnosa.value.unshift(item)
  diagnosa.value.forEach((d, i) => d.status = i === 0 ? 'Primary' : 'Secondary')
  activeSubstitusiDiagnosa.value = null
}

function aksiInacbg(aksi) {
  return ['idrg_ke_inacbg', 'inacbg_diagnosa_get', 'inacbg_procedure_get', 'grouping_inacbg', 'grouping_inacbg_stage_2', 'final_inacbg', 'reedit_inacbg', 'final_klaim', 'reedit_klaim', 'send_claim_individual', 'kirim_klaim'].includes(aksi)
}

function bukaEditorDiagnosa(index) {
  if (idrgSudahFinal.value) return
  activeSubstitusiDiagnosa.value = activeSubstitusiDiagnosa.value === index ? null : index
  activeSubstitusiProsedur.value = null
  substitusiKeyword.value = ''
}

function bukaEditorProsedur(index) {
  if (idrgSudahFinal.value) return
  activeSubstitusiProsedur.value = activeSubstitusiProsedur.value === index ? null : index
  activeSubstitusiDiagnosa.value = null
  substitusiKeyword.value = ''
}

function pilihTabDetail(tab) {
  if (tab === 'idrg' && !bolehBukaIdrg.value) return
  if (tab === 'inacbg' && !bolehBukaInacbg.value) return
  tabAktif.value = tab
}

function dataResponsEclaim(hasil) {
  return hasil?.hasil?.raw?.data
    || hasil?.hasil?.data
    || hasil?.hasil?.response?.data
    || null
}

function normalisasiCodingInacbg(item, jenis) {
  return {
    ...item,
    code: String(item?.code || (jenis === 'diagnosa' ? item?.kd_diag || item?.kd_penyakit : item?.kd_prosedur || item?.kode) || '').trim(),
    display: String(item?.display || (jenis === 'diagnosa' ? item?.nm_diag || item?.nm_penyakit : item?.nm_prosedur || item?.deskripsi) || '').trim(),
    validcode: String(item?.validcode ?? '1'),
    multiplicity: Number(item?.multiplicity || item?.jumlah || 1),
  }
}

function terapkanHasilImportInacbg(hasil) {
  const data = dataResponsEclaim(hasil)
  const daftarDiagnosa = data?.diagnosa?.expanded
  const daftarProsedur = data?.procedure?.expanded
  if (!Array.isArray(daftarDiagnosa) && !Array.isArray(daftarProsedur)) return false

  diagnosaInacbg.value = (daftarDiagnosa || []).map((item) => normalisasiCodingInacbg(item, 'diagnosa'))
  prosedurInacbg.value = (daftarProsedur || []).map((item) => normalisasiCodingInacbg(item, 'prosedur'))
  return true
}

function terapkanDaftarCodingInacbg(hasil, jenis) {
  const daftar = dataResponsEclaim(hasil)?.expanded
  if (!Array.isArray(daftar)) return false
  if (jenis === 'diagnosa') diagnosaInacbg.value = daftar.map((item) => normalisasiCodingInacbg(item, jenis))
  else prosedurInacbg.value = daftar.map((item) => normalisasiCodingInacbg(item, jenis))
  return true
}

async function muatCodingInacbg(jenis) {
  if (!pasienTerpilih.value?.no_sep || loadingProses.value || inacbgSudahFinal.value) return
  const aksi = jenis === 'diagnosa' ? 'inacbg_diagnosa_get' : 'inacbg_procedure_get'
  loadingProses.value = aksi
  try {
    const hasil = await prosesIdrg(props.token, {
      aksi,
      no_rawat: pasienTerpilih.value.no_rawat,
      no_sep: pasienTerpilih.value.no_sep,
      pasien: pasienTerpilih.value,
      klaim: payloadKlaim(aksi),
    })
    if (!terapkanDaftarCodingInacbg(hasil, jenis)) {
      notifikasi.peringatan(`Daftar ${jenis} INA-CBG belum tersedia. Klik Import IDRG terlebih dahulu.`)
    }
  } catch (error) {
    notifikasi.gagal(error.message || `Coding ${jenis} INA-CBG tidak dapat dibaca.`)
  } finally {
    loadingProses.value = ''
  }
}

function kodeCodingTidakValid() {
  return codingInacbgTidakValid.value.map((item) => item.code).filter(Boolean).join(', ')
}

async function importIdrgKeInacbg() {
  if (!pasienTerpilih.value?.no_sep || loadingImportInacbg.value || inacbgSudahFinal.value) return

  loadingImportInacbg.value = true
  try {
    const hasil = await prosesIdrg(props.token, {
      aksi: 'idrg_ke_inacbg',
      no_rawat: pasienTerpilih.value.no_rawat,
      no_sep: pasienTerpilih.value.no_sep,
      pasien: pasienTerpilih.value,
      klaim: payloadKlaim('idrg_ke_inacbg'),
    })
    hasilProses.value = hasil
    if (!prosesBerhasil(hasil)) {
      tampilkanNotifikasiProses(hasil, 'Import IDRG ke INA-CBG gagal.')
      return
    }

    pasienTerpilih.value.status_klaim = 'import_idrg'
    if (!terapkanHasilImportInacbg(hasil)) {
      notifikasi.peringatan('Import berhasil, tetapi daftar coding belum ditemukan dalam respons E-Klaim.')
      return
    }

    notifikasi.sukses('Coding IDRG berhasil diimpor ke INA-CBG.')
    if (codingInacbgTidakValid.value.length > 0) {
      notifikasi.peringatan(`Kode ${kodeCodingTidakValid()} tidak berlaku di INA-CBG. Silakan lakukan substitusi atau hapus sebelum Grouping.`, 'IM Tidak Berlaku')
    }
  } catch (error) {
    if (errorKlaimSudahFinal(error)) {
      notifikasi.peringatan('Coding INA-CBG sudah final dan perlu dibuka kembali sebelum import.', 'INA-CBG Sudah Final')
      return
    }
    notifikasi.gagal(error.message || 'Import IDRG ke INA-CBG gagal.')
  } finally {
    loadingImportInacbg.value = false
  }
}

async function tandaiCodingSudahFinal(aksi) {
  if (aksiInacbg(aksi)) {
    pasienTerpilih.value.status_klaim = 'final_inacbg'
    notifikasi.peringatan('INA-CBG sudah final. Tombol saya ubah ke Edit INA-CBG.', 'Sudah Final')
  } else {
    pasienTerpilih.value.status_klaim = 'final_idrg'
    notifikasi.peringatan('iDRG sudah final. Tombol saya ubah ke Edit Ulang IDRG.', 'Sudah Final')
  }
  // Status final sudah cukup untuk membuka tombol re-edit. Detail dibaca di
  // belakang agar tombol tidak ikut terkunci menunggu respons E-Klaim.
  muatDataKlaimEclaim({ diam: true })
}

function hasilPencarianEclaim(res) {
  const daftar = res?.hasil?.data?.data
    || res?.hasil?.data
    || res?.hasil?.response?.data
    || res?.hasil?.raw?.data
    || res?.hasil?.raw?.response?.data
    || []
  if (!Array.isArray(daftar)) return []

  return daftar.map((item) => {
    const deskripsi = Array.isArray(item) ? String(item[0] || '') : String(item?.display || item?.description || '')
    const kode = Array.isArray(item) ? String(item[1] || '') : String(item?.code || item?.kode || '')
    return { kode, deskripsi, label: `${kode} - ${deskripsi}` }
  }).filter((item) => item.kode)
}

async function onSearchDiagnosaInacbg(event) {
  if (debounceTimeout.value) clearTimeout(debounceTimeout.value)
  debounceTimeout.value = setTimeout(async () => {
    try {
      const res = await prosesIdrg(props.token, {
        aksi: 'search_diagnosis',
        no_rawat: pasienTerpilih.value.no_rawat,
        no_sep: pasienTerpilih.value.no_sep,
        pasien: pasienTerpilih.value,
        klaim: { keyword: event.query },
      })
      searchHasilDiagnosaInacbg.value = hasilPencarianEclaim(res)
    } catch {
      searchHasilDiagnosaInacbg.value = []
    }
  }, 350)
}

async function onSearchProsedurInacbg(event) {
  if (debounceTimeout.value) clearTimeout(debounceTimeout.value)
  debounceTimeout.value = setTimeout(async () => {
    try {
      const res = await prosesIdrg(props.token, {
        aksi: 'search_procedures',
        no_rawat: pasienTerpilih.value.no_rawat,
        no_sep: pasienTerpilih.value.no_sep,
        pasien: pasienTerpilih.value,
        klaim: { keyword: event.query },
      })
      searchHasilProsedurInacbg.value = hasilPencarianEclaim(res)
    } catch {
      searchHasilProsedurInacbg.value = []
    }
  }, 350)
}

function terapkanSubstitusiDiagnosaInacbg(index, event) {
  if (inacbgSudahFinal.value) return
  const pilihan = event?.value
  if (!pilihan?.kode || !diagnosaInacbg.value[index]) return
  const kodePilihan = String(pilihan.kode).trim().toUpperCase()
  const sudahAda = diagnosaInacbg.value.some((item, posisi) => {
    return posisi !== index && String(item.code || '').trim().toUpperCase() === kodePilihan
  })
  if (sudahAda) {
    substitusiDiagnosaInacbg.value = null
    notifikasi.peringatan(`Diagnosis ${pilihan.kode} sudah ada pada daftar pasien.`, 'Duplikat Diagnosa')
    return
  }
  diagnosaInacbg.value[index] = {
    ...diagnosaInacbg.value[index],
    code: pilihan.kode,
    display: pilihan.deskripsi,
    validcode: '1',
    metadata: { code: 200, message: 'Ok' },
  }
  substitusiDiagnosaInacbg.value = null
  activeSubstitusiDiagnosaInacbg.value = null
}

function terapkanSubstitusiProsedurInacbg(index, event) {
  if (inacbgSudahFinal.value) return
  const pilihan = event?.value
  if (!pilihan?.kode || !prosedurInacbg.value[index]) return
  prosedurInacbg.value[index] = {
    ...prosedurInacbg.value[index],
    code: pilihan.kode,
    display: pilihan.deskripsi,
    validcode: '1',
    metadata: { code: 200, message: 'Ok' },
  }
  substitusiProsedurInacbg.value = null
  activeSubstitusiProsedurInacbg.value = null
}

function hapusDiagnosaInacbg(index) {
  if (inacbgSudahFinal.value) return
  diagnosaInacbg.value.splice(index, 1)
  activeSubstitusiDiagnosaInacbg.value = null
}

function hapusProsedurInacbg(index) {
  if (inacbgSudahFinal.value) return
  prosedurInacbg.value.splice(index, 1)
  activeSubstitusiProsedurInacbg.value = null
}

function setPrimaryDiagnosaInacbg(index) {
  if (inacbgSudahFinal.value || index === 0) return
  const item = diagnosaInacbg.value.splice(index, 1)[0]
  diagnosaInacbg.value.unshift(item)
  activeSubstitusiDiagnosaInacbg.value = null
}

function bukaEditorDiagnosaInacbg(index) {
  if (inacbgSudahFinal.value) return
  activeSubstitusiDiagnosaInacbg.value = activeSubstitusiDiagnosaInacbg.value === index ? null : index
  activeSubstitusiProsedurInacbg.value = null
  substitusiDiagnosaInacbg.value = null
}

function bukaEditorProsedurInacbg(index) {
  if (inacbgSudahFinal.value) return
  activeSubstitusiProsedurInacbg.value = activeSubstitusiProsedurInacbg.value === index ? null : index
  activeSubstitusiDiagnosaInacbg.value = null
  substitusiProsedurInacbg.value = null
}

async function jalankanProses(aksi) {
  if (!pasienTerpilih.value || loadingProses.value) {
    return
  }

  if (idrgSudahFinal.value && ['idrg_diagnosa_set', 'idrg_procedure_set', 'grouping_idrg', 'grouping_idrg_stage_1', 'grouping_idrg_stage_2', 'final_idrg'].includes(aksi)) {
    return
  }

  if (inacbgSudahFinal.value && ['idrg_ke_inacbg', 'inacbg_diagnosa_get', 'inacbg_procedure_get', 'grouping_inacbg', 'grouping_inacbg_stage_2', 'final_inacbg'].includes(aksi)) {
    return
  }

  if (aksi === 'final_klaim' && (!inacbgSudahFinal.value || klaimSudahFinal.value)) {
    notifikasi.peringatan('Final INA-CBG harus berhasil dan klaim belum boleh berstatus final.', 'Final Klaim Belum Tersedia')
    return
  }

  if (aksi === 'reedit_klaim' && !klaimSudahFinal.value) {
    notifikasi.peringatan('Klaim belum berstatus final.', 'Edit Ulang Belum Tersedia')
    return
  }

  if (['send_claim_individual', 'kirim_klaim', 'cetak_klaim'].includes(aksi) && !klaimSudahFinal.value) {
    notifikasi.peringatan('Final Klaim harus berhasil terlebih dahulu.', 'Proses Belum Tersedia')
    return
  }

  if (aksi === 'grouping_inacbg') {
    if (codingInacbgTidakValid.value.length > 0) {
      notifikasi.gagal(`Kode ${kodeCodingTidakValid()} bertanda IM tidak berlaku untuk INA-CBG. Ganti melalui Substitusi atau hapus dahulu.`, 'Coding Belum Valid')
      return
    }
    if (diagnosaInacbg.value.length === 0) {
      notifikasi.peringatan('Import IDRG ke INA-CBG terlebih dahulu sebelum grouping.', 'Coding INA-CBG Kosong')
      return
    }
  }

  loadingProses.value = aksi
  try {
    const hasil = await prosesIdrg(props.token, {
      aksi,
      no_rawat: pasienTerpilih.value.no_rawat,
      no_sep: pasienTerpilih.value.no_sep,
      pasien: pasienTerpilih.value,
      klaim: payloadKlaim(aksi),
    })
    hasilProses.value = hasil
    const sukses = prosesBerhasil(hasil)
    terapkanHasilGroupingIdrg(hasil)
    terapkanHasilGroupingInacbg(hasil)
    if (!sukses && hasilKlaimSudahFinal(hasil)) {
      await tandaiCodingSudahFinal(aksi)
      return
    }
    if (sukses && aksi === 'buat_klaim_baru') pasienTerpilih.value.status_klaim = 'new_claim'
    if (sukses && aksi === 'atur_data_klaim') {
      pasienTerpilih.value.status_klaim = 'set_claim_data'
      tabAktif.value = 'idrg'
      await muatCodingPasien()
    }
    if (sukses && aksi === 'final_idrg') {
      pasienTerpilih.value.status_klaim = 'final_idrg'
      if (hasilGroupingIdrg.value) hasilGroupingIdrg.value = { ...hasilGroupingIdrg.value, status_cd: 'final' }
    }
    if (sukses && aksi === 'reedit_idrg') {
      // Respons re-edit berarti kunci E-Klaim sudah terbuka. Lepaskan indikator
      // proses sebelum UI berpindah ke mode edit agar tombol Grouping langsung aktif.
      loadingProses.value = ''
      pasienTerpilih.value.status_klaim = 'set_claim_data'
      if (responseIdrgTerakhir.value) {
        hasilGroupingIdrg.value = { ...responseIdrgTerakhir.value, status_cd: 'normal', status: 'normal' }
      }
    }
    if (sukses && aksi === 'idrg_ke_inacbg') {
      pasienTerpilih.value.status_klaim = 'import_idrg'
      terapkanHasilImportInacbg(hasil)
    }
    if (sukses && ['grouping_inacbg', 'grouping_inacbg_stage_2'].includes(aksi)) pasienTerpilih.value.status_klaim = 'grouping_inacbg'
    if (sukses && aksi === 'final_inacbg') {
      pasienTerpilih.value.status_klaim = 'final_inacbg'
      if (hasilGroupingInacbg.value) hasilGroupingInacbg.value = { ...hasilGroupingInacbg.value, status_cd: 'final' }
    }
    if (sukses && aksi === 'reedit_inacbg') {
      loadingProses.value = ''
      pasienTerpilih.value.status_klaim = 'grouping_inacbg'
      if (responseInacbgTerakhir.value) {
        hasilGroupingInacbg.value = { ...responseInacbgTerakhir.value, status_cd: 'normal', status: 'normal' }
      }
    }
    if (sukses && aksi === 'final_klaim') {
      pasienTerpilih.value.status_klaim = 'final_claim'
      dataKlaimEclaim.value = {
        ...(dataKlaimEclaim.value || {}),
        klaim_status_cd: 'final',
      }
    }
    if (sukses && aksi === 'reedit_klaim') {
      pasienTerpilih.value.status_klaim = 'final_inacbg'
      dataKlaimEclaim.value = {
        ...(dataKlaimEclaim.value || {}),
        klaim_status_cd: 'normal',
        kemenkes_dc_status_cd: 'unsent',
        kemkes_dc_status_cd: 'unsent',
        bpjs_dc_status_cd: 'unsent',
      }
    }
    if (sukses && ['send_claim_individual', 'kirim_klaim'].includes(aksi)) {
      pasienTerpilih.value.status_klaim = 'send_claim_individual'
      dataKlaimEclaim.value = {
        ...(dataKlaimEclaim.value || {}),
        klaim_status_cd: 'final',
        kemenkes_dc_status_cd: 'sent',
        kemkes_dc_status_cd: 'sent',
      }
    }
    if (sukses && aksi === 'cetak_klaim') {
      const base64Data = hasil?.hasil?.data || hasil?.hasil?.response?.data || hasil?.hasil?.response
      if (!unduhPdfKlaim(base64Data)) {
        notifikasi.gagal('Servis E-Klaim tidak mengembalikan berkas PDF yang valid.', 'Cetak Klaim Gagal')
        return
      }
    }
    tampilkanNotifikasiProses(hasil, pesanSuksesProses(aksi))
    if (sukses && aksi === 'idrg_ke_inacbg' && codingInacbgTidakValid.value.length > 0) {
      notifikasi.peringatan(`Kode ${kodeCodingTidakValid()} tidak berlaku di INA-CBG. Silakan lakukan substitusi atau hapus sebelum Grouping.`, 'IM Tidak Berlaku')
    }
  } catch (error) {
    if (errorKlaimSudahFinal(error)) {
      await tandaiCodingSudahFinal(aksi)
      return
    }
    notifikasi.gagal(error.message || 'Proses E-Klaim belum aktif di backend Go.')
  } finally {
    loadingProses.value = ''
  }
}

async function muatDataKlaimEclaim(opsi = {}) {
  if (!props.token || !pasienTerpilih.value?.no_rawat) return false

  try {
    const hasil = await prosesIdrg(props.token, {
      aksi: 'get_claim_data',
      no_rawat: pasienTerpilih.value.no_rawat,
      no_sep: pasienTerpilih.value.no_sep,
      pasien: pasienTerpilih.value,
      klaim: payloadKlaim(),
    })
    hasilProses.value = hasil
    terapkanStatusDariDataKlaim(hasil)
    terapkanHasilGroupingIdrg(hasil)
    terapkanHasilGroupingInacbg(hasil)
    if (opsi.bukaTabJikaAda) {
      if (bolehBukaInacbg.value) tabAktif.value = 'inacbg'
      else if (bolehBukaIdrg.value) tabAktif.value = 'idrg'
      else tabAktif.value = 'data-pasien'
    }
    return true
  } catch (error) {
    if (opsi.diam) return false
    notifikasi.peringatan('Status final terbaca, tapi detail grouping belum bisa dibaca dari E-Klaim.', 'Detail Belum Tampil')
    return false
  }
}

async function kirimNewClaimOtomatis() {
  if (!props.token || !pasienTerpilih.value || loadingNewClaimOtomatis.value) return

  loadingNewClaimOtomatis.value = true
  try {
    const hasil = await prosesIdrg(props.token, {
      aksi: 'buat_klaim_baru',
      no_rawat: pasienTerpilih.value.no_rawat,
      no_sep: pasienTerpilih.value.no_sep,
      pasien: pasienTerpilih.value,
    })
    hasilProses.value = hasil
    if (prosesBerhasil(hasil) && !idrgSudahFinal.value) pasienTerpilih.value.status_klaim = 'new_claim'
    tampilkanNotifikasiProses(hasil, 'New Claim otomatis berhasil dikirim.')
  } catch (error) {
    notifikasi.gagal(error.message || 'New Claim otomatis belum berhasil dikirim.')
  } finally {
    loadingNewClaimOtomatis.value = false
  }
}

function parseTanggal(nilai) {
  if (!nilai) return null
  const tanggal = new Date(`${nilai}T00:00:00`)
  return Number.isNaN(tanggal.getTime()) ? null : tanggal
}

function formatTanggal(nilai) {
  if (!nilai) return ''
  const tanggal = nilai instanceof Date ? nilai : new Date(nilai)
  if (Number.isNaN(tanggal.getTime())) return ''
  const tahun = tanggal.getFullYear()
  const bulan = String(tanggal.getMonth() + 1).padStart(2, '0')
  const hari = String(tanggal.getDate()).padStart(2, '0')
  return `${tahun}-${bulan}-${hari}`
}

function formatTanggalWaktu(nilai) {
  if (!nilai) return ''
  const t = nilai instanceof Date ? nilai : new Date(nilai)
  if (Number.isNaN(t.getTime())) return ''
  const pad = (n) => String(n).padStart(2, '0')
  return `${t.getFullYear()}-${pad(t.getMonth() + 1)}-${pad(t.getDate())} ${pad(t.getHours())}:${pad(t.getMinutes())}:${pad(t.getSeconds())}`
}

function parseWaktu(nilai) {
  if (!nilai) return null
  const [j, m] = String(nilai).split(':')
  const t = new Date()
  t.setHours(Number(j || 0), Number(m || 0), 0, 0)
  return t
}

function formatWaktu(nilai) {
  if (!nilai) return ''
  const t = nilai instanceof Date ? nilai : new Date(nilai)
  if (Number.isNaN(t.getTime())) return ''
  const pad = (n) => String(n).padStart(2, '0')
  return `${pad(t.getHours())}:${pad(t.getMinutes())}`
}

function unduhPdfKlaim(dataPdf) {
  if (typeof dataPdf !== 'string') return false

  const base64 = dataPdf
    .replace(/^data:application\/pdf;base64,/i, '')
    .replace(/\s/g, '')

  if (base64.length < 100) return false

  try {
    const biner = window.atob(base64)
    const potongan = []
    const ukuranPotongan = 8192

    for (let awal = 0; awal < biner.length; awal += ukuranPotongan) {
      const bagian = biner.slice(awal, awal + ukuranPotongan)
      const byte = new Uint8Array(bagian.length)
      for (let index = 0; index < bagian.length; index += 1) {
        byte[index] = bagian.charCodeAt(index)
      }
      potongan.push(byte)
    }

    const url = URL.createObjectURL(new Blob(potongan, { type: 'application/pdf' }))
    const tautan = document.createElement('a')
    tautan.href = url
    tautan.download = `Klaim_${pasienTerpilih.value.no_sep}.pdf`
    document.body.appendChild(tautan)
    tautan.click()
    document.body.removeChild(tautan)
    window.setTimeout(() => URL.revokeObjectURL(url), 1000)
    return true
  } catch {
    return false
  }
}

function labelJenisKelamin(jk) {
  return String(jk).toUpperCase() === 'L' ? 'Laki-laki' : 'Perempuan'
}

function normalisasiStatusKlaim(status) {
  const nilai = String(status || 'belum')
    .trim()
    .toLowerCase()
    .replaceAll('-', '_')
    .replaceAll(' ', '_')

  const peta = {
    '': 'belum',
    belum: 'belum',
    new_claim: 'new_claim',
    set_klaim: 'set_claim_data',
    set_claim: 'set_claim_data',
    set_claim_data: 'set_claim_data',
    final: 'final_idrg',
    final_idrg: 'final_idrg',
    import_idrg: 'import_idrg',
    grouping_inacbg: 'grouping_inacbg',
    grouping_ina_cbg: 'grouping_inacbg',
    final_inacbg: 'final_inacbg',
    final_ina_cbg: 'final_inacbg',
    final_klaim: 'final_claim',
    final_claim: 'final_claim',
    sent: 'send_claim_individual',
    terkirim: 'send_claim_individual',
    send_claim: 'send_claim_individual',
    send_claim_individual: 'send_claim_individual',
  }

  return peta[nilai] || nilai
}

function statusKlaimLabel(status) {
  const daftar = {
    belum: 'Belum',
    new_claim: 'New Claim',
    set_claim_data: 'Set Klaim',
    final_idrg: 'Final IDRG',
    import_idrg: 'Import IDRG',
    grouping_inacbg: 'Grouping INA-CBG',
    final_inacbg: 'Final INA-CBG',
    final_claim: 'Final Klaim',
    send_claim_individual: 'Terkirim',
  }
  const statusNormal = normalisasiStatusKlaim(status)
  return daftar[statusNormal] || status || 'Belum'
}

function statusKlaimHeader(status) {
  return status && status !== 'belum' ? statusKlaimLabel(status) : 'Belum Klaim'
}

function pesanSuksesProses(aksi) {
  const daftar = {
    buat_klaim_baru: 'New Claim berhasil dikirim.',
    atur_data_klaim: 'Set Data Klaim berhasil dikirim.',
    idrg_diagnosa_set: 'Diagnosa IDRG berhasil dikirim.',
    idrg_diagnosa_get: 'Diagnosa IDRG berhasil dibaca.',
    idrg_procedure_set: 'Prosedur IDRG berhasil dikirim.',
    idrg_procedure_get: 'Prosedur IDRG berhasil dibaca.',
    grouping_idrg: 'Grouping IDRG berhasil dijalankan.',
    grouping_idrg_stage_1: 'Grouping IDRG Stage 1 berhasil dijalankan.',
    grouping_idrg_stage_2: 'Grouping IDRG Stage 2 berhasil dijalankan.',
    final_idrg: 'Final IDRG berhasil dijalankan.',
    reedit_idrg: 'Re-edit IDRG berhasil dijalankan.',
    idrg_ke_inacbg: 'Import IDRG ke INA-CBG berhasil dijalankan.',
    grouping_inacbg: 'Grouping INA-CBG berhasil dijalankan.',
    grouping_inacbg_stage_2: 'Grouping INA-CBG Tahap 2 berhasil dijalankan.',
    final_inacbg: 'Final INA-CBG berhasil dijalankan.',
    reedit_inacbg: 'Re-edit INA-CBG berhasil dijalankan.',
    final_klaim: 'Final Klaim berhasil dijalankan.',
    reedit_klaim: 'Edit ulang klaim berhasil dijalankan.',
    send_claim_individual: 'Kirim Klaim berhasil dijalankan.',
    kirim_klaim: 'Kirim Klaim berhasil dijalankan.',
    cetak_klaim: 'Cetak Klaim berhasil diproses.',
    validasi_sitb: 'Validasi SITB berhasil.',
  }
  return daftar[aksi] || 'Proses IDRG berhasil dijalankan.'
}

function pesanDariEclaim(hasil, fallback) {
  return hasil?.hasil?.pesan || hasil?.pesan || fallback
}

function tampilkanNotifikasiProses(hasil, fallback) {
  const pesan = pesanDariEclaim(hasil, fallback)
  if (!prosesBerhasil(hasil)) {
    notifikasi.gagal(pesan)
    return
  }
  notifikasi.sukses(pesan)
}

function prosesBerhasil(hasil) {
  return hasil?.sukses !== false && hasil?.hasil?.sukses !== false
}

function errorKlaimSudahFinal(error) {
  const pesan = String(error?.message || '').toLowerCase()
  return pesan.includes('sudah final') || pesan.includes('coding sudah final')
}

function hasilKlaimSudahFinal(hasil) {
  return kumpulkanPesanProses(hasil).some((pesan) => {
    const teks = String(pesan || '').toLowerCase()
    return teks.includes('sudah final') || teks.includes('coding sudah final')
  })
}

function kumpulkanPesanProses(hasil) {
  const pesan = [
    hasil?.pesan,
    hasil?.hasil?.pesan,
    hasil?.hasil?.metadata?.message,
    hasil?.hasil?.raw?.metadata?.message,
  ]

  if (Array.isArray(hasil?.tahapan)) {
    ;[...hasil.tahapan].reverse().forEach((tahap) => {
      pesan.push(tahap?.hasil?.pesan)
      pesan.push(tahap?.hasil?.metadata?.message)
      pesan.push(tahap?.hasil?.raw?.metadata?.message)
    })
  }

  return pesan.filter(Boolean)
}

function terapkanHasilGroupingIdrg(hasil) {
  const responseIdrg = ekstrakResponseIdrg(hasil)
  if (!responseIdrg) return null

  hasilGroupingIdrg.value = responseIdrg
  waktuGroupingIdrg.value = waktuIndoLengkap(new Date())
  return responseIdrg
}

function terapkanHasilGroupingInacbg(hasil) {
  const responseInacbg = ekstrakResponseInacbg(hasil)
  if (!responseInacbg) return null

  hasilGroupingInacbg.value = responseInacbg
  waktuGroupingInacbg.value = waktuIndoLengkap(new Date())
  return responseInacbg
}

function terapkanStatusDariDataKlaim(hasil) {
  const dataKlaim = hasil?.hasil?.response?.data || hasil?.hasil?.data?.data || hasil?.hasil?.data
  if (dataKlaim && typeof dataKlaim === 'object') {
    dataKlaimEclaim.value = dataKlaim
  }
  const statusIdrg = dataKlaim?.grouper?.response_idrg?.status_cd
  const statusInacbg = dataKlaim?.grouper?.response_inacbg?.status_cd
  const statusKlaim = dataKlaim?.klaim_status_cd
  const statusKemenkes = dataKlaim?.kemenkes_dc_status_cd || dataKlaim?.kemkes_dc_status_cd
  const statusBpjs = dataKlaim?.bpjs_dc_status_cd
  const sudahTerkirim = [statusKemenkes, statusBpjs]
    .some((status) => ['sent', 'terkirim'].includes(String(status || '').toLowerCase()))
  if (sudahTerkirim) {
    pasienTerpilih.value.status_klaim = 'send_claim_individual'
    return
  }
  if (String(statusKlaim || '').toLowerCase() === 'final') {
    pasienTerpilih.value.status_klaim = 'final_claim'
    return
  }
  if (String(statusInacbg || '').toLowerCase() === 'final') {
    pasienTerpilih.value.status_klaim = 'final_inacbg'
    return
  }
  if (String(statusIdrg || '').toLowerCase() === 'final') {
    pasienTerpilih.value.status_klaim = 'final_idrg'
    return
  }
  if (dataKlaim?.grouper?.response_inacbg) {
    pasienTerpilih.value.status_klaim = 'grouping_inacbg'
    return
  }
  if (dataKlaim?.grouper?.response_idrg) {
    pasienTerpilih.value.status_klaim = 'grouping_idrg'
    return
  }
  if (dataKlaim?.tgl_masuk || dataKlaim?.jenis_rawat || dataKlaim?.tarif_rs) {
    pasienTerpilih.value.status_klaim = 'set_claim_data'
  }
}

function kelasStatusKlaim(status) {
  return normalisasiStatusKlaim(status).replaceAll('_', '-')
}

function aksiKlaimDaftar(pasien) {
  const status = normalisasiStatusKlaim(pasien?.status_klaim)
  return status && status !== 'belum' ? 'Ulangi' : 'Grouping'
}

function kelasAksiKlaimDaftar(pasien) {
  return aksiKlaimDaftar(pasien) === 'Ulangi' ? 'warning' : ''
}

function labelTanggalPulang(pasien) {
  const tanggalKeluar = pasien?.tgl_keluar || ''
  const statusPulang = pasien?.status_pulang || ''
  if (pasien?.jenis_rawat_data === 'ralan') return tanggalKeluar
  if (!statusPulang || statusPulang === '-') return tanggalKeluar
  return `${tanggalKeluar} - ${statusPulang}`
}

function labelJenisRawat(pasien) {
  return pasien?.jenis_rawat_data === 'ralan' ? 'Rawat Jalan' : 'Rawat Inap'
}

function labelDiagnosa(pasien) {
  return pasien?.diagnosa_akhir || pasien?.diagnosa_awal || '-'
}

function labelCodingDiagnosa(item) {
  const kode = String(item?.kd_diag || item?.kd_penyakit || item?.kode || item?.code || '').trim()
  const nama = String(item?.nm_diag || item?.nama || item?.deskripsi || item?.display || '').trim()
  if (!kode && !nama) return '-'
  return nama ? `${kode || '-'} - ${nama}` : kode
}

function labelCodingProsedur(item) {
  const kode = String(item?.kd_prosedur || item?.kode || item?.code || '').trim()
  const nama = String(item?.nm_prosedur || item?.nama || item?.deskripsi || item?.display || '').trim()
  if (!kode && !nama) return '-'
  return nama ? `${kode || '-'} - ${nama}` : kode
}

function ubahCodingDiagnosaInacbg(item, nilai) {
  const [kode, ...nama] = String(nilai || '').split(' - ')
  item.kd_diag = kode.trim().toUpperCase()
  if (nama.length) item.nm_diag = nama.join(' - ').trim()
}

function ubahCodingProsedurInacbg(item, nilai) {
  const [kode, ...nama] = String(nilai || '').split(' - ')
  item.kd_prosedur = kode.trim().toUpperCase()
  if (nama.length) item.nm_prosedur = nama.join(' - ').trim()
}

function rupiah(nilai) {
  const angka = Number(nilai || 0)
  return new Intl.NumberFormat('id-ID').format(Number.isNaN(angka) ? 0 : angka)
}

function ubahTarifRS(key, nilai) {
  const angka = Number(String(nilai || '').replace(/[^\d]/g, ''))
  formTarifRS[key] = Number.isNaN(angka) ? 0 : angka
}

function labelOpsi(daftar, nilai, fallback = '-') {
  return daftar.find((item) => String(item.value) === String(nilai))?.label || fallback
}

function tanggalIndo(nilai) {
  if (!nilai) return '-'
  const tanggal = new Date(`${String(nilai).slice(0, 10)}T00:00:00`)
  if (Number.isNaN(tanggal.getTime())) return nilai
  return new Intl.DateTimeFormat('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  }).format(tanggal).replaceAll('.', '')
}

function tanggalWaktuIndo(tanggal, jam) {
  return `${tanggalIndo(tanggal)} ${jam || '00:00'}`
}

function waktuIndoLengkap(nilai) {
  const tanggal = nilai instanceof Date ? nilai : new Date(nilai)
  if (Number.isNaN(tanggal.getTime())) return '-'
  const pad = (n) => String(n).padStart(2, '0')
  return `${tanggalIndo(formatTanggal(tanggal))} ${pad(tanggal.getHours())}:${pad(tanggal.getMinutes())}`
}

function ambilNilai(data, daftarKey, fallback = '-') {
  for (const key of daftarKey) {
    const nilai = data?.[key]
    if (nilai !== undefined && nilai !== null && String(nilai).trim() !== '') return nilai
  }
  return fallback
}

function ambilRupiahGrouping(data, daftarKey) {
  const nilai = ambilNilai(data, daftarKey, '')
  if (nilai === '') return '-'
  if (typeof nilai === 'string' && nilai.toLowerCase().includes('rp')) return nilai
  return `Rp ${rupiah(String(nilai).replace(/[^\d.-]/g, ''))}`
}

function angkaDariNilai(nilai) {
  if (nilai === null || nilai === undefined || nilai === '-') return 0
  const angka = Number(String(nilai).replace(/[^\d.-]/g, ''))
  return Number.isNaN(angka) ? 0 : angka
}

function daftarSpecialCmgTampil(data) {
  const daftarAktif = Array.isArray(data?.special_cmg) ? data.special_cmg : []
  const kategori = ['Special Procedure', 'Special Prosthesis', 'Special Investigation', 'Special Drug']

  return kategori.map((tipe) => {
    const aktif = daftarAktif.find((item) => item.type === tipe)
    return {
      type: tipe,
      code: aktif?.code || '-',
      description: aktif?.description || 'None',
      tariff: angkaDariNilai(aktif?.tariff),
    }
  })
}

function formatBobotGrouping(nilai) {
  if (nilai === '-' || nilai === '' || nilai === null || nilai === undefined) return '-'
  const angka = Number(String(nilai).replace(',', '.'))
  if (Number.isNaN(angka)) return nilai
  return angka.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

function ekstrakResponseIdrg(hasil) {
  const daftarKandidat = [
    hasil,
    hasil?.hasil,
    hasil?.hasil?.raw,
    hasil?.hasil?.data,
    hasil?.hasil?.response,
    hasil?.hasil?.response?.data,
    hasil?.hasil?.response?.data?.grouper,
  ]

  if (Array.isArray(hasil?.tahapan)) {
    ;[...hasil.tahapan].reverse().forEach((tahap) => {
      daftarKandidat.push(tahap?.hasil)
      daftarKandidat.push(tahap?.hasil?.raw)
      daftarKandidat.push(tahap?.hasil?.data)
      daftarKandidat.push(tahap?.hasil?.response)
      daftarKandidat.push(tahap?.hasil?.response?.data)
      daftarKandidat.push(tahap?.hasil?.response?.data?.grouper)
    })
  }

  for (const kandidat of daftarKandidat) {
    const responseIdrg = kandidat?.response_idrg || kandidat?.grouper?.response_idrg
    if (responseIdrg) return responseIdrg
  }

  return null
}

function ekstrakResponseInacbg(hasil) {
  const daftarKandidat = [
    hasil,
    hasil?.hasil,
    hasil?.hasil?.raw,
    hasil?.hasil?.data,
    hasil?.hasil?.response,
    hasil?.hasil?.response?.data,
    hasil?.hasil?.response?.data?.grouper,
    dataKlaimEclaim.value?.grouper,
  ]

  if (Array.isArray(hasil?.tahapan)) {
    hasil.tahapan.forEach((tahap) => {
      daftarKandidat.push(tahap?.hasil)
      daftarKandidat.push(tahap?.hasil?.raw)
      daftarKandidat.push(tahap?.hasil?.data)
      daftarKandidat.push(tahap?.hasil?.response)
      daftarKandidat.push(tahap?.hasil?.response?.data)
      daftarKandidat.push(tahap?.hasil?.response?.data?.grouper)
    })
  }

  for (const kandidat of daftarKandidat) {
    const responseInacbg = kandidat?.response_inacbg || kandidat?.grouper?.response_inacbg
    if (responseInacbg) return responseInacbg
  }

  return null
}

function tipeKlaim() {
  return formKlaim.jenis_rawat === '1' ? 'RI' : 'RJ'
}

function umurPasien() {
  const lahir = new Date(`${pasienTerpilih.value?.tgl_lahir || ''}T00:00:00`)
  const rawat = new Date(`${formKlaim.tgl_masuk || hariIni}T00:00:00`)
  if (Number.isNaN(lahir.getTime()) || Number.isNaN(rawat.getTime())) return '-'

  let umur = rawat.getFullYear() - lahir.getFullYear()
  const belumUlangTahun = rawat.getMonth() < lahir.getMonth() || (rawat.getMonth() === lahir.getMonth() && rawat.getDate() < lahir.getDate())
  if (belumUlangTahun) umur -= 1
  return `${Math.max(umur, 0)} tahun`
}

function losHari() {
  const masuk = new Date(`${formKlaim.tgl_masuk || ''}T00:00:00`)
  const pulang = new Date(`${formKlaim.tgl_pulang || formKlaim.tgl_masuk || ''}T00:00:00`)
  if (Number.isNaN(masuk.getTime()) || Number.isNaN(pulang.getTime())) return '-'

  const selisih = Math.floor((pulang - masuk) / 86400000) + 1
  return `${Math.max(selisih, 1)} hari`
}

function tanggalJamKlaim(tanggal, jam) {
  if (!tanggal) return null
  let waktu = jam || '00:00'
  if (waktu.length === 5) waktu = `${waktu}:00`
  const hasil = new Date(`${tanggal}T${waktu}`)
  return Number.isNaN(hasil.getTime()) ? null : hasil
}

function losJam() {
  const masuk = tanggalJamKlaim(formKlaim.tgl_masuk, formKlaim.jam_masuk)
  const pulang = tanggalJamKlaim(formKlaim.tgl_pulang || formKlaim.tgl_masuk, formKlaim.jam_pulang)
  if (!masuk || !pulang) return '( - jam )'

  const totalMenit = Math.max(0, Math.round((pulang - masuk) / 60000))
  const jam = Math.floor(totalMenit / 60)
  const menit = String(totalMenit % 60).padStart(2, '0')
  return `( ${jam}:${menit} jam )`
}

function totalTarifRS() {
  return daftarTarifRS.value.reduce((total, item) => total + (Number(item.nilai || 0) || 0), 0)
}

function tarifKosong() {
  return definisiTarifRS.reduce((hasil, [key]) => {
    hasil[key] = 0
    return hasil
  }, {})
}

function isiFormKlaim(pasien) {
  const tanggalMasuk = pasien?.tgl_masuk || hariIni
  const tanggalPulang = pasien?.tgl_keluar || tanggalMasuk
  formKlaim.jaminan = 'JKN'
  formKlaim.nomor_kartu_t = 'kartu_jkn'
  formKlaim.cob_cd = '0'
  formKlaim.jenis_rawat = pasien?.jenis_rawat_data === 'ralan' ? '2' : '1'
  formKlaim.kelas_rawat = String(pasien?.kelas_rawat || '3')
  formKlaim.tgl_masuk = tanggalMasuk
  formKlaim.jam_masuk = '08:00'
  formKlaim.tgl_pulang = tanggalPulang
  formKlaim.jam_pulang = '09:00'
  formKlaim.cara_masuk = 'gp'
  formKlaim.discharge_status = statusPulangEklaim(pasien?.status_pulang)
  formKlaim.adl_sub_acute = '0'
  formKlaim.adl_chronic = '0'
  formKlaim.birth_weight = '0'
  formKlaim.sistole = '110'
  formKlaim.diastole = '60'
  formKlaim.kantong_darah = '0'
  formKlaim.alteplase_ind = '0'
  formKlaim.icu_indikator = false
  formKlaim.icu_los = '0'
  formKlaim.ventilator_use_ind = false
  formKlaim.ventilator_hour = '0'
  formKlaim.ventilator_start_dttm = ''
  formKlaim.ventilator_stop_dttm = ''
  formKlaim.upgrade_class_ind = false
  formKlaim.upgrade_class_class = ''
  formKlaim.upgrade_class_los = '0'
  formKlaim.upgrade_class_payor = 'peserta'
  formKlaim.add_payment_pct = '0'
  formKlaim.tarif_poli_eks = '0'
  formKlaim.topup_codes = ''
  formKlaim.nama_dokter = pasien?.nm_dokter || '-'
  formKlaim.kode_tarif = 'BP'
  formKlaim.pasien_tb = false
  formKlaim.nomor_register_sitb = ''
  formKlaim.tarif_dikonfirmasi = true

  const tarif = pasien?.tarif_rs || {}
  definisiTarifRS.forEach(([key]) => {
    formTarifRS[key] = Number(tarif[key] || 0)
  })

  formKlaim.tampilkan_apgar = false
  formKlaim.apgar = {
    menit_1: { appearance: 0, pulse: 0, grimace: 0, activity: 0, respiration: 0 },
    menit_5: { appearance: 0, pulse: 0, grimace: 0, activity: 0, respiration: 0 }
  }
  formKlaim.tampilkan_persalinan = false
  formKlaim.persalinan = {
    usia_kehamilan: '0', gravida: '0', partus: '0', abortus: '0', onset_kontraksi: 'spontan',
    delivery: [
      { delivery_sequence: '1', delivery_method: 'vaginal', delivery_dttm: '', letak_janin: 'kepala', kondisi: 'livebirth', use_manual: '0', use_forcep: '0', use_vacuum: '0', shk_spesimen_ambil: 'tidak', shk_lokasi: 'tumit', shk_alasan: '', shk_spesimen_dttm: '' }
    ]
  }
}

function statusPulangEklaim(status) {
  const nilai = String(status || '').toLowerCase()
  if (nilai.includes('rujuk')) return '2'
  if (nilai.includes('paksa') || nilai.includes('aps') || nilai.includes('permintaan sendiri')) return '3'
  if (nilai.includes('meninggal')) return '4'
  if (nilai.includes('lain')) return '5'
  return '1'
}

function stringDiagnosaKlaim(daftar = diagnosa.value) {
  return daftar
    .map((item) => String(item.kd_diag || item.kd_penyakit || item.kode || item.code || '').trim())
    .filter(Boolean)
    .join('#')
}

function stringProsedurKlaim(daftar = prosedur.value) {
  return daftar
    .map((item) => {
      const kode = String(item.kd_prosedur || item.kode || item.code || '').trim()
      if (!kode) return ''
      const jumlah = Number(item.multiplicity || item.jumlah || 1)
      return jumlah > 1 ? `${kode}+${jumlah}` : kode
    })
    .filter(Boolean)
    .join('#')
}

function payloadKlaim(aksi = '') {
  const payload = { ...formKlaim, tarif_rs: { ...formTarifRS } }
  if (!payload.tampilkan_apgar) delete payload.apgar
  if (!payload.tampilkan_persalinan) delete payload.persalinan
  payload.ventilator_start_dttm = formatTanggalWaktu(payload.ventilator_start_dttm)
  payload.ventilator_stop_dttm = formatTanggalWaktu(payload.ventilator_stop_dttm)
  const memakaiCodingInacbg = String(aksi).startsWith('grouping_inacbg')
  payload.diagnosa = memakaiCodingInacbg ? stringDiagnosaKlaim(diagnosaInacbg.value) : stringDiagnosaKlaim()
  payload.prosedur = memakaiCodingInacbg ? (stringProsedurKlaim(prosedurInacbg.value) || '#') : stringProsedurKlaim()
  if (aksi === 'grouping_idrg') payload.topup_codes = ''
  return payload
}
</script>

<template>
  <section class="idrg-module" :class="{ 'sidebar-closed': sidebarTertutup }">
    <aside class="idrg-sidebar">
      <div class="idrg-brand">
        <i><FileSpreadsheet :size="23" /></i>
        <span v-if="!sidebarTertutup">
          <small>Bridging</small>
          <strong>E-Klaim</strong>
        </span>
        <button type="button" class="idrg-sidebar-toggle" @click="toggleSidebar" :title="sidebarTertutup ? 'Buka Sidebar' : 'Sembunyikan Sidebar'">
          <PanelLeftOpen v-if="sidebarTertutup" :size="18" />
          <PanelLeftClose v-else :size="18" />
        </button>
      </div>

      <nav>
        <button
          v-for="menu in menuSidebar"
          :key="menu.id"
          type="button"
          :class="{ active: menuSidebarAktif === menu.id }"
          @click="pilihMenuSidebar(menu.id)"
        >
          <component :is="menu.icon" :size="18" />
          <span>{{ menu.label }}</span>
          <b v-if="menu.badge">{{ paginasi.total ?? 0 }}</b>
        </button>
      </nav>

      <div class="idrg-sidebar-note">
        <strong>IDRG / INA-CBG</strong>
        <span>Alur dibuat mengikuti SIMRS lama, dipindahkan bertahap ke SIRAVA.</span>
      </div>
    </aside>

    <main class="idrg-main">
      <template v-if="menuSidebarAktif === 'pengajuan'">
        <form v-if="!pasienTerpilih" class="idrg-filter" @submit.prevent="terapkanFilter">
          <label>
            <span>Tanggal</span>
            <DatePicker v-model="tanggalFilter" placeholder="Tanggal pasien" :disabled="loadingDaftar" />
          </label>
          <label>
            <span>Jenis Rawat</span>
            <Select v-model="filter.jenis_rawat" :options="opsiJenisRawat" :disabled="loadingDaftar" />
          </label>
          <label class="idrg-search">
            <span>Pencarian</span>
            <i>
              <Search :size="17" />
              <input v-model="filter.cari" type="search" placeholder="Cari no rawat, RM, pasien, SEP..." :disabled="loadingDaftar" />
            </i>
          </label>
          <button type="submit" :disabled="loadingDaftar">
            <LoaderCircle v-if="loadingDaftar" class="spin" :size="17" />
            <Search v-else :size="17" />
            Cari
          </button>
        </form>

        <div v-if="!pasienTerpilih" class="idrg-card idrg-list">
          <DataTable
            class="idrg-claim-table"
            :rows="daftarPasien"
            data-key="no_rawat"
            :loading="loadingDaftar"
            loading-message="Menarik pasien E-Klaim..."
            empty-message="Data pasien klaim tidak ditemukan."
            paginator
            lazy
            :first="(paginasi.halaman - 1) * paginasi.batas"
            :rows-per-page="paginasi.batas"
            :total-records="paginasi.total"
            @page="gantiHalaman"
          >
            <Column header="No SEP / Rawat">
              <template #body="{ data }">
                <strong class="idrg-sep">{{ data.no_sep || 'SEP belum ada' }}</strong>
                <span>{{ data.no_rawat }}</span>
              </template>
            </Column>
            <Column header="No RM">
              <template #body="{ data }">
                <strong class="idrg-rm">{{ data.no_rkm_medis }}</strong>
              </template>
            </Column>
            <Column header="Nama Pasien">
              <template #body="{ data }">
                <strong>{{ data.nm_pasien }}</strong>
                <span>{{ labelJenisRawat(data) }}</span>
              </template>
            </Column>
            <Column header="Dokter">
              <template #body="{ data }">
                <strong>{{ data.nm_dokter || '-' }}</strong>
                <span>{{ data.penanggung || '-' }}</span>
              </template>
            </Column>
            <Column header="Ruang/Poli">
              <template #body="{ data }">
                <strong>{{ data.kamar || '-' }}</strong>
              </template>
            </Column>
            <Column header="Tanggal">
              <template #body="{ data }">
                <strong>{{ data.tgl_masuk }}</strong>
                <span v-if="labelTanggalPulang(data)">Pulang {{ labelTanggalPulang(data) }}</span>
              </template>
            </Column>
            <Column header="Diagnosa">
              <template #body="{ data }">
                <strong>{{ labelDiagnosa(data) }}</strong>
              </template>
            </Column>
            <Column header="Aksi">
              <template #body="{ data }">
                <div class="idrg-action-cell">
                  <button class="table-action-button" :class="kelasAksiKlaimDaftar(data)" type="button" @click="pilihPasien(data)">
                    <Database :size="13" />
                    {{ aksiKlaimDaftar(data) }}
                  </button>
                </div>
              </template>
            </Column>
          </DataTable>
        </div>

        <div v-else class="idrg-detail-page">
          <header class="idrg-detail-top">
            <div class="idrg-detail-title">
              <button class="idrg-icon-back" type="button" @click="pasienTerpilih = null">
                <ArrowLeft :size="16" />
              </button>
              <div>
                <h2>
                  {{ pasienTerpilih.nm_pasien }}
                  <span class="idrg-status-pill" :class="kelasStatusKlaim(pasienTerpilih.status_klaim)">
                    {{ statusKlaimHeader(pasienTerpilih.status_klaim) }}
                  </span>
                </h2>
                <p>RM {{ pasienTerpilih.no_rkm_medis }} · {{ labelJenisKelamin(pasienTerpilih.jk) }} · SEP: <b>{{ pasienTerpilih.no_sep || '-' }}</b></p>
              </div>
            </div>
            <div class="idrg-detail-tools">
              <button type="button" :disabled="loadingDetail" @click="muatCodingPasien">
                <LoaderCircle v-if="loadingDetail" class="spin" :size="15" />
                <RefreshCw v-else :size="15" />
                Cek Status
              </button>
              <button type="button" :disabled="!pasienTerpilih.no_sep || !klaimSudahFinal || Boolean(loadingProses)" :title="klaimSudahFinal ? '' : 'Final Klaim harus berhasil terlebih dahulu'" @click="jalankanProses('cetak_klaim')">
                <LoaderCircle v-if="loadingProses === 'cetak_klaim'" class="spin" :size="15" />
                <FileText v-else :size="15" />
                Cetak
              </button>
            </div>
          </header>

          <nav class="idrg-tabs idrg-detail-tabs">
            <button :class="{ active: tabAktif === 'data-pasien' }" type="button" @click="pilihTabDetail('data-pasien')">Data Pasien</button>
            <button :class="{ active: tabAktif === 'idrg' }" type="button" :disabled="!bolehBukaIdrg" :title="bolehBukaIdrg ? '' : 'Set Data Klaim harus berhasil terlebih dahulu'" @click="pilihTabDetail('idrg')">IDRG</button>
            <button :class="{ active: tabAktif === 'inacbg' }" type="button" :disabled="!bolehBukaInacbg" :title="bolehBukaInacbg ? '' : 'Final IDRG harus berhasil terlebih dahulu'" @click="pilihTabDetail('inacbg')">INA-CBG</button>
          </nav>

          <section v-if="tabAktif === 'data-pasien'" class="idrg-detail-content">
            <article class="idrg-claim-sheet">
              <div class="claim-summary-grid">
                <span><small>Tanggal Masuk</small><strong>{{ tanggalIndo(formKlaim.tgl_masuk) }}</strong></span>
                <span><small>Tanggal Pulang</small><strong>{{ tanggalIndo(formKlaim.tgl_pulang) }}</strong></span>
                <span><small>Jaminan</small><strong>{{ formKlaim.jaminan }}</strong></span>
                <span><small>No. SEP</small><strong>{{ pasienTerpilih.no_sep || '-' }}</strong></span>
                <span><small>Tipe</small><strong>{{ tipeKlaim() }}</strong></span>
                <span><small>CBG</small><strong>-</strong></span>
                <span><small>Status</small><strong>{{ statusKlaimLabel(pasienTerpilih.status_klaim) }}</strong></span>
                <span><small>Petugas</small><strong>{{ props.user?.name || '-' }}</strong></span>
              </div>

              <div class="claim-sheet-top">
                <label>
                  <small>Jaminan / Cara Bayar</small>
                  <select v-model="formKlaim.jaminan">
                    <option value="JKN">JKN</option>
                    <option value="UMUM">UMUM</option>
                  </select>
                </label>
                <label>
                  <small>No. Peserta</small>
                  <input :value="pasienTerpilih.no_kartu || '-'" type="text" disabled />
                </label>
                <label>
                  <small>No. SEP</small>
                  <input :value="pasienTerpilih.no_sep || '-'" type="text" disabled />
                </label>
                <label>
                  <small>COB</small>
                  <select v-model="formKlaim.cob_cd">
                    <option v-for="item in opsiCob" :key="item.value" :value="item.value">{{ item.label }}</option>
                  </select>
                </label>
              </div>

              <div class="claim-sheet-row">
                <b>Jenis Rawat</b>
                <div class="claim-value">
                  <label class="claim-radio"><input v-model="formKlaim.jenis_rawat" type="radio" value="2" /> Jalan</label>
                  <label class="claim-radio"><input v-model="formKlaim.jenis_rawat" type="radio" value="1" /> Inap</label>
                  <template v-if="formKlaim.jenis_rawat === '2'">
                    <label class="claim-radio claim-help"><span>?</span><input :checked="Number(formKlaim.tarif_poli_eks) > 0" type="checkbox" @change="formKlaim.tarif_poli_eks = $event.target.checked ? '1' : '0'" /> Kelas Eksekutif</label>
                  </template>
                  <template v-else>
                    <label class="claim-radio claim-help"><span>?</span><input v-model="formKlaim.upgrade_class_ind" type="checkbox" /> Naik/Turun Kelas</label>
                    <label class="claim-radio claim-help"><span>?</span><input v-model="formKlaim.icu_indikator" type="checkbox" /> Ada Rawat Intensif</label>
                  </template>
                </div>
                <b>Kelas Hak</b>
                <div class="claim-value">
                  <template v-if="formKlaim.jenis_rawat === '1'">
                    <label v-for="item in opsiKelasRawat" :key="item.value" class="claim-radio">
                      <input v-model="formKlaim.kelas_rawat" type="radio" :value="item.value" />
                      {{ item.label }}
                    </label>
                  </template>
                  <span v-else>-</span>
                </div>
              </div>

              <div v-if="formKlaim.jenis_rawat === '1' && formKlaim.upgrade_class_ind" class="claim-sheet-row claim-upgrade-row">
                <b><span class="claim-help-icon">?</span> Kelas Pelayanan</b>
                <div class="claim-value claim-upgrade-class">
                  <label v-for="item in opsiKelasPelayanan" :key="item.value" class="claim-radio">
                    <input v-model="formKlaim.upgrade_class_class" type="radio" :value="item.value" />
                    {{ item.label }}
                  </label>
                </div>
                <b><span class="claim-help-icon">?</span> Lama <span>(hari)</span></b>
                <div class="claim-value">
                  <input v-model="formKlaim.upgrade_class_los" type="number" min="0" />
                </div>
              </div>

              <div v-if="formKlaim.jenis_rawat === '1' && formKlaim.icu_indikator" class="claim-sheet-row claim-intensive-row">
                <b><span class="claim-help-icon">?</span> Ventilator</b>
                <div class="claim-value claim-intensive-main">
                  <label class="claim-radio"><input v-model="formKlaim.ventilator_use_ind" type="checkbox" /> Ya</label>
                  <template v-if="formKlaim.ventilator_use_ind">
                    <span>Intubasi : <DatePicker v-model="formKlaim.ventilator_start_dttm" :show-icon="false" show-time hour-format="24" placeholder="Waktu Intubasi" /></span>
                    <span>Ekstubasi : <DatePicker v-model="formKlaim.ventilator_stop_dttm" :show-icon="false" show-time hour-format="24" placeholder="Waktu Ekstubasi" /></span>
                  </template>
                </div>
                <b><span class="claim-help-icon">?</span> Rawat Intensif <span>(hari)</span></b>
                <div class="claim-value">
                  <input v-model="formKlaim.icu_los" type="number" min="0" />
                </div>
              </div>

              <div class="claim-sheet-row">
                <b>Tanggal Rawat</b>
                <div class="claim-value claim-date-line">
                  <span>Masuk : 
                    <DatePicker class="w-date" :model-value="parseTanggal(formKlaim.tgl_masuk)" @update:model-value="val => formKlaim.tgl_masuk = formatTanggal(val)" :show-icon="false" placeholder="Tgl Masuk" />
                    <DatePicker class="w-time" :model-value="parseWaktu(formKlaim.jam_masuk)" @update:model-value="val => formKlaim.jam_masuk = formatWaktu(val)" time-only hour-format="24" :show-icon="false" placeholder="Jam" />
                  </span>
                  <span>Pulang : 
                    <DatePicker class="w-date" :model-value="parseTanggal(formKlaim.tgl_pulang)" @update:model-value="val => formKlaim.tgl_pulang = formatTanggal(val)" :show-icon="false" placeholder="Tgl Pulang" />
                    <DatePicker class="w-time" :model-value="parseWaktu(formKlaim.jam_pulang)" @update:model-value="val => formKlaim.jam_pulang = formatWaktu(val)" time-only hour-format="24" :show-icon="false" placeholder="Jam" />
                  </span>
                </div>
                <b>Umur</b>
                <div class="claim-value">{{ umurPasien() }}</div>
              </div>

              <div class="claim-sheet-row single">
                <b>Cara Masuk</b>
                <div class="claim-value claim-select-compact">
                  <select v-model="formKlaim.cara_masuk">
                    <option v-for="item in opsiCaraMasuk" :key="item.value" :value="item.value">{{ item.label }}</option>
                  </select>
                </div>
              </div>

              <div class="claim-sheet-row">
                <b>LOS</b>
                <div class="claim-value claim-los-line">
                  <span>{{ losHari() }}</span>
                  <strong>{{ losJam() }}</strong>
                </div>
                <b>Berat Lahir <span>(gram)</span></b>
                <div class="claim-value"><input v-model="formKlaim.birth_weight" type="number" min="0" /></div>
              </div>

              <div class="claim-sheet-row">
                <b>ADL Score</b>
                <div class="claim-value claim-date-line">
                  <span>Sub Acute : <input v-model="formKlaim.adl_sub_acute" type="number" min="0" /></span>
                  <span>Chronic : <input v-model="formKlaim.adl_chronic" type="number" min="0" /></span>
                </div>
                <b>Cara Pulang</b>
                <div class="claim-value">
                  <select v-model="formKlaim.discharge_status">
                    <option v-for="item in opsiCaraPulang" :key="item.value" :value="item.value">{{ item.label }}</option>
                  </select>
                </div>
              </div>

              <div class="claim-sheet-row tall">
                <b>DPJP</b>
                <div class="claim-value"><textarea v-model="formKlaim.nama_dokter" rows="5"></textarea></div>
                <b>Jenis Tarif</b>
                <div class="claim-value">
                  <select v-model="formKlaim.kode_tarif">
                    <option v-for="item in opsiKodeTarif" :key="item.value" :value="item.value">{{ item.label }}</option>
                  </select>
                </div>
              </div>

              <div class="claim-sheet-row single">
                <b>Pasien TB</b>
                <div class="claim-value claim-tb-line">
                  <label class="claim-radio"><input v-model="formKlaim.pasien_tb" type="checkbox" /> Ya</label>
                  <input v-if="formKlaim.pasien_tb" v-model="formKlaim.nomor_register_sitb" type="text" placeholder="No register SITB" />
                  <button v-if="formKlaim.pasien_tb" type="button" class="table-action-button" :disabled="loadingProses === 'validasi_sitb'" @click="jalankanProses('validasi_sitb')">
                    Validasi SITB
                  </button>
                </div>
              </div>

              <div class="claim-sheet-row">
                <b>Data Apgar</b>
                <div class="claim-value">
                  <label class="claim-radio"><input v-model="formKlaim.tampilkan_apgar" type="checkbox" /> Tambahkan Apgar</label>
                </div>
                <b>Data Persalinan</b>
                <div class="claim-value">
                  <label class="claim-radio"><input v-model="formKlaim.tampilkan_persalinan" type="checkbox" /> Tambahkan Persalinan</label>
                </div>
              </div>

            </article>

            <article class="idrg-claim-sheet idrg-tarif-sheet">
              <header class="claim-tarif-total">
                <span class="claim-help-icon">?</span>
                <em>Tarif Rumah Sakit :</em>
                <strong>Rp {{ rupiah(totalTarifRS()) }}</strong>
              </header>
              <div class="claim-tarif-grid">
                <label v-for="item in daftarTarifRS" :key="item.key" class="claim-tarif-item">
                  <span class="claim-help-icon">?</span>
                  <small>{{ item.label }}</small>
                  <input :value="rupiah(formTarifRS[item.key])" type="text" inputmode="numeric" @input="ubahTarifRS(item.key, $event.target.value)" />
                </label>
              </div>
              <label class="claim-tarif-confirm">
                <input v-model="formKlaim.tarif_dikonfirmasi" type="checkbox" />
                <span>Menyatakan benar bahwa data tarif yang tersebut di atas adalah benar sesuai dengan kondisi yang sesungguhnya.</span>
              </label>
              <div class="idrg-claim-actions">
                <span v-if="loadingNewClaimOtomatis" class="idrg-auto-claim-status">
                  <LoaderCircle class="spin" :size="14" />
                  New Claim otomatis sedang dikirim...
                </span>
                <button type="button" :disabled="loadingNewClaimOtomatis || loadingProses === 'atur_data_klaim' || !formKlaim.tarif_dikonfirmasi" @click="jalankanProses('atur_data_klaim')">
                  <LoaderCircle v-if="loadingProses === 'atur_data_klaim'" class="spin" :size="16" />
                  <ClipboardList v-else :size="16" />
                  Set Data Klaim
                </button>
              </div>
            </article>
          </section>

          <section v-else-if="tabAktif === 'idrg'" class="idrg-panel-card idrg-detail-panel" :class="{ 'idrg-final-locked': idrgSudahFinal }">
            <header class="coding-work-header">
              <div>
                <span>IDRG</span>
                <h3>Coding dan Grouping IDRG</h3>
              </div>
              <div class="coding-work-actions">
                <button type="button" :disabled="loadingCodingIdrg || idrgSudahFinal" @click="muatCodingIdrg('diagnosa')">
                  <LoaderCircle v-if="loadingCodingIdrg === 'diagnosa'" class="spin" :size="15" />
                  <RefreshCw v-else :size="15" />
                  Muat Diagnosa
                </button>
                <button type="button" :disabled="loadingCodingIdrg || idrgSudahFinal" @click="muatCodingIdrg('prosedur')">
                  <LoaderCircle v-if="loadingCodingIdrg === 'prosedur'" class="spin" :size="15" />
                  <RefreshCw v-else :size="15" />
                  Muat Prosedur
                </button>
              </div>
            </header>

            <div class="idrg-coding-grid idrg-coding-grid-table">
              <article>
                <div class="coding-header eklaim-header">
                  <h4>Diagnosa <span>(ICD-10)</span>:</h4>
                </div>
                <div class="idrg-table-container">
                  <div v-if="diagnosa.length === 0" class="idrg-table-empty">Belum ada diagnosa. Klik "Muat Diagnosa".</div>
                  <table v-else class="idrg-coding-table">
                    <thead>
                      <tr>
                        <th>Kode</th>
                        <th>Nama Diagnosa</th>
                        <th>Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      <template v-for="(item, index) in diagnosa" :key="`idrg-diagnosa-${index}`">
                      <tr>
                        <td class="col-kode">
                          <button type="button" class="coding-code-button" :disabled="idrgSudahFinal" @click="bukaEditorDiagnosa(index)">{{ item.kd_diag || '-' }}</button>
                        </td>
                        <td class="col-nama">
                          <button type="button" class="idrg-btn-edit-nama" :disabled="idrgSudahFinal" @click="bukaEditorDiagnosa(index)">
                            {{ item.nm_diag || 'Klik untuk cari diagnosa' }}
                          </button>
                        </td>
                        <td class="col-status">
                          <span class="status-badge" :class="{ 'is-primer': index === 0 }">{{ index === 0 ? 'Primer' : 'Sekunder' }}</span>
                        </td>
                      </tr>
                      <tr v-if="!idrgSudahFinal && activeSubstitusiDiagnosa === index" class="coding-edit-row">
                        <td colspan="3">
                          <div class="coding-inline-editor">
                            <AutoComplete
                              v-model="substitusiKeyword"
                              :suggestions="searchHasilDiagnosa"
                              option-label="label"
                              @complete="onSearchDiagnosa"
                              @item-select="applySubstitusi(index, $event)"
                              placeholder="Substitusi"
                              class="eklaim-substitusi"
                            >
                              <template #option="slotProps">
                                <span v-if="slotProps.option"><b>{{ slotProps.option.kode }}</b> - {{ slotProps.option.deskripsi }}</span>
                              </template>
                            </AutoComplete>
                            <button type="button" class="danger" @click="hapusDiagnosa(index)">Delete</button>
                            <button type="button" :disabled="index === 0" @click="setPrimaryDiagnosa(index)">Set Primary</button>
                          </div>
                        </td>
                      </tr>
                      </template>
                    </tbody>
                  </table>
                </div>
              </article>
              <article>
                <div class="coding-header eklaim-header">
                  <h4>Prosedur <span>(ICD-9CM)</span>:</h4>
                </div>
                <div class="idrg-table-container">
                  <div v-if="prosedur.length === 0" class="idrg-table-empty">Belum ada prosedur. Klik "Muat Prosedur".</div>
                  <table v-else class="idrg-coding-table">
                    <thead>
                      <tr>
                        <th>Kode</th>
                        <th>Nama Prosedur</th>
                        <th>Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      <template v-for="(item, index) in prosedur" :key="`idrg-prosedur-${index}`">
                      <tr>
                        <td class="col-kode">
                          <button type="button" class="coding-code-button" :disabled="idrgSudahFinal" @click="bukaEditorProsedur(index)">{{ item.kd_prosedur || '-' }}</button>
                        </td>
                        <td class="col-nama">
                          <button type="button" class="idrg-btn-edit-nama" :disabled="idrgSudahFinal" @click="bukaEditorProsedur(index)">
                            {{ item.nm_prosedur || 'Klik untuk cari prosedur' }}
                          </button>
                        </td>
                        <td class="col-status">
                          <span class="status-badge">{{ index === 0 ? 'Primer' : 'Sekunder' }}</span>
                        </td>
                      </tr>
                      <tr v-if="!idrgSudahFinal && activeSubstitusiProsedur === index" class="coding-edit-row">
                        <td colspan="3">
                          <div class="coding-inline-editor">
                            <AutoComplete
                              v-model="substitusiKeyword"
                              :suggestions="searchHasilProsedur"
                              option-label="label"
                              @complete="onSearchProsedur"
                              @item-select="applySubstitusiProsedur(index, $event)"
                              placeholder="Substitusi"
                              class="eklaim-substitusi"
                            >
                              <template #option="slotProps">
                                <span v-if="slotProps.option"><b>{{ slotProps.option.kode }}</b> - {{ slotProps.option.deskripsi }}</span>
                              </template>
                            </AutoComplete>
                            <button type="button" class="danger" @click="hapusProsedur(index)">Delete</button>
                          </div>
                        </td>
                      </tr>
                      </template>
                    </tbody>
                  </table>
                </div>
              </article>
            </div>

            <section class="coding-result-card" :class="{ final: idrgGroupingFinal }">
              <header>
                <h4>Hasil Grouping IDRG{{ idrgGroupingFinal ? ' - Final' : '' }}</h4>
                <div class="coding-result-actions">
                  <button v-if="idrgBisaEditUlang" type="button" class="edit-action" :disabled="!pasienTerpilih.no_sep || loadingProses === 'reedit_idrg'" @click="jalankanProses('reedit_idrg')">
                    <LoaderCircle v-if="loadingProses === 'reedit_idrg'" class="spin" :size="15" />
                    <Edit3 v-else :size="15" />
                    Edit Ulang IDRG
                  </button>
                  <template v-else>
                    <button type="button" class="grouping-action" :disabled="!pasienTerpilih.no_sep || loadingProses === 'grouping_idrg'" @click="jalankanProses('grouping_idrg')">
                      <LoaderCircle v-if="loadingProses === 'grouping_idrg'" class="spin" :size="15" />
                      <Layers v-else :size="15" />
                      Grouping
                    </button>
                    <button type="button" class="final-action" :disabled="!pasienTerpilih.no_sep || loadingProses === 'final_idrg' || !responseIdrgTerakhir" @click="jalankanProses('final_idrg')">
                      <LoaderCircle v-if="loadingProses === 'final_idrg'" class="spin" :size="15" />
                      <CheckCircle2 v-else :size="15" />
                      Final IDRG
                    </button>
                  </template>
                </div>
              </header>

              <div v-if="responseIdrgTerakhir" class="idrg-grouping-sheet" :class="{ final: idrgGroupingFinal }">
              <h4>Hasil Grouping iDRG{{ idrgGroupingFinal ? ' - Final' : '' }}</h4>
              <div class="idrg-grouping-table-wrap">
                <table>
                  <tbody>
                    <tr>
                      <th>Info</th>
                      <td colspan="3">{{ hasilGroupingIdrgTabel.info }}</td>
                    </tr>
                    <tr>
                      <th>Jenis Rawat</th>
                      <td colspan="3">{{ hasilGroupingIdrgTabel.jenisRawat }}</td>
                    </tr>
                    <tr>
                      <th>MDC</th>
                      <td>{{ hasilGroupingIdrgTabel.mdcDeskripsi }}</td>
                      <td class="text-right" colspan="2">{{ hasilGroupingIdrgTabel.mdcNomor }}</td>
                    </tr>
                    <tr>
                      <th>DRG</th>
                      <td>{{ hasilGroupingIdrgTabel.drgDeskripsi }}</td>
                      <td class="text-right">{{ hasilGroupingIdrgTabel.drgKode }}</td>
                      <td class="weight-cell"><span>DRG Cost Weight:</span><strong>** {{ hasilGroupingIdrgTabel.drgCostWeight }}</strong></td>
                    </tr>
                    <tr>
                      <th>KRIS</th>
                      <td>{{ hasilGroupingIdrgTabel.kris }}</td>
                      <td></td>
                      <td class="weight-cell"><span>KRIS Cost Weight:</span><strong>** {{ hasilGroupingIdrgTabel.krisCostWeight }}</strong></td>
                    </tr>
                    <tr>
                      <th>NBR</th>
                      <td class="blue-value">** {{ hasilGroupingIdrgTabel.nbr }}</td>
                      <td></td>
                      <td class="weight-cell"><span>Total Cost Weight:</span><strong>** {{ hasilGroupingIdrgTabel.totalCostWeight }}</strong></td>
                    </tr>
                    <tr>
                      <th>Total Klaim</th>
                      <td></td>
                      <td class="text-right">{{ hasilGroupingIdrgTabel.totalKlaim === '-' ? '' : 'Rp' }}</td>
                      <td class="blue-value strong-value">{{ hasilGroupingIdrgTabel.totalKlaim === '-' ? '-' : hasilGroupingIdrgTabel.totalKlaim.replace(/^Rp\s*/i, '') }}</td>
                    </tr>
                    <tr>
                      <th>Status</th>
                      <td colspan="3">{{ hasilGroupingIdrgTabel.status }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <p class="idrg-grouping-note">**) Catatan: Nilai belum final, sewaktu-waktu bisa berubah</p>

              <ul v-if="opsiTopupIdrg.length" class="idrg-grouping-topup">
                <li v-for="item in opsiTopupIdrg" :key="item.code">
                  <b>{{ item.code }}</b>
                  <span>{{ item.description }}</span>
                  <small>Rp {{ rupiah(item.tariff || 0) }}</small>
                </li>
              </ul>

              </div>
              <div v-else class="coding-result-empty">Hasil grouping IDRG belum tersedia.</div>
            </section>

          </section>

          <section v-else class="idrg-panel-card idrg-detail-panel inacbg-panel">
            <section class="inacbg-work-card">
              <header class="inacbg-work-header">
                <div>
                  <span>INA-CBG</span>
                  <h3>Coding dan Grouping INA-CBG</h3>
                </div>
                <div class="inacbg-header-actions">
                  <button type="button" class="purple" :disabled="!pasienTerpilih.no_sep || loadingImportInacbg || inacbgSudahFinal" @click="importIdrgKeInacbg">
                    <LoaderCircle v-if="loadingImportInacbg" class="spin" :size="15" />
                    <FileSpreadsheet v-else :size="15" />
                    Import IDRG ke INA-CBG
                  </button>
                  <button type="button" class="coding-load-action" :disabled="loadingProses || inacbgSudahFinal" @click="muatCodingInacbg('diagnosa')">
                    <LoaderCircle v-if="loadingProses === 'inacbg_diagnosa_get'" class="spin" :size="15" />
                    <RefreshCw v-else :size="15" />
                    Muat Diagnosa
                  </button>
                  <button type="button" class="coding-load-action" :disabled="loadingProses || inacbgSudahFinal" @click="muatCodingInacbg('prosedur')">
                    <LoaderCircle v-if="loadingProses === 'inacbg_procedure_get'" class="spin" :size="15" />
                    <RefreshCw v-else :size="15" />
                    Muat Prosedur
                  </button>
                </div>
              </header>

              <div class="inacbg-code-form">
                <article>
                  <h4>Diagnosa <span>(ICD-10)</span></h4>
                  <div v-if="diagnosaInacbg.length === 0" class="inacbg-empty">Klik Import IDRG ke INA-CBG untuk menampilkan coding.</div>
                  <div v-else class="coding-list-table">
                    <div class="coding-list-head"><span>Kode</span><span>Nama Diagnosa</span><span>Status</span></div>
                    <template v-for="(item, index) in diagnosaInacbg" :key="`inacbg-diagnosa-${index}`">
                      <button type="button" class="coding-list-row" :class="{ invalid: item.validcode === '0' }" :disabled="inacbgSudahFinal" @click="bukaEditorDiagnosaInacbg(index)">
                        <b>{{ item.code || '-' }}</b>
                        <span>{{ item.display || '-' }} <em v-if="item.validcode === '0'">{{ item.metadata?.message || 'Kode tidak berlaku' }}</em></span>
                        <small :class="{ primary: index === 0 }">{{ index === 0 ? 'Primer' : 'Sekunder' }}</small>
                      </button>
                      <div v-if="!inacbgSudahFinal && activeSubstitusiDiagnosaInacbg === index" class="coding-inline-editor">
                        <AutoComplete
                          v-model="substitusiDiagnosaInacbg"
                          :suggestions="searchHasilDiagnosaInacbg"
                          option-label="label"
                          placeholder="Substitusi"
                          @complete="onSearchDiagnosaInacbg"
                          @item-select="terapkanSubstitusiDiagnosaInacbg(index, $event)"
                        />
                        <button type="button" class="danger" @click="hapusDiagnosaInacbg(index)">Delete</button>
                        <button type="button" :disabled="index === 0" @click="setPrimaryDiagnosaInacbg(index)">Set Primary</button>
                      </div>
                    </template>
                  </div>
                </article>

                <article>
                  <h4>Prosedur <span>(ICD-9-CM)</span></h4>
                  <div v-if="prosedurInacbg.length === 0" class="inacbg-empty">Klik Import IDRG ke INA-CBG untuk menampilkan coding.</div>
                  <div v-else class="coding-list-table">
                    <div class="coding-list-head"><span>Kode</span><span>Nama Prosedur</span><span>Status</span></div>
                    <template v-for="(item, index) in prosedurInacbg" :key="`inacbg-prosedur-${index}`">
                      <button type="button" class="coding-list-row" :class="{ invalid: item.validcode === '0' }" :disabled="inacbgSudahFinal" @click="bukaEditorProsedurInacbg(index)">
                        <b>{{ item.code || '-' }}</b>
                        <span>{{ item.display || '-' }} <em v-if="item.validcode === '0'">{{ item.metadata?.message || 'Kode tidak berlaku' }}</em></span>
                        <small>{{ index === 0 ? 'Primer' : 'Sekunder' }}</small>
                      </button>
                      <div v-if="!inacbgSudahFinal && activeSubstitusiProsedurInacbg === index" class="coding-inline-editor">
                        <AutoComplete
                          v-model="substitusiProsedurInacbg"
                          :suggestions="searchHasilProsedurInacbg"
                          option-label="label"
                          placeholder="Substitusi"
                          @complete="onSearchProsedurInacbg"
                          @item-select="terapkanSubstitusiProsedurInacbg(index, $event)"
                        />
                        <button type="button" class="danger" @click="hapusProsedurInacbg(index)">Delete</button>
                      </div>
                    </template>
                  </div>
                </article>
              </div>
            </section>

            <section class="inacbg-result-card" :class="{ final: inacbgGroupingFinal }">
              <header>
                <h4>Hasil Grouping INA-CBG{{ inacbgGroupingFinal ? ' - Final' : '' }}</h4>
                <div class="inacbg-process-actions">
                  <template v-if="klaimSudahFinal">
                    <button type="button" class="secondary" :disabled="!pasienTerpilih.no_sep || Boolean(loadingProses)" @click="jalankanProses('reedit_klaim')">
                      <LoaderCircle v-if="loadingProses === 'reedit_klaim'" class="spin" :size="15" />
                      <Edit3 v-else :size="15" />
                      Edit Ulang Klaim
                    </button>
                    <button type="button" class="teal" :disabled="!pasienTerpilih.no_sep || Boolean(loadingProses)" @click="jalankanProses('send_claim_individual')">
                      <LoaderCircle v-if="loadingProses === 'send_claim_individual'" class="spin" :size="15" />
                      <Send v-else :size="15" />
                      {{ klaimSudahTerkirim ? 'Kirim Ulang Klaim' : 'Kirim Klaim' }}
                    </button>
                  </template>
                  <template v-else-if="inacbgBisaEditUlang">
                    <button type="button" class="secondary" :disabled="!pasienTerpilih.no_sep || Boolean(loadingProses)" @click="jalankanProses('reedit_inacbg')">
                      <LoaderCircle v-if="loadingProses === 'reedit_inacbg'" class="spin" :size="15" />
                      <Edit3 v-else :size="15" />
                      Edit Ulang INA-CBG
                    </button>
                    <button type="button" class="teal" :disabled="!pasienTerpilih.no_sep || Boolean(loadingProses)" @click="jalankanProses('final_klaim')">
                      <LoaderCircle v-if="loadingProses === 'final_klaim'" class="spin" :size="15" />
                      <CheckCircle2 v-else :size="15" />
                      Final Klaim
                    </button>
                  </template>
                  <template v-else>
                    <button type="button" class="purple" :disabled="!pasienTerpilih.no_sep || loadingProses" @click="jalankanProses('grouping_inacbg')">
                      <LoaderCircle v-if="loadingProses === 'grouping_inacbg'" class="spin" :size="15" />
                      <Layers v-else :size="15" />
                      Grouping
                    </button>
                    <button type="button" class="teal" :disabled="!pasienTerpilih.no_sep || loadingProses || !responseInacbgTerakhir" @click="jalankanProses('final_inacbg')">
                      <LoaderCircle v-if="loadingProses === 'final_inacbg'" class="spin" :size="15" />
                      <CheckCircle2 v-else :size="15" />
                      Final INA-CBG
                    </button>
                  </template>
                </div>
              </header>

              <div v-if="responseInacbgTerakhir" class="inacbg-grouping-sheet" :class="{ final: inacbgGroupingFinal }">
                <div class="inacbg-grouping-table-wrap">
                  <table>
                    <tbody>
                      <tr>
                        <th>Info</th>
                        <td colspan="4">{{ hasilGroupingInacbgTabel.info }}</td>
                      </tr>
                      <tr class="main-row">
                        <th>Group</th>
                        <td>{{ hasilGroupingInacbgTabel.deskripsi }}</td>
                        <td class="text-right">{{ hasilGroupingInacbgTabel.kode }}</td>
                        <td class="text-right">Rp</td>
                        <td class="text-right blue-value">{{ rupiah(hasilGroupingInacbgTabel.tarifUtama) }}</td>
                      </tr>
                      <tr>
                        <th>Sub Acute</th>
                        <td>-</td>
                        <td class="text-right">-</td>
                        <td class="text-right">Rp</td>
                        <td class="text-right">{{ rupiah(hasilGroupingInacbgTabel.subAcute) }}</td>
                      </tr>
                      <tr>
                        <th>Chronic</th>
                        <td>-</td>
                        <td class="text-right">-</td>
                        <td class="text-right">Rp</td>
                        <td class="text-right">{{ rupiah(hasilGroupingInacbgTabel.chronic) }}</td>
                      </tr>
                      <tr v-for="item in hasilGroupingInacbgTabel.specialRows" :key="item.type">
                        <th>{{ item.type }}</th>
                        <td>{{ item.description }}</td>
                        <td class="text-right">{{ item.code }}</td>
                        <td class="text-right">Rp</td>
                        <td class="text-right">{{ rupiah(item.tariff) }}</td>
                      </tr>
                      <tr class="total-row">
                        <th colspan="3">Total Klaim</th>
                        <td class="text-right">Rp</td>
                        <td class="text-right">{{ rupiah(hasilGroupingInacbgTabel.total) }}</td>
                      </tr>
                      <tr>
                        <th>Status</th>
                        <td colspan="4">{{ hasilGroupingInacbgTabel.status }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
              <div v-else class="inacbg-empty-result">Hasil grouping INA-CBG belum ada.</div>
            </section>
          </section>

        </div>
      </template>

      <section v-else class="idrg-card idrg-coming-soon">
        <component :is="menuSidebar.find((item) => item.id === menuSidebarAktif)?.icon || FileText" :size="34" />
        <span>{{ menuSidebar.find((item) => item.id === menuSidebarAktif)?.label }}</span>
        <h2>Tahap berikutnya</h2>
        <p>Bagian ini akan disambungkan setelah alur Pengajuan Klaim IDRG stabil dulu, pak.</p>
      </section>
    </main>
  </section>
</template>
