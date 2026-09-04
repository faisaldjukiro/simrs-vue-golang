package laporan_bor_los_toi

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

type Indikator struct {
	Tahun              int     `json:"tahun"`
	Bulan              int     `json:"bulan"`
	Periode            string  `json:"periode"`
	TotalTempatTidur   int     `json:"total_tempat_tidur"`
	JumlahHari         int     `json:"jumlah_hari"`
	TotalHariPerawatan int     `json:"total_hari_perawatan"`
	PasienKeluar       int     `json:"pasien_keluar"`
	BOR                float64 `json:"bor"`
	LOS                float64 `json:"los"`
	TOI                float64 `json:"toi"`
}

type Ringkasan struct {
	JumlahPeriode int     `json:"jumlah_periode"`
	RataBOR       float64 `json:"rata_bor"`
	RataLOS       float64 `json:"rata_los"`
	RataTOI       float64 `json:"rata_toi"`
}

type Hasil struct {
	Data      []Indikator `json:"data"`
	Ringkasan Ringkasan   `json:"ringkasan"`
}

type Repositori struct {
	db *sql.DB
}

func NewRepositori(db *sql.DB) *Repositori {
	return &Repositori{db: db}
}

func (r *Repositori) Daftar(ctx context.Context, mulai, selesai time.Time) (Hasil, error) {
	bulan := daftarBulan(mulai, selesai)
	bagian := make([]string, 0, len(bulan))
	args := make([]any, 0, len(bulan))
	for _, tanggal := range bulan {
		bagian = append(bagian, "SELECT CAST(? AS DATE) AS awal_bulan")
		args = append(args, tanggal.Format("2006-01-02"))
	}

	query := `SELECT
		YEAR(b.awal_bulan),
		MONTH(b.awal_bulan),
		tt.total_tempat_tidur,
		DAY(LAST_DAY(b.awal_bulan)),
		COALESCE(SUM(GREATEST(
			DATEDIFF(
				LEAST(
					CASE
						WHEN ki.tgl_keluar IS NULL OR ki.tgl_keluar = '0000-00-00' THEN CURDATE()
						ELSE ki.tgl_keluar
					END,
					LAST_DAY(b.awal_bulan)
				),
				GREATEST(ki.tgl_masuk, b.awal_bulan)
			) + 1,
			0
		)), 0),
		COUNT(DISTINCT CASE
			WHEN ki.tgl_keluar IS NOT NULL
				AND ki.tgl_keluar <> '0000-00-00'
				AND ki.tgl_keluar BETWEEN b.awal_bulan AND LAST_DAY(b.awal_bulan)
			THEN ki.no_rawat
		END)
	FROM (` + strings.Join(bagian, " UNION ALL ") + `) b
	CROSS JOIN (
		SELECT COUNT(k.kd_kamar) AS total_tempat_tidur
		FROM kamar k
		JOIN bangsal bg ON bg.kd_bangsal = k.kd_bangsal
		WHERE k.statusdata = '1'
			AND bg.status = '1'
			AND (
				(LENGTH(k.kd_kamar) = 3 AND k.kd_kamar REGEXP '^[0-9]+$')
				OR bg.kd_bangsal IN ('17001', '15001')
			)
	) tt
	LEFT JOIN kamar_inap ki
		ON ki.tgl_masuk <= LAST_DAY(b.awal_bulan)
		AND (
			ki.tgl_keluar >= b.awal_bulan
			OR ki.tgl_keluar = '0000-00-00'
			OR ki.tgl_keluar IS NULL
		)
	GROUP BY b.awal_bulan, tt.total_tempat_tidur
	ORDER BY b.awal_bulan`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return Hasil{}, fmt.Errorf("baca indikator BOR LOS TOI: %w", err)
	}
	defer rows.Close()

	hasil := Hasil{Data: make([]Indikator, 0, len(bulan))}
	for rows.Next() {
		var item Indikator
		if err := rows.Scan(
			&item.Tahun,
			&item.Bulan,
			&item.TotalTempatTidur,
			&item.JumlahHari,
			&item.TotalHariPerawatan,
			&item.PasienKeluar,
		); err != nil {
			return Hasil{}, err
		}

		item.Periode = namaBulan(item.Bulan) + " " + fmt.Sprint(item.Tahun)
		kapasitas := item.TotalTempatTidur * item.JumlahHari
		if kapasitas > 0 {
			item.BOR = bulatkan(float64(item.TotalHariPerawatan) / float64(kapasitas) * 100)
		}
		if item.PasienKeluar > 0 {
			item.LOS = bulatkan(float64(item.TotalHariPerawatan) / float64(item.PasienKeluar))
			item.TOI = bulatkan(float64(kapasitas-item.TotalHariPerawatan) / float64(item.PasienKeluar))
		}

		hasil.Data = append(hasil.Data, item)
		hasil.Ringkasan.RataBOR += item.BOR
		hasil.Ringkasan.RataLOS += item.LOS
		hasil.Ringkasan.RataTOI += item.TOI
	}
	if err := rows.Err(); err != nil {
		return Hasil{}, err
	}

	hasil.Ringkasan.JumlahPeriode = len(hasil.Data)
	if hasil.Ringkasan.JumlahPeriode > 0 {
		jumlah := float64(hasil.Ringkasan.JumlahPeriode)
		hasil.Ringkasan.RataBOR = bulatkan(hasil.Ringkasan.RataBOR / jumlah)
		hasil.Ringkasan.RataLOS = bulatkan(hasil.Ringkasan.RataLOS / jumlah)
		hasil.Ringkasan.RataTOI = bulatkan(hasil.Ringkasan.RataTOI / jumlah)
	}
	return hasil, nil
}

func daftarBulan(mulai, selesai time.Time) []time.Time {
	hasil := make([]time.Time, 0)
	for tanggal := mulai; !tanggal.After(selesai); tanggal = tanggal.AddDate(0, 1, 0) {
		hasil = append(hasil, tanggal)
	}
	return hasil
}

func bulatkan(nilai float64) float64 {
	return math.Round(nilai*100) / 100
}

func namaBulan(bulan int) string {
	nama := [...]string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	if bulan < 1 || bulan > 12 {
		return "-"
	}
	return nama[bulan]
}
