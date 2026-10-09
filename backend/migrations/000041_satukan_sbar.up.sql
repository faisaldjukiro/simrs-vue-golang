-- Hanya konfigurasi menu pada DB_*; tidak mengubah tabel transaksi SIMRS.
UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik ON pemilik.kode_sidebar='sbar' AND pemilik.id<>target.id
LEFT JOIN sidebar_pasien kandidat ON kandidat.nama_sidebar IN ('SBAR','SBAR & Verifikasi') AND kandidat.id<target.id
SET target.kode_sidebar='sbar',target.updated_at=NOW()
WHERE target.nama_sidebar IN ('SBAR','SBAR & Verifikasi') AND pemilik.id IS NULL AND kandidat.id IS NULL;

INSERT INTO sidebar_pasien (kode_sidebar,nama_sidebar,ikon,daftar_modul,urutan,aktif,created_at,updated_at)
SELECT 'sbar','SBAR & Verifikasi','MessageSquareText',JSON_ARRAY('Rawat Inap'),3,TRUE,NOW(),NOW()
WHERE NOT EXISTS (SELECT 1 FROM sidebar_pasien WHERE kode_sidebar='sbar');

UPDATE sidebar_pasien
SET nama_sidebar='SBAR & Verifikasi',ikon='MessageSquareText',daftar_modul=JSON_ARRAY('Rawat Inap'),updated_at=NOW()
WHERE kode_sidebar='sbar';

-- Pertahankan entri lama untuk konfigurasi/arsip; cukup nonaktifkan menu duplikat.
UPDATE sidebar_pasien
SET aktif=FALSE,updated_at=NOW()
WHERE kode_sidebar<>'sbar'
  AND (nama_sidebar IN ('SBAR','SBAR & Verifikasi','Verifikasi SBAR','Validasi SBAR')
       OR kode_sidebar IN ('verifikasi_sbar','validasi_sbar'));
