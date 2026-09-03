INSERT INTO permissions (`group`,name,code,created_at,updated_at)
VALUES ('Laporan','Laporan 10 Penyakit','laporan_10_penyakit',NOW(),NOW())
ON DUPLICATE KEY UPDATE name=VALUES(name),updated_at=NOW();

INSERT INTO menu_navigasi
    (tipe,label,deskripsi,ikon,tone,route,permissions,urutan,aktif,created_at,updated_at)
SELECT
    'dashboard','Laporan 10 Penyakit','Rekap 10 penyakit terbanyak berdasarkan periode pelayanan.',
    'ChartBarBig','rose',NULL,JSON_ARRAY('laporan_10_penyakit'),15,TRUE,NOW(),NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM menu_navigasi WHERE tipe='dashboard' AND label='Laporan 10 Penyakit'
);
