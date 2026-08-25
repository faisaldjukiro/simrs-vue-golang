UPDATE sidebar_pasien
SET kode_sidebar = CONCAT('sidebar_', LPAD(id, 4, '0')), updated_at = NOW()
WHERE kode_sidebar = 'ews_ranap';
