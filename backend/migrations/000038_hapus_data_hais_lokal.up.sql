-- HANYA database konfigurasi DB_*, melalui cmd/migrate yang memvalidasi target.
-- Backup terlebih dahulu: seluruh isi tabel lokal terhapus, bukan dipindahkan.
-- Jangan jalankan SQL ini pada SIMRS_DB_*; tabel data_HAIs di SIMRS
-- adalah tabel transaksi utama dan harus tetap dipertahankan.
DROP TABLE IF EXISTS data_hais;
