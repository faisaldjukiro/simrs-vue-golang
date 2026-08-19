package awal_medis_ranap

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInputTidakValid = errors.New("input penilaian awal medis ranap tidak valid")
	ErrSudahAda        = errors.New("penilaian awal medis ranap sudah tersedia")
	ErrTidakDitemukan  = errors.New("penilaian awal medis ranap tidak ditemukan")
	ErrBillingTerkunci = errors.New("kunjungan sudah masuk billing; penilaian awal medis ranap tidak dapat ditambah, diubah, atau dihapus")
)

type Layanan struct{ repo *Repositori }

func NewLayanan(repo *Repositori) *Layanan { return &Layanan{repo: repo} }

func (l *Layanan) Data(ctx context.Context, noRawat string) (Data, error) {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return Data{}, fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	return l.repo.Data(ctx, noRawat)
}

func (l *Layanan) CariDokter(ctx context.Context, kata string) ([]Dokter, error) {
	if len(strings.TrimSpace(kata)) < 2 {
		return []Dokter{}, nil
	}
	return l.repo.CariDokter(ctx, kata)
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
	input.KodeDokter = strings.TrimSpace(input.KodeDokter)
	input.Anamnesis = strings.TrimSpace(input.Anamnesis)
	input.Hubungan = strings.TrimSpace(input.Hubungan)
	input.KeluhanUtama = strings.TrimSpace(input.KeluhanUtama)
	input.RPS = strings.TrimSpace(input.RPS)
	input.RPD = strings.TrimSpace(input.RPD)
	input.RPK = strings.TrimSpace(input.RPK)
	input.RPO = strings.TrimSpace(input.RPO)
}

func validasi(input Input) error {
	if input.NoRawat == "" || input.KodeDokter == "" {
		return fmt.Errorf("%w: pasien dan dokter wajib diisi", ErrInputTidakValid)
	}
	if _, err := time.Parse("2006-01-02 15:04:05", input.Tanggal); err != nil {
		return fmt.Errorf("%w: tanggal penilaian tidak valid", ErrInputTidakValid)
	}
	if input.Anamnesis != "Autoanamnesis" && input.Anamnesis != "Alloanamnesis" {
		return fmt.Errorf("%w: jenis anamnesis tidak valid", ErrInputTidakValid)
	}
	if input.KeluhanUtama == "" || input.RPS == "" || input.RPD == "" || input.RPK == "" || input.RPO == "" {
		return fmt.Errorf("%w: keluhan utama dan seluruh riwayat penyakit wajib diisi", ErrInputTidakValid)
	}
	return nil
}
