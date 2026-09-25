package risiko_jatuh_anak

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

const tabel = "penilaian_lanjutan_resiko_jatuh_anak"

type Skala struct {
	Label   string   `json:"label"`
	Pilihan []string `json:"pilihan"`
	Nilai   []int    `json:"nilai"`
}

// RMPenilaianLanjutanRisikoJatuhAnak: SkalaResiko1..7ItemStateChanged.
var SkalaHumptyDumpty = []Skala{
	{"Umur", []string{"0 - 3 Tahun", "3 - 7 Tahun", "7 - 13 Tahun", "> 13 Tahun"}, []int{4, 3, 2, 1}},
	{"Jenis Kelamin", []string{"Laki-laki", "Perempuan"}, []int{2, 1}},
	{"Diagnosa", []string{"Kelainan Neurologi", "Perubahan Dalam Oksigen(Masalah Saluran Nafas, Dehidrasi, Anemia, Anoreksia / Sakit Kepala, Dll)", "Kelainan Psikis / Perilaku", "Diagnosa Lain"}, []int{4, 3, 2, 1}},
	{"Gangguan Kognitif", []string{"Tidak Sadar Terhadap Keterbatasan", "Lupa Keterbatasan", "Mengetahui Kemampuan Diri"}, []int{3, 2, 1}},
	{"Faktor Lingkungan", []string{"Riwayat Jatuh Dari Tempat Tidur Saat Bayi/Anak", "Pasien Menggunakan Alat Bantu/Box/Mebel", "Pasien Berada Di Tempat Tidur", "Di Luar Ruang Rawat"}, []int{4, 3, 2, 1}},
	{"Efek Obat Penenang/Operasi/Anastesi", []string{"Dalam 24 Jam", "Dalam 48 Jam", "> 48 Jam"}, []int{3, 2, 1}},
	{"Penggunaan Obat", []string{"Bermacam-macam Obat Yang Digunakan : Obat Sedative (Kecuali Pasien ICU Yang Menggunakan sedasi dan paralisis), Hipnotik, Barbiturat, Fenoti-Azin, Antidepresan, Laksans/Diuretika,Narkotik", "Salah Satu Dari Pengobatan Di Atas", "Pengobatan Lain"}, []int{3, 2, 1}},
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
	for i := 1; i <= 7; i++ {
		n := strconv.Itoa(i)
		k = append(k, "penilaian_humptydumpty_skala"+n, "penilaian_humptydumpty_nilai"+n)
	}
	return append(k, "penilaian_humptydumpty_totalnilai", "hasil_skrining", "saran", "nip")
}

func Risiko(total int) string {
	if total < 7 || total > 23 {
		return "Skor tidak valid"
	}
	if total < 12 {
		return "Rendah"
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
	for i, s := range SkalaHumptyDumpty {
		n := strconv.Itoa(i + 1)
		index := -1
		for j, p := range s.Pilihan {
			if in.Data["penilaian_humptydumpty_skala"+n] == p {
				index = j
				break
			}
		}
		if index < 0 {
			return fmt.Errorf("%w: pilih %s", ErrValidasi, s.Label)
		}
		nilai := s.Nilai[index]
		in.Data["penilaian_humptydumpty_nilai"+n] = strconv.Itoa(nilai)
		total += nilai
	}
	in.Data["penilaian_humptydumpty_totalnilai"] = strconv.Itoa(total)
	return nil
}

func (r *Repositori) Daftar(ctx context.Context, no, username string, admin bool) (Hasil, error) {
	h := Hasil{Skala: SkalaHumptyDumpty, Catatan: []Catatan{}}
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
		total, e := strconv.Atoi(c.Data["penilaian_humptydumpty_totalnilai"])
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
