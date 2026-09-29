-- DB_*: konfigurasi sidebar saja. Transaksi menggunakan tabel SIMRS yang sudah tersedia.
-- Gunakan kembali placeholder Rujuk Internal Rawat Inap yang sudah ada.
UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik ON pemilik.kode_sidebar = 'rujukan_internal_ranap' AND pemilik.id <> target.id
SET target.kode_sidebar = 'rujukan_internal_ranap',
    target.nama_sidebar = 'Rujuk Internal Rawat Inap',
    target.daftar_modul = JSON_ARRAY('Rawat Inap'),
    target.ikon = 'ArrowRightToLine', target.aktif = TRUE, target.updated_at = NOW()
WHERE target.nama_sidebar = 'Rujuk Internal' AND pemilik.id IS NULL;

INSERT INTO sidebar_pasien
    (kode_sidebar, nama_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
SELECT 'rujukan_internal_ranap', 'Rujuk Internal Rawat Inap', 'ArrowRightToLine', JSON_ARRAY('Rawat Inap'), 10, TRUE, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM sidebar_pasien WHERE kode_sidebar = 'rujukan_internal_ranap');

INSERT INTO sidebar_pasien
    (kode_sidebar, nama_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
SELECT 'rujukan_internal_poli', 'Rujuk Internal Poli', 'ArrowRightToLine', JSON_ARRAY('Rawat Jalan', 'IGD/UGD'), 10, TRUE, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM sidebar_pasien WHERE kode_sidebar = 'rujukan_internal_poli');
