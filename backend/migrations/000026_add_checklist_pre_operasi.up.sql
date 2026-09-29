-- DB_*: konfigurasi sidebar saja. Transaksi menggunakan tabel SIMRS yang sudah tersedia.
UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik
    ON pemilik.kode_sidebar = 'checklist_pre_operasi' AND pemilik.id <> target.id
SET target.kode_sidebar = 'checklist_pre_operasi', target.updated_at = NOW()
WHERE target.nama_sidebar = 'Checklist Pre Operasi' AND pemilik.id IS NULL;

UPDATE sidebar_pasien SET aktif = TRUE, updated_at = NOW()
WHERE kode_sidebar = 'checklist_pre_operasi';

INSERT INTO sidebar_pasien
    (kode_sidebar, nama_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
SELECT 'checklist_pre_operasi', 'Checklist Pre Operasi', 'ClipboardCheck',
    JSON_ARRAY('Rawat Inap'), 34, TRUE, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM sidebar_pasien WHERE kode_sidebar = 'checklist_pre_operasi'
);
