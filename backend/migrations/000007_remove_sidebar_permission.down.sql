ALTER TABLE sidebar_pasien
    ADD COLUMN permission_code VARCHAR(100) NULL AFTER daftar_modul,
    ADD KEY sidebar_pasien_permission_code_index (permission_code);

INSERT INTO permissions (`group`, name, code, created_at, updated_at) VALUES
    ('Sidebar Pasien', 'CPPT / SOAP', 'pasien.cppt', NOW(), NOW()),
    ('Sidebar Pasien', 'Penanganan Dokter & Petugas', 'pasien.penanganan_dokter_petugas', NOW(), NOW()),
    ('Sidebar Pasien', 'Diagnosa Pasien', 'pasien.diagnosa', NOW(), NOW()),
    ('Sidebar Pasien', 'Riwayat Perawatan', 'pasien.riwayat_perawatan', NOW(), NOW()),
    ('Sidebar Pasien', 'Triase IGD', 'pasien.triase_igd', NOW(), NOW()),
    ('Sidebar Pasien', 'Awal Keperawatan IGD', 'pasien.awal_keperawatan_igd', NOW(), NOW()),
    ('Sidebar Pasien', 'Awal Medis Umum', 'pasien.awal_medis_umum', NOW(), NOW()),
    ('Sidebar Pasien', 'Input Resep', 'pasien.resep', NOW(), NOW()),
    ('Sidebar Pasien', 'Resume Pasien', 'pasien.resume_pasien', NOW(), NOW())
ON DUPLICATE KEY UPDATE
    `group` = VALUES(`group`),
    name = VALUES(name),
    updated_at = NOW();

UPDATE sidebar_pasien SET permission_code = 'pasien.cppt'
WHERE nama_sidebar = 'Cppt/Soap';
UPDATE sidebar_pasien SET permission_code = 'pasien.penanganan_dokter_petugas'
WHERE nama_sidebar = 'Penanganan Dokter & Petugas';
UPDATE sidebar_pasien SET permission_code = 'pasien.diagnosa'
WHERE nama_sidebar = 'Diagnosa';
UPDATE sidebar_pasien SET permission_code = 'pasien.riwayat_perawatan'
WHERE nama_sidebar = 'Riwayat Perawatan';
UPDATE sidebar_pasien SET permission_code = 'permintaan_lab'
WHERE nama_sidebar IN ('Permintaan Lab', 'Permintaan Laboratorium');
UPDATE sidebar_pasien SET permission_code = 'permintaan_radiologi'
WHERE nama_sidebar IN ('Permintaan Rad', 'Permintaan Radiologi');
UPDATE sidebar_pasien SET permission_code = 'pasien.triase_igd'
WHERE nama_sidebar = 'Triase IGD';
UPDATE sidebar_pasien SET permission_code = 'pasien.awal_keperawatan_igd'
WHERE nama_sidebar = 'Awal Keperawatan IGD';
UPDATE sidebar_pasien SET permission_code = 'pasien.awal_medis_umum'
WHERE nama_sidebar = 'Awal Medis Umum';
UPDATE sidebar_pasien SET permission_code = 'pasien.resume_pasien'
WHERE nama_sidebar IN ('Resume Pasien', 'Resume Pasien Ranap');
UPDATE sidebar_pasien SET permission_code = 'pasien.resep'
WHERE nama_sidebar = 'Input Resep';

ALTER TABLE sidebar_pasien
    ADD CONSTRAINT sidebar_pasien_permission_code_foreign
        FOREIGN KEY (permission_code) REFERENCES permissions (code)
        ON UPDATE CASCADE ON DELETE SET NULL;
