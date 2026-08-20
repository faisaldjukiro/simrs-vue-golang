INSERT INTO permissions (`group`, name, code, created_at, updated_at)
VALUES ('Sidebar Pasien', 'Input Resep', 'pasien.resep', NOW(), NOW())
ON DUPLICATE KEY UPDATE name = VALUES(name), updated_at = NOW();

UPDATE sidebar_pasien
SET kode_sidebar = 'input_resep', permission_code = 'pasien.resep'
WHERE nama_sidebar = 'Input Resep';
