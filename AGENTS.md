# SIRAPI Agent Guide

Project ini berisi migrasi bertahap dari SIMRS lama ke SIRAPI (Sistem Informasi Rumah Sakit Pelayanan Terintegrasi).

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

Untuk pekerjaan frontend, UI, layout, warna, form, tabel, dark mode, dan
komponen tampilan, wajib baca:

```text
DESIGN.md
```

Jangan membuat style frontend baru yang bertentangan dengan `DESIGN.md`.
Gunakan component bersama seperti `FormInput.vue`, `InputPencarian.vue`, dan
`Select.vue` agar tampilan antar modul tetap konsisten.

Aturan penting backend:

- Migration hanya boleh menuju database lokal `DB_*`.
- Database SIMRS lama `SIMRS_DB_*` hanya boleh dibaca.
- Jangan memakai prefix `/v1`.
- Gunakan Bahasa Indonesia untuk penamaan module dan response API jika memungkinkan.
- Kerjakan bertahap dan jelaskan perubahan dengan Bahasa Indonesia.
