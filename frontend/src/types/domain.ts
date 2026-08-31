export type Dictionary<T = any> = Record<string, T>

export interface Pasien extends Dictionary {
  no_rawat: string
  no_rkm_medis?: string
  nm_pasien?: string
  jenis_rawat?: 'Rawat Jalan' | 'Rawat Inap' | 'IGD/UGD' | string
  kamar?: string
  poliklinik?: string
}

export interface Pengguna extends Dictionary {
  id?: number | string
  name: string
  username?: string
  nip?: string
  permissions?: string[]
}

export interface Pilihan<T = string> {
  label: string
  value: T
}

export interface ResponsApi<T> {
  success?: boolean
  data?: T
  error?: {
    code?: string
    message?: string
  }
}

export interface KesalahanApi extends Error {
  code?: string
  status?: number
}
