<script setup lang="ts">
import { useId } from 'vue'
import JadwalOperasiPage from './JadwalOperasiPage.vue'
import PermintaanLaboratoriumPage from '../PermintaanLaboratorium/PermintaanLaboratoriumPage.vue'
import RiwayatPerawatanPage from '../RiwayatPerawatan/RiwayatPerawatanPage.vue'
import ChecklistPreOperasiPage from '../ChecklistPreOperasi/ChecklistPreOperasiPage.vue'
import InputResepPage from '../InputResep/InputResepPage.vue'
import RiwayatPendukungPage from './RiwayatPendukungPage.vue'
import { useTabOperasi, type PropsTabOperasi } from './useTabOperasi'
import { modulInputOperasi } from '../../../types/pendukungOperasi'

const props = defineProps<PropsTabOperasi>()
const {
  aktif, sesi, daftar, tabAktif, pilih, navigasi,
  jadwal, sesiPendukung, patientOperasi, pilihJadwal, sinkronkan,
} = useTabOperasi(props)
const id = useId()
const halaman: Record<string, unknown> = {
  permintaan_laboratorium: PermintaanLaboratoriumPage,
  riwayat_perawatan: RiwayatPerawatanPage,
  checklist_pre_operasi: ChecklistPreOperasiPage,
  input_resep: InputResepPage,
}
</script>

<template>
  <section class="operasi-workspace">
    <header class="operasi-tab-header">
      <h2>Pelayanan Operasi</h2>
      <p>Pilih baris jadwal terlebih dahulu, kemudian buka tab layanan pendukung.</p>
    </header>
    <div v-if="jadwal" class="operasi-konteks" role="status">
      <strong>{{ jadwal.nama_paket }}</strong>
      <span>{{ jadwal.tanggal }} · {{ jadwal.jam_mulai }}–{{ jadwal.jam_selesai }}</span>
      <span>Operator: {{ jadwal.nama_dokter }} · Ruang: {{ jadwal.nama_ruang }}</span>
      <button type="button" class="clinical-button secondary" @click="pilih('jadwal_operasi')">
        Lihat / Ganti Jadwal
      </button>
    </div>
    <div
      class="operasi-tabs"
      role="tablist"
      aria-label="Pelayanan operasi pasien"
      @keydown="navigasi"
    >
      <button
        v-for="tab in daftar"
        :id="id + '-tab-' + tab.kode"
        :key="tab.kode"
        type="button"
        role="tab"
        class="clinical-button"
        :class="aktif === tab.kode ? 'primary' : 'secondary'"
        :aria-selected="aktif === tab.kode"
        :aria-controls="id + '-panel'"
        :tabindex="aktif === tab.kode ? 0 : -1"
        :disabled="tab.kode !== 'jadwal_operasi' && !jadwal"
        :title="tab.kode !== 'jadwal_operasi' && !jadwal ? 'Pilih jadwal terlebih dahulu' : tab.label"
        @click="pilih(tab.kode)"
      >
        {{ tab.label }}
        <small v-if="tab.kode !== 'jadwal_operasi' && !halaman[tab.kode] && !modulInputOperasi.has(tab.kode)">Riwayat</small>
      </button>
    </div>
    <div
      v-if="tabAktif"
      :id="id + '-panel'"
      class="operasi-panel"
      role="tabpanel"
      :aria-labelledby="id + '-tab-' + aktif"
    >
      <KeepAlive :key="sesi">
        <JadwalOperasiPage
          v-if="aktif === 'jadwal_operasi'"
          :token="token"
          :patient="patient"
          @pilih-jadwal="pilihJadwal"
          @jadwal-dimuat="sinkronkan"
        />
      </KeepAlive>
      <KeepAlive :key="sesi + ':' + sesiPendukung" :max="18">
        <component
          :is="halaman[aktif] || RiwayatPendukungPage"
          v-if="aktif !== 'jadwal_operasi' && jadwal"
          :key="aktif"
          :token="token"
          :patient="patientOperasi"
          :jadwal-operasi="jadwal"
          :jenis="aktif"
          :judul="tabAktif.label"
          :module-name="moduleName"
        />
      </KeepAlive>
    </div>
    <p v-else class="operasi-tab-help" role="status">
      Tidak ada tab operasi yang aktif untuk layanan ini.
    </p>
  </section>
</template>

<style src="@/Pages/Pasien/JadwalOperasi/tab-operasi.css" scoped></style>
