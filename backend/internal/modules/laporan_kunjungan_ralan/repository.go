package laporan_kunjungan_ralan

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Filter struct {
	TanggalMulai, TanggalSelesai, Status, Poli, Dokter, Penjamin, Kabupaten, Kecamatan, Kelurahan, KataKunci, Jenis string
}

type Kunjungan struct {
	NoRawat      string `json:"no_rawat"`
	Tanggal      string `json:"tanggal"`
	Jam          string `json:"jam"`
	StatusDaftar string `json:"status_daftar"`
	NoRM         string `json:"no_rm"`
	NamaPasien   string `json:"nama_pasien"`
	JenisKelamin string `json:"jenis_kelamin"`
	Umur         string `json:"umur"`
	Alamat       string `json:"alamat"`
	KodeDiagnosa string `json:"kode_diagnosa"`
	Diagnosa     string `json:"diagnosa"`
	Dokter       string `json:"dokter"`
	Poli         string `json:"poli"`
	Penjamin     string `json:"penjamin"`
	NoSEP        string `json:"no_sep"`
}

type KunjunganBerulang struct {
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
	Total         int `json:"total"`
	Baru          int `json:"baru"`
	Lama          int `json:"lama"`
	LakiLaki      int `json:"laki_laki"`
	Perempuan     int `json:"perempuan"`
	Berulang      int `json:"berulang"`
	TidakBerulang int `json:"tidak_berulang"`
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

func (r *Repositori) CariReferensi(ctx context.Context, jenis, kataKunci string) ([]Referensi, error) {
	konfigurasi := map[string]struct{ tabel, kode, nama string }{
		"poli": {"poliklinik", "kd_poli", "nm_poli"}, "dokter": {"dokter", "kd_dokter", "nm_dokter"},
		"penjamin": {"penjab", "kd_pj", "png_jawab"}, "kabupaten": {"kabupaten", "kd_kab", "nm_kab"},
		"kecamatan": {"kecamatan", "kd_kec", "nm_kec"}, "kelurahan": {"kelurahan", "kd_kel", "nm_kel"},
	}
	konfig, ok := konfigurasi[jenis]
	if !ok {
		return []Referensi{}, nil
	}
	like := "%" + strings.TrimSpace(kataKunci) + "%"
	query := fmt.Sprintf("SELECT %s,%s FROM %s WHERE %s LIKE ? OR %s LIKE ? ORDER BY %s LIMIT 30", konfig.kode, konfig.nama, konfig.tabel, konfig.kode, konfig.nama, konfig.nama)
	rows, err := r.db.QueryContext(ctx, query, like, like)
	if err != nil {
		return nil, fmt.Errorf("cari referensi %s: %w", jenis, err)
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
	if f.Jenis == "berulang" {
		return r.berulang(ctx, f)
	}
	args := []any{f.TanggalMulai, f.TanggalSelesai}
	// DlgKunjunganRalan.tampil() tidak membatasi status_lanjut. Laporan lama
	// menarik seluruh registrasi pada tanggal dan status daftar yang dipilih.
	where := []string{"rp.tgl_registrasi BETWEEN ? AND ?"}
	if f.Jenis == "rekap" {
		where = append(where, "(rp.stts <> 'Batal' AND (rp.stts <> 'belum' OR rp.status_bayar = 'Sudah Bayar') AND rp.status_lanjut <> 'Ranap')")
	}
	filterLike := []struct{ nilai, kolom string }{{f.Status, "rp.stts_daftar"}, {f.Poli, "pol.nm_poli"}, {f.Dokter, "d.nm_dokter"}, {f.Penjamin, "pj.png_jawab"}, {f.Kabupaten, "kab.nm_kab"}, {f.Kecamatan, "kec.nm_kec"}, {f.Kelurahan, "kel.nm_kel"}}
	for _, item := range filterLike {
		if strings.TrimSpace(item.nilai) != "" {
			where = append(where, item.kolom+" LIKE ?")
			args = append(args, "%"+strings.TrimSpace(item.nilai)+"%")
		}
	}
	if q := strings.TrimSpace(f.KataKunci); q != "" {
		like := "%" + q + "%"
		where = append(where, `(rp.no_rawat LIKE ? OR rp.no_rkm_medis LIKE ? OR p.nm_pasien LIKE ? OR p.alamat LIKE ? OR pol.nm_poli LIKE ? OR d.nm_dokter LIKE ? OR EXISTS (SELECT 1 FROM diagnosa_pasien dx JOIN penyakit py ON py.kd_penyakit=dx.kd_penyakit WHERE dx.no_rawat=rp.no_rawat AND (py.kd_penyakit LIKE ? OR py.nm_penyakit LIKE ?)))`)
		for i := 0; i < 8; i++ {
			args = append(args, like)
		}
	}
	query := `SELECT rp.no_rawat,DATE_FORMAT(rp.tgl_registrasi,'%Y-%m-%d'),TIME_FORMAT(rp.jam_reg,'%H:%i:%s'),rp.stts_daftar,rp.no_rkm_medis,p.nm_pasien,p.jk,CONCAT(rp.umurdaftar,' ',rp.sttsumur),IFNULL(CONCAT(p.alamat,', ',kel.nm_kel,', ',kec.nm_kec,', ',kab.nm_kab),p.alamat),COALESCE(GROUP_CONCAT(DISTINCT dx.kd_penyakit ORDER BY dx.prioritas SEPARATOR ', '),''),COALESCE(GROUP_CONCAT(DISTINCT py.nm_penyakit ORDER BY dx.prioritas SEPARATOR ', '),''),d.nm_dokter,pol.nm_poli,pj.png_jawab,COALESCE(GROUP_CONCAT(DISTINCT sep.no_sep SEPARATOR ', '),'') FROM reg_periksa rp JOIN dokter d ON d.kd_dokter=rp.kd_dokter JOIN pasien p ON p.no_rkm_medis=rp.no_rkm_medis JOIN poliklinik pol ON pol.kd_poli=rp.kd_poli JOIN penjab pj ON pj.kd_pj=rp.kd_pj LEFT JOIN kelurahan kel ON kel.kd_kel=p.kd_kel LEFT JOIN kecamatan kec ON kec.kd_kec=p.kd_kec LEFT JOIN kabupaten kab ON kab.kd_kab=p.kd_kab LEFT JOIN diagnosa_pasien dx ON dx.no_rawat=rp.no_rawat LEFT JOIN penyakit py ON py.kd_penyakit=dx.kd_penyakit LEFT JOIN bridging_sep sep ON sep.no_rawat=rp.no_rawat WHERE ` + strings.Join(where, " AND ") + ` GROUP BY rp.no_rawat ORDER BY rp.tgl_registrasi,rp.jam_reg`

	// DlgKunjunganRalan.tampil() tanpa filter tambahan dan tampil2() (Kunjungan
	// Non Batal) tidak memakai GROUP BY. Pertahankan perilaku tersebut supaya
	// baris bridging_sep ganda dan total laporan sama persis dengan SIMRS lama.
	tanpaFilterTambahan := strings.TrimSpace(f.Poli) == "" && strings.TrimSpace(f.Dokter) == "" &&
		strings.TrimSpace(f.Penjamin) == "" && strings.TrimSpace(f.Kabupaten) == "" &&
		strings.TrimSpace(f.Kecamatan) == "" && strings.TrimSpace(f.Kelurahan) == "" &&
		strings.TrimSpace(f.KataKunci) == ""
	if f.Jenis == "rekap" || tanpaFilterTambahan {
		query = `SELECT rp.no_rawat,DATE_FORMAT(rp.tgl_registrasi,'%Y-%m-%d'),TIME_FORMAT(rp.jam_reg,'%H:%i:%s'),rp.stts_daftar,rp.no_rkm_medis,p.nm_pasien,p.jk,CONCAT(rp.umurdaftar,' ',rp.sttsumur),IFNULL(CONCAT(p.alamat,', ',kel.nm_kel,', ',kec.nm_kec,', ',kab.nm_kab),p.alamat),COALESCE((SELECT GROUP_CONCAT(py2.kd_penyakit ORDER BY dx2.prioritas SEPARATOR ', ') FROM diagnosa_pasien dx2 JOIN penyakit py2 ON py2.kd_penyakit=dx2.kd_penyakit WHERE dx2.no_rawat=rp.no_rawat),''),COALESCE((SELECT GROUP_CONCAT(py2.nm_penyakit ORDER BY dx2.prioritas SEPARATOR ', ') FROM diagnosa_pasien dx2 JOIN penyakit py2 ON py2.kd_penyakit=dx2.kd_penyakit WHERE dx2.no_rawat=rp.no_rawat),''),d.nm_dokter,pol.nm_poli,pj.png_jawab,COALESCE(sep.no_sep,'') FROM reg_periksa rp JOIN dokter d ON d.kd_dokter=rp.kd_dokter JOIN pasien p ON p.no_rkm_medis=rp.no_rkm_medis JOIN poliklinik pol ON pol.kd_poli=rp.kd_poli JOIN penjab pj ON pj.kd_pj=rp.kd_pj LEFT JOIN bridging_sep sep ON sep.no_rawat=rp.no_rawat LEFT JOIN kabupaten kab ON kab.kd_kab=p.kd_kab LEFT JOIN kecamatan kec ON kec.kd_kec=p.kd_kec LEFT JOIN kelurahan kel ON kel.kd_kel=p.kd_kel WHERE ` + strings.Join(where, " AND ") + ` ORDER BY rp.tgl_registrasi,rp.jam_reg`
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return Hasil{}, fmt.Errorf("baca laporan kunjungan ralan: %w", err)
	}
	defer rows.Close()
	data := make([]Kunjungan, 0)
	ring := Ringkasan{}
	for rows.Next() {
		var x Kunjungan
		if err := rows.Scan(&x.NoRawat, &x.Tanggal, &x.Jam, &x.StatusDaftar, &x.NoRM, &x.NamaPasien, &x.JenisKelamin, &x.Umur, &x.Alamat, &x.KodeDiagnosa, &x.Diagnosa, &x.Dokter, &x.Poli, &x.Penjamin, &x.NoSEP); err != nil {
			return Hasil{}, err
		}
		data = append(data, x)
		ring.Total++
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
	return Hasil{Data: data, Ringkasan: ring}, rows.Err()
}

func (r *Repositori) berulang(ctx context.Context, f Filter) (Hasil, error) {
	q := strings.TrimSpace(f.KataKunci)
	args := []any{f.TanggalMulai, f.TanggalSelesai}
	filterPencarian := ""
	if q != "" {
		filterPencarian = ` WHERE tabel_kunjungan.no_rm LIKE ? OR tabel_kunjungan.nama_pasien LIKE ? OR tabel_kunjungan.alamat LIKE ? OR tabel_kunjungan.kode_diagnosa LIKE ? OR tabel_kunjungan.status_kunjungan LIKE ?`
		for i := 0; i < 5; i++ {
			args = append(args, "%"+q+"%")
		}
	}
	query := `SELECT tabel_kunjungan.no_rm,tabel_kunjungan.nama_pasien,tabel_kunjungan.tanggal_lahir,tabel_kunjungan.alamat,tabel_kunjungan.jenis_kelamin,tabel_kunjungan.kode_diagnosa,tabel_kunjungan.status_kunjungan,tabel_kunjungan.jumlah
		FROM (
			SELECT p.no_rkm_medis AS no_rm,p.nm_pasien AS nama_pasien,DATE_FORMAT(p.tgl_lahir,'%Y-%m-%d') AS tanggal_lahir,p.alamat,p.jk AS jenis_kelamin,
			COALESCE(GROUP_CONCAT(DISTINCT dx.kd_penyakit SEPARATOR ', '),'') AS kode_diagnosa,
			IF(COUNT(DISTINCT rp.no_rawat)>1,'Berulang','Tidak Berulang') AS status_kunjungan,
			COUNT(DISTINCT rp.no_rawat) AS jumlah
			FROM reg_periksa rp
			JOIN pasien p ON p.no_rkm_medis=rp.no_rkm_medis
			JOIN diagnosa_pasien dx ON dx.no_rawat=rp.no_rawat
			JOIN penyakit py ON py.kd_penyakit=dx.kd_penyakit
			WHERE rp.tgl_registrasi BETWEEN ? AND ?
			GROUP BY YEAR(rp.tgl_registrasi),p.no_rkm_medis,p.nm_pasien,p.tgl_lahir,p.alamat,p.jk
		) AS tabel_kunjungan` + filterPencarian + ` ORDER BY tabel_kunjungan.nama_pasien`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return Hasil{}, fmt.Errorf("baca kunjungan berulang: %w", err)
	}
	defer rows.Close()
	data := make([]KunjunganBerulang, 0)
	ring := Ringkasan{}
	for rows.Next() {
		var x KunjunganBerulang
		if err := rows.Scan(&x.NoRM, &x.NamaPasien, &x.TanggalLahir, &x.Alamat, &x.JenisKelamin, &x.KodeDiagnosa, &x.StatusKunjungan, &x.JumlahKunjungan); err != nil {
			return Hasil{}, err
		}
		data = append(data, x)
		ring.Total++
		if x.JenisKelamin == "L" {
			ring.LakiLaki++
		} else if x.JenisKelamin == "P" {
			ring.Perempuan++
		}
		if x.StatusKunjungan == "Berulang" {
			ring.Berulang++
		} else {
			ring.TidakBerulang++
		}
	}
	return Hasil{Data: data, Ringkasan: ring}, rows.Err()
}
