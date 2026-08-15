package awal_keperawatan_igd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrTidakDitemukan = errors.New("penilaian awal keperawatan IGD tidak ditemukan")
	ErrSudahAda       = errors.New("penilaian awal keperawatan IGD sudah ada")
)

type Petugas struct {
	NIP     string `json:"nip"`
	Nama    string `json:"nama"`
	Jabatan string `json:"jabatan"`
}

type Masalah struct {
	Kode     string `json:"kode"`
	Nama     string `json:"nama"`
	Terpilih bool   `json:"terpilih"`
}

type RencanaKeperawatan struct {
	Kode        string `json:"kode"`
	KodeMasalah string `json:"kode_masalah"`
	Nama        string `json:"nama"`
	Terpilih    bool   `json:"terpilih"`
}

type Rekaman struct {
	Input
	NamaPetugas    string `json:"nama_petugas"`
	JabatanPetugas string `json:"jabatan_petugas"`
}

type Data struct {
	Petugas         Petugas              `json:"petugas"`
	Masalah         []Masalah            `json:"masalah_keperawatan"`
	Rencana         []RencanaKeperawatan `json:"rencana_keperawatan"`
	Penilaian       *Rekaman             `json:"penilaian"`
	BillingTerkunci bool                 `json:"billing_terkunci"`
}

type Repositori struct{ db *sql.DB }

func NewRepositori(db *sql.DB) *Repositori { return &Repositori{db: db} }

const kolomUtama = "no_rawat,tanggal,informasi,keluhan_utama,rpd,rpo,status_kehamilan,gravida,para,abortus,hpht,tekanan,pupil,neurosensorik,integumen,turgor,edema,mukosa,perdarahan,jumlah_perdarahan,warna_perdarahan,intoksikasi,bab,xbab,kbab,wbab,bak,xbak,wbak,lbak,psikologis,jiwa,perilaku,dilaporkan,sebutkan,hubungan,tinggal_dengan,ket_tinggal,budaya,ket_budaya,pendidikan_pj,ket_pendidikan_pj,edukasi,ket_edukasi,kemampuan,aktifitas,alat_bantu,ket_bantu,nyeri,provokes,ket_provokes,quality,ket_quality,lokasi,menyebar,skala_nyeri,durasi,nyeri_hilang,ket_nyeri,pada_dokter,ket_dokter,berjalan_a,berjalan_b,berjalan_c,hasil,lapor,ket_lapor,rencana,nip"

func (r *Repositori) Data(ctx context.Context, noRawat, username string) (Data, error) {
	data := Data{Masalah: []Masalah{}, Rencana: []RencanaKeperawatan{}}
	var err error
	data.Petugas, err = r.petugasLogin(ctx, username)
	if err != nil {
		return Data{}, err
	}
	data.Masalah, err = r.masalah(ctx, noRawat)
	if err != nil {
		return Data{}, err
	}
	data.Rencana, err = r.rencana(ctx, noRawat)
	if err != nil {
		return Data{}, err
	}
	data.Penilaian, err = r.penilaian(ctx, noRawat)
	if err != nil {
		return Data{}, err
	}
	data.BillingTerkunci, err = r.BillingTerkunci(ctx, noRawat)
	return data, err
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
		return false, fmt.Errorf("periksa status billing awal keperawatan IGD: %w", err)
	}
	return jumlah > 0, nil
}

func (r *Repositori) petugasLogin(ctx context.Context, username string) (Petugas, error) {
	var item Petugas
	err := r.db.QueryRowContext(ctx, `SELECT nik,COALESCE(nama,''),COALESCE(jbtn,'') FROM pegawai WHERE nik=? LIMIT 1`, strings.TrimSpace(username)).Scan(&item.NIP, &item.Nama, &item.Jabatan)
	if errors.Is(err, sql.ErrNoRows) {
		return Petugas{}, nil
	}
	if err != nil {
		return Petugas{}, fmt.Errorf("baca petugas login: %w", err)
	}
	return item, nil
}

func (r *Repositori) masalah(ctx context.Context, noRawat string) ([]Masalah, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT m.kode_masalah,COALESCE(m.nama_masalah,''),IF(p.no_rawat IS NULL,0,1) FROM master_masalah_keperawatan_igd m LEFT JOIN penilaian_awal_keperawatan_igd_masalah p ON p.kode_masalah=m.kode_masalah AND p.no_rawat=? ORDER BY m.kode_masalah`, noRawat)
	if err != nil {
		return nil, fmt.Errorf("baca masalah keperawatan IGD: %w", err)
	}
	defer rows.Close()
	hasil := make([]Masalah, 0)
	for rows.Next() {
		var item Masalah
		if err := rows.Scan(&item.Kode, &item.Nama, &item.Terpilih); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) rencana(ctx context.Context, noRawat string) ([]RencanaKeperawatan, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT r.kode_rencana,r.kode_masalah,r.rencana_keperawatan,IF(p.no_rawat IS NULL,0,1) FROM master_rencana_keperawatan_igd r LEFT JOIN penilaian_awal_keperawatan_ralan_rencana_igd p ON p.kode_rencana=r.kode_rencana AND p.no_rawat=? ORDER BY r.kode_masalah,r.kode_rencana`, noRawat)
	if err != nil {
		return nil, fmt.Errorf("baca rencana keperawatan IGD: %w", err)
	}
	defer rows.Close()
	hasil := make([]RencanaKeperawatan, 0)
	for rows.Next() {
		var item RencanaKeperawatan
		if err := rows.Scan(&item.Kode, &item.KodeMasalah, &item.Nama, &item.Terpilih); err != nil {
			return nil, err
		}
		hasil = append(hasil, item)
	}
	return hasil, rows.Err()
}

func (r *Repositori) penilaian(ctx context.Context, noRawat string) (*Rekaman, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT "+kolomUtama+" FROM penilaian_awal_keperawatan_igd WHERE no_rawat=?", noRawat)
	if err != nil {
		return nil, fmt.Errorf("baca penilaian awal keperawatan IGD: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var item Rekaman
	if err := pindaiInput(rows, &item.Input); err != nil {
		return nil, err
	}
	_ = r.db.QueryRowContext(ctx, `SELECT COALESCE(nama,''),COALESCE(jbtn,'') FROM pegawai WHERE nik=? LIMIT 1`, item.NIP).Scan(&item.NamaPetugas, &item.JabatanPetugas)
	item.KodeMasalah = kodeTerpilihMasalah(ctx, r.db, noRawat)
	item.KodeRencana = kodeTerpilihRencana(ctx, r.db, noRawat)
	return &item, nil
}

func pindaiInput(rows *sql.Rows, input *Input) error {
	kolom, err := rows.Columns()
	if err != nil {
		return err
	}
	nilai := make([]sql.NullString, len(kolom))
	tujuan := make([]any, len(kolom))
	for i := range nilai {
		tujuan[i] = &nilai[i]
	}
	if err := rows.Scan(tujuan...); err != nil {
		return err
	}
	rv := reflect.ValueOf(input).Elem()
	rt := rv.Type()
	indeks := map[string]int{}
	for i := 0; i < rt.NumField(); i++ {
		tag := strings.Split(rt.Field(i).Tag.Get("json"), ",")[0]
		indeks[tag] = i
	}
	for i, nama := range kolom {
		if field, ok := indeks[nama]; ok && rv.Field(field).CanSet() {
			rv.Field(field).SetString(nilai[i].String)
		}
	}
	return nil
}

func kodeTerpilihMasalah(ctx context.Context, db *sql.DB, noRawat string) []string {
	return kodeTerpilih(ctx, db, `SELECT kode_masalah FROM penilaian_awal_keperawatan_igd_masalah WHERE no_rawat=? ORDER BY kode_masalah`, noRawat)
}
func kodeTerpilihRencana(ctx context.Context, db *sql.DB, noRawat string) []string {
	return kodeTerpilih(ctx, db, `SELECT kode_rencana FROM penilaian_awal_keperawatan_ralan_rencana_igd WHERE no_rawat=? ORDER BY kode_rencana`, noRawat)
}
func kodeTerpilih(ctx context.Context, db *sql.DB, query, noRawat string) []string {
	rows, err := db.QueryContext(ctx, query, noRawat)
	if err != nil {
		return []string{}
	}
	defer rows.Close()
	hasil := []string{}
	for rows.Next() {
		var kode string
		if rows.Scan(&kode) == nil {
			hasil = append(hasil, kode)
		}
	}
	return hasil
}

func (r *Repositori) Simpan(ctx context.Context, input Input, ubah bool) error {
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
	var jumlah int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM reg_periksa WHERE no_rawat=?`, input.NoRawat).Scan(&jumlah); err != nil || jumlah == 0 {
		return fmt.Errorf("%w: nomor rawat tidak terdaftar", ErrInputTidakValid)
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM petugas WHERE nip=?`, input.NIP).Scan(&jumlah); err != nil || jumlah == 0 {
		return fmt.Errorf("%w: petugas tidak terdaftar", ErrInputTidakValid)
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM penilaian_awal_keperawatan_igd WHERE no_rawat=?`, input.NoRawat).Scan(&jumlah); err != nil {
		return err
	}
	if ubah && jumlah == 0 {
		return ErrTidakDitemukan
	}
	if !ubah && jumlah > 0 {
		return ErrSudahAda
	}
	args := inputArgs(input)
	if ubah {
		kolom := strings.Split(kolomUtama, ",")
		set := make([]string, 0, len(kolom)-1)
		for _, nama := range kolom[1:] {
			set = append(set, nama+"=?")
		}
		_, err = tx.ExecContext(ctx, "UPDATE penilaian_awal_keperawatan_igd SET "+strings.Join(set, ",")+" WHERE no_rawat=?", append(args[1:], input.NoRawat)...)
	} else {
		placeholder := strings.TrimSuffix(strings.Repeat("?,", 69), ",")
		_, err = tx.ExecContext(ctx, "INSERT INTO penilaian_awal_keperawatan_igd ("+kolomUtama+") VALUES ("+placeholder+")", args...)
	}
	if err != nil {
		return fmt.Errorf("simpan penilaian awal keperawatan IGD: %w", err)
	}
	if err := simpanRelasi(ctx, tx, input); err != nil {
		return err
	}
	return tx.Commit()
}

func simpanRelasi(ctx context.Context, tx *sql.Tx, input Input) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM penilaian_awal_keperawatan_igd_masalah WHERE no_rawat=?`, input.NoRawat); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM penilaian_awal_keperawatan_ralan_rencana_igd WHERE no_rawat=?`, input.NoRawat); err != nil {
		return err
	}
	for _, kode := range input.KodeMasalah {
		if _, err := tx.ExecContext(ctx, `INSERT INTO penilaian_awal_keperawatan_igd_masalah(no_rawat,kode_masalah) SELECT ?,kode_masalah FROM master_masalah_keperawatan_igd WHERE kode_masalah=?`, input.NoRawat, kode); err != nil {
			return err
		}
	}
	for _, kode := range input.KodeRencana {
		if _, err := tx.ExecContext(ctx, `INSERT INTO penilaian_awal_keperawatan_ralan_rencana_igd(no_rawat,kode_rencana) SELECT ?,kode_rencana FROM master_rencana_keperawatan_igd WHERE kode_rencana=?`, input.NoRawat, kode); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repositori) Hapus(ctx context.Context, noRawat string) error {
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
	result, err := tx.ExecContext(ctx, `DELETE FROM penilaian_awal_keperawatan_igd WHERE no_rawat=?`, noRawat)
	if err != nil {
		return fmt.Errorf("hapus penilaian awal keperawatan IGD: %w", err)
	}
	jumlah, _ := result.RowsAffected()
	if jumlah == 0 {
		return ErrTidakDitemukan
	}
	return tx.Commit()
}

func inputArgs(i Input) []any {
	return []any{
		i.NoRawat, i.Tanggal, i.Informasi, i.KeluhanUtama, i.RPD, i.RPO, i.StatusKehamilan, i.Gravida, i.Para, i.Abortus, i.HPHT, i.Tekanan, i.Pupil, i.Neurosensorik, i.Integumen, i.Turgor, i.Edema, i.Mukosa, i.Perdarahan, i.JumlahPerdarahan, i.WarnaPerdarahan, i.Intoksikasi, i.BAB, i.XBAB, i.KBAB, i.WBAB, i.BAK, i.XBAK, i.WBAK, i.LBAK, i.Psikologis, i.Jiwa, i.Perilaku, i.Dilaporkan, i.Sebutkan, i.Hubungan, i.TinggalDengan, i.KetTinggal, i.Budaya, i.KetBudaya, i.PendidikanPJ, i.KetPendidikanPJ, i.Edukasi, i.KetEdukasi, i.Kemampuan, i.Aktifitas, i.AlatBantu, i.KetBantu, i.Nyeri, i.Provokes, i.KetProvokes, i.Quality, i.KetQuality, i.Lokasi, i.Menyebar, i.SkalaNyeri, i.Durasi, i.NyeriHilang, i.KetNyeri, i.PadaDokter, i.KetDokter, i.BerjalanA, i.BerjalanB, i.BerjalanC, i.Hasil, i.Lapor, i.KetLapor, i.Rencana, i.NIP,
	}
}
