import type { Pasien } from './domain'

export interface PropsObservasi {
  token: string
  patient: Pasien
}

export interface CatatanObservasi {
  data: Record<string, string>
  nama_petugas: string
  bisa_ubah: boolean
}

export interface HasilObservasi {
  catatan: CatatanObservasi[]
}

export const bidangObservasi = [
  { key: 'gcs', label: 'GCS (E, V, M)', batas: 10 },
  { key: 'td', label: 'TD (mmHg)', batas: 8 },
  { key: 'hr', label: 'HR (x/menit)', batas: 5 },
  { key: 'rr', label: 'RR (x/menit)', batas: 5 },
  { key: 'suhu', label: 'Suhu (°C)', batas: 5 },
  { key: 'spo2', label: 'SpO₂ (%)', batas: 3 },
] as const
