-- Hanya konfigurasi sidebar DB_*. Tidak membuat tabel transaksi.
UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik
 ON pemilik.kode_sidebar='perencanaan_pemulangan' AND pemilik.id<>target.id
LEFT JOIN sidebar_pasien kandidat
 ON kandidat.nama_sidebar='Perencanaan Pemulangan' AND kandidat.id<target.id
SET target.kode_sidebar='perencanaan_pemulangan',target.updated_at=NOW()
WHERE target.nama_sidebar='Perencanaan Pemulangan' AND pemilik.id IS NULL AND kandidat.id IS NULL;

INSERT INTO sidebar_pasien (kode_sidebar,nama_sidebar,ikon,daftar_modul,urutan,aktif,created_at,updated_at)
SELECT 'perencanaan_pemulangan','Perencanaan Pemulangan','House',JSON_ARRAY('Rawat Inap'),38,TRUE,NOW(),NOW()
WHERE NOT EXISTS(SELECT 1 FROM sidebar_pasien WHERE kode_sidebar='perencanaan_pemulangan' OR nama_sidebar='Perencanaan Pemulangan');
