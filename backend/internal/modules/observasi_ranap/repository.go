package observasi_ranap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	"simrs-backend/internal/shared/khanzamutasi"
)

var ErrValidasi = errors.New("data observasi tidak valid")
var ErrAkses = errors.New("catatan hanya boleh diubah oleh petugas yang tercatat atau administrator")
var ErrKonflik = errors.New("catatan sudah ada, berubah, atau hilang; muat ulang riwayat")

const tabel = "catatan_observasi_ranap"

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

type Hasil struct {
	Catatan []Catatan `json:"catatan"`
}

type Repositori struct{ db *sql.DB }

func NewRepositori(db *sql.DB) *Repositori { return &Repositori{db: db} }

func Kolom() []string {
	return []string{"tgl_perawatan", "jam_rawat", "gcs", "td", "hr", "rr", "suhu", "spo2", "nip"}
}

func Validasi(in *Input) error {
	if strings.TrimSpace(in.NoRawat) == "" || len(in.NoRawat) > 17 || in.Data == nil {
		return ErrValidasi
	}
	if _, err := time.Parse("2006-01-02", in.Data["tgl_perawatan"]); err != nil {
		return fmt.Errorf("%w: tanggal tidak valid", ErrValidasi)
	}
	if len(in.Data["jam_rawat"]) == 5 {
		in.Data["jam_rawat"] += ":00"
	}
	if _, err := time.Parse("15:04:05", in.Data["jam_rawat"]); err != nil {
		return fmt.Errorf("%w: jam tidak valid", ErrValidasi)
	}
	for k, batas := range map[string]int{"gcs": 10, "td": 8, "hr": 5, "rr": 5, "suhu": 5, "spo2": 3, "nip": 20} {
		v := strings.TrimSpace(in.Data[k])
		if utf8.RuneCountInString(v) > batas || (k == "nip" && v == "") {
			return fmt.Errorf("%w: %s maksimal %d karakter; petugas wajib diisi", ErrValidasi, k, batas)
		}
		in.Data[k] = v
	}
	return nil
}

func (r *Repositori) Daftar(ctx context.Context, no, username string, admin bool) (Hasil, error) {
	h := Hasil{Catatan: []Catatan{}}
	if strings.TrimSpace(no) == "" || len(no) > 17 {
		return h, ErrValidasi
	}
	kolom := Kolom()
	selects := []string{"DATE_FORMAT(c.tgl_perawatan,'%Y-%m-%d')", "TIME_FORMAT(c.jam_rawat,'%H:%i:%s')"}
	for _, k := range kolom[2:] {
		selects = append(selects, "COALESCE(c."+k+",'')")
	}
	selects = append(selects, "COALESCE(p.nama,'')")
	rows, err := r.db.QueryContext(ctx, "SELECT "+strings.Join(selects, ",")+" FROM "+tabel+" c LEFT JOIN petugas p ON p.nip=c.nip WHERE c.no_rawat=? ORDER BY c.tgl_perawatan DESC,c.jam_rawat DESC", no)
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

func (r *Repositori) Mutasi(ctx context.Context, in Input, metode, username string, admin bool) (err error) {
	if in.NoRawat == "" || len(in.NoRawat) > 17 {
		return ErrValidasi
	}
	if metode != "POST" && metode != "PUT" && metode != "DELETE" {
		return ErrValidasi
	}
	if metode != "DELETE" {
		if err = Validasi(&in); err != nil {
			return err
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
