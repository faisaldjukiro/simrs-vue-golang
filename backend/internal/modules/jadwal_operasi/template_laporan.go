package jadwal_operasi

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type TemplateLaporan struct {
	Kode          string `json:"kode"`
	Nama          string `json:"nama"`
	DiagnosisPre  string `json:"diagnosa_preop"`
	DiagnosisPost string `json:"diagnosa_postop"`
	Jaringan      string `json:"jaringan_dieksekusi"`
	Laporan       string `json:"laporan_operasi"`
}

func (r *Repositori) TemplateLaporan(ctx context.Context, q string) ([]TemplateLaporan, error) {
	if len(q) > 200 {
		return nil, ErrValidasi
	}
	q = "%" + strings.TrimSpace(q) + "%"
	rows, err := r.simrs.QueryContext(ctx, `SELECT t.id_template,COALESCE(d.nm_dokter,t.kd_dokter),
 COALESCE(t.diagnosa_preop,''),COALESCE(t.diagnosa_postop,''),COALESCE(t.jaringan_dieksekusi,''),t.laporan_operasi
 FROM template_laporan_operasi t LEFT JOIN dokter d ON d.kd_dokter=t.kd_dokter
 WHERE t.id_template LIKE ? OR d.nm_dokter LIKE ? OR t.diagnosa_preop LIKE ? OR t.laporan_operasi LIKE ?
 ORDER BY t.id_template DESC LIMIT 30`, q, q, q, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hasil := []TemplateLaporan{}
	for rows.Next() {
		var v TemplateLaporan
		if err = rows.Scan(&v.Kode, &v.Nama, &v.DiagnosisPre, &v.DiagnosisPost, &v.Jaringan, &v.Laporan); err != nil {
			return nil, err
		}
		v.Nama += " · " + v.DiagnosisPre
		hasil = append(hasil, v)
	}
	return hasil, rows.Err()
}

// Dipanggil dalam transaksi laporan: kegagalan template membatalkan keduanya.
func simpanTemplateLaporan(ctx context.Context, tx *sql.Tx, in InputPendukung, username string) error {
	var dokterLogin bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM dokter WHERE kd_dokter=?)", username).Scan(&dokterLogin); err != nil {
		return err
	}
	dokter := in.Jadwal.KdDokter
	if dokterLogin {
		dokter = username
	}
	prefix := strings.ReplaceAll(in.Jadwal.Tanggal, "-", "")
	var nomor int
	// Primary key menolak ID bersamaan; error mengembalikan seluruh transaksi.
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(CAST(RIGHT(id_template,4) AS UNSIGNED)),0)+1
 FROM template_laporan_operasi WHERE id_template LIKE ?`, prefix+"%").Scan(&nomor); err != nil {
		return err
	}
	if nomor > 9999 {
		return fmt.Errorf("%w: nomor template harian sudah penuh", ErrKonflik)
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO template_laporan_operasi
 (id_template,kd_dokter,diagnosa_preop,diagnosa_postop,jaringan_dieksekusi,laporan_operasi)
 VALUES (?,?,?,?,?,?)`, fmt.Sprintf("%s%04d", prefix, nomor), dokter, in.Data["diagnosa_preop"], in.Data["diagnosa_postop"], in.Data["jaringan_dieksekusi"], in.Data["laporan_operasi"])
	return err
}
