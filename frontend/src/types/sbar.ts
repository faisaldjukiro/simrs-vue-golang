import type { Pasien } from './domain'

export interface PropsSbar { token: string; patient: Pasien }
export interface DataSbar {
  tgl_perawatan: string
  jam_rawat: string
  nip: string
  situation: string
  background: string
  assesment: string
  recommendation: string
  instruksi: string
}
export interface PilihanSbar { kode: string; nama: string; jabatan: string }
export interface CatatanSbar {
  data: DataSbar
  nama_petugas: string
  jabatan: string
  status: string
  validator: string
  nama_validator: string
  tanggal_validasi: string
  jam_validasi: string
  revisi: string
  bisa_ubah: boolean
  bisa_verifikasi: boolean
  terkunci: boolean
}
export interface HasilSbar {
  catatan: CatatanSbar[]
  petugas_login: PilihanSbar
  dokter_login: PilihanSbar
  dpjp: PilihanSbar[]
  boleh_pilih_petugas: boolean
  batas_instruksi: number
}
export const bidangSbar = [
  { key: 'situation', huruf: 'S', label: 'Situation', petunjuk: 'Situasi atau keluhan yang dilaporkan.' },
  { key: 'background', huruf: 'B', label: 'Background', petunjuk: 'Latar belakang dan informasi yang berkaitan.' },
  { key: 'assesment', huruf: 'A', label: 'Assessment', petunjuk: 'Hasil penilaian petugas.' },
  { key: 'recommendation', huruf: 'R', label: 'Recommendation', petunjuk: 'Rekomendasi atau tindak lanjut yang disampaikan.' },
] as const
