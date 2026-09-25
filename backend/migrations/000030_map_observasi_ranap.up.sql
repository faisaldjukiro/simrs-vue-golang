-- DB_* saja. Hubungkan sidebar yang sudah ada; jangan membuat menu duplikat.
-- Pertahankan nama, urutan, status aktif, dan pengaturan modul milik pengguna.
UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik
    ON pemilik.kode_sidebar = 'observasi_ranap'
    AND pemilik.id <> target.id
SET target.kode_sidebar = 'observasi_ranap', target.updated_at = NOW()
WHERE target.nama_sidebar = 'Observasi Ranap'
    AND pemilik.id IS NULL;
