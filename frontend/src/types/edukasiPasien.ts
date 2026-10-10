import type { Pasien } from './domain'

export interface PropsEdukasi {
  token: string
  patient: Pasien
}

export interface CatatanEdukasi {
  data: Record<string, string>
  nama_petugas: string
  nama_ruangan: string
  nama_verifikator: string
  sumber: string
  bisa_ubah: boolean
  foto_url: string
}

export interface HasilEdukasi {
  catatan: CatatanEdukasi[]
  petugas_login: Record<string, string>
  boleh_pilih_petugas: boolean
}

export const bidangEdukasi = [
  { key: 'metode', label: 'Metode' },
  { key: 'durasi', label: 'Durasi' },
  { key: 'materi', label: 'Materi' },
  { key: 'penerima', label: 'Penerima' },
  { key: 'keterangan', label: 'Keterangan' },
] as const

interface BidangAsesmen {
  key: string
  label: string
  maxlength: number
  jenis?: string
  pilihan?: string[]
}

export const bidangAsesmen: BidangAsesmen[] = [
  { key: 'kemampuan_membaca', label: 'Kemampuan Membaca', maxlength: 30, pilihan: ['Bisa membaca', 'Perlu bantuan', 'Tidak bisa membaca'] },
  { key: 'tingkat_pendidikan', label: 'Pendidikan Penerima Edukasi', maxlength: 30 },
  { key: 'bahasa', label: 'Bahasa yang Dipahami', maxlength: 50 },
  { key: 'motivasi', label: 'Motivasi Menerima Edukasi', maxlength: 100 },
  { key: 'kesediaan_menerima', label: 'Kesediaan Menerima Informasi', maxlength: 30, pilihan: ['Bersedia', 'Belum bersedia', 'Tidak bersedia'] },
  { key: 'hambatan_emosional', label: 'Hambatan Emosional', maxlength: 255, jenis: 'textarea' },
  { key: 'keterbatasan_fisik', label: 'Keterbatasan Fisik', maxlength: 255, jenis: 'textarea' },
  { key: 'keterbatasan_kognitif', label: 'Keterbatasan Kognitif', maxlength: 255, jenis: 'textarea' },
  { key: 'nilai_budaya', label: 'Nilai Budaya / Keyakinan', maxlength: 255, jenis: 'textarea' },
]

export const bidangVerifikasi = [
  { key: 'tingkat_pemahaman', label: 'Tingkat Pemahaman' },
  { key: 'catatan_verifikasi', label: 'Catatan Hasil Verifikasi' },
  { key: 'status_verifikasi', label: 'Status Verifikasi' },
  { key: 'tanggal_verifikasi', label: 'Waktu Verifikasi (WITA)' },
  { key: 'nip_verifikator', label: 'NIP Verifikator' },
] as const

export const pilihanPemahaman = ['Belum memahami', 'Sebagian memahami', 'Memahami']
export const pilihanStatusVerifikasi = ['Belum diverifikasi', 'Terverifikasi']
