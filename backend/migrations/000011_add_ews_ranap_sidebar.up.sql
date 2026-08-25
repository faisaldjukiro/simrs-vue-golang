INSERT INTO sidebar_pasien (
  kode_sidebar, nama_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at
)
SELECT
  'ews_ranap', 'EWS Ranap', 'Activity', JSON_ARRAY('Rawat Inap'), 13, TRUE, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM sidebar_pasien
  WHERE kode_sidebar = 'ews_ranap' OR nama_sidebar = 'EWS Ranap'
);

UPDATE sidebar_pasien
SET
  kode_sidebar = 'ews_ranap',
  nama_sidebar = 'EWS Ranap',
  ikon = 'Activity',
  daftar_modul = JSON_ARRAY('Rawat Inap'),
  urutan = 13,
  aktif = TRUE,
  updated_at = NOW()
WHERE kode_sidebar = 'ews_ranap' OR nama_sidebar = 'EWS Ranap';
