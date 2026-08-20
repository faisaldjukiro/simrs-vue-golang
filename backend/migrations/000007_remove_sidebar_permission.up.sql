ALTER TABLE sidebar_pasien
    DROP FOREIGN KEY sidebar_pasien_permission_code_foreign;

ALTER TABLE sidebar_pasien
    DROP INDEX sidebar_pasien_permission_code_index,
    DROP COLUMN permission_code;

DELETE akses
FROM user_permissions akses
INNER JOIN permissions izin ON izin.id = akses.permission_id
WHERE izin.code LIKE 'pasien.%';

DELETE FROM permissions
WHERE code LIKE 'pasien.%';
