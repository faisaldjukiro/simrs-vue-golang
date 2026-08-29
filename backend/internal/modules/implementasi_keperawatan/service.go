package implementasi_keperawatan

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrInputTidakValid       = errors.New("input implementasi keperawatan tidak valid")
	ErrPetugasTidakDitemukan = errors.New("akun login belum terhubung ke data petugas SIMRS")
	ErrCatatanSudahAda       = errors.New("catatan pada tanggal dan jam tersebut sudah ada")
	ErrBillingTerkunci       = errors.New("kunjungan sudah masuk billing; implementasi keperawatan tidak dapat ditambah, diubah, atau dihapus")
)

type Data struct {
	Petugas            Petugas   `json:"petugas"`
	BisaMemilihPetugas bool      `json:"bisa_memilih_petugas"`
	Catatan            []Catatan `json:"catatan"`
	BillingTerkunci    bool      `json:"billing_terkunci"`
}

type Input struct {
	KunciLama Kunci `json:"kunci_lama"`
	Catatan
}

type Layanan struct{ repositori *Repositori }

func NewLayanan(repositori *Repositori) *Layanan { return &Layanan{repositori: repositori} }

func (l *Layanan) Daftar(ctx context.Context, userID uint64, username, noRawat string) (Data, error) {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return Data{}, fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	petugas, aksesPenuh, err := l.identitas(ctx, userID, username)
	if err != nil {
		return Data{}, err
	}
	catatan, err := l.repositori.Daftar(ctx, noRawat, petugas.NIP, aksesPenuh)
	if err != nil {
		return Data{}, err
	}
	terkunci, err := l.repositori.BillingTerkunci(ctx, noRawat)
	if err != nil {
		return Data{}, err
	}
	return Data{Petugas: petugas, BisaMemilihPetugas: aksesPenuh, Catatan: catatan, BillingTerkunci: terkunci}, nil
}

func (l *Layanan) CariPetugas(ctx context.Context, userID uint64, username, kataKunci string) ([]Petugas, error) {
	petugas, aksesPenuh, err := l.identitas(ctx, userID, username)
	if err != nil {
		return nil, err
	}
	if !aksesPenuh {
		return []Petugas{petugas}, nil
	}
	return l.repositori.CariPetugas(ctx, kataKunci)
}

func (l *Layanan) Simpan(ctx context.Context, userID uint64, username string, input Input) error {
	petugas, aksesPenuh, err := l.identitas(ctx, userID, username)
	if err != nil {
		return err
	}
	input.NIP, err = l.nipInput(ctx, petugas, aksesPenuh, input.NIP)
	if err != nil {
		return err
	}
	if err := validasiCatatan(&input.Catatan); err != nil {
		return err
	}
	if err := l.pastikanBelumBilling(ctx, input.NoRawat); err != nil {
		return err
	}
	return petakanErrorMySQL(l.repositori.Simpan(ctx, input.Catatan))
}

func (l *Layanan) Ubah(ctx context.Context, userID uint64, username string, input Input) error {
	petugas, aksesPenuh, err := l.identitas(ctx, userID, username)
	if err != nil {
		return err
	}
	bersihkanKunci(&input.KunciLama)
	if err := validasiKunci(input.KunciLama); err != nil {
		return err
	}
	if !aksesPenuh {
		input.NIP = petugas.NIP
	} else {
		input.NIP, err = l.nipInput(ctx, petugas, true, input.NIP)
		if err != nil {
			return err
		}
	}
	if err := validasiCatatan(&input.Catatan); err != nil {
		return err
	}
	if err := l.pastikanBelumBilling(ctx, input.KunciLama.NoRawat); err != nil {
		return err
	}
	return petakanErrorMySQL(l.repositori.Ubah(ctx, input.KunciLama, input.Catatan, petugas.NIP, aksesPenuh))
}

func (l *Layanan) Hapus(ctx context.Context, userID uint64, username string, kunci Kunci) error {
	petugas, aksesPenuh, err := l.identitas(ctx, userID, username)
	if err != nil {
		return err
	}
	bersihkanKunci(&kunci)
	if err := validasiKunci(kunci); err != nil {
		return err
	}
	if err := l.pastikanBelumBilling(ctx, kunci.NoRawat); err != nil {
		return err
	}
	return l.repositori.Hapus(ctx, kunci, petugas.NIP, aksesPenuh)
}

func (l *Layanan) identitas(ctx context.Context, userID uint64, username string) (Petugas, bool, error) {
	petugas, err := l.repositori.PetugasByNIP(ctx, username)
	if errors.Is(err, sql.ErrNoRows) {
		return Petugas{}, false, ErrPetugasTidakDitemukan
	}
	if err != nil {
		return Petugas{}, false, err
	}
	aksesPenuh, err := l.repositori.AksesPenuh(ctx, userID)
	return petugas, aksesPenuh, err
}

func (l *Layanan) nipInput(ctx context.Context, login Petugas, aksesPenuh bool, pilihan string) (string, error) {
	pilihan = strings.TrimSpace(pilihan)
	if !aksesPenuh || pilihan == "" {
		return login.NIP, nil
	}
	petugas, err := l.repositori.PetugasByNIP(ctx, pilihan)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: petugas yang dipilih tidak ditemukan", ErrInputTidakValid)
	}
	if err != nil {
		return "", err
	}
	return petugas.NIP, nil
}

func (l *Layanan) pastikanBelumBilling(ctx context.Context, noRawat string) error {
	terkunci, err := l.repositori.BillingTerkunci(ctx, noRawat)
	if err != nil {
		return err
	}
	if terkunci {
		return ErrBillingTerkunci
	}
	return nil
}

func validasiCatatan(catatan *Catatan) error {
	bersihkanCatatan(catatan)
	if catatan.NoRawat == "" {
		return fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	if _, err := time.Parse("2006-01-02", catatan.Tanggal); err != nil {
		return fmt.Errorf("%w: tanggal wajib diisi dengan benar", ErrInputTidakValid)
	}
	if !waktuValid(catatan.Jam) {
		return fmt.Errorf("%w: jam wajib diisi dengan benar", ErrInputTidakValid)
	}
	if catatan.NIP == "" {
		return fmt.Errorf("%w: petugas wajib dipilih", ErrInputTidakValid)
	}
	if catatan.Uraian == "" {
		return fmt.Errorf("%w: uraian implementasi wajib diisi", ErrInputTidakValid)
	}
	if utf8.RuneCountInString(catatan.Uraian) > 1000 {
		return fmt.Errorf("%w: uraian maksimal 1000 karakter", ErrInputTidakValid)
	}
	return nil
}

func validasiKunci(kunci Kunci) error {
	if kunci.NoRawat == "" {
		return fmt.Errorf("%w: nomor rawat catatan lama wajib diisi", ErrInputTidakValid)
	}
	if _, err := time.Parse("2006-01-02", kunci.Tanggal); err != nil {
		return fmt.Errorf("%w: tanggal catatan lama tidak valid", ErrInputTidakValid)
	}
	if !waktuValid(kunci.Jam) {
		return fmt.Errorf("%w: jam catatan lama tidak valid", ErrInputTidakValid)
	}
	return nil
}

func bersihkanCatatan(catatan *Catatan) {
	catatan.NoRawat = strings.TrimSpace(catatan.NoRawat)
	catatan.Tanggal = strings.TrimSpace(catatan.Tanggal)
	catatan.Jam = normalisasiWaktu(catatan.Jam)
	catatan.Uraian = strings.TrimSpace(catatan.Uraian)
	catatan.NIP = strings.TrimSpace(catatan.NIP)
}

func bersihkanKunci(kunci *Kunci) {
	kunci.NoRawat = strings.TrimSpace(kunci.NoRawat)
	kunci.Tanggal = strings.TrimSpace(kunci.Tanggal)
	kunci.Jam = normalisasiWaktu(kunci.Jam)
}

func normalisasiWaktu(nilai string) string {
	nilai = strings.TrimSpace(nilai)
	if len(nilai) == 5 {
		return nilai + ":00"
	}
	return nilai
}

func waktuValid(nilai string) bool {
	_, err := time.Parse("15:04:05", normalisasiWaktu(nilai))
	return err == nil
}

func petakanErrorMySQL(err error) error {
	if err == nil {
		return nil
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrCatatanSudahAda
	}
	return err
}
