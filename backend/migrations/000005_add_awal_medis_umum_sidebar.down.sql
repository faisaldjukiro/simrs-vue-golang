UPDATE sidebar_pasien
SET daftar_modul = JSON_ARRAY('Rawat Inap'),
    permission_code = NULL,
    updated_at = NOW()
WHERE kode_sidebar = 'sidebar_0031'
  AND nama_sidebar = 'Awal Medis Umum';

DELETE FROM permissions WHERE code = 'pasien.awal_medis_umum';
