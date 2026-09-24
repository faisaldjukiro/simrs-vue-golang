import type { JadwalOperasi } from './jadwalOperasi'

export type KategoriLaboratorium = 'PK' | 'PA' | 'MB'

export interface SpesimenLaboratorium {
  pengambilan_bahan: string
  diperoleh_dengan: string
  lokasi_jaringan: string
  diawetkan_dengan: string
  pernah_dilakukan_di: string
  tanggal_pa_sebelumnya: string
  nomor_pa_sebelumnya: string
  diagnosa_pa_sebelumnya: string
}

export interface DetailLaboratorium {
  id: number
  nama: string
  satuan: string
  nilai_rujukan: string
  biaya: number
  status_bayar: string
  dipilih?: boolean
}

export interface PemeriksaanLaboratorium {
  kode: string
  nama: string
  kelas: string
  total: number
  detail: DetailLaboratorium[]
  detail_opsi?: DetailLaboratorium[]
}

export interface PermintaanLaboratorium {
  nomor: string
  kategori: KategoriLaboratorium
  no_rawat: string
  tanggal: string
  jam: string
  tanggal_sampel: string
  jam_sampel: string
  tanggal_hasil: string
  jam_hasil: string
  kode_dokter: string
  nama_dokter: string
  informasi_tambahan: string
  diagnosis_klinis: string
  status_pemeriksaan: string
  status_bayar: string
  pemeriksaan: PemeriksaanLaboratorium[]
  spesimen: SpesimenLaboratorium
  total: number
  dapat_diubah: boolean
  dapat_dihapus: boolean
}

export interface PropsLaboratorium {
  token: string
  jadwalOperasi?: JadwalOperasi
  patient: {
    no_rawat?: string
    nama_pasien?: string
    nm_pasien?: string
    no_rm?: string
    no_rkm_medis?: string
  }
}

export const kategoriLaboratorium: { value: KategoriLaboratorium; label: string }[] = [
  { value: 'PK', label: 'Patologi Klinis' },
  { value: 'PA', label: 'Patologi Anatomi' },
  { value: 'MB', label: 'Mikrobiologi & Biomolekuler' },
]
