import type { Pasien } from './domain'

export interface PropsRisikoJatuh {
  token: string
  patient: Pasien
}

export interface SkalaHumptyDumpty {
  label: string
  pilihan: string[]
  nilai: number[]
}

export interface CatatanRisikoJatuh {
  data: Record<string, string>
  nama_petugas: string
  bisa_ubah: boolean
  risiko: string
}

export interface HasilRisikoJatuh {
  skala: SkalaHumptyDumpty[]
  catatan: CatatanRisikoJatuh[]
}
