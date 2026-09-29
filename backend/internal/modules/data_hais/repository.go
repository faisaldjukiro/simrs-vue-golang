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
	"simrs-backend/internal/shared/arsiplokal"
	"simrs-backend/internal/shared/khanzamutasi"
)

var ErrValidasi = errors.New("data HAIs tidak valid")
var ErrKonflik = errors.New("catatan sudah ada, berubah, atau hilang; muat ulang riwayat")

const tabel = "data_HAIs"

type Input struct {
	NoRawat string            `json:"no_rawat"`
	Sumber  string            `json:"sumber"`
	Data    DataForm          `json:"data"`
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
		if err != nil || n < 0 {
			return fmt.Errorf("%w: %s wajib berupa bilangan bulat tidak negatif", ErrValidasi, k)
		}
		if k != "HAP" && k != "Tinea" && k != "Scabies" && n > 99 {
			return fmt.Errorf("%w: %s maksimal 99", ErrValidasi, k)
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

func (r *Repositori) Daftar(ctx context.Context, no, username string, admin bool, semua bool, mulai, selesai string) (Hasil, error) {
	h := Hasil{Catatan: []Catatan{}}
	if strings.TrimSpace(no) == "" || len(no) > 17 {
		return h, ErrValidasi
	}
	if semua {
		if _, err := time.Parse("2006-01-02", mulai); err != nil {
			return h, fmt.Errorf("%w: tanggal mulai wajib diisi", ErrValidasi)
		}
		if _, err := time.Parse("2006-01-02", selesai); err != nil || mulai > selesai {
			return h, fmt.Errorf("%w: periode tanggal tidak valid", ErrValidasi)
		}
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
		{r.db, "data_hais", "Arsip lokal", false},
		{r.simrs, tabel, "SIMRS", true},
	} {
		if semua && sumber.db == r.db {
			continue
		}
		selects := []string{"DATE_FORMAT(c.tanggal,'%Y-%m-%d')"}
		for _, k := range Kolom()[1:] {
			selects = append(selects, "COALESCE(c."+k+",'')")
		}
		keys := append([]string{}, Kolom()...)
		from := " FROM " + sumber.tabel + " c"
		where := " WHERE c.no_rawat=?"
		args := []any{no}
		if sumber.db == r.simrs {
			from += " INNER JOIN reg_periksa rp ON rp.no_rawat=c.no_rawat INNER JOIN pasien p ON p.no_rkm_medis=rp.no_rkm_medis INNER JOIN kamar k ON k.kd_kamar=c.kd_kamar INNER JOIN bangsal b ON b.kd_bangsal=k.kd_bangsal"
			selects = append(selects, "c.no_rawat", "rp.no_rkm_medis", "p.nm_pasien", "CONCAT(c.kd_kamar,', ',b.nm_bangsal)")
			keys = append(keys, "no_rawat", "no_rkm_medis", "nm_pasien", "kamar_bangsal")
			if semua {
				where = " WHERE c.tanggal BETWEEN ? AND ?"
				args = []any{mulai, selesai}
			}
		}
		rows, err := sumber.db.QueryContext(ctx, "SELECT "+strings.Join(selects, ",")+from+where+" ORDER BY c.tanggal,c.no_rawat", args...)
		if err != nil {
			if sumber.db == r.db && arsiplokal.TabelTidakAda(err) {
				continue
			}
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
			for i, k := range keys {
				c.Data[k] = nilai[i]
			}
			if c.Data["no_rawat"] == "" {
				c.Data["no_rawat"] = no
			}
			c.BisaUbah = c.BisaUbah && c.Data["no_rawat"] == no
			h.Catatan = append(h.Catatan, c)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return h, err
		}
	}
	sort.SliceStable(h.Catatan, func(i, j int) bool { return h.Catatan[i].Data["tanggal"] < h.Catatan[j].Data["tanggal"] })
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
	if metode != "POST" && in.Sumber != "SIMRS" {
		return fmt.Errorf("%w: arsip lokal hanya baca; rekonsiliasi diperlukan", ErrValidasi)
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
		_, err = r.simrs.ExecContext(ctx, "INSERT INTO "+tabel+" ("+strings.Join(kolom, ",")+") VALUES ("+strings.TrimSuffix(strings.Repeat("?,", len(kolom)), ",")+")", baru...)
		return err
	}
	return khanzamutasi.Jalankan(ctx, r.simrs, tabel, kolom, lama, baru, metode == "DELETE")
}
