package autentikasi

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	kunciIDUserSIMRS   = "nur"
	kunciPasswordSIMRS = "windi"
	ContextKeyPengguna = "authenticated_user"
	ContextKeyToken    = "access_token"
)

type Pengguna struct {
	ID           uint64   `json:"id"`
	Username     string   `json:"username"`
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	PasswordHash string   `json:"-"`
	IsActive     bool     `json:"is_active"`
	Permissions  []string `json:"permissions"`
}

type PenggunaSIMRS struct {
	Username string
	Password string
	Nama     string
}

type Repositori struct {
	db      *sql.DB
	simrsDB *sql.DB
}

func NewRepositori(db *sql.DB, simrsDB ...*sql.DB) *Repositori {
	repositori := &Repositori{db: db}
	if len(simrsDB) > 0 {
		repositori.simrsDB = simrsDB[0]
	}
	return repositori
}

func (r *Repositori) CariPenggunaByUsername(ctx context.Context, username string) (Pengguna, error) {
	var user Pengguna
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, name, email, password, is_active
		FROM users
		WHERE username = ?
		LIMIT 1
	`, username).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
	)
	if err != nil {
		return Pengguna{}, err
	}
	if err := r.isiPermissions(ctx, &user); err != nil {
		return Pengguna{}, err
	}
	return user, nil
}

func (r *Repositori) CariPenggunaSIMRS(ctx context.Context, username string) (PenggunaSIMRS, error) {
	if r.simrsDB == nil {
		return PenggunaSIMRS{}, sql.ErrNoRows
	}

	var user PenggunaSIMRS
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT
			CAST(AES_DECRYPT(`+"`id_user`"+`, ?) AS CHAR(255)) AS username_asli,
			CAST(AES_DECRYPT(`+"`password`"+`, ?) AS CHAR(255)) AS password_asli,
			COALESCE(pegawai.nama, '')
		FROM `+"`user`"+`
		LEFT JOIN pegawai
			ON pegawai.nik = CAST(AES_DECRYPT(`+"`user`"+`.`+"`id_user`"+`, ?) AS CHAR(255))
		WHERE CAST(AES_DECRYPT(`+"`id_user`"+`, ?) AS CHAR(255)) = ?
		LIMIT 1
	`, kunciIDUserSIMRS, kunciPasswordSIMRS, kunciIDUserSIMRS, kunciIDUserSIMRS, username).Scan(
		&user.Username,
		&user.Password,
		&user.Nama,
	)
	if err != nil {
		return PenggunaSIMRS{}, err
	}
	return user, nil
}

func (r *Repositori) SimpanPenggunaDariSIMRS(ctx context.Context, username, nama, passwordHash string) (Pengguna, error) {
	if nama == "" {
		nama = username
	}
	email := username + "@simrs.local"

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO users (username, name, email, password, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, TRUE, NOW(), NOW())
	`, username, nama, email, passwordHash)
	if err != nil {
		return Pengguna{}, fmt.Errorf("create local user from SIMRS: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return Pengguna{}, fmt.Errorf("read created SIMRS local user id: %w", err)
	}

	return Pengguna{
		ID:       uint64(userID),
		Username: username,
		Name:     nama,
		Email:    email,
		IsActive: true,
	}, nil
}

func (r *Repositori) SimpanToken(ctx context.Context, userID uint64, tokenHash string, expiresAt time.Time) error {
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO access_tokens (user_id, token_hash, expires_at)
		VALUES (?, ?, ?)
	`, userID, tokenHash, expiresAt); err != nil {
		return fmt.Errorf("store access token: %w", err)
	}
	return nil
}

func (r *Repositori) CariPenggunaByToken(ctx context.Context, tokenHash string) (Pengguna, error) {
	var user Pengguna
	err := r.db.QueryRowContext(ctx, `
		SELECT users.id, users.username, users.name, users.email, users.is_active
		FROM access_tokens
		JOIN users ON users.id = access_tokens.user_id
		WHERE access_tokens.token_hash = ?
		  AND access_tokens.expires_at > UTC_TIMESTAMP()
		  AND users.is_active = TRUE
		LIMIT 1
	`, tokenHash).Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.Email,
		&user.IsActive,
	)
	if err != nil {
		return Pengguna{}, err
	}
	if err := r.isiPermissions(ctx, &user); err != nil {
		return Pengguna{}, err
	}

	_, _ = r.db.ExecContext(ctx, `
		UPDATE access_tokens SET last_used_at = UTC_TIMESTAMP()
		WHERE token_hash = ?
	`, tokenHash)
	return user, nil
}

func (r *Repositori) isiPermissions(ctx context.Context, user *Pengguna) error {
	rows, err := r.db.QueryContext(ctx, `
		SELECT permissions.code
		FROM user_permissions
		INNER JOIN permissions ON permissions.id = user_permissions.permission_id
		WHERE user_permissions.user_id = ?
		ORDER BY permissions.code
	`, user.ID)
	if err != nil {
		return fmt.Errorf("load user permissions: %w", err)
	}
	defer rows.Close()

	user.Permissions = make([]string, 0)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return fmt.Errorf("scan user permission: %w", err)
		}
		user.Permissions = append(user.Permissions, code)
	}
	return rows.Err()
}

func (r *Repositori) HapusToken(ctx context.Context, tokenHash string) error {
	if _, err := r.db.ExecContext(ctx, "DELETE FROM access_tokens WHERE token_hash = ?", tokenHash); err != nil {
		return fmt.Errorf("delete access token: %w", err)
	}
	return nil
}

func (r *Repositori) HapusTokenKedaluwarsa(ctx context.Context) {
	_, _ = r.db.ExecContext(ctx, "DELETE FROM access_tokens WHERE expires_at <= UTC_TIMESTAMP()")
}

func (r *Repositori) MemilikiPermission(ctx context.Context, userID uint64, daftarKode ...string) (bool, error) {
	if len(daftarKode) == 0 {
		return false, nil
	}

	placeholder := make([]string, 0, len(daftarKode))
	parameter := make([]any, 0, len(daftarKode)+1)
	parameter = append(parameter, userID)
	for range daftarKode {
		placeholder = append(placeholder, "?")
	}
	for _, kode := range daftarKode {
		parameter = append(parameter, kode)
	}

	var ada bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM user_permissions akses
			INNER JOIN permissions izin ON izin.id = akses.permission_id
			WHERE akses.user_id = ?
			  AND (izin.code = '*' OR izin.code IN (`+strings.Join(placeholder, ",")+`))
		)
	`, parameter...).Scan(&ada)
	if err != nil {
		return false, fmt.Errorf("check user permission: %w", err)
	}
	return ada, nil
}
