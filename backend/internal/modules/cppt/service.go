package cppt

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
	ErrInputTidakValid       = errors.New("input CPPT tidak valid")
	ErrPetugasTidakDitemukan = errors.New("akun login belum terhubung ke data pegawai SIMRS")
	ErrCatatanSudahAda       = errors.New("catatan CPPT pada tanggal dan jam tersebut sudah ada")
	ErrBillingTerkunci       = errors.New("kunjungan sudah masuk billing; catatan tidak dapat ditambah, diubah, atau dihapus")
)

var DaftarKesadaran = []string{"Compos Mentis", "Apatis", "Somnolence", "Sopor", "Coma"}

type Data struct {
	Petugas            Petugas   `json:"petugas"`
	BisaMemilihPetugas bool      `json:"bisa_memilih_petugas"`
	Kesadaran          []string  `json:"pilihan_kesadaran"`
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

func (l *Layanan) Daftar(ctx context.Context, userID uint64, username, noRawat, jenisRawat string) (Data, error) {
	noRawat = strings.TrimSpace(noRawat)
	jenisRawat = normalisasiJenisRawat(jenisRawat)
	if noRawat == "" {
		return Data{}, fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	petugas, aksesPenuh, err := l.identitas(ctx, userID, username)
	if err != nil {
		return Data{}, err
	}
	var catatan []Catatan
	if jenisRawat == "ralan" {
		catatan, err = l.repositori.DaftarRalan(ctx, noRawat, petugas.NIP, aksesPenuh)
	} else {
		catatan, err = l.repositori.Daftar(ctx, noRawat, petugas.NIP, aksesPenuh)
	}
	if err != nil {
		return Data{}, err
	}
	billingTerkunci, err := l.repositori.BillingTerkunci(ctx, noRawat)
	if err != nil {
		return Data{}, err
	}
	return Data{Petugas: petugas, BisaMemilihPetugas: aksesPenuh, Kesadaran: DaftarKesadaran, Catatan: catatan, BillingTerkunci: billingTerkunci}, nil
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
	if input.Catatan.JenisRawat == "ralan" {
		err = l.repositori.SimpanRalan(ctx, input.Catatan)
	} else {
		err = l.repositori.Simpan(ctx, input.Catatan)
	}
	if err != nil {
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
	if input.KunciLama.JenisRawat != input.Catatan.JenisRawat {
		return fmt.Errorf("%w: jenis rawat catatan lama dan baru harus sama", ErrInputTidakValid)
	}
	if err := l.pastikanBelumBilling(ctx, input.KunciLama.NoRawat); err != nil {
		return err
	}
	if input.Catatan.NoRawat != input.KunciLama.NoRawat {
		if err := l.pastikanBelumBilling(ctx, input.Catatan.NoRawat); err != nil {
			return err
		}
	}
	if input.Catatan.JenisRawat == "ralan" {
		err = l.repositori.UbahRalan(ctx, input.KunciLama, input.Catatan, petugas.NIP, aksesPenuh)
	} else {
		err = l.repositori.Ubah(ctx, input.KunciLama, input.Catatan, petugas.NIP, aksesPenuh)
	}
	if err != nil {
		return petakanErrorMySQL(err)
	}
	return nil
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
	if kunci.JenisRawat == "ralan" {
		return l.repositori.HapusRalan(ctx, kunci, petugas.NIP, aksesPenuh)
	}
	return l.repositori.Hapus(ctx, kunci, petugas.NIP, aksesPenuh)
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
	if kunci.JenisRawat != "ranap" && kunci.JenisRawat != "ralan" {
		return fmt.Errorf("%w: jenis rawat catatan lama tidak valid", ErrInputTidakValid)
	}
	if strings.TrimSpace(kunci.NoRawat) == "" {
		return fmt.Errorf("%w: nomor rawat catatan lama wajib diisi", ErrInputTidakValid)
	}
	if _, err := time.Parse("2006-01-02", kunci.TanggalPerawatan); err != nil {
		return fmt.Errorf("%w: tanggal catatan lama tidak valid", ErrInputTidakValid)
	}
	if !waktuValid(kunci.JamRawat) {
		return fmt.Errorf("%w: jam catatan lama tidak valid", ErrInputTidakValid)
	}
	return nil
}

func validasiCatatan(catatan *Catatan) error {
	bersihkanCatatan(catatan)
	if catatan.JenisRawat != "ranap" && catatan.JenisRawat != "ralan" {
		return fmt.Errorf("%w: jenis rawat tidak valid", ErrInputTidakValid)
	}
	if catatan.NoRawat == "" {
		return fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	if _, err := time.Parse("2006-01-02", catatan.TanggalPerawatan); err != nil {
		return fmt.Errorf("%w: tanggal perawatan tidak valid", ErrInputTidakValid)
	}
	if !waktuValid(catatan.JamRawat) {
		return fmt.Errorf("%w: jam rawat tidak valid", ErrInputTidakValid)
	}
	if !nilaiKesadaranValid(catatan.Kesadaran) {
		return fmt.Errorf("%w: pilihan kesadaran tidak valid", ErrInputTidakValid)
	}
	if !adaIsiKlinis(*catatan) {
		return fmt.Errorf("%w: isi minimal satu pemeriksaan atau catatan SOAP", ErrInputTidakValid)
	}
	panjang := map[string]struct {
		nilai string
		maks  int
	}{
		"suhu tubuh": {catatan.SuhuTubuh, 5}, "tensi": {catatan.Tensi, 8},
		"nadi": {catatan.Nadi, 3}, "respirasi": {catatan.Respirasi, 3},
		"tinggi": {catatan.Tinggi, 5}, "berat": {catatan.Berat, 5},
		"SpO2": {catatan.SpO2, 3}, "GCS": {catatan.GCS, 10},
		"alergi": {catatan.Alergi, 50}, "subjek": {catatan.Subjek, 2000},
		"lingkar perut": {catatan.LingkarPerut, 5},
		"objek":         {catatan.Objek, 2000}, "asesmen": {catatan.Asesmen, 2000},
		"plan": {catatan.Plan, 2000}, "instruksi": {catatan.Instruksi, 2000},
		"evaluasi": {catatan.Evaluasi, 2000},
	}
	for nama, aturan := range panjang {
		if utf8.RuneCountInString(aturan.nilai) > aturan.maks {
			return fmt.Errorf("%w: %s maksimal %d karakter", ErrInputTidakValid, nama, aturan.maks)
		}
	}
	return nil
}

func bersihkanCatatan(catatan *Catatan) {
	catatan.JenisRawat = normalisasiJenisRawat(catatan.JenisRawat)
	catatan.NoRawat = strings.TrimSpace(catatan.NoRawat)
	catatan.TanggalPerawatan = strings.TrimSpace(catatan.TanggalPerawatan)
	catatan.JamRawat = normalisasiWaktu(catatan.JamRawat)
	catatan.SuhuTubuh = strings.TrimSpace(catatan.SuhuTubuh)
	catatan.Tensi = strings.TrimSpace(catatan.Tensi)
	catatan.Nadi = strings.TrimSpace(catatan.Nadi)
	catatan.Respirasi = strings.TrimSpace(catatan.Respirasi)
	catatan.Tinggi = strings.TrimSpace(catatan.Tinggi)
	catatan.Berat = strings.TrimSpace(catatan.Berat)
	catatan.SpO2 = strings.TrimSpace(catatan.SpO2)
	catatan.GCS = strings.TrimSpace(catatan.GCS)
	catatan.Kesadaran = strings.TrimSpace(catatan.Kesadaran)
	catatan.Subjek = strings.TrimSpace(catatan.Subjek)
	catatan.Objek = strings.TrimSpace(catatan.Objek)
	catatan.Alergi = strings.TrimSpace(catatan.Alergi)
	catatan.LingkarPerut = strings.TrimSpace(catatan.LingkarPerut)
	catatan.Asesmen = strings.TrimSpace(catatan.Asesmen)
	catatan.Plan = strings.TrimSpace(catatan.Plan)
	catatan.Instruksi = strings.TrimSpace(catatan.Instruksi)
	catatan.Evaluasi = strings.TrimSpace(catatan.Evaluasi)
	catatan.NIP = strings.TrimSpace(catatan.NIP)
}

func adaIsiKlinis(catatan Catatan) bool {
	nilai := []string{catatan.SuhuTubuh, catatan.Tensi, catatan.Nadi, catatan.Respirasi,
		catatan.Tinggi, catatan.Berat, catatan.SpO2, catatan.GCS, catatan.Subjek,
		catatan.Objek, catatan.Alergi, catatan.LingkarPerut, catatan.Asesmen, catatan.Plan, catatan.Instruksi,
		catatan.Evaluasi}
	for _, item := range nilai {
		if item != "" {
			return true
		}
	}
	return false
}

func bersihkanKunci(kunci *Kunci) {
	kunci.JenisRawat = normalisasiJenisRawat(kunci.JenisRawat)
	kunci.NoRawat = strings.TrimSpace(kunci.NoRawat)
	kunci.TanggalPerawatan = strings.TrimSpace(kunci.TanggalPerawatan)
	kunci.JamRawat = normalisasiWaktu(kunci.JamRawat)
}

func normalisasiJenisRawat(nilai string) string {
	nilai = strings.ToLower(strings.TrimSpace(nilai))
	if nilai == "ralan" || nilai == "igd" || nilai == "rawat jalan" {
		return "ralan"
	}
	return "ranap"
}

func nilaiKesadaranValid(nilai string) bool {
	for _, pilihan := range DaftarKesadaran {
		if nilai == pilihan {
			return true
		}
	}
	return false
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
