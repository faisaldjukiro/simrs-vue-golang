UPDATE menu_navigasi
SET deskripsi = 'Frekuensi rata-rata penggunaan bed per bangsal berdasarkan pasien keluar.',
    updated_at = NOW()
WHERE tipe = 'dashboard'
  AND label = 'Penggunaan Bed & Frekuensi';
