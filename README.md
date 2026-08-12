# SIRAVA

**Sistem Informasi Rumah Sakit Terintegrasi**

SIRAVA adalah project migrasi SIMRS secara bertahap ke backend Go dan frontend Vue. Sistem lama tetap digunakan sebagai sumber referensi alur kerja, tampilan, dan data selama proses migrasi.

## Teknologi

- Backend: Go, Gin, dan MySQL
- Frontend: Vue 3, Vite, dan PrimeVue
- Database aplikasi baru: MySQL lokal
- Integrasi: database SIMRS lama dan Web Service E-Klaim

## Struktur Project

```text
SIMRS-WEB/
|-- backend/       # REST API Go, migration, seeder, dan integrasi SIMRS
|-- frontend/      # aplikasi Vue SIRAVA
|-- simrs-lama/    # referensi sistem lama, tidak boleh masuk Git
|-- AGENTS.md      # panduan kerja untuk AI/developer
|-- .gitignore
`-- README.md
```

Dokumentasi lebih rinci tersedia di:

- [Dokumentasi backend](backend/README.md)
- [Dokumentasi frontend](frontend/README.md)
- [Panduan agent backend](backend/AGENTS.md)

## Persyaratan Development

Siapkan aplikasi berikut:

- Go sesuai versi pada `backend/go.mod`
- Node.js dan npm
- MySQL

## Menyiapkan Backend

Dari root project:

```powershell
cd backend
copy .env.example .env
```

Sesuaikan koneksi database pada `backend/.env`, kemudian jalankan:

```powershell
go run ./cmd/migrate up
go run ./cmd/seed
go run ./cmd/api
```

Backend berjalan pada `http://127.0.0.1:8080` jika memakai konfigurasi bawaan.

### Keamanan Database

Project menggunakan dua koneksi yang berbeda:

```text
DB_*        database lokal milik aplikasi SIRAVA
SIMRS_DB_*  database SIMRS lama untuk dibaca
```

Aturan wajib:

- Migration dan seeder hanya boleh dijalankan pada database `DB_*`.
- Jangan mengarahkan `DB_*` ke database SIMRS lama.
- Koneksi `SIMRS_DB_*` dipakai untuk membaca data SIMRS lama.
- Gunakan akun MySQL read-only untuk `SIMRS_DB_*` jika memungkinkan.
- Jangan commit file `.env` atau kredensial database.

## Menyiapkan Frontend

Buka terminal baru dari root project:

```powershell
cd frontend
copy .env.example .env
npm install
npm run dev
```

Frontend development dapat dibuka melalui `http://localhost:5173`.

Konfigurasi bawaan meneruskan request `/api` ke backend melalui Vite proxy:

```env
VITE_API_URL=
VITE_API_PROXY_TARGET=http://127.0.0.1:8080
```

## Build dan Pemeriksaan

Backend:

```powershell
cd backend
go test ./...
go vet ./...
```

Frontend:

```powershell
cd frontend
npm run build
```

## Aturan Pengembangan

- Jangan mengubah `simrs-lama/` kecuali diminta secara khusus.
- Jangan menyalin seluruh project lama ke project baru; migrasikan modul secara bertahap.
- Gunakan Bahasa Indonesia untuk nama module dan response API jika memungkinkan.
- Endpoint API tidak menggunakan prefix `/v1`.
- Query dan business logic berada di module masing-masing, bukan di komponen tampilan.
- Route developer dipisahkan dalam `backend/internal/routes/faisal` dan `backend/internal/routes/sahrul`.
- Helper API frontend developer dipisahkan dalam `frontend/src/lib/faisal` dan `frontend/src/lib/sahrul`.

## File yang Tidak Masuk Git

`.gitignore` root melindungi:

- seluruh folder `simrs-lama/`;
- file `.env` dan kredensial lokal;
- `node_modules`, hasil build, cache, log, dan file editor;
- file sementara backend dan frontend.

Jika `simrs-lama/` pernah terlanjur masuk commit sebelumnya, `.gitignore` tidak otomatis menghapusnya dari index Git. Hapus hanya dari index tanpa menghapus folder lokal dengan:

```powershell
git rm -r --cached -- simrs-lama
git commit -m "chore: keluarkan simrs-lama dari repository"
```

