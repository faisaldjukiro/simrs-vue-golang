package triase_igd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrTidakDitemukan = errors.New("data triase IGD tidak ditemukan")
	ErrSudahAda       = errors.New("data triase IGD sudah ada")
)

type Petugas struct {
	NIP     string `json:"nip"`
	Nama    string `json:"nama"`
	Jabatan string `json:"jabatan"`
}

type MacamKasus struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

type Pemeriksaan struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

type KriteriaSkala struct {
	Skala           int    `json:"skala"`
	KodePemeriksaan string `json:"kode_pemeriksaan"`
	Kode            string `json:"kode"`
	Pengkajian      string `json:"pengkajian"`
}

type BagianTriase struct {
	IsiUtama         string          `json:"isi_utama"`
	KebutuhanKhusus  string          `json:"kebutuhan_khusus,omitempty"`
	Catatan          string          `json:"catatan"`
	Plan             string          `json:"plan"`
	TanggalTriase    string          `json:"tanggal_triase"`
	NIP              string          `json:"nip"`
	NamaPetugas      string          `json:"nama_petugas"`
	JabatanPetugas   string          `json:"jabatan_petugas"`
	Skala            int             `json:"skala"`
	KriteriaTerpilih []KriteriaSkala `json:"kriteria_terpilih"`
}

type Triase struct {
	NoRawat              string        `json:"no_rawat"`
	TanggalKunjungan     string        `json:"tanggal_kunjungan"`
	CaraMasuk            string        `json:"cara_masuk"`
	AlatTransportasi     string        `json:"alat_transportasi"`
	AlasanKedatangan     string        `json:"alasan_kedatangan"`
	KeteranganKedatangan string        `json:"keterangan_kedatangan"`
	KodeKasus            string        `json:"kode_kasus"`
	NamaKasus            string        `json:"nama_kasus"`
	TekananDarah         string        `json:"tekanan_darah"`
	Nadi                 string        `json:"nadi"`
	Pernapasan           string        `json:"pernapasan"`
	Suhu                 string        `json:"suhu"`
	SaturasiO2           string        `json:"saturasi_o2"`
	Nyeri                string        `json:"nyeri"`
	HandOver             string        `json:"hand_over"`
	Primer               *BagianTriase `json:"primer"`
	Sekunder             *BagianTriase `json:"sekunder"`
}

type Data struct {
	Petugas           Petugas         `json:"petugas"`
	MacamKasus        []MacamKasus    `json:"macam_kasus"`
	Pemeriksaan       []Pemeriksaan   `json:"pemeriksaan"`
	KriteriaSkala     []KriteriaSkala `json:"kriteria_skala"`
	MendukungHandOver bool            `json:"mendukung_hand_over"`
	Triase            *Triase         `json:"triase"`
	BillingTerkunci   bool            `json:"billing_terkunci"`
}

type Repositori struct{ db *sql.DB }

func NewRepositori(db *sql.DB) *Repositori { return &Repositori{db: db} }

func (r *Repositori) Data(ctx context.Context, noRawat, username string) (Data, error) {
	data := Data{
		MacamKasus:    []MacamKasus{},
		Pemeriksaan:   []Pemeriksaan{},
		KriteriaSkala: []KriteriaSkala{},
	}
	var err error
	data.Petugas, err = r.petugasLogin(ctx, username)
	if err != nil {
		return Data{}, err
	}
	if data.MacamKasus, err = r.macamKasus(ctx); err != nil {
		return Data{}, err
	}
	if data.Pemeriksaan, err = r.pemeriksaan(ctx); err != nil {
		return Data{}, err
	}
	if data.KriteriaSkala, err = r.kriteria(ctx); err != nil {
		return Data{}, err
	}
	data.MendukungHandOver, err = r.mendukungHandOver(ctx)
	if err != nil {
		return Data{}, err
	}
	data.Triase, err = r.triase(ctx, noRawat, data.MendukungHandOver)
	if err != nil {
		return Data{}, err
	}
	data.BillingTerkunci, err = r.BillingTerkunci(ctx, noRawat)
	if err != nil {
		return Data{}, err
	}
	return data, nil
}

func (r *Repositori) BillingTerkunci(ctx context.Context, noRawat string) (bool, error) {
	return billingTerkunci(ctx, r.db, noRawat)
}

type pembacaBaris interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func billingTerkunci(ctx context.Context, db pembacaBaris, noRawat string) (bool, error) {
	var jumlah int
	err := db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM billing WHERE no_rawat = ?) +
			(SELECT COUNT(*) FROM reg_periksa WHERE no_rawat = ? AND stts = 'Batal')
	`, strings.TrimSpace(noRawat), strings.TrimSpace(noRawat)).Scan(&jumlah)
	if err != nil {
		return false, fmt.Errorf("periksa status billing triase IGD: %w", err)
	}
	return jumlah > 0, nil
}

func (r *Repositori) petugasLogin(ctx context.Context, username string) (Petugas, error) {
	var item Petugas
	err := r.db.QueryRowContext(ctx, `
		SELECT nik, COALESCE(nama, ''), COALESCE(jbtn, '')
		FROM pegawai WHERE nik = ? LIMIT 1
	`, strings.TrimSpace(username)).Scan(&item.NIP, &item.Nama, &item.Jabatan)
	if errors.Is(err, sql.ErrNoRows) {
		return Petugas{}, nil
	}
	if err != nil {
		return Petugas{}, fmt.Errorf("baca petugas login triase: %w", err)
	}
	return item, nil
}

func (r *Repositori) macamKasus(ctx context.Context) ([]MacamKasus, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT kode_kasus, macam_kasus FROM master_triase_macam_kasus ORDER BY kode_kasus`)
	if err != nil {
		return nil, fmt.Errorf("baca macam kasus triase: %w", err)
	}
	defer rows.Close()
	hasil := make([]MacamKasus, 0)
	for rows.Next() {
		var item MacamKasus
		if err := rows.Scan(&item.Kode, &item.Nama); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) pemeriksaan(ctx context.Context) ([]Pemeriksaan, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT kode_pemeriksaan, nama_pemeriksaan FROM master_triase_pemeriksaan ORDER BY kode_pemeriksaan`)
	if err != nil {
		return nil, fmt.Errorf("baca pemeriksaan triase: %w", err)
	}
	defer rows.Close()
	hasil := make([]Pemeriksaan, 0)
	for rows.Next() {
		var item Pemeriksaan
		if err := rows.Scan(&item.Kode, &item.Nama); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) kriteria(ctx context.Context) ([]KriteriaSkala, error) {
	hasil := make([]KriteriaSkala, 0)
	for skala := 1; skala <= 5; skala++ {
		query := fmt.Sprintf(`SELECT kode_pemeriksaan, kode_skala%d, pengkajian_skala%d FROM master_triase_skala%d ORDER BY kode_pemeriksaan, kode_skala%d`, skala, skala, skala, skala)
		rows, err := r.db.QueryContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("baca master triase skala %d: %w", skala, err)
		}
		for rows.Next() {
			item := KriteriaSkala{Skala: skala}
			if err := rows.Scan(&item.KodePemeriksaan, &item.Kode, &item.Pengkajian); err != nil {
				rows.Close()
				return nil, err
			}
			hasil = append(hasil, item)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	return hasil, nil
}

func (r *Repositori) mendukungHandOver(ctx context.Context) (bool, error) {
	var jumlah int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'data_triase_igd' AND COLUMN_NAME = 'hand_over'
	`).Scan(&jumlah)
	if err != nil {
		return false, fmt.Errorf("periksa struktur tabel triase: %w", err)
	}
	return jumlah > 0, nil
}

func (r *Repositori) triase(ctx context.Context, noRawat string, mendukungHandOver bool) (*Triase, error) {
	item := Triase{}
	err := r.db.QueryRowContext(ctx, `
		SELECT t.no_rawat, DATE_FORMAT(t.tgl_kunjungan, '%Y-%m-%d %H:%i:%s'),
		       COALESCE(t.cara_masuk, ''), COALESCE(t.alat_transportasi, ''),
		       COALESCE(t.alasan_kedatangan, ''), COALESCE(t.keterangan_kedatangan, ''),
		       COALESCE(t.kode_kasus, ''), COALESCE(k.macam_kasus, ''),
		       COALESCE(t.tekanan_darah, ''), COALESCE(t.nadi, ''), COALESCE(t.pernapasan, ''),
		       COALESCE(t.suhu, ''), COALESCE(t.saturasi_o2, ''), COALESCE(t.nyeri, '')
		FROM data_triase_igd t
		LEFT JOIN master_triase_macam_kasus k ON k.kode_kasus = t.kode_kasus
		WHERE t.no_rawat = ? LIMIT 1
	`, noRawat).Scan(
		&item.NoRawat, &item.TanggalKunjungan, &item.CaraMasuk, &item.AlatTransportasi,
		&item.AlasanKedatangan, &item.KeteranganKedatangan, &item.KodeKasus, &item.NamaKasus,
		&item.TekananDarah, &item.Nadi, &item.Pernapasan, &item.Suhu, &item.SaturasiO2,
		&item.Nyeri,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("baca data triase IGD: %w", err)
	}
	if mendukungHandOver {
		if err := r.db.QueryRowContext(ctx, `SELECT COALESCE(hand_over, '') FROM data_triase_igd WHERE no_rawat = ?`, noRawat).Scan(&item.HandOver); err != nil {
			return nil, fmt.Errorf("baca hand over triase IGD: %w", err)
		}
	}

	item.Primer, err = r.bagian(ctx, noRawat, "primer")
	if err != nil {
		return nil, err
	}
	item.Sekunder, err = r.bagian(ctx, noRawat, "sekunder")
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repositori) bagian(ctx context.Context, noRawat, jenis string) (*BagianTriase, error) {
	item := BagianTriase{KriteriaTerpilih: []KriteriaSkala{}}
	var err error
	if jenis == "primer" {
		err = r.db.QueryRowContext(ctx, `
			SELECT p.keluhan_utama, COALESCE(p.kebutuhan_khusus, ''), COALESCE(p.catatan, ''),
			       COALESCE(p.plan, ''), DATE_FORMAT(p.tanggaltriase, '%Y-%m-%d %H:%i:%s'),
			       p.nik, COALESCE(g.nama, ''), COALESCE(g.jbtn, '')
			FROM data_triase_igdprimer p LEFT JOIN pegawai g ON g.nik = p.nik
			WHERE p.no_rawat = ? LIMIT 1
		`, noRawat).Scan(&item.IsiUtama, &item.KebutuhanKhusus, &item.Catatan, &item.Plan, &item.TanggalTriase, &item.NIP, &item.NamaPetugas, &item.JabatanPetugas)
	} else {
		err = r.db.QueryRowContext(ctx, `
			SELECT s.anamnesa_singkat, COALESCE(s.catatan, ''), COALESCE(s.plan, ''),
			       DATE_FORMAT(s.tanggaltriase, '%Y-%m-%d %H:%i:%s'), s.nik,
			       COALESCE(g.nama, ''), COALESCE(g.jbtn, '')
			FROM data_triase_igdsekunder s LEFT JOIN pegawai g ON g.nik = s.nik
			WHERE s.no_rawat = ? LIMIT 1
		`, noRawat).Scan(&item.IsiUtama, &item.Catatan, &item.Plan, &item.TanggalTriase, &item.NIP, &item.NamaPetugas, &item.JabatanPetugas)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("baca triase %s: %w", jenis, err)
	}

	awal, akhir := 1, 2
	if jenis == "sekunder" {
		awal, akhir = 3, 5
	}
	for skala := awal; skala <= akhir; skala++ {
		daftar, err := r.kriteriaTerpilih(ctx, noRawat, skala)
		if err != nil {
			return nil, err
		}
		if len(daftar) > 0 {
			item.Skala = skala
			item.KriteriaTerpilih = append(item.KriteriaTerpilih, daftar...)
		}
	}
	return &item, nil
}

func (r *Repositori) kriteriaTerpilih(ctx context.Context, noRawat string, skala int) ([]KriteriaSkala, error) {
	query := fmt.Sprintf(`
		SELECT m.kode_pemeriksaan, m.kode_skala%d, m.pengkajian_skala%d
		FROM master_triase_skala%d m
		INNER JOIN data_triase_igddetail_skala%d d ON d.kode_skala%d = m.kode_skala%d
		WHERE d.no_rawat = ? ORDER BY m.kode_pemeriksaan, m.kode_skala%d
	`, skala, skala, skala, skala, skala, skala, skala)
	rows, err := r.db.QueryContext(ctx, query, noRawat)
	if err != nil {
		return nil, fmt.Errorf("baca detail triase skala %d: %w", skala, err)
	}
	defer rows.Close()
	hasil := make([]KriteriaSkala, 0)
	for rows.Next() {
		item := KriteriaSkala{Skala: skala}
		if err := rows.Scan(&item.KodePemeriksaan, &item.Kode, &item.Pengkajian); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) Simpan(ctx context.Context, input Input, ubah bool) error {
	mendukungHandOver, err := r.mendukungHandOver(ctx)
	if err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	terkunci, err := billingTerkunci(ctx, tx, input.NoRawat)
	if err != nil {
		return err
	}
	if terkunci {
		return ErrBillingTerkunci
	}

	tabelBagian := "data_triase_igdprimer"
	if input.Jenis == "sekunder" {
		tabelBagian = "data_triase_igdsekunder"
	}
	var jumlah int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+tabelBagian+" WHERE no_rawat = ?", input.NoRawat).Scan(&jumlah); err != nil {
		return err
	}
	if !ubah && jumlah > 0 {
		return ErrSudahAda
	}
	if ubah && jumlah == 0 {
		return ErrTidakDitemukan
	}

	if err := r.validasiReferensi(ctx, tx, input); err != nil {
		return err
	}
	queryKedatangan := `
		INSERT INTO data_triase_igd (
			no_rawat, tgl_kunjungan, cara_masuk, alat_transportasi, alasan_kedatangan,
			keterangan_kedatangan, kode_kasus, tekanan_darah, nadi, pernapasan, suhu, saturasi_o2, nyeri
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			tgl_kunjungan=VALUES(tgl_kunjungan), cara_masuk=VALUES(cara_masuk), alat_transportasi=VALUES(alat_transportasi),
			alasan_kedatangan=VALUES(alasan_kedatangan), keterangan_kedatangan=VALUES(keterangan_kedatangan),
			kode_kasus=VALUES(kode_kasus), tekanan_darah=VALUES(tekanan_darah), nadi=VALUES(nadi),
			pernapasan=VALUES(pernapasan), suhu=VALUES(suhu), saturasi_o2=VALUES(saturasi_o2), nyeri=VALUES(nyeri)`
	argumenKedatangan := []any{input.NoRawat, input.TanggalKunjungan, input.CaraMasuk, input.AlatTransportasi,
		input.AlasanKedatangan, input.KeteranganKedatangan, input.KodeKasus, input.TekananDarah,
		input.Nadi, input.Pernapasan, input.Suhu, input.SaturasiO2, input.Nyeri}
	if mendukungHandOver {
		queryKedatangan = `
			INSERT INTO data_triase_igd (
				no_rawat, tgl_kunjungan, cara_masuk, alat_transportasi, alasan_kedatangan,
				keterangan_kedatangan, kode_kasus, tekanan_darah, nadi, pernapasan, suhu, saturasi_o2, nyeri, hand_over
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				tgl_kunjungan=VALUES(tgl_kunjungan), cara_masuk=VALUES(cara_masuk), alat_transportasi=VALUES(alat_transportasi),
				alasan_kedatangan=VALUES(alasan_kedatangan), keterangan_kedatangan=VALUES(keterangan_kedatangan),
				kode_kasus=VALUES(kode_kasus), tekanan_darah=VALUES(tekanan_darah), nadi=VALUES(nadi),
				pernapasan=VALUES(pernapasan), suhu=VALUES(suhu), saturasi_o2=VALUES(saturasi_o2),
				nyeri=VALUES(nyeri), hand_over=VALUES(hand_over)`
		argumenKedatangan = append(argumenKedatangan, input.HandOver)
	}
	_, err = tx.ExecContext(ctx, queryKedatangan, argumenKedatangan...)
	if err != nil {
		return fmt.Errorf("simpan data kedatangan triase: %w", err)
	}

	awal, akhir := 1, 2
	if input.Jenis == "sekunder" {
		awal, akhir = 3, 5
	}
	for skala := awal; skala <= akhir; skala++ {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM data_triase_igddetail_skala%d WHERE no_rawat = ?", skala), input.NoRawat); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM "+tabelBagian+" WHERE no_rawat = ?", input.NoRawat); err != nil {
		return err
	}
	if input.Jenis == "primer" {
		_, err = tx.ExecContext(ctx, `INSERT INTO data_triase_igdprimer (no_rawat, keluhan_utama, kebutuhan_khusus, catatan, plan, tanggaltriase, nik) VALUES (?, ?, ?, ?, ?, ?, ?)`, input.NoRawat, input.IsiUtama, input.KebutuhanKhusus, input.Catatan, input.Plan, input.TanggalTriase, input.NIP)
	} else {
		_, err = tx.ExecContext(ctx, `INSERT INTO data_triase_igdsekunder (no_rawat, anamnesa_singkat, catatan, plan, tanggaltriase, nik) VALUES (?, ?, ?, ?, ?, ?)`, input.NoRawat, input.IsiUtama, input.Catatan, input.Plan, input.TanggalTriase, input.NIP)
	}
	if err != nil {
		return fmt.Errorf("simpan triase %s: %w", input.Jenis, err)
	}
	for _, kode := range input.KodeKriteria {
		query := fmt.Sprintf("INSERT INTO data_triase_igddetail_skala%d (no_rawat, kode_skala%d) VALUES (?, ?)", input.Skala, input.Skala)
		if _, err := tx.ExecContext(ctx, query, input.NoRawat, kode); err != nil {
			return fmt.Errorf("simpan kriteria triase: %w", err)
		}
	}
	return tx.Commit()
}

func (r *Repositori) validasiReferensi(ctx context.Context, tx *sql.Tx, input Input) error {
	var jumlah int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM reg_periksa WHERE no_rawat = ?`, input.NoRawat).Scan(&jumlah); err != nil || jumlah == 0 {
		return fmt.Errorf("%w: nomor rawat tidak terdaftar", ErrInputTidakValid)
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM master_triase_macam_kasus WHERE kode_kasus = ?`, input.KodeKasus).Scan(&jumlah); err != nil || jumlah == 0 {
		return fmt.Errorf("%w: macam kasus tidak terdaftar", ErrInputTidakValid)
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM pegawai WHERE nik = ?`, input.NIP).Scan(&jumlah); err != nil || jumlah == 0 {
		return fmt.Errorf("%w: petugas triase tidak terdaftar", ErrInputTidakValid)
	}
	query := fmt.Sprintf("SELECT COUNT(*) FROM master_triase_skala%d WHERE kode_skala%d = ?", input.Skala, input.Skala)
	for _, kode := range input.KodeKriteria {
		if err := tx.QueryRowContext(ctx, query, kode).Scan(&jumlah); err != nil || jumlah == 0 {
			return fmt.Errorf("%w: kriteria %s tidak terdaftar pada skala %d", ErrInputTidakValid, kode, input.Skala)
		}
	}
	return nil
}

func (r *Repositori) Hapus(ctx context.Context, noRawat, jenis string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	terkunci, err := billingTerkunci(ctx, tx, noRawat)
	if err != nil {
		return err
	}
	if terkunci {
		return ErrBillingTerkunci
	}
	tabel := "data_triase_igdprimer"
	awal, akhir := 1, 2
	if jenis == "sekunder" {
		tabel, awal, akhir = "data_triase_igdsekunder", 3, 5
	}
	var jumlah int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+tabel+" WHERE no_rawat = ?", noRawat).Scan(&jumlah); err != nil {
		return err
	}
	if jumlah == 0 {
		return ErrTidakDitemukan
	}
	for skala := awal; skala <= akhir; skala++ {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM data_triase_igddetail_skala%d WHERE no_rawat = ?", skala), noRawat); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM "+tabel+" WHERE no_rawat = ?", noRawat); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM data_triase_igdprimer WHERE no_rawat=?) + (SELECT COUNT(*) FROM data_triase_igdsekunder WHERE no_rawat=?)`, noRawat, noRawat).Scan(&jumlah); err != nil {
		return err
	}
	if jumlah == 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM data_triase_igd WHERE no_rawat = ?`, noRawat); err != nil {
			return err
		}
	}
	return tx.Commit()
}
