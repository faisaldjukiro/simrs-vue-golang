# SIRAPI Design Guide

Dokumen ini adalah pegangan tampilan dan interaksi untuk frontend SIRAPI.
Tujuannya supaya setiap modul baru terasa satu keluarga, tidak punya warna,
spacing, tabel, dan form yang berbeda-beda.

## Prinsip Utama

- Ikuti pola tampilan yang sudah disepakati di modul yang matang, terutama:
  - CPPT/SOAP
  - Penanganan Dokter & Petugas
  - Data pasien rawat inap/rawat jalan/IGD
  - E-Klaim, jika modulnya memang bridging/klaim
- Jangan membuat style form sendiri jika sudah ada component bersama.
- Jangan membuat warna hard-code berlebihan di halaman modul.
- Semua tampilan harus mendukung `theme-light` dan `theme-dark`.
- Kalau ragu, pilih tampilan yang mudah dibaca dan tidak terlalu menyala.

## Component Wajib untuk Form

Gunakan component berikut untuk input agar warna, ukuran, label, border, focus,
disabled, dan dark mode konsisten:

```text
frontend/src/Components/Ui/FormInput.vue
frontend/src/Components/Ui/InputPencarian.vue
frontend/src/Components/Ui/Select.vue
```

Aturan:

- Input teks, tanggal, jam, angka, textarea, dan select gunakan `FormInput.vue`.
- Pencarian data referensi seperti dokter, petugas, tindakan, obat, diagnosa,
  prosedur, dan pegawai gunakan component pencarian bersama.
- Jangan override class inti berikut kecuali benar-benar perlu:
  - `.form-input-field`
  - `.form-input-label`
  - `.form-input-element`
  - `.staff-search`
  - `.staff-search-box`
  - `.staff-search-selected`
  - `.staff-search-results`
- Kalau butuh z-index dropdown, override hanya z-index pada wrapper modul.

Contoh yang boleh:

```css
.nama-modul .staff-search-results {
  z-index: 1000;
}
```

Contoh yang dihindari:

```css
.nama-modul .form-input-element {
  background: #111827 !important;
  color: #ffffff !important;
}
```

## Token Warna

Untuk halaman modul, pakai CSS variable global:

```css
var(--surface)
var(--surface-soft)
var(--line)
var(--text)
var(--muted)
```

Untuk aksen utama SIRAPI gunakan hijau/teal secukupnya:

```css
#0d9488
```

Jangan membuat banyak warna baru dalam satu modul. Teal cukup untuk tombol
utama, badge aktif, icon penting, atau border fokus.

## Dark Mode dan Light Mode

Setiap modul wajib aman di dua mode:

- Light mode:
  - Background form jangan terlalu putih menyilaukan jika form panjang.
  - Border input harus terlihat.
  - Text jangan terlalu tipis.
- Dark mode:
  - Jangan pakai background putih pada dropdown/list.
  - Text jangan terlalu glow atau terlalu bold.
  - Header tabel jangan terlalu kontras dari body tabel.

Gunakan selector global yang sudah ada:

```css
.theme-light
.theme-dark
.sirapi-dark
```

Tapi jangan membuat override mode terlalu spesifik jika variable global sudah
cukup.

## Layout Form Modul Pasien

Untuk modul di dalam ruang kerja pasien:

- Header identitas pasien berada di atas.
- Card modul diberi jarak dari header pasien.
- Header card berisi judul modul dan tombol collapse/action utama.
- Area form input diberi padding cukup, jangan menempel pada header.
- Gunakan form yang bisa di-hide/show jika form panjang.
- Saat klik edit, form otomatis tampil.

Rekomendasi spacing:

```css
.module-card {
  margin-top: 12px;
}

.module-form {
  padding: 24px 16px 16px;
}
```

## Struktur Halaman Frontend

Halaman di dalam `frontend/src/Pages` hanya menangani susunan tampilan dan
binding ke controller fitur. Proses pengambilan data, state, validasi, simpan,
edit, dan hapus ditempatkan pada composable `useNamaFitur.js`.

Style yang hanya digunakan satu fitur ditempatkan pada folder fitur. Style
bersama seperti `clinical-form-card`, `clinical-button`, form input, dan tabel
tetap berada pada stylesheet/component global.

Contoh struktur:

```text
Pages/RawatInap/
|-- VentilatorPage.vue
`-- Ventilator/
    |-- useVentilator.js
    `-- ventilator.css
```

Aturan:

- Jangan mengembalikan proses API atau validasi panjang ke dalam file Page.
- Jangan memindahkan style global ke CSS fitur.
- Pecah template menjadi komponen anak hanya jika bagiannya berdiri sendiri,
  digunakan ulang, atau membuat Page sulit dibaca.
- Pertahankan file `*Page.vue` sebagai entry point agar route dan pemetaan
  `kode_sidebar` tetap stabil.

## Tabel

Tabel SIRAPI harus:

- Header jelas tetapi tidak menyakitkan mata.
- Body row zebra/alternating halus jika data panjang.
- Kolom panjang boleh horizontal scroll.
- Text utama cukup tebal, tapi jangan semua dibuat terlalu bold.
- Data kecil seperti kode, tanggal, status, stok, harga boleh memakai ukuran
  lebih kecil.

Untuk tabel yang meniru Khanza, boleh lebih padat, tetapi tetap harus:

- Support dark mode.
- Support horizontal scroll.
- Field editable terlihat jelas.
- Kolom tidak saling menabrak.

## Input Resep

Input resep mengikuti alur `DlgPeresepanDokter.java` dari SIMRS Khanza:

- Satu nomor resep bisa berisi banyak obat.
- User mencari obat, list muncul otomatis.
- Saat obat dicentang, obat langsung masuk ke tabel pilihan.
- User tinggal mengisi `Jumlah` dan `Aturan Pakai` pada tabel.
- Simpan resep dilakukan sekali setelah semua obat/racikan siap.
- Tabel obat terpilih mengikuti gaya Khanza, tetapi tetap memakai warna SIRAPI.

Kolom obat yang diutamakan:

- K
- Jumlah
- Kode Barang
- Nama Barang
- Satuan
- Harga(Rp)
- Jenis Obat
- Aturan Pakai
- H.Beli
- Stok
- Fornas
- Aksi hapus

Kolom `Komposisi` dan `I.F.` tidak ditampilkan di tabel pilihan resep kecuali
user meminta lagi.

## Toast dan Alert

Gunakan PrimeVue toast melalui helper:

```text
frontend/src/lib/shared/useNotifikasi.js
```

Jangan gunakan alert browser untuk flow normal kecuali konfirmasi sederhana
sementara. Untuk error backend, tampilkan pesan yang jelas dan pakai bahasa
Indonesia.

## Sidebar Pasien

Sidebar pasien menggunakan konfigurasi dari database lokal `sidebar_pasien`.
Frontend menentukan halaman berdasarkan `kode_sidebar`, bukan `nama_sidebar`.

Aturan:

- Jangan hard-code fallback sidebar jika API gagal.
- Jika user tidak punya akses, tampilkan informasi tidak memiliki akses.
- Jika jaringan/API bermasalah, tampilkan pesan koneksi bermasalah.
- Jangan membuat permission per sidebar. Jika user memiliki akses ke modul
  besar pasien seperti IGD/UGD, Rawat Jalan, atau Rawat Inap, sidebar yang
  aktif untuk modul tersebut boleh ditampilkan.

## Referensi SIMRS Khanza

Folder `simrs-lama` adalah referensi flow, field, dan perilaku.

Aturan:

- Boleh dibaca untuk meniru proses.
- Jangan diedit kecuali user meminta eksplisit.
- Jika tampilan Khanza terlalu desktop/rapat, adaptasi ke SIRAPI tetapi jangan
  mengubah alur kerja utama.

## Checklist Sebelum Menyerahkan UI

- [ ] Form memakai `FormInput.vue` / component UI bersama.
- [ ] Pencarian memakai component pencarian bersama atau pola yang konsisten.
- [ ] Light mode terbaca.
- [ ] Dark mode terbaca.
- [ ] Dropdown/list tidak transparan dan z-index aman.
- [ ] Tabel panjang bisa horizontal scroll.
- [ ] Tidak ada warna hard-code berlebihan.
- [ ] Tidak ada `!important` kecuali terpaksa untuk override library.
- [ ] Build frontend berhasil:

```powershell
cd frontend
npm.cmd run build
```
