package triase_igd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInputTidakValid = errors.New("input triase IGD tidak valid")

type Input struct {
	NoRawat              string   `json:"no_rawat"`
	Jenis                string   `json:"jenis"`
	TanggalKunjungan     string   `json:"tanggal_kunjungan"`
	CaraMasuk            string   `json:"cara_masuk"`
	AlatTransportasi     string   `json:"alat_transportasi"`
	AlasanKedatangan     string   `json:"alasan_kedatangan"`
	KeteranganKedatangan string   `json:"keterangan_kedatangan"`
	KodeKasus            string   `json:"kode_kasus"`
	TekananDarah         string   `json:"tekanan_darah"`
	Nadi                 string   `json:"nadi"`
	Pernapasan           string   `json:"pernapasan"`
	Suhu                 string   `json:"suhu"`
	SaturasiO2           string   `json:"saturasi_o2"`
	Nyeri                string   `json:"nyeri"`
	HandOver             string   `json:"hand_over"`
	IsiUtama             string   `json:"isi_utama"`
	KebutuhanKhusus      string   `json:"kebutuhan_khusus"`
	Catatan              string   `json:"catatan"`
	Plan                 string   `json:"plan"`
	TanggalTriase        string   `json:"tanggal_triase"`
	NIP                  string   `json:"nip"`
	Skala                int      `json:"skala"`
	KodeKriteria         []string `json:"kode_kriteria"`
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

func (l *Layanan) Hapus(ctx context.Context, noRawat, jenis string) error {
	noRawat, jenis = strings.TrimSpace(noRawat), strings.ToLower(strings.TrimSpace(jenis))
	if noRawat == "" || (jenis != "primer" && jenis != "sekunder") {
		return fmt.Errorf("%w: nomor rawat dan jenis triase wajib diisi", ErrInputTidakValid)
	}
	return l.repo.Hapus(ctx, noRawat, jenis)
}

func bersihkan(input *Input) {
	input.NoRawat = strings.TrimSpace(input.NoRawat)
	input.Jenis = strings.ToLower(strings.TrimSpace(input.Jenis))
	input.TanggalKunjungan = normalisasiTanggal(input.TanggalKunjungan)
	input.TanggalTriase = normalisasiTanggal(input.TanggalTriase)
	input.CaraMasuk = strings.TrimSpace(input.CaraMasuk)
	input.AlatTransportasi = strings.TrimSpace(input.AlatTransportasi)
	input.AlasanKedatangan = strings.TrimSpace(input.AlasanKedatangan)
	input.KeteranganKedatangan = strings.TrimSpace(input.KeteranganKedatangan)
	input.KodeKasus = strings.TrimSpace(input.KodeKasus)
	input.TekananDarah = strings.TrimSpace(input.TekananDarah)
	input.Nadi = strings.TrimSpace(input.Nadi)
	input.Pernapasan = strings.TrimSpace(input.Pernapasan)
	input.Suhu = strings.TrimSpace(input.Suhu)
	input.SaturasiO2 = strings.TrimSpace(input.SaturasiO2)
	input.Nyeri = strings.TrimSpace(input.Nyeri)
	input.HandOver = strings.TrimSpace(input.HandOver)
	input.IsiUtama = strings.TrimSpace(input.IsiUtama)
	input.KebutuhanKhusus = strings.TrimSpace(input.KebutuhanKhusus)
	input.Catatan = strings.TrimSpace(input.Catatan)
	input.Plan = strings.TrimSpace(input.Plan)
	input.NIP = strings.TrimSpace(input.NIP)
	unik := make([]string, 0, len(input.KodeKriteria))
	dilihat := map[string]bool{}
	for _, kode := range input.KodeKriteria {
		kode = strings.TrimSpace(kode)
		if kode != "" && !dilihat[kode] {
			dilihat[kode] = true
			unik = append(unik, kode)
		}
	}
	input.KodeKriteria = unik
}

func normalisasiTanggal(value string) string {
	value = strings.TrimSpace(value)
	if len(value) == 16 {
		return value + ":00"
	}
	return strings.ReplaceAll(value, "T", " ")
}

func validasi(input Input) error {
	if input.NoRawat == "" || input.KodeKasus == "" || input.NIP == "" {
		return fmt.Errorf("%w: pasien, macam kasus, dan petugas wajib diisi", ErrInputTidakValid)
	}
	if input.KeteranganKedatangan == "" {
		return fmt.Errorf("%w: keterangan kedatangan wajib diisi", ErrInputTidakValid)
	}
	if input.Jenis != "primer" && input.Jenis != "sekunder" {
		return fmt.Errorf("%w: jenis harus primer atau sekunder", ErrInputTidakValid)
	}
	if _, err := time.Parse("2006-01-02 15:04:05", input.TanggalKunjungan); err != nil {
		return fmt.Errorf("%w: tanggal kunjungan tidak valid", ErrInputTidakValid)
	}
	if _, err := time.Parse("2006-01-02 15:04:05", input.TanggalTriase); err != nil {
		return fmt.Errorf("%w: tanggal triase tidak valid", ErrInputTidakValid)
	}
	if input.IsiUtama == "" || input.Catatan == "" {
		return fmt.Errorf("%w: %s dan catatan wajib diisi", ErrInputTidakValid, map[bool]string{true: "keluhan utama", false: "anamnesa singkat"}[input.Jenis == "primer"])
	}
	if !pilihanValid(input.CaraMasuk, "Jalan", "Brankar", "Kursi Roda", "Digendong") ||
		!pilihanValid(input.AlatTransportasi, "-", "AGD", "Sendiri", "Swasta") ||
		!pilihanValid(input.AlasanKedatangan, "Datang Sendiri", "Polisi", "Rujukan", "-") ||
		!pilihanValid(input.HandOver, "Bedah", "Penyakit Dalam", "OBGIN", "Anak") {
		return fmt.Errorf("%w: pilihan data kedatangan tidak valid", ErrInputTidakValid)
	}
	if input.Jenis == "primer" && !pilihanValid(input.KebutuhanKhusus, "-", "UPPA", "Airborne", "Dekontaminan") {
		return fmt.Errorf("%w: kebutuhan khusus tidak valid", ErrInputTidakValid)
	}
	if input.TekananDarah == "" || input.Nadi == "" || input.Pernapasan == "" || input.Suhu == "" || input.SaturasiO2 == "" || input.Nyeri == "" {
		return fmt.Errorf("%w: seluruh tanda vital triase wajib diisi", ErrInputTidakValid)
	}
	if len(input.KeteranganKedatangan) > 100 || len(input.Catatan) > 100 || len(input.IsiUtama) > 400 {
		return fmt.Errorf("%w: keterangan/catatan maksimal 100 dan keluhan/anamnesa maksimal 400 karakter", ErrInputTidakValid)
	}
	if len(input.Suhu) > 5 || len(input.Nyeri) > 5 || len(input.TekananDarah) > 8 || len(input.Nadi) > 3 || len(input.SaturasiO2) > 3 || len(input.Pernapasan) > 3 {
		return fmt.Errorf("%w: panjang tanda vital melebihi batas SIMRS", ErrInputTidakValid)
	}
	if len(input.KodeKriteria) == 0 {
		return fmt.Errorf("%w: minimal satu kriteria skala wajib dipilih", ErrInputTidakValid)
	}
	if input.Jenis == "primer" && (input.Skala < 1 || input.Skala > 2) {
		return fmt.Errorf("%w: triase primer hanya memakai skala 1 atau 2", ErrInputTidakValid)
	}
	if input.Jenis == "sekunder" && (input.Skala < 3 || input.Skala > 5) {
		return fmt.Errorf("%w: triase sekunder hanya memakai skala 3, 4, atau 5", ErrInputTidakValid)
	}
	planPrimer := map[string]bool{"Ruang Resusitasi": true, "Ruang Kritis": true, "Zona Kuning": true, "Zona Hijau": true, "Zona Hitam": true}
	planSekunder := map[string]bool{"Zona Kuning": true, "Zona Hijau": true}
	if (input.Jenis == "primer" && !planPrimer[input.Plan]) || (input.Jenis == "sekunder" && !planSekunder[input.Plan]) {
		return fmt.Errorf("%w: plan/keputusan triase tidak valid", ErrInputTidakValid)
	}
	return nil
}

func pilihanValid(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}
