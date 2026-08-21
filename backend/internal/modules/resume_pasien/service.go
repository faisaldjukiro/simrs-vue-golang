package resume_pasien

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrInputTidakValid = errors.New("input resume pasien rawat jalan tidak valid")

type Layanan struct{ repo *Repositori }

func NewLayanan(repo *Repositori) *Layanan { return &Layanan{repo: repo} }

func (l *Layanan) Data(ctx context.Context, noRawat string) (Data, error) {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return Data{}, fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	return l.repo.Data(ctx, noRawat)
}

func (l *Layanan) ValidasiCoding(ctx context.Context, jenis, kode string, utama bool) (HasilValidasiCoding, error) {
	jenis = strings.ToLower(strings.TrimSpace(jenis))
	kode = normalisasiKode(kode)
	if jenis != "diagnosa" && jenis != "prosedur" {
		return HasilValidasiCoding{}, fmt.Errorf("%w: jenis coding harus diagnosa atau prosedur", ErrInputTidakValid)
	}
	if kode == "" {
		return HasilValidasiCoding{}, fmt.Errorf("%w: kode %s wajib diisi", ErrInputTidakValid, jenis)
	}
	return l.repo.ValidasiCoding(ctx, jenis, kode, utama)
}

func (l *Layanan) CariCoding(ctx context.Context, jenis, kataKunci string, utama bool) ([]HasilValidasiCoding, error) {
	jenis = strings.ToLower(strings.TrimSpace(jenis))
	kataKunci = strings.TrimSpace(kataKunci)
	if jenis != "diagnosa" && jenis != "prosedur" {
		return nil, fmt.Errorf("%w: jenis coding harus diagnosa atau prosedur", ErrInputTidakValid)
	}
	if len(kataKunci) < 2 {
		return []HasilValidasiCoding{}, nil
	}
	return l.repo.CariCoding(ctx, jenis, kataKunci, utama)
}

func (l *Layanan) Simpan(ctx context.Context, input Input) error {
	bersihkan(&input)
	if err := validasi(input); err != nil {
		return err
	}
	return l.repo.Simpan(ctx, input)
}

func (l *Layanan) Ubah(ctx context.Context, input Input) error {
	bersihkan(&input)
	if err := validasi(input); err != nil {
		return err
	}
	return l.repo.Ubah(ctx, input)
}

func (l *Layanan) Hapus(ctx context.Context, noRawat string) error {
	noRawat = strings.TrimSpace(noRawat)
	if noRawat == "" {
		return fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	return l.repo.Hapus(ctx, noRawat)
}

func bersihkan(i *Input) {
	i.NoRawat = strings.TrimSpace(i.NoRawat)
	i.KodeDokter = strings.TrimSpace(i.KodeDokter)
	for _, nilai := range []*string{
		&i.KondisiPulang, &i.KeluhanUtama, &i.JalannyaPenyakit, &i.PemeriksaanPenunjang, &i.HasilLaborat,
		&i.DiagnosaUtama, &i.KodeDiagnosaUtama, &i.DiagnosaSekunder, &i.KodeDiagnosaSekunder,
		&i.DiagnosaSekunder2, &i.KodeDiagnosaSekunder2, &i.DiagnosaSekunder3, &i.KodeDiagnosaSekunder3,
		&i.DiagnosaSekunder4, &i.KodeDiagnosaSekunder4, &i.ProsedurUtama, &i.KodeProsedurUtama,
		&i.ProsedurSekunder, &i.KodeProsedurSekunder, &i.ProsedurSekunder2, &i.KodeProsedurSekunder2,
		&i.ProsedurSekunder3, &i.KodeProsedurSekunder3, &i.ObatPulang,
	} {
		*nilai = strings.TrimSpace(*nilai)
	}
	if i.KondisiPulang == "" {
		i.KondisiPulang = "Hidup"
	}
}

func validasi(i Input) error {
	if i.NoRawat == "" {
		return fmt.Errorf("%w: nomor rawat wajib diisi", ErrInputTidakValid)
	}
	if i.KodeDokter == "" {
		return fmt.Errorf("%w: dokter wajib dipilih", ErrInputTidakValid)
	}
	if !ada(i.KondisiPulang, "Hidup", "Meninggal") {
		return fmt.Errorf("%w: kondisi pasien pulang tidak valid", ErrInputTidakValid)
	}
	if i.KeluhanUtama == "" {
		return fmt.Errorf("%w: keluhan utama dan riwayat penyakit wajib diisi", ErrInputTidakValid)
	}
	if i.JalannyaPenyakit == "" {
		return fmt.Errorf("%w: jalannya penyakit wajib diisi", ErrInputTidakValid)
	}
	if i.KodeDiagnosaUtama == "" || i.DiagnosaUtama == "" {
		return fmt.Errorf("%w: diagnosa utama wajib diisi", ErrInputTidakValid)
	}
	return nil
}

func ada(nilai string, pilihan ...string) bool {
	for _, item := range pilihan {
		if nilai == item {
			return true
		}
	}
	return false
}
