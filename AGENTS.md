# SIRAVA Agent Guide

Project ini berisi migrasi bertahap dari SIMRS lama ke SIRAVA (Sistem Informasi Rumah Sakit Terintegrasi).

Folder utama:

```text
backend     backend baru Go, Gin, MySQL
frontend    frontend baru Vue
simrs-lama  project lama sebagai referensi
```

Jangan mengubah `simrs-lama` kecuali user meminta secara eksplisit. Folder itu dipakai sebagai referensi tampilan, flow, dan struktur data.

Untuk pekerjaan backend, baca instruksi detail di:

```text
backend/AGENTS.md
```

Aturan penting backend:

- Migration hanya boleh menuju database lokal `DB_*`.
- Database SIMRS lama `SIMRS_DB_*` hanya boleh dibaca.
- Jangan memakai prefix `/v1`.
- Gunakan Bahasa Indonesia untuk penamaan module dan response API jika memungkinkan.
- Kerjakan bertahap dan jelaskan perubahan dengan Bahasa Indonesia.
