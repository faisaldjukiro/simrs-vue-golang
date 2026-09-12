-- Hanya DB_* lokal. Tidak ada perubahan struktur di Khanza.
ALTER TABLE sirapi_rujukan_internal ADD COLUMN dikirim_pada DATETIME NULL;
