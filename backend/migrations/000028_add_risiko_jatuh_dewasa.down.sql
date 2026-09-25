-- Nonaktifkan menu saja; tidak menghapus rekam medis.
UPDATE sidebar_pasien SET aktif = FALSE, updated_at = NOW()
WHERE kode_sidebar = 'lanjutan_risiko_jatuh_dewasa';
