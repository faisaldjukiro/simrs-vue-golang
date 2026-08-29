INSERT INTO permissions (`group`, name, code, created_at, updated_at)
VALUES ('Master Setting', 'Master Ventilator', 'master_ventilator', NOW(), NOW())
ON DUPLICATE KEY UPDATE name = VALUES(name), `group` = VALUES(`group`), updated_at = NOW();

INSERT INTO menu_navigasi
  (tipe, label, deskripsi, ikon, tone, route, permissions, urutan, aktif, created_at, updated_at)
SELECT
  'dashboard', 'Master Ventilator', 'Kelola perangkat ventilator yang digunakan pada pelayanan rawat inap.',
  'Wind', 'cyan', NULL, JSON_ARRAY('master_ventilator'), 14, TRUE, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM menu_navigasi WHERE tipe = 'dashboard' AND label = 'Master Ventilator'
);

INSERT INTO sidebar_pasien
  (kode_sidebar, nama_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
SELECT
  'ventilator', 'Ventilator', 'Wind', JSON_ARRAY('Rawat Inap'), 49, TRUE, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM sidebar_pasien WHERE kode_sidebar = 'ventilator'
);
