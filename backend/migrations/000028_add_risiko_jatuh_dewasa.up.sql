-- Hanya konfigurasi sidebar pada DB_*; tidak mengubah skema Khanza.
UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik
    ON pemilik.kode_sidebar = 'lanjutan_risiko_jatuh_dewasa'
    AND pemilik.id <> target.id
SET target.kode_sidebar = 'lanjutan_risiko_jatuh_dewasa',
    target.updated_at = NOW()
WHERE target.nama_sidebar = 'Lanjutan Risiko Jatuh Dewasa'
    AND pemilik.id IS NULL;

INSERT INTO sidebar_pasien
    (kode_sidebar, nama_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
SELECT 'lanjutan_risiko_jatuh_dewasa', 'Lanjutan Risiko Jatuh Dewasa',
    'TriangleAlert', JSON_ARRAY('Rawat Inap'), 39, TRUE, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM sidebar_pasien WHERE kode_sidebar = 'lanjutan_risiko_jatuh_dewasa'
);

UPDATE sidebar_pasien SET aktif = TRUE, updated_at = NOW()
WHERE kode_sidebar = 'lanjutan_risiko_jatuh_dewasa';
