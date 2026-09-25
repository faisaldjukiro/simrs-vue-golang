-- Kembalikan kode generik sidebar. Tidak menghapus menu atau catatan klinis.
UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik
    ON pemilik.kode_sidebar = CONCAT('sidebar_', LPAD(target.id, 4, '0'))
    AND pemilik.id <> target.id
SET target.kode_sidebar = CONCAT('sidebar_', LPAD(target.id, 4, '0')),
    target.updated_at = NOW()
WHERE target.kode_sidebar = 'observasi_ranap'
    AND pemilik.id IS NULL;
