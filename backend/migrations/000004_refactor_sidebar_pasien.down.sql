ALTER TABLE sidebar_pasien
    DROP FOREIGN KEY sidebar_pasien_permission_code_foreign,
    DROP INDEX sidebar_pasien_permission_code_index,
    DROP INDEX sidebar_pasien_kode_sidebar_unique,
    DROP COLUMN permission_code,
    DROP COLUMN kode_sidebar,
    CHANGE COLUMN nama_sidebar nama_menu VARCHAR(80) NOT NULL;

RENAME TABLE sidebar_pasien TO menu_workspace_pasien;

DELETE FROM permissions
WHERE code IN (
    'pasien.cppt',
    'pasien.penanganan_dokter_petugas',
    'pasien.diagnosa',
    'pasien.riwayat_perawatan',
    'pasien.triase_igd',
    'pasien.awal_keperawatan_igd',
    'pasien.resume_pasien'
);
