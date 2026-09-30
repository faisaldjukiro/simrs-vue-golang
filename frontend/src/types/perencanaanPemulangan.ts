import type { Pasien } from './domain'
export interface KopPemulangan {
  nama: string
  alamat: string
  kabupaten: string
  propinsi: string
  kontak: string
  email: string
}
export interface PropsPemulangan { token: string; patient: Pasien }
export interface CatatanPemulangan { data: Record<string,string>; nama_petugas: string; bisa_ubah: boolean }
export interface HasilPemulangan { catatan: CatatanPemulangan[]; petugas_login: Record<string,string>; boleh_pilih_petugas: boolean }
export const penilaianPemulangan = [
  {
    "key": "pengaruh_ri_pasien_dan_keluarga",
    "label": "Pengaruh rawat inap terhadap pasien dan keluarga"
  },
  {
    "key": "pengaruh_ri_pekerjaan_sekolah",
    "label": "Pengaruh terhadap pekerjaan / sekolah"
  },
  {
    "key": "pengaruh_ri_keuangan",
    "label": "Pengaruh terhadap keuangan"
  },
  {
    "key": "antisipasi_masalah_saat_pulang",
    "label": "Antisipasi masalah saat pulang"
  },
  {
    "key": "bantuan_diperlukan_dalam",
    "label": "Bantuan yang diperlukan"
  },
  {
    "key": "adakah_yang_membantu_keperluan",
    "label": "Ada yang membantu keperluan"
  },
  {
    "key": "pasien_tinggal_sendiri",
    "label": "Pasien tinggal sendiri"
  },
  {
    "key": "pasien_menggunakan_peralatan_medis",
    "label": "Menggunakan peralatan medis"
  },
  {
    "key": "pasien_memerlukan_alat_bantu",
    "label": "Memerlukan alat bantu"
  },
  {
    "key": "memerlukan_perawatan_khusus",
    "label": "Memerlukan perawatan khusus"
  },
  {
    "key": "bermasalah_memenuhi_kebutuhan",
    "label": "Bermasalah memenuhi kebutuhan"
  },
  {
    "key": "memiliki_nyeri_kronis",
    "label": "Memiliki nyeri kronis"
  },
  {
    "key": "memerlukan_edukasi_kesehatan",
    "label": "Memerlukan edukasi kesehatan"
  },
  {
    "key": "memerlukan_keterampilkan_khusus",
    "label": "Memerlukan keterampilan khusus"
  }
]
export const bantuanPemulangan = ["Menyiapkan Makanan","Edukasi Kesehatan","Makan","Mandi","Diet","Berpakaian","Menyiapkan Obat","Transportasi","Minum Obat"]
