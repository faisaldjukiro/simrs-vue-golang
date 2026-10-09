UPDATE sidebar_pasien
SET nama_sidebar = 'Rujuk Internal Rawat Inap', updated_at = NOW()
WHERE kode_sidebar = 'rujukan_internal_ranap'
  AND nama_sidebar = 'Rujuk Internal';
