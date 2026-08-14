package penanganan_dokter_petugas

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

var ErrInputTidakValid = errors.New("input penanganan dokter dan petugas tidak valid")

type Data struct {
	Catatan         []Catatan `json:"catatan"`
	BillingTerkunci bool      `json:"billing_terkunci"`
	DokterDPJP      *Dokter   `json:"dokter_dpjp,omitempty"`
	PetugasLogin    *Petugas  `json:"petugas_login,omitempty"`
}

type Input struct {
	KunciLama Kunci `json:"kunci_lama"`
	Catatan
}

type InputBanyak struct {
	Catatan []Catatan `json:"catatan"`
}

type Layanan struct{ repo *Repositori }

func NewLayanan(repo *Repositori) *Layanan { return &Layanan{repo: repo} }

func (l *Layanan) Daftar(ctx context.Context, username, noRawat, jenisRawat string) (Data, error) {
	noRawat = strings.TrimSpace(noRawat)
	jenisRawat = normalisasiJenisRawat(jenisRawat)
	if noRawat == "" {
		return Data{}, fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	var catatan []Catatan
	var terkunci bool
	var err error
	if jenisRawat == "ralan" {
		catatan, terkunci, err = l.repo.DaftarRalan(ctx, noRawat)
	} else {
		catatan, terkunci, err = l.repo.Daftar(ctx, noRawat)
	}
	if err != nil {
		return Data{}, err
	}
	data := Data{Catatan: catatan, BillingTerkunci: terkunci}
	var dokter Dokter
	if jenisRawat == "ralan" {
		dokter, err = l.repo.DokterRalan(ctx, noRawat)
	} else {
		dokter, err = l.repo.DokterDPJP(ctx, noRawat)
	}
	if err == nil {
		data.DokterDPJP = &dokter
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Data{}, err
	}
	if petugas, err := l.repo.PetugasLogin(ctx, username); err == nil {
		data.PetugasLogin = &petugas
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Data{}, err
	}
	return data, nil
}

func (l *Layanan) CariDokter(ctx context.Context, kata string) ([]Dokter, error) {
	if len(strings.TrimSpace(kata)) < 2 {
		return []Dokter{}, nil
	}
	return l.repo.CariDokter(ctx, kata)
}
func (l *Layanan) CariPetugas(ctx context.Context, kata string) ([]Petugas, error) {
	if len(strings.TrimSpace(kata)) < 2 {
		return []Petugas{}, nil
	}
	return l.repo.CariPetugas(ctx, kata)
}
func (l *Layanan) CariTindakan(ctx context.Context, noRawat, kata, jenisRawat string) ([]Tindakan, error) {
	if strings.TrimSpace(noRawat) == "" || len(strings.TrimSpace(kata)) < 2 {
		return []Tindakan{}, nil
	}
	if normalisasiJenisRawat(jenisRawat) == "ralan" {
		return l.repo.CariTindakanRalan(ctx, noRawat, kata)
	}
	return l.repo.CariTindakan(ctx, noRawat, kata)
}

func (l *Layanan) Simpan(ctx context.Context, input Input) error {
	bersihkan(&input.Catatan)
	if err := validasiCatatan(input.Catatan); err != nil {
		return err
	}
	if input.Catatan.JenisRawat == "ralan" {
		return petakanMySQL(l.repo.SimpanBanyakRalan(ctx, []Catatan{input.Catatan}))
	}
	return petakanMySQL(l.repo.Simpan(ctx, input.Catatan))
}

func (l *Layanan) SimpanBanyak(ctx context.Context, input InputBanyak) error {
	if len(input.Catatan) == 0 {
		return fmt.Errorf("%w: minimal satu tindakan wajib dipilih", ErrInputTidakValid)
	}
	if len(input.Catatan) > 100 {
		return fmt.Errorf("%w: maksimal 100 tindakan dalam sekali simpan", ErrInputTidakValid)
	}

	kunci := make(map[string]struct{}, len(input.Catatan))
	noRawat := ""
	jenisRawat := ""
	for indeks := range input.Catatan {
		bersihkan(&input.Catatan[indeks])
		item := input.Catatan[indeks]
		if err := validasiCatatan(item); err != nil {
			return fmt.Errorf("tindakan ke-%d: %w", indeks+1, err)
		}
		if noRawat == "" {
			noRawat = item.NoRawat
			jenisRawat = item.JenisRawat
		} else if item.NoRawat != noRawat {
			return fmt.Errorf("%w: seluruh tindakan harus berasal dari nomor rawat yang sama", ErrInputTidakValid)
		} else if item.JenisRawat != jenisRawat {
			return fmt.Errorf("%w: seluruh tindakan harus berasal dari jenis rawat yang sama", ErrInputTidakValid)
		}
		id := strings.Join([]string{item.NoRawat, item.KodeTindakan, item.KodeDokter, item.KodePetugas, item.Tanggal, item.Jam}, "|")
		if _, ada := kunci[id]; ada {
			return fmt.Errorf("%w: tindakan %s dipilih lebih dari satu kali", ErrInputTidakValid, item.KodeTindakan)
		}
		kunci[id] = struct{}{}
	}

	if input.Catatan[0].JenisRawat == "ralan" {
		return petakanMySQL(l.repo.SimpanBanyakRalan(ctx, input.Catatan))
	}
	return petakanMySQL(l.repo.SimpanBanyak(ctx, input.Catatan))
}
func (l *Layanan) Ubah(ctx context.Context, input Input) error {
	bersihkan(&input.Catatan)
	bersihkanKunci(&input.KunciLama)
	if err := validasiKunci(input.KunciLama); err != nil {
		return err
	}
	if err := validasiCatatan(input.Catatan); err != nil {
		return err
	}
	if input.Catatan.JenisRawat != input.KunciLama.JenisRawat {
		return fmt.Errorf("%w: jenis rawat data lama dan baru harus sama", ErrInputTidakValid)
	}
	if input.Catatan.JenisRawat == "ralan" {
		return petakanMySQL(l.repo.UbahRalan(ctx, input.KunciLama, input.Catatan))
	}
	return petakanMySQL(l.repo.Ubah(ctx, input.KunciLama, input.Catatan))
}
func (l *Layanan) Hapus(ctx context.Context, k Kunci) error {
	bersihkanKunci(&k)
	if err := validasiKunci(k); err != nil {
		return err
	}
	if k.JenisRawat == "ralan" {
		return l.repo.HapusRalan(ctx, k)
	}
	return l.repo.Hapus(ctx, k)
}

func bersihkan(c *Catatan) {
	c.JenisRawat = normalisasiJenisRawat(c.JenisRawat)
	c.NoRawat = strings.TrimSpace(c.NoRawat)
	c.KodeTindakan = strings.TrimSpace(c.KodeTindakan)
	c.KodeDokter = strings.TrimSpace(c.KodeDokter)
	c.KodePetugas = strings.TrimSpace(c.KodePetugas)
	c.Tanggal = strings.TrimSpace(c.Tanggal)
	c.Jam = normalisasiJam(c.Jam)
}
func bersihkanKunci(k *Kunci) {
	k.JenisRawat = normalisasiJenisRawat(k.JenisRawat)
	k.NoRawat = strings.TrimSpace(k.NoRawat)
	k.KodeTindakan = strings.TrimSpace(k.KodeTindakan)
	k.KodeDokter = strings.TrimSpace(k.KodeDokter)
	k.KodePetugas = strings.TrimSpace(k.KodePetugas)
	k.Tanggal = strings.TrimSpace(k.Tanggal)
	k.Jam = normalisasiJam(k.Jam)
}
func validasiCatatan(c Catatan) error {
	if c.JenisRawat != "ranap" && c.JenisRawat != "ralan" {
		return fmt.Errorf("%w: jenis rawat tidak valid", ErrInputTidakValid)
	}
	if c.NoRawat == "" || c.KodeTindakan == "" || c.KodeDokter == "" || c.KodePetugas == "" {
		return fmt.Errorf("%w: nomor rawat, dokter, petugas, dan tindakan wajib diisi", ErrInputTidakValid)
	}
	return validasiWaktu(c.Tanggal, c.Jam)
}
func validasiKunci(k Kunci) error {
	if k.JenisRawat != "ranap" && k.JenisRawat != "ralan" {
		return fmt.Errorf("%w: jenis rawat kunci tidak valid", ErrInputTidakValid)
	}
	if k.NoRawat == "" || k.KodeTindakan == "" || k.KodeDokter == "" || k.KodePetugas == "" {
		return fmt.Errorf("%w: kunci catatan lama tidak lengkap", ErrInputTidakValid)
	}
	return validasiWaktu(k.Tanggal, k.Jam)
}
func validasiWaktu(tanggal, jam string) error {
	if _, err := time.Parse("2006-01-02", tanggal); err != nil {
		return fmt.Errorf("%w: tanggal tidak valid", ErrInputTidakValid)
	}
	if _, err := time.Parse("15:04:05", jam); err != nil {
		return fmt.Errorf("%w: jam tidak valid", ErrInputTidakValid)
	}
	return nil
}
func normalisasiJam(v string) string {
	v = strings.TrimSpace(v)
	if len(v) == 5 {
		return v + ":00"
	}
	return v
}
func normalisasiJenisRawat(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "ralan" || v == "igd" || v == "rawat jalan" {
		return "ralan"
	}
	return "ranap"
}
func petakanMySQL(err error) error {
	if err == nil {
		return nil
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return fmt.Errorf("%w: tindakan pada tanggal dan jam tersebut sudah ada", ErrInputTidakValid)
	}
	return err
}
