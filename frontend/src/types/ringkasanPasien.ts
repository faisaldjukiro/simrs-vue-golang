import type { Pasien } from './domain'

export interface PropsRingkasanPasien {
  token: string
  patient: Pasien
  moduleName: string
  kodeSidebar: string[]
}

export interface BagianRingkasan {
  kode: string
  sumber: string
  baris: Record<string, string>[]
  error?: string
}

export interface RingkasanPasien {
  no_rawat: string
  modul: string
  bagian: BagianRingkasan[]
}
