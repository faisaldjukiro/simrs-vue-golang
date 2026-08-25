package ews_ranap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrInputTidakValid       = errors.New("input EWS Ranap tidak valid")
	ErrPetugasTidakDitemukan = errors.New("akun login belum terhubung ke data pegawai SIMRS")
	ErrCatatanSudahAda       = errors.New("catatan EWS Ranap pada tanggal dan jam tersebut sudah ada")
	ErrBillingTerkunci       = errors.New("kunjungan sudah masuk billing; EWS Ranap tidak dapat ditambah, diubah, atau dihapus")
)

var (
	PilihanAlat       = []string{"Ya", "Tidak"}
	PilihanKesadaran  = []string{"A", "P-V-U"}
	PilihanSkalaNyeri = []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
	}
)

type Data struct {
	Petugas            Petugas   `json:"petugas"`
	BisaMemilihPetugas bool      `json:"bisa_memilih_petugas"`
	PilihanAlat        []string  `json:"pilihan_alat"`
	PilihanKesadaran   []string  `json:"pilihan_kesadaran"`
	PilihanSkalaNyeri  []string  `json:"pilihan_skala_nyeri"`
	Catatan            []Catatan `json:"catatan"`
	BillingTerkunci    bool      `json:"billing_terkunci"`
}

type Input struct {
	KunciLama Kunci `json:"kunci_lama"`
	Catatan
}

type Layanan struct {
	repositori *Repositori
}

func NewLayanan(repositori *Repositori) *Layanan {
	return &Layanan{repositori: repositori}
}

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
	billingTerkunci, err := l.repositori.BillingTerkunci(ctx, noRawat)
	if err != nil {
		return Data{}, err
	}
	return Data{
		Petugas: petugas, BisaMemilihPetugas: aksesPenuh, PilihanAlat: PilihanAlat,
		PilihanKesadaran: PilihanKesadaran, PilihanSkalaNyeri: PilihanSkalaNyeri,
		Catatan: catatan, BillingTerkunci: billingTerkunci,
	}, nil
}

func (l *Layanan) CariPetugas(ctx context.Context, userID uint64, username, kataKunci string) ([]Petugas, error) {
	petugasLogin, aksesPenuh, err := l.identitas(ctx, userID, username)
	if err != nil {
		return nil, err
	}
	if !aksesPenuh {
		return []Petugas{petugasLogin}, nil
	}
	return l.repositori.CariPetugas(ctx, kataKunci)
}

func (l *Layanan) Simpan(ctx context.Context, userID uint64, username string, input Input) error {
	petugas, aksesPenuh, err := l.identitas(ctx, userID, username)
	if err != nil {
		return err
	}
	input.Catatan.NIP, err = l.nipPetugasInput(ctx, petugas, aksesPenuh, input.Catatan.NIP)
	if err != nil {
		return err
	}
	if err := validasiCatatan(&input.Catatan); err != nil {
		return err
	}
	if err := l.pastikanBelumBilling(ctx, input.Catatan.NoRawat); err != nil {
		return err
	}
	if err := l.repositori.Simpan(ctx, input.Catatan); err != nil {
		return petakanErrorMySQL(err)
	}
	return nil
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
		input.Catatan.NIP = petugas.NIP
	} else {
		input.Catatan.NIP, err = l.nipPetugasInput(ctx, petugas, true, input.Catatan.NIP)
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
	if input.Catatan.NoRawat != input.KunciLama.NoRawat {
		if err := l.pastikanBelumBilling(ctx, input.Catatan.NoRawat); err != nil {
			return err
		}
	}
	if err := l.repositori.Ubah(ctx, input.KunciLama, input.Catatan, petugas.NIP, aksesPenuh); err != nil {
		return petakanErrorMySQL(err)
	}
	return nil
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

func (l *Layanan) nipPetugasInput(ctx context.Context, petugasLogin Petugas, aksesPenuh bool, nipPilihan string) (string, error) {
	nipPilihan = strings.TrimSpace(nipPilihan)
	if !aksesPenuh || nipPilihan == "" {
		return petugasLogin.NIP, nil
	}
	petugas, err := l.repositori.PetugasLogin(ctx, nipPilihan)
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

func (l *Layanan) identitas(ctx context.Context, userID uint64, username string) (Petugas, bool, error) {
	petugas, err := l.repositori.PetugasLogin(ctx, username)
	if errors.Is(err, sql.ErrNoRows) {
		return Petugas{}, false, ErrPetugasTidakDitemukan
	}
	if err != nil {
		return Petugas{}, false, err
	}
	aksesPenuh, err := l.repositori.AksesPenuh(ctx, userID)
	if err != nil {
		return Petugas{}, false, err
	}
	return petugas, aksesPenuh, nil
}

func validasiKunci(kunci Kunci) error {
	if strings.TrimSpace(kunci.NoRawat) == "" {
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

func validasiCatatan(catatan *Catatan) error {
	bersihkanCatatan(catatan)
	if catatan.NoRawat == "" {
		return fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	if _, err := time.Parse("2006-01-02", catatan.Tanggal); err != nil {
		return fmt.Errorf("%w: tanggal EWS tidak valid", ErrInputTidakValid)
	}
	if !waktuValid(catatan.Jam) {
		return fmt.Errorf("%w: jam EWS tidak valid", ErrInputTidakValid)
	}
	wajib := map[string]string{
		"pernafasan": catatan.Pernafasan, "skor pernafasan": catatan.SkorPernafasan,
		"saturasi": catatan.Saturasi, "skor saturasi": catatan.SkorSaturasi,
		"alat bantu O2": catatan.Alat, "skor alat bantu O2": catatan.SkorAlat,
		"suhu": catatan.Suhu, "skor suhu": catatan.SkorSuhu,
		"denyut": catatan.Denyut, "skor denyut": catatan.SkorDenyut,
		"tekanan sistolik": catatan.Tekanan, "tekanan diastolik": catatan.Diastol,
		"skor tekanan": catatan.SkorTekanan, "kesadaran": catatan.Kesadaran,
		"skor kesadaran": catatan.SkorKesadaran, "total skor": catatan.TotalSkor,
		"klasifikasi": catatan.Klasifikasi, "respon klinis": catatan.Respon,
		"tindakan": catatan.Tindakan, "frekuensi monitoring": catatan.Frekuensi,
	}
	for nama, nilai := range wajib {
		if strings.TrimSpace(nilai) == "" {
			return fmt.Errorf("%w: %s wajib diisi", ErrInputTidakValid, nama)
		}
	}
	return nil
}

func bersihkanCatatan(catatan *Catatan) {
	catatan.NoRawat = strings.TrimSpace(catatan.NoRawat)
	catatan.Tanggal = strings.TrimSpace(catatan.Tanggal)
	catatan.Jam = normalisasiWaktu(catatan.Jam)
	catatan.NIP = strings.TrimSpace(catatan.NIP)
	catatan.Pernafasan = strings.TrimSpace(catatan.Pernafasan)
	catatan.SkorPernafasan = strings.TrimSpace(catatan.SkorPernafasan)
	catatan.Saturasi = strings.TrimSpace(catatan.Saturasi)
	catatan.SkorSaturasi = strings.TrimSpace(catatan.SkorSaturasi)
	catatan.Alat = strings.TrimSpace(catatan.Alat)
	catatan.SkorAlat = strings.TrimSpace(catatan.SkorAlat)
	catatan.Suhu = strings.TrimSpace(catatan.Suhu)
	catatan.SkorSuhu = strings.TrimSpace(catatan.SkorSuhu)
	catatan.Denyut = strings.TrimSpace(catatan.Denyut)
	catatan.SkorDenyut = strings.TrimSpace(catatan.SkorDenyut)
	catatan.Tekanan = strings.TrimSpace(catatan.Tekanan)
	catatan.Diastol = strings.TrimSpace(catatan.Diastol)
	catatan.SkorTekanan = strings.TrimSpace(catatan.SkorTekanan)
	catatan.Kesadaran = strings.TrimSpace(catatan.Kesadaran)
	catatan.SkorKesadaran = strings.TrimSpace(catatan.SkorKesadaran)
	catatan.TotalSkor = strings.TrimSpace(catatan.TotalSkor)
	catatan.Klasifikasi = strings.TrimSpace(catatan.Klasifikasi)
	catatan.Respon = strings.TrimSpace(catatan.Respon)
	catatan.Tindakan = strings.TrimSpace(catatan.Tindakan)
	catatan.Frekuensi = strings.TrimSpace(catatan.Frekuensi)
	catatan.SkalaNyeri = strings.TrimSpace(catatan.SkalaNyeri)
	catatan.BB = strings.TrimSpace(catatan.BB)
	catatan.TB = strings.TrimSpace(catatan.TB)
	catatan.LK = strings.TrimSpace(catatan.LK)
	catatan.LP = strings.TrimSpace(catatan.LP)
	catatan.Masuk1 = angkaDefault(catatan.Masuk1)
	catatan.Masuk2 = angkaDefault(catatan.Masuk2)
	catatan.JumlahMasuk = angkaDefault(catatan.JumlahMasuk)
	catatan.Keluar1 = angkaDefault(catatan.Keluar1)
	catatan.Keluar2 = angkaDefault(catatan.Keluar2)
	catatan.Keluar3 = angkaDefault(catatan.Keluar3)
	catatan.Keluar4 = angkaDefault(catatan.Keluar4)
	catatan.Keluar5 = angkaDefault(catatan.Keluar5)
	catatan.JumlahKeluar = angkaDefault(catatan.JumlahKeluar)
	catatan.BC = angkaDefault(catatan.BC)
}

func bersihkanKunci(kunci *Kunci) {
	kunci.NoRawat = strings.TrimSpace(kunci.NoRawat)
	kunci.Tanggal = strings.TrimSpace(kunci.Tanggal)
	kunci.Jam = normalisasiWaktu(kunci.Jam)
}

func angkaDefault(nilai string) string {
	nilai = strings.TrimSpace(nilai)
	if nilai == "" {
		return "0"
	}
	return nilai
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
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrCatatanSudahAda
	}
	return err
}
