ALTER TABLE aktivitas_log
    ADD COLUMN user_id BIGINT UNSIGNED NULL AFTER id,
    ADD COLUMN request_id CHAR(32) NULL AFTER user_id,
    ADD COLUMN method VARCHAR(10) NULL AFTER modul,
    ADD COLUMN endpoint VARCHAR(255) NULL AFTER method,
    ADD COLUMN target_id VARCHAR(191) NULL AFTER tabel_target,
    ADD COLUMN query_params JSON NULL AFTER target_id,
    ADD COLUMN request_data JSON NULL AFTER query_params,
    ADD COLUMN response_status SMALLINT UNSIGNED NULL AFTER data_sesudah,
    ADD COLUMN berhasil BOOLEAN NOT NULL DEFAULT TRUE AFTER response_status,
    ADD COLUMN durasi_ms BIGINT UNSIGNED NULL AFTER berhasil,
    ADD COLUMN pesan_error VARCHAR(500) NULL AFTER durasi_ms,
    ADD KEY aktivitas_log_user_id_index (user_id),
    ADD UNIQUE KEY aktivitas_log_request_id_unique (request_id),
    ADD KEY aktivitas_log_endpoint_index (endpoint),
    ADD KEY aktivitas_log_target_index (modul, target_id),
    ADD KEY aktivitas_log_status_index (response_status);

INSERT INTO permissions (`group`, name, code, created_at, updated_at)
VALUES ('Sistem', 'Log Aktivitas', 'sistem.audit_log', NOW(), NOW())
ON DUPLICATE KEY UPDATE
    `group` = VALUES(`group`),
    name = VALUES(name),
    updated_at = NOW();
