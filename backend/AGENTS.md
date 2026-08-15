# Backend Agent Guide

File ini adalah instruksi cepat untuk AI coding agent yang bekerja di folder `backend`.

## Konteks Project

Ini backend baru SIRAVA (Sistem Informasi Rumah Sakit Terintegrasi) berbasis Go, Gin, dan MySQL. Project sedang dipakai untuk migrasi bertahap dari `simrs-lama`.

Jangan ubah `simrs-lama` kecuali user meminta secara eksplisit. Pakai folder itu hanya sebagai referensi.

## Aturan Database Paling Penting

Ada dua koneksi database:

- `DB_*` adalah database aplikasi lokal backend baru.
- `SIMRS_DB_*` adalah database SIMRS lama.

Migration dan seeder hanya boleh memakai `DB_*`.

Jangan pernah menjalankan migration, fresh, drop, alter, insert, update, atau delete ke database `SIMRS_DB_*`.

Database SIMRS lama hanya boleh dibaca untuk query tampilan dan migrasi bertahap yang disetujui user.

Default yang diinginkan user:

```env
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=simrs-golang
DB_USERNAME=root
DB_PASSWORD=root
MIGRATION_ALLOWED_HOST=127.0.0.1
MIGRATION_ALLOWED_DATABASE=simrs-golang
```

## Struktur Utama

```text
cmd/api              entry point server HTTP
cmd/migrate          command migration
cmd/seed             command seeder
internal/config      env dan validasi konfigurasi
internal/modules     fitur aplikasi
internal/routes      registrasi route per developer
internal/platform    adapter teknis seperti database
internal/shared      helper umum
migrations           file SQL migration dan seeders
```

## Module Saat Ini

```text
internal/modules/autentikasi
```

Untuk login, logout, token, dan middleware auth.

```text
internal/modules/beranda
```

Untuk query data yang tampil di halaman beranda. Query SQL utama berada di:

```text
internal/modules/beranda/repository.go
```

Gunakan penamaan Bahasa Indonesia untuk module, struct, fungsi, dan variabel baru jika masih nyaman secara Go.

## Route Developer

Route dibagi agar dua developer tidak sering konflik:

```text
internal/routes/faisal/routes.go
internal/routes/sahrul/routes.go
```

Route yang sudah ada sekarang dikelola Faisal:

```text
GET  /health
POST /api/auth/login
GET  /api/auth/me
POST /api/auth/logout
GET  /api/beranda
```

Jika menambah fitur milik Sahrul, tambahkan route di `internal/routes/sahrul/routes.go`. Jika menambah dependency handler baru, tambahkan ke struct `Dependencies` file route milik developer terkait.

Jangan memakai prefix `/v1`; user sudah meminta `/v1` dihapus.

## Pola Penambahan Fitur

Untuk fitur baru, buat module:

```text
internal/modules/nama_fitur/
|-- repository.go
`-- delivery/
    `-- http/
        `-- handler.go
```

Lalu:

1. Inisialisasi repository/handler di `cmd/api/main.go`.
2. Tambahkan dependency ke `internal/routes/<developer>/routes.go`.
3. Daftarkan endpoint di file route developer tersebut.
4. Gunakan response helper dari `internal/shared/httpresponse`.

## Cara Menjalankan

Dari folder `backend`:

```powershell
go run ./cmd/migrate up
go run ./cmd/seed
go run ./cmd/api
```

Test:

```powershell
go test ./...
go vet ./...
```

Jika Go cache di Windows error `Access is denied`, pakai:

```powershell
$env:GOCACHE='C:\project\SIMRS-WEB\backend\tmp\go-cache'
go test ./...
go vet ./...
```

## Gaya Kerja

- Kerjakan pelan-pelan dan bertahap.
- Jelaskan perubahan dengan Bahasa Indonesia.
- Jangan refactor besar tanpa diminta.
- Jangan mengubah endpoint publik tanpa memberi tahu user.
- Pertahankan response API memakai Bahasa Indonesia.
- Setelah mengubah Go code, jalankan `gofmt`, `go test ./...`, dan `go vet ./...` jika memungkinkan.

## Aturan File Test Sementara

- File Go dengan akhiran `_test.go` hanya boleh dibuat sementara untuk memverifikasi implementasi.
- Jalankan pengujian sampai berhasil sebelum menyerahkan hasil pekerjaan.
- Setelah pengujian berhasil, hapus kembali file `_test.go` yang dibuat untuk pekerjaan tersebut agar struktur project tetap ringkas.
- Setelah file test sementara dihapus, jalankan pemeriksaan kompilasi package terkait untuk memastikan kode utama tetap dapat dibangun.
- Untuk integrasi servis eksternal seperti BPJS, berikan URL, parameter, dan cara pengujian Postman kepada user. Pengujian respons nyata dilakukan user melalui Postman.
