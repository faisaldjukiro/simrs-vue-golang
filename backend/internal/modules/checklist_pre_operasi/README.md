# Checklist Pre Operasi

Referensi: `simrs-lama/src/rekammedis/RMChecklistPreOperasi.java`.

- Sidebar: `checklist_pre_operasi`, Rawat Inap.
- API: `/api/checklist-pre-operasi`, mengikuti permission modul pelayanan pasien.
- Form mengikuti 27 kolom referensi: nomor rawat, tanggal/jam, dan 25 bidang
  checklist. Batas SN/CN 25 karakter, tindakan 50, keterangan 20.
- Pilihan klinis kosong saat membuat catatan: petugas wajib memilih hasil
  konfirmasi. Dokter dan kedua petugas wajib dipilih dari referensi Khanza.
- Data baru disimpan sebagai snapshot JSON di tabel lokal
  `sirapi_checklist_pre_operasi`. Tidak menulis ke SIMRS lama.
- Riwayat Khanza berasal dari `checklist_pre_operasi`, hanya baca. Kegagalan
  pembacaan ditampilkan sebagai peringatan, bukan dianggap tidak ada riwayat.
- Edit/hapus lokal hanya untuk pembuat atau pengguna permission `*`.
  Versi diperiksa untuk mencegah perubahan saling menimpa.
- Hapus menandai `deleted_at`; catatan tetap tersimpan untuk penelusuran.
  Kombinasi nomor rawat dan waktu tetap unik, termasuk catatan yang dihapus.
- Cetak detail melalui browser, bukan salinan template cetak desktop Khanza.

## Aktivasi

Jalankan `go run ./cmd/migrate up` dari folder backend menggunakan konfigurasi
DB lokal yang diizinkan. Migration 000026 menambahkan penyimpanan dan pemetaan
sidebar. Restart backend, muat ulang frontend.

Rollback hanya menonaktifkan sidebar dan mempertahankan rekam klinis.
Migration dapat diterapkan ulang tanpa menghapus isi tabel.

## Pemeriksaan

Validasi field/opsi/batas panjang/tanggal telah diuji tanpa database.
Koneksi riwayat Khanza dan CRUD database perlu diverifikasi di lingkungan
pengujian setelah migration. Jangan menggunakan pasien nyata untuk tes.
