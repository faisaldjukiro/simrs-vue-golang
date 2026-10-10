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
  foto_penerima_url: string
}

export interface HasilEdukasi {
  paraf_tersedia: boolean
  paraf_penerima_tersedia: boolean
  foto_penerima_tersedia: boolean
  nama_penerima_tersedia: boolean
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

export interface BidangAsesmen {
  key: string
  label: string
  maxlength: number
  jenis?: string
  pilihan?: string[]
}

export const bidangAsesmen: BidangAsesmen[] = [
  { key: 'kemampuan_membaca', label: 'Kemampuan Membaca', maxlength: 30, pilihan: ['Bisa membaca', 'Perlu bantuan', 'Tidak bisa membaca'] },
  { key: 'tingkat_pendidikan', label: 'Pendidikan Penerima Edukasi', maxlength: 30, pilihan: ['Tidak sekolah', 'SD', 'SLTP', 'SLTA', 'D-I/II', 'D-III', 'D-IV/S1', 'S2', 'S3'] },
  { key: 'bahasa', label: 'Bahasa yang Dipahami', maxlength: 50, pilihan: ['Indonesia', 'Gorontalo', 'Bahasa isyarat'] },
  { key: 'motivasi', label: 'Motivasi Menerima Edukasi', maxlength: 100, pilihan: ['Baik', 'Cukup', 'Kurang', 'Tidak ada motivasi'] },
  { key: 'kesediaan_menerima', label: 'Kesediaan Menerima Informasi', maxlength: 30, pilihan: ['Bersedia', 'Belum bersedia', 'Tidak bersedia'] },
  { key: 'hambatan_emosional', label: 'Hambatan Emosional', maxlength: 255, jenis: 'textarea', pilihan: ['Tidak ada', 'Cemas', 'Gangguan emosi'] },
  { key: 'keterbatasan_fisik', label: 'Keterbatasan Fisik', maxlength: 255, jenis: 'textarea', pilihan: ['Tidak ada', 'Nyeri', 'Penurunan mobilitas fisik', 'Gangguan pendengaran', 'Gangguan penglihatan', 'Gangguan bicara'] },
  { key: 'keterbatasan_kognitif', label: 'Keterbatasan Kognitif', maxlength: 255, jenis: 'textarea', pilihan: ['Tidak ada', 'Gangguan kognitif'] },
  { key: 'nilai_budaya', label: 'Nilai Budaya / Keyakinan', maxlength: 255, jenis: 'textarea', pilihan: ['Tidak ada', 'Ingin dirawat oleh sesama gender'] },
]

export const pilihanPelaksanaan: BidangAsesmen[] = [
  { key: 'durasi', label: 'Durasi', maxlength: 30, pilihan: ['5 menit', '10 menit', '15 menit', '20 menit', '30 menit', '45 menit', '60 menit'] },
  { key: 'penerima', label: 'Penerima Edukasi', maxlength: 30, pilihan: ['Pasien', 'Ibu', 'Bapak', 'Anak kandung', 'Suami', 'Istri', 'Keluarga'] },
]

export const pilihanMateri = [
  'Kondisi kesehatan dan diagnosis penyakit',
  'Hak dan kewajiban pasien', 'Hasil asesmen medis', 'Proses penyakit', 'Hasil asesmen perawat',
  'Perawatan pre dan post operasi', 'Penggunaan obat', 'Program diet dan nutrisi', 'Manajemen nyeri',
  'Pencegahan infeksi di rumah sakit', 'Tindakan / prosedur kedokteran', 'Pencegahan risiko jatuh',
  'Asuhan lanjutan', 'Teknik rehabilitasi', 'Penggunaan peralatan medis secara efektif dan aman',
  'Edukasi kolaborasi', 'Pelayanan spiritual', 'Pencegahan bunuh diri', 'Keterbatasan fasilitas rumah sakit',
  'Hasil asuhan dan pengobatan yang diharapkan (termasuk yang tidak diharapkan)',
  'Manajemen pencegahan risiko dekubitus',
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
