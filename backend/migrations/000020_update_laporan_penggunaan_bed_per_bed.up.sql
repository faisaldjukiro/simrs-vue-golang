UPDATE menu_navigasi
SET deskripsi = 'Frekuensi penggunaan setiap bed berdasarkan periode.',
    updated_at = NOW()
WHERE tipe = 'dashboard'
  AND label = 'Penggunaan Bed & Frekuensi';
