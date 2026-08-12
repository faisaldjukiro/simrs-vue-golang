-- Data awal dari seeder simrs-lama. Jalankan hanya pada database aplikasi.
START TRANSACTION;

INSERT INTO permissions (`group`, name, code, created_at, updated_at) VALUES
    ('Pelayanan', 'Registrasi', 'registrasi', NOW(), NOW()),
    ('Pelayanan', 'IGD/UGD', 'igd', NOW(), NOW()),
    ('Penunjang', 'Laboratorium', 'periksa_lab', NOW(), NOW()),
    ('Penunjang', 'Permintaan Lab', 'permintaan_lab', NOW(), NOW()),
    ('Penunjang', 'Radiologi', 'periksa_radiologi', NOW(), NOW()),
    ('Penunjang', 'Permintaan Radiologi', 'permintaan_radiologi', NOW(), NOW()),
    ('Farmasi', 'Master Obat', 'obat', NOW(), NOW()),
    ('Farmasi', 'Beri Obat', 'beri_obat', NOW(), NOW()),
    ('Farmasi', 'Resep Obat', 'resep_obat', NOW(), NOW()),
    ('Rawat Inap', 'Kamar Inap', 'kamar_inap', NOW(), NOW()),
    ('Rawat Inap', 'Daftar Pasien Ranap', 'daftar_pasien_ranap', NOW(), NOW()),
    ('Rawat Jalan', 'Tindakan Rawat Jalan', 'tindakan_ralan', NOW(), NOW()),
    ('Rawat Jalan', 'Billing Rawat Jalan', 'billing_ralan', NOW(), NOW()),
    ('Klaim', 'E-Klaim iDRG/INACBG', 'eklaim', NOW(), NOW()),
    ('Sistem', 'Kelola Menu Navigasi', 'kelola_menu', NOW(), NOW()),
    ('Sistem', 'Akses Penuh', '*', NOW(), NOW())
ON DUPLICATE KEY UPDATE
    `group` = VALUES(`group`),
    name = VALUES(name),
    updated_at = NOW();

-- Sama dengan akun awal Laravel. Password wajib diganti setelah login pertama.
INSERT INTO users (username, name, email, password, is_active, created_at, updated_at)
VALUES ('admin', 'Administrator', 'admin@gmail.com', '$2y$12$/CBvNutc7PARdjAeInE5NeQ6DCUzxkiSQrfMGyY25LV5EOWFjfysG', TRUE, NOW(), NOW())
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    is_active = TRUE,
    updated_at = NOW();

INSERT INTO user_permissions (user_id, permission_id, granted_by, created_at, updated_at)
SELECT users.id, permissions.id, NULL, NOW(), NOW()
FROM users
JOIN permissions ON permissions.code = '*'
WHERE users.username = 'admin'
ON DUPLICATE KEY UPDATE updated_at = NOW();

DELETE FROM menu_navigasi;

INSERT INTO menu_navigasi
    (tipe, label, deskripsi, ikon, tone, route, permissions, urutan, aktif, created_at, updated_at)
VALUES
    ('ribbon', 'Menu', NULL, 'Home', 'slate', NULL, NULL, 1, TRUE, NOW(), NOW()),
    ('ribbon', 'Registrasi', NULL, 'ClipboardList', 'blue', NULL, JSON_ARRAY('registrasi'), 2, TRUE, NOW(), NOW()),
    ('ribbon', 'IGD/UGD', NULL, 'HeartPulse', 'rose', NULL, JSON_ARRAY('igd'), 3, TRUE, NOW(), NOW()),
    ('ribbon', 'Laborat', NULL, 'FlaskConical', 'amber', NULL, JSON_ARRAY('periksa_lab', 'permintaan_lab'), 4, TRUE, NOW(), NOW()),
    ('ribbon', 'Radiologi', NULL, 'Microscope', 'violet', NULL, JSON_ARRAY('periksa_radiologi', 'permintaan_radiologi'), 5, TRUE, NOW(), NOW()),
    ('ribbon', 'Farmasi', NULL, 'Pill', 'teal', NULL, JSON_ARRAY('obat', 'beri_obat', 'resep_obat'), 6, TRUE, NOW(), NOW()),
    ('ribbon', 'Rawat Inap', NULL, 'Bed', 'indigo', NULL, JSON_ARRAY('kamar_inap', 'daftar_pasien_ranap'), 7, TRUE, NOW(), NOW()),
    ('ribbon', 'Rawat Jalan', NULL, 'Stethoscope', 'cyan', NULL, JSON_ARRAY('registrasi', 'tindakan_ralan', 'billing_ralan'), 8, TRUE, NOW(), NOW()),
    ('ribbon', 'Log Out', NULL, 'LogOut', 'slate', NULL, NULL, 99, TRUE, NOW(), NOW()),
    ('dashboard', 'Registrasi', 'Pendaftaran dan daftar kunjungan pasien hari ini.', 'ClipboardList', 'blue', NULL, JSON_ARRAY('registrasi'), 1, TRUE, NOW(), NOW()),
    ('dashboard', 'IGD/UGD', 'Daftar pasien instalasi gawat darurat.', 'HeartPulse', 'rose', NULL, JSON_ARRAY('igd'), 2, TRUE, NOW(), NOW()),
    ('dashboard', 'Laborat', 'Pemeriksaan dan permintaan laboratorium.', 'FlaskConical', 'amber', NULL, JSON_ARRAY('periksa_lab', 'permintaan_lab'), 3, TRUE, NOW(), NOW()),
    ('dashboard', 'Radiologi', 'Pemeriksaan dan permintaan radiologi.', 'Microscope', 'violet', NULL, JSON_ARRAY('periksa_radiologi', 'permintaan_radiologi'), 4, TRUE, NOW(), NOW()),
    ('dashboard', 'Farmasi', 'Resep, obat, dan pelayanan farmasi.', 'Pill', 'teal', NULL, JSON_ARRAY('obat', 'beri_obat', 'resep_obat'), 5, TRUE, NOW(), NOW()),
    ('dashboard', 'Rawat Inap', 'Kamar inap dan daftar pasien rawat inap.', 'Bed', 'indigo', NULL, JSON_ARRAY('kamar_inap', 'daftar_pasien_ranap'), 6, TRUE, NOW(), NOW()),
    ('dashboard', 'Rawat Jalan', 'Daftar pasien, tindakan, dan billing rawat jalan.', 'Stethoscope', 'cyan', NULL, JSON_ARRAY('registrasi', 'tindakan_ralan', 'billing_ralan'), 7, TRUE, NOW(), NOW()),
    ('dashboard', 'User', 'Kelola permission dan akses menu petugas.', 'UsersRound', 'slate', NULL, JSON_ARRAY('*'), 8, TRUE, NOW(), NOW()),
    ('dashboard', 'IDRG', 'Bridging klaim BPJS E-Klaim iDRG / INA-CBG.', 'FileSpreadsheet', 'emerald', '/faisal/eklaim', JSON_ARRAY('eklaim'), 9, TRUE, NOW(), NOW()),
    ('dashboard', 'Kelola Menu', 'Tambah, ubah, hapus menu navigasi ribbon & dashboard.', 'LayoutDashboard', 'slate', '/admin/menu', JSON_ARRAY('kelola_menu'), 10, TRUE, NOW(), NOW());

DELETE FROM menu_workspace_pasien;

INSERT INTO menu_workspace_pasien
    (nama_menu, ikon, daftar_modul, urutan, aktif, created_at, updated_at)
VALUES
    ('Cppt/Soap', 'FileSignature', JSON_ARRAY('Rawat Inap'), 1, TRUE, NOW(), NOW()),
    ('Penangangan Dokter & Petugas', 'Users', JSON_ARRAY('Rawat Inap'), 2, TRUE, NOW(), NOW()),
    ('SBAR', 'MessageSquareText', JSON_ARRAY('Rawat Inap'), 3, TRUE, NOW(), NOW()),
    ('Diagnosa', 'Stethoscope', JSON_ARRAY('Rawat Inap'), 4, TRUE, NOW(), NOW()),
    ('Riwayat Pasien', 'History', JSON_ARRAY('Rawat Inap'), 5, TRUE, NOW(), NOW()),
    ('Input Resep', 'Pill', JSON_ARRAY('IGD/UGD', 'Rawat Jalan', 'Rawat Inap'), 6, TRUE, NOW(), NOW()),
    ('Copy Resep', 'Copy', JSON_ARRAY('Rawat Inap'), 7, TRUE, NOW(), NOW()),
    ('Resep Luar', 'Receipt', JSON_ARRAY('Rawat Inap'), 8, TRUE, NOW(), NOW()),
    ('Verifikasi SBAR', 'ShieldCheck', JSON_ARRAY('Rawat Inap'), 9, TRUE, NOW(), NOW()),
    ('Rujuk Internal', 'ArrowRightToLine', JSON_ARRAY('Rawat Inap'), 10, TRUE, NOW(), NOW()),
    ('Surat Konsultasi Ke Poli', 'Send', JSON_ARRAY('Rawat Inap'), 11, TRUE, NOW(), NOW()),
    ('Lembar Konsultasi', 'ClipboardList', JSON_ARRAY('Rawat Inap'), 12, TRUE, NOW(), NOW()),
    ('EWS Ranap', 'Activity', JSON_ARRAY('Rawat Inap'), 13, TRUE, NOW(), NOW()),
    ('Permintaan Stok Pasien', 'PackagePlus', JSON_ARRAY('Rawat Inap'), 14, TRUE, NOW(), NOW()),
    ('Permintaan Resep Pulang', 'BriefcaseMedical', JSON_ARRAY('Rawat Inap'), 15, TRUE, NOW(), NOW()),
    ('Input Obat & BHP', 'Syringe', JSON_ARRAY('Rawat Inap'), 16, TRUE, NOW(), NOW()),
    ('Data Obat & BHP', 'Database', JSON_ARRAY('Rawat Inap'), 17, TRUE, NOW(), NOW()),
    ('Berkas Digital', 'FolderOpen', JSON_ARRAY('Rawat Inap'), 18, TRUE, NOW(), NOW()),
    ('Permintaan Lab', 'FlaskConical', JSON_ARRAY('IGD/UGD', 'Rawat Jalan', 'Rawat Inap'), 19, TRUE, NOW(), NOW()),
    ('Permintaan Rad', 'ScanSearch', JSON_ARRAY('IGD/UGD', 'Rawat Jalan', 'Rawat Inap'), 20, TRUE, NOW(), NOW()),
    ('Jadwal Operasi', 'CalendarDays', JSON_ARRAY('Rawat Inap'), 21, TRUE, NOW(), NOW()),
    ('Surat Kontrol', 'CalendarCheck', JSON_ARRAY('Rawat Inap'), 22, TRUE, NOW(), NOW()),
    ('Rujuk Keluar', 'ExternalLink', JSON_ARRAY('Rawat Inap'), 23, TRUE, NOW(), NOW()),
    ('Diagnosa', 'Stethoscope', JSON_ARRAY('Rawat Inap'), 24, TRUE, NOW(), NOW()),
    ('Resume Pasien', 'NotebookText', JSON_ARRAY('Rawat Inap'), 25, TRUE, NOW(), NOW()),
    ('Penilaian Sentinel Pie', 'PieChart', JSON_ARRAY('Rawat Inap'), 26, TRUE, NOW(), NOW()),
    ('Awal Keperawatan Umum', 'HeartPulse', JSON_ARRAY('Rawat Inap'), 27, TRUE, NOW(), NOW()),
    ('Awal Keperawatan Kandungan', 'Baby', JSON_ARRAY('Rawat Inap'), 28, TRUE, NOW(), NOW()),
    ('Awal Fisioterapi', 'Dumbbell', JSON_ARRAY('Rawat Inap'), 29, TRUE, NOW(), NOW()),
    ('Edukasi Pasien', 'BookOpen', JSON_ARRAY('Rawat Inap'), 30, TRUE, NOW(), NOW()),
    ('Registrasi Kanker', 'Ribbon', JSON_ARRAY('Rawat Inap'), 31, TRUE, NOW(), NOW()),
    ('Awal Medis Umum', 'Stethoscope', JSON_ARRAY('Rawat Inap'), 32, TRUE, NOW(), NOW()),
    ('Awal Medis Kandungan', 'Baby', JSON_ARRAY('Rawat Inap'), 33, TRUE, NOW(), NOW()),
    ('Checklist Pre Operasi', 'ClipboardCheck', JSON_ARRAY('Rawat Inap'), 34, TRUE, NOW(), NOW()),
    ('Penilaian Pre Operasi', 'Scissors', JSON_ARRAY('Rawat Inap'), 35, TRUE, NOW(), NOW()),
    ('Penilaian Pre Anestesi', 'Syringe', JSON_ARRAY('Rawat Inap'), 36, TRUE, NOW(), NOW()),
    ('Penilaian Psikolog', 'Brain', JSON_ARRAY('Rawat Inap'), 37, TRUE, NOW(), NOW()),
    ('Perencanaan Pemulangan', 'House', JSON_ARRAY('Rawat Inap'), 38, TRUE, NOW(), NOW()),
    ('Lanjutan Risiko Jatuh Dewasa', 'TriangleAlert', JSON_ARRAY('Rawat Inap'), 39, TRUE, NOW(), NOW()),
    ('Lanjutan Risiko Jatuh Anak', 'TriangleAlert', JSON_ARRAY('Rawat Inap'), 40, TRUE, NOW(), NOW()),
    ('Tambahan Pasien Geriatri', 'UserCog', JSON_ARRAY('Rawat Inap'), 41, TRUE, NOW(), NOW()),
    ('Hasil Pemeriksaan USG', 'Image', JSON_ARRAY('Rawat Inap'), 42, TRUE, NOW(), NOW()),
    ('Catatan Pasien', 'NotepadText', JSON_ARRAY('Rawat Inap'), 43, TRUE, NOW(), NOW()),
    ('Observasi Ranap', 'Bed', JSON_ARRAY('Rawat Inap'), 44, TRUE, NOW(), NOW()),
    ('Observasi Kebidanan', 'Baby', JSON_ARRAY('Rawat Inap'), 45, TRUE, NOW(), NOW()),
    ('Observasi Post Partum', 'Baby', JSON_ARRAY('Rawat Inap'), 46, TRUE, NOW(), NOW()),
    ('Pengkajian Awal MPP', 'ClipboardList', JSON_ARRAY('Rawat Inap'), 47, TRUE, NOW(), NOW()),
    ('Implementasi Keperawatan', 'ClipboardPenLine', JSON_ARRAY('Rawat Inap'), 48, TRUE, NOW(), NOW()),
    ('Imp. Keperawatan', 'ClipboardPenLine', JSON_ARRAY('Rawat Inap'), 49, TRUE, NOW(), NOW()),
    ('Catatan Cek GDS', 'Droplet', JSON_ARRAY('Rawat Inap'), 50, TRUE, NOW(), NOW()),
    ('Pemantauan PEWS Anak', 'Activity', JSON_ARRAY('Rawat Inap'), 51, TRUE, NOW(), NOW()),
    ('Pemantauan NEWS Anak', 'Activity', JSON_ARRAY('Rawat Inap'), 52, TRUE, NOW(), NOW()),
    ('Pemantauan MEOWS', 'Activity', JSON_ARRAY('Rawat Inap'), 53, TRUE, NOW(), NOW()),
    ('Skrining Nutrisi Dewasa', 'Utensils', JSON_ARRAY('Rawat Inap'), 54, TRUE, NOW(), NOW()),
    ('Skrining Nutrisi Lansia', 'Carrot', JSON_ARRAY('Rawat Inap'), 55, TRUE, NOW(), NOW()),
    ('Skrining Nutrisi Anak', 'Apple', JSON_ARRAY('Rawat Inap'), 56, TRUE, NOW(), NOW()),
    ('Skrining Gizi Lanjut', 'Salad', JSON_ARRAY('Rawat Inap'), 57, TRUE, NOW(), NOW()),
    ('Asuhan Gizi', 'Utensils', JSON_ARRAY('Rawat Inap'), 58, TRUE, NOW(), NOW()),
    ('Monitoring Gizi', 'LineChart', JSON_ARRAY('Rawat Inap'), 59, TRUE, NOW(), NOW()),
    ('Konseling Farmasi', 'MessageCircle', JSON_ARRAY('Rawat Inap'), 60, TRUE, NOW(), NOW()),
    ('Informasi Obat', 'Pill', JSON_ARRAY('Rawat Inap'), 61, TRUE, NOW(), NOW()),
    ('Transfer Antar Ruang', 'ArrowRightLeft', JSON_ARRAY('Rawat Inap'), 62, TRUE, NOW(), NOW()),
    ('Riwayat Perawatan', 'ScrollText', JSON_ARRAY('Rawat Inap'), 63, TRUE, NOW(), NOW()),
    ('Triase IGD', 'Stethoscope', JSON_ARRAY('IGD/UGD'), 64, TRUE, NOW(), NOW());

COMMIT;
