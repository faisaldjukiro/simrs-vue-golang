// Package manajemen_pengguna membaca data user aplikasi lokal.
package manajemen_pengguna

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var ErrUsernameAtauEmailSudahAda = errors.New("username or email already exists")
var ErrPermissionTidakValid = errors.New("permission is invalid")
var ErrPenggunaTidakDitemukan = errors.New("user not found")

type Pengguna struct {
	ID         uint64   `json:"id"`
	Username   string   `json:"username"`
	Nama       string   `json:"nama"`
	Email      string   `json:"email"`
	Aktif      bool     `json:"aktif"`
	Permission []string `json:"permission"`
	DibuatPada string   `json:"dibuat_pada"`
}

type Permission struct {
	ID   uint64 `json:"id"`
	Grup string `json:"grup"`
	Nama string `json:"nama"`
	Kode string `json:"kode"`
}

type Pegawai struct {
	NIK     string `json:"nik"`
	Nama    string `json:"nama"`
	Jabatan string `json:"jabatan"`
}

type Ringkasan struct {
	JumlahPengguna int `json:"jumlah_pengguna"`
	JumlahAktif    int `json:"jumlah_aktif"`
	JumlahAdmin    int `json:"jumlah_admin"`
}

type DataManajemenPengguna struct {
	Ringkasan  Ringkasan    `json:"ringkasan"`
	Pengguna   []Pengguna   `json:"pengguna"`
	Permission []Permission `json:"permission"`
}

type InputTambahPengguna struct {
	Username       string
	Nama           string
	Email          string
	Aktif          bool
	KodePermission []string
}

type InputUbahAksesPengguna struct {
	Aktif          bool
	KodePermission []string
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

func (r *Repositori) Daftar(ctx context.Context) (DataManajemenPengguna, error) {
	pengguna, err := r.bacaPengguna(ctx)
	if err != nil {
		return DataManajemenPengguna{}, err
	}

	permission, err := r.bacaPermission(ctx)
	if err != nil {
		return DataManajemenPengguna{}, err
	}

	ringkasan := Ringkasan{JumlahPengguna: len(pengguna)}
	for _, item := range pengguna {
		if item.Aktif {
			ringkasan.JumlahAktif++
		}
		for _, kode := range item.Permission {
			if kode == "*" {
				ringkasan.JumlahAdmin++
				break
			}
		}
	}

	return DataManajemenPengguna{
		Ringkasan:  ringkasan,
		Pengguna:   pengguna,
		Permission: permission,
	}, nil
}

func (r *Repositori) CariPegawai(ctx context.Context, kataKunci string) ([]Pegawai, error) {
	if r.simrsDB == nil {
		return []Pegawai{}, nil
	}

	kataKunci = strings.TrimSpace(kataKunci)
	if len(kataKunci) < 2 {
		return []Pegawai{}, nil
	}

	seperti := "%" + kataKunci + "%"
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT nik, nama, COALESCE(jbtn, '')
		FROM pegawai
		WHERE nik LIKE ?
		   OR nama LIKE ?
		   OR jbtn LIKE ?
		ORDER BY nama
		LIMIT 20
	`, seperti, seperti, seperti)
	if err != nil {
		return nil, fmt.Errorf("read SIMRS employees: %w", err)
	}
	defer rows.Close()

	pegawai := make([]Pegawai, 0)
	for rows.Next() {
		var item Pegawai
		if err := rows.Scan(&item.NIK, &item.Nama, &item.Jabatan); err != nil {
			return nil, fmt.Errorf("scan SIMRS employee: %w", err)
		}
		pegawai = append(pegawai, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate SIMRS employees: %w", err)
	}
	return pegawai, nil
}

func (r *Repositori) Tambah(ctx context.Context, input InputTambahPengguna) (Pengguna, error) {
	username := strings.TrimSpace(input.Username)
	nama := strings.TrimSpace(input.Nama)
	email := strings.TrimSpace(input.Email)
	if nama == "" {
		nama = username
	}
	if email == "" {
		email = username + "@simrs.local"
	}

	passwordHash, err := buatPasswordLokalAcakHash()
	if err != nil {
		return Pengguna{}, fmt.Errorf("hash password: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Pengguna{}, fmt.Errorf("begin create user transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO users (username, name, email, password, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`, username, nama, email, string(passwordHash), input.Aktif)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return Pengguna{}, ErrUsernameAtauEmailSudahAda
		}
		return Pengguna{}, fmt.Errorf("create local user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return Pengguna{}, fmt.Errorf("read created user id: %w", err)
	}

	kodePermission := normalisasiKodePermission(input.KodePermission)
	if len(kodePermission) > 0 {
		daftarID, err := r.cariIDPermission(ctx, tx, kodePermission)
		if err != nil {
			return Pengguna{}, err
		}
		if len(daftarID) != len(kodePermission) {
			return Pengguna{}, ErrPermissionTidakValid
		}

		for _, permissionID := range daftarID {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO user_permissions (user_id, permission_id, granted_by, created_at, updated_at)
				VALUES (?, ?, NULL, NOW(), NOW())
			`, uint64(userID), permissionID); err != nil {
				return Pengguna{}, fmt.Errorf("grant local user permission: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return Pengguna{}, fmt.Errorf("commit create user transaction: %w", err)
	}

	return Pengguna{
		ID:         uint64(userID),
		Username:   username,
		Nama:       nama,
		Email:      email,
		Aktif:      input.Aktif,
		Permission: kodePermission,
	}, nil
}

func (r *Repositori) UbahAkses(ctx context.Context, userID uint64, input InputUbahAksesPengguna) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update user access transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE users
		SET is_active = ?, updated_at = NOW()
		WHERE id = ?
	`, input.Aktif, userID)
	if err != nil {
		return fmt.Errorf("update local user status: %w", err)
	}
	jumlahBerubah, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated user status rows: %w", err)
	}
	if jumlahBerubah == 0 {
		return ErrPenggunaTidakDitemukan
	}

	kodePermission := normalisasiKodePermission(input.KodePermission)
	daftarID, err := r.cariIDPermission(ctx, tx, kodePermission)
	if err != nil {
		return err
	}
	if len(daftarID) != len(kodePermission) {
		return ErrPermissionTidakValid
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM user_permissions
		WHERE user_id = ?
	`, userID); err != nil {
		return fmt.Errorf("delete local user permissions: %w", err)
	}

	for _, permissionID := range daftarID {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO user_permissions (user_id, permission_id, granted_by, created_at, updated_at)
			VALUES (?, ?, NULL, NOW(), NOW())
		`, userID, permissionID); err != nil {
			return fmt.Errorf("grant local user permission: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit update user access transaction: %w", err)
	}
	return nil
}

func (r *Repositori) bacaPengguna(ctx context.Context) ([]Pengguna, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			users.id, users.username, users.name, users.email, users.is_active,
			COALESCE(DATE_FORMAT(users.created_at, '%Y-%m-%d %H:%i'), ''),
			COALESCE(permissions.code, '')
		FROM users
		LEFT JOIN user_permissions ON user_permissions.user_id = users.id
		LEFT JOIN permissions ON permissions.id = user_permissions.permission_id
		ORDER BY users.name, users.id, permissions.`+"`group`"+`, permissions.name
	`)
	if err != nil {
		return nil, fmt.Errorf("read local users: %w", err)
	}
	defer rows.Close()

	urutan := make([]uint64, 0)
	terindeks := make(map[uint64]*Pengguna)
	for rows.Next() {
		var data Pengguna
		var kodePermission string
		if err := rows.Scan(
			&data.ID, &data.Username, &data.Nama, &data.Email, &data.Aktif,
			&data.DibuatPada, &kodePermission,
		); err != nil {
			return nil, fmt.Errorf("scan local user: %w", err)
		}

		item, sudahAda := terindeks[data.ID]
		if !sudahAda {
			data.Permission = make([]string, 0)
			terindeks[data.ID] = &data
			urutan = append(urutan, data.ID)
			item = &data
		}
		if kodePermission != "" {
			item.Permission = append(item.Permission, kodePermission)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate local users: %w", err)
	}

	hasil := make([]Pengguna, 0, len(urutan))
	for _, id := range urutan {
		hasil = append(hasil, *terindeks[id])
	}
	return hasil, nil
}

func (r *Repositori) cariIDPermission(ctx context.Context, tx *sql.Tx, kodePermission []string) ([]uint64, error) {
	if len(kodePermission) == 0 {
		return []uint64{}, nil
	}

	placeholder := strings.TrimSuffix(strings.Repeat("?,", len(kodePermission)), ",")
	args := make([]any, 0, len(kodePermission))
	for _, kode := range kodePermission {
		args = append(args, kode)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id
		FROM permissions
		WHERE code IN (`+placeholder+`)
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("read selected permissions: %w", err)
	}
	defer rows.Close()

	daftarID := make([]uint64, 0, len(kodePermission))
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan selected permission: %w", err)
		}
		daftarID = append(daftarID, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate selected permissions: %w", err)
	}
	return daftarID, nil
}

func (r *Repositori) bacaPermission(ctx context.Context) ([]Permission, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, `+"`group`"+`, name, code
		FROM permissions
		ORDER BY `+"`group`"+`, name
	`)
	if err != nil {
		return nil, fmt.Errorf("read local permissions: %w", err)
	}
	defer rows.Close()

	daftarPermission := make([]Permission, 0)
	for rows.Next() {
		var data Permission
		if err := rows.Scan(&data.ID, &data.Grup, &data.Nama, &data.Kode); err != nil {
			return nil, fmt.Errorf("scan local permission: %w", err)
		}
		daftarPermission = append(daftarPermission, data)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate local permissions: %w", err)
	}
	return daftarPermission, nil
}

func normalisasiKodePermission(kodePermission []string) []string {
	terpilih := make([]string, 0, len(kodePermission))
	sudahAda := make(map[string]bool)
	for _, kode := range kodePermission {
		kode = strings.TrimSpace(kode)
		if kode == "" || sudahAda[kode] {
			continue
		}
		sudahAda[kode] = true
		terpilih = append(terpilih, kode)
	}
	return terpilih
}

func buatPasswordLokalAcakHash() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate random local password: %w", err)
	}

	passwordAcak := hex.EncodeToString(buffer)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(passwordAcak), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(passwordHash), nil
}
