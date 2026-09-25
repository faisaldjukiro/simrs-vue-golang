package pews_anak

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	"simrs-backend/internal/shared/khanzamutasi"
)

var ErrValidasi = errors.New("data PEWS Anak tidak valid")
var ErrAkses = errors.New("catatan hanya boleh disimpan atas nama petugas login dan diubah oleh pembuatnya atau administrator")
var ErrKonflik = errors.New("catatan sudah ada, berubah, atau hilang; muat ulang riwayat")

const tabel = "pemantauan_pews_anak"

type Input struct {
	NoRawat string            `json:"no_rawat"`
	Data    map[string]string `json:"data"`
	Asli    map[string]string `json:"asli"`
}

type Catatan struct {
	Data        map[string]string `json:"data"`
	NamaPetugas string            `json:"nama_petugas"`
	BisaUbah    bool              `json:"bisa_ubah"`
}

type Petugas struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

type Hasil struct {
	Catatan           []Catatan `json:"catatan"`
	Bidang            []Bidang  `json:"bidang"`
	Panduan           []Panduan `json:"panduan"`
	PetugasLogin      Petugas   `json:"petugas_login"`
	BolehPilihPetugas bool      `json:"boleh_pilih_petugas"`
}

type Repositori struct{ db *sql.DB }

func NewRepositori(db *sql.DB) *Repositori { return &Repositori{db: db} }

func Kolom() []string {
	return []string{
		"tanggal", "parameter_perilaku", "skor_perilaku",
		"parameter_crt_atau_warna_kulit", "skor_crt_atau_warna_kulit",
		"parameter_perespirasi", "skor_perespirasi", "skor_total", "parameter_total", "nip",
	}
}

func (r *Repositori) Daftar(ctx context.Context, no, username string, admin bool) (Hasil, error) {
	h := Hasil{Catatan: []Catatan{}, Bidang: BidangPenilaian(), Panduan: PanduanPenilaian(), BolehPilihPetugas: admin}
	if strings.TrimSpace(no) == "" || len(no) > 17 {
		return h, ErrValidasi
	}
	err := r.db.QueryRowContext(ctx, "SELECT nip,nama FROM petugas WHERE nip=?", username).Scan(&h.PetugasLogin.Kode, &h.PetugasLogin.Nama)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return h, err
	}
	kolom := Kolom()
	selects := []string{"DATE_FORMAT(c.tanggal,'%Y-%m-%d %H:%i:%s')"}
	for _, k := range kolom[1:] {
		selects = append(selects, "COALESCE(c."+k+",'')")
	}
	selects = append(selects, "COALESCE(p.nama,'')")
	rows, err := r.db.QueryContext(ctx, "SELECT "+strings.Join(selects, ",")+" FROM "+tabel+" c LEFT JOIN petugas p ON p.nip=c.nip WHERE c.no_rawat=? ORDER BY c.tanggal DESC", no)
	if err != nil {
		return h, err
	}
	defer rows.Close()
	for rows.Next() {
		nilai := make([]string, len(selects))
		tujuan := make([]any, len(nilai))
		for i := range nilai {
			tujuan[i] = &nilai[i]
		}
		if err = rows.Scan(tujuan...); err != nil {
			return h, err
		}
		c := Catatan{Data: map[string]string{}, NamaPetugas: nilai[len(nilai)-1]}
		for i, k := range kolom {
			c.Data[k] = nilai[i]
		}
		c.BisaUbah = admin || (username != "" && c.Data["nip"] == username)
		h.Catatan = append(h.Catatan, c)
	}
	return h, rows.Err()
}

func (r *Repositori) CariPetugas(ctx context.Context, q, username string, admin bool) ([]Petugas, error) {
	hasil := []Petugas{}
	q = strings.TrimSpace(q)
	if len(q) < 2 {
		return hasil, nil
	}
	query := "SELECT nip,nama FROM petugas WHERE (nip LIKE ? OR nama LIKE ?)"
	args := []any{"%" + q + "%", "%" + q + "%"}
	if !admin {
		query += " AND nip=?"
		args = append(args, username)
	}
	rows, err := r.db.QueryContext(ctx, query+" ORDER BY nama LIMIT 30", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p Petugas
		if err = rows.Scan(&p.Kode, &p.Nama); err != nil {
			return nil, err
		}
		hasil = append(hasil, p)
	}
	return hasil, rows.Err()
}

func (r *Repositori) Mutasi(ctx context.Context, in Input, metode, username string, admin bool) (err error) {
	if strings.TrimSpace(in.NoRawat) == "" || len(in.NoRawat) > 17 {
		return ErrValidasi
	}
	if metode != "POST" && metode != "PUT" && metode != "DELETE" {
		return ErrValidasi
	}
	if metode != "DELETE" {
		if err = Validasi(&in); err != nil {
			return err
		}
		if !admin && (username == "" || in.Data["nip"] != username) {
			return ErrAkses
		}
	}
	kolom := append([]string{"no_rawat"}, Kolom()...)
	lama, baru := []any{in.NoRawat}, []any{in.NoRawat}
	for _, k := range kolom[1:] {
		if metode != "POST" {
			if _, ok := in.Asli[k]; !ok {
				return ErrValidasi
			}
		}
		lama = append(lama, in.Asli[k])
		baru = append(baru, in.Data[k])
	}
	if metode != "POST" && !admin && (username == "" || in.Asli["nip"] != username) {
		return ErrAkses
	}
	var ada bool
	if err = r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM reg_periksa WHERE no_rawat=?)", in.NoRawat).Scan(&ada); err != nil {
		return err
	}
	if !ada {
		return fmt.Errorf("%w: kunjungan tidak ditemukan", ErrValidasi)
	}
	if metode != "DELETE" {
		if err = r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM petugas WHERE nip=?)", in.Data["nip"]).Scan(&ada); err != nil {
			return err
		}
		if !ada {
			return fmt.Errorf("%w: petugas tidak ditemukan", ErrValidasi)
		}
	}
	defer func() {
		var e *mysql.MySQLError
		if errors.Is(err, khanzamutasi.ErrKonflik) || (errors.As(err, &e) && e.Number == 1062) {
			err = ErrKonflik
		}
	}()
	if metode == "POST" {
		_, err = r.db.ExecContext(ctx, "INSERT INTO "+tabel+" ("+strings.Join(kolom, ",")+") VALUES ("+strings.TrimSuffix(strings.Repeat("?,", len(kolom)), ",")+")", baru...)
		return err
	}
	return khanzamutasi.Jalankan(ctx, r.db, tabel, kolom, lama, baru, metode == "DELETE")
}
