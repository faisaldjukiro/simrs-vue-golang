# SIRAVA

**Sistem Informasi Rumah Sakit Terintegrasi**

SIRAVA adalah project migrasi SIMRS Khanza secara bertahap ke backend Go dan
frontend Vue. Database aplikasi SIRAVA dipisahkan dari database SIMRS Khanza
agar migration aplikasi baru tidak mengubah struktur database lama.

## Teknologi

- Backend: Go, Gin, dan MySQL
- Frontend: Vue 3, Vite, dan PrimeVue
- Database aplikasi: MySQL lokal
- Integrasi: SIMRS Khanza, E-Klaim, dan BPJS VClaim

## Persyaratan

Pastikan aplikasi berikut sudah tersedia:

- Git
- Go `1.27rc2` atau versi yang sesuai dengan `backend/go.mod`
- Node.js `20.19+` atau `22.12+`
- npm
- MySQL `8+`

Periksa instalasi melalui terminal:

```powershell
git --version
go version
node --version
npm --version
mysql --version
```

## Instalasi Development

### 1. Clone project

```powershell
git clone -b development https://github.com/faisaldjukiro/simrs-vue-golang.git SIMRS-WEB
cd SIMRS-WEB
```

Folder `simrs-lama` tidak ikut di-clone karena hanya digunakan sebagai
referensi lokal dan dikecualikan dari Git.

### 2. Buat database aplikasi lokal

Masuk ke MySQL:

```powershell
mysql -u root -p
```

Buat database baru khusus SIRAVA:

```sql
CREATE DATABASE `simrs-golang`
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;
```

Keluar dari MySQL:

```sql
EXIT;
```

Jangan memakai database `simrs` untuk langkah ini. Database `simrs-golang`
adalah tempat migration, user, permission, token, konfigurasi sidebar, dan log
aktivitas SIRAVA.

### 3. Siapkan backend

Masuk ke folder backend:

```powershell
cd backend
```

Salin konfigurasi contoh. Gunakan salah satu perintah sesuai terminal.

PowerShell:

```powershell
Copy-Item .env.example .env
```

Git Bash:

```bash
cp .env.example .env
```

Pasang dependency Go:

```powershell
go mod download
```

Buka `backend/.env`, lalu periksa bagian database berikut:

```env
# Database aplikasi SIRAVA: boleh ditulis oleh migration dan aplikasi
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=simrs-golang
DB_USERNAME=root
DB_PASSWORD=root

# Pengaman target migration: wajib sama dengan DB_HOST dan DB_DATABASE
MIGRATION_ALLOWED_HOST=127.0.0.1
MIGRATION_ALLOWED_DATABASE=simrs-golang

# Database SIMRS Khanza: hanya dibaca oleh SIRAVA
SIMRS_DB_HOST=127.0.0.1
SIMRS_DB_PORT=3306
SIMRS_DB_DATABASE=simrs
SIMRS_DB_USERNAME=simrs_readonly
SIMRS_DB_PASSWORD=
```

Gunakan alamat, username, dan password MySQL sesuai komputer masing-masing.
Untuk `SIMRS_DB_*`, gunakan user MySQL read-only jika tersedia.

Konfigurasi E-Klaim dan BPJS boleh dikosongkan saat instalasi awal. Modul yang
memerlukan servis tersebut baru dapat digunakan setelah kredensialnya diisi.

### 4. Jalankan migration dan seeder

Pastikan terminal masih berada di folder `backend`, kemudian periksa target:

```powershell
go run ./cmd/migrate status
```

Output harus menunjukkan database `simrs-golang`. Setelah target benar,
jalankan:

```powershell
go run ./cmd/migrate up
go run ./cmd/seed
```

Migration dan seeder hanya membaca konfigurasi `DB_*`. Keduanya ditolak jika:

- `DB_DATABASE` sama dengan `SIMRS_DB_DATABASE`;
- `DB_HOST` berbeda dari `MIGRATION_ALLOWED_HOST`;
- `DB_DATABASE` berbeda dari `MIGRATION_ALLOWED_DATABASE`; atau
- user `root` diarahkan ke host non-lokal.

### 5. Jalankan backend

```powershell
go run ./cmd/api
```

Backend tersedia di:

```text
http://127.0.0.1:8080
```

Coba pemeriksaan sederhana:

```text
GET http://127.0.0.1:8080/health
```

Biarkan terminal backend tetap berjalan.

### 6. Siapkan frontend

Buka terminal baru dari root project, lalu jalankan:

```powershell
cd frontend
```

Salin konfigurasi contoh.

PowerShell:

```powershell
Copy-Item .env.example .env
```

Git Bash:

```bash
cp .env.example .env
```

Konfigurasi development frontend:

```env
VITE_API_URL=
VITE_API_PROXY_TARGET=http://127.0.0.1:8080
```

Pasang dependency dan jalankan Vite:

```powershell
npm install
npm run dev
```

Jika PowerShell menolak script npm, gunakan:

```powershell
npm.cmd install
npm.cmd run dev
```

Buka aplikasi melalui:

```text
http://localhost:5173
```

### 7. Login development

Seeder menyediakan akun lokal awal:

```text
Username: admin
Password: password
```

Akun ini hanya untuk development dan wajib diganti sebelum dipakai pada
environment bersama atau production. User SIMRS yang sudah didaftarkan ke
SIRAVA tetap login memakai password dari tabel `user` SIMRS Khanza.

## Menjalankan Project Setiap Hari

Terminal pertama:

```powershell
cd backend
go run ./cmd/api
```

Terminal kedua:

```powershell
cd frontend
npm run dev
```

Migration dan seeder tidak perlu dijalankan setiap kali aplikasi dibuka.
Jalankan migration kembali hanya setelah menerima file migration baru.

## Perintah Database

Semua perintah berikut wajib dijalankan dari folder `backend`:

```powershell
go run ./cmd/migrate status  # melihat status migration
go run ./cmd/migrate up      # menjalankan migration yang belum diterapkan
go run ./cmd/migrate down    # membatalkan satu kelompok migration terakhir
go run ./cmd/migrate fresh   # membuat ulang seluruh tabel aplikasi lokal
go run ./cmd/seed            # mengisi data awal aplikasi
```

`fresh` menghapus dan membuat ulang tabel pada database aplikasi lokal. Command
akan menampilkan target dan meminta nama database sebagai konfirmasi. Jangan
lanjutkan jika target yang tampil bukan `simrs-golang`.

## Build dan Pemeriksaan

Backend:

```powershell
cd backend
go fmt ./...
go test ./...
go vet ./...
```

Frontend:

```powershell
cd frontend
npm run build
```

Hasil build frontend berada di `frontend/dist`.

## Masalah Umum

### `cmd/migrate: directory not found`

Perintah migration dijalankan dari folder yang salah. Masuk ke folder backend:

```powershell
cd backend
go run ./cmd/migrate up
```

### Migration ditolak

Pastikan nilai berikut cocok:

```env
DB_HOST=127.0.0.1
MIGRATION_ALLOWED_HOST=127.0.0.1
DB_DATABASE=simrs-golang
MIGRATION_ALLOWED_DATABASE=simrs-golang
```

Pastikan pula `SIMRS_DB_DATABASE` tetap berbeda dari `DB_DATABASE`.

### Frontend tidak dapat memanggil API

Periksa bahwa backend masih berjalan pada port `8080`, lalu pastikan:

```env
VITE_API_URL=
VITE_API_PROXY_TARGET=http://127.0.0.1:8080
```

Setelah mengubah `.env` frontend, hentikan dan jalankan ulang `npm run dev`.

### Data SIMRS tidak tampil

Periksa koneksi jaringan ke server SIMRS serta nilai `SIMRS_DB_*`. Koneksi ini
berbeda dari database lokal SIRAVA dan sebaiknya hanya memiliki hak `SELECT`.

## Struktur Project

```text
SIMRS-WEB/
|-- backend/       # REST API Go, migration, seeder, dan integrasi
|-- frontend/      # aplikasi Vue SIRAVA
|-- simrs-lama/    # referensi lokal, tidak masuk Git
|-- AGENTS.md      # panduan kerja agent/developer
|-- .gitignore
`-- README.md
```

Dokumentasi lanjutan:

- [Dokumentasi backend](backend/README.md)
- [Dokumentasi frontend](frontend/README.md)
- [Panduan agent backend](backend/AGENTS.md)

## Aturan Penting

- Jangan mengubah `simrs-lama` kecuali diminta secara khusus.
- Jangan pernah mengarahkan `DB_*` ke database SIMRS Khanza.
- `SIMRS_DB_*` hanya digunakan untuk membaca data SIMRS lama.
- Jangan commit file `.env`, password, token, signature, atau kredensial.
- Endpoint API tidak menggunakan prefix `/v1`.
- Gunakan Bahasa Indonesia untuk module dan response baru jika memungkinkan.
- Migrasikan fitur secara bertahap, bukan menyalin seluruh sistem lama.

Jika `simrs-lama` pernah terlanjur masuk index Git, keluarkan dari index tanpa
menghapus folder lokal:

```powershell
git rm -r --cached -- simrs-lama
git commit -m "chore: keluarkan simrs-lama dari repository"
```
