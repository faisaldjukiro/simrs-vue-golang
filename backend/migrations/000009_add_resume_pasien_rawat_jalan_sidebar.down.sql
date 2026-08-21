UPDATE sidebar_pasien
SET daftar_modul = JSON_ARRAY('Rawat Inap'),
    updated_at = NOW()
WHERE kode_sidebar = 'resume_pasien';
