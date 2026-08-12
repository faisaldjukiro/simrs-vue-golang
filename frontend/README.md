# SIRAVA Frontend

SIRAVA adalah Sistem Informasi Rumah Sakit Terintegrasi.

Frontend Vue 3 untuk backend Go SIRAVA. Tampilan login dan dashboard mengikuti identitas visual `simrs-lama` dengan banner rumah sakit, status bar, ribbon menu, monitoring koneksi, logo RSAS, serta tema terang/gelap.

Dashboard menampilkan ringkasan dan daftar pasien dari endpoint Go untuk registrasi hari ini, rawat jalan, IGD, serta rawat inap aktif. Daftar dapat dicari berdasarkan nama, nomor rekam medis, nomor rawat, poliklinik, dokter, atau kamar.

## Struktur Vue

```text
src/
|-- App.vue                     # state sesi, modal login, toast login/logout
|-- Pages/
|   `-- Home.vue                # halaman utama/dashboard
|-- Components/
|   |-- LoginPage.vue           # form login yang ditampilkan sebagai modal
|   |-- Common/
|   |   `-- PatientFilters.vue  # filter pasien rawat jalan/IGD/rawat inap
|   |-- Layout/
|   |   |-- DashboardLayout.vue # penyusun layout dashboard
|   |   |-- TopStatusBar.vue    # status tanggal, jam, user, tema
|   |   |-- RibbonMenu.vue      # ribbon/menu utama SIRAVA
|   |   |-- AppFooter.vue       # footer aplikasi
|   |   `-- ShutdownScreen.vue  # layar sesi ditutup jika dibutuhkan
|   |-- Ui/
|   |   |-- DataTable.vue       # wrapper PrimeVue DataTable untuk tabel reusable
|   |   |-- DatePicker.vue      # wrapper PrimeVue DatePicker
|   |   |-- Select.vue          # wrapper PrimeVue Select/dropdown
|   |   `-- Toast.vue           # host PrimeVue Toast
|   `-- Tabs/
|       |-- BerandaTab.vue      # isi tab Beranda/Menu
|       |-- ModuleTab.vue       # tabel data pasien untuk tab registrasi/jalan/IGD/inap
|       `-- ModulePlaceholder.vue
`-- lib/
    |-- shared/
    |   |-- http.js             # helper request API bersama
    |   `-- useNotifikasi.js    # helper toast PrimeVue
    |-- faisal/
    |   `-- api.js              # function/query API milik Faisal
    `-- sahrul/
        `-- api.js              # function/query API milik Sahrul
```

Struktur Vue mengikuti pola `simrs-lama`: `Pages/` untuk halaman besar, `Components/Tabs/` untuk isi tab dashboard, `Components/` untuk komponen yang dipakai ulang, lalu `composables/` dan `Data/` bisa ditambahkan nanti jika sudah dibutuhkan.

Untuk menghindari tabrakan kerja, function API/query frontend dipisah per programmer:

- Faisal menaruh function API di `src/lib/faisal/api.js`.
- Sahrul menaruh function API di `src/lib/sahrul/api.js`.
- Helper koneksi API bersama berada di `src/lib/shared/http.js`.

Page CRUD seperti rawat jalan, rawat inap, IGD, farmasi, dan lainnya belum dibuat. Nanti jika modul CRUD mulai dikerjakan, buat file page baru di `src/Pages/` sesuai modulnya.

Komponen UI umum dibungkus di `src/Components/Ui/` supaya PrimeVue tidak dipanggil mentah di banyak page. Jika butuh tabel, dropdown, tanggal, atau toast di modul lain, gunakan wrapper di folder tersebut.

Alert dan notifikasi menggunakan `PrimeVue Toast` melalui helper `useNotifikasi()`.

Filter tanggal pasien menggunakan `DatePicker`, dropdown menggunakan `Select`, dan tabel pasien menggunakan `DataTable` supaya tampilan tetap seragam dan mendukung dark mode. Tabel pasien memakai backend pagination/lazy loading: default 100 data per halaman dan maksimal 500 data per request.

## Menjalankan development

Jalankan backend terlebih dahulu dari folder `backend`:

```powershell
go run ./cmd/migrate up
go run ./cmd/seed
go run ./cmd/api
```

Pada terminal lain, jalankan frontend:

```powershell
cd frontend
npm install
npm run dev
```

Buka `http://localhost:5173`.

Akun development dari seeder:

```text
Username: admin
Password: password
```

Password tersebut wajib diganti sebelum aplikasi digunakan pada environment bersama atau production.

## Koneksi API

Saat development, request `/api` diteruskan oleh proxy Vite ke backend Go.

```env
VITE_API_URL=
VITE_API_PROXY_TARGET=http://127.0.0.1:8080
```

Artinya browser tetap memanggil `/api/...`, lalu Vite meneruskan ke `VITE_API_PROXY_TARGET`.

Untuk deployment terpisah tanpa proxy Vite, isi `VITE_API_URL` langsung ke alamat backend:

```env
VITE_API_URL=http://127.0.0.1:8080
```

## Build production

```powershell
npm run build
```

Hasil build berada di folder `dist`.
