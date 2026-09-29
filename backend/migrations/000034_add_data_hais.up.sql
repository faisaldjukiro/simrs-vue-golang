-- DB_*: konfigurasi sidebar saja. Transaksi menggunakan tabel SIMRS yang sudah tersedia.
INSERT INTO sidebar_pasien (kode_sidebar,nama_sidebar,ikon,daftar_modul,urutan,aktif,created_at,updated_at)
SELECT 'data_hais','Data HAIs','ClipboardList',JSON_ARRAY('Rawat Inap'),65,TRUE,NOW(),NOW()
WHERE NOT EXISTS (SELECT 1 FROM sidebar_pasien WHERE kode_sidebar='data_hais');
