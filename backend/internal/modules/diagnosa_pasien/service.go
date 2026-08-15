package diagnosa_pasien

import (
	"context"
	"fmt"
	"strings"
)

type Layanan struct{ repositori *Repositori }

func NewLayanan(repositori *Repositori) *Layanan { return &Layanan{repositori: repositori} }

func (l *Layanan) Data(ctx context.Context, noRawat, status string) (Data, error) {
	noRawat = strings.TrimSpace(noRawat)
	status, err := normalisasiStatus(status)
	if err != nil {
		return Data{}, err
	}
	if noRawat == "" {
		return Data{}, inputTidakValid("nomor rawat wajib diisi")
	}
	return l.repositori.Data(ctx, noRawat, status)
}

func (l *Layanan) CariCoding(ctx context.Context, jenis, kataKunci string, utama bool) ([]Coding, error) {
	jenis = strings.ToLower(strings.TrimSpace(jenis))
	kataKunci = strings.TrimSpace(kataKunci)
	if jenis != "diagnosa" && jenis != "prosedur" {
		return nil, inputTidakValid("jenis coding harus diagnosa atau prosedur")
	}
	if len(kataKunci) < 3 {
		return []Coding{}, nil
	}
	return l.repositori.CariCoding(ctx, jenis, kataKunci, utama)
}

func (l *Layanan) Hapus(ctx context.Context, input HapusInput) error {
	input.NoRawat = strings.TrimSpace(input.NoRawat)
	input.Jenis = strings.ToLower(strings.TrimSpace(input.Jenis))
	input.Kode = normalisasiKode(input.Kode)
	if input.NoRawat == "" || input.Kode == "" {
		return inputTidakValid("nomor rawat dan kode coding wajib diisi")
	}
	if input.Jenis != "diagnosa" && input.Jenis != "prosedur" {
		return inputTidakValid("jenis coding harus diagnosa atau prosedur")
	}
	status, err := normalisasiStatus(input.Status)
	if err != nil {
		return err
	}
	input.Status = status
	return l.repositori.Hapus(ctx, input)
}

func (l *Layanan) Simpan(ctx context.Context, input Input) error {
	input.NoRawat = strings.TrimSpace(input.NoRawat)
	if input.NoRawat == "" {
		return inputTidakValid("nomor rawat wajib diisi")
	}
	status, err := normalisasiStatus(input.Status)
	if err != nil {
		return err
	}
	input.Status = status
	return l.repositori.Simpan(ctx, input)
}

func normalisasiStatus(status string) (string, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "ranap", "rawat inap":
		return "Ranap", nil
	case "ralan", "rawat jalan", "igd", "igd/ugd":
		return "Ralan", nil
	default:
		return "", fmt.Errorf("%w: status perawatan harus Ralan atau Ranap", ErrInputTidakValid)
	}
}
