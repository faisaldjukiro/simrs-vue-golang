DELETE FROM sidebar_pasien WHERE kode_sidebar = 'rujukan_internal_poli';
UPDATE sidebar_pasien
SET kode_sidebar = CONCAT('sidebar_', LPAD(id, 4, '0')),
    nama_sidebar = 'Rujuk Internal', daftar_modul = JSON_ARRAY('Rawat Inap'), updated_at = NOW()
WHERE kode_sidebar = 'rujukan_internal_ranap';
-- Rollback menghapus rujukan lokal; backup sebelum menjalankan down.
DROP TABLE IF EXISTS sirapi_rujukan_internal;
