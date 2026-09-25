package risiko_jatuh_dewasa

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	"simrs-backend/internal/shared/khanzamutasi"
)

var ErrValidasi = errors.New("data penilaian tidak valid")
var ErrAkses = errors.New("catatan hanya boleh diubah oleh petugas yang tercatat atau administrator")
var ErrKonflik = errors.New("catatan sudah ada, berubah, atau hilang; muat ulang riwayat")

const tabel = "penilaian_lanjutan_resiko_jatuh_dewasa"

type Skala struct {
	Label   string   `json:"label"`
	Pilihan []string `json:"pilihan"`
	Nilai   []int    `json:"nilai"`
}

// RMPenilaianLanjutanRisikoJatuhDewasa: SkalaResiko1..6ItemStateChanged.
var SkalaMorse = []Skala{
	{"Riwayat Jatuh (1 Tahun Terakhir)", []string{"Tidak", "Ya"}, []int{0, 25}},
	{"Diagnosis Sekunder (≥ 2 Diagnosis Medis)", []string{"Tidak", "Ya"}, []int{0, 15}},
	{"Alat Bantu", []string{"Tidak Ada/Kursi Roda/Perawat/Tirah Baring", "Tongkat/Alat Penopang", "Berpegangan Pada Perabot"}, []int{0, 15, 30}},
	{"Terpasang Infus", []string{"Tidak", "Ya"}, []int{0, 20}},
	{"Gaya Berjalan", []string{"Normal/Tirah Baring/Imobilisasi", "Lemah", "Terganggu"}, []int{0, 10, 20}},
	{"Status Mental", []string{"Sadar Akan Kemampuan Diri Sendiri", "Sering Lupa Akan Keterbatasan Yang Dimiliki"}, []int{0, 15}},
}

type Input struct {
	NoRawat string            `json:"no_rawat"`
	Data    map[string]string `json:"data"`
	Asli    map[string]string `json:"asli"`
}

type Catatan struct {
	Data        map[string]string `json:"data"`
	NamaPetugas string            `json:"nama_petugas"`
	BisaUbah    bool              `json:"bisa_ubah"`
	Risiko      string            `json:"risiko"`
}

type Hasil struct {
	Skala   []Skala   `json:"skala"`
	Catatan []Catatan `json:"catatan"`
}

type Repositori struct{ db *sql.DB }

func NewRepositori(db *sql.DB) *Repositori { return &Repositori{db: db} }

func Kolom() []string {
	k := []string{"tanggal"}
	for i := 1; i <= 6; i++ {
		n := strconv.Itoa(i)
		k = append(k, "penilaian_jatuhmorse_skala"+n, "penilaian_jatuhmorse_nilai"+n)
	}
	return append(k, "penilaian_jatuhmorse_totalnilai", "hasil_skrining", "saran", "nip")
}

func Risiko(total int) string {
	if total < 25 {
		return "Rendah"
	}
	if total < 45 {
		return "Sedang"
	}
	return "Tinggi"
}

func Validasi(in *Input) error {
	if strings.TrimSpace(in.NoRawat) == "" || len(in.NoRawat) > 17 || in.Data == nil {
		return ErrValidasi
	}
	in.Data["tanggal"] = strings.ReplaceAll(in.Data["tanggal"], "T", " ")
	if len(in.Data["tanggal"]) == 16 {
		in.Data["tanggal"] += ":00"
	}
	if _, err := time.Parse("2006-01-02 15:04:05", in.Data["tanggal"]); err != nil {
		return fmt.Errorf("%w: tanggal/jam tidak valid", ErrValidasi)
	}
	for k, batas := range map[string]int{"hasil_skrining": 200, "saran": 200, "nip": 20} {
		v := strings.TrimSpace(in.Data[k])
		if v == "" || utf8.RuneCountInString(v) > batas {
			return fmt.Errorf("%w: %s wajib diisi, maksimal %d karakter", ErrValidasi, k, batas)
		}
		in.Data[k] = v
	}
	total := 0
	for i, s := range SkalaMorse {
		n := strconv.Itoa(i + 1)
		index := -1
		for j, p := range s.Pilihan {
			if in.Data["penilaian_jatuhmorse_skala"+n] == p {
				index = j
				break
			}
		}
		if index < 0 {
			return fmt.Errorf("%w: pilih %s", ErrValidasi, s.Label)
		}
		nilai := s.Nilai[index]
		in.Data["penilaian_jatuhmorse_nilai"+n] = strconv.Itoa(nilai)
		total += nilai
	}
	in.Data["penilaian_jatuhmorse_totalnilai"] = strconv.Itoa(total)
	return nil
}

func (r *Repositori) Daftar(ctx context.Context, no, username string, admin bool) (Hasil, error) {
	h := Hasil{Skala: SkalaMorse, Catatan: []Catatan{}}
	if strings.TrimSpace(no) == "" || len(no) > 17 {
		return h, ErrValidasi
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
		total, e := strconv.Atoi(c.Data["penilaian_jatuhmorse_totalnilai"])
		if e == nil {
			c.Risiko = Risiko(total)
		}
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
