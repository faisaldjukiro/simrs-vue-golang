UPDATE menu_navigasi
SET deskripsi = 'BOR, frekuensi penggunaan bed, dan BTO per bangsal.',
    updated_at = NOW()
WHERE tipe = 'dashboard' AND label = 'Penggunaan Bed & Frekuensi';
