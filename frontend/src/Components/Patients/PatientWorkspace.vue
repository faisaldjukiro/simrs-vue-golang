<script setup>
import {
  Activity,
  ArrowLeft,
  Bed,
  ClipboardList,
  FileText,
  FlaskConical,
  HeartPulse,
  History,
  LayoutDashboard,
  NotebookText,
  PanelLeftClose,
  PanelLeftOpen,
  Pill,
  ScanSearch,
  Search,
  Stethoscope,
  Syringe,
  Users,
} from '@lucide/vue'
import { computed, ref, watch } from 'vue'
import PatientIdentityHeader from './PatientIdentityHeader.vue'
import CpptPage from '../../Pages/RawatInap/CpptPage.vue'
import PenangananDokterPetugasPage from '../../Pages/RawatInap/PenangananDokterPetugasPage.vue'
import PermintaanRadiologiPage from '../../Pages/Pasien/PermintaanRadiologiPage.vue'

const props = defineProps({
  moduleName: { type: String, required: true },
  patient: { type: Object, required: true },
  menus: { type: Array, default: () => [] },
  token: { type: String, required: true },
})

const emit = defineEmits(['back'])
const activeSection = ref('Ringkasan')
const menuSearch = ref('')
const sidebarCollapsed = ref(localStorage.getItem('sirava.patient_workspace.sidebar_collapsed') === '1')

const iconMap = {
  Activity,
  Bed,
  ClipboardList,
  FileText,
  FileSignature: FileText,
  FlaskConical,
  HeartPulse,
  History,
  LayoutDashboard,
  NotebookText,
  Pill,
  ScanSearch,
  Stethoscope,
  Syringe,
  Users,
}

const fallbackMenus = [
  { label: 'Input Resep', icon: 'Pill', modules: ['IGD/UGD', 'Rawat Jalan', 'Rawat Inap'] },
  { label: 'Permintaan Lab', icon: 'FlaskConical', modules: ['IGD/UGD', 'Rawat Jalan', 'Rawat Inap'] },
  { label: 'Permintaan Radiologi', icon: 'ScanSearch', modules: ['IGD/UGD', 'Rawat Jalan', 'Rawat Inap'] },
  { label: 'Triase IGD', icon: 'Stethoscope', modules: ['IGD/UGD'] },
]

const menusAktif = computed(() => {
  const sumber = props.menus.length > 0 ? props.menus : fallbackMenus
  const daftar = [{ label: 'Ringkasan', icon: 'LayoutDashboard', modules: [props.moduleName] }]
  const sudahAda = new Set(['Ringkasan'])

  sumber.forEach((menu) => {
    const modules = Array.isArray(menu.modules) ? menu.modules : []
    const namaMenu = menu.label === 'Permintaan Rad' ? 'Permintaan Radiologi' : menu.label
    if (!modules.includes(props.moduleName) || sudahAda.has(namaMenu)) return
    sudahAda.add(namaMenu)
    daftar.push({ ...menu, label: namaMenu })
  })

  return daftar.map((menu) => ({
    ...menu,
    iconComponent: iconMap[menu.icon] || LayoutDashboard,
  }))
})

const menusTersaring = computed(() => {
  const kata = menuSearch.value.trim().toLowerCase()
  if (!kata) return menusAktif.value
  return menusAktif.value.filter((menu) => menu.label.toLowerCase().includes(kata))
})

const ringkasan = computed(() => [
  ['No. Rawat', props.patient.no_rawat],
  ['No. Rekam Medis', props.patient.no_rekam_medis],
  ['No. Registrasi', props.patient.no_registrasi],
  ['Jam Registrasi', props.patient.jam_registrasi],
  ['No. SEP', props.patient.no_sep],
  ['Status Bayar', props.patient.status_bayar],
])

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

watch(sidebarCollapsed, (collapsed) => {
  localStorage.setItem('sirava.patient_workspace.sidebar_collapsed', collapsed ? '1' : '0')
})

watch(() => props.patient.no_rawat, () => {
  activeSection.value = 'Ringkasan'
  menuSearch.value = ''
})
</script>

<template>
  <section class="patient-workspace" :class="{ collapsed: sidebarCollapsed }">
    <aside class="patient-workspace-sidebar">
      <div class="patient-workspace-sidebar-tools">
        <button type="button" title="Kembali ke daftar pasien" @click="emit('back')">
          <ArrowLeft :size="17" />
          <span>Daftar Pasien</span>
        </button>
        <button type="button" :title="sidebarCollapsed ? 'Besarkan sidebar' : 'Perkecil sidebar'" @click="toggleSidebar">
          <PanelLeftOpen v-if="sidebarCollapsed" :size="18" />
          <PanelLeftClose v-else :size="18" />
        </button>
      </div>

      <label class="patient-workspace-search">
        <Search :size="16" />
        <input v-model="menuSearch" type="search" placeholder="Cari menu pasien..." />
      </label>

      <nav class="patient-workspace-menu">
        <button
          v-for="menu in menusTersaring"
          :key="menu.label"
          type="button"
          :class="{ active: activeSection === menu.label }"
          :title="menu.label"
          @click="activeSection = menu.label"
        >
          <i><component :is="menu.iconComponent" :size="17" /></i>
          <span>{{ menu.label }}</span>
        </button>
        <p v-if="menusTersaring.length === 0">Menu tidak ditemukan.</p>
      </nav>
    </aside>

    <main class="patient-workspace-content">
      <PatientIdentityHeader
        :active-section="activeSection"
        :module-name="moduleName"
        :patient="patient"
      />

      <section v-if="activeSection === 'Ringkasan'" class="patient-workspace-summary">
        <article v-for="item in ringkasan" :key="item[0]">
          <span>{{ item[0] }}</span>
          <strong>{{ item[1] || '-' }}</strong>
        </article>
      </section>

      <CpptPage
        v-else-if="activeSection === 'Cppt/Soap'"
        :token="token"
        :patient="patient"
      />

      <PenangananDokterPetugasPage
        v-else-if="['Penangangan Dokter & Petugas', 'Penanganan Dokter & Petugas'].includes(activeSection)"
        :token="token"
        :patient="patient"
      />

      <PermintaanRadiologiPage
        v-else-if="['Permintaan Rad', 'Permintaan Radiologi'].includes(activeSection)"
        :token="token"
        :patient="patient"
      />

      <section v-else class="patient-workspace-placeholder">
        <component :is="menusAktif.find((menu) => menu.label === activeSection)?.iconComponent || LayoutDashboard" :size="32" />
        <span>Administrasi Pasien</span>
        <h3>{{ activeSection }}</h3>
        <p>Ruang kerja {{ activeSection }} untuk pasien ini sudah disiapkan. Form dan prosesnya akan dipindahkan bertahap dari SIMRS lama.</p>
      </section>
    </main>
  </section>
</template>
