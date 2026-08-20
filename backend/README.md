# SIRAPI Backend

SIRAPI adalah Sistem Informasi Rumah Sakit Pelayanan Terintegrasi.

Backend baru SIRAPI dibuat dengan Go, Gin, dan MySQL. Project ini dipakai sebagai backend baru untuk migrasi bertahap dari `simrs-lama`.

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
GET  /api/aktivitas-log
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
|-- 000003_expand_aktivitas_log.up.sql
|-- 000003_expand_aktivitas_log.down.sql
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

Seeder mengisi data awal seperti permission, user admin development, menu navigasi, dan sidebar pasien.

### Sidebar Pasien

Konfigurasi sidebar ruang kerja pasien disimpan pada tabel lokal
`sidebar_pasien`. Kolom pentingnya:

```text
kode_sidebar     kode stabil yang dipakai frontend untuk membuka halaman
nama_sidebar     nama yang ditampilkan kepada user
ikon             nama ikon sidebar
daftar_modul     daftar modul tempat sidebar berlaku dalam format JSON
urutan           urutan tampilan sidebar
aktif            status tampil atau tidak
```

Frontend memakai `kode_sidebar`, bukan `nama_sidebar`, sehingga nama tampilan
bisa diubah tanpa merusak pemetaan halaman. Endpoint `GET /api/beranda`
mengembalikan field `sidebar_pasien` aktif. Sidebar tidak punya permission
sendiri-sendiri; aksesnya mengikuti menu besar pelayanan pasien seperti
IGD/UGD, Rawat Jalan, dan Rawat Inap.

User dengan permission `kelola_menu` atau `*` dapat mengatur sidebar melalui:

```text
GET    /api/kelola-menu
POST   /api/kelola-menu/sidebar
PUT    /api/kelola-menu/sidebar/:id
DELETE /api/kelola-menu/sidebar/:id
```

Perubahan nama, ikon, modul, urutan, dan status aktif langsung dibaca kembali
oleh Patient Sidebar. `kode_sidebar` hanya boleh diubah jika
pemetaan halaman frontend ikut menggunakan kode baru tersebut.

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

Log Aktivitas:

```text
GET http://localhost:8080/api/aktivitas-log?halaman=1&batas=25&tanggal_mulai=2026-08-01&tanggal_selesai=2026-08-16
```

Endpoint ini hanya dapat dibuka oleh user dengan permission `*` atau
`sistem.audit_log`. Audit trail disimpan ke database aplikasi lokal `DB_*`,
bukan ke database SIMRS Khanza. Password, token, signature, dan secret selalu
disamarkan. Log mencatat login/logout, GET, tambah, ubah, hapus, proses,
status HTTP, durasi, alamat IP, serta snapshot request perubahan.

## Mencoba Signature BPJS VClaim

Struktur awal module BPJS:

```text
internal/modules/bpjs/
|-- signature.go                    # generator signature bersama
|-- response.go                     # membuka response AES + LZString bersama
|-- delivery/http/handler.go        # endpoint uji signature
`-- vclaim/
    `-- monitoring/
        `-- data_klaim/
            |-- service.go          # fungsi Monitoring Data Klaim
            `-- delivery/http/
                `-- handler.go
```

Backend membaca konfigurasi BPJS dari file `.env` saat dijalankan. Pastikan
nilai berikut sudah terisi di `.env` (bukan hanya di `.env.example`):

```env
BPJS_CONS_ID=
BPJS_SECRET=
BPJS_USER_KEY=
BPJS_VCLAIM_URL=
```

Jalankan API, login, lalu gunakan Bearer Token dari response login untuk
mencoba endpoint berikut di Postman:

```text
GET http://localhost:8080/api/bpjs/signature
Authorization: Bearer <access_token>
```

Endpoint ini hanya membuat signature lokal dan **belum memanggil servis
BPJS**. Rumus yang digunakan sama dengan Khanza:

```text
timestamp = Unix time dalam detik
data      = BPJS_CONS_ID + "&" + timestamp
signature = Base64(HMAC-SHA256(data, BPJS_SECRET))
```

Response hanya menampilkan `X-timestamp` dan `X-signature`. Nilai
`BPJS_SECRET` dan `BPJS_USER_KEY` tidak dikirim ke frontend/Postman.

### Mencoba Monitoring Data Klaim

Setelah signature berhasil, koneksi VClaim dapat diuji melalui endpoint:

```text
GET http://localhost:8080/api/bpjs/monitoring/klaim?tanggal_pulang=2026-08-14&jenis_pelayanan=1&status_klaim=1
Authorization: Bearer <access_token>
```

Untuk mengambil data dalam rentang tanggal (maksimal 31 hari):

```text
GET http://localhost:8080/api/bpjs/monitoring/klaim?tanggal_mulai=2026-08-01&tanggal_selesai=2026-08-14&jenis_pelayanan=1&status_klaim=1
Authorization: Bearer <access_token>
```

Parameter:

- `tanggal_pulang`: format `yyyy-mm-dd`.
- `tanggal_mulai` dan `tanggal_selesai`: format `yyyy-mm-dd`, wajib diisi
  bersamaan untuk mengambil rentang maksimal 31 hari. Jika parameter rentang
  dipakai, `tanggal_pulang` tidak perlu dikirim.
- `jenis_pelayanan`: `1` rawat inap, `2` rawat jalan, atau `semua` untuk
  menarik keduanya dalam satu request.
- `status_klaim`: `1` proses verifikasi, `2` pending verifikasi, atau `3` klaim.

Backend membuat header BPJS `X-cons-id`, `X-timestamp`, `X-signature`,
`user_key`, `Content-Type: application/json`, `Accept: application/json`, dan
`User-Agent: SIMRS-BRIDGING-BPJS/1.0`, lalu memanggil endpoint Monitoring Klaim.
Jika response VClaim terenkripsi, backend otomatis melakukan dekripsi
AES-256-CBC dan dekompresi LZString sebelum mengirim JSON ke Postman.
Karena endpoint resmi BPJS hanya menerima satu tanggal pulang, backend akan
memanggil VClaim per tanggal lalu menggabungkan seluruh item `klaim` menjadi
satu response. Field `periode` pada response menunjukkan tanggal mulai,
tanggal selesai, dan jumlah hari yang diproses. Jika BPJS mengembalikan kode
`404` dengan pesan bahwa data tidak ditemukan, tanggal tersebut dilewati dan
dicantumkan pada field `tanggal_tanpa_data`; data dari tanggal lain tetap
dikembalikan. Respons `404` dengan pesan `Silakan coba lagi nanti` dianggap
sebagai gangguan VClaim. Pada request satu tanggal, backend mencoba ulang
otomatis maksimal tiga kali. Pada request rentang, tanggal tersebut dicatat
di field `tanggal_gagal` dan proses dilanjutkan ke tanggal berikutnya agar
data tanggal lain tetap dapat ditarik.

Respons BPJS `201` dengan pesan `Data Tidak Ada` juga diperlakukan sebagai
tanggal tanpa data, bukan sebagai kegagalan aplikasi.

Tambah user mengikuti pola `simrs-lama`: pilih pegawai dari tabel `pegawai` pada database SIMRS lama, lalu berikan permission di database lokal SIRAPI. Password tidak dibuat di SIRAPI; saat login user tetap memakai password dari tabel `user` SIMRS lama.

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

## Catatan Penting untuk Migrasi dari SIMRS Lama ke SIRAPI

- Jangan mengubah folder `simrs-lama` kecuali memang diminta.
- Gunakan `simrs-lama` sebagai referensi tampilan, flow, dan nama tabel.
- Backend baru tidak masuk ke mode/module `simrs` dulu.
- Migration backend baru harus tetap masuk ke database lokal `simrs-golang`.
- Query ke database SIMRS lama harus bersifat baca data saja.
