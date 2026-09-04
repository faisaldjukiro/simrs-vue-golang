INSERT INTO permissions (`group`, name, code, created_at, updated_at)
VALUES ('Laporan', 'Laporan BOR, LOS & TOI', 'laporan_bor_los_toi', NOW(), NOW())
ON DUPLICATE KEY UPDATE name = VALUES(name), updated_at = NOW();

INSERT INTO menu_navigasi (
  tipe,
  label,
  deskripsi,
  ikon,
  tone,
  route,
  permissions,
  urutan,
  aktif,
  created_at,
  updated_at
)
SELECT
  'dashboard',
  'Laporan BOR, LOS & TOI',
  'Indikator pemanfaatan tempat tidur rawat inap per bulan.',
  'BedDouble',
  'teal',
  NULL,
  JSON_ARRAY('laporan_bor_los_toi'),
  17,
  TRUE,
  NOW(),
  NOW()
WHERE NOT EXISTS (
  SELECT 1
  FROM menu_navigasi
  WHERE tipe = 'dashboard'
    AND label = 'Laporan BOR, LOS & TOI'
);
