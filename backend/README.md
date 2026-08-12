# SIRAVA Backend

SIRAVA adalah Sistem Informasi Rumah Sakit Terintegrasi.

Backend baru SIRAVA dibuat dengan Go, Gin, dan MySQL. Project ini dipakai sebagai backend baru untuk migrasi bertahap dari `simrs-lama`.

Yang paling penting: database aplikasi lokal dan database SIMRS lama dipisah. Migration hanya boleh masuk ke database lokal `DB_*`, sedangkan database SIMRS lama `SIMRS_DB_*` hanya dibaca untuk kebutuhan data.

## Struktur Project

```text
backend/
|-- cmd/
|   |-- api/            # entry point server HTTP Gin
|   |-- migrate/        # command migration database aplikasi lokal
|   `-- seed/           # command seeder data awal aplikasi
|-- internal/
|   |-- config/         # baca env dan validasi konfigurasi
|   |-- modules/        # fitur aplikasi
|   |   |-- autentikasi/        # login, token, middleware auth
|   |   |-- beranda/            # query data untuk halaman beranda
|   |   `-- manajemen_pengguna/ # daftar user dan permission aplikasi
|   |-- platform/
|   |   `-- database/   # koneksi MySQL, migrator, seeder
|   |-- routes/         # pembagian route per developer
|   |   |-- faisal/
|   |   `-- sahrul/
|   `-- shared/         # helper umum seperti format response HTTP
|-- migrations/         # file SQL migration dan seeders
|-- .env.example        # contoh konfigurasi lokal
|-- go.mod
`-- README.md
```

## Cara Membaca Alur Kode

Alur request API:

```text
cmd/api/main.go
-> internal/routes/<developer>/routes.go
-> internal/modules/<nama-module>/delivery/http/handler.go
-> internal/modules/<nama-module>/repository.go atau service.go
```

Contoh endpoint beranda:

```text
GET /api/beranda
-> internal/routes/faisal/routes.go
-> internal/modules/beranda/delivery/http/handler.go
-> internal/modules/beranda/repository.go
```

Query data beranda berada di:

```text
internal/modules/beranda/repository.go
```

Di file itu saat ini ada fungsi:

```text
bacaRingkasan
bacaRegistrasiHariIni
bacaRawatInapAktif
```

## Pembagian Module

`internal/modules/autentikasi`

Module untuk login, token, logout, dan middleware Bearer Token. Endpoint publiknya tetap memakai `/api/auth/...`.

`internal/modules/beranda`

Module untuk data yang tampil di halaman beranda/dashboard. Module ini membaca database SIMRS lama memakai koneksi `SIMRS_DB_*` dengan transaksi read-only.

Jika nanti membuat fitur baru, buat module baru di:

```text
internal/modules/nama_module/
```

Pola yang disarankan:

```text
internal/modules/nama_module/
|-- repository.go
`-- delivery/
    `-- http/
        `-- handler.go
```

Gunakan penamaan Bahasa Indonesia agar mudah dibaca tim.

## Pembagian Route Developer

Route dipisahkan agar Faisal dan Sahrul tidak sering mengubah file yang sama.

```text
internal/routes/
|-- faisal/routes.go
`-- sahrul/routes.go
```

Aturan route:

- Route yang sudah dibuat saat ini berada di `internal/routes/faisal/routes.go`.
- Sahrul menambahkan route baru di `internal/routes/sahrul/routes.go`.
- Nama developer hanya untuk organisasi kode, tidak menjadi URL publik.
- Handler dan business logic tetap berada di `internal/modules`.
- Jangan mendaftarkan method dan path yang sama pada dua file route, karena Gin akan error saat startup.

Endpoint yang sudah tersedia:

```text
GET  /health
POST /api/auth/login
GET  /api/auth/me
POST /api/auth/logout
GET  /api/beranda
GET  /api/user-management
GET  /api/user-management/pegawai?q=fa
POST /api/user-management
PUT  /api/user-management/:id/akses
```

## Database

Ada dua koneksi database:

```text
DB_*        -> database aplikasi lokal backend baru
SIMRS_DB_*  -> database SIMRS lama, hanya untuk dibaca
```

Contoh lokal:

```env
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=simrs-golang
DB_USERNAME=root
DB_PASSWORD=root
MIGRATION_ALLOWED_HOST=127.0.0.1
MIGRATION_ALLOWED_DATABASE=simrs-golang

SIMRS_DB_HOST=192.168.20.1
SIMRS_DB_PORT=3306
SIMRS_DB_DATABASE=simrs
SIMRS_DB_USERNAME=root
SIMRS_DB_PASSWORD=
```

Penting:

- Migration hanya membaca `DB_*`.
- Migration tidak pernah memakai `SIMRS_DB_*`.
- `DB_DATABASE` tidak boleh sama dengan `SIMRS_DB_DATABASE`.
- Untuk database SIMRS lama, idealnya gunakan user MySQL khusus read-only.
- Jangan menjalankan SQL migration manual ke database SIMRS lama.

## Menyiapkan Environment

Copy contoh env:

```powershell
copy .env.example .env
```

Lalu sesuaikan nilai database lokal dan database SIMRS lama di `.env`.

Pastikan target migration tetap database lokal:

```env
DB_DATABASE=simrs-golang
MIGRATION_ALLOWED_DATABASE=simrs-golang
```

## Migration

File migration berada di:

```text
migrations/
|-- 000001_create_application_schema.up.sql
|-- 000001_create_application_schema.down.sql
|-- 000002_create_access_tokens.up.sql
|-- 000002_create_access_tokens.down.sql
`-- seeders/
    `-- 000001_seed_application_data.sql
```

Command migration:

```powershell
go run ./cmd/migrate up
go run ./cmd/migrate status
go run ./cmd/migrate down
go run ./cmd/migrate fresh
```

`fresh` hanya untuk database aplikasi lokal. Command ini meminta konfirmasi nama database sebelum menghapus tabel aplikasi.

Setelah migration, jalankan seeder:

```powershell
go run ./cmd/seed
```

Seeder mengisi data awal seperti permission, user admin development, menu navigasi, dan menu workspace pasien.

## Menjalankan API

Dari folder `backend`:

```powershell
go run ./cmd/migrate up
go run ./cmd/seed
go run ./cmd/api
```

Default API:

```text
http://localhost:8080
```

## Mencoba Login di Postman

Login:

```text
POST http://localhost:8080/api/auth/login
```

Body JSON:

```json
{
  "username": "198901010001",
  "password": "password-simrs"
}
```

Login akan mencoba akun SIMRS lama lebih dulu memakai tabel `user` pada koneksi `SIMRS_DB_*`. Jika koneksi/akun SIMRS tidak cocok, backend memakai fallback akun lokal, misalnya akun development `admin` / `password`.

Ambil `access_token` dari response, lalu pakai sebagai Bearer Token.

Melihat user aktif:

```text
GET http://localhost:8080/api/auth/me
```

Logout:

```text
POST http://localhost:8080/api/auth/logout
```

Beranda:

```text
GET http://localhost:8080/api/beranda
```

User Management:

```text
GET http://localhost:8080/api/user-management
GET http://localhost:8080/api/user-management/pegawai?q=faisal
POST http://localhost:8080/api/user-management
PUT http://localhost:8080/api/user-management/2/akses
```

Tambah user mengikuti pola `simrs-lama`: pilih pegawai dari tabel `pegawai` pada database SIMRS lama, lalu berikan permission di database lokal SIRAVA. Password tidak dibuat di SIRAVA; saat login user tetap memakai password dari tabel `user` SIMRS lama.

Body tambah user:

```json
{
  "username": "198901010001",
  "nama": "Sahrul",
  "aktif": true,
  "all_access": false,
  "permission": ["registrasi", "tindakan_ralan"]
}
```

Daftar `permission` diambil dari tabel `permissions`. Jika ingin memberi semua akses seperti admin di `simrs-lama`, gunakan `all_access: true`. Backend akan menyimpan permission `*`.

```json
{
  "all_access": true,
  "permission": ["*"]
}
```

Body ubah akses user:

```json
{
  "aktif": true,
  "all_access": false,
  "permission": ["registrasi", "daftar_pasien_ranap"]
}
```

Filter pasien mengikuti pola `simrs-lama`:

```text
GET /api/beranda?rj_date_from=2026-08-09&rj_date_to=2026-08-09&rj_status=Belum&rj_poly=INT&rj_search=andi&rj_page=1&rj_limit=100
GET /api/beranda?igd_date_from=2026-08-09&igd_date_to=2026-08-09&igd_status=Sudah&igd_search=andi&igd_page=1&igd_limit=100
GET /api/beranda?ri_date_from=2026-08-09&ri_date_to=2026-08-09&ri_status=-&ri_search=andi&ri_belum_pulang=1&ri_page=1&ri_limit=100
```

Pagination pasien dilakukan dari backend/service. Default `limit` adalah 100 dan maksimal 500 data per halaman.

Response beranda memakai nama Bahasa Indonesia:

```json
{
  "koneksi_database": {},
  "ringkasan": {},
  "registrasi": [],
  "rawat_jalan": [],
  "igd": [],
  "rawat_inap": [],
  "paginasi": {
    "rawat_jalan": {
      "halaman": 1,
      "batas": 100,
      "total": 0,
      "total_halaman": 0
    }
  },
  "poliklinik": [],
  "pilihan_status": {
    "periksa": [],
    "rawat_inap": []
  }
}
```

## Development

Format kode:

```powershell
gofmt -w ./cmd ./internal
```

Test:

```powershell
go test ./...
go vet ./...
```

Jika Go cache di Windows bermasalah `Access is denied`, arahkan cache ke folder project:

```powershell
$env:GOCACHE='C:\project\SIMRS-WEB\backend\tmp\go-cache'
go test ./...
```

## Catatan Penting untuk Migrasi dari SIMRS Lama ke SIRAVA

- Jangan mengubah folder `simrs-lama` kecuali memang diminta.
- Gunakan `simrs-lama` sebagai referensi tampilan, flow, dan nama tabel.
- Backend baru tidak masuk ke mode/module `simrs` dulu.
- Migration backend baru harus tetap masuk ke database lokal `simrs-golang`.
- Query ke database SIMRS lama harus bersifat baca data saja.
