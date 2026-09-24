import { computed, ref, watch } from 'vue'
import type { JadwalOperasi } from '../../../types/jadwalOperasi'

export interface PropsTabOperasi {
  token: string
  patient: Record<string, unknown>
  moduleName: string
  kodeSidebar: string[]
}

const tabOperasi = [
  { kode: 'jadwal_operasi', label: 'Jadwal Operasi' },
  { kode: 'kamar_inap', label: 'Kamar Inap' },
  { kode: 'permintaan_laboratorium', label: 'Permintaan Lab' },
  { kode: 'riwayat_perawatan', label: 'Riwayat Perawatan' },
  { kode: 'penilaian_pre_induksi', label: 'Penilaian Pre Induksi' },
  { kode: 'checklist_pre_operasi', label: 'Checklist Pre Operasi' },
  { kode: 'signin_sebelum_anestesi', label: 'Sign-In Sebelum Anestesi' },
  { kode: 'timeout_sebelum_insisi', label: 'Time-Out Sebelum Insisi' },
  { kode: 'signout_sebelum_menutup_luka', label: 'Sign-Out Sebelum Menutup Luka' },
  { kode: 'checklist_post_operasi', label: 'Checklist Post Operasi' },
  { kode: 'penilaian_pre_operasi', label: 'Penilaian Pre Operasi' },
  { kode: 'penilaian_pre_anestesi', label: 'Penilaian Pre Anestesi' },
  { kode: 'laporan_operasi', label: 'Laporan Operasi' },
  { kode: 'tagihan_operasi', label: 'Tagihan Operasi/VK' },
  { kode: 'input_resep', label: 'Permintaan Resep' },
  { kode: 'transfer_pasien_antar_ruang', label: 'Transfer Antar Ruang' },
  { kode: 'skor_aldrette_pasca_anestesi', label: 'Skor Aldrette Pasca Anestesi' },
  { kode: 'skor_steward_pasca_anestesi', label: 'Skor Steward Pasca Anestesi' },
  { kode: 'skor_bromage_pasca_anestesi', label: 'Skor Bromage Pasca Anestesi' },
]

export function useTabOperasi(props: PropsTabOperasi) {
  const aktif = ref('jadwal_operasi')
  const sesi = ref(0)
  const jadwal = ref<JadwalOperasi | null>(null)
  const sesiPendukung = ref(0)
  // Menu ini adalah bagian Jadwal Operasi, bukan permission sidebar baru.
  const daftar = computed(() => props.kodeSidebar.includes('jadwal_operasi') ? tabOperasi : [])
  const tabAktif = computed(() => daftar.value.find(tab => tab.kode === aktif.value))
  const patientOperasi = computed(() => ({
    ...props.patient,
    kd_dokter: jadwal.value?.kd_dokter,
    nama_dokter: jadwal.value?.nama_dokter,
  }))

  function pilihJadwal(baris: JadwalOperasi) {
    if (baris.no_rawat !== String(props.patient.no_rawat || '')) return
    if (JSON.stringify(jadwal.value) === JSON.stringify(baris)) return
    if (jadwal.value && !window.confirm('Ganti jadwal? Isian tab pendukung yang belum disimpan akan dikosongkan.')) return
    jadwal.value = { ...baris }
    sesiPendukung.value++
  }

  function sinkronkan(baris: JadwalOperasi[]) {
    if (!jadwal.value) return
    const lama = jadwal.value
    if (!baris.some(r => Object.keys(r).every(k => r[k as keyof JadwalOperasi] === lama[k as keyof JadwalOperasi]))) {
      jadwal.value = null
      sesiPendukung.value++
      aktif.value = 'jadwal_operasi'
    }
  }

  function pilih(kode: string) {
    if (kode !== 'jadwal_operasi' && !jadwal.value) return
    if (daftar.value.some(tab => tab.kode === kode)) aktif.value = kode
  }

  function navigasi(event: KeyboardEvent) {
    if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return
    const buttons = Array.from((event.currentTarget as HTMLElement)
      .querySelectorAll<HTMLButtonElement>('[role="tab"]:not(:disabled)'))
    if (!buttons.length) return
    const index = buttons.indexOf(event.target as HTMLButtonElement)
    if (index < 0) return
    event.preventDefault()
    const target = event.key === 'Home' ? 0
      : event.key === 'End' ? buttons.length - 1
        : (index + (event.key === 'ArrowRight' ? 1 : -1) + buttons.length) % buttons.length
    buttons[target].focus()
    buttons[target].click()
    buttons[target].scrollIntoView({ block: 'nearest', inline: 'nearest' })
  }

  watch(
    () => [props.token, props.patient.no_rawat, props.moduleName, props.kodeSidebar.join('|')],
    () => {
      // Buang cache formulir ketika pasien, sesi login, atau akses berganti.
      sesi.value++
      jadwal.value = null
      sesiPendukung.value++
      aktif.value = daftar.value.find(tab => tab.kode === 'jadwal_operasi')?.kode
        || daftar.value[0]?.kode || ''
    },
    { immediate: true, flush: 'sync' },
  )

  return {
    aktif, sesi, daftar, tabAktif, pilih, navigasi,
    jadwal, sesiPendukung, patientOperasi, pilihJadwal, sinkronkan,
  }
}
