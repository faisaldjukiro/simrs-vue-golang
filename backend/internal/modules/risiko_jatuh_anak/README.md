# Lanjutan Risiko Jatuh Anak

Referensi: `simrs-lama/src/rekammedis/RMPenilaianLanjutanRisikoJatuhAnak.java`
dan skema `penilaian_lanjutan_resiko_jatuh_anak` di dump Khanza.

## Aktivasi

Jalankan `go run ./cmd/migrate up` dari backend, restart backend, lalu refresh
frontend. Migration 000029 hanya mengatur sidebar lokal `DB_*` dengan kode
`lanjutan_risiko_jatuh_anak` untuk Rawat Inap. Tidak perlu seed ulang.
Tidak ada migrasi skema atau penulisan data klinis saat aktivasi.

## Proses

Sesuai pola modul dewasa yang diminta pengguna, simpan/edit/hapus langsung ke
Khanza. Endpoint `/api/risiko-jatuh-anak`:

- GET dengan query `no_rawat`: pilihan, nilai, dan riwayat kunjungan.
- POST JSON `{no_rawat, data}`: simpan.
- PUT JSON `{no_rawat, data, asli}`: edit.
- DELETE JSON `{no_rawat, asli}`: hapus dengan konfirmasi pada UI.

`data` berisi kolom Khanza selain `no_rawat`; `asli` adalah snapshot dari GET.
Backend menghitung ulang tujuh nilai dan total Humpty Dumpty (7–23).
Pilihan, bobot, 20 kolom, serta primary key `no_rawat, tanggal` mengikuti Java
dan skema lama. Tanggal input WITA. Hasil skrining dan saran wajib diisi petugas,
masing-masing maksimum 200 karakter. Pilihan tidak diisi otomatis.

Catatan perbedaan yang disengaja: `isTotalResikoJatuh()` Java tidak memperbarui
label saat total kembali >=12 (cabang kedua `<7` juga tidak terjangkau).
SIRAPI selalu menghitung kategori ulang menurut batas yang tertulis di form
Java: 7–11 rendah, >=12 tinggi. Nilai historis di luar 7–23 diberi label
“Skor tidak valid”, tidak diam-diam diganti. Tidak membuat rekomendasi klinis
otomatis atau mengisi jawaban berdasarkan identitas pasien.

Akses API mengikuti modul pelayanan pasien. Edit/hapus mensyaratkan username
sama dengan NIP pencatat atau permission administrator `*`, dengan pemeriksaan
snapshot dalam transaksi untuk mencegah penimpaan perubahan bersamaan.
Audit mengikuti middleware aplikasi. Referensi petugas memakai endpoint bersama
Checklist Pre Operasi. Tampilan memakai kelas klinis CPPT dalam light/dark mode.

## Pengujian

Tes sementara mencocokkan pilihan dengan sumber Java, seluruh 3.456 kombinasi
skor, batas 7/11/12/23, penolakan data tidak valid, serta kepemilikan edit/hapus.
Tidak menjalankan simpan/edit/hapus terhadap database Khanza nyata saat pengembangan.
Verifikasi menggunakan kunjungan uji sebelum digunakan dalam pelayanan produksi.
