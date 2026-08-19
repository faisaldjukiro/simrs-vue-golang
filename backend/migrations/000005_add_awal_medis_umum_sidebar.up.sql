INSERT INTO permissions (`group`, name, code, created_at, updated_at)
VALUES ('Sidebar Pasien', 'Awal Medis Umum', 'pasien.awal_medis_umum', NOW(), NOW())
ON DUPLICATE KEY UPDATE
    `group` = VALUES(`group`),
    name = VALUES(name),
    updated_at = NOW();

UPDATE sidebar_pasien
SET kode_sidebar = CONCAT('sidebar_lama_', LPAD(id, 4, '0'))
WHERE kode_sidebar = 'sidebar_0031'
  AND nama_sidebar <> 'Awal Medis Umum';

INSERT INTO sidebar_pasien
    (kode_sidebar, nama_sidebar, ikon, daftar_modul, permission_code, urutan, aktif, created_at, updated_at)
SELECT
    'sidebar_0031', 'Awal Medis Umum', 'Stethoscope', JSON_ARRAY('Rawat Jalan'),
    'pasien.awal_medis_umum', 31, TRUE, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM sidebar_pasien WHERE nama_sidebar = 'Awal Medis Umum'
);

UPDATE sidebar_pasien
SET kode_sidebar = 'sidebar_0031',
    nama_sidebar = 'Awal Medis Umum',
    ikon = 'Stethoscope',
    daftar_modul = JSON_ARRAY('Rawat Jalan'),
    permission_code = 'pasien.awal_medis_umum',
    aktif = TRUE,
    updated_at = NOW()
WHERE nama_sidebar = 'Awal Medis Umum';
