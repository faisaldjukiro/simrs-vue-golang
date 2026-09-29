-- HANYA database konfigurasi DB_*, dijalankan melalui cmd/migrate.
-- Permintaan penghapusan tabel transaksi duplikat. Backup sebelum menjalankan:
-- seluruh isi empat tabel berikut akan hilang, bukan dipindahkan ke SIMRS.
-- Tidak mengubah tabel akun, permission, sidebar, audit, atau SIMRS_DB_*.
DROP TABLE IF EXISTS sirapi_jadwal_operasi_lock;
DROP TABLE IF EXISTS sirapi_jadwal_operasi;
DROP TABLE IF EXISTS sirapi_checklist_pre_operasi;
DROP TABLE IF EXISTS sirapi_rujukan_internal;
