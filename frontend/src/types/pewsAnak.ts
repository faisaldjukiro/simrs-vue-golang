import type { Pasien } from './domain'

export interface PropsPewsAnak {
  token: string
  patient: Pasien
}

export interface CatatanPews {
  data: Record<string, string>
  nama_petugas: string
  bisa_ubah: boolean
}

export interface BidangPews {
  key: string
  skor: string
  label: string
  pilihan: string[]
}

export interface PanduanPews {
  minimal: number
  maksimal: number
  parameter: string
}

export interface HasilPews {
  catatan: CatatanPews[]
  bidang: BidangPews[]
  panduan: PanduanPews[]
  petugas_login: Record<string, string>
  boleh_pilih_petugas: boolean
}
