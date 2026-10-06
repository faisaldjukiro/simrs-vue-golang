export interface KlaimMonitoring {
  noSEP?: string
  noFPK?: string
  peserta?: { noMR?: string; noKartu?: string; nama?: string }
  jenisPelayanan?: string
  poli?: string
  kelasRawat?: string
  Inacbg?: { kode?: string; nama?: string }
  tglSep?: string
  tglPulang?: string
  status?: string
  biaya?: Record<string, string | number | null>
  dokter_simrs?: {
    no_rawat: string
    nama_dokter: string
    sumber: string
    keterangan: string
  }
}

export interface FilterKlaim {
  tanggal_mulai: string
  tanggal_selesai: string
  jenis_pelayanan: string
  status_klaim: string
}

export interface HasilMonitoringKlaim {
  response?: { klaim?: KlaimMonitoring[] }
  peringatan_simrs?: string
  periode?: {
    tanggal_mulai: string
    tanggal_selesai: string
    jumlah_tanggal_berhasil: number
  }
  tanggal_tanpa_data?: string[]
  tanggal_gagal?: { tanggal: string; jenis_pelayanan?: string; pesan: string }[]
}
