package data_hais

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	"simrs-backend/internal/shared/khanzamutasi"
)

var ErrValidasi = errors.New("data HAIs tidak valid")
var ErrKonflik = errors.New("catatan sudah ada, berubah, atau hilang; muat ulang riwayat")

const tabel = "data_hais"

type Input struct {
	NoRawat string            `json:"no_rawat"`
	Data    map[string]string `json:"data"`
	Asli    map[string]string `json:"asli"`
}

type Catatan struct {
	Data     map[string]string `json:"data"`
	Sumber   string            `json:"sumber"`
	BisaUbah bool              `json:"bisa_ubah"`
}

type Hasil struct {
	Catatan []Catatan `json:"catatan"`
	Kamar   string    `json:"kamar"`
}

type Repositori struct{ db, simrs *sql.DB }

func NewRepositori(db, simrs *sql.DB) *Repositori { return &Repositori{db: db, simrs: simrs} }

func Kolom() []string {
	return []string{"tanggal", "ETT", "CVL", "IVL", "UC", "VAP", "IAD", "PLEB", "ISK", "ILO", "HAP", "Tinea", "Scabies", "DEKU", "SPUTUM", "DARAH", "URINE", "ANTIBIOTIK", "kd_kamar"}
}

func Validasi(in *Input) error {
	if strings.TrimSpace(in.NoRawat) == "" || len(in.NoRawat) > 17 || in.Data == nil {
		return ErrValidasi
	}

	if _, err := time.Parse("2006-01-02", in.Data["tanggal"]); err != nil {
		return fmt.Errorf("%w: tanggal tidak valid", ErrValidasi)
	}
	for _, k := range Kolom()[1:13] {
		v := strings.TrimSpace(in.Data[k])
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 || n > 99 {
			return fmt.Errorf("%w: %s wajib berupa bilangan bulat 0 sampai 99", ErrValidasi, k)
		}
		in.Data[k] = strconv.Itoa(n)
	}
	if in.Data["DEKU"] != "IYA" && in.Data["DEKU"] != "TIDAK" {
		return fmt.Errorf("%w: pilihan dekubitus tidak valid", ErrValidasi)
	}
	for _, k := range []string{"SPUTUM", "DARAH", "URINE", "ANTIBIOTIK"} {
		in.Data[k] = strings.TrimSpace(in.Data[k])
		if utf8.RuneCountInString(in.Data[k]) > 200 {
			return fmt.Errorf("%w: %s maksimal 200 karakter", ErrValidasi, k)
		}
	}

	return nil
}

func (r *Repositori) Daftar(ctx context.Context, no, username string, admin bool) (Hasil, error) {
	h := Hasil{Catatan: []Catatan{}}
	if strings.TrimSpace(no) == "" || len(no) > 17 {
		return h, ErrValidasi
	}

	kamar, err := r.kamar(ctx, no)
	if err != nil {
		return h, err
	}
	h.Kamar = kamar
	for _, sumber := range []struct {
		db          *sql.DB
		tabel, nama string
		ubah        bool
	}{
		{r.db, tabel, "SIRAPI", true},
		{r.simrs, "data_HAIs", "Riwayat lama", false},
	} {
		selects := []string{"DATE_FORMAT(c.tanggal,'%Y-%m-%d')"}
		for _, k := range Kolom()[1:] {
			selects = append(selects, "COALESCE(c."+k+",'')")
		}
		rows, err := sumber.db.QueryContext(ctx, "SELECT "+strings.Join(selects, ",")+" FROM "+sumber.tabel+" c WHERE c.no_rawat=? ORDER BY c.tanggal DESC", no)
		if err != nil {
			return h, err
		}
		for rows.Next() {
			nilai := make([]string, len(selects))
			tujuan := make([]any, len(nilai))
			for i := range nilai {
				tujuan[i] = &nilai[i]
			}
			if err := rows.Scan(tujuan...); err != nil {
				rows.Close()
				return h, err
			}
			c := Catatan{Data: map[string]string{}, Sumber: sumber.nama, BisaUbah: sumber.ubah}
			for i, k := range Kolom() {
				c.Data[k] = nilai[i]
			}
			h.Catatan = append(h.Catatan, c)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return h, err
		}
	}
	sort.SliceStable(h.Catatan, func(i, j int) bool { return h.Catatan[i].Data["tanggal"] > h.Catatan[j].Data["tanggal"] })
	return h, nil
}

// Hanya SELECT pada koneksi SIMRS lama; termasuk kamar ibu untuk rawat gabung.
func (r *Repositori) kamar(ctx context.Context, no string) (string, error) {
	var kamar string
	err := r.simrs.QueryRowContext(ctx, `SELECT COALESCE(ki.kd_kamar,'') FROM kamar_inap ki
 WHERE ki.no_rawat=COALESCE((SELECT rg.no_rawat FROM ranap_gabung rg WHERE rg.no_rawat2=? LIMIT 1),?)
 ORDER BY ki.tgl_masuk DESC,ki.jam_masuk DESC LIMIT 1`, no, no).Scan(&kamar)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return kamar, err
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
	var ada bool
	if err = r.simrs.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM reg_periksa WHERE no_rawat=?)", in.NoRawat).Scan(&ada); err != nil {
		return err
	}
	if !ada {
		return fmt.Errorf("%w: kunjungan tidak ditemukan", ErrValidasi)
	}

	if metode != "DELETE" {
		if metode == "POST" {
			in.Data["kd_kamar"], err = r.kamar(ctx, in.NoRawat)
			if err != nil {
				return err
			}
			if in.Data["kd_kamar"] == "" {
				return fmt.Errorf("%w: kamar rawat inap belum ditemukan", ErrValidasi)
			}
		} else {
			in.Data["kd_kamar"] = in.Asli["kd_kamar"]
		}
		baru[len(baru)-1] = in.Data["kd_kamar"]
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
