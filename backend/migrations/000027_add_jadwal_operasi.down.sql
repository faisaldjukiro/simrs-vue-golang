-- Pertahankan jadwal dan audit saat rollback.
UPDATE sidebar_pasien SET aktif = FALSE, updated_at = NOW()
WHERE kode_sidebar = 'jadwal_operasi';
