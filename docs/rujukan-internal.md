# Rujukan Internal

## Cakupan

- `rujukan_internal_poli`: sidebar **Rujuk Internal Poli** hanya pada Rawat Jalan dan IGD/UGD, mengikuti `DlgRujukanPoliInternal.java`.
- `rujukan_internal_ranap`: sidebar **Rujuk Internal Rawat Inap** hanya pada Rawat Inap, mengikuti `DlgRujukanPoliInternalRanap.java`.
- Form memilih poli/unit dan dokter aktif. Rawat Inap menambahkan tanggal dan jam rujukan, dengan default waktu WITA yang bisa disesuaikan.
- Riwayat hanya untuk nomor rawat yang sedang dibuka. Sumber SIRAPI dan Khanza diberi label terpisah.
- Dialog acuan menyediakan simpan. Atas permintaan user, SIRAPI juga menyediakan edit, hapus dengan konfirmasi, dan kirim rujukan lokal lama. Surat konsultasi tidak termasuk.

## Penyimpanan

Pada 12 September 2026, user memberikan izin khusus untuk penyimpanan langsung
ke Khanza serta edit/hapus rujukan internal. Pengecualian terhadap aturan
baca-saja hanya berlaku untuk data pada `rujukan_internal_poli` dan
`rujukan_internal_ranap`, melalui tindakan petugas yang berhak di modul ini.
Tabel lain tetap hanya dibaca. Tidak ada migration atau perubahan struktur
database Khanza. Pengujian CRUD menggunakan database tiruan, bukan pasien nyata.

Rujukan baru langsung disimpan ke tabel Khanza yang sesuai. Tabel lokal
`sirapi_rujukan_internal` menampung rujukan lama yang belum dikirim dan arsip
pengirimannya. Tidak ada pengiriman otomatis atau massal.

- **Edit Khanza:** mengubah rujukan langsung di Khanza.
- **Edit SIRAPI:** hanya mengubah rujukan lokal yang belum dikirim.
- **Hapus:** konfirmasi menampilkan kunjungan, poli, dokter, dan sumber.
  Hanya baris pada sumber yang dipilih yang dihapus; bukan kedua sumber.
- **Kirim ke Khanza:** konfirmasi per baris, lalu salinan lokal diarsipkan
  dengan `dikirim_pada` setelah Khanza berhasil menyimpan. Arsip tidak
  ditampilkan sebagai rujukan lokal tertunda lagi.

Aturan duplikat mengikuti primary key tabel sumber yang diverifikasi:

| Jenis | Tabel sumber | Kunci unik |
| --- | --- | --- |
| Rawat Jalan / IGD | `rujukan_internal_poli` | `no_rawat`, `kd_dokter` |
| Rawat Inap | `rujukan_internal_ranap` | `no_rawat`, `kd_poli` |

Kunci yang sudah ada dengan isi berbeda ditolak, bukan ditimpa dengan upsert.
Riwayat menandai konflik tersebut dan menonaktifkan tombol kirim. Petugas
dapat memeriksa kedua sumber, mengedit bila memang perlu, atau menghapus
salinan lokal yang tidak diperlukan. Perbedaan tanggal/jam Ranap juga
dianggap konflik.

Jika kunci dan seluruh isi identik, pengiriman tidak membuat duplikat: hanya
mengarsipkan salinan lokal. Commit Khanza dilakukan lebih dahulu. Jika
pengarsipan lokal gagal sesudahnya, API mengembalikan sukses dengan
`peringatan: true` yang menjelaskan bahwa rujukan **sudah** tersimpan di
Khanza; kirim ulang untuk menyelesaikan pengarsipan. Dua database ini bukan
satu transaksi atomik. Jika koneksi terputus saat commit Khanza, muat ulang
untuk memeriksa hasil sebelum mencoba lagi.

Edit/hapus memakai transaksi, `SELECT ... FOR UPDATE`, dan membandingkan
snapshot asal dengan isi database. Perubahan petugas lain ditolak sebagai
konflik; petugas harus memuat ulang. Unique key menangani tabrakan tujuan.

Backend memvalidasi nomor rawat, status perawatan, kunjungan batal,
referensi aktif, tanggal/jam, dan sesi login. Semua mutasi ditolak jika
status bayar `Sudah Bayar` atau kunjungan telah memiliki data billing.
Riwayat tetap dapat dilihat. Audit endpoint mengikuti middleware aktivitas
aplikasi yang menyimpan log ke database lokal.

## Aktivasi

```powershell
cd backend
go run ./cmd/migrate up
go run ./cmd/api
```

Migration `000024_add_rujukan_internal` membuat tabel lokal dan konfigurasi
sidebar. Migration `000025_add_rujukan_internal_delivery` menambah penanda
arsip pengiriman pada tabel lokal tersebut. Keduanya hanya menuju `DB_*`.
Tidak perlu menjalankan seeder ulang pada instalasi aktif. Restart API setelah
migration. Akun database Khanza memerlukan izin CRUD pada dua tabel rujukan
di atas; jangan memberikan izin DDL untuk kebutuhan ini.

Rollback 25 menghilangkan penanda arsip sehingga salinan yang pernah dikirim
dapat muncul kembali sebagai tertunda. Rollback 24 menghapus tabel lokal.
Backup dan evaluasi dampaknya dahulu; rollback tidak membatalkan rujukan
yang sudah tersimpan di Khanza.

## API

Semua endpoint memerlukan `Authorization: Bearer <token>`.

| Endpoint | Metode | Kegunaan |
| --- | --- | --- |
| `/api/rujukan-internal-poli?no_rawat=...` | GET | Riwayat Rawat Jalan/IGD |
| `/api/rujukan-internal-ranap?no_rawat=...` | GET | Riwayat Rawat Inap |
| `/api/rujukan-internal-poli` | POST | Simpan rujukan poli langsung di Khanza |
| `/api/rujukan-internal-ranap` | POST | Simpan rujukan ranap langsung di Khanza |
| `/<endpoint>` | PUT | Edit rujukan sesuai sumber |
| `/<endpoint>` | DELETE | Hapus rujukan sesuai sumber |
| `/<endpoint>/kirim` | POST | Kirim satu rujukan lokal ke Khanza |
| `/<endpoint>/referensi/poli?q=...` | GET | Cari poli aktif, minimal 2 karakter |
| `/<endpoint>/referensi/dokter?q=...` | GET | Cari dokter aktif, minimal 2 karakter |

`<endpoint>` adalah `api/rujukan-internal-poli` atau
`api/rujukan-internal-ranap`.

Payload POST Ranap:

```json
{
  "no_rawat": "2026/09/12/000001",
  "kd_dokter": "D001",
  "kd_poli": "P001",
  "tanggal": "2026-09-12",
  "jam": "08:30:00"
}
```

Tanggal/jam tidak disimpan untuk Rawat Jalan/IGD, sesuai tabel acuan Khanza.
PUT menerima `{ "asal": <baris dari GET>, "baru": <payload rujukan> }`.
DELETE dan POST `/kirim` menerima `{ "asal": <baris dari GET> }`.
Kirim seluruh snapshot asal tanpa mengubah `id`, `sumber`, kode, tanggal,
atau jamnya. Nomor rawat tidak dapat dipindahkan melalui edit. GET hanya
menampilkan salinan lokal yang belum diarsipkan dan riwayat Khanza.

Jangan mengirim contoh tersebut ke kunjungan nyata untuk sekadar uji coba.
Akses mengikuti permission modul besar: `igd`, `tindakan_ralan`, atau
`billing_ralan` untuk Poli; `kamar_inap` atau `daftar_pasien_ranap` untuk Ranap.

## Verifikasi implementasi

- Driver SQL tiruan: simpan Ralan/Ranap langsung ke tabel yang tepat,
  kunci edit/hapus, snapshot yang berubah, duplikat, billing, kunjungan batal,
  perubahan lokal tanpa menulis Khanza, daftar konflik, dan urutan commit/arsip.
- Pengiriman diuji untuk baris baru, isi identik, konflik isi, snapshot lokal
  berubah, dan kegagalan arsip setelah commit Khanza.
- Browser dengan API tiruan: Rawat Jalan, IGD, Rawat Inap; desktop dan ponsel;
  terang/gelap; pilih referensi, simpan baru, edit gagal/berhasil, batal
  konfirmasi tanpa mutasi, kirim, hapus, dan batas lebar halaman.
- File uji sementara dibersihkan setelah lulus sesuai panduan backend.
  Tidak ada mutasi data klinis nyata selama pengujian.
- `go test ./...` dan `go vet ./...` masih terhalang dua deklarasi `main`
  pada file lama `backend/cek.go` dan `backend/desc_igd.go`. Package modul
  rujukan dan entry point API diperiksa terpisah; file lama tersebut tidak diubah.
