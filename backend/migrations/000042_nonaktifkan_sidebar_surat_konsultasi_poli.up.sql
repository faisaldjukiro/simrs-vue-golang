-- Hanya konfigurasi sidebar pada DB_*; surat dan rujukan di SIMRS tetap utuh.
-- Pertahankan konfigurasi lama agar dapat diaktifkan lagi melalui Kelola Menu.
UPDATE sidebar_pasien
SET aktif = FALSE, updated_at = NOW()
WHERE nama_sidebar = 'Surat Konsultasi Ke Poli'
   OR kode_sidebar IN ('surat_konsultasi_ke_poli', 'surat_konsultasi_poli');
