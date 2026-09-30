# Perencanaan Pemulangan

Referensi: `simrs-lama/src/rekammedis/RMPerencanaanPemulangan.java`.

- CRUD langsung pada 34 kolom tabel SIMRS `perencanaan_pemulangan`.
- Nama kolom `memerlukan_keterampilkan_khusus` mengikuti ejaan tabel lama.
- Diagnosis maksimal 50 karakter, alasan masuk 150, setiap keterangan 100,
  dan nama pasien/keluarga 50, mengikuti batas input Java.
- Satu catatan per nomor rawat; duplikasi ditolak, bukan ditimpa.
- Edit/hapus memakai snapshot lengkap dan dibatasi petugas pencatat/admin.
- Halaman berada dalam kunjungan pasien Rawat Inap. Cetak menggunakan browser.
- Pengambilan tanda tangan/foto keluarga, antrean perangkat tanda tangan,
  laporan Jasper, dan daftar lintas pasien belum diimplementasikan.
- Migrasi 000039 hanya konfigurasi sidebar DB_*, tidak menyentuh tabel SIMRS.

Pemeriksaan dilakukan tanpa mutasi data pasien nyata. Uji simpan, muat ulang,
edit, dan hapus hanya pada kunjungan uji yang disetujui rumah sakit.
