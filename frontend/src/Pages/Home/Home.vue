<script setup lang="ts">
import DashboardLayout from "../../Components/Layout/DashboardLayout.vue"
import PatientSidebar from "../../Components/Patients/PatientSidebar.vue"
import BerandaTab from "../../Components/Tabs/BerandaTab.vue"
import IdrgPage from "../Eklaim/Idrg/IdrgPage.vue"
import MonitoringDataKlaimPage from "../BPJS/MonitoringDataKlaim/MonitoringDataKlaimPage.vue"
import AktivitasLogPage from "../Sistem/AktivitasLog/AktivitasLogPage.vue"
import KelolaMenuPage from "../Sistem/KelolaMenu/KelolaMenuPage.vue"
import WhatsAppGatewayPage from "../Integrasi/WhatsAppGateway/WhatsAppGatewayPage.vue"
import MasterVentilatorPage from "../Sistem/MasterVentilator/MasterVentilatorPage.vue"
import KunjunganRalanPage from "../Laporan/KunjunganRalan/KunjunganRalanPage.vue"
import KunjunganRanapPage from "../Laporan/KunjunganRanap/KunjunganRanapPage.vue"
import ModulePlaceholder from "../../Components/Tabs/ModulePlaceholder.vue"
import ModuleTab from "../../Components/Tabs/ModuleTab.vue"
import UserManagementTab from "../../Components/Tabs/UserManagementTab.vue"
import { useHome } from "./useHome"

const props = defineProps({
  user: { type: Object, default: null },
  token: { type: String, default: '' },
  loading: Boolean,
})
const emit = defineEmits(['login', 'logout'])

const {
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
} = useHome(props, emit)
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
      :is-menu-disabled="isMenuDisabled"
      @select-tab="selectMenu"
    />

    <PatientSidebar
      v-else-if="selectedPatient"
      :module-name="selectedPatientModule"
      :patient="selectedPatient"
      :sidebar="dashboard.sidebar_pasien"
      :sidebar-loading="dashboardLoading"
      :sidebar-error="dashboardError"
      :token="token"
      @back="closePatientSidebar"
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

    <MonitoringDataKlaimPage
      v-else-if="currentTab === 'Monitoring Klaim BPJS'"
      :token="token"
    />

    <AktivitasLogPage
      v-else-if="currentTab === 'Log Aktivitas'"
      :token="token"
    />

    <KelolaMenuPage
      v-else-if="currentTab === 'Kelola Menu'"
      :token="token"
      @saved="loadDashboard"
    />

    <WhatsAppGatewayPage
      v-else-if="currentTab === 'WhatsApp Gateway'"
      :token="token"
    />

    <MasterVentilatorPage
      v-else-if="currentTab === 'Master Ventilator'"
      :token="token"
    />

    <KunjunganRalanPage
      v-else-if="currentTab === 'Laporan Kunjungan Ralan'"
      :token="token"
    />
    <KunjunganRanapPage v-else-if="currentTab === 'Laporan Kunjungan Ranap'" :token="token" />

    <ModulePlaceholder
      v-else
      :current-tab="currentTab"
      :ribbon-menus="ribbonMenus"
      @back="currentTab = 'Menu'"
    />
  </DashboardLayout>
</template>
