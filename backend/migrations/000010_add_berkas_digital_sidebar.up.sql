INSERT INTO sidebar_pasien
    (kode_sidebar, nama_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
SELECT
    'berkas_digital',
    'Berkas Digital',
    'FolderOpen',
    JSON_ARRAY('IGD/UGD', 'Rawat Jalan', 'Rawat Inap'),
    18,
    TRUE,
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM sidebar_pasien WHERE kode_sidebar = 'berkas_digital' OR nama_sidebar = 'Berkas Digital'
);

UPDATE sidebar_pasien
SET kode_sidebar = 'berkas_digital',
    nama_sidebar = 'Berkas Digital',
    ikon = 'FolderOpen',
    daftar_modul = JSON_ARRAY('IGD/UGD', 'Rawat Jalan', 'Rawat Inap'),
    aktif = TRUE,
    updated_at = NOW()
WHERE kode_sidebar = 'berkas_digital'
   OR nama_sidebar = 'Berkas Digital';
