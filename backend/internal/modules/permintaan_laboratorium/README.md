# Permintaan laboratorium

Referensi: `simrs-lama/src/permintaan/DlgPermintaanLaboratorium.java`.

## Kategori dan kontrak API

Endpoint tetap `/api/permintaan-laboratorium`, tanpa perubahan prefix.
GET data, pencarian tindakan, detail tindakan, dan DELETE menerima query
`kategori=PK|PA|MB`. POST/PUT menerima `kategori` dalam JSON.
Kategori yang tidak dikirim tetap dianggap PK untuk kompatibilitas klien lama.
Kategori selain ketiganya ditolak. Klien harus mengirim kategori yang sama
dengan riwayat yang akan diedit/dihapus.

| Kategori | Nomor | Header | Pemeriksaan | Detail |
| --- | --- | --- | --- | --- |
| PK | PK + tanggal + 4 digit | permintaan_lab | permintaan_pemeriksaan_lab | permintaan_detail_permintaan_lab |
| PA | PA + tanggal + 4 digit | permintaan_labpa | permintaan_pemeriksaan_labpa | Tidak diinput dari form permintaan |
| MB | MB + tanggal + 4 digit | permintaan_labmb | permintaan_pemeriksaan_labmb | permintaan_detail_permintaan_labmb |

Input PA memakai objek `spesimen`: `pengambilan_bahan`, `diperoleh_dengan`,
`lokasi_jaringan`, `diawetkan_dengan`, `pernah_dilakukan_di`,
`tanggal_pa_sebelumnya`, `nomor_pa_sebelumnya`, `diagnosa_pa_sebelumnya`.
Tanggal pengambilan wajib; tanggal PA sebelumnya wajib jika tempat PA sebelumnya
diisi. Tanpa tempat PA sebelumnya, tanggal riwayat disimpan sebagai tanggal kosong
Khanza (`0000-00-00`).

## Tarif dan billing

- Kategori dan status aktif tindakan selalu diperiksa.
- Cara bayar dan kelas mengikuti `set_tarif.cara_bayar_lab` / `kelas_lab`.
  Jika tidak ada baris konfigurasi, default Yes/Yes seperti dialog lama.
  Kegagalan query konfigurasi tidak dianggap izin menampilkan seluruh tarif.
- Penjamin menggunakan `reg_periksa.kd_pj` asli, tidak dipetakan menjadi BPJ/A09.
- Kelas rawat inap memakai kamar dengan `stts_pulang='-'`, termasuk rujukan
  kunjungan ibu pada `ranap_gabung`. Tindakan dengan penjamin/kelas `-` tetap
  mengikuti pengecualian umum Khanza.
- Validasi tarif juga diterapkan pada simpan/edit, bukan hanya pencarian.
- `LAB_AKTIFKAN_BILLING_PARSIAL=no` adalah default. Set `yes` hanya jika sesuai
  `AKTIFKANBILLINGPARSIAL` pada instalasi Khanza. Pengecualian hanya berlaku
  untuk permintaan baru dengan penjamin terdaftar di `set_input_parsial`.
  Edit/hapus tetap terkunci oleh billing. Kunjungan batal selalu terkunci.
- Permintaan yang sudah diterima/diproses/dibayar tidak boleh diubah/dihapus.

## Batas adaptasi dan verifikasi

- Form web menyimpan satu kategori per proses. Dialog Java dapat mengirim
  beberapa kategori dalam satu aksi. Ini belum merupakan kesetaraan 100%.
- Cetak web adalah lembar permintaan dari data tersimpan, bukan replika Jasper
  atau tanda tangan elektronik Khanza.
- Tidak ada migration atau perubahan pada `simrs-lama`. Jangan menjalankan
  pengujian mutasi terhadap database SIMRS produksi.
- Pengujian otomatis memakai driver database tiruan, tidak membaca `.env`
  atau mengirim transaksi ke server SIMRS.
- Sebelum dipakai operasional, verifikasi di lingkungan uji yang disetujui:
  pencarian tiap kategori, kombinasi tarif Yes/No, rawat jalan/rawat inap/gabung,
  spesimen PA, riwayat, cetak, edit/hapus sebelum diproses, serta penolakan
  setelah sampel diterima atau billing terkunci. Cocokkan struktur tabel dengan
  instalasi Khanza yang dipakai; build saja tidak membuktikan kompatibilitas DB.
