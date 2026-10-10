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

## Foto Edukasi Pasien

Kolom `catatan_edukasi.foto` (`varchar(255)`, nullable) ditambahkan pengguna
langsung pada SIMRS. Aplikasi tidak menjalankan DDL atau migration untuk kolom
ini. Nilai NULL dibaca sebagai foto kosong. Deploy backend dan frontend bersama.

Form Edukasi menyediakan foto petugas dan foto penerima dari kamera, masing-masing maksimal 10 MB dan
40 megapiksel, pratinjau, penggantian foto, dan tombol Lihat Foto di riwayat.
Edit tanpa file pengganti mempertahankan foto sebelumnya. Reset atau pindah
pasien membuang pilihan file yang belum disimpan.

Upload memakai `BERKAS_DIGITAL_UPLOAD_URL` dan `BERKAS_DIGITAL_LOKASI_PREFIX`
yang sama dengan Berkas Digital. Default endpoint adalah
`<SIMRS_WEB_BASE_URL>/berkasrawat/uploadsep.php`, dengan prefix `pages/upload`.
Backend mengirim multipart `file`, `no_rawat`, dan `kode=edukasi`, serta
mengharuskan respons sukses `UPLOAD_BERHASIL`. Nama file unik dibuat server;
kolom `foto` berisi lokasi relatif, bukan isi gambar. Tidak membuat baris
tambahan di `berkas_digital_perawatan`.

Upload dijalankan setelah validasi akses, snapshot edit, dan penulisan catatan
dalam transaksi InnoDB. Upload gagal membatalkan perubahan catatan. File pada
server berkas tidak ikut transaksi database: jika upload berhasil lalu database
gagal, API meminta petugas memuat ulang riwayat sebelum mencoba lagi. File lama
saat penggantian atau penghapusan catatan tidak dihapus secara fisik otomatis.

Pengujian integrasi melalui Postman pada lingkungan uji:

1. Gunakan Bearer token dan `GET {{api_url}}/api/edukasi-pasien?no_rawat={{no_rawat}}`.
2. Untuk foto baru, kirim `POST {{api_url}}/api/edukasi-pasien` dengan Body
   **form-data**: `foto` dan `foto_penerima` bertipe File, serta `payload` bertipe Text berisi JSON
   `{ "no_rawat": "<nomor rawat uji>", "data": { "tgl_perawatan": "2026-10-08",
   "jam_rawat": "10:00:00", "nip": "<kode petugas uji>", "kd_ruangan": "",
   "metode": "Lisan", "durasi": "10 menit", "materi": "Materi uji",
   "penerima": "Pasien", "nama_penerima": "Penerima Uji", "keterangan": "", "foto": "", "foto_penerima": "" } }`.
   Jangan mengisi header Content-Type manual; Postman membuat boundary multipart.
3. Untuk mengganti foto, gunakan `PUT` pada URL yang sama; sertakan
   `sumber: "SIMRS"` dan `asli` yang berisi seluruh `data` hasil GET, termasuk
   `foto` dan `foto_penerima`. `data` berisi nilai yang diperbarui dan lokasi foto lama.
4. GET kembali, periksa `data.foto`, `data.foto_penerima`, `foto_url`, dan `foto_penerima_url`, lalu buka foto. Uji juga
   edit JSON tanpa file pengganti, file bukan gambar, dan server upload gagal.

Verifikasi otomatis memakai driver SQL dan server HTTP tiruan; belum menguji
upload ke server SIMRS nyata atau melakukan mutasi data pasien nyata.

## Asesmen dan verifikasi Edukasi Pasien

Form menggunakan 14 kolom asesmen/verifikasi yang sudah ditambahkan pengguna
pada `catatan_edukasi` SIMRS; `materi` dan `catatan_verifikasi` bertipe TEXT.
Tidak ada migration tambahan. Deploy backend dan frontend bersamaan.

Asesmen mencatat kemampuan membaca, pendidikan penerima, bahasa, motivasi,
kesediaan menerima informasi, hambatan emosional, keterbatasan fisik/kognitif,
dan nilai budaya. Isian yang belum dikaji boleh kosong. Nilai NULL pada catatan
lama ditampilkan sebagai belum dicatat, bukan otomatis terverifikasi.

Status baru default `Belum diverifikasi`. Pemilihan `Terverifikasi` memerlukan
seluruh asesmen, materi, penerima, catatan hasil verifikasi, salah satu bukti (paraf atau foto),
tingkat pemahaman, waktu WITA yang tidak mendahului edukasi, dan petugas
verifikator. Validasi berlaku pada frontend dan backend. Petugas biasa memakai identitas login; administrator dapat memilih
petugas. Status terverifikasi mencatat bahwa penilaian sudah dilakukan, sehingga
hasil `Belum memahami` tetap diperbolehkan dan dapat disertai catatan tindak lanjut.
Pengubahan status menjadi belum diverifikasi mengosongkan waktu/NIP verifikator.

Riwayat menampilkan ringkasan verifikasi dan tombol Detail Edukasi. Cetak memakai
tabel per catatan agar semua asesmen, hasil verifikasi, dan nama petugas terbaca.
Upload foto tetap tersedia. Tanggal verifikasi kosong disimpan sebagai SQL NULL.
Edit/hapus menyertakan seluruh field terbaru pada snapshot `asli`.

Endpoint Postman tetap `GET/POST/PUT/DELETE {{api_url}}/api/edukasi-pasien`, dengan
Bearer token dan bentuk payload yang sama. Tambahkan field baru pada `data`;
untuk status `Terverifikasi`, lengkapi asesmen, materi, penerima, catatan hasil
verifikasi, dan salah satu bukti; kirim misalnya `tingkat_pemahaman: "Sebagian memahami"`,
`tanggal_verifikasi: "2026-10-10 10:00:00"`, dan `nip_verifikator` milik petugas uji.
Gunakan waktu setelah tanggal/jam edukasi. Uji GET, edit, dan foto pada lingkungan
uji; pemeriksaan otomatis dilakukan dengan data tiruan tanpa mutasi pasien nyata.

Paraf digambar langsung pada kanvas aplikasi dan disimpan bersama catatan pada
kolom `paraf_petugas` dan `paraf_penerima` (MEDIUMTEXT ASCII dengan collation ascii_bin), sebagai
data URL PNG 720 x 240 piksel, maksimal 256 KiB termasuk encoding. Backend
memeriksa format PNG, ukuran, dan menolak gambar kosong. Paraf merupakan gambar
goresan petugas dan penerima edukasi; fitur ini bukan tanda tangan elektronik tersertifikasi.
Nama penerima disimpan pada `nama_penerima` (VARCHAR(100), nullable),
terpisah dari `penerima` yang menyimpan hubungan/kategori penerima.
Kedua paraf ditampilkan pada detail dan cetakan dengan nama masing-masing tanpa NIP.
Paraf disamarkan pada audit aplikasi dan tidak dicari sebagai teks riwayat.

Administrator menjalankan query berikut secara manual di database SIMRS bila
kolom belum ada (jangan dimasukkan ke migration aplikasi):

```sql
ALTER TABLE catatan_edukasi
  ADD COLUMN paraf_petugas MEDIUMTEXT
  CHARACTER SET ascii COLLATE ascii_bin NULL
  AFTER nip_verifikator;
```

Tambahan untuk penerima edukasi, jalankan hanya bila kedua kolom belum ada:

```sql
ALTER TABLE catatan_edukasi
  ADD COLUMN nama_penerima VARCHAR(100) NULL,
  ADD COLUMN paraf_penerima MEDIUMTEXT
    CHARACTER SET ascii COLLATE ascii_bin NULL;
```

Aplikasi hanya memeriksa keberadaan kolom, tidak menjalankan
DDL. Sebelum kolom tersedia, riwayat tetap dapat dibaca dan catatan belum
diverifikasi tetap dapat disimpan. Verifikasi dengan dua foto memerlukan
`foto_penerima` dan `nama_penerima`, tanpa mewajibkan kolom paraf.
Setelah kolom ditambahkan, muat ulang riwayat untuk mengaktifkan kotak paraf.
Deploy backend dan frontend bersama.

Perubahan isi/nama penerima/petugas/ruangan/verifikator di form membatalkan kedua paraf
sehingga harus digambar ulang. Menggambar paraf kedua tidak membatalkan paraf pertama.
Backend menolak penggunaan paraf snapshot
yang sama untuk perubahan catatan. Snapshot edit/hapus menyertakan
`paraf_petugas`, `nama_penerima`, dan `paraf_penerima` agar tidak menimpa perubahan pengguna lain.

Form memilih jenis bukti paraf atau foto. Mengganti pilihan mengosongkan bukti
lain pada form; snapshot asli tetap utuh hingga simpan. Catatan belum diverifikasi
boleh tanpa bukti atau menyimpan paraf yang belum lengkap. Verifikasi memerlukan
tepat satu jenis bukti: dua paraf (petugas dan penerima, dengan nama penerima)
atau dua foto (petugas dan penerima, dengan nama penerima). Backend
menolak paraf bersama lokasi foto maupun bersama file upload baru. Foto baru
dihitung sebagai bukti saat validasi; kegagalan upload tetap membatalkan transaksi.
File foto lama tidak dihapus fisik ketika catatan beralih ke paraf.
Cetak hanya menampilkan bukti yang tersimpan tanpa ruang paraf kosong untuk
pilihan foto. Catatan lama dengan dua bukti perlu diedit untuk memilih salah satu
sebelum dicetak ulang. Foto bukti yang gagal dimuat membatalkan cetak.

Foto petugas tetap memakai kolom `foto`; foto penerima memakai kolom baru berikut.
Jalankan manual di Navicat hanya jika kolom belum ada (bukan migration aplikasi):

```sql
ALTER TABLE catatan_edukasi ADD COLUMN foto_penerima VARCHAR(255) NULL;
```

Form foto memakai kamera langsung dengan `getUserMedia`, tanpa pemilih file.
Buka kamera, izinkan akses, ambil foto, dan periksa pratinjau sebelum menyimpan.
Kamera memerlukan HTTPS atau localhost; alamat IP LAN melalui HTTP tidak mendukung
akses kamera. Tombol ganti kamera meminta kamera depan/belakang jika perangkat mendukung.
Kamera dihentikan setelah foto diambil, saat ditutup, form disembunyikan, berganti
jenis bukti/pasien, dan saat meninggalkan halaman. Hasil kamera berupa JPEG,
sisi terpanjang maksimal 1600 piksel. Pengambilan ulang yang dibatalkan tetap
mempertahankan foto sebelumnya. Maksimal satu kamera aktif pada form.

Kedua foto diunggah berurutan ke server berkas yang sama; lokasi keduanya diperbarui
dalam satu transaksi catatan. Jika upload kedua gagal, transaksi database dibatalkan.
Foto pertama yang sudah terkirim tidak dapat dibatalkan secara transaksional pada
server berkas; API memberi pesan khusus dan tidak mengklaim catatan berhasil.
Catatan lama dengan satu foto tetap terbaca dan dapat dicetak dengan keterangan
foto penerima belum dilampirkan. Saat disimpan sebagai terverifikasi, kedua foto wajib lengkap.
