import type { Pasien } from './domain'

export interface PropsRujukanInternal {
  token: string
  patient: Pasien
  moduleName: string
}

export interface ReferensiRujukan {
  kode?: string
  nama?: string
}

export interface InputRujukan {
  no_rawat: string
  kd_dokter: string
  kd_poli: string
  tanggal: string
  jam: string
}

export interface RujukanInternal extends InputRujukan {
  id: number
  konflik_khanza: boolean
  nama_dokter: string
  nama_poli: string
  sumber: 'SIRAPI' | 'Khanza'
}

export interface DataRujukanInternal {
  daftar: RujukanInternal[]
  boleh_simpan: boolean
  pesan: string
}
