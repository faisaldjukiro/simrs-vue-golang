-- Pertahankan data klinis ketika rollback. Nonaktifkan menu saja.
UPDATE sidebar_pasien SET aktif=FALSE,updated_at=NOW() WHERE kode_sidebar='data_hais';
