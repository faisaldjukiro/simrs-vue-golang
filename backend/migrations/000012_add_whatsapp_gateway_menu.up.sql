INSERT INTO permissions (`group`, name, code, created_at, updated_at)
VALUES ('Integrasi', 'WhatsApp Gateway', 'whatsapp_gateway', NOW(), NOW())
ON DUPLICATE KEY UPDATE name = VALUES(name), `group` = VALUES(`group`), updated_at = NOW();

INSERT INTO menu_navigasi
  (tipe, label, deskripsi, ikon, tone, route, permissions, urutan, aktif, created_at, updated_at)
SELECT
  'dashboard', 'WhatsApp Gateway', 'Kelola perangkat dan kirim pesan WhatsApp dari SIRAPI.',
  'MessageCircle', 'teal', NULL, JSON_ARRAY('whatsapp_gateway'), 12, TRUE, NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM menu_navigasi WHERE tipe = 'dashboard' AND label = 'WhatsApp Gateway'
);
