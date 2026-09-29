-- DB_*: konfigurasi sidebar saja. Transaksi menggunakan tabel SIMRS yang sudah tersedia.
UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik
 ON pemilik.kode_sidebar='edukasi_pasien' AND pemilik.id<>target.id
LEFT JOIN sidebar_pasien kandidat
 ON kandidat.nama_sidebar='Edukasi Pasien' AND kandidat.id<target.id
SET target.kode_sidebar='edukasi_pasien',target.updated_at=NOW()
WHERE target.nama_sidebar='Edukasi Pasien' AND pemilik.id IS NULL AND kandidat.id IS NULL;

INSERT INTO sidebar_pasien (kode_sidebar,nama_sidebar,ikon,daftar_modul,urutan,aktif,created_at,updated_at)
SELECT 'edukasi_pasien','Edukasi Pasien','BookOpen',JSON_ARRAY('Rawat Inap'),30,TRUE,NOW(),NOW()
WHERE NOT EXISTS(SELECT 1 FROM sidebar_pasien WHERE kode_sidebar='edukasi_pasien' OR nama_sidebar='Edukasi Pasien');
