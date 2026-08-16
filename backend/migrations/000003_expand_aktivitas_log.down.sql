DELETE FROM permissions WHERE code = 'sistem.audit_log';

ALTER TABLE aktivitas_log
    DROP INDEX aktivitas_log_status_index,
    DROP INDEX aktivitas_log_target_index,
    DROP INDEX aktivitas_log_endpoint_index,
    DROP INDEX aktivitas_log_request_id_unique,
    DROP INDEX aktivitas_log_user_id_index,
    DROP COLUMN pesan_error,
    DROP COLUMN durasi_ms,
    DROP COLUMN berhasil,
    DROP COLUMN response_status,
    DROP COLUMN request_data,
    DROP COLUMN query_params,
    DROP COLUMN target_id,
    DROP COLUMN endpoint,
    DROP COLUMN method,
    DROP COLUMN request_id,
    DROP COLUMN user_id;
