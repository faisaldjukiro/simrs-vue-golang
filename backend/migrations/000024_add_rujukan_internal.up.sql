-- Hanya database lokal DB_*. Tidak membuat atau mengubah tabel SIMRS_DB_*.
CREATE TABLE sirapi_rujukan_internal (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    jenis_rawat ENUM('Ralan', 'Ranap') NOT NULL,
    no_rawat VARCHAR(17) NOT NULL,
    kd_dokter VARCHAR(20) NOT NULL,
    nama_dokter VARCHAR(255) NOT NULL,
    kd_poli VARCHAR(5) NOT NULL,
    nama_poli VARCHAR(255) NOT NULL,
    kunci_tujuan VARCHAR(20) NOT NULL,
    tanggal DATE NULL,
    jam TIME NULL,
    dibuat_oleh VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY rujukan_kunjungan_tujuan_unique (jenis_rawat, no_rawat, kunci_tujuan)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Gunakan kembali placeholder Rujuk Internal Rawat Inap yang sudah ada.
UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik ON pemilik.kode_sidebar = 'rujukan_internal_ranap' AND pemilik.id <> target.id
SET target.kode_sidebar = 'rujukan_internal_ranap',
    target.nama_sidebar = 'Rujuk Internal Rawat Inap',
    target.daftar_modul = JSON_ARRAY('Rawat Inap'),
    target.ikon = 'ArrowRightToLine', target.aktif = TRUE, target.updated_at = NOW()
WHERE target.nama_sidebar = 'Rujuk Internal' AND pemilik.id IS NULL;

INSERT INTO sidebar_pasien
    (kode_sidebar, nama_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
SELECT 'rujukan_internal_ranap', 'Rujuk Internal Rawat Inap', 'ArrowRightToLine', JSON_ARRAY('Rawat Inap'), 10, TRUE, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM sidebar_pasien WHERE kode_sidebar = 'rujukan_internal_ranap');

INSERT INTO sidebar_pasien
    (kode_sidebar, nama_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
SELECT 'rujukan_internal_poli', 'Rujuk Internal Poli', 'ArrowRightToLine', JSON_ARRAY('Rawat Jalan', 'IGD/UGD'), 10, TRUE, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM sidebar_pasien WHERE kode_sidebar = 'rujukan_internal_poli');
