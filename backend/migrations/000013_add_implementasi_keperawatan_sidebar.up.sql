INSERT INTO sidebar_pasien (
  kode_sidebar, nama_sidebar, ikon, daftar_modul, urutan, aktif, created_at, updated_at
)
SELECT
  'implementasi_keperawatan', 'Implementasi Keperawatan', 'ClipboardPenLine', JSON_ARRAY('Rawat Inap'), 48, TRUE, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM sidebar_pasien
  WHERE kode_sidebar = 'implementasi_keperawatan' OR nama_sidebar = 'Implementasi Keperawatan'
);

UPDATE sidebar_pasien target
LEFT JOIN sidebar_pasien pemilik_kode
  ON pemilik_kode.kode_sidebar = 'implementasi_keperawatan'
  AND pemilik_kode.id <> target.id
SET
  target.kode_sidebar = 'implementasi_keperawatan',
  target.nama_sidebar = 'Implementasi Keperawatan',
  target.ikon = 'ClipboardPenLine',
  target.daftar_modul = JSON_ARRAY('Rawat Inap'),
  target.urutan = 48,
  target.aktif = TRUE,
  target.updated_at = NOW()
WHERE target.nama_sidebar = 'Implementasi Keperawatan'
  AND pemilik_kode.id IS NULL;

UPDATE sidebar_pasien
SET aktif = FALSE, updated_at = NOW()
WHERE nama_sidebar = 'Imp. Keperawatan';
