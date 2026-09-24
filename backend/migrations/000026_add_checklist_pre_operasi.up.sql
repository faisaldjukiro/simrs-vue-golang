-- Hanya DB_* lokal. SIMRS lama tetap hanya dibaca.
CREATE TABLE IF NOT EXISTS sirapi_checklist_pre_operasi (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    no_rawat VARCHAR(17) NOT NULL,
    tanggal DATETIME NOT NULL,
    data_checklist JSON NOT NULL,
    dibuat_oleh BIGINT UNSIGNED NOT NULL,
    diubah_oleh BIGINT UNSIGNED NULL,
    versi INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    UNIQUE KEY checklist_kunjungan_waktu (no_rawat, tanggal)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik
    ON pemilik.kode_sidebar = 'checklist_pre_operasi' AND pemilik.id <> target.id
SET target.kode_sidebar = 'checklist_pre_operasi', target.updated_at = NOW()
WHERE target.nama_sidebar = 'Checklist Pre Operasi' AND pemilik.id IS NULL;

UPDATE sidebar_pasien SET aktif = TRUE, updated_at = NOW()
WHERE kode_sidebar = 'checklist_pre_operasi';

INSERT INTO sidebar_pasien
    (kode_sidebar, nama_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
SELECT 'checklist_pre_operasi', 'Checklist Pre Operasi', 'ClipboardCheck',
    JSON_ARRAY('Rawat Inap'), 34, TRUE, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM sidebar_pasien WHERE kode_sidebar = 'checklist_pre_operasi'
);
