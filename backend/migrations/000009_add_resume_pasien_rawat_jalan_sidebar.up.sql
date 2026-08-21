UPDATE sidebar_pasien
SET daftar_modul = JSON_ARRAY('Rawat Jalan', 'Rawat Inap'),
    nama_sidebar = 'Resume Pasien',
    ikon = 'NotebookText',
    aktif = TRUE,
    updated_at = NOW()
WHERE kode_sidebar = 'resume_pasien';

INSERT INTO sidebar_pasien (nama_sidebar, kode_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
SELECT 'Resume Pasien', 'resume_pasien', 'NotebookText', JSON_ARRAY('Rawat Jalan', 'Rawat Inap'), 25, TRUE, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM sidebar_pasien WHERE kode_sidebar = 'resume_pasien'
);
