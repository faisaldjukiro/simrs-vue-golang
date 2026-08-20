UPDATE sidebar_pasien SET kode_sidebar = CONCAT('sidebar_', LPAD(id, 4, '0')), permission_code = NULL
WHERE nama_sidebar = 'Input Resep';
DELETE FROM permissions WHERE code = 'pasien.resep';
