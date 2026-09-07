package monitoring_bed

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Bed struct {
	Kamar           string `json:"kamar"`
	Kelas           string `json:"kelas"`
	Bangsal         string `json:"bangsal"`
	StatusKamar     string `json:"status_kamar"`
	NamaPasien      string `json:"nama_pasien"`
	NoRawat         string `json:"no_rawat"`
	TanggalMasuk    string `json:"tanggal_masuk"`
	LamaDigunakan   int    `json:"lama_digunakan_hari"`
	TerakhirDipakai string `json:"terakhir_dipakai"`
}

type Hasil struct {
	Data []Bed `json:"data"`
}

type Filter struct {
	KataKunci string
	Status    string
}

type Repositori struct {
	db *sql.DB
}

func NewRepositori(db *sql.DB) *Repositori {
	return &Repositori{db: db}
}

func (r *Repositori) Daftar(ctx context.Context, f Filter) (Hasil, error) {
	where := []string{"k.statusdata = '1'"}
	args := []any{}

	if status := strings.TrimSpace(f.Status); status != "" {
		where = append(where, "k.status = ?")
		args = append(args, status)
	}

	if q := strings.TrimSpace(f.KataKunci); q != "" {
		like := "%" + q + "%"
		where = append(where, "(k.kd_kamar LIKE ? OR b.nm_bangsal LIKE ? OR p.nm_pasien LIKE ?)")
		args = append(args, like, like, like)
	}

	query := `SELECT 
		k.kd_kamar,
		MAX(k.kelas),
		COALESCE(MAX(b.nm_bangsal), ''),
		MAX(k.status),
		COALESCE(MAX(p.nm_pasien), ''),
		COALESCE(MAX(ki_aktif.no_rawat), ''),
		IF(MAX(ki_aktif.tgl_masuk) IS NOT NULL, DATE_FORMAT(MAX(ki_aktif.tgl_masuk), '%Y-%m-%d'), ''),
		COALESCE(IF(MAX(k.status) = 'ISI' AND MAX(ki_aktif.tgl_masuk) IS NOT NULL, DATEDIFF(CURDATE(), MAX(ki_aktif.tgl_masuk)), 0), 0) AS lama_digunakan_hari,
		COALESCE(IF(MAX(k.status) = 'KOSONG' AND MAX(ki_terakhir.tgl_keluar) IS NOT NULL, DATE_FORMAT(MAX(ki_terakhir.tgl_keluar), '%Y-%m-%d'), ''), '') AS terakhir_dipakai
	FROM kamar k
	LEFT JOIN bangsal b ON b.kd_bangsal = k.kd_bangsal
	LEFT JOIN kamar_inap ki_aktif ON ki_aktif.kd_kamar = k.kd_kamar AND ki_aktif.stts_pulang = '-'
	LEFT JOIN reg_periksa rp ON rp.no_rawat = ki_aktif.no_rawat
	LEFT JOIN pasien p ON p.no_rkm_medis = rp.no_rkm_medis
	LEFT JOIN (
		SELECT kd_kamar, MAX(tgl_keluar) as tgl_keluar 
		FROM kamar_inap 
		WHERE stts_pulang != '-' AND tgl_keluar != '0000-00-00' 
		GROUP BY kd_kamar
	) ki_terakhir ON ki_terakhir.kd_kamar = k.kd_kamar
	WHERE ` + strings.Join(where, " AND ") + `
	GROUP BY k.kd_kamar
	ORDER BY MAX(b.nm_bangsal), k.kd_kamar`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return Hasil{}, fmt.Errorf("baca monitoring bed: %w", err)
	}
	defer rows.Close()

	data := make([]Bed, 0)
	for rows.Next() {
		var x Bed
		if err := rows.Scan(
			&x.Kamar,
			&x.Kelas,
			&x.Bangsal,
			&x.StatusKamar,
			&x.NamaPasien,
			&x.NoRawat,
			&x.TanggalMasuk,
			&x.LamaDigunakan,
			&x.TerakhirDipakai,
		); err != nil {
			return Hasil{}, err
		}
		data = append(data, x)
	}
	return Hasil{Data: data}, rows.Err()
}
