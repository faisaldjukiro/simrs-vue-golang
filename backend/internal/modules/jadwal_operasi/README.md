# Jadwal Operasi

Referensi: `simrs-lama/src/permintaan/DlgBookingOperasi.java` dan
`keuangan/DlgCariDaftarOperasi.java`.

Aktifkan dengan migration 000027 pada database aplikasi DB_*.
Sidebar `jadwal_operasi` menggunakan konfigurasi Jadwal Operasi Rawat Inap.
API `/api/jadwal-operasi` mengikuti permission modul pelayanan pasien.

## Cakupan

- Jadwal untuk kunjungan pasien aktif: tanggal, mulai/selesai, paket operasi,
  operator, ruang OK, status, dokter anestesi, perawat.
- Status: Permintaan, Menunggu, Proses Operasi, Selesai.
- Paket aktif disaring mengikuti kelas dan cara bayar di set_tarif, termasuk
  kelas kunjungan induk ranap_gabung. Dokter anestesi/perawat berupa teks,
  seperti pada referensi.
- Riwayat pasien berasal dari booking_operasi Khanza. Kegagalan pembacaan
  diberi peringatan, tidak dianggap riwayat kosong.
- Atas izin eksplisit user, jadwal baru langsung INSERT ke booking_operasi Khanza
  (10 kolom sesuai referensi). Tidak ada fallback/salinan lokal apabila gagal.
- Jadwal lokal lama hanya-baca apabila tabel arsip masih ada. Mutasi jadwal
  selalu memakai tabel SIMRS; tidak ada pengiriman massal data lokal.
- Edit/hapus Khanza telah diizinkan user melalui PUT/DELETE /khanza.
  Snapshot asli seluruh kolom wajib disertakan. Perubahan data/target ambigu
  ditolak; hapus permanen membutuhkan konfirmasi di UI. Billing tetap diperiksa.
- Billing terverifikasi atau kunjungan Batal mengunci mutasi.
- Konflik versi mencegah perubahan saling menimpa.

## Waktu dan bentrok

Tanggal/jam menggunakan WITA. Mengikuti Java, jam mulai 00:00:00 melewati
pemeriksaan bentrok. Format waktu tetap divalidasi, tanpa menambahkan aturan
urutan jam yang tidak ada di dialog lama.

Pemeriksaan bentrok sama dengan query desktop: tanggal/ruang sama, nomor rawat
berbeda, dan jam mulai jadwal lain BETWEEN mulai dan selesai (inklusif).
Ini bukan deteksi semua kemungkinan tumpang tindih interval. Duplikasi
paket/tanggal/jam mulai tetap ditolak sesuai primary key tabel.

Simpan memakai named lock pada koneksi SIMRS, tanpa tabel pengunci lokal.
Lock ini tidak mengoordinasikan aplikasi lain yang tidak memakai lock yang sama.

Edit mengikuti delapan kolom Java; dokteranastesi dan perawat dipertahankan,
dan kedua input dinonaktifkan saat edit. Snapshot seluruh kolom tetap diperiksa.
Batas booking_operasi yang dikonfirmasi pengguna: kode_paket 15, kd_ruang_ok 3,
kd_dokter 20, dokteranastesi/perawat masing-masing 50 karakter.

POST jadwal baru tidak lagi ditandai sebagai arsip SIRAPI oleh handler.

## Template laporan

GET `/api/jadwal-operasi/template-laporan?q=...` mencari tabel SIMRS
`template_laporan_operasi`. Pengguna memilih lalu menerapkan template secara
eksplisit; penggantian isian yang sudah ada membutuhkan konfirmasi.
POST/PUT pendukung laporan menerima `simpan_template: true` untuk menyimpan
laporan dan template dalam transaksi yang sama. Dokter mengikuti akun dokter
login, atau operator jadwal bila akun bukan dokter, seperti Java.
ID memakai tanggal laporan/jadwal ditambah urutan empat digit; bentrok ID
membatalkan transaksi, tidak menimpa template. Tidak ada DDL/migration SIMRS.
Template yang melebihi batas form laporan harus disesuaikan oleh pengguna;
teks tidak dipotong otomatis. Skema laporan_operasi belum dikonfirmasi ulang.

## Batas modul

Ini sidebar jadwal per pasien, bukan dashboard jadwal semua pasien.
Workspace memuat Jadwal Operasi dan 18 tab pendukung sesuai urutan desktop.
Pilih jadwal terlebih dahulu. Ganti jadwal mengosongkan cache form pendukung
setelah konfirmasi; perubahan/hapus jadwal terpilih juga membatalkan konteks.
Operator diteruskan ke Permintaan Lab dan Resep, tindakan/operator ke Checklist
Pre Operasi, tanggal/jam selesai ke Resep. Jawaban checklist tidak diisi otomatis.

Empat modul yang tersedia (Lab, Riwayat Perawatan, Checklist Pre Operasi,
Resep) tetap menggunakan halaman/validasi masing-masing. Dua belas tab telah
ditambahkan form, validasi server, simpan/edit/hapus langsung ke Khanza:

- Penilaian Pre Induksi, Pre Operasi, Pre Anestesi.
- Sign-In, Time-Out, Sign-Out, Checklist Post Operasi.
- Laporan Operasi.
- Skor Aldrette, Steward, Bromage.
- Transfer Antar Ruang (dokumen serah-terima, bukan mutasi bed).

Endpoint GET
`/api/jadwal-operasi/pendukung?jenis=...&no_rawat=...&tanggal=YYYY-MM-DD`
menggunakan whitelist tabel dari referensi. Laporan Operasi disaring tanggal
jadwal; riwayat lainnya seluruh kunjungan. POST/PUT/DELETE pada path yang sama
menerima `jenis`, `no_rawat`, `jadwal`, `data`, dan `asli` (PUT/DELETE).
Sumber tabel/kolom selalu whitelist server, bukan identifier dari request.
Tanggal wajib valid; skor dihitung ulang server; pilihan klinis tidak dipilih
otomatis. Dokter/petugas diverifikasi terhadap tabel referensi.

Mutasi berada dalam transaksi dengan lock reg_periksa per kunjungan, billing
dan pembatalan diperiksa ulang, jadwal terpilih harus masih ada di Khanza.
Edit/hapus cocokkan snapshot seluruh field dan dibatasi satu baris. Hak ubah
mengikuti dokter/petugas pada catatan atau admin `*`; username akun SIRAPI
harus sama dengan kode pegawai Khanza, bukan dibandingkan nama tampilan.
Transfer baru harus oleh salah satu petugas serah-terima atau admin.
Laporan baru yang sudah ada pada hari yang sama ditolak; gunakan Edit agar
catatan tidak dihapus/ditimpa diam-diam seperti pola delete-insert lama.

**Masih baca saja:** Kamar Inap dan Tagihan Operasi/VK. Tagihan menampilkan
catatan tabel operasi, belum rincian obat/jurnal/tagihan lengkap DlgTagihanOperasi.
Kamar Inap belum menyediakan pindah kamar/pulang.

Pekerjaan lanjutan: proses Kamar Inap, Tagihan Operasi (tarif/obat/jurnal),
template laporan operasi dan cetak Jasper khusus masing-masing dialog.
Cetak catatan saat ini adalah cetak tabel detail browser, bukan dokumen bertanda tangan.
Jangan menyebut workspace ini sudah setara penuh dengan Khanza.
Cetak jadwal menggunakan tabel browser sesuai filter, bukan template Jasper.

Rollback hanya menonaktifkan sidebar dan mempertahankan data klinis.
Belum diuji terhadap database nyata; lakukan verifikasi pada data uji setelah
migration. Verifikasi INSERT nyata dilakukan pengguna di lingkungan uji yang
disetujui; agent tidak menjalankan INSERT ke database nyata.

## Verifikasi sebelum produksi

Gunakan kunjungan/jadwal uji, bukan rekam medis pasien produksi:

1. Buka setiap tab, isi catatan, simpan; pastikan muncul di dialog Khanza.
2. Muat ulang, edit satu catatan, pastikan tidak menambah duplikat.
3. Ubah catatan dari Khanza, coba simpan snapshot lama dari web: harus konflik.
4. Hapus catatan uji dengan konfirmasi; pastikan hanya target yang terhapus.
5. Uji akun nonpemilik, admin, serta billing terkunci.
6. Periksa nilai skor sesuai pilihan dan pastikan tanggal WITA tidak bergeser.
7. Verifikasi kasus kateter tidak ada pada SQL mode server. Jika zero-date ditolak,
   mutasi gagal dan di-rollback; jangan menonaktifkan SQL mode secara global.
