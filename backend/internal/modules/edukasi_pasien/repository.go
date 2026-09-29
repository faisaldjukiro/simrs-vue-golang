package edukasi_pasien

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	"simrs-backend/internal/shared/khanzamutasi"
)

var ErrValidasi = errors.New("catatan edukasi tidak valid")
var ErrAkses = errors.New("catatan hanya boleh diubah oleh petugas pencatat atau administrator")
var ErrKonflik = errors.New("catatan sudah ada, berubah, atau hilang; muat ulang riwayat")

const tabel = "catatan_edukasi"

type Input struct {
	NoRawat string            `json:"no_rawat"`
	Sumber  string            `json:"sumber"`
	Data    map[string]string `json:"data"`
	Asli    map[string]string `json:"asli"`
}

type Catatan struct {
	Data        map[string]string `json:"data"`
	Sumber      string            `json:"sumber"`
	BisaUbah    bool              `json:"bisa_ubah"`
	NamaPetugas string            `json:"nama_petugas"`
	NamaRuangan string            `json:"nama_ruangan"`
}

type Hasil struct {
	Catatan           []Catatan `json:"catatan"`
	PetugasLogin      Pilihan   `json:"petugas_login"`
	BolehPilihPetugas bool      `json:"boleh_pilih_petugas"`
}

type Pilihan struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

type Repositori struct{ simrs *sql.DB }

func NewRepositori(simrs *sql.DB) *Repositori { return &Repositori{simrs: simrs} }

func Kolom() []string {
	return []string{"tgl_perawatan", "jam_rawat", "kd_ruangan", "nip", "metode", "durasi", "materi", "penerima", "keterangan"}
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
	switch in.Data["metode"] {
	case "Audio", "Demonstrasi", "Lisan", "Tulisan", "Visual":
	default:
		return fmt.Errorf("%w: metode edukasi tidak valid", ErrValidasi)
	}
	// Sesuai struktur catatan_edukasi SIMRS yang dikonfirmasi pengguna.
	// Tolak kelebihan panjang, jangan memotong isi catatan klinis.
	for k, batas := range map[string]int{"kd_ruangan": 30, "nip": 20, "durasi": 30, "materi": 50, "penerima": 30, "keterangan": 255} {
		in.Data[k] = strings.TrimSpace(in.Data[k])
		if utf8.RuneCountInString(in.Data[k]) > batas {
			return fmt.Errorf("%w: %s maksimal %d karakter", ErrValidasi, k, batas)
		}
	}
	if in.Data["nip"] == "" {
		return fmt.Errorf("%w: petugas wajib diisi", ErrValidasi)
	}

	return nil
}

func (r *Repositori) Daftar(ctx context.Context, no, username string, admin bool) (Hasil, error) {
	h := Hasil{Catatan: []Catatan{}}
	if strings.TrimSpace(no) == "" || len(no) > 17 {
		return h, ErrValidasi
	}

	h.BolehPilihPetugas = admin
	err := r.simrs.QueryRowContext(ctx, "SELECT nip,nama FROM petugas WHERE nip=?", username).Scan(&h.PetugasLogin.Kode, &h.PetugasLogin.Nama)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return h, err
	}

	for _, sumber := range []struct {
		db          *sql.DB
		tabel, nama string
		ubah        bool
	}{
		{r.simrs, tabel, "SIMRS", true},
	} {
		selects := []string{"DATE_FORMAT(c.tgl_perawatan,'%Y-%m-%d')", "TIME_FORMAT(c.jam_rawat,'%H:%i:%s')"}
		for _, k := range Kolom()[2:] {
			selects = append(selects, "COALESCE(c."+k+",'')")
		}
		rows, err := sumber.db.QueryContext(ctx, "SELECT "+strings.Join(selects, ",")+" FROM "+sumber.tabel+" c WHERE c.no_rawat=? ORDER BY c.tgl_perawatan DESC,c.jam_rawat DESC", no)
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
			c := Catatan{Data: map[string]string{}, Sumber: sumber.nama, BisaUbah: sumber.ubah && (admin || (username != "" && nilai[3] == username))}
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
	sort.SliceStable(h.Catatan, func(i, j int) bool {
		return h.Catatan[i].Data["tgl_perawatan"]+h.Catatan[i].Data["jam_rawat"] > h.Catatan[j].Data["tgl_perawatan"]+h.Catatan[j].Data["jam_rawat"]
	})
	if err := r.lengkapiNama(ctx, h.Catatan); err != nil {
		return h, err
	}
	return h, nil
}

// Referensi selalu dibaca dari koneksi SIMRS lama.
func (r *Repositori) Referensi(ctx context.Context, jenis, q, username string, admin bool) ([]Pilihan, error) {
	query := ""
	switch jenis {
	case "petugas":
		query = "SELECT nip,nama FROM petugas WHERE (nip LIKE ? OR nama LIKE ?)"
	case "ruangan":
		query = "SELECT kd_ruangan,nama_ruangan FROM ruangan WHERE (kd_ruangan LIKE ? OR nama_ruangan LIKE ?)"
	default:
		return nil, ErrValidasi
	}
	q = strings.TrimSpace(q)
	if len(q) > 100 {
		return nil, ErrValidasi
	}
	args := []any{"%" + q + "%", "%" + q + "%"}
	if jenis == "petugas" && !admin {
		query += " AND nip=?"
		args = append(args, username)
	}
	rows, err := r.simrs.QueryContext(ctx, query+" ORDER BY 2 LIMIT 30", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hasil := []Pilihan{}
	for rows.Next() {
		var p Pilihan
		if err = rows.Scan(&p.Kode, &p.Nama); err != nil {
			return nil, err
		}
		hasil = append(hasil, p)
	}
	return hasil, rows.Err()
}

func (r *Repositori) lengkapiNama(ctx context.Context, catatan []Catatan) error {
	for _, ref := range []struct{ kolom, tabel, nama string }{
		{"nip", "petugas", "nama"}, {"kd_ruangan", "ruangan", "nama_ruangan"},
	} {
		keys := []any{}
		seen := map[string]bool{}
		for _, c := range catatan {
			v := c.Data[ref.kolom]
			if v != "" && !seen[v] {
				keys = append(keys, v)
				seen[v] = true
			}
		}
		if len(keys) == 0 {
			continue
		}
		marks := strings.TrimSuffix(strings.Repeat("?,", len(keys)), ",")
		rows, err := r.simrs.QueryContext(ctx, "SELECT "+ref.kolom+","+ref.nama+" FROM "+ref.tabel+" WHERE "+ref.kolom+" IN ("+marks+")", keys...)
		if err != nil {
			return err
		}
		nama := map[string]string{}
		for rows.Next() {
			var k, n string
			if err = rows.Scan(&k, &n); err != nil {
				rows.Close()
				return err
			}
			nama[k] = n
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		for i := range catatan {
			if ref.kolom == "nip" {
				catatan[i].NamaPetugas = nama[catatan[i].Data["nip"]]
			} else {
				catatan[i].NamaRuangan = nama[catatan[i].Data["kd_ruangan"]]
			}
		}
	}
	return nil
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
	if err = r.simrs.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM reg_periksa WHERE no_rawat=?)", in.NoRawat).Scan(&ada); err != nil {
		return err
	}
	if !ada {
		return fmt.Errorf("%w: kunjungan tidak ditemukan", ErrValidasi)
	}

	if metode != "DELETE" {
		for _, ref := range []struct{ tabel, kolom string }{{"petugas", "nip"}, {"ruangan", "kd_ruangan"}} {
			// Java tidak mewajibkan ruangan; nilai kosong tetap mengikuti aturan tabel SIMRS.
			if ref.kolom == "kd_ruangan" && in.Data[ref.kolom] == "" {
				continue
			}
			if err = r.simrs.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM "+ref.tabel+" WHERE "+ref.kolom+"=?)", in.Data[ref.kolom]).Scan(&ada); err != nil {
				return err
			}
			if !ada {
				return fmt.Errorf("%w: %s tidak ditemukan", ErrValidasi, ref.tabel)
			}
		}
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
