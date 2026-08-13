package permintaan_radiologi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInputTidakValid = errors.New("input permintaan radiologi tidak valid")

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

func (l *Layanan) CariTindakan(ctx context.Context, noRawat, kata string) ([]Tindakan, error) {
	if strings.TrimSpace(noRawat) == "" || len(strings.TrimSpace(kata)) < 2 {
		return []Tindakan{}, nil
	}
	return l.repo.CariTindakan(ctx, noRawat, kata)
}

func (l *Layanan) Simpan(ctx context.Context, input Input) (string, error) {
	bersihkan(&input)
	if err := validasi(input); err != nil {
		return "", err
	}
	return l.repo.Simpan(ctx, input)
}

func (l *Layanan) Ubah(ctx context.Context, nomor string, input Input) error {
	nomor = strings.TrimSpace(nomor)
	bersihkan(&input)
	if nomor == "" {
		return fmt.Errorf("%w: nomor permintaan wajib diisi", ErrInputTidakValid)
	}
	if err := validasi(input); err != nil {
		return err
	}
	return l.repo.Ubah(ctx, nomor, input)
}

func (l *Layanan) Hapus(ctx context.Context, noRawat, nomor string) error {
	noRawat = strings.TrimSpace(noRawat)
	nomor = strings.TrimSpace(nomor)
	if noRawat == "" || nomor == "" {
		return fmt.Errorf("%w: nomor rawat dan nomor permintaan wajib diisi", ErrInputTidakValid)
	}
	return l.repo.Hapus(ctx, noRawat, nomor)
}

func bersihkan(input *Input) {
	input.NoRawat = strings.TrimSpace(input.NoRawat)
	input.Tanggal = strings.TrimSpace(input.Tanggal)
	input.Jam = normalisasiJam(input.Jam)
	input.KodeDokter = strings.TrimSpace(input.KodeDokter)
	input.InformasiTambahan = strings.TrimSpace(input.InformasiTambahan)
	input.DiagnosisKlinis = strings.TrimSpace(input.DiagnosisKlinis)
	unik := make([]string, 0, len(input.KodeTindakan))
	dilihat := make(map[string]struct{}, len(input.KodeTindakan))
	for _, kode := range input.KodeTindakan {
		kode = strings.TrimSpace(kode)
		if kode == "" {
			continue
		}
		if _, ada := dilihat[kode]; ada {
			continue
		}
		dilihat[kode] = struct{}{}
		unik = append(unik, kode)
	}
	input.KodeTindakan = unik
}

func validasi(input Input) error {
	if input.NoRawat == "" || input.KodeDokter == "" {
		return fmt.Errorf("%w: nomor rawat dan dokter perujuk wajib diisi", ErrInputTidakValid)
	}
	if input.InformasiTambahan == "" || input.DiagnosisKlinis == "" {
		return fmt.Errorf("%w: informasi tambahan dan diagnosis klinis wajib diisi", ErrInputTidakValid)
	}
	if len(input.InformasiTambahan) > 60 || len(input.DiagnosisKlinis) > 80 {
		return fmt.Errorf("%w: informasi tambahan maksimal 60 karakter dan diagnosis klinis maksimal 80 karakter", ErrInputTidakValid)
	}
	if len(input.KodeTindakan) == 0 {
		return fmt.Errorf("%w: minimal satu tindakan radiologi wajib dipilih", ErrInputTidakValid)
	}
	if len(input.KodeTindakan) > 100 {
		return fmt.Errorf("%w: maksimal 100 tindakan dalam satu permintaan", ErrInputTidakValid)
	}
	if _, err := time.Parse("2006-01-02", input.Tanggal); err != nil {
		return fmt.Errorf("%w: tanggal permintaan tidak valid", ErrInputTidakValid)
	}
	if _, err := time.Parse("15:04:05", input.Jam); err != nil {
		return fmt.Errorf("%w: jam permintaan tidak valid", ErrInputTidakValid)
	}
	return nil
}

func normalisasiJam(value string) string {
	value = strings.TrimSpace(value)
	if len(value) == 5 {
		return value + ":00"
	}
	return value
}
