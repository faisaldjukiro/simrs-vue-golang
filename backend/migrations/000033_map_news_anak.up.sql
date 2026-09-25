-- DB_* saja. Gunakan sidebar yang sudah ada; pertahankan pengaturan pengguna.
UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik
    ON pemilik.kode_sidebar = 'news_anak'
    AND pemilik.id <> target.id
LEFT JOIN sidebar_pasien kandidat
    ON kandidat.nama_sidebar IN ('Pemantauan NEWS Anak', 'NEWS Anak')
    AND kandidat.id < target.id
SET target.kode_sidebar = 'news_anak', target.updated_at = NOW()
WHERE target.nama_sidebar IN ('Pemantauan NEWS Anak', 'NEWS Anak')
    AND pemilik.id IS NULL
    AND kandidat.id IS NULL;
