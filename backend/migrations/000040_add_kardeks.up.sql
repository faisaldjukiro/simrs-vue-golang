-- Konfigurasi DB_* saja; tidak membuat tabel transaksi.
UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik ON pemilik.kode_sidebar='kardeks' AND pemilik.id<>target.id
LEFT JOIN sidebar_pasien kandidat ON kandidat.nama_sidebar IN ('Kardeks','Kardex','Kardeks ICU') AND kandidat.id<target.id
SET target.kode_sidebar='kardeks',target.updated_at=NOW()
WHERE target.nama_sidebar IN ('Kardeks','Kardex','Kardeks ICU') AND pemilik.id IS NULL AND kandidat.id IS NULL;
INSERT INTO sidebar_pasien (kode_sidebar,nama_sidebar,ikon,daftar_modul,urutan,aktif,created_at,updated_at)
SELECT 'kardeks','Kardeks ICU','Activity',JSON_ARRAY('Rawat Inap'),39,TRUE,NOW(),NOW()
WHERE NOT EXISTS(SELECT 1 FROM sidebar_pasien WHERE kode_sidebar='kardeks' OR nama_sidebar IN ('Kardeks','Kardex','Kardeks ICU'));
