# Lanjutan Risiko Jatuh Dewasa

Referensi: `simrs-lama/src/rekammedis/RMPenilaianLanjutanRisikoJatuhDewasa.java`.

Sidebar `lanjutan_risiko_jatuh_dewasa` ditampilkan untuk Rawat Inap. Jalankan
migration `000028` dengan `go run ./cmd/migrate up` dari folder backend,
kemudian restart backend. Migration hanya mengubah konfigurasi sidebar di
database aplikasi `DB_*`, bukan skema Khanza. Tidak perlu menjalankan ulang seed.

## Penyimpanan

Atas permintaan pengguna, simpan/edit/hapus langsung menggunakan koneksi Khanza
dan tabel `penilaian_lanjutan_resiko_jatuh_dewasa`, bukan draft lokal.
Primary key mengikuti Khanza: `no_rawat, tanggal`.

- GET `/api/risiko-jatuh-dewasa?no_rawat=...`: riwayat kunjungan dan pilihan Morse.
- POST `/api/risiko-jatuh-dewasa`: `{no_rawat, data}`.
- PUT `/api/risiko-jatuh-dewasa`: `{no_rawat, data, asli}`.
- DELETE `/api/risiko-jatuh-dewasa`: `{no_rawat, asli}`.

`data`/`asli` memakai nama kolom Khanza selain `no_rawat`. `asli` adalah snapshot
lengkap dari GET. PUT/DELETE memeriksa snapshot dalam transaksi agar perubahan
bersamaan tidak tertimpa. Riwayat hanya boleh diubah oleh username yang sama
dengan NIP pencatat atau administrator dengan permission `*`.
API tetap membutuhkan Bearer token dan akses modul pelayanan pasien.
Pencarian petugas memakai endpoint referensi petugas Checklist Pre Operasi yang
sudah ada. Audit mutasi mengikuti middleware aplikasi.

Enam pilihan dan bobot Morse mengikuti Java; backend menghitung ulang skor,
bukan mempercayai nilai kiriman browser. Risiko rendah <25, sedang 25–44,
tinggi >=45 mengikuti implementasi Java. Hasil skrining dan saran wajib diisi
petugas (masing-masing maksimum 200 karakter), tidak dihasilkan otomatis.
Tanggal input menggunakan WITA. Form baru tidak otomatis memilih jawaban.

## Verifikasi

144 kombinasi pilihan dan batas kategori skor telah diuji secara lokal.
Type-check/build frontend serta test/kompilasi dan vet `./cmd/... ./internal/...`
lulus. File unit test sementara dihapus sesuai panduan repository.
Belum dilakukan uji simpan/edit/hapus ke database Khanza atau inspeksi visual
browser. Verifikasi dengan kunjungan uji dan akun petugas sebelum pemakaian
produksi, termasuk konflik tanggal, hak edit, dan perubahan bersamaan.
