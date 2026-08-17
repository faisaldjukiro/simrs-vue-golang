// Package kelola_menu mengelola konfigurasi sidebar pasien pada database aplikasi.
package kelola_menu

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrSidebarTidakDitemukan = errors.New("sidebar tidak ditemukan")
	ErrKodeSidebarDigunakan  = errors.New("kode sidebar sudah digunakan")
	ErrPermissionTidakValid  = errors.New("permission tidak valid")
)

var polaKodeSidebar = regexp.MustCompile(`^[a-z0-9_]+$`)

type Sidebar struct {
	ID             uint64   `json:"id"`
	Kode           string   `json:"kode"`
	Nama           string   `json:"nama"`
	Ikon           string   `json:"ikon"`
	DaftarModul    []string `json:"daftar_modul"`
	PermissionCode string   `json:"permission_code"`
	Urutan         uint16   `json:"urutan"`
	Aktif          bool     `json:"aktif"`
}

type Permission struct {
	Grup string `json:"grup"`
	Nama string `json:"nama"`
	Kode string `json:"kode"`
}

type Data struct {
	Sidebar      []Sidebar    `json:"sidebar"`
	Permission   []Permission `json:"permission"`
	PilihanModul []string     `json:"pilihan_modul"`
}

type InputSidebar struct {
	Kode           string
	Nama           string
	Ikon           string
	DaftarModul    []string
	PermissionCode string
	Urutan         uint16
	Aktif          bool
}

type Repositori struct{ db *sql.DB }

func NewRepositori(db *sql.DB) *Repositori { return &Repositori{db: db} }

func (r *Repositori) Daftar(ctx context.Context) (Data, error) {
	sidebar, err := r.bacaSidebar(ctx)
	if err != nil {
		return Data{}, err
	}
	permission, err := r.bacaPermission(ctx)
	if err != nil {
		return Data{}, err
	}
	return Data{
		Sidebar:      sidebar,
		Permission:   permission,
		PilihanModul: []string{"IGD/UGD", "Rawat Jalan", "Rawat Inap"},
	}, nil
}

func (r *Repositori) Tambah(ctx context.Context, input InputSidebar) (Sidebar, error) {
	input = normalisasiInput(input)
	if err := r.validasiPermission(ctx, input.PermissionCode); err != nil {
		return Sidebar{}, err
	}
	daftarModul, err := json.Marshal(input.DaftarModul)
	if err != nil {
		return Sidebar{}, fmt.Errorf("encode sidebar modules: %w", err)
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO sidebar_pasien
			(kode_sidebar,nama_sidebar,ikon,daftar_modul,permission_code,urutan,aktif,created_at,updated_at)
		VALUES (?,?,?,?,NULLIF(?,''),?,?,NOW(),NOW())
	`, input.Kode, input.Nama, input.Ikon, daftarModul, input.PermissionCode, input.Urutan, input.Aktif)
	if duplikat(err) {
		return Sidebar{}, ErrKodeSidebarDigunakan
	}
	if err != nil {
		return Sidebar{}, fmt.Errorf("create sidebar: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Sidebar{}, fmt.Errorf("read created sidebar id: %w", err)
	}
	return Sidebar{ID: uint64(id), Kode: input.Kode, Nama: input.Nama, Ikon: input.Ikon, DaftarModul: input.DaftarModul, PermissionCode: input.PermissionCode, Urutan: input.Urutan, Aktif: input.Aktif}, nil
}

func (r *Repositori) Ubah(ctx context.Context, id uint64, input InputSidebar) error {
	input = normalisasiInput(input)
	if err := r.validasiPermission(ctx, input.PermissionCode); err != nil {
		return err
	}
	daftarModul, err := json.Marshal(input.DaftarModul)
	if err != nil {
		return fmt.Errorf("encode sidebar modules: %w", err)
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE sidebar_pasien
		SET kode_sidebar=?,nama_sidebar=?,ikon=?,daftar_modul=?,permission_code=NULLIF(?,''),urutan=?,aktif=?,updated_at=NOW()
		WHERE id=?
	`, input.Kode, input.Nama, input.Ikon, daftarModul, input.PermissionCode, input.Urutan, input.Aktif, id)
	if duplikat(err) {
		return ErrKodeSidebarDigunakan
	}
	if err != nil {
		return fmt.Errorf("update sidebar: %w", err)
	}
	jumlah, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated sidebar rows: %w", err)
	}
	if jumlah == 0 {
		var ada bool
		if err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM sidebar_pasien WHERE id=?)", id).Scan(&ada); err != nil {
			return fmt.Errorf("check sidebar: %w", err)
		}
		if !ada {
			return ErrSidebarTidakDitemukan
		}
	}
	return nil
}

func (r *Repositori) Hapus(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM sidebar_pasien WHERE id=?", id)
	if err != nil {
		return fmt.Errorf("delete sidebar: %w", err)
	}
	jumlah, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted sidebar rows: %w", err)
	}
	if jumlah == 0 {
		return ErrSidebarTidakDitemukan
	}
	return nil
}

func (r *Repositori) bacaSidebar(ctx context.Context) ([]Sidebar, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id,kode_sidebar,nama_sidebar,ikon,daftar_modul,COALESCE(permission_code,''),urutan,aktif
		FROM sidebar_pasien
		ORDER BY urutan,nama_sidebar
	`)
	if err != nil {
		return nil, fmt.Errorf("read sidebar: %w", err)
	}
	defer rows.Close()

	daftar := make([]Sidebar, 0)
	for rows.Next() {
		var item Sidebar
		var daftarModul []byte
		if err := rows.Scan(&item.ID, &item.Kode, &item.Nama, &item.Ikon, &daftarModul, &item.PermissionCode, &item.Urutan, &item.Aktif); err != nil {
			return nil, fmt.Errorf("scan sidebar: %w", err)
		}
		if err := json.Unmarshal(daftarModul, &item.DaftarModul); err != nil {
			return nil, fmt.Errorf("decode sidebar modules: %w", err)
		}
		daftar = append(daftar, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sidebar: %w", err)
	}
	return daftar, nil
}

func (r *Repositori) bacaPermission(ctx context.Context) ([]Permission, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+"`group`"+`,name,code FROM permissions
		WHERE code <> '*'
		ORDER BY `+"`group`"+`,name
	`)
	if err != nil {
		return nil, fmt.Errorf("read permissions: %w", err)
	}
	defer rows.Close()

	daftar := make([]Permission, 0)
	for rows.Next() {
		var item Permission
		if err := rows.Scan(&item.Grup, &item.Nama, &item.Kode); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		daftar = append(daftar, item)
	}
	return daftar, rows.Err()
}

func (r *Repositori) validasiPermission(ctx context.Context, kode string) error {
	if kode == "" {
		return nil
	}
	var ada bool
	if err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM permissions WHERE code=?)", kode).Scan(&ada); err != nil {
		return fmt.Errorf("check permission: %w", err)
	}
	if !ada {
		return ErrPermissionTidakValid
	}
	return nil
}

func ValidasiInput(input InputSidebar) bool {
	input = normalisasiInput(input)
	if !polaKodeSidebar.MatchString(input.Kode) || len(input.Kode) > 80 || input.Nama == "" || len(input.Nama) > 80 || input.Ikon == "" || len(input.Ikon) > 40 || len(input.DaftarModul) == 0 {
		return false
	}
	modulValid := map[string]bool{"IGD/UGD": true, "Rawat Jalan": true, "Rawat Inap": true}
	for _, modul := range input.DaftarModul {
		if !modulValid[modul] {
			return false
		}
	}
	return true
}

func normalisasiInput(input InputSidebar) InputSidebar {
	input.Kode = strings.ToLower(strings.TrimSpace(input.Kode))
	input.Nama = strings.TrimSpace(input.Nama)
	input.Ikon = strings.TrimSpace(input.Ikon)
	input.PermissionCode = strings.TrimSpace(input.PermissionCode)
	unik := make([]string, 0, len(input.DaftarModul))
	terpakai := make(map[string]bool)
	for _, modul := range input.DaftarModul {
		modul = strings.TrimSpace(modul)
		if modul != "" && !terpakai[modul] {
			terpakai[modul] = true
			unik = append(unik, modul)
		}
	}
	input.DaftarModul = unik
	return input
}

func duplikat(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
