// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { Activity, Bed, BedDouble, ClipboardList, FileSpreadsheet, FlaskConical, HeartPulse, Home, LayoutDashboard, LogIn, LogOut, Microscope, MessageCircle, Pill, ScrollText, Stethoscope, UsersRound, Wind } from "@lucide/vue"
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue"
import { dashboardData, koneksiBPJS, koneksiEKlaim, userManagementData } from "../../lib/faisal/api"
import { useNotifikasi } from "../../lib/shared/useNotifikasi"

export function useHome(props, emit) {
  const PREFIX_NAVIGASI = 'sirapi.navigation.'
  const isDark = ref(localStorage.getItem('simrs_theme') !== 'light')
  const currentTab = ref('Menu')
  const menuOpen = ref(false)
  const menuSearch = ref('')
  const selectedPatient = ref(null)
  const selectedPatientModule = ref('')
  const now = ref(new Date())
  const dashboardLoading = ref(true)
  const dashboardError = ref('')
  const koneksiEKlaimStatus = ref({ memeriksa: false, dikonfigurasi: false, terhubung: false, latensi_ms: null })
  const koneksiBPJSStatus = ref({ memeriksa: false, dikonfigurasi: false, terhubung: false, latensi_ms: null })
  const userManagementLoading = ref(false)
  const userManagementError = ref('')
  const notifikasi = useNotifikasi()
  let urutanRequestDashboard = 0
  const hariIni = new Date().toLocaleDateString('en-CA')
  
  function terapkanClassDarkMode() {
    document.body.classList.toggle('sirapi-dark', isDark.value)
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
  const paginationKosong = () => ({ halaman: 1, batas: 10, total: 0, total_halaman: 0 })
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
    sidebar_pasien: [],
  })
  const userManagement = ref({
    ringkasan: { jumlah_pengguna: 0, jumlah_aktif: 0, jumlah_admin: 0 },
    pengguna: [],
    permission: [],
  })
  
  const isAuthenticated = computed(() => Boolean(props.user && props.token))
  
  function kunciNavigasi() {
    const identitas = props.user?.id || props.user?.username
    return identitas ? `${PREFIX_NAVIGASI}${identitas}` : ''
  }
  
  function simpanNavigasi() {
    const kunci = kunciNavigasi()
    if (!kunci || !isAuthenticated.value) return
  
    localStorage.setItem(kunci, JSON.stringify({
      menu: currentTab.value,
      modul_pasien: selectedPatientModule.value,
      pasien: selectedPatient.value,
    }))
  }
  
  function pulihkanNavigasi() {
    const kunci = kunciNavigasi()
    if (!kunci || !isAuthenticated.value) return
  
    try {
      const tersimpan = JSON.parse(localStorage.getItem(kunci) || '{}')
      const menu = typeof tersimpan.menu === 'string' ? tersimpan.menu : 'Menu'
      const modulPasienValid = ['Rawat Jalan', 'IGD/UGD', 'Rawat Inap'].includes(tersimpan.modul_pasien)
      const pasienValid = tersimpan.pasien && typeof tersimpan.pasien === 'object' && tersimpan.pasien.no_rawat
  
      currentTab.value = ['Login', 'Logout'].includes(menu) ? 'Menu' : menu
      selectedPatientModule.value = modulPasienValid && pasienValid ? tersimpan.modul_pasien : ''
      selectedPatient.value = modulPasienValid && pasienValid ? tersimpan.pasien : null
    } catch {
      localStorage.removeItem(kunci)
    }
  }
  
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
    { label: 'Registrasi', description: 'Pendaftaran dan daftar kunjungan pasien hari ini.', icon: ClipboardList, tone: 'blue', category: 'Layanan Medis' },
    { label: 'IGD/UGD', description: 'Daftar pasien instalasi gawat darurat.', icon: HeartPulse, tone: 'rose', category: 'Layanan Medis' },
    { label: 'Laborat', description: 'Pemeriksaan dan permintaan laboratorium.', icon: FlaskConical, tone: 'amber', category: 'Penunjang Medis' },
    { label: 'Radiologi', description: 'Pemeriksaan dan permintaan radiologi.', icon: Microscope, tone: 'violet', category: 'Penunjang Medis' },
    { label: 'Farmasi', description: 'Resep, obat, dan pelayanan farmasi.', icon: Pill, tone: 'teal', category: 'Penunjang Medis' },
    { label: 'Rawat Inap', description: 'Kamar inap dan daftar pasien rawat inap.', icon: Bed, tone: 'indigo', category: 'Layanan Medis' },
    { label: 'Rawat Jalan', description: 'Daftar pasien dan pelayanan rawat jalan.', icon: Stethoscope, tone: 'cyan', category: 'Layanan Medis' },
    { label: 'IDRG', description: 'Bridging klaim BPJS E-Klaim iDRG / INA-CBG.', icon: FileSpreadsheet, tone: 'emerald', category: 'Integrasi & Klaim' },
    { label: 'Monitoring Klaim BPJS', description: 'Monitoring data klaim VClaim berdasarkan periode.', icon: FileSpreadsheet, tone: 'blue', category: 'Integrasi & Klaim' },
    { label: 'WhatsApp Gateway', description: 'Kelola perangkat dan kirim pesan WhatsApp dari SIRAPI.', icon: MessageCircle, tone: 'teal', category: 'Integrasi & Klaim' },
    { label: 'Kelola Menu', description: 'Pengaturan sidebar pasien dan hak aksesnya.', icon: LayoutDashboard, tone: 'slate', category: 'Sistem' },
    { label: 'Master Ventilator', description: 'Kelola perangkat ventilator dan jadwal pemeliharaannya.', icon: Wind, tone: 'cyan', category: 'Sistem' },
    { label: 'User Management', description: 'Kelola user, status akun, dan permission aplikasi.', icon: UsersRound, tone: 'slate', category: 'Sistem' },
    { label: 'Log Aktivitas', description: 'Audit login, akses data, perubahan, dan kegagalan proses.', icon: ScrollText, tone: 'slate', category: 'Sistem' },
    { label: 'Laporan Kunjungan Ralan', description: 'Laporan kunjungan rawat jalan berdasarkan periode dan pelayanan.', icon: FileSpreadsheet, tone: 'emerald', category: 'Laporan' },
    { label: 'Laporan Kunjungan Ranap', description: 'Laporan pasien masuk, pulang, dan kunjungan berulang rawat inap.', icon: FileSpreadsheet, tone: 'indigo', category: 'Laporan' },
    { label: 'Laporan 10 Penyakit', description: 'Rekap 10 penyakit terbanyak berdasarkan periode pelayanan.', icon: FileSpreadsheet, tone: 'rose', category: 'Laporan' },
    { label: 'Penggunaan Bed & Frekuensi', description: 'Frekuensi rata-rata penggunaan bed per bangsal berdasarkan pasien keluar.', icon: Bed, tone: 'indigo', category: 'BED' },
    { label: 'Laporan BOR, LOS & TOI', description: 'Indikator pemanfaatan tempat tidur rawat inap per bulan.', icon: Bed, tone: 'teal', category: 'BED' },
    { label: 'Monitoring Bed', description: 'Laporan ketersediaan dan penggunaan hari bed.', icon: BedDouble, tone: 'cyan', category: 'BED' },
  ]
  
  const filteredMenus = computed(() => {
    const keyword = menuSearch.value.trim().toLowerCase()
    if (!keyword) return dashboardMenus
    return dashboardMenus.filter((menu) =>
      `${menu.label} ${menu.description} ${menu.category}`.toLowerCase().includes(keyword),
    )
  })
  
  const quickStats = computed(() => [
    { label: 'Registrasi Hari Ini', value: dashboard.value.ringkasan.jumlah_registrasi, icon: ClipboardList, tone: 'blue', tab: 'Registrasi' },
    { label: 'Antrian IGD', value: dashboard.value.ringkasan.jumlah_igd, icon: Activity, tone: 'rose', tab: 'IGD/UGD' },
    { label: 'Rawat Jalan', value: dashboard.value.ringkasan.jumlah_rawat_jalan, icon: Stethoscope, tone: 'cyan', tab: 'Rawat Jalan' },
    { label: 'Pasien Rawat Inap', value: dashboard.value.ringkasan.jumlah_rawat_inap, icon: Bed, tone: 'teal', tab: 'Rawat Inap' },
  ])
  
  const serviceStatuses = computed(() => [
    { label: 'Web Servis', value: 'Online', tone: 'online' },
    {
      label: 'Simrs Khanza',
      value: !isAuthenticated.value ? 'Perlu login' : dashboardLoading.value ? 'Memeriksa...' : dashboardError.value ? 'Offline' : `Online - ${dashboard.value.koneksi_database.latensi_ms} ms`,
      tone: !isAuthenticated.value || dashboardLoading.value ? 'pending' : dashboardError.value ? 'offline' : 'online',
    },
    {
      label: 'E-Klaim',
      value: !isAuthenticated.value
        ? 'Perlu login'
        : koneksiEKlaimStatus.value.memeriksa
          ? 'Memeriksa...'
          : !koneksiEKlaimStatus.value.dikonfigurasi
            ? 'Belum dikonfigurasi'
            : koneksiEKlaimStatus.value.terhubung
              ? `Online - ${koneksiEKlaimStatus.value.latensi_ms} ms`
              : 'Offline',
      tone: !isAuthenticated.value || koneksiEKlaimStatus.value.memeriksa || !koneksiEKlaimStatus.value.dikonfigurasi
        ? 'pending'
        : koneksiEKlaimStatus.value.terhubung ? 'online' : 'offline',
    },
    {
      label: 'BPJS VClaim',
      value: !isAuthenticated.value
        ? 'Perlu login'
        : koneksiBPJSStatus.value.memeriksa
          ? 'Memeriksa...'
          : !koneksiBPJSStatus.value.dikonfigurasi
            ? 'Belum dikonfigurasi'
            : koneksiBPJSStatus.value.terhubung
              ? `Online - ${koneksiBPJSStatus.value.latensi_ms} ms`
              : 'Offline',
      tone: !isAuthenticated.value || koneksiBPJSStatus.value.memeriksa || !koneksiBPJSStatus.value.dikonfigurasi
        ? 'pending'
        : koneksiBPJSStatus.value.terhubung ? 'online' : 'offline',
    },
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
      parameter[`${prefix}_limit`] = paginasi.batas || 10
  
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
    if (['Menu', 'Beranda', 'Login', 'Logout'].includes(label)) return false
    if (!isAuthenticated.value) return true

    const permissions = Array.isArray(props.user?.permissions) ? props.user.permissions : []
    if (permissions.includes('*')) return false

    const aksesMenu = {
      Registrasi: ['registrasi'],
      'IGD/UGD': ['igd'],
      Laborat: ['periksa_lab', 'permintaan_lab'],
      Radiologi: ['periksa_radiologi', 'permintaan_radiologi'],
      Farmasi: ['obat', 'beri_obat', 'resep_obat'],
      'Rawat Inap': ['kamar_inap', 'daftar_pasien_ranap'],
      'Rawat Jalan': ['registrasi', 'tindakan_ralan', 'billing_ralan'],
      IDRG: ['eklaim'],
      'WhatsApp Gateway': ['whatsapp_gateway'],
      'Kelola Menu': ['kelola_menu'],
      'Master Ventilator': ['master_ventilator'],
      'User Management': ['*'],
      'Log Aktivitas': ['sistem.audit_log'],
      'Laporan Kunjungan Ralan': ['laporan_kunjungan_ralan'],
      'Laporan Kunjungan Ranap': ['laporan_kunjungan_ranap'],
      'Laporan 10 Penyakit': ['laporan_10_penyakit'],
      'Penggunaan Bed & Frekuensi': ['laporan_penggunaan_bed'],
      'Laporan BOR, LOS & TOI': ['laporan_bor_los_toi'],
      'Monitoring Bed': ['monitoring_bed'],
    }
    const dibutuhkan = aksesMenu[label] || []
    return dibutuhkan.length === 0 || !dibutuhkan.some((kode) => permissions.includes(kode))
  }
  
  async function loadDashboard(filterPerModul = patientFilters.value) {
    if (!isAuthenticated.value) {
      dashboardLoading.value = false
      dashboardError.value = 'Silakan login untuk membaca data pasien dari database SIMRS KHANZA.'
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
  
  async function loadKoneksiEKlaim() {
    if (!isAuthenticated.value) {
      koneksiEKlaimStatus.value = { memeriksa: false, dikonfigurasi: false, terhubung: false, latensi_ms: null }
      return
    }
  
    koneksiEKlaimStatus.value.memeriksa = true
    try {
      const hasil = await koneksiEKlaim(props.token)
      koneksiEKlaimStatus.value = { memeriksa: false, ...hasil }
    } catch {
      koneksiEKlaimStatus.value = {
        ...koneksiEKlaimStatus.value,
        memeriksa: false,
        dikonfigurasi: true,
        terhubung: false,
        latensi_ms: null,
      }
    }
  }
  
  async function loadKoneksiBPJS() {
    if (!isAuthenticated.value) {
      koneksiBPJSStatus.value = { memeriksa: false, dikonfigurasi: false, terhubung: false, latensi_ms: null }
      return
    }
  
    koneksiBPJSStatus.value.memeriksa = true
    try {
      const hasil = await koneksiBPJS(props.token)
      koneksiBPJSStatus.value = { memeriksa: false, ...hasil }
    } catch {
      koneksiBPJSStatus.value = {
        ...koneksiBPJSStatus.value,
        memeriksa: false,
        dikonfigurasi: true,
        terhubung: false,
        latensi_ms: null,
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
    pulihkanNavigasi()
    loadDashboard()
    loadKoneksiEKlaim()
    loadKoneksiBPJS()
  })
  watch(() => props.token, () => {
    pulihkanNavigasi()
    loadDashboard()
    loadKoneksiEKlaim()
    loadKoneksiBPJS()
  })
  watch(isAuthenticated, (loggedIn) => {
    if (!loggedIn) {
      currentTab.value = 'Menu'
      selectedPatient.value = null
      selectedPatientModule.value = ''
    }
  })
  watch([currentTab, selectedPatientModule, selectedPatient], simpanNavigasi, { deep: true })
  
  function toggleTheme() {
    isDark.value = !isDark.value
    localStorage.setItem('simrs_theme', isDark.value ? 'dark' : 'light')
    terapkanClassDarkMode()
  }
  
  function selectMenu(label) {
    if (isMenuDisabled(label)) {
      notifikasi.peringatan(isAuthenticated.value ? 'Anda tidak memiliki akses ke modul ini.' : 'Silakan login dulu untuk membuka menu ini.')
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
      notifikasi.info('Mengakhiri sesi SIRAPI...')
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
  
  function closePatientSidebar() {
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
  return {
    isDark,
    currentTab,
    menuOpen,
    menuSearch,
    selectedPatient,
    selectedPatientModule,
    dashboardLoading,
    dashboardError,
    userManagementLoading,
    userManagementError,
    dashboard,
    userManagement,
    isAuthenticated,
    formattedDate,
    formattedTime,
    ribbonMenus,
    filteredMenus,
    quickStats,
    serviceStatuses,
    patientRows,
    filteredPatientRows,
    filterAktif,
    paginasiAktif,
    daftarPoliklinik,
    daftarDokter,
    pilihanStatus,
    isMenuDisabled,
    loadDashboard,
    loadUserManagement,
    toggleTheme,
    selectMenu,
    selectPatient,
    closePatientSidebar,
    applyPatientFilters,
    applyPatientPage,
  }
}
