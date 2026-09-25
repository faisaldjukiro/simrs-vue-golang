# Observasi Ranap

Referensi: `simrs-lama/src/rekammedis/RMDataCatatanObservasiRanap.java` dan tabel
`catatan_observasi_ranap`. Tampilan menggunakan form/tabel klinis CPPT.

## Aktivasi sidebar yang sudah ada

Dari backend jalankan `go run ./cmd/migrate up`, restart backend, dan refresh
frontend. Migration 000030 hanya memetakan kode sidebar lokal **Observasi Ranap**
menjadi `observasi_ranap`. Tidak menambah menu, mengubah status aktif/urutan/modul,
atau memigrasikan skema Khanza. Tidak perlu seed ulang. Jika menu telah diubah
namanya, sesuaikan kode sidebar melalui pengaturan menu yang sudah ada.

## Data dan proses

- Tanggal perawatan, jam rawat WITA, GCS (10 karakter), TD (8), HR (5), RR (5),
  suhu (5), SpO2 (3), dan NIP petugas (20).
- Tanda vital berupa teks dan boleh kosong mengikuti Java. Tidak ada nilai
  normal bawaan, konversi otomatis, atau interpretasi klinis.
- Simpan/edit/hapus langsung ke Khanza, mengikuti pola modul klinis sebelumnya.
- Primary key: `no_rawat, tgl_perawatan, jam_rawat`. Konflik tanggal/jam mendapat
  respons 409; edit/hapus memakai snapshot lengkap dan transaksi untuk menghindari
  penimpaan perubahan bersamaan.
- Edit/hapus hanya oleh username yang sama dengan NIP pencatat atau admin `*`.
  Endpoint tetap memakai autentikasi dan akses modul pelayanan pasien.
- Riwayat dibatasi kunjungan aktif; seluruh catatan kunjungan diambil tanpa limit
  tersembunyi. Filter tanggal inklusif, pencarian, dan pagination ada di frontend.
- Cetak browser A4 landscape mengikuti hasil filter, bukan Jasper legacy.
  Isi pasien dirender memakai textContent agar tidak mengeksekusi HTML.
- Referensi petugas memakai endpoint bersama Checklist Pre Operasi; audit
  mutasi mengikuti middleware aplikasi lokal.

## API

Bearer token wajib. Endpoint `/api/observasi-ranap`:

- GET `?no_rawat=...` mengembalikan `{catatan: [...]}`.
- POST `{no_rawat, data}` untuk menyimpan.
- PUT `{no_rawat, data, asli}` untuk mengedit.
- DELETE `{no_rawat, asli}` untuk menghapus.

`data` berisi sembilan kolom selain no_rawat. `asli` adalah snapshot `data`
dari GET. Nama kolom sama dengan tabel Khanza.

Pengujian pengembangan tidak melakukan mutasi pada database Khanza nyata.
Verifikasi kunjungan uji, hak akses petugas, konflik waktu, edit/hapus, dan hasil
cetak sebelum digunakan pada pelayanan produksi.
