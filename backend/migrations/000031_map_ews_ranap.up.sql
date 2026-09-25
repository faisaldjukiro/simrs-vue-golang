-- DB_* saja. Hubungkan menu yang sudah ada tanpa mengubah konfigurasi pengguna.
UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik
    ON pemilik.kode_sidebar = 'ews_ranap'
    AND pemilik.id <> target.id
SET target.kode_sidebar = 'ews_ranap', target.updated_at = NOW()
WHERE target.nama_sidebar = 'EWS Ranap'
    AND pemilik.id IS NULL;
