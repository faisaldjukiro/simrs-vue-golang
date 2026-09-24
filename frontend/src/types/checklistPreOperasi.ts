import type { JadwalOperasi } from './jadwalOperasi'

export interface PropsChecklist {
  token: string
  patient: Record<string, unknown>
  jadwalOperasi?: JadwalOperasi
}

export interface CatatanChecklist {
  id: number
  versi: number
  no_rawat: string
  tanggal: string
  data: Record<string, string>
  sumber: string
  bisa_ubah: boolean
}

export interface BidangChecklist {
  kode: string
  label: string
  maxlength?: number
  pilihan?: string[]
  jenis?: 'dokter' | 'petugas'
}

export const bidangChecklist: BidangChecklist[] = [
  {
    "kode": "sncn",
    "label": "SN/CN",
    "maxlength": 25
  },
  {
    "kode": "tindakan",
    "label": "Tindakan",
    "maxlength": 50
  },
  {
    "kode": "kd_dokter_bedah",
    "label": "Dokter Bedah",
    "maxlength": 20,
    "jenis": "dokter"
  },
  {
    "kode": "kd_dokter_anestesi",
    "label": "Dokter Anestesi",
    "maxlength": 20,
    "jenis": "dokter"
  },
  {
    "kode": "identitas",
    "label": "Identitas",
    "pilihan": [
      "Ya",
      "Tidak"
    ]
  },
  {
    "kode": "surat_ijin_bedah",
    "label": "Surat Izin Bedah",
    "pilihan": [
      "Ada",
      "Tidak Ada"
    ]
  },
  {
    "kode": "surat_ijin_anestesi",
    "label": "Surat Izin Anestesi",
    "pilihan": [
      "Ada",
      "Tidak Ada"
    ]
  },
  {
    "kode": "surat_ijin_transfusi",
    "label": "Surat Izin Transfusi",
    "pilihan": [
      "Ada",
      "Tidak Ada",
      "Tidak Diperlukan"
    ]
  },
  {
    "kode": "penandaan_area_operasi",
    "label": "Penandaan Area Operasi",
    "pilihan": [
      "Ada",
      "Tidak Ada",
      "Tidak Diperlukan"
    ]
  },
  {
    "kode": "keadaan_umum",
    "label": "Keadaan Umum",
    "pilihan": [
      "Baik",
      "Sedang",
      "Lemah"
    ]
  },
  {
    "kode": "pemeriksaan_penunjang_rontgen",
    "label": "RONTGEN",
    "pilihan": [
      "Ada",
      "Tidak Ada",
      "Tidak Diperlukan"
    ]
  },
  {
    "kode": "keterangan_pemeriksaan_penunjang_rontgen",
    "label": "Keterangan RONTGEN",
    "maxlength": 20
  },
  {
    "kode": "pemeriksaan_penunjang_ekg",
    "label": "EKG",
    "pilihan": [
      "Ada",
      "Tidak Ada",
      "Tidak Diperlukan"
    ]
  },
  {
    "kode": "keterangan_pemeriksaan_penunjang_ekg",
    "label": "Keterangan EKG",
    "maxlength": 20
  },
  {
    "kode": "pemeriksaan_penunjang_usg",
    "label": "USG",
    "pilihan": [
      "Ada",
      "Tidak Ada",
      "Tidak Diperlukan"
    ]
  },
  {
    "kode": "keterangan_pemeriksaan_penunjang_usg",
    "label": "Keterangan USG",
    "maxlength": 20
  },
  {
    "kode": "pemeriksaan_penunjang_ctscan",
    "label": "CTSCAN",
    "pilihan": [
      "Ada",
      "Tidak Ada",
      "Tidak Diperlukan"
    ]
  },
  {
    "kode": "keterangan_pemeriksaan_penunjang_ctscan",
    "label": "Keterangan CTSCAN",
    "maxlength": 20
  },
  {
    "kode": "pemeriksaan_penunjang_mri",
    "label": "MRI",
    "pilihan": [
      "Ada",
      "Tidak Ada",
      "Tidak Diperlukan"
    ]
  },
  {
    "kode": "keterangan_pemeriksaan_penunjang_mri",
    "label": "Keterangan MRI",
    "maxlength": 20
  },
  {
    "kode": "persiapan_darah",
    "label": "Persiapan Darah",
    "pilihan": [
      "Ada",
      "Tidak Ada",
      "Tidak Diperlukan"
    ]
  },
  {
    "kode": "keterangan_persiapan_darah",
    "label": "Keterangan Persiapan Darah",
    "maxlength": 20
  },
  {
    "kode": "perlengkapan_khusus",
    "label": "Perlengkapan Khusus",
    "pilihan": [
      "Ada",
      "Tidak Ada",
      "Tidak Diperlukan"
    ]
  },
  {
    "kode": "nip_petugas_ruangan",
    "label": "Petugas Ruangan",
    "maxlength": 20,
    "jenis": "petugas"
  },
  {
    "kode": "nip_perawat_ok",
    "label": "Petugas OK",
    "maxlength": 20,
    "jenis": "petugas"
  }
]

export const kelompokChecklist = [
  { judul: 'Informasi operasi', kode: ['sncn', 'tindakan', 'kd_dokter_bedah', 'kd_dokter_anestesi'] },
  { judul: 'Konfirmasi dan izin', kode: ['identitas', 'keadaan_umum', 'surat_ijin_bedah', 'surat_ijin_anestesi', 'surat_ijin_transfusi', 'penandaan_area_operasi'] },
  { judul: 'Pemeriksaan penunjang', kode: bidangChecklist.filter(b => b.kode.includes('penunjang')).map(b => b.kode) },
  { judul: 'Persiapan dan petugas', kode: ['persiapan_darah', 'keterangan_persiapan_darah', 'perlengkapan_khusus', 'nip_petugas_ruangan', 'nip_perawat_ok'] },
].map(k => ({
  judul: k.judul,
  bidang: k.kode.map(kode => bidangChecklist.find(b => b.kode === kode)!),
}))
