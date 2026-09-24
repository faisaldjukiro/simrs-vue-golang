-- Hanya database aplikasi DB_*. Tidak menulis ke SIMRS_DB_*.
CREATE TABLE IF NOT EXISTS sirapi_jadwal_operasi (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    no_rawat VARCHAR(17) NOT NULL,
    kode_paket VARCHAR(30) NOT NULL,
    tanggal DATE NOT NULL,
    jam_mulai TIME NOT NULL,
    jam_selesai TIME NOT NULL,
    status VARCHAR(20) NOT NULL,
    kd_dokter VARCHAR(20) NOT NULL,
    kd_ruang_ok VARCHAR(20) NOT NULL,
    dokteranastesi VARCHAR(255) NOT NULL DEFAULT '',
    perawat VARCHAR(255) NOT NULL DEFAULT '',
    nama_paket VARCHAR(255) NOT NULL,
    nama_dokter VARCHAR(255) NOT NULL,
    nama_ruang VARCHAR(255) NOT NULL,
    dibuat_oleh BIGINT UNSIGNED NOT NULL,
    diubah_oleh BIGINT UNSIGNED NULL,
    versi INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    KEY jadwal_kunjungan (no_rawat, tanggal),
    KEY jadwal_ruang (tanggal, kd_ruang_ok)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS sirapi_jadwal_operasi_lock (
    id INT NOT NULL PRIMARY KEY
) ENGINE=InnoDB;
INSERT IGNORE INTO sirapi_jadwal_operasi_lock (id) VALUES (1);

UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik
    ON pemilik.kode_sidebar = 'jadwal_operasi' AND pemilik.id <> target.id
SET target.kode_sidebar = 'jadwal_operasi', target.updated_at = NOW()
WHERE target.nama_sidebar = 'Jadwal Operasi' AND pemilik.id IS NULL;

INSERT INTO sidebar_pasien
    (kode_sidebar, nama_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
SELECT 'jadwal_operasi', 'Jadwal Operasi', 'CalendarDays',
    JSON_ARRAY('Rawat Inap'), 21, TRUE, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM sidebar_pasien WHERE kode_sidebar = 'jadwal_operasi');

UPDATE sidebar_pasien SET aktif = TRUE, updated_at = NOW()
WHERE kode_sidebar = 'jadwal_operasi';
