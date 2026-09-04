package laporan_penggunaan_bed

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Filter struct{ TanggalMulai, TanggalSelesai, Bangsal string }
type Penggunaan struct {
	KodeBangsal  string  `json:"kode_bangsal"`
	NamaBangsal  string  `json:"nama_bangsal"`
	TotalBed     int     `json:"total_bed"`
	PasienKeluar int     `json:"pasien_keluar"`
	Frekuensi    float64 `json:"frekuensi"`
}
type Ringkasan struct {
	JumlahBangsal     int     `json:"jumlah_bangsal"`
	TotalBed          int     `json:"total_bed"`
	TotalPasienKeluar int     `json:"total_pasien_keluar"`
	RataRataFrekuensi float64 `json:"rata_rata_frekuensi"`
}
type Hasil struct {
	Data      []Penggunaan `json:"data"`
	Ringkasan Ringkasan    `json:"ringkasan"`
}
type Referensi struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}
type Repositori struct{ db *sql.DB }

func NewRepositori(db *sql.DB) *Repositori { return &Repositori{db: db} }

func (r *Repositori) CariBangsal(ctx context.Context, q string) ([]Referensi, error) {
	like := "%" + strings.TrimSpace(q) + "%"
	rows, err := r.db.QueryContext(ctx, `SELECT kd_bangsal,nm_bangsal FROM bangsal WHERE kd_bangsal LIKE ? OR nm_bangsal LIKE ? ORDER BY nm_bangsal LIMIT 30`, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hasil := make([]Referensi, 0)
	for rows.Next() {
		var item Referensi
		if err := rows.Scan(&item.Kode, &item.Nama); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) Daftar(ctx context.Context, f Filter) (Hasil, error) {
	query := `SELECT
		b.kd_bangsal,
		b.nm_bangsal,
		COUNT(DISTINCT k.kd_kamar),
		COUNT(ki.no_rawat),
		ROUND(
			COUNT(ki.no_rawat) /
			NULLIF(COUNT(DISTINCT k.kd_kamar), 0),
			2
		)
	FROM kamar_inap ki
	JOIN kamar k
		ON ki.kd_kamar = k.kd_kamar
	JOIN bangsal b
		ON k.kd_bangsal = b.kd_bangsal
	WHERE ki.tgl_keluar BETWEEN ? AND ?
		AND ki.stts_pulang <> '-'`
	args := []any{f.TanggalMulai, f.TanggalSelesai}
	if nama := strings.TrimSpace(f.Bangsal); nama != "" {
		query += ` AND (b.kd_bangsal LIKE ? OR b.nm_bangsal LIKE ?)`
		args = append(args, "%"+nama+"%")
		args = append(args, "%"+nama+"%")
	}
	query += `
	GROUP BY b.kd_bangsal, b.nm_bangsal
	ORDER BY b.nm_bangsal`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return Hasil{}, fmt.Errorf("baca penggunaan bed: %w", err)
	}
	defer rows.Close()
	hasil := Hasil{Data: make([]Penggunaan, 0)}
	for rows.Next() {
		var item Penggunaan
		if err := rows.Scan(
			&item.KodeBangsal,
			&item.NamaBangsal,
			&item.TotalBed,
			&item.PasienKeluar,
			&item.Frekuensi,
		); err != nil {
			return Hasil{}, err
		}
		hasil.Data = append(hasil.Data, item)
		hasil.Ringkasan.JumlahBangsal++
		hasil.Ringkasan.TotalBed += item.TotalBed
		hasil.Ringkasan.TotalPasienKeluar += item.PasienKeluar
	}
	if hasil.Ringkasan.TotalBed > 0 {
		hasil.Ringkasan.RataRataFrekuensi = float64(hasil.Ringkasan.TotalPasienKeluar) / float64(hasil.Ringkasan.TotalBed)
	}
	return hasil, rows.Err()
}
