package awal_keperawatan_igd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInputTidakValid = errors.New("input penilaian awal keperawatan IGD tidak valid")
	ErrBillingTerkunci = errors.New("kunjungan sudah masuk billing; penilaian awal keperawatan IGD tidak dapat ditambah, diubah, atau dihapus")
)

type Input struct {
	NoRawat          string   `json:"no_rawat"`
	Tanggal          string   `json:"tanggal"`
	Informasi        string   `json:"informasi"`
	KeluhanUtama     string   `json:"keluhan_utama"`
	RPD              string   `json:"rpd"`
	RPO              string   `json:"rpo"`
	StatusKehamilan  string   `json:"status_kehamilan"`
	Gravida          string   `json:"gravida"`
	Para             string   `json:"para"`
	Abortus          string   `json:"abortus"`
	HPHT             string   `json:"hpht"`
	Tekanan          string   `json:"tekanan"`
	Pupil            string   `json:"pupil"`
	Neurosensorik    string   `json:"neurosensorik"`
	Integumen        string   `json:"integumen"`
	Turgor           string   `json:"turgor"`
	Edema            string   `json:"edema"`
	Mukosa           string   `json:"mukosa"`
	Perdarahan       string   `json:"perdarahan"`
	JumlahPerdarahan string   `json:"jumlah_perdarahan"`
	WarnaPerdarahan  string   `json:"warna_perdarahan"`
	Intoksikasi      string   `json:"intoksikasi"`
	BAB              string   `json:"bab"`
	XBAB             string   `json:"xbab"`
	KBAB             string   `json:"kbab"`
	WBAB             string   `json:"wbab"`
	BAK              string   `json:"bak"`
	XBAK             string   `json:"xbak"`
	WBAK             string   `json:"wbak"`
	LBAK             string   `json:"lbak"`
	Psikologis       string   `json:"psikologis"`
	Jiwa             string   `json:"jiwa"`
	Perilaku         string   `json:"perilaku"`
	Dilaporkan       string   `json:"dilaporkan"`
	Sebutkan         string   `json:"sebutkan"`
	Hubungan         string   `json:"hubungan"`
	TinggalDengan    string   `json:"tinggal_dengan"`
	KetTinggal       string   `json:"ket_tinggal"`
	Budaya           string   `json:"budaya"`
	KetBudaya        string   `json:"ket_budaya"`
	PendidikanPJ     string   `json:"pendidikan_pj"`
	KetPendidikanPJ  string   `json:"ket_pendidikan_pj"`
	Edukasi          string   `json:"edukasi"`
	KetEdukasi       string   `json:"ket_edukasi"`
	Kemampuan        string   `json:"kemampuan"`
	Aktifitas        string   `json:"aktifitas"`
	AlatBantu        string   `json:"alat_bantu"`
	KetBantu         string   `json:"ket_bantu"`
	Nyeri            string   `json:"nyeri"`
	Provokes         string   `json:"provokes"`
	KetProvokes      string   `json:"ket_provokes"`
	Quality          string   `json:"quality"`
	KetQuality       string   `json:"ket_quality"`
	Lokasi           string   `json:"lokasi"`
	Menyebar         string   `json:"menyebar"`
	SkalaNyeri       string   `json:"skala_nyeri"`
	Durasi           string   `json:"durasi"`
	NyeriHilang      string   `json:"nyeri_hilang"`
	KetNyeri         string   `json:"ket_nyeri"`
	PadaDokter       string   `json:"pada_dokter"`
	KetDokter        string   `json:"ket_dokter"`
	BerjalanA        string   `json:"berjalan_a"`
	BerjalanB        string   `json:"berjalan_b"`
	BerjalanC        string   `json:"berjalan_c"`
	Hasil            string   `json:"hasil"`
	Lapor            string   `json:"lapor"`
	KetLapor         string   `json:"ket_lapor"`
	Rencana          string   `json:"rencana"`
	NIP              string   `json:"nip"`
	KodeMasalah      []string `json:"kode_masalah"`
	KodeRencana      []string `json:"kode_rencana"`
}

type Layanan struct{ repo *Repositori }

func NewLayanan(repo *Repositori) *Layanan { return &Layanan{repo: repo} }

func (l *Layanan) Data(ctx context.Context, noRawat, username string) (Data, error) {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return Data{}, fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	return l.repo.Data(ctx, noRawat, username)
}

func (l *Layanan) Simpan(ctx context.Context, input Input, ubah bool) error {
	bersihkan(&input)
	if err := validasi(input); err != nil {
		return err
	}
	return l.repo.Simpan(ctx, input, ubah)
}

func (l *Layanan) Hapus(ctx context.Context, noRawat string) error {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	return l.repo.Hapus(ctx, noRawat)
}

func bersihkan(input *Input) {
	input.NoRawat = strings.TrimSpace(input.NoRawat)
	input.Tanggal = strings.ReplaceAll(strings.TrimSpace(input.Tanggal), "T", " ")
	if len(input.Tanggal) == 16 {
		input.Tanggal += ":00"
	}
	input.Informasi = strings.TrimSpace(input.Informasi)
	input.KeluhanUtama = strings.TrimSpace(input.KeluhanUtama)
	input.RPD = strings.TrimSpace(input.RPD)
	input.RPO = strings.TrimSpace(input.RPO)
	input.NIP = strings.TrimSpace(input.NIP)
	input.KodeMasalah = unik(input.KodeMasalah)
	input.KodeRencana = unik(input.KodeRencana)
}

func validasi(input Input) error {
	if input.NoRawat == "" || input.NIP == "" {
		return fmt.Errorf("%w: pasien dan petugas wajib diisi", ErrInputTidakValid)
	}
	if input.KeluhanUtama == "" || input.RPD == "" || input.RPO == "" {
		return fmt.Errorf("%w: riwayat penyakit sekarang, dahulu, dan penggunaan obat wajib diisi", ErrInputTidakValid)
	}
	if _, err := time.Parse("2006-01-02 15:04:05", input.Tanggal); err != nil {
		return fmt.Errorf("%w: tanggal penilaian tidak valid", ErrInputTidakValid)
	}
	if input.Informasi != "Autoanamnesis" && input.Informasi != "Alloanamnesis" {
		return fmt.Errorf("%w: sumber informasi tidak valid", ErrInputTidakValid)
	}
	return nil
}

func unik(values []string) []string {
	hasil, dilihat := make([]string, 0, len(values)), map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !dilihat[value] {
			dilihat[value] = true
			hasil = append(hasil, value)
		}
	}
	return hasil
}
