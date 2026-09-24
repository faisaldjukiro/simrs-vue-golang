-- Menonaktifkan modul tanpa membuang rekam klinis yang telah disimpan.
UPDATE sidebar_pasien SET aktif = FALSE, updated_at = NOW()
WHERE kode_sidebar = 'checklist_pre_operasi';
