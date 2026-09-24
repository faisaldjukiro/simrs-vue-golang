export interface BidangOperasi {
  kode: string
  label: string
  jenis: string
  batas: number
  wajib: boolean
  pilihan?: string[]
  nilai_ke?: string
}

export interface HasilPendukung {
  kolom: string[]
  catatan: Record<string, string>[]
  form: BidangOperasi[] | null
  pesan_kunci: string
}

export const modulInputOperasi = new Set([
  'penilaian_pre_operasi',
  'penilaian_pre_anestesi',
  'penilaian_pre_induksi',
  'transfer_pasien_antar_ruang',
  'signin_sebelum_anestesi',
  'timeout_sebelum_insisi',
  'signout_sebelum_menutup_luka',
  'checklist_post_operasi',
  'laporan_operasi',
  'skor_aldrette_pasca_anestesi',
  'skor_steward_pasca_anestesi',
  'skor_bromage_pasca_anestesi',
])
