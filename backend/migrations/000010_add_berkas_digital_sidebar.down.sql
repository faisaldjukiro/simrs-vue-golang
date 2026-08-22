UPDATE sidebar_pasien
SET kode_sidebar = CONCAT('sidebar_', LPAD(id, 4, '0')),
    daftar_modul = JSON_ARRAY('Rawat Inap'),
    updated_at = NOW()
WHERE kode_sidebar = 'berkas_digital';
