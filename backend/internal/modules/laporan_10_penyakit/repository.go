package laporan_10_penyakit

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Filter struct {
	TanggalMulai   string
	TanggalSelesai string
	Status         string
	KataKunci      string
}

type Penyakit struct {
	Kode               string `json:"kode"`
	Nama               string `json:"nama"`
	DiagnosaLain       int    `json:"diagnosa_lain"`
	LakiHidup          int    `json:"laki_hidup"`
	PerempuanHidup     int    `json:"perempuan_hidup"`
	LakiMeninggal      int    `json:"laki_meninggal"`
	PerempuanMeninggal int    `json:"perempuan_meninggal"`
	Jumlah             int    `json:"jumlah"`
}

type Ringkasan struct {
	JumlahPenyakit int `json:"jumlah_penyakit"`
	JumlahDiagnosa int `json:"jumlah_diagnosa"`
	LakiLaki       int `json:"laki_laki"`
	Perempuan      int `json:"perempuan"`
	Meninggal      int `json:"meninggal"`
}

type Hasil struct {
	Data      []Penyakit `json:"data"`
	Ringkasan Ringkasan  `json:"ringkasan"`
}

type Repositori struct {
	db *sql.DB
}

func NewRepositori(db *sql.DB) *Repositori {
	return &Repositori{db: db}
}

func (r *Repositori) Daftar(ctx context.Context, filter Filter) (Hasil, error) {
	where := []string{"rp.tgl_registrasi BETWEEN ? AND ?"}
	args := []any{filter.TanggalMulai, filter.TanggalSelesai}

	switch filter.Status {
	case "Ralan":
		where = append(where,
			"rp.status_lanjut = 'Ralan'",
			"dp.status = 'Ralan'",
			"py.nm_penyakit NOT LIKE '%Follow-up examination%'",
			"py.kd_penyakit NOT LIKE '%Z%'",
		)
	case "Ranap":
		where = append(where,
			"rp.status_lanjut = 'Ranap'",
			"dp.status = 'Ranap'",
		)
	}

	if kataKunci := strings.TrimSpace(filter.KataKunci); kataKunci != "" {
		where = append(where, "(py.kd_penyakit LIKE ? OR py.nm_penyakit LIKE ?)")
		like := "%" + kataKunci + "%"
		args = append(args, like, like)
	}

	query := `
		SELECT
			py.kd_penyakit,
			SUBSTRING(py.nm_penyakit, 1, 80),
			COUNT(py.kd_penyakit) AS jumlah_diagnosa,
			SUM(CASE WHEN p.jk = 'L' AND COALESCE(rp.stts, '') <> 'Meninggal' THEN 1 ELSE 0 END),
			SUM(CASE WHEN p.jk = 'P' AND COALESCE(rp.stts, '') <> 'Meninggal' THEN 1 ELSE 0 END),
			SUM(CASE WHEN p.jk = 'L' AND rp.stts = 'Meninggal' THEN 1 ELSE 0 END),
			SUM(CASE WHEN p.jk = 'P' AND rp.stts = 'Meninggal' THEN 1 ELSE 0 END)
		FROM penyakit py
		JOIN diagnosa_pasien dp ON dp.kd_penyakit = py.kd_penyakit
		JOIN reg_periksa rp ON rp.no_rawat = dp.no_rawat
		JOIN dokter d ON d.kd_dokter = rp.kd_dokter
		JOIN pasien p ON p.no_rkm_medis = rp.no_rkm_medis
		JOIN poliklinik pol ON pol.kd_poli = rp.kd_poli
		JOIN penjab pj ON pj.kd_pj = rp.kd_pj
		LEFT JOIN kabupaten kab ON kab.kd_kab = p.kd_kab
		LEFT JOIN kecamatan kec ON kec.kd_kec = p.kd_kec
		LEFT JOIN kelurahan kel ON kel.kd_kel = p.kd_kel
		WHERE ` + strings.Join(where, " AND ") + `
		GROUP BY py.kd_penyakit
		ORDER BY jumlah_diagnosa DESC
		LIMIT 10`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return Hasil{}, fmt.Errorf("baca laporan 10 penyakit: %w", err)
	}
	defer rows.Close()

	hasil := Hasil{Data: make([]Penyakit, 0)}
	for rows.Next() {
		var item Penyakit
		if err := rows.Scan(
			&item.Kode,
			&item.Nama,
			&item.DiagnosaLain,
			&item.LakiHidup,
			&item.PerempuanHidup,
			&item.LakiMeninggal,
			&item.PerempuanMeninggal,
		); err != nil {
			return Hasil{}, err
		}
		item.Jumlah = item.DiagnosaLain

		hasil.Data = append(hasil.Data, item)
		hasil.Ringkasan.JumlahPenyakit++
		hasil.Ringkasan.JumlahDiagnosa += item.Jumlah
		hasil.Ringkasan.LakiLaki += item.LakiHidup + item.LakiMeninggal
		hasil.Ringkasan.Perempuan += item.PerempuanHidup + item.PerempuanMeninggal
		hasil.Ringkasan.Meninggal += item.LakiMeninggal + item.PerempuanMeninggal
	}

	return hasil, rows.Err()
}
