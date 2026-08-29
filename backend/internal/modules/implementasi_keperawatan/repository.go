package implementasi_keperawatan

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrCatatanTidakDitemukan = errors.New("catatan implementasi keperawatan tidak ditemukan")
	ErrTidakBerhak           = errors.New("catatan hanya dapat diubah atau dihapus oleh petugas yang membuatnya")
)

type Petugas struct {
	NIP     string `json:"nip"`
	Nama    string `json:"nama"`
	Jabatan string `json:"jabatan"`
}

type Catatan struct {
	NoRawat     string `json:"no_rawat"`
	Tanggal     string `json:"tanggal"`
	Jam         string `json:"jam"`
	Uraian      string `json:"uraian"`
	NIP         string `json:"nip"`
	NamaPetugas string `json:"nama_petugas"`
	Jabatan     string `json:"jabatan"`
	BisaDiubah  bool   `json:"bisa_diubah"`
}

type Kunci struct {
	NoRawat string `json:"no_rawat"`
	Tanggal string `json:"tanggal"`
	Jam     string `json:"jam"`
}

type Repositori struct {
	aplikasiDB *sql.DB
	simrsDB    *sql.DB
}

func NewRepositori(aplikasiDB, simrsDB *sql.DB) *Repositori {
	return &Repositori{aplikasiDB: aplikasiDB, simrsDB: simrsDB}
}

func (r *Repositori) BillingTerkunci(ctx context.Context, noRawat string) (bool, error) {
	var jumlah int
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM billing WHERE no_rawat = ?) +
			(SELECT COUNT(*) FROM reg_periksa WHERE no_rawat = ? AND stts = 'Batal')
	`, strings.TrimSpace(noRawat), strings.TrimSpace(noRawat)).Scan(&jumlah)
	if err != nil {
		return false, fmt.Errorf("periksa status billing implementasi keperawatan: %w", err)
	}
	return jumlah > 0, nil
}

func (r *Repositori) PetugasByNIP(ctx context.Context, nip string) (Petugas, error) {
	var petugas Petugas
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT p.nip, COALESCE(p.nama, ''), COALESCE(pg.jbtn, '')
		FROM petugas p
		LEFT JOIN pegawai pg ON pg.nik = p.nip
		WHERE p.nip = ?
		LIMIT 1
	`, strings.TrimSpace(nip)).Scan(&petugas.NIP, &petugas.Nama, &petugas.Jabatan)
	return petugas, err
}

func (r *Repositori) CariPetugas(ctx context.Context, kataKunci string) ([]Petugas, error) {
	kataKunci = strings.TrimSpace(kataKunci)
	if len(kataKunci) < 2 {
		return []Petugas{}, nil
	}
	seperti := "%" + kataKunci + "%"
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT p.nip, COALESCE(p.nama, ''), COALESCE(pg.jbtn, '')
		FROM petugas p
		LEFT JOIN pegawai pg ON pg.nik = p.nip
		WHERE p.nip LIKE ? OR p.nama LIKE ? OR pg.jbtn LIKE ?
		ORDER BY p.nama
		LIMIT 20
	`, seperti, seperti, seperti)
	if err != nil {
		return nil, fmt.Errorf("cari petugas implementasi keperawatan: %w", err)
	}
	defer rows.Close()

	daftar := make([]Petugas, 0)
	for rows.Next() {
		var item Petugas
		if err := rows.Scan(&item.NIP, &item.Nama, &item.Jabatan); err != nil {
			return nil, fmt.Errorf("scan petugas implementasi keperawatan: %w", err)
		}
		daftar = append(daftar, item)
	}
	return daftar, rows.Err()
}

func (r *Repositori) AksesPenuh(ctx context.Context, userID uint64) (bool, error) {
	var jumlah int
	err := r.aplikasiDB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM user_permissions
		INNER JOIN permissions ON permissions.id = user_permissions.permission_id
		WHERE user_permissions.user_id = ? AND permissions.code = '*'
	`, userID).Scan(&jumlah)
	return jumlah > 0, err
}

func (r *Repositori) Daftar(ctx context.Context, noRawat, nipLogin string, aksesPenuh bool) ([]Catatan, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT c.no_rawat, DATE_FORMAT(c.tanggal, '%Y-%m-%d'), TIME_FORMAT(c.jam, '%H:%i:%s'),
		       COALESCE(c.uraian, ''), c.nip, COALESCE(p.nama, ''), COALESCE(pg.jbtn, '')
		FROM catatan_keperawatan_ranap c
		INNER JOIN petugas p ON p.nip = c.nip
		LEFT JOIN pegawai pg ON pg.nik = c.nip
		WHERE c.no_rawat = ?
		ORDER BY c.tanggal DESC, c.jam DESC
	`, strings.TrimSpace(noRawat))
	if err != nil {
		return nil, fmt.Errorf("baca implementasi keperawatan: %w", err)
	}
	defer rows.Close()

	daftar := make([]Catatan, 0)
	for rows.Next() {
		var item Catatan
		if err := rows.Scan(&item.NoRawat, &item.Tanggal, &item.Jam, &item.Uraian, &item.NIP, &item.NamaPetugas, &item.Jabatan); err != nil {
			return nil, fmt.Errorf("scan implementasi keperawatan: %w", err)
		}
		item.BisaDiubah = aksesPenuh || item.NIP == nipLogin
		daftar = append(daftar, item)
	}
	return daftar, rows.Err()
}

func (r *Repositori) Simpan(ctx context.Context, catatan Catatan) error {
	_, err := r.simrsDB.ExecContext(ctx, `
		INSERT INTO catatan_keperawatan_ranap (tanggal, jam, no_rawat, uraian, nip)
		VALUES (?, ?, ?, ?, ?)
	`, catatan.Tanggal, catatan.Jam, catatan.NoRawat, catatan.Uraian, catatan.NIP)
	if err != nil {
		return fmt.Errorf("simpan implementasi keperawatan: %w", err)
	}
	return nil
}

func (r *Repositori) Ubah(ctx context.Context, kunci Kunci, catatan Catatan, nipLogin string, aksesPenuh bool) error {
	return r.dalamTransaksiDenganHak(ctx, kunci, nipLogin, aksesPenuh, func(tx *sql.Tx, nipLama string) error {
		nipBaru := nipLama
		if aksesPenuh && strings.TrimSpace(catatan.NIP) != "" {
			nipBaru = catatan.NIP
		}
		hasil, err := tx.ExecContext(ctx, `
			UPDATE catatan_keperawatan_ranap
			SET no_rawat = ?, tanggal = ?, jam = ?, uraian = ?, nip = ?
			WHERE no_rawat = ? AND tanggal = ? AND jam = ?
		`, catatan.NoRawat, catatan.Tanggal, catatan.Jam, catatan.Uraian, nipBaru,
			kunci.NoRawat, kunci.Tanggal, kunci.Jam)
		if err != nil {
			return fmt.Errorf("ubah implementasi keperawatan: %w", err)
		}
		jumlah, err := hasil.RowsAffected()
		if err != nil {
			return err
		}
		if jumlah == 0 {
			return ErrCatatanTidakDitemukan
		}
		return nil
	})
}

func (r *Repositori) Hapus(ctx context.Context, kunci Kunci, nipLogin string, aksesPenuh bool) error {
	return r.dalamTransaksiDenganHak(ctx, kunci, nipLogin, aksesPenuh, func(tx *sql.Tx, _ string) error {
		hasil, err := tx.ExecContext(ctx, `
			DELETE FROM catatan_keperawatan_ranap
			WHERE no_rawat = ? AND tanggal = ? AND jam = ?
		`, kunci.NoRawat, kunci.Tanggal, kunci.Jam)
		if err != nil {
			return fmt.Errorf("hapus implementasi keperawatan: %w", err)
		}
		jumlah, err := hasil.RowsAffected()
		if err != nil {
			return err
		}
		if jumlah == 0 {
			return ErrCatatanTidakDitemukan
		}
		return nil
	})
}

func (r *Repositori) dalamTransaksiDenganHak(ctx context.Context, kunci Kunci, nipLogin string, aksesPenuh bool, aksi func(*sql.Tx, string) error) error {
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var pemilik string
	err = tx.QueryRowContext(ctx, `
		SELECT nip FROM catatan_keperawatan_ranap
		WHERE no_rawat = ? AND tanggal = ? AND jam = ?
		FOR UPDATE
	`, kunci.NoRawat, kunci.Tanggal, kunci.Jam).Scan(&pemilik)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCatatanTidakDitemukan
	}
	if err != nil {
		return err
	}
	if !aksesPenuh && pemilik != nipLogin {
		return ErrTidakBerhak
	}
	if err := aksi(tx, pemilik); err != nil {
		return err
	}
	return tx.Commit()
}
