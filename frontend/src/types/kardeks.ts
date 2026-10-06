import type { Pasien } from './domain'
export interface PropsKardeks { token: string; patient: Pasien }
export interface BagianKardeks {
  kode: string
  nama: string
  baris: Record<string, string>[]
  error?: string
}
export interface HasilKardeks { mulai: string; selesai: string; bagian: BagianKardeks[] }
export const parameterKardeks = [
  { key: 'td', label: 'Tekanan darah', unit: 'mmHg' },
  { key: 'nadi', label: 'Nadi', unit: '/menit' },
  { key: 'respirasi', label: 'Respirasi', unit: '/menit' },
  { key: 'suhu', label: 'Suhu', unit: '°C' },
  { key: 'spo2', label: 'SpO₂', unit: '%' },
  { key: 'gcs', label: 'GCS (sesuai catatan)', unit: '' },
]
