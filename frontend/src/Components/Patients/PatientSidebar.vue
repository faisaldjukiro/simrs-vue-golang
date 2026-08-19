<script setup>
import * as LucideIcons from '@lucide/vue'
import { computed, ref, watch } from 'vue'
import PatientIdentityHeader from './PatientIdentityHeader.vue'
import CpptPage from '../../Pages/RawatInap/CpptPage.vue'
import PenangananDokterPetugasPage from '../../Pages/Pasien/PenangananDokterPetugasPage.vue'
import PermintaanRadiologiPage from '../../Pages/Pasien/PermintaanRadiologiPage.vue'
import PermintaanLaboratoriumPage from '../../Pages/Pasien/PermintaanLaboratoriumPage.vue'
import RiwayatPerawatanPage from '../../Pages/Pasien/RiwayatPerawatanPage.vue'
import TriaseIgdPage from '../../Pages/IGD/TriaseIgdPage.vue'
import AwalKeperawatanIgdPage from '../../Pages/IGD/AwalKeperawatanIgdPage.vue'
import ResumePasienRanapPage from '../../Pages/RawatInap/ResumePasienRanapPage.vue'
import DiagnosaPasienPage from '../../Pages/Pasien/DiagnosaPasienPage.vue'
import AwalMedisUmumPage from '../../Pages/RawatJalan/AwalMedisUmumPage.vue'
import AwalMedisRanapPage from '../../Pages/RawatInap/AwalMedisRanapPage.vue'
import AwalMedisIgdPage from '../../Pages/IGD/AwalMedisIgdPage.vue'

const { ArrowLeft, LayoutDashboard, PanelLeftClose, PanelLeftOpen, Search, ShieldX, WifiOff } = LucideIcons

const props = defineProps({
  moduleName: { type: String, required: true },
  patient: { type: Object, required: true },
  sidebar: { type: Array, default: () => [] },
  sidebarLoading: Boolean,
  sidebarError: { type: String, default: '' },
  token: { type: String, required: true },
})

const emit = defineEmits(['back'])
const kodeSidebarAktif = ref('ringkasan')
const pencarianSidebar = ref('')
const sidebarCollapsed = ref(localStorage.getItem('sirava.patient_sidebar.collapsed') === '1')

const halamanSidebar = {
  cppt_soap: CpptPage,
  penanganan_dokter_petugas: PenangananDokterPetugasPage,
  permintaan_radiologi: PermintaanRadiologiPage,
  permintaan_laboratorium: PermintaanLaboratoriumPage,
  riwayat_perawatan: RiwayatPerawatanPage,
  triase_igd: TriaseIgdPage,
  awal_keperawatan_igd: AwalKeperawatanIgdPage,
  resume_pasien: ResumePasienRanapPage,
  diagnosa: DiagnosaPasienPage,
  awal_medis_igd: AwalMedisIgdPage,
  sidebar_0031: AwalMedisUmumPage,
}

const daftarSidebarAktif = computed(() => {
  const sumberModul = props.sidebar.filter((item) => {
    const daftarModul = Array.isArray(item.daftar_modul) ? item.daftar_modul : []
    return item.kode && daftarModul.includes(props.moduleName)
  })
  if (sumberModul.length === 0) return []

  const daftar = [{ kode: 'ringkasan', nama: 'Ringkasan', ikon: 'LayoutDashboard' }]
  const kodeTerpakai = new Set(['ringkasan'])
  sumberModul.forEach((item) => {
    if (kodeTerpakai.has(item.kode)) return
    kodeTerpakai.add(item.kode)
    daftar.push(item)
  })

  return daftar.map((item) => ({
    ...item,
    iconComponent: LucideIcons[item.ikon] || LayoutDashboard,
  }))
})

const aksesSidebarTersedia = computed(() => !props.sidebarLoading && !props.sidebarError && daftarSidebarAktif.value.length > 0)
const statusSidebar = computed(() => {
  if (props.sidebarLoading) return { judul: 'Memeriksa akses sidebar', pesan: 'Mohon tunggu sebentar.', ikon: Search }
  if (props.sidebarError) return { judul: 'Koneksi jaringan bermasalah', pesan: 'Daftar sidebar pasien tidak dapat dimuat. Silakan kembali dan coba lagi.', ikon: WifiOff }
  return { judul: 'Tidak memiliki akses', pesan: 'Akun Anda belum memiliki akses ke menu pelayanan pasien ini.', ikon: ShieldX }
})

const daftarSidebarTersaring = computed(() => {
  const kata = pencarianSidebar.value.trim().toLowerCase()
  if (!kata) return daftarSidebarAktif.value
  return daftarSidebarAktif.value.filter((item) => item.nama.toLowerCase().includes(kata))
})

const sidebarAktif = computed(() => daftarSidebarAktif.value.find((item) => item.kode === kodeSidebarAktif.value))
const komponenSidebarAktif = computed(() => {
  if (kodeSidebarAktif.value === 'sidebar_0031') {
    return props.moduleName === 'Rawat Inap' ? AwalMedisRanapPage : AwalMedisUmumPage
  }
  return halamanSidebar[kodeSidebarAktif.value] || null
})
const propertiHalamanAktif = computed(() => {
  const properti = { token: props.token, patient: props.patient }
  if (['cppt_soap', 'penanganan_dokter_petugas'].includes(kodeSidebarAktif.value)) {
    properti.jenisRawat = props.moduleName === 'Rawat Inap' ? 'ranap' : 'ralan'
    properti.namaModul = props.moduleName
  }
  if (kodeSidebarAktif.value === 'diagnosa') properti.moduleName = props.moduleName
  return properti
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
  localStorage.setItem('sirava.patient_sidebar.collapsed', collapsed ? '1' : '0')
})

watch(() => props.patient.no_rawat, () => {
  kodeSidebarAktif.value = aksesSidebarTersedia.value ? 'ringkasan' : ''
  pencarianSidebar.value = ''
})

watch(aksesSidebarTersedia, (tersedia) => {
  if (!tersedia) kodeSidebarAktif.value = ''
  else if (!daftarSidebarAktif.value.some((item) => item.kode === kodeSidebarAktif.value)) kodeSidebarAktif.value = 'ringkasan'
}, { immediate: true })
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

      <label v-if="aksesSidebarTersedia" class="patient-workspace-search">
        <Search :size="16" />
        <input v-model="pencarianSidebar" type="search" placeholder="Cari sidebar pasien..." />
      </label>

      <nav v-if="aksesSidebarTersedia" class="patient-workspace-menu">
        <button
          v-for="item in daftarSidebarTersaring"
          :key="item.kode"
          type="button"
          :class="{ active: kodeSidebarAktif === item.kode }"
          :title="item.nama"
          @click="kodeSidebarAktif = item.kode"
        >
          <i><component :is="item.iconComponent" :size="17" /></i>
          <span>{{ item.nama }}</span>
        </button>
        <p v-if="daftarSidebarTersaring.length === 0">Sidebar tidak ditemukan.</p>
      </nav>

      <div v-else class="patient-workspace-menu-state">
        <i><component :is="statusSidebar.ikon" :size="22" /></i>
        <strong>{{ statusSidebar.judul }}</strong>
        <p>{{ statusSidebar.pesan }}</p>
      </div>
    </aside>

    <main class="patient-workspace-content">
      <section v-if="!aksesSidebarTersedia" class="patient-workspace-access-state">
        <i><component :is="statusSidebar.ikon" :size="34" /></i>
        <span>Akses Pelayanan Pasien</span>
        <h3>{{ statusSidebar.judul }}</h3>
        <p>{{ statusSidebar.pesan }}</p>
      </section>

      <template v-else>
        <PatientIdentityHeader
          :active-section="sidebarAktif?.nama || 'Ringkasan'"
          :module-name="moduleName"
          :patient="patient"
        />

      <section v-if="kodeSidebarAktif === 'ringkasan'" class="patient-workspace-summary">
        <article v-for="item in ringkasan" :key="item[0]">
          <span>{{ item[0] }}</span>
          <strong>{{ item[1] || '-' }}</strong>
        </article>
      </section>

      <component
        :is="komponenSidebarAktif"
        v-else-if="komponenSidebarAktif"
        v-bind="propertiHalamanAktif"
      />

      <section v-else class="patient-workspace-placeholder">
        <component :is="sidebarAktif?.iconComponent || LayoutDashboard" :size="32" />
        <span>Sidebar Pasien</span>
        <h3>{{ sidebarAktif?.nama }}</h3>
        <p>Ruang kerja {{ sidebarAktif?.nama }} untuk pasien ini sudah disiapkan. Form dan prosesnya akan dipindahkan bertahap dari SIMRS lama.</p>
      </section>
      </template>
    </main>
  </section>
</template>
