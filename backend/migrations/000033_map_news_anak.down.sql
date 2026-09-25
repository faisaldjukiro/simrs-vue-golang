-- Tidak menghapus sidebar atau konfigurasi akses pengguna.
UPDATE sidebar_pasien
SET kode_sidebar = CONCAT('sidebar_lama_', LPAD(id, 4, '0')), updated_at = NOW()
WHERE kode_sidebar = 'news_anak';
