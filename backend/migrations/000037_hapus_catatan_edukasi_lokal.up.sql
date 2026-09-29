-- HANYA database konfigurasi DB_*, melalui cmd/migrate yang memvalidasi target.
-- Backup terlebih dahulu: isi tabel lokal terhapus, bukan dipindahkan ke SIMRS.
-- Jangan menjalankan SQL ini pada SIMRS_DB_*: tabel catatan_edukasi di sana
-- adalah tabel transaksi utama dan harus tetap dipertahankan.
DROP TABLE IF EXISTS catatan_edukasi;
