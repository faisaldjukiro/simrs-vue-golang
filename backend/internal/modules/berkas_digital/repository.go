package berkas_digital

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

var (
	ErrInputTidakValid      = errors.New("input berkas digital tidak valid")
	ErrBillingTerkunci      = errors.New("billing sudah terverifikasi atau kunjungan dibatalkan. Berkas digital hanya dapat dilihat")
	ErrMasterTidakDitemukan = errors.New("jenis berkas digital tidak ditemukan")
	ErrBerkasTidakDitemukan = errors.New("berkas digital tidak ditemukan")
)

type MasterBerkas struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

type Berkas struct {
	NoRawat    string `json:"no_rawat"`
	Kode       string `json:"kode"`
	Nama       string `json:"nama"`
	LokasiFile string `json:"lokasi_file"`
	URL        string `json:"url"`
}

type Data struct {
	Master          []MasterBerkas `json:"master"`
	Berkas          []Berkas       `json:"berkas"`
	BillingTerkunci bool           `json:"billing_terkunci"`
}

type Input struct {
	NoRawat    string `json:"no_rawat"`
	Kode       string `json:"kode"`
	LokasiFile string `json:"lokasi_file"`
}

type Kunci struct {
	NoRawat    string `json:"no_rawat"`
	Kode       string `json:"kode"`
	LokasiFile string `json:"lokasi_file"`
}

type Repositori struct {
	simrsDB    *sql.DB
	webBaseURL string
}

func NewRepositori(simrsDB *sql.DB, webBaseURL string) *Repositori {
	return &Repositori{simrsDB: simrsDB, webBaseURL: strings.TrimRight(strings.TrimSpace(webBaseURL), "/")}
}

func (r *Repositori) Data(ctx context.Context, noRawat string) (Data, error) {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return Data{}, ErrInputTidakValid
	}

	master, err := r.Master(ctx)
	if err != nil {
		return Data{}, err
	}
	berkas, err := r.Berkas(ctx, noRawat)
	if err != nil {
		return Data{}, err
	}
	terkunci, err := r.billingTerkunci(ctx, r.simrsDB, noRawat)
	if err != nil {
		return Data{}, fmt.Errorf("periksa status billing berkas digital: %w", err)
	}
	return Data{Master: master, Berkas: berkas, BillingTerkunci: terkunci}, nil
}

func (r *Repositori) Master(ctx context.Context) ([]MasterBerkas, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT kode, COALESCE(nama, '')
		FROM master_berkas_digital
		ORDER BY kode
	`)
	if err != nil {
		return nil, fmt.Errorf("baca master berkas digital: %w", err)
	}
	defer rows.Close()

	hasil := make([]MasterBerkas, 0)
	for rows.Next() {
		var item MasterBerkas
		if err := rows.Scan(&item.Kode, &item.Nama); err != nil {
			return nil, fmt.Errorf("scan master berkas digital: %w", err)
		}
		hasil = append(hasil, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate master berkas digital: %w", err)
	}
	return hasil, nil
}

func (r *Repositori) Berkas(ctx context.Context, noRawat string) ([]Berkas, error) {
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT b.no_rawat, b.kode, COALESCE(m.nama, ''), b.lokasi_file
		FROM berkas_digital_perawatan b
		LEFT JOIN master_berkas_digital m ON m.kode = b.kode
		WHERE b.no_rawat = ?
		ORDER BY m.nama, b.lokasi_file
	`, strings.TrimSpace(noRawat))
	if err != nil {
		return nil, fmt.Errorf("baca berkas digital perawatan: %w", err)
	}
	defer rows.Close()

	hasil := make([]Berkas, 0)
	for rows.Next() {
		var item Berkas
		if err := rows.Scan(&item.NoRawat, &item.Kode, &item.Nama, &item.LokasiFile); err != nil {
			return nil, fmt.Errorf("scan berkas digital perawatan: %w", err)
		}
		item.URL = r.urlBerkas(item.LokasiFile)
		hasil = append(hasil, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate berkas digital perawatan: %w", err)
	}
	return hasil, nil
}

func (r *Repositori) Simpan(ctx context.Context, input Input) error {
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mulai transaksi berkas digital: %w", err)
	}
	defer tx.Rollback()

	if terkunci, err := r.billingTerkunci(ctx, tx, input.NoRawat); err != nil {
		return err
	} else if terkunci {
		return ErrBillingTerkunci
	}

	var adaMaster bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM master_berkas_digital WHERE kode = ?)`, input.Kode).Scan(&adaMaster); err != nil {
		return fmt.Errorf("validasi master berkas digital: %w", err)
	}
	if !adaMaster {
		return ErrMasterTidakDitemukan
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO berkas_digital_perawatan (no_rawat, kode, lokasi_file)
		VALUES (?, ?, ?)
	`, input.NoRawat, input.Kode, input.LokasiFile); err != nil {
		return fmt.Errorf("simpan berkas digital perawatan: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit berkas digital perawatan: %w", err)
	}
	return nil
}

func (r *Repositori) PastikanBisaSimpan(ctx context.Context, noRawat, kode string) error {
	if terkunci, err := r.billingTerkunci(ctx, r.simrsDB, noRawat); err != nil {
		return err
	} else if terkunci {
		return ErrBillingTerkunci
	}

	var adaMaster bool
	if err := r.simrsDB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM master_berkas_digital WHERE kode = ?)`, kode).Scan(&adaMaster); err != nil {
		return fmt.Errorf("validasi master berkas digital: %w", err)
	}
	if !adaMaster {
		return ErrMasterTidakDitemukan
	}
	return nil
}

func (r *Repositori) Hapus(ctx context.Context, kunci Kunci) error {
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mulai transaksi hapus berkas digital: %w", err)
	}
	defer tx.Rollback()

	if terkunci, err := r.billingTerkunci(ctx, tx, kunci.NoRawat); err != nil {
		return err
	} else if terkunci {
		return ErrBillingTerkunci
	}

	result, err := tx.ExecContext(ctx, `
		DELETE FROM berkas_digital_perawatan
		WHERE no_rawat = ? AND kode = ? AND lokasi_file = ?
	`, kunci.NoRawat, kunci.Kode, kunci.LokasiFile)
	if err != nil {
		return fmt.Errorf("hapus berkas digital perawatan: %w", err)
	}
	jumlah, _ := result.RowsAffected()
	if jumlah == 0 {
		return ErrBerkasTidakDitemukan
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit hapus berkas digital: %w", err)
	}
	return nil
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (r *Repositori) billingTerkunci(ctx context.Context, q queryer, noRawat string) (bool, error) {
	var jumlah int
	err := q.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM billing WHERE no_rawat = ?) +
			(SELECT COUNT(*) FROM reg_periksa WHERE no_rawat = ? AND stts = 'Batal')
	`, strings.TrimSpace(noRawat), strings.TrimSpace(noRawat)).Scan(&jumlah)
	if err != nil {
		return false, fmt.Errorf("periksa status billing: %w", err)
	}
	return jumlah > 0, nil
}

func (r *Repositori) urlBerkas(lokasi string) string {
	lokasi = strings.TrimSpace(lokasi)
	if lokasi == "" || lokasi == "-" || r.webBaseURL == "" {
		return ""
	}
	lokasi = normalisasiLokasiBerkas(lokasi)
	base, err := url.Parse(r.webBaseURL)
	if err != nil {
		return ""
	}
	base.Path = path.Join(base.Path, "berkasrawat", strings.TrimLeft(lokasi, "/"))
	return base.String()
}

func normalisasiLokasiBerkas(lokasi string) string {
	lokasi = strings.TrimLeft(strings.TrimSpace(lokasi), "/")
	if strings.HasPrefix(lokasi, "pages/upload/") {
		return "" + lokasi
	}
	return lokasi
}
