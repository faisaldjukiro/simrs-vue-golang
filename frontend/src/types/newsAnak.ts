import type { Pasien } from './domain'

export interface PropsNewsAnak {
  token: string
  patient: Pasien
}

export interface CatatanNews {
  data: Record<string, string>
  nama_petugas: string
  bisa_ubah: boolean
}

export interface BidangNews {
  key: string
  skor: string
  label: string
  pilihan: string[]
  nilai: number[]
}

export interface PanduanNews {
  kode: string
  syarat: string
  respon: string
  eskalasi: string
}

export interface HasilNews {
  catatan: CatatanNews[]
  bidang: BidangNews[]
  panduan: PanduanNews[]
  petugas_login: Record<string, string>
  boleh_pilih_petugas: boolean
}
