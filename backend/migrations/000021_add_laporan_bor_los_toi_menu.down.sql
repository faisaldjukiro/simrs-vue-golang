DELETE FROM menu_navigasi
WHERE tipe = 'dashboard'
  AND label = 'Laporan BOR, LOS & TOI';

DELETE FROM role_permissions
WHERE permission_id IN (
  SELECT id FROM permissions WHERE code = 'laporan_bor_los_toi'
);

DELETE FROM permissions WHERE code = 'laporan_bor_los_toi';
