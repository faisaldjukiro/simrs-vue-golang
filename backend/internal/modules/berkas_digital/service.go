package berkas_digital

import (
	"context"
	"strings"
)

type Layanan struct {
	repositori *Repositori
}

func NewLayanan(repositori *Repositori) *Layanan {
	return &Layanan{repositori: repositori}
}

func (l *Layanan) Data(ctx context.Context, noRawat string) (Data, error) {
	return l.repositori.Data(ctx, noRawat)
}

func (l *Layanan) Simpan(ctx context.Context, input Input) error {
	input.NoRawat = strings.TrimSpace(input.NoRawat)
	input.Kode = strings.TrimSpace(input.Kode)
	input.LokasiFile = strings.TrimSpace(input.LokasiFile)
	if input.NoRawat == "" || input.Kode == "" || input.LokasiFile == "" {
		return ErrInputTidakValid
	}
	return l.repositori.Simpan(ctx, input)
}

func (l *Layanan) PastikanBisaUpload(ctx context.Context, noRawat, kode string) error {
	noRawat = strings.TrimSpace(noRawat)
	kode = strings.TrimSpace(kode)
	if noRawat == "" || kode == "" {
		return ErrInputTidakValid
	}
	return l.repositori.PastikanBisaSimpan(ctx, noRawat, kode)
}

func (l *Layanan) Hapus(ctx context.Context, kunci Kunci) error {
	kunci.NoRawat = strings.TrimSpace(kunci.NoRawat)
	kunci.Kode = strings.TrimSpace(kunci.Kode)
	kunci.LokasiFile = strings.TrimSpace(kunci.LokasiFile)
	if kunci.NoRawat == "" || kunci.Kode == "" || kunci.LokasiFile == "" {
		return ErrInputTidakValid
	}
	return l.repositori.Hapus(ctx, kunci)
}
