DELETE FROM sidebar_pasien WHERE kode_sidebar = 'ventilator';
DELETE FROM menu_navigasi WHERE tipe = 'dashboard' AND label = 'Master Ventilator';
DELETE FROM permissions WHERE code = 'master_ventilator';
