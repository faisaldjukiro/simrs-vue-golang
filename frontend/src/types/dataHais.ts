import type { Pasien } from './domain'

export interface PropsHais {
  token: string
  patient: Pasien
}

export interface CatatanHais {
  data: Record<string, string>
  sumber: string
  bisa_ubah: boolean
}

export interface HasilHais {
  catatan: CatatanHais[]
  kamar: string
}

export const bidangHais = [
  { key: 'ETT', label: 'ETT', grup: 'Hari Pemasangan Alat', angka: true },
  { key: 'CVL', label: 'CVL', grup: 'Hari Pemasangan Alat', angka: true },
  { key: 'IVL', label: 'IVL', grup: 'Hari Pemasangan Alat', angka: true },
  { key: 'UC', label: 'UC', grup: 'Hari Pemasangan Alat', angka: true },
  { key: 'VAP', label: 'VAP', grup: 'Infeksi RS', angka: true },
  { key: 'IAD', label: 'IAD', grup: 'Infeksi RS', angka: true },
  { key: 'PLEB', label: 'PLEB', grup: 'Infeksi RS', angka: true },
  { key: 'ISK', label: 'ISK', grup: 'Infeksi RS', angka: true },
  { key: 'ILO', label: 'ILO', grup: 'Infeksi RS', angka: true },
  { key: 'HAP', label: 'HAP', grup: 'Infeksi RS', angka: true },
  { key: 'Tinea', label: 'Tinea', grup: 'Infeksi RS', angka: true },
  { key: 'Scabies', label: 'Scabies', grup: 'Infeksi RS', angka: true },
  { key: 'SPUTUM', label: 'Sputum', grup: 'Kultur dan Antibiotik', angka: false },
  { key: 'DARAH', label: 'Darah', grup: 'Kultur dan Antibiotik', angka: false },
  { key: 'URINE', label: 'Urine', grup: 'Kultur dan Antibiotik', angka: false },
  { key: 'ANTIBIOTIK', label: 'Antibiotik', grup: 'Kultur dan Antibiotik', angka: false },
]
