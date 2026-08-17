RENAME TABLE menu_workspace_pasien TO sidebar_pasien;

ALTER TABLE sidebar_pasien
    CHANGE COLUMN nama_menu nama_sidebar VARCHAR(80) NOT NULL,
    ADD COLUMN kode_sidebar VARCHAR(80) NULL AFTER id,
    ADD COLUMN permission_code VARCHAR(100) NULL AFTER daftar_modul,
    ADD UNIQUE KEY sidebar_pasien_kode_sidebar_unique (kode_sidebar),
    ADD KEY sidebar_pasien_permission_code_index (permission_code);

DELETE sidebar_duplikat
FROM sidebar_pasien sidebar_duplikat
INNER JOIN sidebar_pasien sidebar_utama
    ON sidebar_utama.nama_sidebar = sidebar_duplikat.nama_sidebar
   AND sidebar_utama.id < sidebar_duplikat.id;

UPDATE sidebar_pasien
SET kode_sidebar = CONCAT('sidebar_', LPAD(id, 4, '0'));

UPDATE sidebar_pasien SET kode_sidebar = 'cppt_soap', permission_code = 'pasien.cppt'
WHERE nama_sidebar = 'Cppt/Soap';
UPDATE sidebar_pasien SET kode_sidebar = 'penanganan_dokter_petugas', permission_code = 'pasien.penanganan_dokter_petugas'
WHERE nama_sidebar = 'Penanganan Dokter & Petugas';
UPDATE sidebar_pasien SET kode_sidebar = 'diagnosa', permission_code = 'pasien.diagnosa'
WHERE nama_sidebar = 'Diagnosa';
UPDATE sidebar_pasien SET kode_sidebar = 'riwayat_perawatan', permission_code = 'pasien.riwayat_perawatan'
WHERE nama_sidebar = 'Riwayat Perawatan';
UPDATE sidebar_pasien SET kode_sidebar = 'permintaan_laboratorium', permission_code = 'permintaan_lab'
WHERE nama_sidebar IN ('Permintaan Lab', 'Permintaan Laboratorium')
ORDER BY id LIMIT 1;
UPDATE sidebar_pasien SET kode_sidebar = 'permintaan_radiologi', permission_code = 'permintaan_radiologi'
WHERE nama_sidebar IN ('Permintaan Rad', 'Permintaan Radiologi')
ORDER BY id LIMIT 1;
UPDATE sidebar_pasien SET kode_sidebar = 'triase_igd', permission_code = 'pasien.triase_igd'
WHERE nama_sidebar = 'Triase IGD';
UPDATE sidebar_pasien SET kode_sidebar = 'awal_keperawatan_igd', permission_code = 'pasien.awal_keperawatan_igd'
WHERE nama_sidebar = 'Awal Keperawatan IGD';
UPDATE sidebar_pasien SET kode_sidebar = 'resume_pasien', permission_code = 'pasien.resume_pasien'
WHERE nama_sidebar IN ('Resume Pasien', 'Resume Pasien Ranap')
ORDER BY id LIMIT 1;

INSERT INTO permissions (`group`, name, code, created_at, updated_at) VALUES
    ('Sidebar Pasien', 'CPPT / SOAP', 'pasien.cppt', NOW(), NOW()),
    ('Sidebar Pasien', 'Penanganan Dokter & Petugas', 'pasien.penanganan_dokter_petugas', NOW(), NOW()),
    ('Sidebar Pasien', 'Diagnosa Pasien', 'pasien.diagnosa', NOW(), NOW()),
    ('Sidebar Pasien', 'Riwayat Perawatan', 'pasien.riwayat_perawatan', NOW(), NOW()),
    ('Sidebar Pasien', 'Triase IGD', 'pasien.triase_igd', NOW(), NOW()),
    ('Sidebar Pasien', 'Awal Keperawatan IGD', 'pasien.awal_keperawatan_igd', NOW(), NOW()),
    ('Sidebar Pasien', 'Resume Pasien', 'pasien.resume_pasien', NOW(), NOW())
ON DUPLICATE KEY UPDATE
    `group` = VALUES(`group`),
    name = VALUES(name),
    updated_at = NOW();

ALTER TABLE sidebar_pasien
    ADD CONSTRAINT sidebar_pasien_permission_code_foreign
        FOREIGN KEY (permission_code) REFERENCES permissions (code)
        ON UPDATE CASCADE ON DELETE SET NULL;
