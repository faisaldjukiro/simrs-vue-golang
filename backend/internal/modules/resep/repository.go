package resep

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInputTidakValid = errors.New("input resep tidak valid")
	ErrResepTidakAda   = errors.New("resep tidak ditemukan")
	ErrBillingTerkunci = errors.New("billing sudah terverifikasi atau kunjungan dibatalkan")
)

type InfoPasien struct {
	NoRawat         string `json:"no_rawat"`
	NoRKM           string `json:"no_rkm_medis"`
	NamaPasien      string `json:"nm_pasien"`
	Umur            string `json:"umur"`
	KdPoli          string `json:"kd_poli"`
	StatusRawat     string `json:"status_rawat"`
	BillingTerkunci bool   `json:"billing_terkunci"`
}
type InfoDokter struct {
	KdDokter string `json:"kd_dokter"`
	NmDokter string `json:"nm_dokter"`
}
type DepoGudang struct {
	KdBangsal string `json:"kd_bangsal"`
	NmBangsal string `json:"nm_bangsal"`
}
type Obat struct {
	KodeBarang     string  `json:"kode_brng"`
	NamaBarang     string  `json:"nama_brng"`
	KodeSatuan     string  `json:"kode_sat"`
	Jenis          string  `json:"jenis"`
	Harga          float64 `json:"harga"`
	HargaBeli      float64 `json:"h_beli"`
	Stok           float64 `json:"stok"`
	LetakBarang    string  `json:"letak_barang"`
	NamaIndustri   string  `json:"nama_industri"`
	KategoriFornas string  `json:"kategori_fornas"`
}
type ItemResep struct {
	KodeBarang     string  `json:"kode_brng"`
	NamaBarang     string  `json:"nama_brng"`
	KodeSatuan     string  `json:"kode_sat"`
	Jumlah         float64 `json:"jml"`
	Harga          float64 `json:"harga"`
	HargaBeli      float64 `json:"h_beli"`
	Stok           float64 `json:"stok"`
	AturanPakai    string  `json:"aturan_pakai"`
	Jenis          string  `json:"jenis"`
	Ikhtisar       string  `json:"ikhtisar_farmasi"`
	KategoriFornas string  `json:"kategori_fornas"`
}
type RacikanDetail struct {
	KodeBarang string  `json:"kode_brng"`
	NamaBarang string  `json:"nama_brng"`
	KodeSatuan string  `json:"kode_sat"`
	Jumlah     float64 `json:"jml"`
	Harga      float64 `json:"harga"`
	Stok       float64 `json:"stok"`
	P1         float64 `json:"p1"`
	P2         float64 `json:"p2"`
	Kandungan  string  `json:"kandungan"`
}
type RacikanHeader struct {
	NoRacik     string          `json:"no_racik"`
	NamaRacik   string          `json:"nama_racik"`
	KdRacik     string          `json:"kd_racik"`
	NmMetode    string          `json:"nm_racik"`
	JumlahDr    float64         `json:"jml_dr"`
	AturanPakai string          `json:"aturan_pakai"`
	Keterangan  string          `json:"keterangan"`
	Detail      []RacikanDetail `json:"detail"`
}
type ResepHeader struct {
	NoResep      string  `json:"no_resep"`
	TglPeresepan string  `json:"tgl_peresepan"`
	Jam          string  `json:"jam"`
	NoRawat      string  `json:"no_rawat"`
	KdDokter     string  `json:"kd_dokter"`
	NmDokter     string  `json:"nm_dokter"`
	Status       string  `json:"status"`
	KdBangsal    string  `json:"kd_bangsal"`
	NmBangsal    string  `json:"nm_bangsal"`
	Judul        string  `json:"judul"`
	Total        float64 `json:"total"`
	JumlahObat   int     `json:"jumlah_obat"`
	JumlahRacik  int     `json:"jumlah_racik"`
}
type ResepLengkap struct {
	Header  ResepHeader     `json:"header"`
	Obat    []ItemResep     `json:"obat"`
	Racikan []RacikanHeader `json:"racikan"`
}
type InputSimpanResep struct {
	NoResep      string          `json:"no_resep"`
	TglPeresepan string          `json:"tgl_peresepan"`
	Jam          string          `json:"jam"`
	NoRawat      string          `json:"no_rawat"`
	KdDokter     string          `json:"kd_dokter"`
	Status       string          `json:"status"`
	KdBangsal    string          `json:"kd_bangsal"`
	Judul        string          `json:"judul"`
	Obat         []ItemResep     `json:"obat"`
	Racikan      []RacikanHeader `json:"racikan"`
	IsUbah       bool            `json:"is_ubah"`
}
type MetodeRacik struct {
	KdRacik   string  `json:"kd_racik"`
	NmRacik   string  `json:"nm_racik"`
	Kapasitas float64 `json:"kapasitas"`
}
type Repositori struct{ simrsDB *sql.DB }

func NewRepositori(simrsDB, _ *sql.DB) *Repositori { return &Repositori{simrsDB: simrsDB} }

func (r *Repositori) InfoPasien(ctx context.Context, noRawat string) (InfoPasien, error) {
	var p InfoPasien
	err := r.simrsDB.QueryRowContext(ctx, `SELECT rp.no_rawat,rp.no_rkm_medis,p.nm_pasien,p.umur,rp.kd_poli,LOWER(rp.status_lanjut),EXISTS(SELECT 1 FROM billing b WHERE b.no_rawat=rp.no_rawat) OR rp.stts='Batal' FROM reg_periksa rp JOIN pasien p ON p.no_rkm_medis=rp.no_rkm_medis WHERE rp.no_rawat=? LIMIT 1`, noRawat).Scan(&p.NoRawat, &p.NoRKM, &p.NamaPasien, &p.Umur, &p.KdPoli, &p.StatusRawat, &p.BillingTerkunci)
	if err == sql.ErrNoRows {
		return InfoPasien{}, ErrResepTidakAda
	}
	if err != nil {
		return p, fmt.Errorf("info pasien: %w", err)
	}
	return p, nil
}
func (r *Repositori) CariDokter(ctx context.Context, q string) ([]InfoDokter, error) {
	like := "%" + q + "%"
	rows, e := r.simrsDB.QueryContext(ctx, `SELECT kd_dokter,nm_dokter FROM dokter WHERE (kd_dokter LIKE ? OR nm_dokter LIKE ?) AND status='1' ORDER BY nm_dokter LIMIT 30`, like, like)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []InfoDokter{}
	for rows.Next() {
		var d InfoDokter
		if e = rows.Scan(&d.KdDokter, &d.NmDokter); e != nil {
			return nil, e
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
func (r *Repositori) DepoDefault(ctx context.Context, noRawat, status string) (string, string) {
	var kd, nm string
	if strings.EqualFold(status, "ralan") {
		var poli string
		_ = r.simrsDB.QueryRowContext(ctx, `SELECT kd_poli FROM reg_periksa WHERE no_rawat=?`, noRawat).Scan(&poli)
		_ = r.simrsDB.QueryRowContext(ctx, `SELECT kd_bangsal FROM set_depo_ralan WHERE kd_poli=? LIMIT 1`, poli).Scan(&kd)
	} else {
		var kdBangsal string
		_ = r.simrsDB.QueryRowContext(ctx, `SELECT k.kd_bangsal FROM kamar_inap ki INNER JOIN kamar k ON k.kd_kamar=ki.kd_kamar WHERE ki.no_rawat=? ORDER BY ki.tgl_masuk DESC,ki.jam_masuk DESC LIMIT 1`, noRawat).Scan(&kdBangsal)
		_ = r.simrsDB.QueryRowContext(ctx, `SELECT kd_depo FROM set_depo_ranap WHERE kd_bangsal=? LIMIT 1`, kdBangsal).Scan(&kd)
		if kd == "" {
			kd = kdBangsal
		}
	}
	if kd == "" {
		_ = r.simrsDB.QueryRowContext(ctx, `SELECT kd_bangsal FROM set_lokasi LIMIT 1`).Scan(&kd)
	}
	_ = r.simrsDB.QueryRowContext(ctx, `SELECT nm_bangsal FROM bangsal WHERE kd_bangsal=?`, kd).Scan(&nm)
	return kd, nm
}
func (r *Repositori) CariDepoGudang(ctx context.Context, q string) ([]DepoGudang, error) {
	like := "%" + q + "%"
	rows, e := r.simrsDB.QueryContext(ctx, `SELECT kd_bangsal,nm_bangsal FROM bangsal WHERE status='1' AND (kd_bangsal LIKE ? OR nm_bangsal LIKE ?) ORDER BY nm_bangsal LIMIT 40`, like, like)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []DepoGudang{}
	for rows.Next() {
		var d DepoGudang
		if e = rows.Scan(&d.KdBangsal, &d.NmBangsal); e != nil {
			return nil, e
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
func (r *Repositori) CariObat(ctx context.Context, q, kdBangsal, kelas string) ([]Obat, error) {
	col := normalisasiKolumHarga(kelas)
	like := "%" + q + "%"
	query := fmt.Sprintf(`SELECT db.kode_brng,db.nama_brng,COALESCE(db.kode_sat,''),COALESCE(j.nama,''),COALESCE(db.%s,0),COALESCE(db.h_beli,0),COALESCE(SUM(gb.stok),0),COALESCE(db.letak_barang,''),COALESCE(i.nama_industri,''),COALESCE(ofn.kategori,'') FROM databarang db LEFT JOIN jenis j ON j.kdjns=db.kdjns LEFT JOIN industrifarmasi i ON i.kode_industri=db.kode_industri LEFT JOIN gudangbarang gb ON gb.kode_brng=db.kode_brng AND gb.kd_bangsal=? LEFT JOIN obat_fornas_non_fornas ofn ON ofn.kode_brng=db.kode_brng WHERE db.status='1' AND (db.kode_brng LIKE ? OR db.nama_brng LIKE ? OR j.nama LIKE ?) GROUP BY db.kode_brng ORDER BY db.nama_brng LIMIT 100`, col)
	rows, e := r.simrsDB.QueryContext(ctx, query, kdBangsal, like, like, like)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Obat{}
	for rows.Next() {
		var o Obat
		if e = rows.Scan(&o.KodeBarang, &o.NamaBarang, &o.KodeSatuan, &o.Jenis, &o.Harga, &o.HargaBeli, &o.Stok, &o.LetakBarang, &o.NamaIndustri, &o.KategoriFornas); e != nil {
			return nil, e
		}
		out = append(out, o)
	}
	return out, rows.Err()
}
func (r *Repositori) DaftarMetodeRacik(ctx context.Context) ([]MetodeRacik, error) {
	rows, e := r.simrsDB.QueryContext(ctx, `SELECT kd_racik,nm_racik,kapasitas FROM metode_racik ORDER BY nm_racik`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []MetodeRacik{}
	for rows.Next() {
		var m MetodeRacik
		if e = rows.Scan(&m.KdRacik, &m.NmRacik, &m.Kapasitas); e != nil {
			return nil, e
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (r *Repositori) AutoNomorResep(ctx context.Context, tanggal string) (string, error) {
	t, e := time.Parse("2006-01-02", tanggal)
	if e != nil {
		return "", inputTidakValid("format tanggal tidak valid")
	}
	var n int
	_ = r.simrsDB.QueryRowContext(ctx, `SELECT COALESCE(MAX(CAST(RIGHT(no_resep,4) AS UNSIGNED)),0) FROM resep_obat WHERE tgl_peresepan=?`, tanggal).Scan(&n)
	return fmt.Sprintf("%s%04d", t.Format("20060102"), n+1), nil
}
func (r *Repositori) DaftarResep(ctx context.Context, noRawat string) ([]ResepHeader, error) {
	rows, e := r.simrsDB.QueryContext(ctx, `SELECT ro.no_resep,DATE_FORMAT(IF(ro.tgl_peresepan IS NULL OR ro.tgl_peresepan='0000-00-00',ro.tgl_perawatan,ro.tgl_peresepan),'%Y-%m-%d'),TIME_FORMAT(IF(ro.jam_peresepan IS NULL OR ro.jam_peresepan='00:00:00',ro.jam,ro.jam_peresepan),'%H:%i:%s'),ro.no_rawat,ro.kd_dokter,COALESCE(d.nm_dokter,''),COALESCE(ro.status,''),'' AS kd_bangsal,'' AS nm_bangsal,COALESCE(ro.judul,''),COALESCE((SELECT SUM(rd.jml*COALESCE(db.ralan,0)) FROM resep_dokter rd LEFT JOIN databarang db ON db.kode_brng=rd.kode_brng WHERE rd.no_resep=ro.no_resep),0),(SELECT COUNT(*) FROM resep_dokter rd WHERE rd.no_resep=ro.no_resep),(SELECT COUNT(*) FROM resep_dokter_racikan rr WHERE rr.no_resep=ro.no_resep) FROM resep_obat ro LEFT JOIN dokter d ON d.kd_dokter=ro.kd_dokter WHERE ro.no_rawat=? ORDER BY ro.tgl_peresepan DESC,ro.jam_peresepan DESC`, noRawat)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []ResepHeader{}
	for rows.Next() {
		var h ResepHeader
		if e = rows.Scan(&h.NoResep, &h.TglPeresepan, &h.Jam, &h.NoRawat, &h.KdDokter, &h.NmDokter, &h.Status, &h.KdBangsal, &h.NmBangsal, &h.Judul, &h.Total, &h.JumlahObat, &h.JumlahRacik); e != nil {
			return nil, e
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
func (r *Repositori) DaftarResepPasien(ctx context.Context, noRKM string) ([]ResepHeader, error) {
	rows, e := r.simrsDB.QueryContext(ctx, `SELECT ro.no_resep,DATE_FORMAT(IF(ro.tgl_peresepan IS NULL OR ro.tgl_peresepan='0000-00-00',ro.tgl_perawatan,ro.tgl_peresepan),'%Y-%m-%d'),TIME_FORMAT(IF(ro.jam_peresepan IS NULL OR ro.jam_peresepan='00:00:00',ro.jam,ro.jam_peresepan),'%H:%i:%s'),ro.no_rawat,ro.kd_dokter,COALESCE(d.nm_dokter,''),COALESCE(ro.status,''),'' AS kd_bangsal,'' AS nm_bangsal,COALESCE(ro.judul,''),COALESCE((SELECT SUM(rd.jml*COALESCE(db.ralan,0)) FROM resep_dokter rd LEFT JOIN databarang db ON db.kode_brng=rd.kode_brng WHERE rd.no_resep=ro.no_resep),0),(SELECT COUNT(*) FROM resep_dokter rd WHERE rd.no_resep=ro.no_resep),(SELECT COUNT(*) FROM resep_dokter_racikan rr WHERE rr.no_resep=ro.no_resep) FROM resep_obat ro INNER JOIN reg_periksa rp ON rp.no_rawat=ro.no_rawat LEFT JOIN dokter d ON d.kd_dokter=ro.kd_dokter WHERE rp.no_rkm_medis=? ORDER BY ro.tgl_peresepan DESC,ro.jam_peresepan DESC,ro.tgl_perawatan DESC,ro.jam DESC LIMIT 200`, noRKM)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []ResepHeader{}
	for rows.Next() {
		var h ResepHeader
		if e = rows.Scan(&h.NoResep, &h.TglPeresepan, &h.Jam, &h.NoRawat, &h.KdDokter, &h.NmDokter, &h.Status, &h.KdBangsal, &h.NmBangsal, &h.Judul, &h.Total, &h.JumlahObat, &h.JumlahRacik); e != nil {
			return nil, e
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
func (r *Repositori) DetailResep(ctx context.Context, no string) (ResepLengkap, error) {
	var h ResepLengkap
	e := r.simrsDB.QueryRowContext(ctx, `SELECT ro.no_resep,IF(ro.tgl_peresepan IS NULL OR ro.tgl_peresepan='0000-00-00',ro.tgl_perawatan,ro.tgl_peresepan),IF(ro.jam_peresepan IS NULL OR ro.jam_peresepan='00:00:00',ro.jam,ro.jam_peresepan),ro.no_rawat,ro.kd_dokter,COALESCE(d.nm_dokter,''),COALESCE(ro.status,''),'' AS kd_bangsal,'' AS nm_bangsal,COALESCE(ro.judul,'') FROM resep_obat ro LEFT JOIN dokter d ON d.kd_dokter=ro.kd_dokter WHERE ro.no_resep=?`, no).Scan(&h.Header.NoResep, &h.Header.TglPeresepan, &h.Header.Jam, &h.Header.NoRawat, &h.Header.KdDokter, &h.Header.NmDokter, &h.Header.Status, &h.Header.KdBangsal, &h.Header.NmBangsal, &h.Header.Judul)
	if e == sql.ErrNoRows {
		return h, ErrResepTidakAda
	}
	if e != nil {
		return h, e
	}
	h.Obat = []ItemResep{}
	rows, e := r.simrsDB.QueryContext(ctx, `SELECT rd.kode_brng,COALESCE(db.nama_brng,''),COALESCE(db.kode_sat,''),rd.jml,COALESCE(db.ralan,0),COALESCE(db.h_beli,0),COALESCE(rd.aturan_pakai,''),COALESCE(j.nama,''),'','' FROM resep_dokter rd LEFT JOIN databarang db ON db.kode_brng=rd.kode_brng LEFT JOIN jenis j ON j.kdjns=db.kdjns WHERE rd.no_resep=?`, no)
	if e != nil {
		return h, e
	}
	for rows.Next() {
		var i ItemResep
		if e = rows.Scan(&i.KodeBarang, &i.NamaBarang, &i.KodeSatuan, &i.Jumlah, &i.Harga, &i.HargaBeli, &i.AturanPakai, &i.Jenis, &i.Ikhtisar, &i.KategoriFornas); e != nil {
			rows.Close()
			return h, e
		}
		h.Obat = append(h.Obat, i)
	}
	rows.Close()
	h.Racikan = []RacikanHeader{}
	racikRows, re := r.simrsDB.QueryContext(ctx, `SELECT rr.no_racik,rr.nama_racik,rr.kd_racik,COALESCE(m.nm_racik,''),rr.jml_dr,rr.aturan_pakai,COALESCE(rr.keterangan,'') FROM resep_dokter_racikan rr LEFT JOIN metode_racik m ON m.kd_racik=rr.kd_racik WHERE rr.no_resep=? ORDER BY rr.no_racik`, no)
	if re != nil {
		return h, re
	}
	for racikRows.Next() {
		var rc RacikanHeader
		if re = racikRows.Scan(&rc.NoRacik, &rc.NamaRacik, &rc.KdRacik, &rc.NmMetode, &rc.JumlahDr, &rc.AturanPakai, &rc.Keterangan); re != nil {
			racikRows.Close()
			return h, re
		}
		detailRows, de := r.simrsDB.QueryContext(ctx, `SELECT d.kode_brng,COALESCE(b.nama_brng,''),COALESCE(b.kode_sat,''),d.jml,COALESCE(b.ralan,0),COALESCE(g.stok,0),COALESCE(d.p1,1),COALESCE(d.p2,1),COALESCE(d.kandungan,'') FROM resep_dokter_racikan_detail d LEFT JOIN databarang b ON b.kode_brng=d.kode_brng LEFT JOIN (SELECT kode_brng,SUM(stok) stok FROM gudangbarang GROUP BY kode_brng) g ON g.kode_brng=d.kode_brng WHERE d.no_resep=? AND d.no_racik=?`, no, rc.NoRacik)
		if de != nil {
			racikRows.Close()
			return h, de
		}
		for detailRows.Next() {
			var d RacikanDetail
			if de = detailRows.Scan(&d.KodeBarang, &d.NamaBarang, &d.KodeSatuan, &d.Jumlah, &d.Harga, &d.Stok, &d.P1, &d.P2, &d.Kandungan); de != nil {
				detailRows.Close()
				racikRows.Close()
				return h, de
			}
			rc.Detail = append(rc.Detail, d)
		}
		detailRows.Close()
		h.Racikan = append(h.Racikan, rc)
	}
	racikRows.Close()
	return h, nil
}
func (r *Repositori) SimpanResep(ctx context.Context, in InputSimpanResep) error {
	tx, e := r.simrsDB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var terkunci bool
	_ = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM billing WHERE no_rawat=?) OR EXISTS(SELECT 1 FROM reg_periksa WHERE no_rawat=? AND stts='Batal')`, in.NoRawat, in.NoRawat).Scan(&terkunci)
	if terkunci {
		return ErrBillingTerkunci
	}
	if in.IsUbah {
		for _, t := range []string{"resep_dokter_racikan_detail", "resep_dokter_racikan", "resep_dokter"} {
			if _, e = tx.ExecContext(ctx, "DELETE FROM "+t+" WHERE no_resep=?", in.NoResep); e != nil {
				return e
			}
		}
		_, e = tx.ExecContext(ctx, `UPDATE resep_obat SET tgl_peresepan=?,jam_peresepan=?,no_rawat=?,kd_dokter=?,status=?,judul=? WHERE no_resep=?`, in.TglPeresepan, in.Jam, in.NoRawat, in.KdDokter, in.Status, in.Judul, in.NoResep)
	} else {
		_, e = tx.ExecContext(ctx, `INSERT INTO resep_obat(no_resep,tgl_perawatan,jam,no_rawat,kd_dokter,tgl_peresepan,jam_peresepan,status,tgl_penyerahan,jam_penyerahan,judul) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, in.NoResep, "0000-00-00", "00:00:00", in.NoRawat, in.KdDokter, in.TglPeresepan, in.Jam, in.Status, "0000-00-00", "00:00:00", in.Judul)
	}
	if e != nil {
		return e
	}
	for _, i := range in.Obat {
		if i.KodeBarang == "" || i.Jumlah <= 0 {
			continue
		}
		if _, e = tx.ExecContext(ctx, `INSERT INTO resep_dokter(no_resep,kode_brng,jml,aturan_pakai) VALUES(?,?,?,?)`, in.NoResep, i.KodeBarang, i.Jumlah, i.AturanPakai); e != nil {
			return e
		}
	}
	for n, rc := range in.Racikan {
		no := fmt.Sprintf("%d", n+1)
		if _, e = tx.ExecContext(ctx, `INSERT INTO resep_dokter_racikan(no_resep,no_racik,nama_racik,kd_racik,jml_dr,aturan_pakai,keterangan) VALUES(?,?,?,?,?,?,?)`, in.NoResep, no, rc.NamaRacik, rc.KdRacik, rc.JumlahDr, rc.AturanPakai, rc.Keterangan); e != nil {
			return e
		}
		for _, d := range rc.Detail {
			if _, e = tx.ExecContext(ctx, `INSERT INTO resep_dokter_racikan_detail(no_resep,no_racik,kode_brng,jml,p1,p2,kandungan) VALUES(?,?,?,?,?,?,?)`, in.NoResep, no, d.KodeBarang, d.Jumlah, d.P1, d.P2, d.Kandungan); e != nil {
				return e
			}
		}
	}
	return tx.Commit()
}
func (r *Repositori) HapusResep(ctx context.Context, no string) error {
	tx, e := r.simrsDB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var rawat string
	if e = tx.QueryRowContext(ctx, `SELECT no_rawat FROM resep_obat WHERE no_resep=?`, no).Scan(&rawat); e == sql.ErrNoRows {
		return ErrResepTidakAda
	}
	if e != nil {
		return e
	}
	var terkunci bool
	_ = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM billing WHERE no_rawat=?) OR EXISTS(SELECT 1 FROM reg_periksa WHERE no_rawat=? AND stts='Batal')`, rawat, rawat).Scan(&terkunci)
	if terkunci {
		return ErrBillingTerkunci
	}
	for _, t := range []string{"resep_dokter_racikan_detail", "resep_dokter_racikan", "resep_dokter", "resep_obat"} {
		if _, e = tx.ExecContext(ctx, "DELETE FROM "+t+" WHERE no_resep=?", no); e != nil {
			return e
		}
	}
	return tx.Commit()
}
func normalisasiKolumHarga(k string) string {
	switch strings.ToLower(strings.ReplaceAll(strings.TrimSpace(k), " ", "")) {
	case "karyawan", "beliluar", "kelas1", "kelas2", "kelas3", "vip", "vvip":
		return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(k), " ", ""))
	default:
		return "ralan"
	}
}
func inputTidakValid(msg string) error { return fmt.Errorf("%w: %s", ErrInputTidakValid, msg) }
