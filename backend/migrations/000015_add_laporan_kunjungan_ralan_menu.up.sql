INSERT INTO permissions (`group`,name,code,created_at,updated_at) VALUES ('Laporan','Laporan Kunjungan Rawat Jalan','laporan_kunjungan_ralan',NOW(),NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name),updated_at=NOW();
INSERT INTO menu_navigasi (tipe,label,deskripsi,ikon,tone,route,permissions,urutan,aktif,created_at,updated_at)
SELECT 'dashboard','Laporan Kunjungan Ralan','Laporan kunjungan rawat jalan berdasarkan periode dan filter pelayanan.','FileBarChart','emerald',NULL,JSON_ARRAY('laporan_kunjungan_ralan'),13,TRUE,NOW(),NOW()
WHERE NOT EXISTS(SELECT 1 FROM menu_navigasi WHERE tipe='dashboard' AND label='Laporan Kunjungan Ralan');
