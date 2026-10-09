-- Hanya label sidebar pada DB_*; kode dan cakupan modul tetap.
UPDATE sidebar_pasien
SET nama_sidebar = 'Rujuk Internal', updated_at = NOW()
WHERE kode_sidebar = 'rujukan_internal_ranap';
