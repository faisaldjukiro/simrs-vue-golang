INSERT INTO sidebar_pasien (nama_sidebar, kode_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
SELECT 'Copy Resep', 'copy_resep', 'Copy', JSON_ARRAY('IGD/UGD', 'Rawat Jalan', 'Rawat Inap'), 7, TRUE, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM sidebar_pasien
    WHERE nama_sidebar = 'Copy Resep' OR kode_sidebar = 'copy_resep'
);

UPDATE sidebar_pasien
SET kode_sidebar = 'copy_resep',
    daftar_modul = JSON_ARRAY('IGD/UGD', 'Rawat Jalan', 'Rawat Inap')
WHERE nama_sidebar = 'Copy Resep';
