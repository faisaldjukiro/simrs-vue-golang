-- DB_*: konfigurasi sidebar saja. Transaksi menggunakan tabel SIMRS yang sudah tersedia.
UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik
    ON pemilik.kode_sidebar = 'jadwal_operasi' AND pemilik.id <> target.id
SET target.kode_sidebar = 'jadwal_operasi', target.updated_at = NOW()
WHERE target.nama_sidebar = 'Jadwal Operasi' AND pemilik.id IS NULL;

INSERT INTO sidebar_pasien
    (kode_sidebar, nama_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
SELECT 'jadwal_operasi', 'Jadwal Operasi', 'CalendarDays',
    JSON_ARRAY('Rawat Inap'), 21, TRUE, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM sidebar_pasien WHERE kode_sidebar = 'jadwal_operasi');

UPDATE sidebar_pasien SET aktif = TRUE, updated_at = NOW()
WHERE kode_sidebar = 'jadwal_operasi';
