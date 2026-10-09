import type { Pasien } from './domain'

export interface PropsEdukasi {
  token: string
  patient: Pasien
}

export interface CatatanEdukasi {
  data: Record<string, string>
  nama_petugas: string
  nama_ruangan: string
  sumber: string
  bisa_ubah: boolean
  foto_url: string
}

export interface HasilEdukasi {
  catatan: CatatanEdukasi[]
  petugas_login: Record<string, string>
  boleh_pilih_petugas: boolean
}

export const bidangEdukasi = [
  { key: 'metode', label: 'Metode' },
  { key: 'durasi', label: 'Durasi' },
  { key: 'materi', label: 'Materi' },
  { key: 'penerima', label: 'Penerima' },
  { key: 'keterangan', label: 'Keterangan' },
] as const
