<script setup>
import {
  Activity,
  Bed,
  ClipboardList,
  FileSpreadsheet,
  FlaskConical,
  HeartPulse,
  Home,
  LayoutDashboard,
  LogIn,
  LogOut,
  Microscope,
  Pill,
  Stethoscope,
  UsersRound,
} from '@lucide/vue'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import DashboardLayout from '../Components/Layout/DashboardLayout.vue'
import PatientWorkspace from '../Components/Patients/PatientWorkspace.vue'
import BerandaTab from '../Components/Tabs/BerandaTab.vue'
import IdrgPage from './Eklaim/IdrgPage.vue'
import ModulePlaceholder from '../Components/Tabs/ModulePlaceholder.vue'
import ModuleTab from '../Components/Tabs/ModuleTab.vue'
import UserManagementTab from '../Components/Tabs/UserManagementTab.vue'
import { dashboardData, userManagementData } from '../lib/faisal/api'
import { useNotifikasi } from '../lib/shared/useNotifikasi'

const props = defineProps({
  user: { type: Object, default: null },
  token: { type: String, default: '' },
  loading: Boolean,
})
const emit = defineEmits(['login', 'logout'])

const isDark = ref(localStorage.getItem('simrs_theme') !== 'light')
const currentTab = ref('Menu')
const menuOpen = ref(false)
const menuSearch = ref('')
const selectedPatient = ref(null)
const selectedPatientModule = ref('')
const now = ref(new Date())
const dashboardLoading = ref(true)
const dashboardError = ref('')
const userManagementLoading = ref(false)
const userManagementError = ref('')
const notifikasi = useNotifikasi()
let urutanRequestDashboard = 0
const hariIni = new Date().toLocaleDateString('en-CA')

function terapkanClassDarkMode() {
  document.body.classList.toggle('sirava-dark', isDark.value)
}
const filterKosong = (belumPulang = false) => ({
  date_from: hariIni,
  date_to: hariIni,
  status: '',
  status_bayar: '',
  dokter: '',
  poly: '',
  search: '',
  belum_pulang: belumPulang,
})
const patientFilters = ref({
  'Rawat Jalan': filterKosong(),
  'IGD/UGD': filterKosong(),
  'Rawat Inap': filterKosong(true),
})
const paginationKosong = () => ({ halaman: 1, batas: 100, total: 0, total_halaman: 0 })
const patientPagination = ref({
  Registrasi: paginationKosong(),
  'Rawat Jalan': paginationKosong(),
  'IGD/UGD': paginationKosong(),
  'Rawat Inap': paginationKosong(),
})
const dashboard = ref({
  koneksi_database: { terhubung: false, latensi_ms: null },
  ringkasan: { jumlah_registrasi: 0, jumlah_rawat_jalan: 0, jumlah_igd: 0, jumlah_rawat_inap: 0 },
  registrasi: [],
  rawat_jalan: [],
  igd: [],
  rawat_inap: [],
  paginasi: {
    registrasi: paginationKosong(),
    rawat_jalan: paginationKosong(),
    igd: paginationKosong(),
    rawat_inap: paginationKosong(),
  },
  poliklinik: [],
  dokter: [],
  pilihan_status: { periksa: [], rawat_inap: [], status_bayar: [] },
  menu_workspace_pasien: [],
})
const userManagement = ref({
  ringkasan: { jumlah_pengguna: 0, jumlah_aktif: 0, jumlah_admin: 0 },
  pengguna: [],
  permission: [],
})

const isAuthenticated = computed(() => Boolean(props.user && props.token))

const clockTimer = window.setInterval(() => {
  now.value = new Date()
}, 1000)
onBeforeUnmount(() => window.clearInterval(clockTimer))

const formattedDate = computed(() => new Intl.DateTimeFormat('id-ID', {
  weekday: 'long',
  day: '2-digit',
  month: 'long',
  year: 'numeric',
  timeZone: 'Asia/Makassar',
}).format(now.value))

const formattedTime = computed(() => new Intl.DateTimeFormat('id-ID', {
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
  hour12: false,
  timeZone: 'Asia/Makassar',
}).format(now.value).replaceAll('.', ':'))

const moduleRibbonMenus = [
  { label: 'Registrasi', icon: ClipboardList, tone: 'blue' },
  { label: 'IGD/UGD', icon: HeartPulse, tone: 'rose' },
  { label: 'Laborat', icon: FlaskConical, tone: 'amber' },
  { label: 'Radiologi', icon: Microscope, tone: 'violet' },
  { label: 'Farmasi', icon: Pill, tone: 'teal' },
  { label: 'Rawat Inap', icon: Bed, tone: 'indigo' },
  { label: 'Rawat Jalan', icon: Stethoscope, tone: 'cyan' },
]

const ribbonMenus = computed(() => [
  { label: currentTab.value === 'Menu' ? 'Menu' : 'Beranda', icon: Home, tone: 'slate' },
  ...moduleRibbonMenus,
  isAuthenticated.value
    ? { label: 'Logout', icon: LogOut, tone: 'red', authAction: true }
    : { label: 'Login', icon: LogIn, tone: 'teal', authAction: true },
])

const dashboardMenus = [
  { label: 'Registrasi', description: 'Pendaftaran dan daftar kunjungan pasien hari ini.', icon: ClipboardList, tone: 'blue' },
  { label: 'IGD/UGD', description: 'Daftar pasien instalasi gawat darurat.', icon: HeartPulse, tone: 'rose' },
  { label: 'Laborat', description: 'Pemeriksaan dan permintaan laboratorium.', icon: FlaskConical, tone: 'amber' },
  { label: 'Radiologi', description: 'Pemeriksaan dan permintaan radiologi.', icon: Microscope, tone: 'violet' },
  { label: 'Farmasi', description: 'Resep, obat, dan pelayanan farmasi.', icon: Pill, tone: 'teal' },
  { label: 'Rawat Inap', description: 'Kamar inap dan daftar pasien rawat inap.', icon: Bed, tone: 'indigo' },
  { label: 'Rawat Jalan', description: 'Daftar pasien dan pelayanan rawat jalan.', icon: Stethoscope, tone: 'cyan' },
  { label: 'IDRG', description: 'Bridging klaim BPJS E-Klaim iDRG / INA-CBG.', icon: FileSpreadsheet, tone: 'emerald' },
  { label: 'Kelola Menu', description: 'Pengaturan menu navigasi dan hak akses.', icon: LayoutDashboard, tone: 'slate' },
  { label: 'User Management', description: 'Kelola user, status akun, dan permission aplikasi.', icon: UsersRound, tone: 'slate' },
]

const filteredMenus = computed(() => {
  const keyword = menuSearch.value.trim().toLowerCase()
  if (!keyword) return dashboardMenus
  return dashboardMenus.filter((menu) => `${menu.label} ${menu.description}`.toLowerCase().includes(keyword))
})

const quickStats = computed(() => [
  { label: 'Registrasi Hari Ini', value: dashboard.value.ringkasan.jumlah_registrasi, icon: ClipboardList, tone: 'blue', tab: 'Registrasi' },
  { label: 'Antrian IGD', value: dashboard.value.ringkasan.jumlah_igd, icon: Activity, tone: 'rose', tab: 'IGD/UGD' },
  { label: 'Rawat Jalan', value: dashboard.value.ringkasan.jumlah_rawat_jalan, icon: Stethoscope, tone: 'cyan', tab: 'Rawat Jalan' },
  { label: 'Pasien Rawat Inap', value: dashboard.value.ringkasan.jumlah_rawat_inap, icon: Bed, tone: 'teal', tab: 'Rawat Inap' },
])

const serviceStatuses = computed(() => [
  { label: 'Koneksi API', value: 'Online', tone: 'online' },
  {
    label: 'Koneksi SIMRS Lama',
    value: !isAuthenticated.value ? 'Perlu login' : dashboardLoading.value ? 'Memeriksa...' : dashboardError.value ? 'Offline' : `Online - ${dashboard.value.koneksi_database.latensi_ms} ms`,
    tone: !isAuthenticated.value || dashboardLoading.value ? 'pending' : dashboardError.value ? 'offline' : 'online',
  },
  { label: 'Koneksi E-Klaim', value: 'Belum tersedia', tone: 'pending' },
])

const patientRows = computed(() => ({
  Registrasi: dashboard.value.registrasi,
  'Rawat Jalan': dashboard.value.rawat_jalan,
  'IGD/UGD': dashboard.value.igd,
  'Rawat Inap': dashboard.value.rawat_inap,
}[currentTab.value] ?? null))

const kunciDataPasien = {
  Registrasi: 'registrasi',
  'Rawat Jalan': 'rawat_jalan',
  'IGD/UGD': 'igd',
  'Rawat Inap': 'rawat_inap',
}
const kunciPaginasiPasien = {
  Registrasi: 'registrasi',
  'Rawat Jalan': 'rawat_jalan',
  'IGD/UGD': 'igd',
  'Rawat Inap': 'rawat_inap',
}
const menuDataPasien = Object.keys(kunciDataPasien)

const filteredPatientRows = computed(() => patientRows.value ?? [])
const filterAktif = computed(() => patientFilters.value[currentTab.value] ?? filterKosong())
const paginasiAktif = computed(() => patientPagination.value[currentTab.value] ?? paginationKosong())
const daftarPoliklinik = computed(() => dashboard.value.poliklinik ?? [])
const daftarDokter = computed(() => dashboard.value.dokter ?? [])
const pilihanStatus = computed(() => dashboard.value.pilihan_status ?? { periksa: [], rawat_inap: [], status_bayar: [] })

const prefixFilterPasien = {
  'Rawat Jalan': 'rj',
  'IGD/UGD': 'igd',
  'Rawat Inap': 'ri',
}

function parameterFilterPasien(filterPerModul) {
  const parameter = {}

  Object.entries(filterPerModul).forEach(([namaModul, filter]) => {
    const prefix = prefixFilterPasien[namaModul]
    if (!prefix) return

    parameter[`${prefix}_date_from`] = filter.date_from || ''
    parameter[`${prefix}_date_to`] = filter.date_to || filter.date_from || ''
    parameter[`${prefix}_status`] = filter.status || ''
    parameter[`${prefix}_status_bayar`] = filter.status_bayar || ''
    parameter[`${prefix}_dokter`] = filter.dokter || ''
    parameter[`${prefix}_search`] = filter.search || ''

    const paginasi = patientPagination.value[namaModul] ?? paginationKosong()
    parameter[`${prefix}_page`] = paginasi.halaman || 1
    parameter[`${prefix}_limit`] = paginasi.batas || 100

    if (namaModul === 'Rawat Jalan') {
      parameter[`${prefix}_poly`] = filter.poly || ''
    }

    if (namaModul === 'Rawat Inap') {
      parameter[`${prefix}_belum_pulang`] = filter.belum_pulang ? 1 : 0
    }
  })

  if (currentTab.value !== 'Menu') parameter.tab = currentTab.value
  return parameter
}

function isMenuDisabled(label) {
  if (isAuthenticated.value) return false
  return !['Menu', 'Beranda', 'Login', 'Logout'].includes(label)
}

async function loadDashboard(filterPerModul = patientFilters.value) {
  if (!isAuthenticated.value) {
    dashboardLoading.value = false
    dashboardError.value = 'Silakan login untuk membaca data pasien dari database SIMRS lama.'
    return
  }

  const requestAktif = urutanRequestDashboard + 1
  urutanRequestDashboard = requestAktif
  dashboardLoading.value = true
  dashboardError.value = ''
  try {
    const dataBeranda = await dashboardData(props.token, parameterFilterPasien(filterPerModul))
    if (requestAktif !== urutanRequestDashboard) return

    dashboard.value = dataBeranda
    sinkronkanPaginasi(dataBeranda.paginasi)
    tampilkanToastDataKosong(currentTab.value, dashboard.value)
  } catch (error) {
    if (requestAktif !== urutanRequestDashboard) return

    dashboardError.value = error.message
    notifikasi.gagal(error.message || 'Gagal memuat data beranda.')
  } finally {
    if (requestAktif === urutanRequestDashboard) {
      dashboardLoading.value = false
    }
  }
}

function sinkronkanPaginasi(paginasi = {}) {
  patientPagination.value = {
    Registrasi: paginasi.registrasi ?? patientPagination.value.Registrasi ?? paginationKosong(),
    'Rawat Jalan': paginasi.rawat_jalan ?? patientPagination.value['Rawat Jalan'] ?? paginationKosong(),
    'IGD/UGD': paginasi.igd ?? patientPagination.value['IGD/UGD'] ?? paginationKosong(),
    'Rawat Inap': paginasi.rawat_inap ?? patientPagination.value['Rawat Inap'] ?? paginationKosong(),
  }
}

async function loadUserManagement() {
  if (!isAuthenticated.value) return

  userManagementLoading.value = true
  userManagementError.value = ''
  try {
    userManagement.value = await userManagementData(props.token)
  } catch (error) {
    userManagementError.value = error.message
    notifikasi.gagal(error.message || 'Gagal memuat data user management.')
  } finally {
    userManagementLoading.value = false
  }
}

onMounted(() => {
  terapkanClassDarkMode()
  loadDashboard()
})
watch(() => props.token, loadDashboard)
watch(isAuthenticated, (loggedIn) => {
  if (!loggedIn) {
    currentTab.value = 'Menu'
    selectedPatient.value = null
    selectedPatientModule.value = ''
  }
})

function toggleTheme() {
  isDark.value = !isDark.value
  localStorage.setItem('simrs_theme', isDark.value ? 'dark' : 'light')
  terapkanClassDarkMode()
}

function selectMenu(label) {
  if (isMenuDisabled(label)) {
    notifikasi.peringatan('Silakan login dulu untuk membuka menu ini.')
    return
  }

  if (label === 'Login') {
    emit('login')
    return
  }
  if (label === 'Menu') {
    openMenu()
    return
  }
  if (label === 'Beranda') {
    currentTab.value = 'Menu'
    selectedPatient.value = null
    selectedPatientModule.value = ''
    menuOpen.value = false
    return
  }
  if (label === 'Logout') {
    notifikasi.info('Mengakhiri sesi SIRAVA...')
    emit('logout')
    return
  }
  selectedPatient.value = null
  selectedPatientModule.value = ''
  currentTab.value = label
  menuOpen.value = false

  if (menuDataPasien.includes(label)) {
    loadDashboard(patientFilters.value)
    return
  }

  if (label === 'User Management') {
    loadUserManagement()
    return
  }

  if (!dashboardLoading.value) tampilkanToastDataKosong(label, dashboard.value)
}

function selectPatient(patient) {
  if (!patient || !['Rawat Jalan', 'IGD/UGD', 'Rawat Inap'].includes(currentTab.value)) return
  selectedPatient.value = patient
  selectedPatientModule.value = currentTab.value
}

function closePatientWorkspace() {
  selectedPatient.value = null
  selectedPatientModule.value = ''
}

function openMenu() {
  menuSearch.value = ''
  menuOpen.value = true
}

function applyPatientFilters(filter) {
  if (!prefixFilterPasien[currentTab.value]) return
  if (dashboardLoading.value) return

  patientPagination.value = {
    ...patientPagination.value,
    [currentTab.value]: { ...paginasiAktif.value, halaman: 1 },
  }

  const filterPerModul = {
    ...patientFilters.value,
    [currentTab.value]: filter,
  }
  patientFilters.value = filterPerModul
  loadDashboard(filterPerModul)
}

function applyPatientPage(paginasi) {
  if (!prefixFilterPasien[currentTab.value]) return
  if (dashboardLoading.value) return

  patientPagination.value = {
    ...patientPagination.value,
    [currentTab.value]: {
      ...paginasiAktif.value,
      halaman: paginasi.halaman,
      batas: paginasi.batas,
    },
  }
  loadDashboard(patientFilters.value)
}

function tampilkanToastDataKosong(namaTab, dataBeranda) {
  if (dashboardError.value) return

  const kunci = kunciDataPasien[namaTab]
  if (!kunci) return

  const daftarData = dataBeranda?.[kunci] ?? []
  if (daftarData.length > 0) return

  notifikasi.peringatan(`Data ${namaTab} tidak ditemukan untuk filter yang dipilih.`, 'Data Kosong')
}
</script>

<template>
  <DashboardLayout
    v-model:menu-open="menuOpen"
    v-model:menu-search="menuSearch"
    :is-dark="isDark"
    :formatted-date="formattedDate"
    :formatted-time="formattedTime"
    :user="user"
    :current-tab="currentTab"
    :ribbon-menus="ribbonMenus"
    :filtered-menus="filteredMenus"
    :is-menu-disabled="isMenuDisabled"
    @toggle-theme="toggleTheme"
    @select-menu="selectMenu"
  >
    <BerandaTab
      v-if="currentTab === 'Menu'"
      :dashboard-loading="dashboardLoading"
      :dashboard-error="dashboardError"
      :quick-stats="quickStats"
      :service-statuses="serviceStatuses"
      :is-authenticated="isAuthenticated"
      @select-tab="selectMenu"
      @open-menu="openMenu"
    />

    <PatientWorkspace
      v-else-if="selectedPatient"
      :module-name="selectedPatientModule"
      :patient="selectedPatient"
      :menus="dashboard.menu_workspace_pasien"
      @back="closePatientWorkspace"
    />

    <ModuleTab
      v-else-if="patientRows !== null"
      :current-tab="currentTab"
      :patient-rows="patientRows"
      :filtered-patient-rows="filteredPatientRows"
      :filter="filterAktif"
      :daftar-poliklinik="daftarPoliklinik"
      :daftar-dokter="daftarDokter"
      :daftar-status-periksa="pilihanStatus.periksa"
      :daftar-status-rawat-inap="pilihanStatus.rawat_inap"
      :daftar-status-bayar="pilihanStatus.status_bayar"
      :pagination="paginasiAktif"
      :is-dark="isDark"
      :dashboard-error="dashboardError"
      :dashboard-loading="dashboardLoading"
      @apply-filters="applyPatientFilters"
      @page-change="applyPatientPage"
      @select-patient="selectPatient"
    />

    <UserManagementTab
      v-else-if="currentTab === 'User Management'"
      :data="userManagement"
      :token="token"
      :loading="userManagementLoading"
      :error="userManagementError"
      @saved="loadUserManagement"
    />

    <IdrgPage
      v-else-if="currentTab === 'IDRG'"
      :token="token"
      :user="user"
    />

    <ModulePlaceholder
      v-else
      :current-tab="currentTab"
      :ribbon-menus="ribbonMenus"
      @back="currentTab = 'Menu'"
    />
  </DashboardLayout>
</template>
