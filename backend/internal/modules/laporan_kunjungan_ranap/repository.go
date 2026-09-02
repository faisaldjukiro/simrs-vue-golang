package laporan_kunjungan_ranap

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Filter struct{ TanggalMulai, TanggalSelesai, Jenis, Status, Bangsal, Dokter, Penjamin, Kabupaten, Kecamatan, Kelurahan, KataKunci string }
type Kunjungan struct {
	NoRawat       string `json:"no_rawat"`
	TanggalMasuk  string `json:"tanggal_masuk"`
	TanggalKeluar string `json:"tanggal_keluar"`
	StatusDaftar  string `json:"status_daftar"`
	NoRM          string `json:"no_rm"`
	NamaPasien    string `json:"nama_pasien"`
	JenisKelamin  string `json:"jenis_kelamin"`
	Umur          string `json:"umur"`
	Alamat        string `json:"alamat"`
	KodeDiagnosa  string `json:"kode_diagnosa"`
	Diagnosa      string `json:"diagnosa"`
	Ruang         string `json:"ruang"`
	StatusPulang  string `json:"status_pulang"`
	DPJP          string `json:"dpjp"`
	LamaRawat     int    `json:"lama_rawat"`
	Kelas         string `json:"kelas"`
	Penjamin      string `json:"penjamin"`
	NoSEP         string `json:"no_sep"`
}
type Berulang struct {
	NoRM            string `json:"no_rm"`
	NamaPasien      string `json:"nama_pasien"`
	TanggalLahir    string `json:"tanggal_lahir"`
	Alamat          string `json:"alamat"`
	JenisKelamin    string `json:"jenis_kelamin"`
	KodeDiagnosa    string `json:"kode_diagnosa"`
	StatusKunjungan string `json:"status_kunjungan"`
	JumlahKunjungan int    `json:"jumlah_kunjungan"`
}
type Ringkasan struct {
	Total          int `json:"total"`
	Baru           int `json:"baru"`
	Lama           int `json:"lama"`
	LakiLaki       int `json:"laki_laki"`
	Perempuan      int `json:"perempuan"`
	Berulang       int `json:"berulang"`
	TidakBerulang  int `json:"tidak_berulang"`
	TotalLamaRawat int `json:"total_lama_rawat"`
}
type Hasil struct {
	Data      any       `json:"data"`
	Ringkasan Ringkasan `json:"ringkasan"`
}
type Referensi struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}
type Repositori struct{ db *sql.DB }

func NewRepositori(db *sql.DB) *Repositori { return &Repositori{db: db} }
func (r *Repositori) CariReferensi(ctx context.Context, jenis, q string) ([]Referensi, error) {
	cfg := map[string][3]string{"bangsal": {"bangsal", "kd_bangsal", "nm_bangsal"}, "dokter": {"dokter", "kd_dokter", "nm_dokter"}, "penjamin": {"penjab", "kd_pj", "png_jawab"}, "kabupaten": {"kabupaten", "kd_kab", "nm_kab"}, "kecamatan": {"kecamatan", "kd_kec", "nm_kec"}, "kelurahan": {"kelurahan", "kd_kel", "nm_kel"}}
	c, ok := cfg[jenis]
	if !ok {
		return []Referensi{}, nil
	}
	like := "%" + strings.TrimSpace(q) + "%"
	query := fmt.Sprintf("SELECT %s,%s FROM %s WHERE %s LIKE ? OR %s LIKE ? ORDER BY %s LIMIT 30", c[1], c[2], c[0], c[1], c[2], c[2])
	rows, e := r.db.QueryContext(ctx, query, like, like)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Referensi{}
	for rows.Next() {
		var x Referensi
		if e = rows.Scan(&x.Kode, &x.Nama); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repositori) Daftar(ctx context.Context, f Filter) (Hasil, error) {
	if f.Jenis == "berulang" {
		return r.berulang(ctx, f)
	}
	tanggal := "ki.tgl_masuk"
	if f.Jenis == "pulang" {
		tanggal = "ki.tgl_keluar"
	}
	where := []string{"rp.status_lanjut='Ranap'", "rp.stts<>'Batal'", "ki.stts_pulang<>'Pindah Kamar'", tanggal + " BETWEEN ? AND ?"}
	args := []any{f.TanggalMulai, f.TanggalSelesai}
	filters := [][2]string{{f.Status, "rp.stts_daftar"}, {f.Bangsal, "b.nm_bangsal"}, {f.Penjamin, "pj.png_jawab"}, {f.Kabupaten, "kab.nm_kab"}, {f.Kecamatan, "kec.nm_kec"}, {f.Kelurahan, "kel.nm_kel"}}
	for _, x := range filters {
		if strings.TrimSpace(x[0]) != "" {
			where = append(where, x[1]+" LIKE ?")
			args = append(args, "%"+strings.TrimSpace(x[0])+"%")
		}
	}
	if d := strings.TrimSpace(f.Dokter); d != "" {
		where = append(where, "(d.nm_dokter LIKE ? OR EXISTS(SELECT 1 FROM dpjp_ranap dr JOIN dokter dd ON dd.kd_dokter=dr.kd_dokter WHERE dr.no_rawat=rp.no_rawat AND dd.nm_dokter LIKE ?))")
		args = append(args, "%"+d+"%", "%"+d+"%")
	}
	if q := strings.TrimSpace(f.KataKunci); q != "" {
		like := "%" + q + "%"
		where = append(where, "(rp.no_rawat LIKE ? OR rp.no_rkm_medis LIKE ? OR p.nm_pasien LIKE ? OR p.alamat LIKE ? OR ki.kd_kamar LIKE ? OR b.nm_bangsal LIKE ? OR dx.kd_penyakit LIKE ? OR py.nm_penyakit LIKE ?)")
		for i := 0; i < 8; i++ {
			args = append(args, like)
		}
	}
	query := `SELECT rp.no_rawat,DATE_FORMAT(ki.tgl_masuk,'%Y-%m-%d'),DATE_FORMAT(ki.tgl_keluar,'%Y-%m-%d'),rp.stts_daftar,rp.no_rkm_medis,p.nm_pasien,p.jk,CONCAT(rp.umurdaftar,' ',rp.sttsumur),CONCAT_WS(', ',NULLIF(p.alamat,''),NULLIF(kel.nm_kel,''),NULLIF(kec.nm_kec,''),NULLIF(kab.nm_kab,'')),COALESCE(GROUP_CONCAT(DISTINCT dx.kd_penyakit ORDER BY dx.prioritas SEPARATOR ', '),''),COALESCE(GROUP_CONCAT(DISTINCT py.nm_penyakit ORDER BY dx.prioritas SEPARATOR ', '),''),CONCAT(ki.kd_kamar,' ',b.nm_bangsal),ki.stts_pulang,COALESCE(GROUP_CONCAT(DISTINCT dp.nm_dokter SEPARATOR ', '),d.nm_dokter),GREATEST(DATEDIFF(IF(ki.tgl_keluar='0000-00-00' OR ki.tgl_keluar IS NULL,CURDATE(),ki.tgl_keluar),ki.tgl_masuk),0),k.kelas,pj.png_jawab,COALESCE(GROUP_CONCAT(DISTINCT sep.no_sep SEPARATOR ', '),'') FROM reg_periksa rp JOIN pasien p ON p.no_rkm_medis=rp.no_rkm_medis JOIN kamar_inap ki ON ki.no_rawat=rp.no_rawat JOIN kamar k ON k.kd_kamar=ki.kd_kamar JOIN bangsal b ON b.kd_bangsal=k.kd_bangsal JOIN dokter d ON d.kd_dokter=rp.kd_dokter JOIN penjab pj ON pj.kd_pj=rp.kd_pj LEFT JOIN dpjp_ranap dr ON dr.no_rawat=rp.no_rawat LEFT JOIN dokter dp ON dp.kd_dokter=dr.kd_dokter LEFT JOIN kabupaten kab ON kab.kd_kab=p.kd_kab LEFT JOIN kecamatan kec ON kec.kd_kec=p.kd_kec LEFT JOIN kelurahan kel ON kel.kd_kel=p.kd_kel LEFT JOIN diagnosa_pasien dx ON dx.no_rawat=rp.no_rawat LEFT JOIN penyakit py ON py.kd_penyakit=dx.kd_penyakit LEFT JOIN bridging_sep sep ON sep.no_rawat=rp.no_rawat WHERE ` + strings.Join(where, " AND ") + ` GROUP BY rp.no_rawat ORDER BY ` + tanggal + `,rp.no_rawat`
	rows, e := r.db.QueryContext(ctx, query, args...)
	if e != nil {
		return Hasil{}, fmt.Errorf("baca laporan ranap: %w", e)
	}
	defer rows.Close()
	data := []Kunjungan{}
	ring := Ringkasan{}
	for rows.Next() {
		var x Kunjungan
		if e = rows.Scan(&x.NoRawat, &x.TanggalMasuk, &x.TanggalKeluar, &x.StatusDaftar, &x.NoRM, &x.NamaPasien, &x.JenisKelamin, &x.Umur, &x.Alamat, &x.KodeDiagnosa, &x.Diagnosa, &x.Ruang, &x.StatusPulang, &x.DPJP, &x.LamaRawat, &x.Kelas, &x.Penjamin, &x.NoSEP); e != nil {
			return Hasil{}, e
		}
		data = append(data, x)
		ring.Total++
		ring.TotalLamaRawat += x.LamaRawat
		if x.StatusDaftar == "Baru" {
			ring.Baru++
		} else if x.StatusDaftar == "Lama" {
			ring.Lama++
		}
		if x.JenisKelamin == "L" {
			ring.LakiLaki++
		} else if x.JenisKelamin == "P" {
			ring.Perempuan++
		}
	}
	return Hasil{data, ring}, rows.Err()
}
func (r *Repositori) berulang(ctx context.Context, f Filter) (Hasil, error) {
	args := []any{f.TanggalMulai, f.TanggalSelesai}
	having := ""
	if q := strings.TrimSpace(f.KataKunci); q != "" {
		having = " HAVING p.no_rkm_medis LIKE ? OR p.nm_pasien LIKE ? OR p.alamat LIKE ? OR kode_diagnosa LIKE ? OR status_kunjungan LIKE ?"
		for i := 0; i < 5; i++ {
			args = append(args, "%"+q+"%")
		}
	}
	rows, e := r.db.QueryContext(ctx, `SELECT p.no_rkm_medis,p.nm_pasien,DATE_FORMAT(p.tgl_lahir,'%Y-%m-%d'),p.alamat,p.jk,COALESCE(GROUP_CONCAT(DISTINCT dx.kd_penyakit SEPARATOR ', '),''),IF(COUNT(DISTINCT rp.no_rawat)>1,'Berulang','Tidak Berulang') status_kunjungan,COUNT(DISTINCT rp.no_rawat) FROM reg_periksa rp JOIN pasien p ON p.no_rkm_medis=rp.no_rkm_medis JOIN kamar_inap ki ON ki.no_rawat=rp.no_rawat LEFT JOIN diagnosa_pasien dx ON dx.no_rawat=rp.no_rawat WHERE rp.status_lanjut='Ranap' AND ki.stts_pulang<>'Pindah Kamar' AND ki.tgl_masuk BETWEEN ? AND ? GROUP BY YEAR(ki.tgl_masuk),p.no_rkm_medis,p.nm_pasien,p.tgl_lahir,p.alamat,p.jk`+having+` ORDER BY p.nm_pasien`, args...)
	if e != nil {
		return Hasil{}, e
	}
	defer rows.Close()
	data := []Berulang{}
	ring := Ringkasan{}
	for rows.Next() {
		var x Berulang
		if e = rows.Scan(&x.NoRM, &x.NamaPasien, &x.TanggalLahir, &x.Alamat, &x.JenisKelamin, &x.KodeDiagnosa, &x.StatusKunjungan, &x.JumlahKunjungan); e != nil {
			return Hasil{}, e
		}
		data = append(data, x)
		ring.Total++
		if x.StatusKunjungan == "Berulang" {
			ring.Berulang++
		} else {
			ring.TidakBerulang++
		}
		if x.JenisKelamin == "L" {
			ring.LakiLaki++
		} else if x.JenisKelamin == "P" {
			ring.Perempuan++
		}
	}
	return Hasil{data, ring}, rows.Err()
}
