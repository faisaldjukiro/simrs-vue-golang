# Penyimpanan transaksi SIMRS

`DB_*` hanya untuk akun, hak akses, menu, sesi, dan audit aplikasi. `SIMRS_DB_*` digunakan untuk membaca dan melakukan CRUD pada tabel SIMRS yang sudah tersedia. Migration, seeder, dan DDL tidak boleh diarahkan ke SIMRS.

## Perbaikan jalur penyimpanan

| Modul | Tabel tujuan SIMRS |
| --- | --- |
| HAIs | `data_HAIs` (perhatikan huruf besar/kecil di Linux) |
| Edukasi Pasien | `catatan_edukasi` |
| Checklist Pre Operasi | `checklist_pre_operasi` |
| Jadwal Operasi | `booking_operasi` |
| Rujukan Internal | `rujukan_internal_poli`, `rujukan_internal_ranap` |

Transaksi baru tidak disimpan sebagai draft di database konfigurasi. Kegagalan SIMRS harus dikembalikan sebagai error, bukan dialihkan ke database lokal. Edit/hapus memakai sumber dan snapshot asli untuk menghindari perubahan baris yang salah.

Arsip transaksi lokal yang telanjur tersimpan tetap ditampilkan hanya-baca. Tabel arsip bersifat opsional; instalasi baru tidak membuat tabel tersebut. Jadwal Operasi memakai named lock pada koneksi SIMRS, tanpa tabel pengunci lokal. Lock ini mengoordinasikan penulisan SIRAPI, bukan menjamin penguncian aplikasi lain yang tidak memakai lock yang sama.

## Data lama dan deployment

- Tidak ada data pasien yang dipindahkan atau dihapus otomatis dalam perbaikan ini.
- Backup kedua database sebelum rekonsiliasi. Bandingkan nomor rawat, tanggal/jam, identitas petugas, dan seluruh isi terhadap tabel tujuan. Baris identik tidak perlu disalin; konflik harus ditinjau, bukan ditimpa.
- Rujukan Internal masih memiliki aksi Kirim khusus arsip lama dengan konfirmasi dan pemeriksaan konflik. Pembaruan lokal pada aksi itu hanya penanda rekonsiliasi, bukan penyimpanan transaksi baru.
- Arsip modul lainnya memerlukan rekonsiliasi terpisah. Jangan melakukan INSERT SELECT massal tanpa pemeriksaan kunci dan isi.
- Deploy backend dan frontend bersamaan: edit/hapus HAIs dan Edukasi kini menyertakan sumber data.
- Build ulang backend dan restart proses Supervisor yang sesuai; deploy hasil build frontend. Pastikan kredensial SIMRS memiliki SELECT/INSERT/UPDATE/DELETE pada tabel yang diperlukan, tanpa izin DDL.
- Script migration lama yang sebelumnya membuat tabel duplikat sudah dibatasi ke konfigurasi sidebar untuk instalasi baru. Migration yang sudah tercatat tidak dijalankan ulang; tabel lama tetap dipertahankan. Jangan menjalankan reset/fresh atau menghapus riwayat migration untuk menerapkan perbaikan ini.
- Migration 000025 kini hanya `SELECT 1`, tanpa perubahan struktur arsip. Nomor versi dipertahankan supaya riwayat migration tetap dikenali. Arsip Rujukan dari instalasi sangat lama yang belum memiliki kolom `dikirim_pada` perlu penanganan rekonsiliasi tersendiri; kolom tidak ditambahkan otomatis.

Pengujian kode dilakukan tanpa menjalankan CRUD pasien nyata atau migration pada server SIMRS. Verifikasi integrasi selanjutnya harus menggunakan database uji dengan skema SIMRS yang sama.

## Penghapusan empat tabel lokal atas permintaan pengguna

Migration `000036_hapus_tabel_transaksi_lokal` menghapus `sirapi_checklist_pre_operasi`,
`sirapi_jadwal_operasi`, `sirapi_jadwal_operasi_lock`, dan `sirapi_rujukan_internal`
dari `DB_*`. Ini menggantikan kebijakan mempertahankan keempat arsip tersebut
setelah migration dijalankan. Tabel lokal lainnya tidak ikut dihapus.

Backup dahulu dan pastikan tidak ada catatan yang masih perlu direkonsiliasi.
Deploy backend terbaru yang sudah menangani tabel arsip tidak ditemukan.
Dari folder `backend`, periksa target dan seluruh migration pending:

```sh
go run ./cmd/migrate status
go run ./cmd/migrate up
```

`up` menjalankan semua migration pending, bukan hanya 000036. Pastikan target
yang ditampilkan adalah database konfigurasi. Jangan memakai `fresh` atau seed.
Migration ini tidak memindahkan isi tabel ke SIMRS. `down` tidak mengembalikan
tabel maupun data yang dihapus; pemulihan membutuhkan backup.

Migration `000037_hapus_catatan_edukasi_lokal` menambahkan penghapusan
`catatan_edukasi` pada `DB_*` atas permintaan pengguna. Tabel bernama sama pada
`SIMRS_DB_*` tetap dipertahankan. Migration terpisah ini tetap dijalankan meskipun
000036 sudah selesai. Backup dahulu; isi tabel lokal tidak dipindahkan otomatis
dan tidak dapat dipulihkan melalui `down`. Jalankan melalui `cmd/migrate`, bukan
dengan mengeksekusi SQL pada koneksi SIMRS.

Migration `000038_hapus_data_hais_lokal` menghapus `data_hais` pada `DB_*`
atas permintaan pengguna. Tabel `data_HAIs` pada `SIMRS_DB_*` tidak dihapus.
Backup dahulu; data lokal tidak dipindahkan otomatis dan `down` tidak dapat
mengembalikannya. Migration ini terpisah agar tetap berjalan jika 000036 dan
000037 sudah selesai. Gunakan perintah `cmd/migrate` di atas, bukan SQL langsung
pada koneksi SIMRS.

Edukasi Pasien kini memakai satu koneksi repository, yaitu SIMRS, tanpa membaca
arsip `DB_*`. Field dan pilihan metode mengikuti
`RMDataCatatanInformasiEdukasiPasien.java`; ruangan kosong tidak ditolak oleh
form, tetapi constraint tabel SIMRS tetap berlaku. Reset mempertahankan pilihan
petugas, ruangan, dan metode dalam kunjungan yang sama. Error MySQL dikategorikan
dengan kode numerik tanpa menampilkan SQL atau isi catatan pasien. Struktur
database server dan penyebab error simpan sebelumnya belum terverifikasi langsung.
