UPDATE menu_navigasi
SET deskripsi = 'Jumlah bed aktif dan frekuensi penggunaannya per bangsal.',
    updated_at = NOW()
WHERE tipe = 'dashboard' AND label = 'Penggunaan Bed & Frekuensi';
