export interface PropsJadwalOperasi {
  token: string
  patient: Record<string, unknown>
}

export interface JadwalOperasi {
  id: number
  versi: number
  no_rawat: string
  kode_paket: string
  tanggal: string
  jam_mulai: string
  jam_selesai: string
  status: string
  kd_dokter: string
  kd_ruang_ok: string
  dokteranastesi: string
  perawat: string
  nama_paket: string
  nama_dokter: string
  nama_ruang: string
  sumber: string
  bisa_ubah: boolean
}

export const statusOperasi = ['Permintaan', 'Menunggu', 'Proses Operasi', 'Selesai']
  .map(value => ({ label: value, value }))
