package awal_medis_ranap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Dokter struct {
	Kode      string `json:"kode"`
	Nama      string `json:"nama"`
	Spesialis string `json:"spesialis"`
}

type Input struct {
	NoRawat      string `json:"no_rawat"`
	Tanggal      string `json:"tanggal"`
	KodeDokter   string `json:"kode_dokter"`
	Anamnesis    string `json:"anamnesis"`
	Hubungan     string `json:"hubungan"`
	KeluhanUtama string `json:"keluhan_utama"`
	RPS          string `json:"rps"`
	RPD          string `json:"rpd"`
	RPK          string `json:"rpk"`
	RPO          string `json:"rpo"`
	Alergi       string `json:"alergi"`
	Keadaan      string `json:"keadaan"`
	GCS          string `json:"gcs"`
	Kesadaran    string `json:"kesadaran"`
	TD           string `json:"td"`
	Nadi         string `json:"nadi"`
	RR           string `json:"rr"`
	Suhu         string `json:"suhu"`
	SPO2         string `json:"spo2"`
	BB           string `json:"bb"`
	TB           string `json:"tb"`
	Kepala       string `json:"kepala"`
	Mata         string `json:"mata"`
	Gigi         string `json:"gigi"`
	THT          string `json:"tht"`
	Thoraks      string `json:"thoraks"`
	Jantung      string `json:"jantung"`
	Paru         string `json:"paru"`
	Abdomen      string `json:"abdomen"`
	Genital      string `json:"genital"`
	Ekstremitas  string `json:"ekstremitas"`
	Kulit        string `json:"kulit"`
	KetFisik     string `json:"ket_fisik"`
	KetLokalis   string `json:"ket_lokalis"`
	Lab          string `json:"lab"`
	Rad          string `json:"rad"`
	Penunjang    string `json:"penunjang"`
	Diagnosis    string `json:"diagnosis"`
	Tata         string `json:"tata"`
	Edukasi      string `json:"edukasi"`
}

type Pilihan struct {
	Anamnesis   []string `json:"anamnesis"`
	Keadaan     []string `json:"keadaan"`
	Kesadaran   []string `json:"kesadaran"`
	Pemeriksaan []string `json:"pemeriksaan"`
}

type Data struct {
	Tersedia        bool    `json:"tersedia"`
	BillingTerkunci bool    `json:"billing_terkunci"`
	Penilaian       Input   `json:"penilaian"`
	Dokter          *Dokter `json:"dokter,omitempty"`
	Pilihan         Pilihan `json:"pilihan"`
}

type Repositori struct{ simrsDB *sql.DB }

func NewRepositori(simrsDB *sql.DB) *Repositori { return &Repositori{simrsDB: simrsDB} }

func (r *Repositori) Data(ctx context.Context, noRawat string) (Data, error) {
	data := Data{
		Penilaian: defaultInput(noRawat),
		Pilihan: Pilihan{
			Anamnesis:   []string{"Autoanamnesis", "Alloanamnesis"},
			Keadaan:     []string{"Sehat", "Sakit Ringan", "Sakit Sedang", "Sakit Berat"},
			Kesadaran:   []string{"Compos Mentis", "Apatis", "Somnolen", "Sopor", "Koma"},
			Pemeriksaan: []string{"Normal", "Abnormal", "Tidak Diperiksa"},
		},
	}
	terkunci, err := r.billingTerkunci(ctx, r.simrsDB, noRawat)
	if err != nil {
		return Data{}, fmt.Errorf("periksa status billing pasien: %w", err)
	}
	data.BillingTerkunci = terkunci

	item, dokter, err := r.ambil(ctx, noRawat)
	if err == nil {
		data.Tersedia, data.Penilaian, data.Dokter = true, item, &dokter
		return data, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Data{}, err
	}

	dokter, err = r.dokterRegistrasi(ctx, noRawat)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Data{}, err
	}
	if dokter.Kode != "" {
		data.Penilaian.KodeDokter = dokter.Kode
		data.Dokter = &dokter
	}
	return data, nil
}

func defaultInput(noRawat string) Input {
	return Input{
		NoRawat: noRawat, Tanggal: time.Now().Format("2006-01-02 15:04:05"),
		Anamnesis: "Autoanamnesis", Keadaan: "Sehat", Kesadaran: "Compos Mentis",
		Kepala: "Normal", Mata: "Normal", Gigi: "Normal", THT: "Normal", Thoraks: "Normal",
		Jantung: "Normal", Paru: "Normal", Abdomen: "Normal", Genital: "Normal", Ekstremitas: "Normal", Kulit: "Normal",
	}
}

func (r *Repositori) ambil(ctx context.Context, noRawat string) (Input, Dokter, error) {
	var item Input
	var dokter Dokter
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT p.no_rawat, DATE_FORMAT(p.tanggal,'%Y-%m-%d %H:%i:%s'), p.kd_dokter,
			p.anamnesis,p.hubungan,p.keluhan_utama,p.rps,p.rpd,p.rpk,p.rpo,p.alergi,
			p.keadaan,p.gcs,p.kesadaran,p.td,p.nadi,p.rr,p.suhu,p.spo,p.bb,p.tb,
			p.kepala,p.mata,p.gigi,p.tht,p.thoraks,p.jantung,p.paru,p.abdomen,p.ekstremitas,
			p.genital,p.kulit,p.ket_fisik,p.ket_lokalis,p.lab,p.rad,p.penunjang,p.diagnosis,p.tata,p.edukasi,
			COALESCE(d.nm_dokter,''),COALESCE(s.nm_sps,'')
		FROM penilaian_medis_ranap p
		LEFT JOIN dokter d ON d.kd_dokter=p.kd_dokter
		LEFT JOIN spesialis s ON s.kd_sps=d.kd_sps
		WHERE p.no_rawat=? LIMIT 1`, noRawat).Scan(nilaiScan(&item, &dokter)...)
	dokter.Kode = item.KodeDokter
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Input{}, Dokter{}, fmt.Errorf("baca penilaian awal medis ranap: %w", err)
	}
	return item, dokter, err
}

func (r *Repositori) dokterRegistrasi(ctx context.Context, noRawat string) (Dokter, error) {
	var item Dokter
	err := r.simrsDB.QueryRowContext(ctx, `
		SELECT d.kd_dokter,COALESCE(d.nm_dokter,''),COALESCE(s.nm_sps,'')
		FROM reg_periksa rp INNER JOIN dokter d ON d.kd_dokter=rp.kd_dokter
		LEFT JOIN spesialis s ON s.kd_sps=d.kd_sps WHERE rp.no_rawat=? LIMIT 1`, noRawat).
		Scan(&item.Kode, &item.Nama, &item.Spesialis)
	return item, err
}

func (r *Repositori) CariDokter(ctx context.Context, kata string) ([]Dokter, error) {
	seperti := "%" + strings.TrimSpace(kata) + "%"
	rows, err := r.simrsDB.QueryContext(ctx, `
		SELECT d.kd_dokter,COALESCE(d.nm_dokter,''),COALESCE(s.nm_sps,'')
		FROM dokter d LEFT JOIN spesialis s ON s.kd_sps=d.kd_sps
		WHERE d.status='1' AND (d.kd_dokter LIKE ? OR d.nm_dokter LIKE ? OR s.nm_sps LIKE ?)
		ORDER BY d.nm_dokter LIMIT 30`, seperti, seperti, seperti)
	if err != nil {
		return nil, fmt.Errorf("cari dokter: %w", err)
	}
	defer rows.Close()
	daftar := make([]Dokter, 0)
	for rows.Next() {
		var item Dokter
		if err := rows.Scan(&item.Kode, &item.Nama, &item.Spesialis); err != nil {
			return nil, err
		}
		daftar = append(daftar, item)
	}
	return daftar, rows.Err()
}

func (r *Repositori) Simpan(ctx context.Context, input Input, ubah bool) error {
	return r.dalamTransaksi(ctx, input.NoRawat, func(tx *sql.Tx) error {
		if ubah {
			hasil, err := tx.ExecContext(ctx, `UPDATE penilaian_medis_ranap SET tanggal=?,kd_dokter=?,anamnesis=?,hubungan=?,keluhan_utama=?,rps=?,rpd=?,rpk=?,rpo=?,alergi=?,keadaan=?,gcs=?,kesadaran=?,td=?,nadi=?,rr=?,suhu=?,spo=?,bb=?,tb=?,kepala=?,mata=?,gigi=?,tht=?,thoraks=?,jantung=?,paru=?,abdomen=?,ekstremitas=?,genital=?,kulit=?,ket_fisik=?,ket_lokalis=?,lab=?,rad=?,penunjang=?,diagnosis=?,tata=?,edukasi=? WHERE no_rawat=?`, nilaiSimpan(input, false)...)
			if err != nil {
				return fmt.Errorf("ubah penilaian awal medis ranap: %w", err)
			}
			jumlah, err := hasil.RowsAffected()
			if err != nil {
				return err
			}
			if jumlah == 0 {
				var ada int
				if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM penilaian_medis_ranap WHERE no_rawat=?`, input.NoRawat).Scan(&ada); err != nil {
					return err
				}
				if ada == 0 {
					return ErrTidakDitemukan
				}
			}
			return nil
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO penilaian_medis_ranap (no_rawat,tanggal,kd_dokter,anamnesis,hubungan,keluhan_utama,rps,rpd,rpk,rpo,alergi,keadaan,gcs,kesadaran,td,nadi,rr,suhu,spo,bb,tb,kepala,mata,gigi,tht,thoraks,jantung,paru,abdomen,ekstremitas,genital,kulit,ket_fisik,ket_lokalis,lab,rad,penunjang,diagnosis,tata,edukasi) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, nilaiSimpan(input, true)...)
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrSudahAda
		}
		if err != nil {
			return fmt.Errorf("simpan penilaian awal medis ranap: %w", err)
		}
		return nil
	})
}

func (r *Repositori) Hapus(ctx context.Context, noRawat string) error {
	return r.dalamTransaksi(ctx, noRawat, func(tx *sql.Tx) error {
		hasil, err := tx.ExecContext(ctx, `DELETE FROM penilaian_medis_ranap WHERE no_rawat=?`, noRawat)
		if err != nil {
			return fmt.Errorf("hapus penilaian awal medis ranap: %w", err)
		}
		jumlah, err := hasil.RowsAffected()
		if err != nil {
			return err
		}
		if jumlah == 0 {
			return ErrTidakDitemukan
		}
		return nil
	})
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (r *Repositori) billingTerkunci(ctx context.Context, q queryer, noRawat string) (bool, error) {
	var jumlah int
	err := q.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM billing WHERE no_rawat=?) + (SELECT COUNT(*) FROM reg_periksa WHERE no_rawat=? AND stts='Batal')`, noRawat, noRawat).Scan(&jumlah)
	return jumlah > 0, err
}

func (r *Repositori) dalamTransaksi(ctx context.Context, noRawat string, proses func(*sql.Tx) error) error {
	tx, err := r.simrsDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	terkunci, err := r.billingTerkunci(ctx, tx, noRawat)
	if err != nil {
		return err
	}
	if terkunci {
		return ErrBillingTerkunci
	}
	if err := proses(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func nilaiSimpan(i Input, sertakanNoRawat bool) []any {
	nilai := []any{i.Tanggal, i.KodeDokter, i.Anamnesis, i.Hubungan, i.KeluhanUtama, i.RPS, i.RPD, i.RPK, i.RPO, i.Alergi, i.Keadaan, i.GCS, i.Kesadaran, i.TD, i.Nadi, i.RR, i.Suhu, i.SPO2, i.BB, i.TB, i.Kepala, i.Mata, i.Gigi, i.THT, i.Thoraks, i.Jantung, i.Paru, i.Abdomen, i.Ekstremitas, i.Genital, i.Kulit, i.KetFisik, i.KetLokalis, i.Lab, i.Rad, i.Penunjang, i.Diagnosis, i.Tata, i.Edukasi}
	if sertakanNoRawat {
		return append([]any{i.NoRawat}, nilai...)
	}
	return append(nilai, i.NoRawat)
}

func nilaiScan(i *Input, d *Dokter) []any {
	return []any{&i.NoRawat, &i.Tanggal, &i.KodeDokter, &i.Anamnesis, &i.Hubungan, &i.KeluhanUtama, &i.RPS, &i.RPD, &i.RPK, &i.RPO, &i.Alergi, &i.Keadaan, &i.GCS, &i.Kesadaran, &i.TD, &i.Nadi, &i.RR, &i.Suhu, &i.SPO2, &i.BB, &i.TB, &i.Kepala, &i.Mata, &i.Gigi, &i.THT, &i.Thoraks, &i.Jantung, &i.Paru, &i.Abdomen, &i.Ekstremitas, &i.Genital, &i.Kulit, &i.KetFisik, &i.KetLokalis, &i.Lab, &i.Rad, &i.Penunjang, &i.Diagnosis, &i.Tata, &i.Edukasi, &d.Nama, &d.Spesialis}
}
